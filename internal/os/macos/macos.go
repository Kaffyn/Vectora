//go:build darwin

package macos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type MacosManager struct {
	cmd   *exec.Cmd
	state string
}

func NewManager() *MacosManager {
	return &MacosManager{state: "STOPPED"}
}

func (m *MacosManager) GetAppDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Padrão estrutural do Apple File System (APFS) para suporte nativo a aplicações locais Sandboxed.
	return filepath.Join(home, "Library", "Application Support", "Vectora"), nil
}

func (m *MacosManager) StartLlamaEngine(modelPath string, port int) error {
	m.state = "STARTING"
	baseDir, err := m.GetAppDataDir()
	if err != nil {
		m.state = "ERROR"
		return err
	}

	binaryPath := filepath.Join(baseDir, "llama-server")

	// Roda o Binário Darwin com injeção paralela baseada em aceleradores de hardware (Toda Mac Apple Silicon M1-M4 fluí via Metal nativamente).
	m.cmd = exec.Command(binaryPath, "-m", modelPath, "--port", fmt.Sprintf("%d", port), "-ngl", "99")

	err = m.cmd.Start()
	if err != nil {
		m.state = "ERROR"
		return err
	}

	m.state = "RUNNING"
	
	// Vigília Sub-Thread Mestre Catcher
	go func() {
		m.cmd.Wait()
		m.state = "STOPPED"
	}()

	return nil
}

func (m *MacosManager) StopLlamaEngine() error {
	if m.cmd != nil && m.cmd.Process != nil {
		err := m.cmd.Process.Kill()
		m.state = "STOPPED"
		m.cmd = nil
		return err
	}
	return nil
}

func (m *MacosManager) GetEngineState() string {
	return m.state
}
