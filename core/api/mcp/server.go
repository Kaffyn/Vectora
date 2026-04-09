package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kaffyn/Vectora/core/engine"
)

// Server implements the Model Context Protocol (MCP) over HTTP.
type Server struct {
	Engine *engine.Engine
}

func NewServer(eng *engine.Engine) *Server {
	return &Server{Engine: eng}
}

// Start runs the HTTP server on the given address.
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", s.handleRPC)
	mux.HandleFunc("/mcp/v1/tools", s.handleListTools)
	mux.HandleFunc("/mcp/v1/call", s.handleCallTool)

	fmt.Printf("Vectora MCP Server listening on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      any             `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, -32700, "Parse error", nil)
		return
	}

	var result any
	var err error

	switch req.Method {
	case "initialize":
		result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]string{"name": "Vectora Core", "version": "0.1.0"},
		}
	case "tools/list":
		result = s.listTools()
	case "tools/call":
		result, err = s.callTool(r.Context(), req.Params)
	default:
		s.writeError(w, -32601, "Method not found", req.ID)
		return
	}

	if err != nil {
		s.writeError(w, -32000, err.Error(), req.ID)
		return
	}

	s.writeResponse(w, result, req.ID)
}

func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	s.writeResponse(w, s.listTools(), nil)
}

func (s *Server) handleCallTool(w http.ResponseWriter, r *http.Request) {
	var params json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := s.callTool(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeResponse(w, result, nil)
}

func (s *Server) listTools() any {
	allTools := s.Engine.Tools.GetAll()
	mcpTools := make([]map[string]any, 0, len(allTools))
	for _, t := range allTools {
		mcpTools = append(mcpTools, map[string]any{
			"name":        t.Name(),
			"description": t.Description(),
			"inputSchema": t.Schema(),
		})
	}
	return map[string]any{"tools": mcpTools}
}

func (s *Server) callTool(ctx context.Context, params json.RawMessage) (any, error) {
	var req struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	res, err := s.Engine.ExecuteTool(ctx, req.Name, req.Arguments)
	if err != nil {
		return nil, err
	}

	// MCP expects content array
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": res.Output},
		},
		"isError": res.IsError,
	}, nil
}

func (s *Server) writeResponse(w http.ResponseWriter, result any, id any) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"jsonrpc": "2.0",
		"result":  result,
	}
	if id != nil {
		resp["id"] = id
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) writeError(w http.ResponseWriter, code int, message string, id any) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"jsonrpc": "2.0",
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
	if id != nil {
		resp["id"] = id
	}
	json.NewEncoder(w).Encode(resp)
}
