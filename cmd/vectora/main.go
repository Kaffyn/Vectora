package main

import (
	"log"

	"github.com/Kaffyn/vectora/internal/infra"
	"github.com/Kaffyn/vectora/internal/tray"
)

func main() {
	// 1. Initialize Logger
	if err := infra.SetupLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	infra.Logger.Info("Starting Vectora Daemon...")

	// 2. Load Configuration
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey == "" {
		infra.Logger.Warn("GEMINI_API_KEY is not set. Cloud features may be unavailable.")
	}

	// 3. Start the Systray Engine (This is blocking)
	tray.Setup()
}
