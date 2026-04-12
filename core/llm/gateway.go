package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
)

// GatewayProvider is a flexible provider that uses the OpenAI SDK
// to connect to arbitrary gateways (OpenRouter, Anannas, etc.)
// It supports all 10 LLM families catalogued in AGENTS.md (Standard April 2026).
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
	// Try the live /models endpoint first — OpenRouter and Anannas both support it,
	// returning the full real-time catalog of available models.
	models, err := p.OpenAIProvider.ListModels(ctx)
	if err == nil && len(models) > 0 {
		return models, nil
	}
	// Fallback: return the static catalog for all 10 LLM families (AGENTS.md April 2026).
	return GatewayModelsForProvider(p.providerName), nil
}

// Embed overrides OpenAIProvider.Embed to detect the correct embedding model for all
// 10 LLM families, supporting both "provider/model" (OpenRouter) and plain model names.
//
// Family detection examples:
//
//	"anthropic/claude-4.6-sonnet" → text-embedding-3-large (no native gateway embedding)
//	"google/gemini-3.1-pro"       → text-embedding-3-large (Gemini not accessible via gateway)
//	"qwen/qwen3.6-plus"           → qwen3-embedding-8b     (native Qwen embedding)
//	"openai/gpt-5.4-pro"          → text-embedding-3-large
//	"meta-llama/llama-4-70b"      → text-embedding-3-large (no native embedding)
//	"microsoft/phi-4-medium"      → text-embedding-3-large
//	"deepseek/deepseek-v3.2"      → text-embedding-3-large
//	"mistralai/mistral-large-3"   → text-embedding-3-large
//	"x-ai/grok-4.20"              → text-embedding-3-large
//	"zhipuai/glm-5.1"             → text-embedding-3-large
func (p *GatewayProvider) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	// EmbeddingModelForGateway uses the shared family-detection logic from gateway_models.go.
	embeddingModel := EmbeddingModelForGateway(model)
	if embeddingModel == "" {
		return nil, fmt.Errorf("no native embedding for %s family, router should fallback", FamilyFromModel(model))
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
