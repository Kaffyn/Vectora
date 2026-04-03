//go:build !darwin

package macos

import (
	"errors"
	"net"
	"os/exec"
)

type MacosManager struct {
	cmd   *exec.Cmd
	state string
	portListener net.Listener
}

func NewManager() *MacosManager {
	return &MacosManager{state: "STOPPED"}
}

func (m *MacosManager) GetAppDataDir() (string, error) { return "", nil }
func (m *MacosManager) GetInstallDir() (string, error) { return "", nil }
func (m *MacosManager) IsRunningAsAdmin() bool { return false }
func (m *MacosManager) StartLlamaEngine(modelPath string, port int) error { return nil }
func (m *MacosManager) StopLlamaEngine() error { return nil }
func (m *MacosManager) GetEngineState() string { return "" }
func (m *MacosManager) IsInstalled() string { return "" }
func (m *MacosManager) RegisterApp(installDir string) {}
func (m *MacosManager) UnregisterApp(installDir string) {}
func (m *MacosManager) EnforceSingleInstance() error { return nil }
