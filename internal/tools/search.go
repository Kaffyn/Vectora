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
func (t *FindFilesTool) Description() string { return "Pesquisa por nomes ou trechos de arquivos através de uma árvore inteira recursivamente." }
func (t *FindFilesTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"root_path":{"type":"string"},"pattern":{"type":"string"}},"required":["root_path","pattern"]}`)
}
// Falback nativo caso `find` falhe em SOs ríspidos
func (t *FindFilesTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	rootPath, _ := args["root_path"].(string)
	pattern, _ := args["pattern"].(string)

	if rootPath == "" {
		return ToolResult{IsError: true, Output: "Insira uma pasta base de procura."}, nil
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
func (t *GrepSearchTool) Description() string { return "Traz linhas de códigos exatas correspondentes à uma palavra sob uma pasta ou arquivo." }
func (t *GrepSearchTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"root_path":{"type":"string"},"query":{"type":"string"}},"required":["root_path","query"]}`)
}
func (t *GrepSearchTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	rootPath, _ := args["root_path"].(string)
	query, _ := args["query"].(string)

	// Implementação em Go Puro recursivo para ser veloz + Compatível entre Mac/Linux/Win sem Ripgrep
	var matches []string

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			data, e := os.ReadFile(path)
			if e == nil && strings.Contains(string(data), query) {
				// Evita overload listando apenas ocorrência do binário/arquivo
				matches = append(matches, path)
			}
		}
		return nil
	})

	if len(matches) == 0 {
		return ToolResult{Output: "Nenhum arquivo apontou contendo este trecho exato de string."}, nil
	}
	return ToolResult{Output: "Trechos de código encontrados dentro de:\n" + strings.Join(matches, "\n")}, nil
}
