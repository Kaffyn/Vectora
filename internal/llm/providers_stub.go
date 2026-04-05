package llm

import (
	"context"
)

// GeminiProvider is a stub implementation
type GeminiProvider struct {
}

// Complete returns a stub response
func (p *GeminiProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{
		FinishReason: "stop",
		Content: "Stub response - Gemini provider not yet implemented",
	}, nil
}

// QwenProvider is a stub implementation
type QwenProvider struct {
}

// Complete returns a stub response
func (p *QwenProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{
		FinishReason: "stop",
		Content: "Stub response - Qwen provider not yet implemented",
	}, nil
}
