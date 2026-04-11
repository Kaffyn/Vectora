package manager

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// ResourceConfig define os limites de recursos por tenant.
type ResourceConfig struct {
	MaxParallelLLMPerTenant int // Máximo de requisições LLM simultâneas por projeto (padrão: 2)
	MaxConcurrentIndexing   int // Máximo de threads de indexação global
}

// ResourcePool gerencia a distribuição de recursos entre múltiplos tenants.
type ResourcePool struct {
	mu            sync.RWMutex
	llmSemaphores map[string]*semaphore.Weighted
	config        ResourceConfig
}

func NewResourcePool(config ResourceConfig) *ResourcePool {
	return &ResourcePool{
		llmSemaphores: make(map[string]*semaphore.Weighted),
		config:        config,
	}
}

// AcquireLLMSlot solicita permissão para executar uma chamada de LLM para um tenant específico.
func (rp *ResourcePool) AcquireLLMSlot(ctx context.Context, tenantID string) error {
	rp.mu.Lock()
	sem, exists := rp.llmSemaphores[tenantID]
	if !exists {
		sem = semaphore.NewWeighted(int64(rp.config.MaxParallelLLMPerTenant))
		rp.llmSemaphores[tenantID] = sem
	}
	rp.mu.Unlock()

	return sem.Acquire(ctx, 1)
}

// ReleaseLLMSlot libera o slot de execução ocupado por uma chamada de LLM.
func (rp *ResourcePool) ReleaseLLMSlot(tenantID string) {
	rp.mu.RLock()
	sem, exists := rp.llmSemaphores[tenantID]
	rp.mu.RUnlock()

	if exists {
		sem.Release(1)
	}
}

// Nota: A PriorityQueue para indexação será implementada conforme a necessidade
// do engine de indexação em fases posteriores, utilizando este pool como base.
