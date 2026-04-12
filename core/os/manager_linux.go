//go:build linux

package os

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

func NewManager() (OSManager, error) {
	return &LinuxManager{state: string(EngineStopped)}, nil
}

func (m *LinuxManager) GetAppDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".Vectora"), nil
}

func (m *LinuxManager) GetInstallDir() (string, error) {
	return "/opt/vectora", nil
}

func (m *LinuxManager) IsRunningAsAdmin() bool {
	return os.Geteuid() == 0
}

func (m *LinuxManager) StartLlamaEngine(modelPath string, port int) error {
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

func (m *LinuxManager) StopLlamaEngine() error {
	if m.cmd != nil && m.cmd.Process != nil {
		err := m.cmd.Process.Kill()
		m.state = string(EngineStopped)
		m.cmd = nil
		return err
	}
	return nil
}

func (m *LinuxManager) GetEngineState() string {
	return m.state
}

func (m *LinuxManager) IsInstalled() string {
	home, _ := os.UserHomeDir()
	desktopFile := filepath.Join(home, ".local", "share", "applications", "vectora.desktop")
	if _, err := os.Stat(desktopFile); err == nil {
		p, _ := m.GetAppDataDir()
		return p
	}
	return ""
}

func (m *LinuxManager) RegisterApp(installDir string) {
	desktopContent := `[Desktop Entry]
Name=Vectora
Exec=` + filepath.Join(installDir, "vectora") + `
Icon=vectora
Type=Application
Categories=Utility;
Terminal=false`

	home, _ := os.UserHomeDir()
	desktopDir := filepath.Join(home, ".local", "share", "applications")
	os.MkdirAll(desktopDir, 0755)
	os.WriteFile(filepath.Join(desktopDir, "vectora.desktop"), []byte(desktopContent), 0644)
}

func (m *LinuxManager) UnregisterApp(installDir string) {
	home, _ := os.UserHomeDir()
	os.Remove(filepath.Join(home, ".local", "share", "applications", "vectora.desktop"))
	_ = installDir
}

func (m *LinuxManager) GetSystemLanguage() string {
	if lang := os.Getenv("LANG"); lang != "" {
		return lang[:2]
	}
	return "en"
}
