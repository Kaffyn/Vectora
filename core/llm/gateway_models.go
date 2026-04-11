package llm

import "strings"

// GatewayModelCatalog lists all known models available via gateways (OpenRouter, Anannas).
// Source of truth: AGENTS.md (Standard April 2026).
// Organized by the provider prefix used on OpenRouter (e.g. "openai/gpt-5.4-pro").
//
// Usage:
//   - GatewayProvider.ListModels() returns this full catalog.
//   - EmbeddingModelForGateway() selects the right embedding model for any LLM.
var GatewayModelCatalog = map[string][]string{
	// 1. OpenAI — GPT-5.4 series (launched March 2026, 1.05M context)
	"openai": {
		"openai/gpt-5.4-pro",
		"openai/gpt-5.4-mini",
		"openai/gpt-5.4-nano",
		"openai/gpt-5-o1",
	},

	// 2. Anthropic — Claude 4.6 series (launched February 2026, 1M context)
	"anthropic": {
		"anthropic/claude-4.6-sonnet",
		"anthropic/claude-4.6-opus",
		"anthropic/claude-4.5-haiku",
		"anthropic/claude-4.5-sonnet",
		"anthropic/claude-4.5-opus",
	},

	// 3. Google — Gemini 3.1 & 3.0 (launched February 2026)
	"google": {
		"google/gemini-3.1-pro-preview",
		"google/gemini-3-flash-preview",
		"google/gemma-4-31b",
		"google/gemini-2.5-pro",
		"google/gemini-2.5-flash",
	},

	// 4. Alibaba — Qwen 3.6 series (launched April 2026, 1M context, reasoning always-on)
	"qwen": {
		"qwen/qwen3.6-plus",
		"qwen/qwen3.6-turbo",
		"qwen/qwen3.5-omni",
		"qwen/qwen-max",
	},

	// 5. Meta — Muse & Llama 4 (Muse Spark launched April 2026)
	"meta-llama": {
		"meta-llama/llama-4-70b",
		"meta-llama/llama-4-maverick",
		"meta-llama/llama-4-scout",
		"meta-llama/muse-spark",
	},

	// 6. Microsoft — Phi-4 series (Phi-4-Reasoning-Vision launched March 2026)
	"microsoft": {
		"microsoft/phi-4-mini",
		"microsoft/phi-4-medium",
		"microsoft/phi-4-reasoning-vision-15b",
	},

	// 7. DeepSeek — V3 series (V3.2 launched December 2025, 671B MoE)
	"deepseek": {
		"deepseek/deepseek-v3.2",
		"deepseek/deepseek-v3.2-speciale",
	},

	// 8. Mistral AI — Cloud & Open series (Mistral Small 4 launched March 2026)
	"mistralai": {
		"mistralai/mistral-large-3",
		"mistralai/mistral-small-4",
	},

	// 9. xAI — Grok 4 series (launched 2026)
	"x-ai": {
		"x-ai/grok-4.1",
		"x-ai/grok-4.20",
	},

	// 10. Zhipu AI (Z.ai) — GLM-5 series (GLM-5.1 launched April 2026, MIT License)
	"zhipuai": {
		"zhipuai/glm-5.1",
		"zhipuai/glm-5-flash",
	},
}

// AllGatewayModels returns a flat slice of every known gateway model.
// Suitable for use as a static fallback in ListModels().
func AllGatewayModels() []string {
	var all []string
	for _, models := range GatewayModelCatalog {
		all = append(all, models...)
	}
	return all
}

// GatewayModelsForProvider returns models whose prefix matches the given provider name.
// For example, GatewayModelsForProvider("openrouter") returns all models.
// GatewayModelsForProvider("anthropic") returns only Anthropic models.
func GatewayModelsForProvider(providerName string) []string {
	lower := strings.ToLower(providerName)

	// OpenRouter and Anannas expose ALL providers.
	if lower == "openrouter" || lower == "anannas" || lower == "openai" {
		return AllGatewayModels()
	}

	// Otherwise match catalog key prefix.
	if models, ok := GatewayModelCatalog[lower]; ok {
		return models
	}

	return AllGatewayModels()
}

// EmbeddingModelForGateway selects the appropriate embedding model for a given LLM
// when using a gateway. Supports both "provider/model" format (OpenRouter) and plain
// model names.
//
// Decision logic (from gateway-support.md Family Detection):
//
//  1. Qwen family → qwen3-embedding-8b (native Qwen embedding)
//  2. OpenAI/GPT family → text-embedding-3-large (native OpenAI embedding)
//  3. All others (Claude, Gemini, LLaMA, Phi, Mistral, DeepSeek, Grok, GLM, Muse)
//     → text-embedding-3-large as fallback (Voyage AI is handled at router level)
func EmbeddingModelForGateway(model string) string {
	lower := strings.ToLower(model)

	// Extract provider prefix from "provider/model" format.
	family := lower
	subModel := lower
	if idx := strings.Index(lower, "/"); idx != -1 {
		family = lower[:idx]
		subModel = lower[idx+1:]
	}

	// Qwen: has its own high-quality embedding model.
	if family == "qwen" || strings.Contains(subModel, "qwen") || strings.Contains(family, "qwen") {
		return "qwen3-embedding-8b"
	}

	// OpenAI: use large embedding model.
	if family == "openai" || strings.Contains(subModel, "gpt") || strings.Contains(family, "gpt") {
		return "text-embedding-3-large"
	}

	// All other families (Anthropic, Google, Meta, Microsoft, DeepSeek, Mistral, xAI, Zhipu):
	// No native embedding accessible via gateway → use OpenAI text-embedding-3-large.
	// The router will further fall back to Voyage AI if this also fails.
	return "text-embedding-3-large"
}

// FamilyFromModel extracts the provider family name from a model string.
// Supports both "provider/model" and plain model name formats.
//
// Examples:
//
//	"anthropic/claude-4.6-sonnet" → "anthropic"
//	"google/gemini-3.1-pro"       → "google"
//	"gemini-3.1-pro-preview"      → "google"
//	"gpt-5.4-pro"                 → "openai"
//	"claude-4.6-sonnet"           → "anthropic"
//	"llama-4-70b"                 → "meta-llama"
//	"mistral-large-3"             → "mistralai"
func FamilyFromModel(model string) string {
	lower := strings.ToLower(model)

	// Explicit "provider/model" format (OpenRouter/Anannas).
	if idx := strings.Index(lower, "/"); idx != -1 {
		return lower[:idx]
	}

	// Plain model name — infer family from well-known keywords.
	switch {
	case strings.Contains(lower, "gpt") || strings.Contains(lower, "o1") || strings.Contains(lower, "o3"):
		return "openai"
	case strings.Contains(lower, "claude"):
		return "anthropic"
	case strings.Contains(lower, "gemini") || strings.Contains(lower, "gemma"):
		return "google"
	case strings.Contains(lower, "qwen"):
		return "qwen"
	case strings.Contains(lower, "llama") || strings.Contains(lower, "muse"):
		return "meta-llama"
	case strings.Contains(lower, "phi"):
		return "microsoft"
	case strings.Contains(lower, "deepseek"):
		return "deepseek"
	case strings.Contains(lower, "mistral"):
		return "mistralai"
	case strings.Contains(lower, "grok"):
		return "x-ai"
	case strings.Contains(lower, "glm"):
		return "zhipuai"
	case strings.Contains(lower, "voyage"):
		return "voyage"
	default:
		return "unknown"
	}
}
