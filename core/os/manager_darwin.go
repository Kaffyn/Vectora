//go:build darwin

package os

import (
	mac "github.com/Kaffyn/Vectora/core/os/macos"
)

type EngineState string

const (
	EngineStopped  EngineState = "STOPPED"
	EngineStarting EngineState = "STARTING"
	EngineRunning  EngineState = "RUNNING"
	EngineError    EngineState = "ERROR"
)

type OSManager interface {
	GetAppDataDir() (string, error)
	GetInstallDir() (string, error)
	IsRunningAsAdmin() bool
	StartLlamaEngine(modelPath string, port int) error
	StopLlamaEngine() error
	GetEngineState() string
	IsInstalled() string
	RegisterApp(installDir string)
	UnregisterApp(installDir string)
	EnforceSingleInstance() error
}

func NewManager() (OSManager, error) {
	return mac.NewManager(), nil
}
