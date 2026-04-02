package llm

import (
	"context"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

type Message struct {
	Role    Role
	Content string
}

type ToolDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Schema      []byte   `json:"schema"` // JSON Schema bruto via string
}

type ToolCall struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Args string `json:"args"` // JSON string de argumentos gerados
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CompletionRequest struct {
	Messages     []Message
	SystemPrompt string
	MaxTokens    int
	Temperature  float32
	Tools        []ToolDefinition
}

type CompletionResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     TokenUsage `json:"usage"`
}

// Provider abstrai qualquer LLM Backend (Qwen/Gemini/OpenAI) sob a interface unida do Vectora,
// assegurado internamente via langchaingo.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	Embed(ctx context.Context, input string) ([]float32, error)
	Name() string
	IsConfigured() bool
}
