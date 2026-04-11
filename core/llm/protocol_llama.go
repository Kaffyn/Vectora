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
	tokens, errs := p.StreamPrompt(ctx, prompt, opts)
	var sb strings.Builder
	for {
		select {
		case token, ok := <-tokens:
			if !ok {
				return sb.String(), nil
			}
			sb.WriteString(token)
		case err := <-errs:
			if err != nil {
				return "", err
			}
			return sb.String(), nil
		case <-ctx.Done():
			return sb.String(), ctx.Err()
		}
	}
}

func (p *LlamaProcess) StreamPrompt(ctx context.Context, prompt string, opts CompletionRequest) (<-chan string, <-chan error) {
	p.mu.Lock()

	tokenChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer p.mu.Unlock()
		defer close(tokenChan)
		defer close(errChan)

		req := LlamaRequest{
			Prompt:      prompt,
			MaxTokens:   opts.MaxTokens,
			Temperature: opts.Temperature,
			Stop:        []string{"</s>", "[INST]", "[/INST]"},
		}

		if err := json.NewEncoder(p.stdin).Encode(req); err != nil {
			errChan <- fmt.Errorf("stdin_write_failed: %v", err)
			return
		}

		for p.stdout.Scan() {
			text := p.stdout.Text()

			var token LlamaToken
			if err := json.Unmarshal([]byte(text), &token); err != nil {
				continue
			}

			if token.Error != "" {
				errChan <- fmt.Errorf("llama_inference_error: %s", token.Error)
				return
			}

			tokenChan <- token.Token
			if token.Done {
				break
			}
		}
	}()

	return tokenChan, errChan
}

func (p *LlamaProcess) Close() error {
	if p.stdin != nil {
		p.stdin.Close()
	}
	return p.cmd.Wait()
}
