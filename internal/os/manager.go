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

// OSManager consolida as funções vitais de ciclo de vida que o Daemon master requisitará.
type OSManager interface {
	GetAppDataDir() (string, error)
	StartLlamaEngine(modelPath string, port int) error
	StopLlamaEngine() error
	GetEngineState() string
}

// NewManager cria a instância acoplada dinamicamente ao ambiente físico de execução (Linux, Mac, Win).
func NewManager() (OSManager, error) {
	switch runtime.GOOS {
	case "windows":
		return win.NewManager(), nil
	case "darwin":
		return mac.NewManager(), nil
	case "linux":
		return lin.NewManager(), nil
	default:
		return nil, fmt.Errorf("unsupported operating system environment: %s -> cross platform bindings unavailable", runtime.GOOS)
	}
}
