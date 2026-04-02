package ai

import (
	"context"
	// Removido fmt, os/exec, time, config pois não são mais necessários diretamente aqui
	// "fmt"
	// "os/exec"
	"path/filepath"
	"sync"

	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/domain" // Importar a interface OSManager
)

// SidecarManager agora usa a interface OSManager para abstrair as operações específicas do OS.
type SidecarManager struct {
	osManager    domain.OSManager
	paths        config.VectoraPaths
	mu           sync.RWMutex
	activeModels map[string]string
}

// NewSidecarManager cria uma nova instância do SidecarManager, recebendo um OSManager.
func NewSidecarManager(osManager domain.OSManager, paths config.VectoraPaths) *SidecarManager {
	return &SidecarManager{
		osManager:    osManager,
		paths:        paths,
		activeModels: make(map[string]string),
	}
}

// StartLLAMA inicia um servidor llama.cpp (text ou embedding) usando o OSManager.
func (m *SidecarManager) StartLLAMA(ctx context.Context, id string, modelName string, port int, enableVulkan bool) error {
	modelPath := filepath.Join(m.paths.Models, modelName, modelName+".gguf")

	err := m.osManager.StartLLAMAServer(ctx, id, modelPath, port, enableVulkan)
	if err == nil {
		m.mu.Lock()
		m.activeModels[id] = modelName
		m.mu.Unlock()
	}
	return err
}

// Stop encerra um servidor específico.
func (m *SidecarManager) Stop(id string) error {
	err := m.osManager.StopLLAMAServer(id)
	m.mu.Lock()
	delete(m.activeModels, id)
	m.mu.Unlock()
	return err
}

// StopAll encerra todos os servidores.
func (m *SidecarManager) StopAll() error {
	err := m.osManager.StopAllLLAMAServers()
	m.mu.Lock()
	m.activeModels = make(map[string]string)
	m.mu.Unlock()
	return err
}

// IsRunning verifica se um servidor específico está ativo.
func (m *SidecarManager) IsRunning(id string) bool {
	return m.osManager.IsLLAMAServerRunning(id)
}

// EnsureRunning garante que um servidor específico está ativo com o modelo solicitado.
func (m *SidecarManager) EnsureRunning(ctx context.Context, id string, modelName string, port int, enableVulkan bool) error {
	m.mu.RLock()
	currentModel, isTracked := m.activeModels[id]
	m.mu.RUnlock()

	isRunning := m.IsRunning(id)

	if isRunning && isTracked && currentModel == modelName {
		return nil // Já está rodando e com o modelo correto
	}

	if isRunning {
		// Derrubar o antigo para Hot Swap
		m.Stop(id)
	}

	return m.StartLLAMA(ctx, id, modelName, port, enableVulkan)
}

// GetActiveModel retorna o nome do modelo atualmente atrelado a um ID.
func (m *SidecarManager) GetActiveModel(id string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeModels[id]
}
