package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/austinfhunter/voyageai"
)

type VoyageProvider struct {
	apiKey string
	client *voyageai.VoyageClient
}

func NewVoyageProvider(ctx context.Context, apiKey string) (*VoyageProvider, error) {
	if apiKey == "" {
		return nil, errors.New("voyage_api_key_required: Voyage API key was not provided")
	}

	client := voyageai.NewClient(&voyageai.VoyageClientOpts{
		Key: apiKey,
	})

	return &VoyageProvider{
		apiKey: apiKey,
		client: client,
	}, nil
}

func (p *VoyageProvider) ListModels(ctx context.Context) ([]string, error) {
	// Voyage é especializado em embeddings — modelos fixos
	return []string{
		"voyage-3",
		"voyage-3-lite",
		"voyage-code-3",
		"voyage-3-large",
	}, nil
}

func (p *VoyageProvider) Name() string {
	return "voyage"
}

func (p *VoyageProvider) IsConfigured() bool {
	return p.apiKey != ""
}

func (p *VoyageProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	return CompletionResponse{}, errors.New("voyage_only_supports_embeddings")
}

func (p *VoyageProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, <-chan error) {
	errChan := make(chan error, 1)
	errChan <- errors.New("voyage_only_supports_embeddings")
	close(errChan)
	return nil, errChan
}

func (p *VoyageProvider) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	// Voyage SDK supports batch embedding.
	resp, err := p.client.Embed([]string{input}, voyageai.ModelVoyageCode3, &voyageai.EmbeddingRequestOpts{
		InputType: voyageai.Opt("document"),
	})
	if err != nil {
		return nil, fmt.Errorf("voyage_sdk_error: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("voyage_no_embeddings_returned")
	}

	return resp.Data[0].Embedding, nil
}
