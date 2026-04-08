package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// LlamaToken is emitted via stdout for each generated token.
type LlamaToken struct {
	Token string `json:"token"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// LlamaProcess manages the lifecycle and communication via Pipes with the llama-cli.
type LlamaProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	mu     sync.Mutex
}

func NewLlamaProcess(ctx context.Context, binPath string, args ...string) (*LlamaProcess, error) {
	cmd := exec.CommandContext(ctx, binPath, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &LlamaProcess{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
	}, nil
}

func (p *LlamaProcess) SendPrompt(ctx context.Context, prompt string, opts CompletionRequest) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	req := LlamaRequest{
		Prompt:      prompt,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		Stop:        []string{"</s>", "[INST]", "[/INST]"},
	}

	if err := json.NewEncoder(p.stdin).Encode(req); err != nil {
		return "", fmt.Errorf("stdin_write_failed: %v", err)
	}

	var sb strings.Builder
	for p.stdout.Scan() {
		text := p.stdout.Text()

		var token LlamaToken
		if err := json.Unmarshal([]byte(text), &token); err != nil {
			continue
		}

		if token.Error != "" {
			return "", fmt.Errorf("llama_inference_error: %s", token.Error)
		}

		sb.WriteString(token.Token)
		if token.Done {
			break
		}
	}

	return sb.String(), nil
}

func (p *LlamaProcess) Close() error {
	if p.stdin != nil {
		p.stdin.Close()
	}
	return p.cmd.Wait()
}
