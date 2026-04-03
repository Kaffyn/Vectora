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

// OSManager encapsulates Native Host Kernel interactions so that the Daemon and Installer are cross-platform.
type OSManager interface {
	GetAppDataDir() (string, error)
	GetInstallDir() (string, error)
	IsRunningAsAdmin() bool
	StartLlamaEngine(modelPath string, port int) error
	StopLlamaEngine() error
	GetEngineState() string

	// Installer-Level Registration System
	IsInstalled() string
	RegisterApp(installDir string)
	UnregisterApp(installDir string)
	
	// Concurrency Prevention
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
