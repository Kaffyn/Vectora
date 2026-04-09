# Vectora Testing Audit

## 1. Current Methodology

The Vectora project currently employs a dual-language testing strategy, separating the Go-based Core from the TypeScript-based VS Code Extension.

### Go Core Methodology

- **Framework:** Uses the standard Go `testing` package.
- **Test Types:** Primarily **Unit Tests** for individual components (ACP, I18n, Policies, Tools).
- **Mocks & Stubs:** Implements manual mock structures (e.g., `mockEngine`, `testKV`) to isolate logic from external LLM providers, file systems, and database layers.
- **Execution:** Tests are executed via `go test ./...` from the root or within specific packages.

### VS Code Extension Methodology

- **Framework:** Uses `@vscode/test-electron` (formerly `vscode-test`) combined with the **Mocha** test runner and **Assert** library.
- **Test Types:** Focuses on **Smoke Tests** and **Integration Tests** within the VS Code environment.
- **Execution:** Run via `npm test` or `node ./out/test/runTest.js`.
- **Core Interaction:** Current automated tests primarily verify extension presence, activation, and command registration. They do not yet perform deep end-to-end integration testing against the live Go Core process.

---

## 2. Test Suite Structure

```text
Vectora/
├── core/
│   ├── api/acp/server_test.go      # ACP Protocol logic tests
│   ├── i18n/i18n_test.go           # Translation and fallback tests
│   ├── llm/router_test.go          # Multi-provider routing tests
│   ├── policies/guardian_test.go   # Security and trust boundary tests
│   └── tools/tools_test.go         # Agent tools execution tests
└── extensions/vscode/
    └── src/test/                   # Extension test root
        ├── runTest.ts              # Test runner configuration
        └── suite/
            ├── extension.test.ts   # Core extension integration tests
            └── index.ts            # Mocha suite entry point
```

---

## 3. Existing Test Files (Cleaned)

### [core/api/acp/server_test.go](file:///c:/Users/bruno/Desktop/Vectora/core/api/acp/server_test.go)

```go
package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type mockEngine struct{}

func (m *mockEngine) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *mockEngine) Query(ctx context.Context, query string, workspaceID string) (string, error) {
	return "This is a mock response for: " + query, nil
}

func (m *mockEngine) ExecuteTool(ctx context.Context, name string, args map[string]any) (ToolResult, error) {
	return ToolResult{Output: "Mock tool output for: " + name}, nil
}

func (m *mockEngine) ReadFile(ctx context.Context, path string) (string, error) {
	return "Mock file content for: " + path, nil
}

func (m *mockEngine) WriteFile(ctx context.Context, path, content string) error {
	return nil
}

func (m *mockEngine) RunCommand(ctx context.Context, cwd, command string) (string, error) {
	return "Command output: " + command, nil
}

func (m *mockEngine) CompleteCode(ctx context.Context, path, prefix, suffix, language string) (string, error) {
	return prefix + " /* mock completion */ " + suffix, nil
}

func TestACPInitialize(t *testing.T) {
	engine := &mockEngine{}
	server := NewServer(engine)

	input := `{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":1,"clientCapabilities":{"fs":{"readTextFile":true,"writeTextFile":true},"terminal":true},"clientInfo":{"name":"test-client","title":"Test Client","version":"1.0.0"}}}`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var raw map[string]json.RawMessage
	json.Unmarshal([]byte(input), &raw)

	result, errMsg := server.handleInitialize(ctx, raw["params"])

	if errMsg != "" {
		t.Fatalf("initialize failed: %s", errMsg)
	}

	resp, ok := result.(InitializeResponse)
	if !ok {
		t.Fatalf("expected InitializeResponse, got %T", result)
	}

	if resp.ProtocolVersion != 1 {
		t.Errorf("expected protocol version 1, got %d", resp.ProtocolVersion)
	}
	if resp.AgentInfo.Name != "vectora" {
		t.Errorf("expected agent name 'vectora', got '%s'", resp.AgentInfo.Name)
	}
	if resp.AgentCapabilities == nil {
		t.Error("expected agent capabilities")
	}
	if !resp.AgentCapabilities.LoadSession {
		t.Error("expected loadSession to be true")
	}

	fmt.Printf("✅ Initialize response: %+v\n", resp)
}

func TestACPSessionNew(t *testing.T) {
	engine := &mockEngine{}
	server := NewServer(engine)

	input := `{"cwd":"/home/user/project"}`
	var params map[string]any
	json.Unmarshal([]byte(input), &params)
	paramsJSON, _ := json.Marshal(params)

	result, errMsg := server.handleSessionNew(context.Background(), paramsJSON)
	if errMsg != "" {
		t.Fatalf("session/new failed: %s", errMsg)
	}

	resp, ok := result.(SessionNewResponse)
	if !ok {
		t.Fatalf("expected SessionNewResponse, got %T", result)
	}
	if resp.SessionID == "" {
		t.Error("expected non-empty session ID")
	}
	if !strings.HasPrefix(resp.SessionID, "sess_") {
		t.Errorf("expected session ID to start with 'sess_', got '%s'", resp.SessionID)
	}

	fmt.Printf("✅ Session created: %s\n", resp.SessionID)
}

func TestACPFSRead(t *testing.T) {
	engine := &mockEngine{}
	server := NewServer(engine)

	server.sessions["sess_test"] = &Session{
		ID:  "sess_test",
		CWD: "/test",
	}

	input := `{"sessionId":"sess_test","path":"/test/file.go"}`
	var params map[string]any
	json.Unmarshal([]byte(input), &params)
	paramsJSON, _ := json.Marshal(params)

	result, errMsg := server.handleFSRead(context.Background(), paramsJSON)
	if errMsg != "" {
		t.Fatalf("fs/read_text_file failed: %s", errMsg)
	}

	resp, ok := result.(FSReadResponse)
	if !ok {
		t.Fatalf("expected FSReadResponse, got %T", result)
	}
	if !strings.Contains(resp.Content, "Mock file content") {
		t.Errorf("expected mock content, got '%s'", resp.Content)
	}

	fmt.Printf("✅ FSRead content: %s\n", resp.Content)
}

func TestACPPrompt(t *testing.T) {
	engine := &mockEngine{}
	server := NewServer(engine)

	server.sessions["sess_test"] = &Session{
		ID:           "sess_test",
		CWD:          "/test",
		Updates:      make(chan SessionUpdate, 100),
		PermissionCh: make(chan PermissionResponse, 1),
	}

	input := `{"sessionId":"sess_test","prompt":[{"type":"text","text":"What is this codebase about?"}]}`
	var params map[string]any
	json.Unmarshal([]byte(input), &params)
	paramsJSON, _ := json.Marshal(params)

	result, errMsg := server.handleSessionPrompt(context.Background(), paramsJSON)
	if errMsg != "" {
		t.Fatalf("session/prompt failed: %s", errMsg)
	}

	resp, ok := result.(PromptResponse)
	if !ok {
		t.Fatalf("expected PromptResponse, got %T", result)
	}
	if resp.StopReason != StopEndTurn {
		t.Errorf("expected stop reason 'end_turn', got '%s'", resp.StopReason)
	}

	fmt.Printf("✅ Prompt completed with stop reason: %s\n", resp.StopReason)
}

func TestACPFullFlow(t *testing.T) {
	engine := &mockEngine{}
	server := NewServer(engine)

	ctx := context.Background()

	initReq := InitializeRequest{
		ProtocolVersion: 1,
		ClientInfo:      &Info{Name: "test", Title: "Test Client", Version: "1.0.0"},
	}
	initResult, _ := server.handleInitialize(ctx, toJSON(t, initReq))
	initResp := initResult.(InitializeResponse)
	if initResp.ProtocolVersion != 1 {
		t.Fatalf("version mismatch: got %d", initResp.ProtocolVersion)
	}

	newReq := SessionNewRequest{CWD: "/test/project"}
	newResult, _ := server.handleSessionNew(ctx, toJSON(t, newReq))
	sessionResp := newResult.(SessionNewResponse)
	if sessionResp.SessionID == "" {
		t.Fatal("no session ID")
	}

	promptReq := SessionPromptRequest{
		SessionID: sessionResp.SessionID,
		Prompt:    []ContentBlock{{Type: "text", Text: "Explain this code"}},
	}
	promptResult, errMsg := server.handleSessionPrompt(ctx, toJSON(t, promptReq))
	if errMsg != "" {
		t.Fatalf("prompt failed: %s", errMsg)
	}
	promptResp := promptResult.(PromptResponse)
	if promptResp.StopReason != StopEndTurn {
		t.Errorf("unexpected stop: %s", promptResp.StopReason)
	}

	fmt.Println("✅ Full ACP flow passed: initialize → session/new → session/prompt")
}

func toJSON(t *testing.T, v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return data
}
```

### [core/i18n/i18n_test.go](file:///c:/Users/bruno/Desktop/Vectora/core/i18n/i18n_test.go)

```go
package i18n

import "testing"

func TestI18nFallback(t *testing.T) {
	val := T("tray_status")
	if val == "" {
		t.Error("expected non-empty translation for tray_status")
	}
}

func TestI18nMultipleLangs(t *testing.T) {
	SetLanguage("pt")
	val := T("tray_status")
	if val == "" || val == "tray_status" {
		t.Errorf("expected Portuguese translation, got '%s'", val)
	}

	SetLanguage("en")
	valEn := T("tray_status")
	if valEn != "Status: Running" {
		t.Errorf("expected 'Status: Running', got '%s'", valEn)
	}

	SetLanguage("es")
	valEs := T("tray_status")
	if valEs != "Estado: En ejecución" {
		t.Errorf("expected 'Estado: En ejecución', got '%s'", valEs)
	}

	SetLanguage("fr")
	valFr := T("tray_status")
	if valFr != "Statut : En cours" {
		t.Errorf("expected 'Statut : En cours', got '%s'", valFr)
	}

	SetLanguage("pt")
}

func TestI18nUnknownKey(t *testing.T) {
	SetLanguage("en")
	val := T("unknown_key_xyz")
	if val != "unknown_key_xyz" {
		t.Errorf("expected key returned for unknown key, got '%s'", val)
	}
}
```

### [core/llm/router_test.go](file:///c:/Users/bruno/Desktop/Vectora/core/llm/router_test.go)

```go
package llm

import (
	"context"
	"testing"
)

type mockProvider struct {
	configured bool
	name       string
}

func (m *mockProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	return CompletionResponse{Content: "mock completion"}, nil
}

func (m *mockProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	return []float32{0.1}, nil
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) IsConfigured() bool {
	return m.configured
}

func TestRouter_Registration(t *testing.T) {
	router := NewRouter()
	p1 := &mockProvider{configured: true, name: "p1"}
	p2 := &mockProvider{configured: false, name: "p2"}

	router.RegisterProvider("p1", p1, true)
	router.RegisterProvider("p2", p2, false)

	if router.defaultProvider != "p1" {
		t.Errorf("expected default provider p1, got %s", router.defaultProvider)
	}

	p, err := router.GetProvider("p2")
	if err != nil {
		t.Errorf("failed to get provider p2: %v", err)
	}
	if p.Name() != p2.Name() {
		t.Error("provider mismatch for p2")
	}
}

func TestRouter_IsConfigured(t *testing.T) {
	router := NewRouter()

	if router.IsConfigured() {
		t.Error("empty router should not be configured")
	}

	p1 := &mockProvider{configured: false, name: "p1"}
	router.RegisterProvider("p1", p1, true)
	if router.IsConfigured() {
		t.Error("router with unconfigured provider should not be configured")
	}

	p1.configured = true
	if !router.IsConfigured() {
		t.Error("router with configured provider should be configured")
	}
}
```

### [core/policies/guardian_test.go](file:///c:/Users/bruno/Desktop/Vectora/core/policies/guardian_test.go)

```go
package policies

import (
	"path/filepath"
	"testing"
)

func TestGuardianBlocksProtectedFiles(t *testing.T) {
	g := NewGuardian("/trust")

	protectedFiles := []string{".env", "secrets.yml", "id_rsa", "test.key", "cert.pem", "db.sqlite"}
	for _, f := range protectedFiles {
		if !g.IsProtected(f) {
			t.Errorf("expected %s to be protected", f)
		}
	}
}

func TestGuardianBlocksProtectedExtensions(t *testing.T) {
	g := NewGuardian("/trust")

	protectedExts := []string{"file.db", "file.sqlite", "file.exe", "file.dll", "file.key", "file.pem", "file.log"}
	for _, f := range protectedExts {
		if !g.IsProtected(f) {
			t.Errorf("expected %s to be protected by extension", f)
		}
	}
}

func TestGuardianPathSafe(t *testing.T) {
	g := NewGuardian("/trust")

	if !g.IsPathSafe("/trust/file.txt") {
		t.Error("expected /trust/file.txt to be path-safe")
	}
	if !g.IsPathSafe("/trust/sub/file.txt") {
		t.Error("expected nested path to be safe")
	}
}

func TestGuardianExcludedDirs(t *testing.T) {
	g := NewGuardian("/trust")

	excluded := []string{".git", "node_modules", "vendor", "dist", "build"}
	for _, d := range excluded {
		if !g.IsExcludedDir(d) {
			t.Errorf("expected %s to be excluded", d)
		}
	}
}

func TestGuardianSanitizeOutput(t *testing.T) {
	g := NewGuardian("/trust")

	input := "normal text AKIAIOSFODNN7EXAMPLE more text"
	output := g.SanitizeOutput(input)
	if output == input {
		t.Error("expected secret to be redacted")
	}
}

func TestGuardianBlocksPathTraversal(t *testing.T) {
	g := NewGuardian("/trust")

	traversal := filepath.Join("/trust", "..", "escape.txt")
	if g.IsPathSafe(traversal) {
		t.Error("expected path traversal to be blocked")
	}
}
```

### [core/tools/tools_test.go](file:///c:/Users/bruno/Desktop/Vectora/core/tools/tools_test.go)

```go
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
```

### [extensions/vscode/src/test/runTest.ts](file:///c:/Users/bruno/Desktop/Vectora/extensions/vscode/src/test/runTest.ts)

```typescript
import * as path from "path";
import * as os from "os";
import { runTests } from "@vscode/test-electron";

async function main() {
  try {
    let vscodeExecutablePath: string | undefined = process.env.VECTORA_VSCODE_PATH;

    if (!vscodeExecutablePath && os.platform() === "win32") {
      const localAppData = process.env.LOCALAPPDATA || path.join(os.homedir(), "AppData", "Local");
      const standardPath = path.join(localAppData, "Programs", "Microsoft VS Code", "Code.exe");
      vscodeExecutablePath = standardPath;
    }

    const extensionDevelopmentPath = path.resolve(__dirname, "../../");

    const extensionTestsPath = path.resolve(__dirname, "./suite/index");

    console.log(`Using VS Code executable: ${vscodeExecutablePath}`);

    await runTests({
      vscodeExecutablePath,
      extensionDevelopmentPath,
      extensionTestsPath,
      launchArgs: ["--disable-extensions"],
    });

    console.log("[TEST] Tests completed. Waiting 10 seconds before closing VS Code...");
    await new Promise((resolve) => setTimeout(resolve, 10000));
  } catch (err) {
    console.error("Failed to run tests");
    process.exit(1);
  }
}

main();
```

### [extensions/vscode/src/test/suite/extension.test.ts](file:///c:/Users/bruno/Desktop/Vectora/extensions/vscode/src/test/suite/extension.test.ts)

```typescript
import * as assert from "assert";
import * as vscode from "vscode";

suite("Vectora Extension Test Suite", () => {
  vscode.window.showInformationMessage("Start all tests.");

  test("Extension should be present", () => {
    assert.ok(vscode.extensions.getExtension("kaffyn.vectora"));
  });

  test("Extension should activate", async () => {
    const extension = vscode.extensions.getExtension("kaffyn.vectora");
    await extension?.activate();
    assert.strictEqual(extension?.isActive, true);
  });

  test("Commands should be registered", async () => {
    const commands = await vscode.commands.getCommands(true);
    assert.ok(commands.includes("vectora.newSession"));
    assert.ok(commands.includes("vectora.explainCode"));
    assert.ok(commands.includes("vectora.refactorCode"));
  });

  test("Chat view should be registered", async () => {
    try {
      await vscode.commands.executeCommand("vectora.chatView.focus");
      assert.ok(true);
    } catch (e) {
      assert.fail("Chat view was not registered correctly");
    }
  });
});
```

### [extensions/vscode/src/test/suite/index.ts](file:///c:/Users/bruno/Desktop/Vectora/extensions/vscode/src/test/suite/index.ts)

```typescript
import * as path from "path";
import Mocha from "mocha";
import { glob } from "glob";

export async function run(): Promise<void> {
  const mocha = new Mocha({
    ui: "tdd",
    color: true,
  });

  const testsRoot = __dirname;
  const testFile = path.resolve(testsRoot, "extension.test.js");

  console.log(`[TEST] Suite Root: ${testsRoot}`);
  console.log(`[TEST] Adding file: ${testFile}`);

  if (require("fs").existsSync(testFile)) {
    mocha.addFile(testFile);
  } else {
    throw new Error(`Test file not found at: ${testFile}`);
  }

  try {
    return new Promise((c, e) => {
      mocha.run((failures: number) => {
        if (failures > 0) {
          e(new Error(`${failures} tests failed.`));
        } else {
          c();
        }
      });
    });
  } catch (err) {
    console.error(err);
    throw err;
  }
}
```

---

## 4. Proposed Improvements & Open Questions

- [ ] Implement integrated E2E tests between the VS Code Extension and the Go Core.
- [ ] Add more comprehensive unit tests for the RAG engine and specific MCP integrations.
- [ ] Explore automated UI testing for the Chat Panel webview.
- [ ] Establish a CI action to run all test suites on every pull request.
- [ ] Decide on a shared testing data format for multi-language components.
