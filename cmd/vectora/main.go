package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/getlantern/systray"

	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/db"
	"github.com/Kaffyn/Vectora/src/core/domain"
	coreos "github.com/Kaffyn/Vectora/src/core/os"
	"github.com/Kaffyn/Vectora/src/core/rag"
	"github.com/Kaffyn/Vectora/src/core/server"
	"github.com/Kaffyn/Vectora/src/core/tool"
	"github.com/Kaffyn/Vectora/src/lib/bbolt"
	"github.com/Kaffyn/Vectora/src/lib/chromem"
	"github.com/Kaffyn/Vectora/src/lib/gemini"
	"github.com/Kaffyn/Vectora/src/lib/git"
	"github.com/Kaffyn/Vectora/src/lib/llama"
)

var (
	// These will be removed/refactored as the architecture evolves
	// engineCmd *exec.Cmd
	// webCmd    *exec.Cmd
	acpLevel = "guarded" // Default ACP level
)

func main() {
	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to signal when core logic is ready
	coreReady := make(chan struct{})

	// Run the core logic in a goroutine
	go func() {
		defer close(coreReady) // Signal that core is initialized
		runCore(ctx)
	}()

	// Run the systray in a goroutine
	go func() {
		<-coreReady // Wait for core to be ready
		systray.Run(onReady, onExit)
	}()

	// Monitor signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		fmt.Printf("
🛑 Recebido sinal de encerramento. Encerrando Vectora...
")
		// Trigger shutdown for core components
		cancel()
	case <-ctx.Done():
		// Core logic finished, or was cancelled
		fmt.Printf("
🛑 Core logic encerrada. Encerrando Vectora...
")
	}

	// Give a moment for cleanup
	// time.Sleep(time.Second) // Consider if needed
	os.Exit(0)
}

func runCore(ctx context.Context) {
	// 1. Configuração de Caminhos
	os.Setenv("VECTORA_DEV_MODE", "true") // Simulando DEV para test local ou falso
	paths := config.GetDefaultPaths()

	userHome, _ := os.UserHomeDir()
	vectoraDir := filepath.Join(userHome, ".Vectora")

	// 2. Inicialização do Banco de Dados (bbolt)
	bbDB, err := db.NewBboltDB(filepath.Join(paths.DB, "vectora.db"))
	if err != nil {
		log.Fatalf("Falha ao iniciar bbolt: %v", err)
	}
	defer bbDB.Close()

	// 3. Inicialização do Repositório Vetorial (chromem-go)
	vectorRepo := chromem.NewVectorRepo()

	// Carregar índices existentes (.idx) de paths.Data
	indices, _ := os.ReadDir(paths.Data)
	for _, entry := range indices {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".idx" {
			id := strings.TrimSuffix(entry.Name(), ".idx")
			indexPath := filepath.Join(paths.Data, entry.Name())
			if err := vectorRepo.LoadIndex(ctx, id, indexPath); err != nil {
				log.Printf("⚠️ Falha ao carregar índice [%s]: %v", id, err)
			} else {
				log.Printf("📚 Índice carregado: %s", id)
			}
		}
	}

	// 4. Inicialização do OS Manager dinâmico (MacOS/Linux/Windows)
	osManager := coreos.NewOSManager(paths)

	// 4.1 - Inicialização da Ponte Autônoma (ACP + Git)
	gitBridge := git.NewGitBridge(paths.Root)
	acpHandler := server.NewACPHandler(gitBridge)

	// 4.2 - Inicialização de Provedores Cognitivos Padrão
	var embedder domain.EmbeddingProvider

	settings := config.LoadSettings()
	if settings.ActiveProvider == "gemini" {
		embedder = gemini.NewGeminiEmbedder("text-embedding-004")
	} else {
		embedder = llama.NewLlamaEmbedder("http://localhost:8082")
	}

	// 5. Ingestão e Busca Service
	convRepo, _ := bbolt.NewConversationRepo(bbDB)
	chunkRepo, _ := bbolt.NewChunkRepo(bbDB)

	// 5.1 - Inicialização do Catálogo Global de Ferramentas
	registry := tool.NewRegistry()
	registry.InitDefaultTools(osManager, chunkRepo, embedder)

	// Mock Search Service (Unindo metadados bbolt + vetores chromem)
	searchService := rag.NewSearchService(chunkRepo, vectorRepo)

	fmt.Println("🚀 Vectora Core Iniciado!")
	fmt.Printf("📂 Base: %s
", vectoraDir)

	// 5.5 - Lógica Híbrida Gemini vs Local Qwen (Settings)
	applyCognitiveSettings(settings, osManager)

	http.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(config.LoadSettings())
			return
		}

		if r.Method == http.MethodPost {
			var incoming config.VectoraSettings
			if err := json.NewDecoder(r.Body).Decode(&incoming); err == nil {
				config.SaveSettings(incoming)
				applyCognitiveSettings(incoming, osManager)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "ok"}`))
			} else {
				http.Error(w, "Bad JSON", http.StatusBadRequest)
			}
		}
	})

	go func() {
		fmt.Println("🌐 API Backend Core Operante na Porta :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Falha no servidor Core HTTP: %v", err)
		}
	}()

	// 5.7 - Chat Service logic
	http.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method != http.MethodPost {
			return
		}

		var req struct {
			Message        string `json:"message"`
			ConversationID string `json:"conversationId"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		s := config.LoadSettings()

		// Atualizar provedores se necessário (Simplificado para consistência)
		var currentEmbedder domain.EmbeddingProvider
		var currentLLM domain.LLMProvider

		if s.ActiveProvider == "gemini" {
			currentEmbedder = gemini.NewGeminiEmbedder("text-embedding-004")
			currentLLM = gemini.NewGeminiProvider(s.GeminiModel)
		} else {
			currentEmbedder = llama.NewLlamaEmbedder("http://localhost:8082")
			currentLLM = gemini.NewGeminiProvider("qwen-mock")
		}

		chatSvc := rag.NewChatService(convRepo, osManager, currentEmbedder, currentLLM, searchService)
		resp, err := chatSvc.SendMessage(ctx, req.ConversationID, req.Message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"reply": resp})
	})

	// 5.8 - Registry de Protocolos (MCP & ACP)
	mcpHandler := server.NewMCPHandler(searchService, acpHandler, registry)
	http.Handle("/api/mcp/", mcpHandler)
	http.Handle("/api/acp/", acpHandler)

	// 5.9 - Dashboard UI (React/Next sidecar)

	// ... rest ...

	// Keep core running until context is cancelled
	<-ctx.Done()
	fmt.Println("Core logic shutting down...")
	// Perform cleanup tasks here before exiting
	osManager.StopLLAMAServer("text")
	osManager.StopLLAMAServer("embedding")
}

func onReady() {
	systray.SetTitle("Vectora")
	systray.SetTooltip("Vectora - AI Engine")

	// Usaremos um botão vazio por não termos ícone agora
	// Para uso real, deve se carregar um byte slice .ico
	systray.SetIcon(getDummyIcon())

	// These buttons will be refactored as the architecture changes
	mEngineStart := systray.AddMenuItem("Start RAG Engine (N/A)", "Engine runs automatically")
	mEngineStop := systray.AddMenuItem("Stop RAG Engine (N/A)", "Engine runs automatically")
	systray.AddSeparator()

	mWebStart := systray.AddMenuItem("Start Web UI", "Starts the Next.js Frontend")
	mWebStop := systray.AddMenuItem("Stop Web UI", "Stops the Next.js Frontend")
	systray.AddSeparator()

	mACP := systray.AddMenuItem("Autonomy Level (ACP)", "Set AI agent autonomy level")
	mLevelAsk := mACP.AddSubMenuItem("Ask Everything", "Confirm every AI action")
	mLevelGuarded := mACP.AddSubMenuItem("Guarded (Safe)", "Auto-approve reads, confirm writes")
	mLevelYOLO := mACP.AddSubMenuItem("YOLO (Full Auto)", "IA does everything without asking")
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit Vectora", "Exits the manager and all services")

	// Default to guarded, but reflect actual acpLevel if it's set on startup
	if acpLevel == "ask-any" {
		mLevelAsk.Check()
	} else if acpLevel == "yolo" {
		mLevelYOLO.Check()
	} else {
		mLevelGuarded.Check()
	}

	mEngineStart.Disable() // Engine is part of this process
	mEngineStop.Disable()  // Engine is part of this process
	mWebStop.Disable()     // Web UI start/stop logic needs to be integrated with IPC

	rootPath, _ := os.Getwd()

	go func() {
		for {
			select {
			case <-mEngineStart.ClickedCh:
				// Engine is always running, maybe add a restart option?
				fmt.Println("Engine is already running as part of the daemon.")
			case <-mEngineStop.ClickedCh:
				// Engine is always running, maybe add a restart option?
				fmt.Println("Engine is already running as part of the daemon.")

			case <-mWebStart.ClickedCh:
				mWebStart.Disable()
				// This will eventually call the IPC client to start the Wails app
				fmt.Printf("Web UI start requested. Placeholder for IPC call. RootPath: %s
", rootPath)
				// if err := startWeb(rootPath); err != nil {
				// 	fmt.Printf("Web start error: %v
", err)
				// 	mWebStart.Enable()
				// } else {
				// 	mWebStop.Enable()
				// }
			case <-mWebStop.ClickedCh:
				mWebStop.Disable()
				// This will eventually call the IPC client to stop the Wails app
				fmt.Println("Web UI stop requested. Placeholder for IPC call.")
				// stopWeb()
				// mWebStart.Enable()

			case <-mLevelAsk.ClickedCh:
				mLevelAsk.Check()
				mLevelGuarded.Uncheck()
				mLevelYOLO.Uncheck()
				setACPLevel("ask-any")
			case <-mLevelGuarded.ClickedCh:
				mLevelAsk.Uncheck()
				mLevelGuarded.Check()
				mLevelYOLO.Uncheck()
				setACPLevel("guarded")
			case <-mLevelYOLO.ClickedCh:
				mLevelAsk.Uncheck()
				mLevelGuarded.Uncheck()
				mLevelYOLO.Check()
				setACPLevel("yolo")

			case <-mQuit.ClickedCh:
				// Signal the main goroutine to shut down
				// This will trigger the context cancellation in main
				fmt.Println("Quit requested from systray.")
				// No direct os.Exit(0) here, let main handle graceful shutdown
				// For now, systray.Quit() will be called, which will lead to program exit.
				// In a full IPC system, this would send a shutdown signal to the daemon.
				systray.Quit()
				return
			}
		}
	}()
}

func setACPLevel(level string) {
	fmt.Printf("🛡️ Tray -> Core: Alterando nível ACP para [%s]
", level)
	// This will eventually communicate via IPC to the core to set the ACP level
	// http.Post("http://localhost:8080/api/settings", ... (Ajustar quando endpoint Settings suportar ACP))
	acpLevel = level // Update global variable for now
}

func onExit() {
	// Cleanup all child processes cleanly (if any were started directly by systray)
	// These will be handled by the main daemon process now
	// stopWeb()
	// stopEngine()
	fmt.Println("Systray exiting.")
}

// These functions will be removed or heavily refactored later
// as the daemon will manage starting/stopping web UI via IPC
// and the engine is integrated.

// var (
// 	engineCmd *exec.Cmd
// 	webCmd    *exec.Cmd
// )

// func startEngine(rootPath string) error {
// 	exePath := filepath.Join(rootPath, "vectora-cli.exe") // This path will change
// 	engineCmd = exec.Command(exePath, "start")
	
// 	engineCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

// 	return engineCmd.Start()
// }

// func stopEngine() {
// 	if engineCmd != nil && engineCmd.Process != nil {
// 		engineCmd.Process.Kill()
// 		engineCmd.Wait()
// 		engineCmd = nil
// 	}
// }

// func startWeb(rootPath string) error {
// 	webPath := filepath.Join(rootPath, "src", "web") // This path will change
	
// 	webCmd = exec.Command("bun", "dev")
// 	webCmd.Dir = webPath
	
// 	webCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

// 	return webCmd.Start()
// }

// func stopWeb() {
// 	if webCmd != nil && webCmd.Process != nil {
// 		webCmd.Process.Kill()
// 		webCmd.Wait()
// 		webCmd = nil
// 	}
// }


func getDummyIcon() []byte {
	// 1x1 transparent ICO just to avoid panic
	return []byte{
		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00,
		0x18, 0x00, 0x30, 0x00, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func applyCognitiveSettings(s config.VectoraSettings, osm domain.OSManager) {
	fmt.Printf("🧠 Ajustando Redes Cognitivas -> Provedor: %s
", s.ActiveProvider)
	if s.ActiveProvider == "gemini" {
		// Stop local LLM
		fmt.Println("☁️  Gemini Ativado! Suspendendo Llama.cpp e liberando 2GB RAM...")
		osm.StopLLAMAServer("text")
		osm.StopLLAMAServer("embedding")
		os.Setenv("GEMINI_API_KEY", s.GeminiAPIKey)
	} else {
		// Start Local Qwen
		fmt.Println("💻 Qwen Local Selecionado! Bootando binários Native Llama.cpp...")
		os.Unsetenv("GEMINI_API_KEY")
		// (Isso depende de ter a flag `--model` ou config de default)
		// osm.EnsureLLAMAServer(..., "text")
	}
}