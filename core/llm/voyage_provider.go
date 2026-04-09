package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type VoyageProvider struct {
	apiKey string
	model  string
}

func NewVoyageProvider(ctx context.Context, apiKey string) (*VoyageProvider, error) {
	if apiKey == "" {
		return nil, errors.New("voyage_api_key_required: Voyage API key was not provided")
	}

	return &VoyageProvider{
		apiKey: apiKey,
		model:  "voyage-code-3",
	}, nil
}

func (p *VoyageProvider) Name() string {
	return "voyage"
}

func (p *VoyageProvider) IsConfigured() bool {
	return p.apiKey != ""
}

func (p *VoyageProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	return CompletionResponse{}, errors.New("voyage_only_supports_embeddings: Use Gemini or Claude for reasoning")
}

func (p *VoyageProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	url := "https://api.voyageai.com/v1/embeddings"
	
	requestBody, err := json.Marshal(map[string]interface{}{
		"input": []string{input},
		"model": p.model,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("voyage API error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, errors.New("voyage_no_embeddings_returned")
	}

	return result.Data[0].Embedding, nil
}
