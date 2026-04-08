//go:build !linux

package linux

import "errors"

type LinuxManager struct{}

func NewManager() *LinuxManager {
	return &LinuxManager{}
}

func (m *LinuxManager) GetAppDataDir() (string, error) {
	return "", errors.New("unsupported OS")
}
func (m *LinuxManager) GetInstallDir() (string, error) {
	return "", errors.New("unsupported OS")
}
func (m *LinuxManager) IsRunningAsAdmin() bool {
	return false
}
func (m *LinuxManager) StartLlamaEngine(string, int) error {
	return errors.New("unsupported OS")
}
func (m *LinuxManager) StopLlamaEngine() error {
	return errors.New("unsupported OS")
}
func (m *LinuxManager) GetEngineState() string {
	return "ERROR"
}
func (m *LinuxManager) IsInstalled() string {
	return ""
}
func (m *LinuxManager) RegisterApp(string)   {}
func (m *LinuxManager) UnregisterApp(string) {}
func (m *LinuxManager) EnforceSingleInstance() error {
	return errors.New("unsupported OS")
}
