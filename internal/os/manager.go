package os

import (
	"fmt"
	"runtime"

	lin "github.com/Kaffyn/vectora/internal/os/linux"
	mac "github.com/Kaffyn/vectora/internal/os/macos"
	win "github.com/Kaffyn/vectora/internal/os/windows"
)

type EngineState string

const (
	EngineStopped  EngineState = "STOPPED"
	EngineStarting EngineState = "STARTING"
	EngineRunning  EngineState = "RUNNING"
	EngineError    EngineState = "ERROR"
)

// OSManager encapsula interações Nativas do Kernel Host para que Daemon e Instalador sejam universais.
type OSManager interface {
	GetAppDataDir() (string, error)
	StartLlamaEngine(modelPath string, port int) error
	StopLlamaEngine() error
	GetEngineState() string

	// Sistema de Registro do Installer Level
	IsInstalled() string
	RegisterApp(installDir string)
	UnregisterApp(installDir string)
	
	// Prevenção de Concorrência
	EnforceSingleInstance() error
}

func NewManager() (OSManager, error) {
	switch runtime.GOOS {
	case "windows":
		return win.NewManager(), nil
	case "darwin":
		return mac.NewManager(), nil
	case "linux":
		return lin.NewManager(), nil
	default:
		return nil, fmt.Errorf("unsupported operating system environment: %s", runtime.GOOS)
	}
}
