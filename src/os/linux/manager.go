package linux

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/domain"
)

// LinuxManager implementa a interface domain.OSManager para o Linux.
type LinuxManager struct {
	paths config.VectoraPaths
	cmds  map[string]*exec.Cmd // Comandos para os processos do llama.cpp
}

// NewLinuxManager cria uma nova instância do LinuxManager.
func NewLinuxManager(paths config.VectoraPaths) *LinuxManager {
	return &LinuxManager{
		paths: paths,
		cmds:  make(map[string]*exec.Cmd),
	}
}

// StartLLAMAServer inicia o servidor llama.cpp para o Linux (Otimizado para Vulkan).
func (m *LinuxManager) StartLLAMAServer(ctx context.Context, id string, modelPath string, port int, enableGPU bool) error {
	binaryPath := m.GetLlamaServerBinaryPath()

	args := []string{
		"--model", modelPath,
		"--port", fmt.Sprintf("%d", port),
		"--ctx-size", "2048",
		"--threads", "4",
		"--embedding",
	}

	if enableGPU {
		// No Linux, --n-gpu-layers ativa aceleração Vulkan ou CUDA se os drivers estiverem no sistema.
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
func (m *LinuxManager) StopLLAMAServer(id string) error {
	if cmd, ok := m.cmds[id]; ok && cmd.Process != nil {
		err := cmd.Process.Kill()
		delete(m.cmds, id)
		return err
	}
	return nil
}

// StopAllLLAMAServers encerra todos os servidores ativos.
func (m *LinuxManager) StopAllLLAMAServers() error {
	for id := range m.cmds {
		m.StopLLAMAServer(id)
	}
	return nil
}

// IsLLAMAServerRunning verifica se o servidor llama.cpp está ativo no Linux.
func (m *LinuxManager) IsLLAMAServerRunning(id string) bool {
	if cmd, ok := m.cmds[id]; ok {
		return cmd.ProcessState == nil
	}
	return false
}

// GetLlamaServerBinaryPath retorna o caminho para o binário llama-server no Linux.
func (m *LinuxManager) GetLlamaServerBinaryPath() string {
	return m.paths.GetBinaryPath("llama-server")
}

// RunCommand executa um comando de sistema no Linux de forma síncrona.
func (m *LinuxManager) RunCommand(ctx context.Context, cmdName string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, cmdName, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// Garante que LinuxManager implementa a interface domain.OSManager
var _ domain.OSManager = (*LinuxManager)(nil)
