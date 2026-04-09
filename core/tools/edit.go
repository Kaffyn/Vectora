package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kaffyn/Vectora/core/policies"
)

type EditTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *EditTool) Name() string        { return "edit" }
func (t *EditTool) Description() string { return "Replaces a specific block of text in a file." }
func (t *EditTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"path":{"type":"string"},"target":{"type":"string"},"replacement":{"type":"string"}},"required":["path","target","replacement"]}`)
}

func (t *EditTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Path        string `json:"path"`
		Target      string `json:"target"`
		Replacement string `json:"replacement"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	safePath := filepath.Join(t.TrustFolder, params.Path)
	if !t.Guardian.IsPathSafe(safePath) || t.Guardian.IsProtected(safePath) {
		return &ToolResult{Output: "Access Denied", IsError: true}, nil
	}
	if t.Guardian.IsModificationBlocked(safePath) {
		return &ToolResult{Output: "Modification Denied: File protected by .vectoraignore", IsError: true}, nil
	}

	data, err := os.ReadFile(safePath)
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}

	snapID := backupFileLocally(safePath)
	strData := string(data)

	if !strings.Contains(strData, params.Target) {
		return &ToolResult{Output: "Target string not found. Edit canceled.", IsError: true}, nil
	}

	newData := strings.Replace(strData, params.Target, params.Replacement, 1)
	err = os.WriteFile(safePath, []byte(newData), 0644)
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}

	return &ToolResult{Output: "File edited successfully.", SnapshotID: snapID}, nil
}
