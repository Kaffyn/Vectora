package llm

import (
	"context"
	"testing"
)

type mockProvider struct {
	configured bool
	name       string
}

func (m *mockProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	return CompletionResponse{Content: "mock completion"}, nil
}

func (m *mockProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, <-chan error) {
	return nil, nil
}

func (m *mockProvider) Embed(ctx context.Context, input string, taskType string) ([]float32, error) {
	return []float32{0.1}, nil
}

func (m *mockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"model1"}, nil
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) IsConfigured() bool {
	return m.configured
}

func TestRouter_Registration(t *testing.T) {
	router := NewRouter()
	p1 := &mockProvider{configured: true, name: "p1"}
	p2 := &mockProvider{configured: false, name: "p2"}

	router.RegisterProvider("p1", p1, true)
	router.RegisterProvider("p2", p2, false)

	if router.defaultProvider != "p1" {
		t.Errorf("expected default provider p1, got %s", router.defaultProvider)
	}

	p, err := router.GetProvider("p2")
	if err != nil {
		t.Errorf("failed to get provider p2: %v", err)
	}
	// Use type assertion or compare as interface
	if p.Name() != p2.Name() {
		t.Error("provider mismatch for p2")
	}
}

func TestRouter_IsConfigured(t *testing.T) {
	router := NewRouter()

	if router.IsConfigured() {
		t.Error("empty router should not be configured")
	}

	p1 := &mockProvider{configured: false, name: "p1"}
	router.RegisterProvider("p1", p1, true)
	if router.IsConfigured() {
		t.Error("router with unconfigured provider should not be configured")
	}

	p1.configured = true
	if !router.IsConfigured() {
		t.Error("router with configured provider should be configured")
	}
}
