package main

import (
	"log"
	"os"

	"github.com/Kaffyn/vectora/internal/infra"
	"github.com/Kaffyn/vectora/internal/tray"
	"github.com/gen2brain/beeep"
	"golang.org/x/sys/windows"
)

func main() {
	// 0. Bloqueio de Múltiplas Instâncias via OS Mutex (Single-Instance Singleton)
	mutexName, _ := windows.UTF16PtrFromString("Global\\VectoraDaemonMutex_v1")
	_, err := windows.CreateMutex(nil, false, mutexName)
	if err == windows.ERROR_ALREADY_EXISTS {
		beeep.Notify("Vectora", "O Vectora já está em execução na bandeja do sistema. Verifique o ícone próximo ao relógio.", "info")
		os.Exit(0)
	}

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

	// 3. Notificação OS visual para Feedback de Inicialização
	err = beeep.Notify("Vectora", "O Vectora foi iniciado e está pronto para começar.", "")
	if err != nil {
		infra.Logger.Warn("Falha ao disparar notificacao de S.O: " + err.Error())
	}

	// 4. Start the Systray Engine (This is blocking)
	tray.Setup()
}
