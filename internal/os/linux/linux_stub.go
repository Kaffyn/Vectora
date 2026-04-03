//go:build !linux

package linux

import (
	"errors"
	"net"
	"os/exec"
)

type LinuxManager struct {
	cmd   *exec.Cmd
	state string
	portListener net.Listener
}

func NewManager() *LinuxManager {
	return &LinuxManager{state: "STOPPED"}
}

func (m *LinuxManager) GetAppDataDir() (string, error) { return "", nil }
func (m *LinuxManager) GetInstallDir() (string, error) { return "", nil }
func (m *LinuxManager) IsRunningAsAdmin() bool { return false }
func (m *LinuxManager) StartLlamaEngine(modelPath string, port int) error { return nil }
func (m *LinuxManager) StopLlamaEngine() error { return nil }
func (m *LinuxManager) GetEngineState() string { return "" }
func (m *LinuxManager) IsInstalled() string { return "" }
func (m *LinuxManager) RegisterApp(installDir string) {}
func (m *LinuxManager) UnregisterApp(installDir string) {}
func (m *LinuxManager) EnforceSingleInstance() error { return nil }
