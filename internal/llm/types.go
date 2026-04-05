package llm

import "encoding/json"

// CompletionRequest represents an LLM completion request
type CompletionRequest struct {
	SystemPrompt string
	UserMessage  string
	Messages     []map[string]interface{}
	Context      string
	Tools        []map[string]interface{}
	Model        string
	MaxTokens    int
	Temperature  float32
}

// CompletionResponse represents an LLM completion response
type CompletionResponse struct {
	FinishReason string
	Content      string
	ToolCalls    []ToolCall
	Thinking     string
	Usage        TokenUsage
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID        string
	Name      string
	ToolName  string
	Args      json.RawMessage
	Arguments json.RawMessage
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// LoadMasterInstructions loads system instructions
func LoadMasterInstructions() (string, error) {
	return `Você é o Vectora, um assistente de IA avançado integrado ao seu computador.`, nil
}
