package engines

import "time"

// Build representa uma versão compilada específica do llama.cpp.
type Build struct {
	ID          string `json:"id"`
	OS          string `json:"os"`           // "windows", "linux", "darwin"
	Architecture string `json:"architecture"` // "x86_64", "arm64", etc
	CPUFeatures  []string `json:"cpu_features"`  // "avx2", "avx512", "neon"
	GPU         string `json:"gpu"`          // "none", "cuda", "vulkan", "metal"
	GPUVersion  string `json:"gpu_version"`  // "11.8", "12.0" para CUDA
	DownloadURL string `json:"download_url"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	Description string `json:"description"`
}

// Catalog é o catálogo embarcado de builds disponíveis.
type Catalog struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Builds    []Build   `json:"builds"`
}

// Hardware descreve as capacidades do computador.
type Hardware struct {
	OS            string
	Architecture  string
	CPUFeatures   []string
	GPUType       string // "none", "cuda", "vulkan", "metal"
	GPUVersion    string
	RAM           int64
	CoreCount     int
}

// InstallationInfo contém informações sobre uma instalação de engine.
type InstallationInfo struct {
	BuildID     string
	Path        string
	Installed   time.Time
	SHA256      string
	SizeBytes   int64
	IsActive    bool
}

// DownloadProgress é enviado durante o download.
type DownloadProgress struct {
	Current int64
	Total   int64
	Speed   float64 // bytes/sec
}

// EngineProcess encapsula um processo llama.cpp em execução.
type EngineProcess struct {
	PID       int
	BuildID   string
	Listening bool
	Port      int // se disponível
}

// Error codes específicos do engines package
const (
	ErrHardwareDetectionFailed = "hardware_detection_failed"
	ErrBuildNotFound           = "build_not_found"
	ErrDownloadFailed          = "download_failed"
	ErrIntegrityCheckFailed    = "integrity_check_failed"
	ErrExtractionFailed        = "extraction_failed"
	ErrProcessSpawnFailed      = "process_spawn_failed"
	ErrAlreadyInstalled        = "already_installed"
	ErrNotInstalled            = "not_installed"
)
