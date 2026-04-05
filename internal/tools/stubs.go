package tools

import (
	"context"
	"encoding/json"
	"github.com/Kaffyn/Vectora/internal/db"
)

// SaveMemoryTool is a stub
type SaveMemoryTool struct {
	KV db.KVStore
}

// PlanModeTool is a stub
type PlanModeTool struct {
}

// Name returns the tool name
func (t *SaveMemoryTool) Name() string {
	return "save_memory"
}

// Description returns the tool description
func (t *SaveMemoryTool) Description() string {
	return "Save information to memory"
}

// Schema returns the tool schema
func (t *SaveMemoryTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type": "object"}`)
}

// Execute executes the tool
func (t *SaveMemoryTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	return ToolResult{Output: "Memory saved", IsError: false}, nil
}

// Name returns the tool name
func (t *PlanModeTool) Name() string {
	return "plan_mode"
}

// Description returns the tool description
func (t *PlanModeTool) Description() string {
	return "Activate plan mode"
}

// Schema returns the tool schema
func (t *PlanModeTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type": "object"}`)
}

// Execute executes the tool
func (t *PlanModeTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	return ToolResult{Output: "Plan mode activated", IsError: false}, nil
}
