package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesystemTools(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "vectora-tools-*")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	testFile := filepath.Join(tmpDir, "hello.txt")

	// 1. Tool: write_file
	write := &WriteFileTool{}
	res, err := write.Execute(ctx, map[string]any{"path": testFile, "content": "hello world"})
	if err != nil || res.IsError {
		t.Errorf("write_file falhou: %v / %s", err, res.Output)
	}

	// 2. Tool: read_file
	read := &ReadFileTool{}
	res, err = read.Execute(ctx, map[string]any{"path": testFile})
	if err != nil || res.IsError || res.Output != "hello world" {
		t.Errorf("read_file obteve lixo: %s", res.Output)
	}

	// 3. Tool: edit
	edit := &EditTool{}
	res, err = edit.Execute(ctx, map[string]any{
		"path":        testFile,
		"target":      "world",
		"replacement": "vectora",
	})
	if err != nil || res.IsError {
		t.Errorf("edit falhou: %v", err)
	}

	// Double check edit
	res, _ = read.Execute(ctx, map[string]any{"path": testFile})
	if res.Output != "hello vectora" {
		t.Errorf("Esperado 'hello vectora', obtido '%s'", res.Output)
	}
}

func TestSearchTools(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "vectora-search-*")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	f1 := filepath.Join(tmpDir, "file1.go")
	os.WriteFile(f1, []byte("func main() { fmt.Println(\"ai\") }"), 0644)

	// Tool: grep_search
	gs := &GrepSearchTool{}
	res, err := gs.Execute(ctx, map[string]any{"root_path": tmpDir, "query": "fmt"})
	if err != nil || res.IsError || res.Output == "" {
		t.Errorf("grep_search falhou em achar constant: %v", err)
	}
}
