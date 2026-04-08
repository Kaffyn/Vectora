//go:build !darwin

package macos

import "errors"

type MacosManager struct{}

func NewManager() *MacosManager {
	return &MacosManager{}
}

func (m *MacosManager) GetAppDataDir() (string, error) {
	return "", errors.New("unsupported OS")
}
func (m *MacosManager) GetInstallDir() (string, error) {
	return "", errors.New("unsupported OS")
}
func (m *MacosManager) IsRunningAsAdmin() bool {
	return false
}
func (m *MacosManager) StartLlamaEngine(string, int) error {
	return errors.New("unsupported OS")
}
func (m *MacosManager) StopLlamaEngine() error {
	return errors.New("unsupported OS")
}
func (m *MacosManager) GetEngineState() string {
	return "ERROR"
}
func (m *MacosManager) IsInstalled() string {
	return ""
}
func (m *MacosManager) RegisterApp(string)   {}
func (m *MacosManager) UnregisterApp(string) {}
func (m *MacosManager) EnforceSingleInstance() error {
	return errors.New("unsupported OS")
}
