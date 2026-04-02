//go:build darwin

package macos

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
)

type MacosManager struct {
	cmd   *exec.Cmd
	state string
	portListener net.Listener
}

func NewManager() *MacosManager {
	return &MacosManager{state: "STOPPED"}
}

func (m *MacosManager) GetAppDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".Vectora"), nil
}

func (m *MacosManager) StartLlamaEngine(modelPath string, port int) error {
	m.state = "STARTING"
	baseDir, err := m.GetAppDataDir()
	if err != nil {
		m.state = "ERROR"
		return err
	}

	binaryPath := filepath.Join(baseDir, "llama-server")
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

// ==== EXTENSÕES MAC OS ====

func (m *MacosManager) IsInstalled() string {
	p, _ := m.GetAppDataDir()
	if _, err := os.Stat(filepath.Join(p, "vectora")); err == nil {
		return p
	}
	return ""
}

func (m *MacosManager) RegisterApp(installDir string) {
	// Apenas instanciar o dir estrutural
	os.MkdirAll(installDir, 0755)
}

func (m *MacosManager) UnregisterApp(installDir string) {
	// Cleanup manual em Unix
}

// Para MacOS, Mutex Win32 inexiste. Faremos lock via TCP invisível (Standard UNIX approach sem Syscalls libc sujas).
func (m *MacosManager) EnforceSingleInstance() error {
	l, err := net.Listen("tcp", "127.0.0.1:41785")
	if err != nil {
		return errors.New("instance_already_running")
	}
	m.portListener = l
	return nil
}
