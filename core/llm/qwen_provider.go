package llm

import (
	"context"
	"fmt"
)

type QwenProvider struct {
	process      *LlamaProcess
	modelPath    string
	binPath      string
	systemPrompt string
	toolsSpec    string
}

// NewQwenProvider initializes the engine via Zero-Port architecture.
func NewQwenProvider(ctx context.Context, binPath string, modelPath string) (*QwenProvider, error) {
	if binPath == "" {
		return nil, fmt.Errorf("llama_binary_not_found")
	}

	proc, err := NewLlamaProcess(ctx, binPath,
		"--model", modelPath,
		"--interactive",
		"--simple-io",
		"--no-display-prompt",
	)
	if err != nil {
		return nil, fmt.Errorf("llama_start_failed: %v", err)
	}

	return &QwenProvider{
		process:      proc,
		modelPath:    modelPath,
		binPath:      binPath,
		systemPrompt: "",
		toolsSpec:    "",
	}, nil
}

func (p *QwenProvider) Name() string {
	return "qwen"
}

func (p *QwenProvider) IsConfigured() bool {
	return p.process != nil
}

func (p *QwenProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	if p.process == nil {
		return CompletionResponse{}, fmt.Errorf("qwen_offline: motor de inferência não iniciado")
	}

	masterPrompt := req.SystemPrompt
	if masterPrompt == "" {
		masterPrompt = p.systemPrompt
	}

	var lastMsg string
	if len(req.Messages) > 0 {
		m := req.Messages[len(req.Messages)-1]
		lastMsg = fmt.Sprintf("[%s]: %s\n", m.Role, m.Content)
	}

	content, err := p.process.SendPrompt(ctx, lastMsg, req)
	if err != nil {
		return CompletionResponse{}, err
	}

	return CompletionResponse{
		Content: content,
	}, nil
}

func (p *QwenProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	return nil, fmt.Errorf("qwen_embed_not_implemented: inference via pipes focuses on chat")
}

func (p *QwenProvider) Shutdown() {
	if p.process != nil {
		p.process.Close()
	}
}
