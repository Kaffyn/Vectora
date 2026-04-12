package llm

import (
	"context"
	"encoding/json"
)

type Message struct {
	Role       Role           `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"` // For thought_signature, etc.
}

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"` // JSON Schema bruto
}

type ToolCall struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Args string `json:"args"` // JSON string of generated arguments
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CompletionRequest struct {
	Model         string
	FallbackModel string
	Messages      []Message
	SystemPrompt  string
	MaxTokens     int
	Temperature   float32
	Tools         []ToolDefinition
}

type CompletionResponse struct {
	Content   string         `json:"content"`
	ToolCalls []ToolCall     `json:"tool_calls,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Usage     TokenUsage     `json:"usage"`
}

// Provider abstracts any LLM Backend (Qwen/Gemini/OpenAI) under Vectora's unified interface.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	StreamComplete(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, <-chan error)
	Embed(ctx context.Context, input string, model string) ([]float32, error)
	ListModels(ctx context.Context) ([]string, error)
	Name() string
	IsConfigured() bool
}
