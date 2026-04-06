package models

// Model representa um modelo de IA que pode ser gerenciado pelo MPM
type Model struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	SizeBytes      int64    `json:"size_bytes"`
	RequiredRAMGB  float64  `json:"required_ram_gb"`
	RequiredVRAMGB float64  `json:"required_vram_gb"`
	Capabilities   []string `json:"capabilities"`
	Tags           []string `json:"tags"`
	SHA256         string   `json:"sha256"`
	HuggingFaceID  string   `json:"huggingface_id"`
}

// Catalog representa o catálogo completo de modelos disponíveis
type Catalog struct {
	Models []Model `json:"models"`
}

// Hardware descreve as capacidades de hardware do sistema
type Hardware struct {
	OS           string
	Architecture string
	CPUFeatures  []string
	CoreCount    int
	RAM          int64 // em bytes
	GPUType      string // "none", "cuda", "vulkan", "metal"
	GPUVersion   string
}

// ModelManager orquestra operações com modelos
type ModelManager struct {
	catalog   *Catalog
	modelsDir string
	metadata  map[string]*InstalledModel
}

// InstalledModel representa metadados de um modelo instalado
type InstalledModel struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Installed bool   `json:"installed"`
	SHA256    string `json:"sha256"`
	Size      int64  `json:"size"`
}

// DownloadProgress rastreia progresso de download
type DownloadProgress struct {
	Downloaded      int64
	Total           int64
	PercentComplete float64
	Speed           float64 // bytes por segundo
}

// IPCError define erro de comunicação IPC
type IPCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Erros personalizados do MPM
const (
	ErrModelNotFound   = "mpm_model_not_found"
	ErrInsufficientRAM = "mpm_insufficient_ram"
	ErrDownloadFailed  = "mpm_download_failed"
	ErrVerifyFailed    = "mpm_verify_failed"
	ErrInstallFailed   = "mpm_install_failed"
)
