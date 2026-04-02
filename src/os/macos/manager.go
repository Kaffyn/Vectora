package macos

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/domain"
)

// MacOSManager implementa a interface domain.OSManager para o macOS.
type MacOSManager struct {
	paths config.VectoraPaths
	cmds  map[string]*exec.Cmd // Comandos para os processos do llama.cpp
}

// NewMacOSManager cria uma nova instância do MacOSManager.
func NewMacOSManager(paths config.VectoraPaths) *MacOSManager {
	return &MacOSManager{
		paths: paths,
		cmds:  make(map[string]*exec.Cmd),
	}
}

// StartLLAMAServer inicia o servidor llama.cpp para o macOS (Otimizado para Metal/Apple Silicon).
func (m *MacOSManager) StartLLAMAServer(ctx context.Context, id string, modelPath string, port int, enableGPU bool) error {
	binaryPath := m.GetLlamaServerBinaryPath()

	args := []string{
		"--model", modelPath,
		"--port", fmt.Sprintf("%d", port),
		"--ctx-size", "2048",
		"--threads", "4",
		"--embedding",
	}

	if enableGPU {
		// No macOS moderno (ARM), --n-gpu-layers ativa aceleração Metal automaticamente se compilado.
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
func (m *MacOSManager) StopLLAMAServer(id string) error {
	if cmd, ok := m.cmds[id]; ok && cmd.Process != nil {
		err := cmd.Process.Kill()
		delete(m.cmds, id)
		return err
	}
	return nil
}

// StopAllLLAMAServers encerra todos os servidores ativos.
func (m *MacOSManager) StopAllLLAMAServers() error {
	for id := range m.cmds {
		m.StopLLAMAServer(id)
	}
	return nil
}

// IsLLAMAServerRunning verifica se o servidor llama.cpp está ativo no macOS.
func (m *MacOSManager) IsLLAMAServerRunning(id string) bool {
	if cmd, ok := m.cmds[id]; ok {
		return cmd.ProcessState == nil
	}
	return false
}

// GetLlamaServerBinaryPath retorna o caminho para o binário llama-server no macOS.
func (m *MacOSManager) GetLlamaServerBinaryPath() string {
	return m.paths.GetBinaryPath("llama-server")
}

// RunCommand executa um comando de sistema no macOS de forma síncrona.
func (m *MacOSManager) RunCommand(ctx context.Context, cmdName string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, cmdName, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// Garante que MacOSManager implementa a interface domain.OSManager
var _ domain.OSManager = (*MacOSManager)(nil)
