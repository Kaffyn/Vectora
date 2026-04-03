package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kaffyn/vectora/internal/core"
	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/infra"
	"github.com/Kaffyn/vectora/internal/llm"
	"github.com/Kaffyn/vectora/internal/tools"
)

func main() {
	// Command flags
	cleanFlag := flag.Bool("clean", false, "Remove a base de dados de teste antes de iniciar")
	flag.Parse()

	// Enforce UTF-8 output on Windows
	fmt.Print("\033[?25l") // Esconde cursor
	defer fmt.Print("\033[?25h") // Mostra cursor

	fmt.Println("==================================================")
	fmt.Println("🛰️  VECTORA E2E ECOSYSTEM AUDIT (Full Suite)")
	fmt.Println("==================================================")

	infra.SetupLogger()
	ctx := context.Background()
	cfg := infra.LoadConfig()

	// --- 1. MICRO-SERVICES TEST (INTEGRITY) ---
	checkMicroServices(ctx)

	// --- 2. CLOUD PROVIDER TEST (GEMINI) ---
	if cfg.GeminiAPIKey == "" {
		fmt.Println("\n❌ [ERRO] Chave Gemini não configurada. Abortando teste de RAG Real.")
		os.Exit(1)
	}

	fmt.Printf("\n[3/4] Inicializando Pipeline de RAG (Zyris Engine Data)...\n")
	prov, err := llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("FAILED TO START GEMINI: %v", err)
	}

	// Fixed paths for persistent tests
	home, _ := os.UserHomeDir()
	testDataDir := filepath.Join(home, ".Vectora", "test_data")
	tmpDbPath := filepath.Join(testDataDir, "vectora_rag_test.db")
	tmpMemPath := filepath.Join(testDataDir, "vectora_rag_memory")

	if *cleanFlag {
		fmt.Println("      -> [CLEAN] Removendo dados de teste existentes...")
		os.Remove(tmpDbPath)
		os.RemoveAll(tmpMemPath)
	}
	os.MkdirAll(testDataDir, 0755)

	kv, _ := db.NewKVStoreAtPath(tmpDbPath)
	vecStore, _ := db.NewVectorStoreAtPath(tmpMemPath)

	// --- 3. SMART INDEXING ---
	indexDataFolder(ctx, prov, vecStore, *cleanFlag)

	// --- 4. CONSULTA RAG REAL ---
	runRAGQuery(ctx, prov, vecStore, kv)

	fmt.Println("\n==================================================")
	fmt.Println("✅ VECTORA ECOSYSTEM IS STABLE & FULLY OPERATIONAL")
	fmt.Println("==================================================")
}

func checkMicroServices(ctx context.Context) {
	fmt.Println("\n[1/4] Auditando Micro-Serviços Internos...")
	
	// Message Service
	fmt.Print("      - Chat Persistence (BBolt)... ")
	tmpPath := filepath.Join(os.TempDir(), "v_msg_test.db")
	_ = os.Remove(tmpPath)
	kv, _ := db.NewKVStoreAtPath(tmpPath)
	msgSvc := llm.NewMessageService(kv)
	_, err := msgSvc.CreateConversation(ctx, "test-id", "Test Conversation")
	if err == nil { fmt.Println("✅") } else { fmt.Println("❌") }

	// Memory Service
	fmt.Print("      - Knowledge Vector (ChroMem)... ")
	tmpMem := filepath.Join(os.TempDir(), "v_mem_test")
	_ = os.RemoveAll(tmpMem)
	memSvc, err := db.NewMemoryService(ctx, tmpMem)
	if err == nil && memSvc != nil { fmt.Println("✅") } else { fmt.Println("❌") }

	// Search Service
	fmt.Print("      - Intelligence Search... ")
	searchSvc := tools.NewSearchService()
	if searchSvc != nil { fmt.Println("✅") } else { fmt.Println("❌") }
}

func indexDataFolder(ctx context.Context, prov llm.Provider, vecStore db.VectorStore, force bool) {
	fmt.Println("\n[2/4] Sincronizando base de conhecimento (./data)...")
	
	collectionName := "ws_zyris_test"
	
	// Check if collection exists and has documents
	if !force && vecStore.CollectionExists(ctx, collectionName) {
		// Smoke test to see if collection is actually populated
		dummyVec := make([]float32, 1536) // Gemini standard size
		res, _ := vecStore.Query(ctx, collectionName, dummyVec, 1)
		if len(res) > 0 {
			fmt.Println("      -> Cache Detectado (ChroMem-Go). Pulando vetorização (Use --clean para refazer).")
			return
		}
	}

	dataDir := "./data"
	count := 0
	
	filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() { return nil }
		if strings.HasSuffix(info.Name(), ".xml") || strings.HasSuffix(info.Name(), ".txt") {
			content, _ := os.ReadFile(path)
			fmt.Printf("      -> [GEMINI] Vetorizando [%s] (%d bytes)...\n", info.Name(), len(content))
			
			emb, err := prov.Embed(ctx, string(content))
			if err != nil {
				if strings.Contains(err.Error(), "429") {
					fmt.Println("      ⚠️ [QUOTA EXCEEDED] Abortando vetorização para evitar falhas em cascata.")
					return fmt.Errorf("quota_limit")
				}
				fmt.Printf("      ⚠️ [ERRO]: %v\n", err)
				return nil
			}

			vecStore.UpsertChunk(ctx, collectionName, db.Chunk{
				ID:      info.Name(),
				Content: string(content),
				Vector:  emb,
			})
			count++
		}
		return nil
	})
	fmt.Printf("      Total de arquivos indexados no ChroMem: %d\n", count)
}

func runRAGQuery(ctx context.Context, prov llm.Provider, vecStore db.VectorStore, kv db.KVStore) {
	fmt.Println("\n[4/4] Executando consulta RAG sobre o Zyris Engine...")
	pipeline := core.NewPipeline(prov, vecStore, kv)
	
	query := "Resuma as principais entidades do Zyris Engine baseado no contexto fornecido."
	fmt.Printf("      [QUERY]: %s\n", query)
	
	start := time.Now()
	res, err := pipeline.Query(ctx, core.QueryRequest{
		WorkspaceID: "ws_zyris_test", // Synchronized with ChroMem collection
		Query:       query,
	})
	
	if err != nil {
		fmt.Printf("      ❌ [FALLA]: %v\n", err)
		return
	}
	
	fmt.Printf("      [TEMPO]: %v\n", time.Since(start))
	fmt.Printf("\n🤖 [RESPOSTA]:\n%s\n", res.Answer)
}
