//go:build !windows

package windows

import "errors"

type WindowsManager struct{}

func NewManager() *WindowsManager {
	return &WindowsManager{}
}

func (m *WindowsManager) GetAppDataDir() (string, error)     { return "", errors.New("unsupported OS") }
func (m *WindowsManager) GetInstallDir() (string, error)     { return "", errors.New("unsupported OS") }
func (m *WindowsManager) IsRunningAsAdmin() bool             { return false }
func (m *WindowsManager) StartLlamaEngine(string, int) error { return errors.New("unsupported OS") }
func (m *WindowsManager) StopLlamaEngine() error             { return errors.New("unsupported OS") }
func (m *WindowsManager) GetEngineState() string             { return "ERROR" }
func (m *WindowsManager) IsInstalled() string                { return "" }
func (m *WindowsManager) RegisterApp(string)                 {}
func (m *WindowsManager) UnregisterApp(string)               {}
func (m *WindowsManager) EnforceSingleInstance() error       { return errors.New("unsupported OS") }
