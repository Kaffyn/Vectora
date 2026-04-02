package llama_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kaffyn/Vectora/src/lib/llama"
)

func TestLlamaEmbedder_300Percent(t *testing.T) {
	// 1. HAPPY PATH: Mock llama-server response
	t.Run("HappyPath_GenerateEmbedding", func(t *testing.T) {
		expectedVector := []float32{0.1, 0.2, 0.3}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/embedding" {
				t.Errorf("Path esperado /embedding, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"embedding": expectedVector,
			})
		}))
		defer server.Close()

		embedder := llama.NewLlamaEmbedder(server.URL)
		got, err := embedder.Generate(context.Background(), "test text")

		if err != nil {
			t.Errorf("Generate falhou: %v", err)
		}

		if len(got) != 3 || got[0] != 0.1 {
			t.Errorf("Vetor incorreto, got %v", got)
		}
	})

	// 2. NEGATIVE: Server Error (500)
	t.Run("Negative_ServerError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		embedder := llama.NewLlamaEmbedder(server.URL)
		_, err := embedder.Generate(context.Background(), "test text")

		if err == nil {
			t.Error("Esperava erro para status 500, got nil")
		}
	})

	// 3. EDGE CASE: Empty text
	t.Run("EdgeCase_EmptyText", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"embedding": []float32{},
			})
		}))
		defer server.Close()

		embedder := llama.NewLlamaEmbedder(server.URL)
		got, err := embedder.Generate(context.Background(), "")
		if err != nil {
			t.Errorf("Empty text fail: %v", err)
		}
		if len(got) != 0 {
			t.Error("Expected empty vector for empty input")
		}
	})
}
