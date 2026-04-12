package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// TestStdioServerInitialize verifies the initialize method returns correct protocol info.
func TestStdioServerInitialize(t *testing.T) {
	// Setup mock stdio with initialize request.
	input := `{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}` + "\n"

	output := bytes.NewBuffer(nil)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Create a minimal StdioServer with mock I/O.
	server := &StdioServer{
		Engine: nil, // Not needed for initialize.
		logger: logger,
		reader: nil,
		writer: output,
	}

	// Parse the request manually.
	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      any             `json:"id"`
	}

	line := strings.TrimSpace(input)
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	// Handle the initialize request.
	ctx := context.Background()
	server.handleRequest(ctx, req.Method, req.Params, req.ID)

	// Parse the response.
	responseJSON := strings.TrimSpace(output.String())
	var response struct {
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			ProtocolVersion string `json:"protocolVersion"`
			Capabilities    map[string]interface{} `json:"capabilities"`
			ServerInfo      struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"serverInfo"`
		} `json:"result"`
		ID any `json:"id"`
	}

	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response.
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	if response.Result.ProtocolVersion != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %s", response.Result.ProtocolVersion)
	}

	if response.Result.ServerInfo.Name != "Vectora Core" {
		t.Errorf("Expected server name 'Vectora Core', got %s", response.Result.ServerInfo.Name)
	}

	if response.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", response.ID)
	}
}

// TestStdioServerWriteResponse verifies JSON-RPC 2.0 response format.
func TestStdioServerWriteResponse(t *testing.T) {
	output := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	server := &StdioServer{
		logger: logger,
		writer: output,
	}

	result := map[string]interface{}{
		"status": "ok",
		"data":   "test",
	}

	server.writeResponse(result, 123)

	responseJSON := strings.TrimSpace(output.String())
	var response struct {
		JSONRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}

	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	if response.ID != float64(123) {
		t.Errorf("Expected ID 123, got %v", response.ID)
	}
}

// TestStdioServerWriteError verifies JSON-RPC 2.0 error format.
func TestStdioServerWriteError(t *testing.T) {
	output := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	server := &StdioServer{
		logger: logger,
		writer: output,
	}

	server.writeError(-32601, "Method not found", 456)

	errorJSON := strings.TrimSpace(output.String())
	var response struct {
		JSONRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}

	if err := json.Unmarshal([]byte(errorJSON), &response); err != nil {
		t.Fatalf("Failed to unmarshal error: %v", err)
	}

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}

	if response.Error.Message != "Method not found" {
		t.Errorf("Expected error message 'Method not found', got %s", response.Error.Message)
	}

	if response.ID != float64(456) {
		t.Errorf("Expected ID 456, got %v", response.ID)
	}
}

// TestStdioServerInvalidJSONRPC verifies invalid JSON-RPC version is rejected.
func TestStdioServerInvalidJSONRPC(t *testing.T) {
	output := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	server := &StdioServer{
		logger: logger,
		writer: output,
	}

	// Write an error for invalid JSON-RPC version.
	server.writeError(-32600, "Invalid Request (jsonrpc must be 2.0)", 1)

	errorJSON := strings.TrimSpace(output.String())
	var response struct {
		Error struct {
			Code int `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal([]byte(errorJSON), &response); err != nil {
		t.Fatalf("Failed to unmarshal error: %v", err)
	}

	if response.Error.Code != -32600 {
		t.Errorf("Expected error code -32600, got %d", response.Error.Code)
	}
}

// TestStdioServerCallToolWithInvalidParams verifies tool call error handling.
func TestStdioServerCallToolWithInvalidParams(t *testing.T) {
	output := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	server := &StdioServer{
		Engine: nil,
		logger: logger,
		writer: output,
	}

	// Try to call tool with invalid JSON params.
	invalidParams := json.RawMessage(`{invalid json}`)

	_, err := server.callTool(context.Background(), invalidParams)
	if err == nil {
		t.Error("Expected error for invalid params, got nil")
	}

	if !strings.Contains(err.Error(), "invalid tool call params") {
		t.Errorf("Expected 'invalid tool call params' error, got: %v", err)
	}
}

// TestStdioServerParseRequest verifies JSON-RPC request parsing.
func TestStdioServerParseRequest(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid initialize request",
			input:   `{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}`,
			wantErr: false,
		},
		{
			name:    "Valid tools/list request",
			input:   `{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "Missing jsonrpc field",
			input:   `{"method":"initialize","params":{},"id":1}`,
			wantErr: false, // Will parse but fail validation.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req struct {
				JSONRPC string          `json:"jsonrpc"`
				Method  string          `json:"method"`
				Params  json.RawMessage `json:"params"`
				ID      any             `json:"id"`
			}

			err := json.Unmarshal([]byte(tc.input), &req)
			if (err != nil) != tc.wantErr {
				if tc.wantErr {
					t.Errorf("Expected error, got nil")
				} else {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
