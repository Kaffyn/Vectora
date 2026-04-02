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

func main() {
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
			if err := vectorRepo.LoadIndex(context.Background(), id, indexPath); err != nil {
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
	fmt.Printf("📂 Base: %s\n", vectoraDir)

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
		resp, err := chatSvc.SendMessage(context.Background(), req.ConversationID, req.Message)
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

	// 6. Monitor de Sinais para shutdown gracioso
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Printf("\n🛑 Recebido sinal %v. Encerrando Sidecars...\n", sig)
		osManager.StopLLAMAServer("text")
		osManager.StopLLAMAServer("embedding")
		cancel()
		os.Exit(0)
	}()

	// Manter vivo para os testes
	<-ctx.Done()
}

func applyCognitiveSettings(s config.VectoraSettings, osm domain.OSManager) {
	fmt.Printf("🧠 Ajustando Redes Cognitivas -> Provedor: %s\n", s.ActiveProvider)
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
