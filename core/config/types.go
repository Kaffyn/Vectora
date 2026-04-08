package config

import (
	"time"
)

// Config representa o arquivo ~/.vectora/config.yaml
type Config struct {
	Version     string       `yaml:"version"`
	LastRun     time.Time    `yaml:"last_run"`

	// Global Settings
	DefaultModel  string `yaml:"default_model"`  // ex: "gemini-1.5-pro"
	FallbackCloud bool   `yaml:"fallback_cloud"` // Se true, usa cloud se local falhar

	// Network (Dev Mode)
	TCPPort int `yaml:"tcp_port,omitempty"` // 0 = desativado (apenas stdio/unix)

	// Providers (Chaves Criptografadas)
	Providers ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Gemini ProviderCreds `yaml:"gemini"`
	Claude ProviderCreds `yaml:"claude"`
	Qwen   ProviderCreds `yaml:"qwen"`
}

type ProviderCreds struct {
	APIKeyEncrypted string `yaml:"api_key_encrypted"` // Chave AES-GCM base64
	UseEnvVar       bool   `yaml:"use_env_var"`       // Se true, ignora encrypted e usa ENV
	EnvVarName      string `yaml:"env_var_name"`      // ex: "GEMINI_API_KEY"
}

// WorkspaceContext mantém o estado volátil de um workspace ativo em memória
type WorkspaceContext struct {
	Path        string
	ID          string // Hash do Path
	StoragePath string // ~/.vectora/workspaces/{ID}
	IsActive    bool
}
