package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestACPInitialize(t *testing.T) {
	engine := NewStatefulMockEngine()
	server := NewServer(engine)

	// Send initialize request
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

	fmt.Printf("✅ Initialize response: %+v\n", resp)
}

func TestACPSessionNew(t *testing.T) {
	engine := NewStatefulMockEngine()
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

	fmt.Printf("✅ Session created: %s\n", resp.SessionID)
}

func TestACPFSRead(t *testing.T) {
	engine := NewStatefulMockEngine()
	server := NewServer(engine)

	// Inject file directly into stateful mock
	engine.WriteFile(context.Background(), "/test/file.go", "package main\n\nfunc main() {}")

	// Create session
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
	if !strings.Contains(resp.Content, "package main") {
		t.Errorf("expected file content, got '%s'", resp.Content)
	}

	fmt.Printf("✅ FSRead content: %s\n", resp.Content)
}

func TestStatefulPersistenceFlow(t *testing.T) {
	engine := NewStatefulMockEngine()
	server := NewServer(engine)
	ctx := context.Background()

	server.sessions["sess_test"] = &Session{ID: "sess_test", CWD: "/test"}

	// 1. Write File via ACP handler
	writeInput := `{"sessionId":"sess_test","path":"/test/persistent.go","content":"package persistence\n\nconst ID = 42"}`
	var writeParams map[string]any
	json.Unmarshal([]byte(writeInput), &writeParams)
	_, errMsg := server.handleFSWrite(ctx, toJSON(t, writeParams))
	if errMsg != "" {
		t.Fatalf("fs/write failed: %s", errMsg)
	}

	// 2. Read File back and verify
	readInput := `{"sessionId":"sess_test","path":"/test/persistent.go"}`
	var readParams map[string]any
	json.Unmarshal([]byte(readInput), &readParams)
	readResult, errMsg := server.handleFSRead(ctx, toJSON(t, readParams))
	if errMsg != "" {
		t.Fatalf("fs/read failed: %s", errMsg)
	}
	readResp := readResult.(FSReadResponse)
	if !strings.Contains(readResp.Content, "const ID = 42") {
		t.Errorf("expected persistent content, got %q", readResp.Content)
	}

	// 3. Search via Grep handler
	grepInput := `{"pattern":"persistence","path":"/test"}`
	var grepParams map[string]any
	json.Unmarshal([]byte(grepInput), &grepParams)
	grepResult, errMsg := server.handleGrepSearch(ctx, toJSON(t, grepParams))
	if errMsg != "" {
		t.Fatalf("grep failed: %s", errMsg)
	}
	grepResp := grepResult.(map[string]string)
	if !strings.Contains(grepResp["output"], "persistent.go") {
		t.Errorf("expected grep to find file, got %q", grepResp["output"])
	}

	fmt.Println("✅ Stateful persistence flow passed: Write → Read → Grep")
}

func TestACPPrompt(t *testing.T) {
	engine := NewStatefulMockEngine()
	server := NewServer(engine)

	server.sessions["sess_test"] = &Session{
		ID:           "sess_test",
		CWD:          "/test",
		Updates:      make(chan SessionUpdate, 100),
		PermissionCh: make(chan PermissionResponse, 1),
	}

	// Test keyword-based response from stateful mock
	input := `{"sessionId":"sess_test","prompt":[{"type":"text","text":"Say hello to me!"}]}`
	var params map[string]any
	json.Unmarshal([]byte(input), &params)
	paramsJSON, _ := json.Marshal(params)

	result, errMsg := server.handleSessionPrompt(context.Background(), paramsJSON)
	if errMsg != "" {
		t.Fatalf("session/prompt failed: %s", errMsg)
	}

	resp := result.(PromptResponse)
	if resp.StopReason != StopEndTurn {
		t.Errorf("expected stop reason 'end_turn', got '%s'", resp.StopReason)
	}

	fmt.Printf("✅ Prompt completed with stop reason: %s\n", resp.StopReason)
}

func toJSON(t *testing.T, v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return data
}
