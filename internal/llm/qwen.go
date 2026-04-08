package llm

import (
	"context"
	"fmt"
)

// QwenProvider for internal/ (legacy - use core/llm instead)
type QwenProvider struct {
	modelPath string
}

func NewQwenProvider(ctx context.Context, modelPath string) (*QwenProvider, error) {
	return nil, fmt.Errorf("qwen not available in internal/ - use core/llm instead")
}

func (p *QwenProvider) Name() string       { return "qwen" }
func (p *QwenProvider) IsConfigured() bool { return false }
func (p *QwenProvider) Shutdown()          {}
func (p *QwenProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	return CompletionResponse{}, fmt.Errorf("qwen not configured")
}
func (p *QwenProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	return nil, fmt.Errorf("qwen not configured")
}
