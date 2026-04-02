package windows

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/domain"
)

// WindowsManager implementa a interface domain.OSManager para o Windows.
type WindowsManager struct {
	paths config.VectoraPaths
	cmds  map[string]*exec.Cmd // Comandos para os processos do llama.cpp
}

// NewWindowsManager cria uma nova instância do WindowsManager.
func NewWindowsManager(paths config.VectoraPaths) *WindowsManager {
	return &WindowsManager{
		paths: paths,
		cmds:  make(map[string]*exec.Cmd),
	}
}

// StartLLAMAServer inicia o servidor llama.cpp para o Windows (Otimizado para Vulkan).
func (m *WindowsManager) StartLLAMAServer(ctx context.Context, id string, modelPath string, port int, enableGPU bool) error {
	binaryPath := m.GetLlamaServerBinaryPath()

	args := []string{
		"--model", modelPath,
		"--port", fmt.Sprintf("%d", port),
		"--ctx-size", "2048",
		"--threads", "4",
		"--embedding",
	}

	if enableGPU {
		// No Windows, usamos Vulkan como o driver unificador CPU/GPU.
		args = append(args, "--n-gpu-layers", "100")
		args = append(args, "--flash-attn")
	}

	cmd := exec.CommandContext(ctx, binaryPath, args...)
	m.cmds[id] = cmd

	// Inicia o processo em segundo plano
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar o binário llama-server [%s] em %s: %w", id, binaryPath, err)
	}

	// Espera por uma verificação de saúde conceitual (o ideal seria esperar pelo endpoint HTTP)
	time.Sleep(2 * time.Second)

	return nil
}

// StopLLAMAServer encerra um servidor específico.
func (m *WindowsManager) StopLLAMAServer(id string) error {
	if cmd, ok := m.cmds[id]; ok && cmd.Process != nil {
		err := cmd.Process.Kill()
		delete(m.cmds, id)
		return err
	}
	return nil
}

// StopAllLLAMAServers encerra todos os servidores ativos.
func (m *WindowsManager) StopAllLLAMAServers() error {
	for id := range m.cmds {
		m.StopLLAMAServer(id)
	}
	return nil
}

// IsLLAMAServerRunning verifica se o servidor llama.cpp está ativo no Windows.
func (m *WindowsManager) IsLLAMAServerRunning(id string) bool {
	if cmd, ok := m.cmds[id]; ok {
		return cmd.ProcessState == nil
	}
	return false
}

// GetLlamaServerBinaryPath retorna o caminho para o binário llama-server no Windows.
func (m *WindowsManager) GetLlamaServerBinaryPath() string {
	return m.paths.GetBinaryPath("llama-server")
}

// RunCommand executa um comando de sistema no Windows de forma síncrona.
func (m *WindowsManager) RunCommand(ctx context.Context, cmdName string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, cmdName, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// Garante que WindowsManager implementa a interface domain.OSManager
var _ domain.OSManager = (*WindowsManager)(nil)
