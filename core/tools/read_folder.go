package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kaffyn/Vectora/core/policies"
)

type ReadFolderTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *ReadFolderTool) Name() string { return "read_folder" }
func (t *ReadFolderTool) Description() string {
	return "Lists files and directories within a base directory."
}
func (t *ReadFolderTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}

func (t *ReadFolderTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	safePath := filepath.Join(t.TrustFolder, params.Path)
	if !t.Guardian.IsPathSafe(safePath) {
		return &ToolResult{Output: "Access Denied", IsError: true}, nil
	}

	entries, err := os.ReadDir(safePath)
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}

	var dirs []string
	for _, e := range entries {
		kind := "[File]"
		if e.IsDir() {
			kind = "[Folder]"
		}
		dirs = append(dirs, fmt.Sprintf("%s %s", kind, e.Name()))
	}

	return &ToolResult{Output: strings.Join(dirs, "\n")}, nil
}
