//go:build windows

package os

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

type WindowsManager struct {
	cmd   *exec.Cmd
	state string
}

func NewManager() (OSManager, error) {
	return &WindowsManager{state: string(EngineStopped)}, nil
}

func (m *WindowsManager) GetAppDataDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".Vectora"), nil
	}
	return filepath.Join(appData, "Vectora"), nil
}

func (m *WindowsManager) GetInstallDir() (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			progFiles := os.Getenv("ProgramFiles")
			if progFiles == "" {
				progFiles = `C:\Program Files`
			}
			return filepath.Join(progFiles, "Vectora"), nil
		}
		return filepath.Join(home, "AppData", "Local", "Vectora"), nil
	}
	return filepath.Join(localAppData, "Programs", "Vectora"), nil
}

func (m *WindowsManager) IsRunningAsAdmin() bool {
	f, err := os.Open(`\\.\PHYSICALDRIVE0`)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func (m *WindowsManager) StartLlamaEngine(modelPath string, port int) error {
	m.state = string(EngineStarting)
	installDir, err := m.GetInstallDir()
	if err != nil {
		m.state = string(EngineError)
		return err
	}

	binaryPath := filepath.Join(installDir, "llama-server.exe")
	m.cmd = exec.Command(binaryPath, "-m", modelPath, "--port", fmt.Sprintf("%d", port), "-ngl", "99")
	m.cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW

	err = m.cmd.Start()
	if err != nil {
		m.state = string(EngineError)
		return err
	}

	m.state = string(EngineRunning)
	go func() {
		_ = m.cmd.Wait()
		m.state = string(EngineStopped)
	}()
	return nil
}

func (m *WindowsManager) StopLlamaEngine() error {
	if m.cmd != nil && m.cmd.Process != nil {
		err := m.cmd.Process.Kill()
		m.state = string(EngineStopped)
		m.cmd = nil
		return err
	}
	return nil
}

func (m *WindowsManager) GetEngineState() string {
	return m.state
}

func (m *WindowsManager) IsInstalled() string {
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
	if err == nil {
		defer key.Close()
		val, _, err := key.GetStringValue("InstallLocation")
		if err == nil && val != "" {
			return val
		}
	}
	return ""
}

func (m *WindowsManager) RegisterApp(installDir string) {
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`
	key, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.ALL_ACCESS)
	if err == nil {
		defer key.Close()
		_ = key.SetStringValue("DisplayName", "Vectora")
		_ = key.SetStringValue("DisplayVersion", "1.0.0")
		_ = key.SetStringValue("Publisher", "Kaffyn")
		_ = key.SetStringValue("DisplayIcon", filepath.Join(installDir, "vectora.exe"))
		_ = key.SetStringValue("InstallLocation", installDir)
	}
}

func (m *WindowsManager) UnregisterApp(installDir string) {
	_ = registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`)
	_ = installDir
}
