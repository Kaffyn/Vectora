package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = "config.yaml"

type Manager struct {
	ConfigPath string
	Config     *Config
}

func NewManager(homeDir string) (*Manager, error) {
	configDir := filepath.Join(homeDir, ".vectora")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return nil, err
	}

	m := &Manager{
		ConfigPath: filepath.Join(configDir, ConfigFileName),
		Config:     &Config{},
	}

	// Tenta carregar existente
	if _, err := os.Stat(m.ConfigPath); err == nil {
		if err := m.Load(); err != nil {
			return nil, err
		}
	} else {
		// Cria default
		m.Config = &Config{
			Version:       "1.0.0",
			DefaultModel:  "gemini-1.5-flash",
			FallbackCloud: true,
			TCPPort:       0, // Desativado por padrão
		}
		if err := m.Save(); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, m.Config)
}

func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.Config)
	if err != nil {
		return err
	}
	return os.WriteFile(m.ConfigPath, data, 0600)
}

// SetAPIKey criptografa e salva a chave no config
func (m *Manager) SetAPIKey(provider string, key string) error {
	enc, err := EncryptSecret(key)
	if err != nil {
		return err
	}

	switch provider {
	case "gemini":
		m.Config.Providers.Gemini.APIKeyEncrypted = enc
		m.Config.Providers.Gemini.UseEnvVar = false
	case "claude":
		m.Config.Providers.Claude.APIKeyEncrypted = enc
		m.Config.Providers.Claude.UseEnvVar = false
	}

	return m.Save()
}

// GetAPIKey descriptografa e retorna a chave
func (m *Manager) GetAPIKey(provider string) (string, error) {
	var creds ProviderCreds
	switch provider {
	case "gemini":
		creds = m.Config.Providers.Gemini
	case "claude":
		creds = m.Config.Providers.Claude
	default:
		return "", nil
	}

	if creds.UseEnvVar {
		return os.Getenv(creds.EnvVarName), nil
	}

	if creds.APIKeyEncrypted == "" {
		return "", nil
	}

	return DecryptSecret(creds.APIKeyEncrypted)
}
