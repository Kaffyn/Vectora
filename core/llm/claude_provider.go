package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ClaudeProvider struct {
	apiKey string
	client *http.Client
}

func NewClaudeProvider(ctx context.Context, apiKey string) (*ClaudeProvider, error) {
	if apiKey == "" {
		return nil, errors.New("claude_api_key_required: Claude API key was not provided")
	}

	return &ClaudeProvider{
		apiKey: apiKey,
		client: &http.Client{},
	}, nil
}

func (p *ClaudeProvider) Name() string {
	return "claude"
}

func (p *ClaudeProvider) IsConfigured() bool {
	return p.apiKey != ""
}

type claudeMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // can be string or []interface{}
}

type claudeTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

type claudeRequest struct {
	Model       string          `json:"model"`
	Messages    []claudeMessage `json:"messages"`
	System      string          `json:"system,omitempty"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float32         `json:"temperature"`
	Tools       []claudeTool    `json:"tools,omitempty"`
}

type claudeResponse struct {
	ID      string `json:"id"`
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text,omitempty"`
		ID    string          `json:"id,omitempty"`   // for tool_use
		Name  string          `json:"name,omitempty"` // for tool_use
		Input json.RawMessage `json:"input,omitempty"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *ClaudeProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	modelName := req.Model
	if modelName == "" {
		modelName = "claude-3-5-sonnet-20241022"
	}

	// Map "4.6" aliases to real models
	if strings.Contains(modelName, "4.6-sonnet") {
		modelName = "claude-3-5-sonnet-20241022"
	} else if strings.Contains(modelName, "4.6-opus") {
		modelName = "claude-3-opus-20240229"
	} else if strings.Contains(modelName, "4.6-haiku") {
		modelName = "claude-3-5-haiku-20241022"
	}

	url := "https://api.anthropic.com/v1/messages"

	var messages []claudeMessage
	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			continue
		}
		messages = append(messages, claudeMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	claudeReq := claudeRequest{
		Model:       modelName,
		Messages:    messages,
		System:      req.SystemPrompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	if len(req.Tools) > 0 {
		for _, t := range req.Tools {
			var schema interface{}
			json.Unmarshal(t.Schema, &schema)
			claudeReq.Tools = append(claudeReq.Tools, claudeTool{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: schema,
			})
		}
	}

	body, _ := json.Marshal(claudeReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return CompletionResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return CompletionResponse{}, fmt.Errorf("claude API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var cResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return CompletionResponse{}, err
	}

	var content string
	var tCalls []ToolCall
	for _, item := range cResp.Content {
		if item.Type == "text" {
			content += item.Text
		} else if item.Type == "tool_use" {
			tCalls = append(tCalls, ToolCall{
				ID:   item.ID,
				Name: item.Name,
				Args: string(item.Input),
			})
		}
	}

	return CompletionResponse{
		Content:   content,
		ToolCalls: tCalls,
		Usage: TokenUsage{
			PromptTokens:     cResp.Usage.InputTokens,
			CompletionTokens: cResp.Usage.OutputTokens,
			TotalTokens:      cResp.Usage.InputTokens + cResp.Usage.OutputTokens,
		},
	}, nil
}

func (p *ClaudeProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	return nil, errors.New("claude_no_native_embedding")
}
