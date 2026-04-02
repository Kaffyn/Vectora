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
	return &WindowsManager{
		state: "STOPPED",
	}
}

func (m *WindowsManager) GetAppDataDir() (string, error) {
	// Para dados do usuário e configurações (como o .env), continuamos usando a pasta do usuário
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".Vectora"), nil
}

func (m *WindowsManager) GetInstallDir() (string, error) {
	// Para os executáveis, o padrão é Program Files no Windows
	progFiles := os.Getenv("ProgramFiles")
	if progFiles == "" {
		progFiles = `C:\Program Files`
	}
	return filepath.Join(progFiles, "Vectora"), nil
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
	m.cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}

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

// ==== EXTENSÕES DE REGISTRO E SINGLETON ====

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
		key.SetStringValue("UninstallString", filepath.Join(installDir, "vectora-uninstaller.exe")+" --uninstall")
		key.SetStringValue("InstallLocation", installDir)
	}

	appData := os.Getenv("APPDATA")
	if appData != "" {
		programsDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Vectora")
		os.MkdirAll(programsDir, 0755)
		script := `
$WShell = New-Object -ComObject WScript.Shell
$Shortcut = $WShell.CreateShortcut("` + filepath.Join(programsDir, "Vectora.lnk") + `")
$Shortcut.TargetPath = "` + filepath.Join(installDir, "vectora.exe") + `"
$Shortcut.WorkingDirectory = "` + installDir + `"
$Shortcut.IconLocation = "` + filepath.Join(installDir, "vectora.exe") + `,0"
$Shortcut.Save()
`
		exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script).Run()
	}
}

func (m *WindowsManager) UnregisterApp(installDir string) {
	registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`)
	appData := os.Getenv("APPDATA")
	if appData != "" {
		os.RemoveAll(filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Vectora"))
	}
}

func (m *WindowsManager) EnforceSingleInstance() error {
	mutexName, _ := windows.UTF16PtrFromString("Global\\VectoraDaemonMutex_v1")
	_, err := windows.CreateMutex(nil, false, mutexName)
	if err == windows.ERROR_ALREADY_EXISTS {
		return errors.New("instance_already_running")
	}
	return nil
}
