package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kaffyn/vectora/internal/core"
	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/infra"
	"github.com/Kaffyn/vectora/internal/llm"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 VECTORA E2E ZYRIS ENGINE TEST SUITE (NO MOCKS)")
	fmt.Println("==================================================")

	// 1. Inicializa o ambiente e Configurações Reais
	infra.SetupLogger()
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey == "" {
		log.Fatalf("[ERRO] Variável GEMINI_API_KEY não encontrada. Teste End-to-End necessita de chave real.")
	}

	ctx := context.Background()

	// 2. Cria Bancos Temporários Mas Reais (BBolt e Chromem)
	fmt.Println("[1/4] Montando Bancos de Dados Temporários...")
	tmpDbPath := filepath.Join(os.TempDir(), "vectora_e2e_bbolt.db")
	_ = os.Remove(tmpDbPath) // Limpeza de Execuções Antigas
	
	kvStore, err := db.NewKVStoreAtPath(tmpDbPath)
	if err != nil {
		log.Fatalf("Falha no BBolt: %v", err)
	}

	vecStore, err := db.NewVectorStore()
	if err != nil {
		log.Fatalf("Falha no Chromem: %v", err)
	}

	// 3. Setup Gemini Provider (Instanciamos o client real com a chave do usuário)
	fmt.Println("[2/4] Iniciando Conexão Real com Gemini AI (Auth)...")
	// Precisamos simular a factory do LLM com base na configuração do Tray
	// Aqui testamos a pipeline de RAG inteira via injeção
	geminiProvider, err := llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Erro ao montar o Gemini Provider: %v", err)
	}

	// 4. Ingestão e Vetorização Real
	fmt.Println("[3/4] Indexando pasta ./data (XML files)...")
	workspaceID := "zyris_engine_test"
	colName := "ws_" + workspaceID

	// Procurar XMLs
	dataDir := "./data"
	xmlFiles := 0
	err = filepath.Walk(dataDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil { return nil }
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".xml") {
			xmlFiles++
			contentBytes, _ := os.ReadFile(path)
			content := string(contentBytes)
			
			// Chunking Ingênuo de Teste (Simples Split) e Vetorização Unida
			fmt.Printf("      - Vetorizando %s (%d bytes)...\n", info.Name(), len(content))
			
			// Gera o Vector de embedding usando Gemini
			vec, embErr := geminiProvider.Embed(ctx, content)
			if embErr != nil {
				fmt.Printf("      [ERRO VETORIZACAO]: %v\n", embErr)
				return embErr
			}
			
			chunkID := fmt.Sprintf("chunk_%s_%d", info.Name(), time.Now().UnixNano())
			
			// Salva o Chunk Fisicamente
			if err := vecStore.UpsertChunk(ctx, colName, db.Chunk{
				ID:       chunkID,
				Content:  content,
				Vector:   vec,
				Metadata: map[string]string{"filename": info.Name(), "type": "xml"},
			}); err != nil {
				fmt.Printf("      [ERRO UPSERT]: %v\n", err)
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Falha crítica no processamento de dados: %v", err)
	}
	
	if xmlFiles == 0 {
		fmt.Printf("Aviso: zero arquivos XML encontrados no '%s'.\n", dataDir)
		log.Println("Crie um arquivo na pasta data para validação real (ex: zyris_player.xml).")
	}

	// 5. Querying Real (RAG em Ação)
	fmt.Println("[4/4] Consultando LLM usando o Cérebro local Vetorizado...")
	
	pipeline := core.NewPipeline(geminiProvider, vecStore, kvStore)
	
	queryStr := "Resuma como funciona a entidade base do Zyris Engine baseado no contexto que você tem guardado."
	req := core.QueryRequest{
		WorkspaceID: workspaceID,
		Query:       queryStr,
	}

	fmt.Printf("\n[PERGUNTA]: %s\n\n", queryStr)
	
	startQuery := time.Now()
	res, qErr := pipeline.Query(ctx, req)
	if qErr != nil {
		fmt.Printf("[ERRO QUERY PIPELINE]: %v\n", qErr)
		log.Fatalf("Falha na pipeline de execução: %v", qErr)
	}
	elapsed := time.Since(startQuery)

	fmt.Println("==================================================")
	fmt.Printf("[RESPOSTA GERADA] (Tempo de inferência: %v)\n%s\n", elapsed, res.Answer)
	fmt.Println("==================================================")

	fmt.Println("[SUCESSO] RAG Físico Completo executado adequadamente!")
}
