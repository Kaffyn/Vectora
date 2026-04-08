package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Kaffyn/Vectora/core/policies"
)

const MAX_READ_BYTES = 50 * 1024 // 50KB

type ReadFileTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *ReadFileTool) Name() string { return "read_file" }
func (t *ReadFileTool) Description() string {
	return "Reads the content of a file within the Trust Folder."
}
func (t *ReadFileTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}

func (t *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	safePath := filepath.Join(t.TrustFolder, params.Path)

	if !t.Guardian.IsPathSafe(safePath) || t.Guardian.IsProtected(safePath) {
		return &ToolResult{Output: "Access Denied", IsError: true}, nil
	}

	data, err := os.ReadFile(safePath)
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}

	content := string(data)
	truncated := false
	if len(data) > MAX_READ_BYTES {
		content = string(data[:MAX_READ_BYTES])
		truncated = true
	}

	if truncated {
		content += "\n... [TRUNCATED: Use grep_search for specific content] ..."
	}

	return &ToolResult{
		Output:   content,
		Metadata: map[string]interface{}{"truncated": truncated, "size": len(data)},
	}, nil
}
