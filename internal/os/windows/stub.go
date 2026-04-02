//go:build !windows

package windows

import "errors"

type WindowsManager struct{}

func NewManager() *WindowsManager { return nil }
func (m *WindowsManager) GetAppDataDir() (string, error) { return "", errors.New("não aplicável na infraestrutura atual") }
func (m *WindowsManager) GetInstallDir() (string, error) { return "", errors.New("não aplicável") }
func (m *WindowsManager) StartLlamaEngine(string, int) error { return errors.New("inválido") }
func (m *WindowsManager) StopLlamaEngine() error { return errors.New("inválido") }
func (m *WindowsManager) GetEngineState() string { return "ERROR" }
func (m *WindowsManager) IsInstalled() string { return "" }
func (m *WindowsManager) RegisterApp(string) {}
func (m *WindowsManager) UnregisterApp(string) {}
func (m *WindowsManager) EnforceSingleInstance() error { return errors.New("inválido") }
