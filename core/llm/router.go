package llm

import (
	"context"
	"fmt"
	"strings"
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
	// 1. Try family-based routing
	lowerModel := strings.ToLower(model)

	// Se o modelo for explicitamente de um provedor que conhecemos, tentamos o embedding dele primeiro
	var targetProvider string
	if strings.Contains(lowerModel, "gemini") {
		targetProvider = "gemini"
	} else if strings.Contains(lowerModel, "openai") || strings.Contains(lowerModel, "gpt") {
		targetProvider = "openai"
	} else if strings.Contains(lowerModel, "qwen") {
		targetProvider = "qwen"
	}

	if targetProvider != "" {
		if p, err := r.GetProvider(targetProvider); err == nil && p.IsConfigured() {
			vec, err := p.Embed(ctx, input, model)
			if err == nil {
				return vec, nil
			}
		}
	}

	// 2. Fallback: Use the default provider's embedding
	p := r.GetDefault()
	if p != nil {
		vec, err := p.Embed(ctx, input, model)
		if err == nil {
			return vec, nil
		}
	}

	// 3. Global Fallback: Voyage (High quality)
	if vp, err := r.GetProvider("voyage"); err == nil && vp.IsConfigured() {
		return vp.Embed(ctx, input, model)
	}

	return nil, fmt.Errorf("no suitable embedding provider found for model %s (fallback failed)", model)
}

func (r *Router) IsConfigured() bool {
	return r.GetDefault() != nil && r.GetDefault().IsConfigured()
}
