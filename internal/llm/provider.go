package llm

import (
	"context"
	"os"
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
	Args string `json:"args"` // JSON string of generated arguments
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CompletionRequest struct {
	Model        string
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

// Provider abstracts any LLM Backend (Qwen/Gemini/OpenAI) under Vectora's unified interface.
// assegurado internamente via langchaingo.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	Embed(ctx context.Context, input string) ([]float32, error)
	Name() string
	IsConfigured() bool
}

// LoadMasterInstructions reads prompt and tool files from the instruct folder.
func LoadMasterInstructions() (string, string) {
	promptPath := `C:\Users\bruno\Desktop\Vectora\internal\llm\instruct\prompt.txt`
	toolsPath := `C:\Users\bruno\Desktop\Vectora\internal\llm\instruct\tools.json`

	prompt, _ := os.ReadFile(promptPath)
	tools, _ := os.ReadFile(toolsPath)

	return string(prompt), string(tools)
}
