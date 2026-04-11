package llm

import (
	"strings"
	"testing"
)

func TestFamilyFromModel(t *testing.T) {
	tests := []struct {
		model  string
		family string
	}{
		// OpenRouter "provider/model" format
		{"openai/gpt-5.4-pro", "openai"},
		{"anthropic/claude-4.6-sonnet", "anthropic"},
		{"google/gemini-3.1-pro-preview", "google"},
		{"qwen/qwen3.6-plus", "qwen"},
		{"meta-llama/llama-4-70b", "meta-llama"},
		{"meta-llama/muse-spark", "meta-llama"},
		{"microsoft/phi-4-medium", "microsoft"},
		{"deepseek/deepseek-v3.2", "deepseek"},
		{"mistralai/mistral-large-3", "mistralai"},
		{"x-ai/grok-4.20", "x-ai"},
		{"zhipuai/glm-5.1", "zhipuai"},
		// Plain model name format
		{"gpt-5.4-pro", "openai"},
		{"claude-4.6-sonnet", "anthropic"},
		{"gemini-3.1-pro-preview", "google"},
		{"gemma-4-31b", "google"},
		{"qwen3.6-plus", "qwen"},
		{"llama-4-70b", "meta-llama"},
		{"muse-spark", "meta-llama"},
		{"phi-4-mini", "microsoft"},
		{"deepseek-v3.2", "deepseek"},
		{"mistral-large-3", "mistralai"},
		{"grok-4.20", "x-ai"},
		{"glm-5.1", "zhipuai"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := FamilyFromModel(tt.model)
			if got != tt.family {
				t.Errorf("FamilyFromModel(%q) = %q, want %q", tt.model, got, tt.family)
			}
		})
	}
}

func TestEmbeddingModelForGateway(t *testing.T) {
	tests := []struct {
		model         string
		wantEmbedding string
	}{
		// Qwen → native embedding
		{"qwen/qwen3.6-plus", "qwen3-embedding-8b"},
		{"qwen3.6-plus", "qwen3-embedding-8b"},
		// OpenAI → large embedding
		{"openai/gpt-5.4-pro", "text-embedding-3-large"},
		{"gpt-5.4-pro", "text-embedding-3-large"},
		// All other 8 families → text-embedding-3-large (no gateway native embedding)
		{"anthropic/claude-4.6-sonnet", "text-embedding-3-large"},
		{"google/gemini-3.1-pro-preview", "text-embedding-3-large"},
		{"meta-llama/llama-4-70b", "text-embedding-3-large"},
		{"microsoft/phi-4-medium", "text-embedding-3-large"},
		{"deepseek/deepseek-v3.2", "text-embedding-3-large"},
		{"mistralai/mistral-large-3", "text-embedding-3-large"},
		{"x-ai/grok-4.20", "text-embedding-3-large"},
		{"zhipuai/glm-5.1", "text-embedding-3-large"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := EmbeddingModelForGateway(tt.model)
			if got != tt.wantEmbedding {
				t.Errorf("EmbeddingModelForGateway(%q) = %q, want %q", tt.model, got, tt.wantEmbedding)
			}
		})
	}
}

func TestGatewayModelCatalog_Coverage(t *testing.T) {
	// Ensure all 10 LLM families from AGENTS.md are represented in the catalog.
	requiredFamilies := []string{
		"openai", "anthropic", "google", "qwen",
		"meta-llama", "microsoft", "deepseek",
		"mistralai", "x-ai", "zhipuai",
	}

	for _, family := range requiredFamilies {
		models, ok := GatewayModelCatalog[family]
		if !ok {
			t.Errorf("family %q missing from GatewayModelCatalog", family)
			continue
		}
		if len(models) == 0 {
			t.Errorf("family %q has no models in GatewayModelCatalog", family)
		}
	}
}

func TestAllGatewayModels(t *testing.T) {
	all := AllGatewayModels()
	if len(all) == 0 {
		t.Fatal("AllGatewayModels() returned empty slice")
	}

	// Each model must use "provider/model" format.
	for _, m := range all {
		if !strings.Contains(m, "/") {
			t.Errorf("model %q does not use provider/model format", m)
		}
	}
}

func TestGatewayModelsForProvider(t *testing.T) {
	// OpenRouter and Anannas return the full catalog.
	for _, gw := range []string{"openrouter", "anannas"} {
		models := GatewayModelsForProvider(gw)
		if len(models) == 0 {
			t.Errorf("GatewayModelsForProvider(%q) returned empty slice", gw)
		}
	}

	// Specific family returns only that family's models.
	anthropicModels := GatewayModelsForProvider("anthropic")
	for _, m := range anthropicModels {
		if !strings.HasPrefix(m, "anthropic/") {
			t.Errorf("expected anthropic model, got %q", m)
		}
	}
}
