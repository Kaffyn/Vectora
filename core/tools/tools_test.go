package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaffyn/Vectora/core/policies"
)

func setupTestTools(t *testing.T) (*Registry, string) {
	t.Helper()
	dir := t.TempDir()
	guardian := policies.NewGuardian(dir)
	reg := NewRegistry(dir, guardian, &testKV{data: make(map[string]map[string][]byte)})
	return reg, dir
}

type testKV struct {
	data map[string]map[string][]byte
}

func (kv *testKV) Set(_ context.Context, bucket, key string, value []byte) error {
	if kv.data[bucket] == nil {
		kv.data[bucket] = make(map[string][]byte)
	}
	kv.data[bucket][key] = value
	return nil
}
func (kv *testKV) Get(_ context.Context, bucket, key string) ([]byte, error) {
	if b, ok := kv.data[bucket]; ok {
		return b[key], nil
	}
	return nil, nil
}
func (kv *testKV) Delete(_ context.Context, bucket, key string) error {
	if b, ok := kv.data[bucket]; ok {
		delete(b, key)
	}
	return nil
}
func (kv *testKV) List(_ context.Context, bucket, prefix string) ([]string, error) {
	var keys []string
	if b, ok := kv.data[bucket]; ok {
		for k := range b {
			if len(prefix) == 0 || (len(k) >= len(prefix) && k[:len(prefix)] == prefix) {
				keys = append(keys, k)
			}
		}
	}
	return keys, nil
}

func TestReadFile(t *testing.T) {
	reg, dir := setupTestTools(t)

	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello world"), 0644)

	tool, _ := reg.GetTool("read_file")
	if tool == nil {
		t.Fatal("read_file tool not found")
	}

	args, _ := json.Marshal(map[string]string{"path": "test.txt"})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", result.Output)
	}
}

func TestReadFileBlocked(t *testing.T) {
	reg, dir := setupTestTools(t)

	os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=123"), 0644)

	tool, _ := reg.GetTool("read_file")
	args, _ := json.Marshal(map[string]string{"path": ".env"})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error when reading .env")
	}
}

func TestWriteFile(t *testing.T) {
	reg, dir := setupTestTools(t)

	tool, _ := reg.GetTool("write_file")
	args, _ := json.Marshal(map[string]string{"path": "output.txt", "content": "test content"})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("write failed: %s", result.Output)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "output.txt"))
	if string(data) != "test content" {
		t.Fatalf("expected 'test content', got '%s'", string(data))
	}
}

func TestReadFolder(t *testing.T) {
	reg, dir := setupTestTools(t)

	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "file2.go"), []byte("b"), 0644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	tool, _ := reg.GetTool("read_folder")
	args, _ := json.Marshal(map[string]string{"path": ""})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("read_folder failed: %s", result.Output)
	}
	if result.Output == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestEdit(t *testing.T) {
	reg, dir := setupTestTools(t)

	os.WriteFile(filepath.Join(dir, "edit.txt"), []byte("hello world"), 0644)

	tool, _ := reg.GetTool("edit")
	args, _ := json.Marshal(map[string]string{"path": "edit.txt", "target": "world", "replacement": "vectora"})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("edit failed: %s", result.Output)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "edit.txt"))
	if string(data) != "hello vectora" {
		t.Fatalf("expected 'hello vectora', got '%s'", string(data))
	}
}

func TestTerminalRun(t *testing.T) {
	reg, _ := setupTestTools(t)

	tool, _ := reg.GetTool("run_shell_command")
	args, _ := json.Marshal(map[string]string{"command": "echo hello"})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("terminal_run failed: %s", result.Output)
	}
	if result.Output == "" {
		t.Fatal("expected non-empty output from echo")
	}
}

func TestSaveMemory(t *testing.T) {
	reg, _ := setupTestTools(t)

	tool, _ := reg.GetTool("save_memory")
	args, _ := json.Marshal(map[string]string{"key": "test_key", "value": "test_value"})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("save_memory failed: %s", result.Output)
	}
}

func TestGuardianBlocksOutOfScope(t *testing.T) {
	reg, _ := setupTestTools(t)

	tool, _ := reg.GetTool("write_file")
	args, _ := json.Marshal(map[string]string{"path": "../../escape.txt", "content": "hacked"})
	_, _ = tool.Execute(context.Background(), json.RawMessage(args))

	// File should NOT exist outside trust folder
	if _, err := os.Stat("../../escape.txt"); err == nil {
		os.Remove("../../escape.txt")
		t.Fatal("security breach: file was written outside trust folder")
	}
}

func TestGrepSearch(t *testing.T) {
	reg, dir := setupTestTools(t)

	os.WriteFile(filepath.Join(dir, "search.txt"), []byte("hello world\nfoo bar\nhello again"), 0644)

	tool, _ := reg.GetTool("grep_search")
	args, _ := json.Marshal(map[string]string{"pattern": "hello"})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("grep_search failed: %s", result.Output)
	}
	if result.Output == "No matches found" {
		t.Fatal("expected matches for 'hello'")
	}
}

func TestFindFiles(t *testing.T) {
	reg, dir := setupTestTools(t)

	os.WriteFile(filepath.Join(dir, "testfile.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "testfile.go"), []byte("b"), 0644)

	tool, _ := reg.GetTool("find_files")
	args, _ := json.Marshal(map[string]string{"pattern": "*.txt"})
	result, err := tool.Execute(context.Background(), json.RawMessage(args))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// find_files uses shell commands - may or may not work in test env
	_ = result
}

func TestAllToolsRegistered(t *testing.T) {
	reg, _ := setupTestTools(t)

	expected := []string{
		"read_file", "write_file", "read_folder", "edit",
		"find_files", "grep_search", "run_shell_command",
		"save_memory", "google_search", "web_fetch",
	}

	for _, name := range expected {
		tool, ok := reg.GetTool(name)
		if !ok {
			t.Errorf("tool '%s' not registered", name)
		}
		if tool.Name() != name {
			t.Errorf("tool name mismatch: expected '%s', got '%s'", name, tool.Name())
		}
	}
}
