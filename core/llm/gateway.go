package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
)

// GatewayProvider is a flexible provider that uses the OpenAI SDK
// to connect to arbitrary gateways (OpenRouter, Anannas, etc.)
type GatewayProvider struct {
	*OpenAIProvider
	providerName string
}

func NewGatewayProvider(apiKey string, baseURL string, name string) *GatewayProvider {
	return &GatewayProvider{
		OpenAIProvider: NewOpenAIProvider(apiKey, baseURL, name),
		providerName:   name,
	}
}

func (p *GatewayProvider) Name() string {
	return p.providerName
}

func (p *GatewayProvider) ListModels(ctx context.Context) ([]string, error) {
	// Re-uses OpenAIProvider's ListModels
	return p.OpenAIProvider.ListModels(ctx)
}

// Embed is overridden if we want to add family-based embedding logic here,
// but the plan says the ROUTER handles this logic.
// However, the GatewayProvider can provide hints.
func (p *GatewayProvider) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	// If the model name hints at a specific family, we can choose the embedding model accordingly.
	embeddingModel := "text-embedding-3-small" // Default fallback

	lowerModel := strings.ToLower(model)
	if strings.Contains(lowerModel, "qwen") {
		embeddingModel = "qwen3-embedding-8b"
	} else if strings.Contains(lowerModel, "gemini") || strings.Contains(lowerModel, "gemma") {
		embeddingModel = "gemini-embedding-2-preview"
	} else if strings.Contains(lowerModel, "gpt") || strings.Contains(lowerModel, "openai") {
		embeddingModel = "text-embedding-3-large"
	} else if strings.Contains(lowerModel, "llama") || strings.Contains(lowerModel, "phi") ||
		strings.Contains(lowerModel, "mistral") || strings.Contains(lowerModel, "deepseek") ||
		strings.Contains(lowerModel, "grok") || strings.Contains(lowerModel, "glm") ||
		strings.Contains(lowerModel, "muse") {
		embeddingModel = "text-embedding-3-large"
	}

	params := openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(embeddingModel),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(input),
		},
	}

	resp, err := p.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) > 0 {
		vec := make([]float32, len(resp.Data[0].Embedding))
		for i, v := range resp.Data[0].Embedding {
			vec[i] = float32(v)
		}
		return vec, nil
	}

	return nil, fmt.Errorf("no embedding returned")
}
