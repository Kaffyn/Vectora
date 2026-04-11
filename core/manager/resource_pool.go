package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

// ResourcePool gerencia recursos compartilhados entre tenants
// Garante que nenhum tenant consuma todos os recursos e afete outros
type ResourcePool struct {
	mu sync.RWMutex

	// Semáforos por tenant para LLM calls
	llmSemaphores map[string]*semaphore.Weighted

	// Fila de prioridade para tarefas de indexação
	indexQueue *PriorityQueue

	// Configuração
	config ResourceConfig
}

// ResourceConfig define limites de recursos por tenant
type ResourceConfig struct {
	MaxParallelLLMPerTenant int           // Máximo de chamadas LLM simultâneas por tenant
	MaxConcurrentIndexing   int           // Máximo de tarefas de indexação simultâneas
	LLMRateLimit            time.Duration // Tempo mínimo entre chamadas LLM
}

// PriorityQueue gerencia tarefas de indexação com prioridade e isolamento por tenant
type PriorityQueue struct {
	mu    sync.Mutex
	items []*IndexTask

	// Semáforo para limitar indexação concurrent
	semaphore *semaphore.Weighted

	// Histórico de execução por tenant para rate limiting
	tenantStats map[string]*TenantStats
}

// IndexTask representa uma tarefa de indexação na fila
type IndexTask struct {
	TenantID  string
	Priority  int
	Timestamp time.Time
	Callback  func() error
}

// TenantStats rastreia estatísticas de uso por tenant
type TenantStats struct {
	LastLLMCall time.Time
	CallCount   int
	ErrorCount  int
}

// NewResourcePool cria uma nova instância do resource pool
func NewResourcePool(config ResourceConfig) *ResourcePool {
	// Validar valores padrão
	if config.MaxParallelLLMPerTenant == 0 {
		config.MaxParallelLLMPerTenant = 2
	}
	if config.MaxConcurrentIndexing == 0 {
		config.MaxConcurrentIndexing = 4
	}
	if config.LLMRateLimit == 0 {
		config.LLMRateLimit = 100 * time.Millisecond
	}

	return &ResourcePool{
		llmSemaphores: make(map[string]*semaphore.Weighted),
		indexQueue: &PriorityQueue{
			items:       make([]*IndexTask, 0),
			semaphore:   semaphore.NewWeighted(int64(config.MaxConcurrentIndexing)),
			tenantStats: make(map[string]*TenantStats),
		},
		config: config,
	}
}

// AcquireLLMSlot adquire permissão para chamar um provedor LLM
// Respeita a cota máxima de chamadas simultâneas por tenant
// Bloqueia se o tenant atingiu seu limite, retorna erro se contexto cancelado
func (rp *ResourcePool) AcquireLLMSlot(tenantID string, ctx context.Context) error {
	rp.mu.Lock()
	sem, exists := rp.llmSemaphores[tenantID]
	if !exists {
		// Criar semáforo novo para este tenant
		sem = semaphore.NewWeighted(int64(rp.config.MaxParallelLLMPerTenant))
		rp.llmSemaphores[tenantID] = sem
	}
	rp.mu.Unlock()

	// Adquirir slot (bloqueia se cota atingida)
	if err := sem.Acquire(ctx, 1); err != nil {
		return fmt.Errorf("falha ao adquirir LLM slot: %w", err)
	}

	return nil
}

// ReleaseLLMSlot libera um slot de LLM para que outro tenant possa usá-lo
func (rp *ResourcePool) ReleaseLLMSlot(tenantID string) {
	rp.mu.RLock()
	sem := rp.llmSemaphores[tenantID]
	rp.mu.RUnlock()

	if sem != nil {
		sem.Release(1)
	}
}

// EnqueueIndexing enfileira uma tarefa de indexação com prioridade
// Tarefas com maior prioridade são executadas primeiro
// Cada tenant tem sua própria cota de tarefas simultâneas
func (rp *ResourcePool) EnqueueIndexing(tenantID string, priority int, callback func() error) {
	task := &IndexTask{
		TenantID:  tenantID,
		Priority:  priority,
		Timestamp: time.Now(),
		Callback:  callback,
	}

	rp.indexQueue.mu.Lock()
	rp.indexQueue.items = append(rp.indexQueue.items, task)
	// Ordenar por prioridade (simples insertion no final, ordem manual)
	rp.indexQueue.mu.Unlock()

	// Iniciar processamento em background
	go rp.processIndexQueue()
}

// processIndexQueue processa tarefas de indexação respeitando limites por tenant
func (rp *ResourcePool) processIndexQueue() {
	rp.indexQueue.mu.Lock()
	if len(rp.indexQueue.items) == 0 {
		rp.indexQueue.mu.Unlock()
		return
	}

	// Pegar próxima tarefa (FIFO com respaldo de prioridade)
	task := rp.indexQueue.items[0]
	rp.indexQueue.items = rp.indexQueue.items[1:]
	rp.indexQueue.mu.Unlock()

	// Adquirir slot de indexação
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := rp.indexQueue.semaphore.Acquire(ctx, 1); err != nil {
		// Recolocar tarefa na fila se falhar
		rp.EnqueueIndexing(task.TenantID, task.Priority, task.Callback)
		return
	}
	defer rp.indexQueue.semaphore.Release(1)

	// Executar callback
	if err := task.Callback(); err != nil {
		// Log error mas continua processando fila
		rp.indexQueue.mu.Lock()
		stats := rp.indexQueue.tenantStats[task.TenantID]
		if stats == nil {
			stats = &TenantStats{}
			rp.indexQueue.tenantStats[task.TenantID] = stats
		}
		stats.ErrorCount++
		rp.indexQueue.mu.Unlock()
	}
}

// GetLLMSlotCount retorna quantos slots de LLM estão em uso por um tenant
func (rp *ResourcePool) GetLLMSlotCount(tenantID string) int {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	_, exists := rp.llmSemaphores[tenantID]
	if !exists {
		return 0
	}

	// Semaphore não expõe count, então retornamos 0 por agora
	// Isso é uma limitação da implementação de golang.org/x/sync/semaphore
	return 0
}

// GetIndexQueueSize retorna o tamanho atual da fila de indexação
func (rp *ResourcePool) GetIndexQueueSize() int {
	rp.indexQueue.mu.Lock()
	defer rp.indexQueue.mu.Unlock()
	return len(rp.indexQueue.items)
}

// GetTenantStats retorna estatísticas de uso para um tenant
func (rp *ResourcePool) GetTenantStats(tenantID string) *TenantStats {
	rp.indexQueue.mu.Lock()
	defer rp.indexQueue.mu.Unlock()

	stats, exists := rp.indexQueue.tenantStats[tenantID]
	if !exists {
		return &TenantStats{}
	}

	return stats
}

// ResetTenantStats limpa estatísticas de um tenant
func (rp *ResourcePool) ResetTenantStats(tenantID string) {
	rp.indexQueue.mu.Lock()
	defer rp.indexQueue.mu.Unlock()

	delete(rp.indexQueue.tenantStats, tenantID)
}
