package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/infra"
	"github.com/Kaffyn/vectora/internal/ipc"
	"github.com/Kaffyn/vectora/internal/llm"
	vecos "github.com/Kaffyn/vectora/internal/os"
	"github.com/Kaffyn/vectora/internal/tray"
)

func main() {
	// CLI Flags & Mode handling
	for _, arg := range os.Args {
		if arg == "--tests" {
			runSystemIntegrityTests()
			return
		}
	}

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

	// 4. Inicializa Motores de Back-End (Bancos + Servidor Sockets)
	infra.Logger.Info("Alocando Camada Oculta: Bancos BBolt/Chromem & IPC Sockets...")
	kvStore, dbErr := db.NewKVStore()
	if dbErr != nil {
		infra.Logger.Error("Fatal: BBolt Mount", "err", dbErr)
	}
	vecStore, vecErr := db.NewVectorStore()
	if vecErr != nil {
		infra.Logger.Error("Fatal: Chromem Mount", "err", vecErr)
	}
	
	ipcServer, ipcErr := ipc.NewServer()
	if ipcErr != nil {
		infra.Logger.Warn("Não foi possivel abrir o Socket UNIX IPC", "err", ipcErr)
	} else {
		// Mapeia todas as rotas centralizadas RPC
		ipc.RegisterRoutes(ipcServer, kvStore, vecStore, func() llm.Provider {
			return tray.ActiveProvider
		})
		
		go ipcServer.Start()
		defer ipcServer.Shutdown()
	}

	// 5. Notificação OS visual para Feedback de Inicialização rápida
	err = infra.NotifyOS("Vectora", "O assistente Vectora foi invocado no sistema e está residente.")
	if err != nil {
		infra.Logger.Warn("Falha ao disparar notificacao de S.O: " + err.Error())
	}

	// 6. Start the Systray Engine (This is blocking)
	tray.Setup()
}

// Smoke Tests Integrados (Validação de Sanidade das Engrenagens)
func runSystemIntegrityTests() {
	log.Println("=== VECTORA INTEGRITY TEST SUITE ===")
	
	// 1. Teste de Banco (Memória)
	tmpFile := filepath.Join(os.TempDir(), "vectora_smoke.db")
	defer os.Remove(tmpFile)
	log.Print("Checking KV Store (BBolt)... ")
	kv, err := db.NewKVStoreAtPath(tmpFile)
	if err != nil {
		log.Fatalf("FAILED: %v", err)
	}
	ctx := context.Background()
	kv.Set(ctx, "smoke", "ok", []byte("1"))
	val, _ := kv.Get(ctx, "smoke", "ok")
	if string(val) != "1" {
		log.Fatal("FAILED: Recovery Data Error")
	}
	log.Println("SUCCESS")

	// 2. Teste de Sockets (IPC)
	log.Print("Checking IPC Socket Engine... ")
	server, err := ipc.NewServer()
	if err != nil {
		log.Fatalf("FAILED: Socket Listen Error: %v", err)
	}
	_ = server
	log.Println("SUCCESS")

	// 3. Teste de Kernel O.S
	log.Print("Checking OS Manager... ")
	osMgr, err := vecos.NewManager()
	if err != nil {
		log.Fatalf("FAILED: Manager Failure: %v", err)
	}
	_ = osMgr // Ignora p/ silenciar linter se for apenas smoke check
	log.Println("SUCCESS (Detected Host: " + runtime.GOOS + ")")

	log.Println("====================================")
	log.Println("ALL CORE SYSTEMS OPERATIONAL (100%)")
}

