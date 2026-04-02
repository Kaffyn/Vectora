package commands

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Kaffyn/Vectora/src/core/ai"
	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/db"
	coreos "github.com/Kaffyn/Vectora/src/core/os"
	"github.com/Kaffyn/Vectora/src/core/rag"
	"github.com/Kaffyn/Vectora/src/core/server"
	"github.com/Kaffyn/Vectora/src/core/tool"
	"github.com/Kaffyn/Vectora/src/lib/bbolt"
	"github.com/Kaffyn/Vectora/src/lib/chromem"
	"github.com/Kaffyn/Vectora/src/lib/git"
	"github.com/Kaffyn/Vectora/src/lib/llama"

	"github.com/spf13/cobra"
)

var (
	llamaPort    int
	httpPort     int
	enableVulkan bool
)

func init() {
	startCmd.Flags().IntVarP(&llamaPort, "llama-port", "l", 8081, "Porta para o servidor llama.cpp")
	startCmd.Flags().IntVarP(&httpPort, "http-port", "p", 8080, "Porta para o servidor HTTP principal do Vectora")
	startCmd.Flags().BoolVarP(&enableVulkan, "vulkan", "v", false, "Habilita aceleração GPU via Vulkan para o llama.cpp")
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o servidor Vectora e o sidecar llama.cpp",
	Long: `Este comando inicia todos os componentes necessários do Vectora,
incluindo o servidor HTTP principal e o sidecar llama.cpp para embeddings e inferência.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func startServer() {
	ctx := context.Background()

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}
	dbPath := filepath.Join(dataDir, "vectora_rag.db")

	paths := config.GetDefaultPaths()
	if err := paths.EnsureDirectories(); err != nil {
		log.Fatalf("[Vectora] Falha ao garantir diretórios: %v", err)
	}

	isDev := os.Getenv("VECTORA_DEV_MODE") == "true"
	log.Printf("[Vectora] Mode: %s | Root: %s", func() string {
		if isDev {
			return "DEV"
		}
		return "PROD"
	}(), paths.Root)

	// 4. Inicialização do OS Manager dinâmico (MacOS/Linux/Windows)
	osManager := coreos.NewOSManager(paths)
	log.Printf("[Vectora] Usando OS Manager para: %s (Driver: %s)", runtime.GOOS, func() string {
		if runtime.GOOS == "darwin" {
			return "Metal"
		}
		return "Vulkan"
	}())

	sidecarManager := ai.NewSidecarManager(osManager, paths)

	textModel := "qwen3-0.6b"
	embeddingModel := "qwen3-embedding-0.6b"

	log.Printf("[Vectora] Iniciando Sidecars Qwen3 (Vulkan: %t)...", enableVulkan)

	// 1. Iniciar Sidecar de Texto (Porta 8081)
	if err := sidecarManager.StartLLAMA(ctx, "text", textModel, llamaPort, enableVulkan); err != nil {
		log.Printf("[Vectora] Erro: Falha ao iniciar sidecar de TEXTO: %v", err)
	} else {
		log.Printf("[Vectora] Sidecar de TEXTO iniciado na porta %d", llamaPort)
	}

	// 2. Iniciar Sidecar de Embedding (Porta 8082)
	embeddingPort := llamaPort + 1
	if err := sidecarManager.StartLLAMA(ctx, "embedding", embeddingModel, embeddingPort, enableVulkan); err != nil {
		log.Printf("[Vectora] Erro: Falha ao iniciar sidecar de EMBEDDING: %v", err)
	} else {
		log.Printf("[Vectora] Sidecar de EMBEDDING iniciado na porta %d", embeddingPort)
	}

	defer func() {
		log.Printf("[Vectora] Encerrando todos os sidecars...")
		sidecarManager.StopAll()
	}()

	bbDB, err := db.NewBboltDB(dbPath)
	if err != nil {
		log.Fatalf("[Vectora] Failed to open Bbolt at %s: %v", dbPath, err)
	}
	defer bbDB.Close()
	log.Printf("[Vectora] Bbolt ready at %s", dbPath)

	convRepo, _ := bbolt.NewConversationRepo(bbDB)
	metadataRepo, _ := bbolt.NewChunkRepo(bbDB)

	// 2. Initialize Chromem (In-process Vector DB)
	vectorRepo := chromem.NewVectorRepo()
	for _, entry := range []string{"engine-docs", "user-context"} {
		indexPath := filepath.Join(paths.Data, entry+".idx")
		if _, err := os.Stat(indexPath); err == nil {
			vectorRepo.LoadIndex(ctx, entry, indexPath)
		}
	}

	// 3. Initialize Services
	searchService := rag.NewSearchService(metadataRepo, vectorRepo)

	llamaURL := fmt.Sprintf("http://localhost:%d", llamaPort)
	llamaEmbedURL := fmt.Sprintf("http://localhost:%d", embeddingPort)
	embedder := llama.NewLlamaEmbedder(llamaEmbedURL)
	llm := llama.NewLlamaInference(llamaURL)

	chatService := rag.NewChatService(convRepo, osManager, embedder, llm, searchService)

	// 4. Initialize Protocols & Handlers
	gitBridge := git.NewGitBridge(paths.Root)
	acpHandler := server.NewACPHandler(gitBridge)

	registry := tool.NewRegistry()
	registry.InitDefaultTools(osManager, metadataRepo, embedder)

	settingsHandler := server.NewSettingsHandler(textModel, embeddingModel, sidecarManager)
	convHandler := server.NewConversationHandler(convRepo)
	searchHandler := server.NewSearchHandler(searchService)
	chatHandler := server.NewChatHandler(chatService)
	mcpHandler := server.NewMCPHandler(searchService, acpHandler, registry)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		textOk := sidecarManager.IsRunning("text")
		embeddingOk := sidecarManager.IsRunning("embedding")

		if textOk && embeddingOk {
			w.Write([]byte(`{"status":"ok","engine":"Vectora"}`))
		} else {
			// Retornar código de erro se o motor RAG / sidecars caíram
			http.Error(w, `{"status":"degraded","error":"Sidecars are offline"}`, http.StatusServiceUnavailable)
		}
	})

	mux.Handle("/api/settings", settingsHandler)
	mux.Handle("/api/chat", chatHandler)
	mux.Handle("/api/search", searchHandler)

	mux.Handle("/api/conversations", convHandler)
	mux.Handle("/api/conversations/", convHandler)

	mux.Handle("/api/mcp/tools", mcpHandler)
	mux.Handle("/api/mcp/tools/call", mcpHandler)

	addr := fmt.Sprintf(":%d", httpPort)
	log.Printf("[Vectora] Core HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[Vectora] Server error: %v", err)
	}
}
