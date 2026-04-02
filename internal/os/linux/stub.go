//go:build !linux

package linux

import "errors"

type LinuxManager struct{}

func NewManager() *LinuxManager { return nil }
func (m *LinuxManager) GetAppDataDir() (string, error) { return "", errors.New("não aplicável na infraestrutura atual") }
func (m *LinuxManager) StartLlamaEngine(string, int) error { return errors.New("inválido") }
func (m *LinuxManager) StopLlamaEngine() error { return errors.New("inválido") }
func (m *LinuxManager) GetEngineState() string { return "ERROR" }
