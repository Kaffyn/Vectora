package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/config"
)

// TestEnsureDirectories segue o padrão Vectora 300% de Testes (Happy, Negative, Edge).
func TestEnsureDirectories(t *testing.T) {
	// 1. HAPPY PATH: Criação bem-sucedida em local temporário
	t.Run("HappyPath_CreatesAllDirectories", func(t *testing.T) {
		tempRoot := filepath.Join(os.TempDir(), "vectora_test_happy")
		_ = os.RemoveAll(tempRoot)
		defer os.RemoveAll(tempRoot)

		paths := config.VectoraPaths{
			Root:   tempRoot,
			Bin:    filepath.Join(tempRoot, "bin"),
			DB:     filepath.Join(tempRoot, "db"),
			Models: filepath.Join(tempRoot, "models", "qwen"),
			Data:   filepath.Join(tempRoot, "models", "data"),
			Docs:   filepath.Join(tempRoot, "docs"),
		}

		if err := paths.EnsureDirectories(); err != nil {
			t.Fatalf("❌ Falha no Happy Path: %v", err)
		}

		// Validar existência
		dirs := []string{paths.Bin, paths.DB, paths.Models, paths.Data, paths.Docs}
		for _, d := range dirs {
			if _, err := os.Stat(d); os.IsNotExist(err) {
				t.Errorf("📁 Diretório esperado não foi criado: %s", d)
			}
		}
	})

	// 2. EDGE CASE: Idempotência (executar duas vezes não causa erro)
	t.Run("EdgeCase_Idempotency", func(t *testing.T) {
		tempRoot := filepath.Join(os.TempDir(), "vectora_test_edge")
		_ = os.RemoveAll(tempRoot)
		defer os.RemoveAll(tempRoot)

		paths := config.VectoraPaths{
			Root:   tempRoot,
			Bin:    filepath.Join(tempRoot, "bin"),
			DB:     filepath.Join(tempRoot, "db"),
			Models: filepath.Join(tempRoot, "models", "qwen"),
			Data:   filepath.Join(tempRoot, "models", "data"),
			Docs:   filepath.Join(tempRoot, "docs"),
		}

		// Primeira execução
		if err := paths.EnsureDirectories(); err != nil {
			t.Fatalf("❌ Falha na primeira execução: %v", err)
		}

		// Segunda execução (deve passar sem erros)
		if err := paths.EnsureDirectories(); err != nil {
			t.Errorf("❌ Falha na segunda execução (idempotência): %v", err)
		}
	})

	// 3. NEGATIVE: Tentativa de criação em local proibido/inválido
	t.Run("Negative_InvalidPath", func(t *testing.T) {
		// Tentaremos um caminho que não pode ser criado (ex: diretório com nome de arquivo existente)
		tempFile := filepath.Join(os.TempDir(), "vectora_test_conflict.txt")
		_ = os.WriteFile(tempFile, []byte("conflito"), 0644)
		defer os.Remove(tempFile)

		paths := config.VectoraPaths{
			Root: tempFile,
			Bin:  filepath.Join(tempFile, "sub"), // Falhará porque root é um arquivo
		}

		err := paths.EnsureDirectories()
		if err == nil {
			t.Error("❌ Esperava erro ao tentar criar subdiretório em um arquivo existente, mas obteve nil")
		} else {
			t.Logf("✅ Erro esperado capturado: %v", err)
		}
	})
}
