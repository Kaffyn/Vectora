//go:build !darwin

package macos

import "errors"

type MacosManager struct{}

func NewManager() *MacosManager { return nil }
func (m *MacosManager) GetAppDataDir() (string, error) { return "", errors.New("não aplicável na infraestrutura atual") }
func (m *MacosManager) StartLlamaEngine(string, int) error { return errors.New("inválido") }
func (m *MacosManager) StopLlamaEngine() error { return errors.New("inválido") }
func (m *MacosManager) GetEngineState() string { return "ERROR" }
