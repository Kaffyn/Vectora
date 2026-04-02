package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

type LlamaCPPProvider struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewLlamaCPPProvider(baseURL string) *LlamaCPPProvider {
	return &LlamaCPPProvider{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type embeddingRequest struct {
	Content string `json:"content"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (p *LlamaCPPProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	reqBody, err := json.Marshal(map[string]string{"content": text})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/embedding", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to llama.cpp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llama.cpp returned status: %s", resp.Status)
	}

	var res struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res.Embedding, nil
}

func (p *LlamaCPPProvider) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, 0, len(texts))
	for _, text := range texts {
		emb, err := p.EmbedQuery(ctx, text)
		if err != nil {
			return nil, err
		}
		results = append(results, emb)
	}
	return results, nil
}

// Ensure LlamaCPPProvider implements domain.EmbeddingProvider
var _ domain.EmbeddingProvider = (*LlamaCPPProvider)(nil)
