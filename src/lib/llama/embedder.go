package llama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

type LlamaEmbedder struct {
	serverURL string
	client    *http.Client
}

func NewLlamaEmbedder(serverURL string) *LlamaEmbedder {
	return &LlamaEmbedder{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type llamaEmbeddingRequest struct {
	Content string `json:"content"`
}

type llamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (e *LlamaEmbedder) Generate(ctx context.Context, text string) ([]float32, error) {
	reqBody, _ := json.Marshal(llamaEmbeddingRequest{Content: text})

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/embedding", e.serverURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llama server access error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llama server returned status: %s", resp.Status)
	}

	var embResp llamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	return embResp.Embedding, nil
}

// Ensure implementation
func (e *LlamaEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return e.Generate(ctx, text)
}

func (e *LlamaEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, 0, len(texts))
	for _, text := range texts {
		emb, err := e.Generate(ctx, text)
		if err != nil {
			return nil, err
		}
		results = append(results, emb)
	}
	return results, nil
}

var _ domain.EmbeddingProvider = (*LlamaEmbedder)(nil)
