//go:build windows

package windows

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type WindowsManager struct {
	cmd   *exec.Cmd
	state string
}

func NewManager() *WindowsManager {
	return &WindowsManager{state: "STOPPED"}
}

func (m *WindowsManager) GetAppDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".Vectora"), nil
}

func (m *WindowsManager) GetInstallDir() (string, error) {
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

func (m *WindowsManager) IsRunningAsAdmin() bool {
	f, err := os.Open(`\\.\PHYSICALDRIVE0`)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func (m *WindowsManager) StartLlamaEngine(modelPath string, port int) error {
	m.state = "STARTING"
	baseDir, err := m.GetAppDataDir()
	if err != nil {
		m.state = "ERROR"
		return err
	}

	binaryPath := filepath.Join(baseDir, "llama-server.exe")
	m.cmd = exec.Command(binaryPath, "-m", modelPath, "--port", fmt.Sprintf("%d", port), "-ngl", "99")
	m.cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW

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

func (m *WindowsManager) StopLlamaEngine() error {
	if m.cmd != nil && m.cmd.Process != nil {
		err := m.cmd.Process.Kill()
		m.state = "STOPPED"
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
		key.SetStringValue("DisplayName", "Vectora")
		key.SetStringValue("DisplayVersion", "1.0.0")
		key.SetStringValue("Publisher", "Kaffyn")
		key.SetStringValue("DisplayIcon", filepath.Join(installDir, "vectora.exe"))
		key.SetStringValue("InstallLocation", installDir)
	}
}

func (m *WindowsManager) UnregisterApp(installDir string) {
	registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`)
	_ = installDir
}

func (m *WindowsManager) EnforceSingleInstance() error {
	mutexName, _ := windows.UTF16PtrFromString("Global\\VectoraCoreMutex_v1")
	_, err := windows.CreateMutex(nil, false, mutexName)
	if err == windows.ERROR_ALREADY_EXISTS {
		return errors.New("instance_already_running")
	}
	return nil
}
