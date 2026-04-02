package tool_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/config"
	"github.com/Kaffyn/Vectora/src/core/domain"
	coreos "github.com/Kaffyn/Vectora/src/core/os"
	"github.com/Kaffyn/Vectora/src/core/tool"
)

// TestFileSystemTools (100%): Cenário de sucesso para Read, Write e Edit.
func TestFileSystemTools(t *testing.T) {
	dir, _ := os.MkdirTemp("", "tool-test-*")
	defer os.RemoveAll(dir)

	ctx := context.Background()
	filePath := filepath.Join(dir, "test.txt")

	// 1. Testar Write
	writer := tool.NewWriteFileTool()
	_, err := writer.Execute(ctx, map[string]interface{}{
		"path":    filePath,
		"content": "versao 1\ncontexto unico",
	})
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// 2. Testar Read
	reader := tool.NewReadFileTool()
	res, _ := reader.Execute(ctx, map[string]interface{}{"path": filePath})
	m := res.(map[string]string)
	if m["content"] != "versao 1\ncontexto unico" {
		t.Errorf("ReadFile mismatch. Got: %s", m["content"])
	}

	// 3. Testar Edit
	editor := tool.NewEditTool()
	_, errEdit := editor.Execute(ctx, map[string]interface{}{
		"path":        filePath,
		"target":      "versao 1",
		"replacement": "versao 2",
	})
	if errEdit != nil {
		t.Fatalf("EditTool failed: %v", errEdit)
	}

	// Validar resultado do edit
	res2, _ := reader.Execute(ctx, map[string]interface{}{"path": filePath})
	m2 := res2.(map[string]string)
	if m2["content"] != "versao 2\ncontexto unico" {
		t.Errorf("Edit mismatch. Got: %s", m2["content"])
	}
}

// TestSearchTools (200%): Cenário de sucesso para FindFiles e GrepSearch.
func TestSearchTools(t *testing.T) {
	dir, _ := os.MkdirTemp("", "search-test-*")
	defer os.RemoveAll(dir)

	ctx := context.Background()
	filePath := filepath.Join(dir, "search_me.go")
	os.WriteFile(filePath, []byte("package test\nfunc TargetFunction() {}"), 0644)

	// 1. Testar FindFiles (Glob)
	finder := tool.NewFindFilesTool()
	resFind, _ := finder.Execute(ctx, map[string]interface{}{"pattern": filepath.Join(dir, "*.go")})
	mf := resFind.(map[string]interface{})
	if mf["count"].(int) != 1 {
		t.Errorf("FindFiles failed to locate file. Count: %d", mf["count"])
	}

	// 2. Testar GrepSearch (Conteúdo)
	grepper := tool.NewGrepSearchTool()
	resGrep, _ := grepper.Execute(ctx, map[string]interface{}{
		"pattern": "TargetFunction",
		"root":    dir,
	})
	mg := resGrep.(map[string]interface{})
	if mg["count"].(int) != 1 {
		t.Errorf("GrepSearch failed to find text. Count: %d", mg["count"])
	}
}

// TestNegativeScenarios (300%): Testando falhas e erros de validação.
func TestNegativeScenarios(t *testing.T) {
	ctx := context.Background()

	// Erro: Arquivo inexistente
	reader := tool.NewReadFileTool()
	_, err := reader.Execute(ctx, map[string]interface{}{"path": "non-existent.txt"})
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}

	// Erro: Edit com alvo ambíguo (não único)
	dir, _ := os.MkdirTemp("", "neg-test-*")
	defer os.RemoveAll(dir)
	fPath := filepath.Join(dir, "ambiguous.txt")
	os.WriteFile(fPath, []byte("repetido\nrepetido"), 0644)

	editor := tool.NewEditTool()
	_, errEdit := editor.Execute(ctx, map[string]interface{}{
		"path":        fPath,
		"target":      "repetido",
		"replacement": "novo",
	})
	if errEdit == nil || !strings.Contains(errEdit.Error(), "não é único") {
		t.Errorf("Expected ambiguity error, got: %v", errEdit)
	}
}

// TestShellTool (200%): Cenário de sucesso para execução de shell via OSManager.
func TestShellTool(t *testing.T) {
	paths := config.GetDefaultPaths()
	osm := coreos.NewOSManager(paths)
	shell := tool.NewShellTool(osm)

	ctx := context.Background()
	res, err := shell.Execute(ctx, map[string]interface{}{
		"command":   "go",
		"arguments": []interface{}{"version"},
	})
	if err != nil {
		t.Fatalf("Shell execution failed: %v", err)
	}
	m := res.(map[string]string)
	if m["status"] != "success" {
		t.Errorf("Expected success, got %s", m["status"])
	}
}

// TestIntegratedSearchAndFile (300%): Ciclo completo de criação e descoberta técnica.
func TestIntegratedSearchAndFile(t *testing.T) {
	dir, _ := os.MkdirTemp("", "inter-test-*")
	defer os.RemoveAll(dir)

	ctx := context.Background()
	secretKey := "KEY-DE-INTEGRACAO-VECTORA-99"
	filePath := filepath.Join(dir, "vault.secret")

	// 1. Ação: Escrever o segredo
	writer := tool.NewWriteFileTool()
	writer.Execute(ctx, map[string]interface{}{
		"path":    filePath,
		"content": secretKey,
	})

	// 2. Descoberta: Procurar o segredo via GrepSearch
	grepper := tool.NewGrepSearchTool()
	res, _ := grepper.Execute(ctx, map[string]interface{}{
		"pattern": secretKey,
		"root":    dir,
	})

	mg := res.(map[string]interface{})
	if mg["count"].(int) != 1 {
		t.Errorf("Integrated search failed to discover the written file. Count: %d", mg["count"])
	}
}

// Mocks para Memória Semântica
type MockEmbedder struct{}

func (m *MockEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2}, nil
}
func (m *MockEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	return [][]float32{{0.1, 0.2}}, nil
}

type MockRepo struct{}

func (m *MockRepo) Save(ctx context.Context, chunk *domain.Chunk, emb []float32) error {
	return nil
}
func (m *MockRepo) Search(ctx context.Context, query string, limit int) ([]*domain.Chunk, error) {
	return []*domain.Chunk{}, nil
}
func (m *MockRepo) VectorSearch(ctx context.Context, emb []float32, limit int) ([]*domain.Chunk, error) {
	return []*domain.Chunk{}, nil
}
func (m *MockRepo) GetByDocumentID(ctx context.Context, docID string) ([]*domain.Chunk, error) {
	return []*domain.Chunk{}, nil
}
func (m *MockRepo) Delete(ctx context.Context, id string) error { return nil }

// TestMemoryIntegration (300%): Cenário de sucesso para memorização de fatos.
func TestMemoryIntegration(t *testing.T) {
	ctx := context.Background()
	saver := tool.NewSaveMemoryTool(&MockRepo{}, &MockEmbedder{})

	res, err := saver.Execute(ctx, map[string]interface{}{
		"knowledge": "Vectora suporta Vulkan nativamente",
		"tag":       "engine",
	})

	if err != nil {
		t.Fatalf("SaveMemory failed: %v", err)
	}

	m := res.(map[string]string)
	if m["status"] != "memorized" {
		t.Errorf("Expected status memorized, got %s", m["status"])
	}
}
