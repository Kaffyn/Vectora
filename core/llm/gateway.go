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

// Embed overrides OpenAIProvider.Embed to add family-based embedding model detection
// that understands the "provider/model" format used by OpenRouter and Anannas.
//
// Examples:
//
//	"anthropic/claude-4.6-sonnet" → text-embedding-3-large (Claude has no native embedding)
//	"google/gemini-3.1-pro"       → text-embedding-3-large (Gemini embedding not via gateway)
//	"qwen/qwen3.6-plus"           → qwen3-embedding-8b
//	"openai/gpt-5.4-pro"          → text-embedding-3-large

func (p *GatewayProvider) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	lowerModel := strings.ToLower(model)

	// Extract provider family from "provider/model" format (OpenRouter, Anannas style).
	family := lowerModel
	subModel := lowerModel
	if idx := strings.Index(lowerModel, "/"); idx != -1 {
		family = lowerModel[:idx]     // e.g. "anthropic", "google", "qwen", "openai"
		subModel = lowerModel[idx+1:] // e.g. "claude-4.6-sonnet", "gemini-3.1-pro"
	}

	// Choose embedding model based on detected family.
	// Qwen has its own embedding model; everything else uses OpenAI's.
	// Note: Claude, Gemini, LLaMA, Phi, Mistral, etc. have no native embedding accessible
	// via gateways, so we fall back to text-embedding-3-large.
	var embeddingModel string
	if strings.Contains(family, "qwen") || strings.Contains(subModel, "qwen") {
		embeddingModel = "qwen3-embedding-8b"
	} else {
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
