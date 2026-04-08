package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"vectora/core/policies"
)

type GrepSearchTool struct {
	TrustFolder string
	Guardian    *policies.Guardian
}

func (t *GrepSearchTool) Name() string        { return "grep_search" }
func (t *GrepSearchTool) Description() string { return "Busca por padrão regex em arquivos do projeto." }
func (t *GrepSearchTool) Schema() string {
	return `{"type": "object", "properties": {"pattern": {"type": "string"}, "case_sensitive": {"type": "boolean"}}, "required": ["pattern"]}`
}

func (t *GrepSearchTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Pattern       string `json:"pattern"`
		CaseSensitive bool   `json:"case_sensitive"`
	}
	json.Unmarshal(args, &params)

	var re *regexp.Regexp
	var err error
	if params.CaseSensitive {
		re, err = regexp.Compile(params.Pattern)
	} else {
		re, err = regexp.Compile("(?i)" + params.Pattern)
	}
	if err != nil {
		return &ToolResult{Output: "Invalid Regex", IsError: true}, nil
	}

	var matches []string
	filepath.WalkDir(t.TrustFolder, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || t.Guardian.IsProtected(path) {
			return nil
		}

		data, _ := os.ReadFile(path)
		if re.Match(data) {
			rel, _ := filepath.Rel(t.TrustFolder, path)
			matches = append(matches, rel)
		}
		return nil
	})

	if len(matches) == 0 {
		return &ToolResult{Output: "No matches found"}, nil
	}

	return &ToolResult{
		Output:   strings.Join(matches, "\n"),
		Metadata: map[string]interface{}{"count": len(matches)},
	}, nil
}
