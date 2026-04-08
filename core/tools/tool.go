package tools

import (
	"context"
	"encoding/json"
)

// ToolResult is the standardized response for agentic tool execution
type ToolResult struct {
	Output     string                 `json:"output"`
	IsError    bool                   `json:"is_error"`
	SnapshotID string                 `json:"snapshot_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Tool defines the baseline contract for agentic execution
type Tool interface {
	Name() string
	Description() string
	Schema() json.RawMessage
	Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}
