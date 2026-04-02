//go:build windows

package windows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
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
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Diretório Nativo exigido e estabelecido pelo OS Windows p/ Programas portáteis de Usuário
	return filepath.Join(home, "AppData", "Local", "Programs", "Vectora"), nil
}

func (m *WindowsManager) StartLlamaEngine(modelPath string, port int) error {
	m.state = "STARTING"
	baseDir, err := m.GetAppDataDir()
	if err != nil {
		m.state = "ERROR"
		return err
	}

	binaryPath := filepath.Join(baseDir, "llama-server.exe")

	// Prepara a infraestrutura de GPU local (Vulkan/CUDA genérico para Win32 compilado).
	m.cmd = exec.Command(binaryPath, "-m", modelPath, "--port", fmt.Sprintf("%d", port), "-ngl", "99")

	// Syscall brutal para Windows: CREATE_NO_WINDOW proibe abrir aquele terminal feio e pop-up preto do cmd.
	m.cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}

	err = m.cmd.Start()
	if err != nil {
		m.state = "ERROR"
		return err
	}

	m.state = "RUNNING"
	
	// Cadeira auxiliar de Vigia de Morte do Processo (Evita Zumbis caso crashe).
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
