package infra

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiAPIKey string
	ClaudeAPIKey string
}

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
