package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Kaffyn/Vectora/core/db"
)

// ClaudeProvider implements the Provider interface for Anthropic Claude via HTTP API.
type ClaudeProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewClaudeProvider creates a new Claude provider using the Anthropic API.
func NewClaudeProvider(ctx context.Context, apiKey string) (*ClaudeProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("claude_api_key_required: Claude API key was not provided")
	}

	return &ClaudeProvider{
		apiKey: apiKey,
		model:  "claude-3-5-sonnet-20241022",
		client: &http.Client{},
	}, nil
}

func (p *ClaudeProvider) Name() string {
	return "claude"
}

func (p *ClaudeProvider) IsConfigured() bool {
	return p.apiKey != "" && p.client != nil
}

// Complete sends a completion request to Claude API.
func (p *ClaudeProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	// Build messages array for Claude API
	var messages []map[string]string
	for _, msg := range req.Messages {
		role := string(msg.Role)
		if role == "system" {
			// Claude handles system prompt separately
			continue
		}
		if role == "model" {
			role = "assistant"
		}
		messages = append(messages, map[string]string{
			"role":    role,
			"content": msg.Content,
		})
	}

	// Extract system prompt
	var systemPrompt string
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			break
		}
	}

	// Build request body
	body := map[string]any{
		"model":       p.model,
		"messages":    messages,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}
	if systemPrompt != "" {
		body["system"] = systemPrompt
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyJSON))
	if err != nil {
		return CompletionResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("claude request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return CompletionResponse{}, fmt.Errorf("claude API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var claudeResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return CompletionResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	var content string
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return CompletionResponse{
		Content: content,
		Usage: TokenUsage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}, nil
}

// Embed generates an embedding using Claude. Note: Claude doesn't natively support embeddings,
// so this falls back to a simple hash-based embedding for compatibility.
// In production, use a dedicated embedding model (e.g., Gemini Embedding).
func (p *ClaudeProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	// Claude doesn't have a native embedding API.
	// Return a deterministic hash-based embedding for compatibility.
	// For production RAG, configure Gemini as the embedding provider.
	return db.GenerateDummyEmbedding(input, 768), nil
}

// Close releases resources held by the provider.
func (p *ClaudeProvider) Close() {
	// No persistent connections to close
}
