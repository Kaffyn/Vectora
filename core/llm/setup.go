package llm

import (
	"context"
	"github.com/Kaffyn/Vectora/core/infra"
)

// SetupRouter initializes a router with all available providers based on configuration.
func SetupRouter(ctx context.Context, cfg *infra.Config, prefs *infra.UserPreferences) *Router {
	router := NewRouter()
	router.SetFallbackProvider("gemini")
	if cfg.DefaultFallbackProvider != "" {
		router.SetFallbackProvider(cfg.DefaultFallbackProvider)
	}

	// Default fallback models (April 2026 Standards)
	router.SetFallbackModel("gemini", "gemini-3-flash-preview")
	router.SetFallbackModel("claude", "claude-4.5-haiku")
	router.SetFallbackModel("openai", "gpt-5.4-mini")
	router.SetFallbackModel("qwen", "qwen3")

	// Set fallback models from config (overrides defaults)
	if cfg.GeminiFallbackModel != "" {
		router.SetFallbackModel("gemini", cfg.GeminiFallbackModel)
	}
	if cfg.ClaudeFallbackModel != "" {
		router.SetFallbackModel("claude", cfg.ClaudeFallbackModel)
	}
	if cfg.OpenAIFallbackModel != "" {
		router.SetFallbackModel("openai", cfg.OpenAIFallbackModel)
	}
	if cfg.QwenFallbackModel != "" {
		router.SetFallbackModel("qwen", cfg.QwenFallbackModel)
	}

	// Use preference for default provider if available
	defaultProv := prefs.DefaultProvider

	// Se há um gateway ativo, ele se torna o provider padrão para completion
	// Os providers nativos ainda são registrados para embeddings
	if cfg.ActiveGateway == "openrouter" && cfg.OpenRouterAPIKey != "" {
		defaultProv = "openrouter"
	} else if cfg.ActiveGateway == "anannas" && cfg.AnannasAPIKey != "" {
		defaultProv = "anannas"
	}

	// Register Native Providers
	if cfg.GeminiAPIKey != "" {
		p, _ := NewGeminiProvider(ctx, cfg.GeminiAPIKey)
		router.RegisterProvider("gemini", p, defaultProv == "gemini")
	}
	if cfg.ClaudeAPIKey != "" {
		p, _ := NewClaudeProvider(ctx, cfg.ClaudeAPIKey)
		router.RegisterProvider("claude", p, defaultProv == "claude")
	}
	if cfg.OpenAIAPIKey != "" {
		p := NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, "openai")
		router.RegisterProvider("openai", p, defaultProv == "openai")
	}
	if cfg.QwenAPIKey != "" {
		p := NewOpenAIProvider(cfg.QwenAPIKey, cfg.QwenBaseURL, "qwen")
		router.RegisterProvider("qwen", p, defaultProv == "qwen")
	}

	// Register Embedding-Specialized Providers
	if cfg.VoyageAPIKey != "" {
		p, err := NewVoyageProvider(ctx, cfg.VoyageAPIKey)
		if err == nil {
			router.RegisterProvider("voyage", p, defaultProv == "voyage")
		}
	}

	// Register Gateway Providers
	if cfg.OpenRouterAPIKey != "" {
		p := NewGatewayProvider(cfg.OpenRouterAPIKey, "https://openrouter.ai/api/v1", "openrouter")
		router.RegisterProvider("openrouter", p, defaultProv == "openrouter")
	}
	if cfg.AnannasAPIKey != "" {
		p := NewGatewayProvider(cfg.AnannasAPIKey, "https://api.anannas.ai/v1", "anannas")
		router.RegisterProvider("anannas", p, defaultProv == "anannas")
	}

	return router
}
