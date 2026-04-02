package infra

import (
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
	
	// Ensure the directory exists
	if _, err := os.Stat(filepath.Dir(envPath)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(envPath), 0755)
	}

	// Try to load the .env file if it exists
	if err := godotenv.Load(envPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	return &Config{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	}
}
