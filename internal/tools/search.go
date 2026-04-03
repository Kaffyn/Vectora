package tools

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ----------------------------------------------------
// 1. Tool: find_files
// ----------------------------------------------------
type FindFilesTool struct{}

func (t *FindFilesTool) Name() string        { return "find_files" }
func (t *FindFilesTool) Description() string { return "Searches for file names or fragments across an entire tree recursively." }
func (t *FindFilesTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"root_path":{"type":"string"},"pattern":{"type":"string"}},"required":["root_path","pattern"]}`)
}
// Native fallback if `find` fails on restrictive OS environments
func (t *FindFilesTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	rootPath, _ := args["root_path"].(string)
	pattern, _ := args["pattern"].(string)

	if rootPath == "" {
		return ToolResult{IsError: true, Output: "Please provide a base search folder."}, nil
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-c", "Get-ChildItem", "-Path", rootPath, "-Recurse", "-Filter", pattern, "-Name")
	} else {
		cmd = exec.CommandContext(ctx, "find", rootPath, "-name", pattern)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return ToolResult{IsError: true, Output: string(out)}, nil
	}
	return ToolResult{Output: string(out)}, nil
}

// ----------------------------------------------------
// 2. Tool: grep_search
// ----------------------------------------------------
type GrepSearchTool struct{}

func (t *GrepSearchTool) Name() string        { return "grep_search" }
func (t *GrepSearchTool) Description() string { return "Finds exact lines of code matching a query within a folder or file." }
func (t *GrepSearchTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"root_path":{"type":"string"},"query":{"type":"string"}},"required":["root_path","query"]}`)
}
func (t *GrepSearchTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	rootPath, _ := args["root_path"].(string)
	query, _ := args["query"].(string)

	// Pure Go recursive implementation for speed and cross-platform compatibility (Mac/Linux/Win) without Ripgrep
	var matches []string

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			data, e := os.ReadFile(path)
			if e == nil && strings.Contains(string(data), query) {
				// Avoid overload by listing only the file occurrences.
				matches = append(matches, path)
			}
		}
		return nil
	})

	if len(matches) == 0 {
		return ToolResult{Output: "No files found containing this exact string fragment."}, nil
	}
	return ToolResult{Output: "Code fragments found within:\n" + strings.Join(matches, "\n")}, nil
}
