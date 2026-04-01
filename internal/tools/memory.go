package tools

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/Vectora/internal/db"
)

// Tool linkable to persistent memory (In BBolt)
type SaveMemoryTool struct {
	KV db.KVStore
}

func (t *SaveMemoryTool) Name() string { return "save_memory" }
func (t *SaveMemoryTool) Description() string {
	return "Stores original base instructions about the environment in permanent chronic memory."
}
func (t *SaveMemoryTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"key":{"type":"string"},"value":{"type":"string"}},"required":["key","value"]}`)
}
func (t *SaveMemoryTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	if t.KV == nil {
		return ToolResult{IsError: true, Output: "BBolt KVStore Unavailable at Runtime."}, nil
	}

	key, _ := args["key"].(string)
	val, _ := args["value"].(string)

	err := t.KV.Set(ctx, "memories", key, []byte(val))
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}

	return ToolResult{Output: "Synapse stored! (This will stay with me forever)"}, nil
}

// Temporal Fragmentation Tool (Meta Agent)
type PlanModeTool struct{}

func (t *PlanModeTool) Name() string { return "enter_plan_mode" }
func (t *PlanModeTool) Description() string {
	return "Activates sub-task breakdown. Use when facing massive software engineering requirements."
}
func (t *PlanModeTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"plan_description":{"type":"string"}},"required":["plan_description"]}`)
}
func (t *PlanModeTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	plan, _ := args["plan_description"].(string)
	return ToolResult{Output: "Planning initialized locally and emitted via IPC to the interface: \n" + plan}, nil
}
