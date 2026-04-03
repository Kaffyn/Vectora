package llm

import (
	"context"
	"fmt"
	"runtime"

	"github.com/Kaffyn/vectora/internal/engines"
)

type QwenProvider struct {
	process      *LlamaProcess
	engineMgr    *engines.EngineManager
	modelPath    string
	binPath      string
	systemPrompt string
	toolsSpec    string
}

// NewQwenProvider initializes the engine via Zero-Port architecture.
func NewQwenProvider(ctx context.Context, modelPath string) (*QwenProvider, error) {
	mgr, err := engines.NewManager()
	if err != nil {
		return nil, err
	}

	binPath := mgr.GetBinaryPath("llama", "llama-cli")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	// Starts the interactive Llama process with flags for STDIO IPC
	proc, err := NewLlamaProcess(ctx, binPath, 
		"--model", modelPath, 
		"--interactive", 
		"--simple-io",
		"--no-display-prompt",
	)
	if err != nil {
		return nil, fmt.Errorf("llama_start_failed: %v", err)
	}

	// Load instruction and tool specs from the master instructions.

	return &QwenProvider{
		process:      proc,
		engineMgr:    mgr,
		modelPath:    modelPath,
		binPath:      binPath,
		systemPrompt: pInstr,
		toolsSpec:    tSpec,
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

	// Resolve Master Instruction
	masterPrompt := req.SystemPrompt
	if masterPrompt == "" {
		masterPrompt = p.systemPrompt + "\n\nAVAILABLE TOOLS (JSON SCHEMA):\n" + p.toolsSpec
	}

	// Builds the formatted history for interactive llama-cli
	// In interactive mode, Llama keeps previous history in the KV Cache.
	// We only send what is NEW.
	var lastMsg string
	if len(req.Messages) > 0 {
		m := req.Messages[len(req.Messages)-1]
		lastMsg = fmt.Sprintf("[%s]: %s\n", m.Role, m.Content)
	}

	// TODO: On first Chat execution, p.process.SendPrompt sends the 'masterPrompt'.
	// The return of this first call is currently ignored (Silent Setup).
	
	content, err := p.process.SendPrompt(ctx, lastMsg, req)
	if err != nil {
		return CompletionResponse{}, err
	}

	return CompletionResponse{
		Content: content,
	}, nil
}

func (p *QwenProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	return nil, fmt.Errorf("qwen_embed_not_implemented: inferência via pipes ainda foca em chat")
}

func (p *QwenProvider) Shutdown() {
	if p.process != nil {
		p.process.Close()
	}
}
