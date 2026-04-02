//go:build linux

package linux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type LinuxManager struct {
	cmd   *exec.Cmd
	state string
}

func NewManager() *LinuxManager {
	return &LinuxManager{state: "STOPPED"}
}

func (m *LinuxManager) GetAppDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Freedesktop.org Padrões FHS estritos para software invisível de usuário-espaço.
	return filepath.Join(home, ".local", "share", "Vectora"), nil
}

func (m *LinuxManager) StartLlamaEngine(modelPath string, port int) error {
	m.state = "STARTING"
	baseDir, err := m.GetAppDataDir()
	if err != nil {
		m.state = "ERROR"
		return err
	}

	binaryPath := filepath.Join(baseDir, "llama-server")

	// Puxa dependências nativas Vulkan / System Libraries em Linux Host para LLM off-load pesado.
	m.cmd = exec.Command(binaryPath, "-m", modelPath, "--port", fmt.Sprintf("%d", port), "-ngl", "99")

	err = m.cmd.Start()
	if err != nil {
		m.state = "ERROR"
		return err
	}

	m.state = "RUNNING"
	go func() {
		m.cmd.Wait()
		m.state = "STOPPED"
	}()

	return nil
}

func (m *LinuxManager) StopLlamaEngine() error {
	if m.cmd != nil && m.cmd.Process != nil {
		err := m.cmd.Process.Kill()
		m.state = "STOPPED"
		m.cmd = nil
		return err
	}
	return nil
}

func (m *LinuxManager) GetEngineState() string {
	return m.state
}
