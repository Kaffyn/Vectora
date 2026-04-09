package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// mockEngine implements Engine for testing
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

func TestACPInitialize(t *testing.T) {
	engine := &mockEngine{}
	server := NewServer(engine)

	// Send initialize request
	input := `{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":1,"clientCapabilities":{"fs":{"readTextFile":true,"writeTextFile":true},"terminal":true},"clientInfo":{"name":"test-client","title":"Test Client","version":"1.0.0"}}}`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse the input and test the handler directly
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

	// Create session first
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

	// Create session
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

	// 1. Initialize
	initReq := InitializeRequest{
		ProtocolVersion: 1,
		ClientInfo:      &Info{Name: "test", Title: "Test Client", Version: "1.0.0"},
	}
	initResult, _ := server.handleInitialize(ctx, toJSON(t, initReq))
	initResp := initResult.(InitializeResponse)
	if initResp.ProtocolVersion != 1 {
		t.Fatalf("version mismatch: got %d", initResp.ProtocolVersion)
	}

	// 2. session/new
	newReq := SessionNewRequest{CWD: "/test/project"}
	newResult, _ := server.handleSessionNew(ctx, toJSON(t, newReq))
	sessionResp := newResult.(SessionNewResponse)
	if sessionResp.SessionID == "" {
		t.Fatal("no session ID")
	}

	// 3. session/prompt
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
