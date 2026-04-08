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
	}
}

// SaveConfig persists sensitive keys like GeminiAPIKey to the user's .env file.
func SaveConfig(cfg *Config) error {
	userProfile, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	envPath := filepath.Join(userProfile, ".Vectora", ".env")
	content := fmt.Sprintf("GEMINI_API_KEY=%s\n", cfg.GeminiAPIKey)
	return os.WriteFile(envPath, []byte(content), 0600)
}
