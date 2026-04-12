package llm

import "strings"

// ModelInfo defines the identifiers for a specific LLM, mapping standard names
// to both native SDK IDs and gateway (provider/id) formats.
// Reference: AGENTS.md (April 2026 Standards).
type ModelInfo struct {
	NativeID  string // Exact ID for official SDK (e.g. "gemini-3.1-pro-preview")
	GatewayID string // Uniform ID for Gateways (e.g. "google/gemini-3.1-pro-preview")
	Family    string // Canonical provider family (e.g. "google", "openai")
}

// ModelRegistry is the single source of truth for all supported models in Vectora.
var ModelRegistry = map[string]ModelInfo{
	// 1. Google (Gemini 3.1 & 3.0)
	"gemini-3.1-pro": {
		NativeID:  "gemini-3.1-pro-preview",
		GatewayID: "google/gemini-3.1-pro-preview",
		Family:    "google",
	},
	"gemini-3-flash": {
		NativeID:  "gemini-3-flash-preview",
		GatewayID: "google/gemini-3-flash-preview",
		Family:    "google",
	},
	"gemma-4-31b": {
		NativeID:  "gemma-4-31b",
		GatewayID: "google/gemma-4-31b",
		Family:    "google",
	},

	// 2. Anthropic (Claude 4.6 Series)
	"claude-4.6-sonnet": {
		NativeID:  "claude-4.6-sonnet",
		GatewayID: "anthropic/claude-4.6-sonnet",
		Family:    "anthropic",
	},
	"claude-4.6-opus": {
		NativeID:  "claude-4.6-opus",
		GatewayID: "anthropic/claude-4.6-opus",
		Family:    "anthropic",
	},
	"claude-4.5-haiku": {
		NativeID:  "claude-4.5-haiku",
		GatewayID: "anthropic/claude-4.5-haiku",
		Family:    "anthropic",
	},

	// 3. OpenAI (GPT-5.4 Series)
	"gpt-5.4-pro": {
		NativeID:  "gpt-5.4-pro",
		GatewayID: "openai/gpt-5.4-pro",
		Family:    "openai",
	},
	"gpt-5.4-mini": {
		NativeID:  "gpt-5.4-mini",
		GatewayID: "openai/gpt-5.4-mini",
		Family:    "openai",
	},
	"gpt-5-o1": {
		NativeID:  "gpt-5-o1",
		GatewayID: "openai/gpt-5-o1",
		Family:    "openai",
	},

	// 4. Alibaba (Qwen 3.6 Series)
	"qwen3.6-plus": {
		NativeID:  "qwen3.6-plus",
		GatewayID: "qwen/qwen3.6-plus",
		Family:    "qwen",
	},
	"qwen3.6-turbo": {
		NativeID:  "qwen3.6-turbo",
		GatewayID: "qwen/qwen3.6-turbo",
		Family:    "qwen",
	},
	"qwen-max": {
		NativeID:  "qwen-max",
		GatewayID: "qwen/qwen-max",
		Family:    "qwen",
	},

	// 5. Meta (Muse & Llama 4)
	"llama-4-70b": {
		NativeID:  "llama-4-70b",
		GatewayID: "meta-llama/llama-4-70b",
		Family:    "meta-llama",
	},
	"muse-spark": {
		NativeID:  "muse-spark",
		GatewayID: "meta-llama/muse-spark",
		Family:    "meta-llama",
	},

	// 6. DeepSeek (V3.2 Series)
	"deepseek-v3.2": {
		NativeID:  "deepseek-v3.2",
		GatewayID: "deepseek/deepseek-v3.2",
		Family:    "deepseek",
	},

	// 7. Mistral (Mistral Small 4)
	"mistral-small-4": {
		NativeID:  "mistral-small-4",
		GatewayID: "mistralai/mistral-small-4",
		Family:    "mistralai",
	},

	// 8. Grok (Grok 4.20)
	"grok-4.20": {
		NativeID:  "grok-4.20",
		GatewayID: "x-ai/grok-4.20",
		Family:    "x-ai",
	},

	// 9. Zhipu (GLM-5.1)
	"glm-5.1": {
		NativeID:  "glm-5.1",
		GatewayID: "zhipuai/glm-5.1",
		Family:    "zhipuai",
	},
}

// FamilyEmbeddingRegistry maps LLM families to their native embedding models.
// Reference: AGENTS.md. Families not listed here (e.g. Anthropic, Meta, DeepSeek)
// will trigger the global router fallback to Voyage AI (voyage-3-large).
var FamilyEmbeddingRegistry = map[string]string{
	"google": "gemini-embedding-2-preview",
	"openai": "text-embedding-3-large",
	"qwen":   "qwen3-embedding-8b",
}

// GlobalAliases provide shorthand names for common models.
var GlobalAliases = map[string]string{
	"gemini": "gemini-3-flash",
	"flash":  "gemini-3-flash",
	"pro":    "gemini-3.1-pro",
	"sonnet": "claude-4.6-sonnet",
	"opus":   "claude-4.6-opus",
	"haiku":  "claude-4.5-haiku",
	"gpt5":   "gpt-5.4-pro",
	"mini":   "gpt-5.4-mini",
}

// ResolveModel resolves a model shorthand or ID to the requested format based on provider.
func ResolveModel(provider, model string) string {
	if model == "" {
		if provider == "gemini" {
			model = "gemini"
		} else if provider == "openrouter" || provider == "anannas" {
			model = "gemini-3.1-pro"
		} else {
			return model
		}
	}

	// 1. Resolve alias if exists
	if target, ok := GlobalAliases[model]; ok {
		model = target
	}

	// 2. Detect if we need Gateway ID (provider/id)
	useGateway := (provider == "openrouter" || provider == "anannas")

	// 3. Lookup in registry
	if info, ok := ModelRegistry[model]; ok {
		if useGateway {
			return info.GatewayID
		}
		return info.NativeID
	}

	// 4. Passthrough if not in registry
	return model
}

// FamilyFromModel deduces the provider family from a model name/ID.
func FamilyFromModel(model string) string {
	lower := strings.ToLower(model)

	// Check registry first (handles standard names)
	if info, ok := ModelRegistry[model]; ok {
		return info.Family
	}

	// Handle "provider/model" format used by gateways
	if idx := strings.Index(lower, "/"); idx != -1 {
		return lower[:idx]
	}

	// Fallback detection
	switch {
	case strings.Contains(lower, "gpt") || strings.Contains(lower, "o1"):
		return "openai"
	case strings.Contains(lower, "claude"):
		return "anthropic"
	case strings.Contains(lower, "gemini") || strings.Contains(lower, "gemma"):
		return "google"
	case strings.Contains(lower, "qwen"):
		return "qwen"
	case strings.Contains(lower, "llama"):
		return "meta-llama"
	case strings.Contains(lower, "voyage"):
		return "voyage"
	default:
		return "unknown"
	}
}

// ProviderModels maps families to their primary model IDs (registry keys).
var ProviderModels = map[string][]string{
	"gemini":     {"gemini-3.1-pro", "gemini-3-flash", "gemma-4-31b"},
	"claude":     {"claude-4.6-sonnet", "claude-4.6-opus", "claude-4.5-haiku"},
	"openai":     {"gpt-5.4-pro", "gpt-5.4-mini", "gpt-5-o1"},
	"openrouter": {"gemini-3.1-pro", "claude-4.6-sonnet", "llama-4-70b", "deepseek-v3.2"},
	"anannas":    {"claude-4.6-sonnet", "gemini-3.1-pro", "gpt-5.4-pro"},
	"deepseek":   {"deepseek-v3.2"},
	"mistral":    {"mistral-small-4"},
	"grok":       {"grok-4.20"},
	"zhipu":      {"glm-5.1"},
}

// KnownModels fallback lists for OpenAI/Qwen style providers.
var KnownModels = map[string][]string{
	"openai": {"gpt-5.4-pro", "gpt-5.4-mini", "gpt-5-o1"},
	"qwen":   {"qwen3.6-plus", "qwen3.6-turbo", "qwen-max"},
}
