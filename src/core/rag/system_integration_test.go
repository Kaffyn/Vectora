package rag_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/db"
	"github.com/Kaffyn/Vectora/src/lib/bbolt"
	"github.com/Kaffyn/Vectora/src/lib/chromem"
	"github.com/Kaffyn/Vectora/src/os/windows"
)

// TestFullStackIntegration realiza um teste real entre llama, bbolt e chromem-go.
// Requer que o modelo GGUF esteja em models/gguf/qwen-0.5b-coder/model.gguf
func TestFullStackIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste de integração de sistema (long-running)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Configuração de Caminhos para teste
	testDir := filepath.Join(os.TempDir(), "vectora_test")
	_ = os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	paths := config.NewVectoraPaths(testDir)

	// Criar estrutura necessária
	os.MkdirAll(paths.GetDBDir(), 0755)
	os.MkdirAll(paths.GetIndicesDir(), 0755)

	// 2. Inicializar Persistência
	bbDB, err := db.NewBboltDB(paths.GetDBPath("test.db"))
	if err != nil {
		t.Fatalf("Erro ao abrir bbolt: %v", err)
	}
	defer bbDB.Close()

	chunkMetaRepo, _ := bbolt.NewChunkRepo(bbDB)
	vectorRepo := chromem.NewVectorRepo()

	// 3. Inicializar e Subir LLAMA (SIDE-CAR)
	osManager := windows.NewWindowsManager(paths)
	modelPath := filepath.Join("c:/Users/bruno/Desktop/Vectora/models/gguf/qwen3-0.6b", "qwen3-0.6b.gguf")

	err = osManager.StartLLAMAServer(ctx, "text", modelPath, 8080, false)
	if err != nil {
		t.Fatalf("Erro ao iniciar llama-server: %v", err)
	}
	defer osManager.StopAllLLAMAServers()

	// 4. Fluxo de Ingestão Real
	// Aqui o IngestService precisaria de um Embedder Real conversando com o llama via HTTP
	// TODO: Implementar o LLAMAEmbedder para fechar o ciclo
	fmt.Println("🚀 Integração Llama + Bbolt + Chromem pronta para testes.")
}
