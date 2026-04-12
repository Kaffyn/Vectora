// Package handlers contains the business logic for JSON-RPC methods.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kaffyn/Vectora/core/api/jsonrpc"
	"github.com/Kaffyn/Vectora/core/engine"
)

// InitRequest represents the initialize params.
type InitRequest struct {
	ProtocolVersion string `json:"protocolVersion"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

// InitResponse represents the initialize result.
type InitResponse struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	ServerInfo      map[string]string      `json:"serverInfo"`
	Capabilities    map[string]interface{} `json:"capabilities"`
}

// HandleInitialize processes the initialize method.
func HandleInitialize(params json.RawMessage) (interface{}, error) {
	var req InitRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewError(-32602, "Invalid initialize params")
	}

	return InitResponse{
		ProtocolVersion: "2.0",
		ServerInfo: map[string]string{
			"name":    "vectora-core",
			"version": "0.1.0",
		},
		Capabilities: map[string]interface{}{
			"tools":           true,
			"guardianEnabled": true,
		},
	}, nil
}

// ToolCallRequest represents a generic tool call.
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResponse represents the result of a tool call.
type ToolCallResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError,omitempty"`
}

// HandleToolsCall processes a tool call request.
func HandleToolsCall(eng *engine.Engine, params json.RawMessage) (interface{}, error) {
	var req ToolCallRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewError(-32602, "Invalid params")
	}

	// Execute tool via engine
	result, err := eng.ExecuteTool(context.Background(), req.Name, req.Arguments)
	if err != nil {
		return nil, jsonrpc.NewError(-32000, err.Error())
	}

	if result.IsError {
		return nil, jsonrpc.NewErrorWithData(-32001, result.Output, map[string]string{
			"tool": req.Name,
		})
	}

	return ToolCallResponse{
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{{Type: "text", Text: result.Output}},
	}, nil
}

// HandleToolsList returns the list of available tools.
func HandleToolsList() (interface{}, error) {
	// Return a placeholder — tools are dynamic based on engine configuration
	return map[string]interface{}{
		"tools": []map[string]string{
			{"name": "read_file", "description": "Read file content"},
			{"name": "write_file", "description": "Write file content"},
			{"name": "grep_search", "description": "Search file contents"},
			{"name": "run_shell_command", "description": "Execute shell command"},
			{"name": "save_memory", "description": "Save persistent memory"},
		},
	}, nil
}

// SessionNewRequest represents a session/new request.
type SessionNewRequest struct {
	CWD        string      `json:"cwd"`
	MCPServers []MCPServer `json:"mcpServers,omitempty"`
}

// MCPServer represents an MCP server configuration.
type MCPServer struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Env     []string `json:"env,omitempty"`
}

// HandleNewSession creates a new ACP session.
func HandleNewSession(eng *engine.Engine, params json.RawMessage) (interface{}, error) {
	var req SessionNewRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewError(-32602, "Invalid session params")
	}

	// Generate a session ID (in production, this would be managed by the engine)
	sessionID := fmt.Sprintf("sess_%s", eng.GetStatus())

	return map[string]string{
		"sessionId": sessionID,
	}, nil
}

// SessionPromptRequest represents a session/prompt request.
type SessionPromptRequest struct {
	SessionID string         `json:"sessionId"`
	Prompt    []ContentBlock `json:"prompt"`
}

// ContentBlock represents a content block in a prompt.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// HandleSessionPrompt processes a prompt request.
func HandleSessionPrompt(eng *engine.Engine, params json.RawMessage) (interface{}, error) {
	var req SessionPromptRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, jsonrpc.NewError(-32602, "Invalid prompt params")
	}

	// Extract text from prompt blocks
	var queryText string
	for _, block := range req.Prompt {
		if block.Type == "text" {
			queryText += block.Text
		}
	}

	if queryText == "" {
		return nil, jsonrpc.NewError(-32602, "Empty prompt")
	}

	// Execute query via engine using Query instead of ProcessQuery (which might be deprecated/renamed)
	answer, _, err := eng.Query(context.Background(), queryText, "default", "gemini-3-flash-preview", "fast", "ask")
	if err != nil {
		return nil, jsonrpc.NewError(-32000, err.Error())
	}

	return map[string]string{
		"stopReason": "end_turn",
		"answer":     answer,
	}, nil
}
