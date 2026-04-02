package domain

import "context"

type EmbeddingProvider interface {
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
}

type LLMProvider interface {
	GenerateResponse(ctx context.Context, prompt string, context string) (string, error)
}
