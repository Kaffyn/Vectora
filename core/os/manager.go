package os

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
	GetSystemLanguage() string
}
