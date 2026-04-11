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
		return CompletionResponse{}, fmt.Errorf("qwen_offline: Motor de inferência não iniciado")
	}

	lastMsg := p.prepareLastMsg(req)
	content, err := p.process.SendPrompt(ctx, lastMsg, req)
	if err != nil {
		return CompletionResponse{}, err
	}

	return CompletionResponse{
		Content: content,
	}, nil
}

func (p *QwenProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, <-chan error) {
	respChan := make(chan CompletionResponse, 1)
	errChan := make(chan error, 1)

	if p.process == nil {
		errChan <- fmt.Errorf("qwen_offline")
		close(respChan)
		close(errChan)
		return respChan, errChan
	}

	go func() {
		defer close(respChan)
		defer close(errChan)

		lastMsg := p.prepareLastMsg(req)
		tokens, errs := p.process.StreamPrompt(ctx, lastMsg, req)

		for {
			select {
			case token, ok := <-tokens:
				if !ok {
					return
				}
				respChan <- CompletionResponse{
					Content: token,
				}
			case err := <-errs:
				if err != nil {
					errChan <- err
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return respChan, errChan
}

func (p *QwenProvider) prepareLastMsg(req CompletionRequest) string {
	if len(req.Messages) > 0 {
		m := req.Messages[len(req.Messages)-1]
		return fmt.Sprintf("[%s]: %s\n", m.Role, m.Content)
	}
	return ""
}

func (p *QwenProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	return nil, fmt.Errorf("qwen_embed_not_implemented: inference via pipes focuses on chat")
}

func (p *QwenProvider) Shutdown() {
	if p.process != nil {
		p.process.Close()
	}
}
