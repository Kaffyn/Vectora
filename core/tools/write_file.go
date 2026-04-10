package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Kaffyn/Vectora/core/policies"
)

type WriteFileTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Creates or overwrites a file within the Trust Folder."
}
func (t *WriteFileTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"}},"required":["path","content"]}`)
}

func (t *WriteFileTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	safePath := filepath.Join(t.TrustFolder, params.Path)
	if !t.Guardian.IsPathSafe(safePath) || t.Guardian.IsProtected(safePath) {
		return &ToolResult{Output: "Access Denied: Out of Trust Folder or Core Protected", IsError: true}, nil
	}
	if t.Guardian.IsModificationBlocked(safePath) {
		return &ToolResult{Output: "Modification Denied: File protected by .vectoraignore", IsError: true}, nil
	}

	// Backup existing file before overwrite
	snapID := ""
	if _, err := os.Stat(safePath); err == nil {
		snapID = backupFileLocally(safePath)
	}

	os.MkdirAll(filepath.Dir(safePath), 0755)
	err := os.WriteFile(safePath, []byte(params.Content), 0644)
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}

	return &ToolResult{
		Output:     "File written successfully.",
		SnapshotID: snapID,
	}, nil
}

// backupFileLocally creates a local backup of a file and returns a snapshot ID.
func backupFileLocally(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	backups := filepath.Join(home, ".Vectora", "backups")
	os.MkdirAll(backups, 0755)

	srcFile, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer srcFile.Close()

	dstPath := filepath.Join(backups, filepath.Base(path)+".bak")
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return ""
	}
	defer dstFile.Close()

	_, _ = dstFile.ReadFrom(srcFile)
	return dstPath
}
