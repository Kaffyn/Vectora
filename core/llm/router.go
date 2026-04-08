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
	p := r.GetDefault()
	if p == nil {
		return nil, fmt.Errorf("no LLM provider configured")
	}
	return p.Embed(ctx, input)
}

func (r *Router) IsConfigured() bool {
	return r.GetDefault() != nil && r.GetDefault().IsConfigured()
}
