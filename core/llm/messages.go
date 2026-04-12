package llm

import "time"

// Role defines the role in the conversation and in the provider abstraction.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ChatMessage represents a single entry in the conversation.
type ChatMessage struct {
	Role       Role           `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// Conversation represents the full history of a chat.
type Conversation struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Messages  []ChatMessage `json:"messages"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// LlamaRequest is the envelope sent to the stdin of the llama-cli process.
type LlamaRequest struct {
	Prompt      string   `json:"prompt,omitempty"`
	SystemShell string   `json:"system_prompt,omitempty"`
	MaxTokens   int      `json:"n_predict"`
	Temperature float32  `json:"temperature"`
	Stop        []string `json:"stop,omitempty"`
	Stream      bool     `json:"stream"`
}
