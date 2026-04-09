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

func (r *Router) Embed(ctx context.Context, input string) ([]float32, error) {
	// 1. Try default provider first
	p := r.GetDefault()
	if p != nil {
		vec, err := p.Embed(ctx, input)
		if err == nil {
			return vec, nil
		}
	}

	// 2. Smart Fallback: Look for dedicated embedding providers
	// Priority 1: Voyage (High quality)
	if vp, err := r.GetProvider("voyage"); err == nil && vp.IsConfigured() {
		return vp.Embed(ctx, input)
	}

	// Priority 2: Gemini
	if gp, err := r.GetProvider("gemini"); err == nil && gp.IsConfigured() {
		return gp.Embed(ctx, input)
	}

	return nil, fmt.Errorf("no suitable embedding provider found (default failed, no Voyage/Gemini fallback)")
}

func (r *Router) IsConfigured() bool {
	return r.GetDefault() != nil && r.GetDefault().IsConfigured()
}
