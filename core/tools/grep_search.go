package tools

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Kaffyn/Vectora/core/policies"
)

type FindFilesTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *FindFilesTool) Name() string { return "find_files" }
func (t *FindFilesTool) Description() string {
	return "Searches for files matching a pattern recursively."
}
func (t *FindFilesTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"pattern":{"type":"string"}},"required":["pattern"]}`)
}

func (t *FindFilesTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Pattern string `json:"pattern"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-c", "Get-ChildItem", "-Path", t.TrustFolder, "-Recurse", "-Filter", params.Pattern, "-Name")
	} else {
		cmd = exec.CommandContext(ctx, "find", t.TrustFolder, "-name", params.Pattern)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ToolResult{Output: string(out)}, nil
	}
	return &ToolResult{Output: string(out)}, nil
}

// ---- GrepSearchTool ----

type GrepSearchTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *GrepSearchTool) Name() string        { return "grep_search" }
func (t *GrepSearchTool) Description() string { return "Searches file contents using regex." }
func (t *GrepSearchTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"pattern":{"type":"string"},"case_sensitive":{"type":"boolean"}},"required":["pattern"]}`)
}

func (t *GrepSearchTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Pattern       string `json:"pattern"`
		CaseSensitive bool   `json:"case_sensitive"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	var matches []string
	_ = filepath.WalkDir(t.TrustFolder, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if t.Guardian.IsProtected(path) {
			return nil
		}
		data, e := os.ReadFile(path)
		if e == nil {
			if params.CaseSensitive {
				if stringsContains(data, params.Pattern) {
					rel, _ := filepath.Rel(t.TrustFolder, path)
					matches = append(matches, rel)
				}
			} else {
				if stringsContainsFold(data, params.Pattern) {
					rel, _ := filepath.Rel(t.TrustFolder, path)
					matches = append(matches, rel)
				}
			}
		}
		return nil
	})

	if len(matches) == 0 {
		return &ToolResult{Output: "No matches found"}, nil
	}

	return &ToolResult{
		Output:   stringsJoin(matches, "\n"),
		Metadata: map[string]interface{}{"count": len(matches)},
	}, nil
}

// Simple string contains check without importing strings
func stringsContains(data []byte, substr string) bool {
	s := string(data)
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func stringsContainsFold(data []byte, substr string) bool {
	s := string(data)
	lowerS := toLower(s)
	lowerSub := toLower(substr)
	return findSubstring(lowerS, lowerSub)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func stringsJoin(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
