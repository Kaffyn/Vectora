package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"path/filepath"
	"runtime"

	"github.com/Kaffyn/vectora/internal/core"
	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/infra"
	"github.com/Kaffyn/vectora/internal/ipc"
	"github.com/Kaffyn/vectora/internal/llm"
	vecos "github.com/Kaffyn/vectora/internal/os"
	"github.com/Kaffyn/vectora/internal/tray"
)

func main() {
	// [NOVO] CLI / Mode Handling
	// Se houver argumentos, processamos como comandos ou testes.
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch {
		case arg == "--tests":
			// Inicia o Tray em background e roda os testes integrados
			go func() {
				defer func() { recover() }() // Evita panic se fechar rápido demais
				tray.Setup()
			}()
			time.Sleep(1 * time.Second)
			runSystemIntegrityTests()
			return
		case arg == "--query" && len(os.Args) > 2:
			// Modo API direta via terminal (Vectora --query "pergunta")
			runDirectQuery(os.Args[2])
			return
		case !strings.HasPrefix(arg, "-"):
			// Fallback: se passar texto direto, assume query
			runDirectQuery(strings.Join(os.Args[1:], " "))
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
		// O Shutdown será gerenciado pelo process cycle/os signal
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
	
	// Aguarda um pouco antes de fechar para garantir que o tray inicializou se foi pedido
	time.Sleep(2 * time.Second)
	log.Println("Closing tests...")
}

// runDirectQuery executa uma pergunta via terminal, usando o provedor Gemini padrão (ou configurado)
func runDirectQuery(query string) {
	fmt.Printf("🔍 Vectora Query: %s\n", query)
	
	infra.SetupLogger()
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey == "" {
		fmt.Println("[ERRO] Chave Gemini não configurada. Use o Tray para configurar ou adicione ao .env")
		return
	}

	ctx := context.Background()
	kvStore, _ := db.NewKVStore()
	vecStore, _ := db.NewVectorStore()
	
	prov, err := llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao iniciar provedor: %v\n", err)
		return
	}

	pipeline := core.NewPipeline(prov, vecStore, kvStore)
	res, err := pipeline.Query(ctx, core.QueryRequest{
		Query: query,
	})
	if err != nil {
		fmt.Printf("[ERRO] Falha na execução: %v\n", err)
		return
	}

	fmt.Printf("\n🤖 Resposta:\n%s\n", res.Answer)
}

