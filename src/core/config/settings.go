package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type VectoraSettings struct {
	ActiveProvider string `json:"active_provider"` // "qwen" ou "gemini"
	LocalModel     string `json:"local_model"`     // qwen3-0.6b, etc.
	GeminiModel    string `json:"gemini_model"`    // gemini-1.5-flash, etc.
	GeminiAPIKey   string `json:"gemini_api_key"`
	InstallDir     string `json:"install_dir"`
}

func GetSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".Vectora", "settings.json")
}

func LoadSettings() VectoraSettings {
	var s VectoraSettings
	data, err := os.ReadFile(GetSettingsPath())
	if err == nil {
		_ = json.Unmarshal(data, &s)
	}

	// Default fallback
	if s.ActiveProvider == "" {
		s.ActiveProvider = "qwen"
	}
	if s.LocalModel == "" {
		s.LocalModel = "qwen3-0.6b"
	}
	if s.GeminiModel == "" {
		s.GeminiModel = "gemini-1.5-flash"
	}
	return s
}

func SaveSettings(s VectoraSettings) error {
	path := GetSettingsPath()
	os.MkdirAll(filepath.Dir(path), 0755)

	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}
