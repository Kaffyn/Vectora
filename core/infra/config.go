// Package infra provides infrastructure utilities for Vectora.
//
// Configuration:
// Vectora uses a simple .env file for configuration, located at:
//   - Windows: %USERPROFILE%\.Vectora\.env
//   - Linux/macOS: ~/.Vectora/.env
//
// Supported environment variables:
//   - GEMINI_API_KEY: API key for Google Gemini (required for embeddings and chat)
//   - CLAUDE_API_KEY: API key for Anthropic Claude (optional, alternative provider)
//
// Example .env file:
//
//	GEMINI_API_KEY=AIzaSy...
//	CLAUDE_API_KEY=sk-ant-...
//
// Note: API keys are stored in plaintext. For production deployments with
// strict security requirements, consider using OS-level secret management
// (Windows Credential Manager, macOS Keychain, Linux libsecret) and
// loading keys from there instead.
package infra

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds the runtime configuration for Vectora.
type Config struct {
	GeminiAPIKey string
	ClaudeAPIKey string
}

// LoadConfig loads configuration from %USERPROFILE%\.Vectora\.env.
// If the file doesn't exist, it returns a Config with empty keys.
// This is the official configuration method for Vectora.
func LoadConfig() *Config {
	userProfile, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not find user home directory: %v", err)
	}

	envPath := filepath.Join(userProfile, ".Vectora", ".env")

	if _, err := os.Stat(filepath.Dir(envPath)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(envPath), 0755)
	}

	if err := godotenv.Overload(envPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	return &Config{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		ClaudeAPIKey: os.Getenv("CLAUDE_API_KEY"),
	}
}

// SaveConfig persists API keys to the user's .env file.
// Only keys that are non-empty are written.
func SaveConfig(cfg *Config) error {
	userProfile, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	envPath := filepath.Join(userProfile, ".Vectora", ".env")

	content := ""
	if cfg.GeminiAPIKey != "" {
		content += fmt.Sprintf("GEMINI_API_KEY=%s\n", cfg.GeminiAPIKey)
	}
	if cfg.ClaudeAPIKey != "" {
		content += fmt.Sprintf("CLAUDE_API_KEY=%s\n", cfg.ClaudeAPIKey)
	}

	return os.WriteFile(envPath, []byte(content), 0600)
}
