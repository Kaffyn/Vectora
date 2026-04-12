package llm

import "strings"

// ProviderModels defines the official recommended model IDs for each provider family.
// Reference: AGENTS.md (April 2026 Standards)
var ProviderModels = map[string][]string{
	"gemini":     {"gemini-3.1-pro-preview", "gemini-3-flash-preview", "gemma-4-31b"},
	"claude":     {"claude-4.6-sonnet", "claude-4.6-opus", "claude-4.5-haiku"},
	"openai":     {"gpt-5.4-pro", "gpt-5.4-mini", "gpt-5-o1"},
	"openrouter": {"google/gemini-3.1-pro", "anthropic/claude-4.6-sonnet", "meta-llama/llama-4-70b"},
	"anannas":    {"anthropic/claude-4.6-sonnet", "google/gemini-3.1-pro", "openai/gpt-5.4-pro"},
	"deepseek":   {"deepseek-v3.2", "deepseek-v3.2-speciale"},
	"mistral":    {"mistral-large-3", "mistral-small-4"},
	"grok":       {"grok-4.20", "grok-4.1"},
	"zhipu":      {"glm-5.1", "glm-5-flash"},
}

// FamilyFromModel attempts to deduce the provider family from a model name/ID.
// Supports both "provider/model" (OpenRouter) and plain name formats.
func FamilyFromModel(model string) string {
	lower := strings.ToLower(model)

	// Explicit "provider/model" format (OpenRouter/Anannas).
	if idx := strings.Index(lower, "/"); idx != -1 {
		return lower[:idx]
	}

	// Plain model name — infer family from well-known keywords.
	switch {
	case strings.Contains(lower, "gpt") || strings.Contains(lower, "o1") || strings.Contains(lower, "o3") || strings.Contains(lower, "openai"):
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
