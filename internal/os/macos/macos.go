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

func (m *MacosManager) GetInstallDir() (string, error) {
	return "/Applications/Vectora.app", nil
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

// ==== MAC OS EXTENSIONS ====

func (m *MacosManager) IsInstalled() string {
	p, _ := m.GetAppDataDir()
	if _, err := os.Stat(filepath.Join(p, "vectora")); err == nil {
		return p
	}
	return ""
}

// Only instantiate the structural directory
func (m *MacosManager) RegisterApp(installDir string) {
	os.MkdirAll(installDir, 0755)
}

func (m *MacosManager) UnregisterApp(installDir string) {
	// Manual cleanup on Unix
}

// For macOS, Win32 Mutex does not exist. We use an invisible TCP lock (Standard UNIX approach).
func (m *MacosManager) EnforceSingleInstance() error {
	l, err := net.Listen("tcp", "127.0.0.1:41785")
	if err != nil {
		return errors.New("instance_already_running")
	}
	m.portListener = l
	return nil
}
