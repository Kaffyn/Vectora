package gemini

import (
	"context"
	"fmt"
	"os"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiEmbedder struct {
	modelName string
}

func NewGeminiEmbedder(modelName string) *GeminiEmbedder {
	if modelName == "" {
		modelName = "text-embedding-004"
	}
	return &GeminiEmbedder{modelName: modelName}
}

func (e *GeminiEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY não configurada")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.EmbeddingModel(e.modelName)
	res, err := model.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, err
	}

	return res.Embedding.Values, nil
}

func (e *GeminiEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY não configurada")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.EmbeddingModel(e.modelName)
	batch := model.NewBatch()
	for _, text := range texts {
		batch.AddContent(genai.Text(text))
	}

	res, err := model.BatchEmbedContents(ctx, batch)
	if err != nil {
		return nil, err
	}

	results := make([][]float32, 0, len(res.Embeddings))
	for _, emb := range res.Embeddings {
		results = append(results, emb.Values)
	}

	return results, nil
}

var _ domain.EmbeddingProvider = (*GeminiEmbedder)(nil)
