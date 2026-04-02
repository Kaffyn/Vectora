package main

import (
	"log"
	"os"

	"github.com/Kaffyn/vectora/internal/infra"
	vecos "github.com/Kaffyn/vectora/internal/os"
	"github.com/Kaffyn/vectora/internal/tray"
)

func main() {
	// 0. Verifica S.O Physical Support e Trata
	systemManager, err := vecos.NewManager()
	if err != nil {
		log.Fatalf("Critical Hardware OS Failure: %v", err)
	}

	// 1. Bloqueio de Múltiplas Instâncias Agilizado Baseado no S.O Host (Mac/Linux TPC vs Win32 Mutex)
	if err := systemManager.EnforceSingleInstance(); err != nil {
		infra.NotifyOS("Vectora em Rodagem", "O Vectora já está em execução na bandeja do sistema. Verifique o ícone próximo ao relógio.")
		os.Exit(0)
	}

	// 2. Initialize Logger
	if err := infra.SetupLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	infra.Logger.Info("Starting Vectora Daemon...")

	// 3. Load Configuration
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey == "" {
		infra.Logger.Warn("GEMINI_API_KEY is not set. Cloud features may be unavailable.")
	}

	// 4. Notificação OS visual para Feedback de Inicialização rápida
	err = infra.NotifyOS("Vectora", "O assistente Vectora foi invocado no sistema e está residente.")
	if err != nil {
		infra.Logger.Warn("Falha ao disparar notificacao de S.O: " + err.Error())
	}

	// 5. Start the Systray Engine (This is blocking)
	tray.Setup()
}
