package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Kaffyn/Vectora/core/db"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/Kaffyn/Vectora/core/policies"
)

// Tenant representa um workspace isolado com seus próprios recursos.
type Tenant struct {
	ID           string
	Root         string
	ProjectName  string
	VectorStore  *db.ChromemStore // Type confirmed from core/db/vector.go
	KVStore      *db.BBoltStore   // Type confirmed from core/db/store.go
	Guardian     *policies.Guardian
	LastActivity time.Time
}

// TenantManager gerencia o ciclo de vida (carregamento/eviction) dos tenants ativos.
type TenantManager struct {
	mu             sync.RWMutex
	activeTenants  map[string]*Tenant
	evictionPolicy EvictionPolicy
	osMgr          vecos.OSManager
}

// EvictionPolicy define as regras para descarregar um tenant da memória.
type EvictionPolicy struct {
	IdleTimeout time.Duration // Tempo de inatividade antes do unload (padrão: 30m)
	MaxTenants  int           // Máximo de tenants ativos simultâneos
}

func NewTenantManager(policy EvictionPolicy) (*TenantManager, error) {
	osMgr, err := vecos.NewManager()
	if err != nil {
		return nil, err
	}

	return &TenantManager{
		activeTenants:  make(map[string]*Tenant),
		evictionPolicy: policy,
		osMgr:          osMgr,
	}, nil
}

// GetOrCreateTenant obtém um tenant existente ou inicializa um novo para o workspace.
func (tm *TenantManager) GetOrCreateTenant(wsID, wsRoot, projName string) (*Tenant, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 1. Verificar se já está na memória
	if t, exists := tm.activeTenants[wsID]; exists {
		t.LastActivity = time.Now()
		return t, nil
	}

	// 2. Preparar diretórios de isolamento
	baseDir, _ := tm.osMgr.GetAppDataDir()
	tenantDataDir := filepath.Join(baseDir, "workspaces", wsID)
	if err := os.MkdirAll(tenantDataDir, 0755); err != nil {
		return nil, fmt.Errorf("mtp_err: failed to create tenant directory: %w", err)
	}

	// 3. Inicializar Stores isolados
	vecStore, err := db.NewVectorStoreAtPath(filepath.Join(tenantDataDir, "chromadb"))
	if err != nil {
		return nil, fmt.Errorf("mtp_err: failed to init tenant vector store: %w", err)
	}

	kvStore, err := db.NewKVStoreAtPath(filepath.Join(tenantDataDir, "chat_history.db"))
	if err != nil {
		vecStore.Close()
		return nil, fmt.Errorf("mtp_err: failed to init tenant kv store: %w", err)
	}

	// 4. Criar o Tenant
	tenant := &Tenant{
		ID:           wsID,
		Root:         wsRoot,
		ProjectName:  projName,
		VectorStore:  vecStore,
		KVStore:      kvStore,
		Guardian:     policies.NewGuardian(wsRoot),
		LastActivity: time.Now(),
	}

	tm.activeTenants[wsID] = tenant
	return tenant, nil
}

// ReleaseTenant força o fechamento e descarregamento de um tenant.
func (tm *TenantManager) ReleaseTenant(wsID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if t, exists := tm.activeTenants[wsID]; exists {
		t.VectorStore.Close()
		t.KVStore.Close()
		delete(tm.activeTenants, wsID)
	}

	return nil
}

// StartEvictionRoutine inicia uma goroutine que limpa tenants inativos em background.
func (tm *TenantManager) StartEvictionRoutine(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				tm.evictIdleTenants()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (tm *TenantManager) evictIdleTenants() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()
	for id, t := range tm.activeTenants {
		if now.Sub(t.LastActivity) > tm.evictionPolicy.IdleTimeout {
			t.VectorStore.Close()
			t.KVStore.Close()
			delete(tm.activeTenants, id)
		}
	}
}
