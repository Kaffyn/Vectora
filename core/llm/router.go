package llm

import (
	"context"
	"fmt"
)

type Router struct {
	providers       map[string]Provider
	defaultProvider string
}

func NewRouter() *Router {
	return &Router{
		providers: make(map[string]Provider),
	}
}

func (r *Router) RegisterProvider(name string, p Provider, asDefault bool) {
	r.providers[name] = p
	if asDefault || r.defaultProvider == "" {
		r.defaultProvider = name
	}
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
	if p, ok := r.providers[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("provider %s not found", name)
}

func (r *Router) GetDefault() Provider {
	if r.defaultProvider == "" {
		return nil
	}
	return r.providers[r.defaultProvider]
}

func (r *Router) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	p := r.GetDefault()
	if p == nil {
		return CompletionResponse{}, fmt.Errorf("no LLM provider configured")
	}
	return p.Complete(ctx, req)
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
