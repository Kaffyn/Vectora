//go:build linux

package linux

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
)

type LinuxManager struct {
	cmd   *exec.Cmd
	state string
	portListener net.Listener
}

func NewManager() *LinuxManager {
	return &LinuxManager{state: "STOPPED"}
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

// ==== REGISTRY AND O.S FILES EXTENSIONS (LINUX) ====

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
	os.MkdirAll(installDir, 0755)
	
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
}

// Adoption of silent Socket Binding for Single-Instance System-wide Lock
func (m *LinuxManager) EnforceSingleInstance() error {
	l, err := net.Listen("tcp", "127.0.0.1:41785")
	if err != nil {
		return errors.New("instance_already_running")
	}
	m.portListener = l
	return nil
}
