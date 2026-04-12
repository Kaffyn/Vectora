package llm

import (
	"context"
	"fmt"
)

type Router struct {
	Providers        map[string]Provider
	DefaultProvider  string
	FallbackProvider string
	FallbackModels   map[string]string
}

func NewRouter() *Router {
	return &Router{
		Providers:      make(map[string]Provider),
		FallbackModels: make(map[string]string),
	}
}

func (r *Router) RegisterProvider(name string, p Provider, asDefault bool) {
	r.Providers[name] = p
	if asDefault || r.DefaultProvider == "" {
		r.DefaultProvider = name
	}
}

func (r *Router) SetFallbackProvider(name string) {
	r.FallbackProvider = name
}

func (r *Router) SetFallbackModel(providerName, model string) {
	r.FallbackModels[providerName] = model
}

func (r *Router) GetFallbackModel(providerName string) string {
	return r.FallbackModels[providerName]
}

func (r *Router) ListModels(ctx context.Context, providerName string) ([]string, error) {
	if providerName == "" {
		p := r.GetDefault()
		if p == nil {
			return nil, fmt.Errorf("no default provider configured")
		}
		return p.ListModels(ctx)
	}

	p, err := r.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return p.ListModels(ctx)
}

func (r *Router) GetProvider(name string) (Provider, error) {
	if p, ok := r.Providers[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("provider %s not found", name)
}

func (r *Router) GetDefault() Provider {
	if r.DefaultProvider == "" {
		return nil
	}
	return r.Providers[r.DefaultProvider]
}

func (r *Router) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	p := r.GetDefault()
	if p == nil {
		return CompletionResponse{}, fmt.Errorf("no LLM provider configured")
	}

	res, err := p.Complete(ctx, req)
	if err != nil && r.FallbackProvider != "" && r.FallbackProvider != r.DefaultProvider {
		// Try fallback provider (e.g., Gemini)
		fp, ok := r.Providers[r.FallbackProvider]
		if ok && fp.IsConfigured() {
			// Optional: log that we are falling back
			return fp.Complete(ctx, req)
		}
	}

	return res, err
}

func (r *Router) Embed(ctx context.Context, input string, model string) ([]float32, error) {
	// 1. Family-based routing — map the model to the best available native provider.
	//    FamilyFromModel understands both "provider/model" (OpenRouter) and plain names.
	//    All 10 LLM families from AGENTS.md are handled.
	family := FamilyFromModel(model)

	// Map family → provider name registered in the router.
	familyToProvider := map[string]string{
		"google":    "gemini",
		"openai":    "openai",
		"qwen":      "qwen",
		"anthropic": "claude",
		// The remaining families have no dedicated native provider in the router;
		// they are accessed via gateway (openrouter/anannas) or the default provider.
		"meta-llama": "openrouter",
		"microsoft":  "openrouter",
		"deepseek":   "openrouter",
		"mistralai":  "openrouter",
		"x-ai":       "openrouter",
		"zhipuai":    "openrouter",
	}

	if providerName, ok := familyToProvider[family]; ok {
		if p, err := r.GetProvider(providerName); err == nil && p.IsConfigured() {
			vec, err := p.Embed(ctx, input, model)
			if err == nil {
				return vec, nil
			}
			// Provider found but embedding failed — continue to fallback chain.
		}
	}

	// 2. Default provider embedding.
	if p := r.GetDefault(); p != nil {
		vec, err := p.Embed(ctx, input, model)
		if err == nil {
			return vec, nil
		}
	}

	// 3. Global fallback: Voyage AI (high-quality code embeddings, voyage-3-large).
	if vp, err := r.GetProvider("voyage"); err == nil && vp.IsConfigured() {
		return vp.Embed(ctx, input, model)
	}

	return nil, fmt.Errorf("no suitable embedding provider found for model %s (all fallbacks exhausted)", model)
}

func (r *Router) IsConfigured() bool {
	return r.GetDefault() != nil && r.GetDefault().IsConfigured()
}
