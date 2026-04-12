//go:build darwin

package os

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

func NewManager() (OSManager, error) {
	return &MacosManager{state: string(EngineStopped)}, nil
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

func (m *MacosManager) IsRunningAsAdmin() bool {
	return os.Geteuid() == 0
}

func (m *MacosManager) StartLlamaEngine(modelPath string, port int) error {
	m.state = string(EngineStarting)
	baseDir, err := m.GetAppDataDir()
	if err != nil {
		m.state = string(EngineError)
		return err
	}

	binaryPath := filepath.Join(baseDir, "llama-server")
	m.cmd = exec.Command(binaryPath, "-m", modelPath, "--port", fmt.Sprintf("%d", port), "-ngl", "99")

	err = m.cmd.Start()
	if err != nil {
		m.state = string(EngineError)
		return err
	}

	m.state = string(EngineRunning)
	go func() {
		m.cmd.Wait()
		m.state = string(EngineStopped)
	}()

	return nil
}

func (m *MacosManager) StopLlamaEngine() error {
	if m.cmd != nil && m.cmd.Process != nil {
		err := m.cmd.Process.Kill()
		m.state = string(EngineStopped)
		m.cmd = nil
		return err
	}
	return nil
}

func (m *MacosManager) GetEngineState() string {
	return m.state
}

func (m *MacosManager) IsInstalled() string {
	p, _ := m.GetAppDataDir()
	if _, err := os.Stat(filepath.Join(p, "vectora")); err == nil {
		return p
	}
	return ""
}

func (m *MacosManager) RegisterApp(installDir string) {
	os.MkdirAll(installDir, 0755)
}

func (m *MacosManager) UnregisterApp(installDir string) {
	_ = installDir
}

func (m *MacosManager) GetSystemLanguage() string {
	if lang := os.Getenv("APPLE_LANGUAGE"); lang != "" {
		return lang[:2]
	}
	if lang := os.Getenv("LANG"); lang != "" {
		return lang[:2]
	}
	return "en"
}
