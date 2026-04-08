package tools

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/policies"
)

type SaveMemoryTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
	KV          db.KVStore
}

func (t *SaveMemoryTool) Name() string        { return "save_memory" }
func (t *SaveMemoryTool) Description() string { return "Stores a key-value pair in persistent memory." }
func (t *SaveMemoryTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"key":{"type":"string"},"value":{"type":"string"}},"required":["key","value"]}`)
}

func (t *SaveMemoryTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	if t.KV == nil {
		return &ToolResult{Output: "KVStore unavailable", IsError: true}, nil
	}

	err := t.KV.Set(ctx, "memories", params.Key, []byte(params.Value))
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}

	return &ToolResult{Output: "Memory stored successfully."}, nil
}
