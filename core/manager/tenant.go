package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/engine"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/policies"
	"github.com/Kaffyn/Vectora/core/tools"
)

// Tenant representa um workspace isolado no daemon singleton
// Cada tenant possui seus próprios índices vetoriais, histórico de chat e regras Guardian
type Tenant struct {
	ID           string
	Root         string
	ProjectName  string
	VectorStore  db.VectorStore
	KVStore      db.KVStore
	Guardian     *policies.Guardian
	Engine       *engine.Engine
	LastActivity time.Time
	mu           sync.RWMutex
}

// TenantManager gerencia o lifecycle de múltiplos tenants
// Responsável por criar, recuperar, liberar e limpar tenants inativos
type TenantManager struct {
	mu             sync.RWMutex
	activeTenants  map[string]*Tenant
	evictionPolicy EvictionPolicy
	evictionTicker *time.Ticker
	stopEviction   chan struct{}
	dataDir        string
	llmRouter      *llm.Router
}

// EvictionPolicy define os parâmetros de limpeza automática de tenants
type EvictionPolicy struct {
	IdleTimeout   time.Duration
	MaxTenants    int
	CheckInterval time.Duration
}

// NewTenantManager cria uma nova instância do gerenciador de tenants
func NewTenantManager(policy EvictionPolicy, llmRouter *llm.Router) *TenantManager {
	if policy.IdleTimeout == 0 {
		policy.IdleTimeout = 30 * time.Minute
	}
	if policy.MaxTenants == 0 {
		policy.MaxTenants = 10
	}
	if policy.CheckInterval == 0 {
		policy.CheckInterval = 5 * time.Minute
	}

	appDataDir := os.Getenv("APPDATA")
	if appDataDir == "" {
		appDataDir = os.Getenv("HOME")
	}
	dataDir := filepath.Join(appDataDir, "Vectora", "workspaces")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		// Log do erro mas continua - o erro real aparecerá no GetOrCreateTenant
		fmt.Fprintf(os.Stderr, "Aviso: falha ao criar diretório de workspaces: %v\n", err)
	}

	return &TenantManager{
		activeTenants:  make(map[string]*Tenant),
		evictionPolicy: policy,
		stopEviction:   make(chan struct{}),
		dataDir:        dataDir,
		llmRouter:      llmRouter,
	}
}

// GetOrCreateTenant obtém um tenant existente ou cria um novo
func (tm *TenantManager) GetOrCreateTenant(wsID, wsRoot, projName string) (*Tenant, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if t, exists := tm.activeTenants[wsID]; exists {
		t.mu.Lock()
		t.LastActivity = time.Now()
		t.mu.Unlock()
		return t, nil
	}

	if len(tm.activeTenants) >= tm.evictionPolicy.MaxTenants {
		return nil, fmt.Errorf("limite de tenants ativo atingido: %d", tm.evictionPolicy.MaxTenants)
	}

	tenant := &Tenant{
		ID:           wsID,
		Root:         wsRoot,
		ProjectName:  projName,
		LastActivity: time.Now(),
	}

	workspaceDir := filepath.Join(tm.dataDir, wsID)
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório do workspace: %w", err)
	}

	vecStorePath := filepath.Join(workspaceDir, "chromadb")
	vecStore, err := db.NewVectorStoreAtPath(vecStorePath)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar vector store: %w", err)
	}
	tenant.VectorStore = vecStore

	kvStorePath := filepath.Join(workspaceDir, "chat_history.db")
	kvStore, err := db.NewKVStoreAtPath(kvStorePath)
	if err != nil {
		vecStore.Close()
		return nil, fmt.Errorf("falha ao criar kv store: %w", err)
	}
	tenant.KVStore = kvStore

	// Inicializar Guardian para este tenant
	tenant.Guardian = policies.NewGuardian(wsRoot)

	// Inicializar Engine com o registro de ferramentas local
	toolRegistry := tools.NewRegistry(wsRoot, tenant.Guardian, tenant.KVStore)
	tenant.Engine = engine.NewEngine(
		vecStore,
		kvStore,
		tm.llmRouter,
		toolRegistry,
		tenant.Guardian,
		nil,
	)

	tm.activeTenants[wsID] = tenant
	return tenant, nil
}

// ReleaseTenant descarrega um tenant da memória
func (tm *TenantManager) ReleaseTenant(wsID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if t, exists := tm.activeTenants[wsID]; exists {
		t.mu.Lock()
		defer t.mu.Unlock()

		if t.VectorStore != nil {
			if err := t.VectorStore.Close(); err != nil {
				return fmt.Errorf("erro ao fechar vector store: %w", err)
			}
		}

		if t.KVStore != nil {
			if err := t.KVStore.Close(); err != nil {
				return fmt.Errorf("erro ao fechar kv store: %w", err)
			}
		}

		delete(tm.activeTenants, wsID)
	}

	return nil
}

// GetTenant obtém um tenant ativo por ID
func (tm *TenantManager) GetTenant(wsID string) (*Tenant, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if t, exists := tm.activeTenants[wsID]; exists {
		return t, nil
	}

	return nil, fmt.Errorf("tenant não encontrado: %s", wsID)
}

// GetActiveTenantCount retorna o número de tenants ativos
func (tm *TenantManager) GetActiveTenantCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.activeTenants)
}

// StartEvictionRoutine inicia a rotina de background que limpa tenants inativos
func (tm *TenantManager) StartEvictionRoutine() {
	tm.evictionTicker = time.NewTicker(tm.evictionPolicy.CheckInterval)

	go func() {
		for {
			select {
			case <-tm.evictionTicker.C:
				tm.evictIdleTenants()
			case <-tm.stopEviction:
				tm.evictionTicker.Stop()
				return
			}
		}
	}()
}

// evictIdleTenants verifica e remove tenants que ultrapassaram IdleTimeout
func (tm *TenantManager) evictIdleTenants() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()
	var toEvict []string

	for wsID, tenant := range tm.activeTenants {
		tenant.mu.RLock()
		if now.Sub(tenant.LastActivity) > tm.evictionPolicy.IdleTimeout {
			toEvict = append(toEvict, wsID)
		}
		tenant.mu.RUnlock()
	}

	for _, wsID := range toEvict {
		tenant := tm.activeTenants[wsID]
		tenant.mu.Lock()
		if tenant.VectorStore != nil {
			tenant.VectorStore.Close()
		}
		if tenant.KVStore != nil {
			tenant.KVStore.Close()
		}
		tenant.mu.Unlock()

		delete(tm.activeTenants, wsID)
	}
}

// StopEvictionRoutine para a rotina de background de eviction
func (tm *TenantManager) StopEvictionRoutine() {
	close(tm.stopEviction)
}

// GetTenantDataDir retorna o caminho da pasta de dados para um tenant
func (tm *TenantManager) GetTenantDataDir(wsID string) string {
	return filepath.Join(tm.dataDir, wsID)
}

// ListActiveTenants retorna uma lista de IDs de tenants ativos
func (tm *TenantManager) ListActiveTenants() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var tenants []string
	for wsID := range tm.activeTenants {
		tenants = append(tenants, wsID)
	}
	return tenants
}
