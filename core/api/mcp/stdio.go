package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/engine"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/policies"
	"github.com/Kaffyn/Vectora/core/tools"
)

// StdioServer implements the Model Context Protocol (MCP) over stdin/stdout.
// This allows Claude Code and other agents to invoke Vectora as a sub-agent.
// Communicates via newline-delimited JSON-RPC 2.0 messages.
type StdioServer struct {
	Engine *engine.Engine
	logger *slog.Logger

	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
}

// NewStdioServer creates a new MCP server that communicates via stdin/stdout.
func NewStdioServer(eng *engine.Engine, logger *slog.Logger) *StdioServer {
	return &StdioServer{
		Engine: eng,
		logger: logger,
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// NewStdioServerFromMCP creates a StdioServer from MCP components.
// This is used during startup to initialize the MCP server with all dependencies.
func NewStdioServerFromMCP(
	mcpServer *VectoraMCPServer,
	kvStore db.KVStore,
	vecStore db.VectorStore,
	router *llm.Router,
	logger *slog.Logger,
) *StdioServer {
	// Create Engine from MCP server's dependencies.
	// Guardian defaults to root path "/" for MCP mode (all tools available).
	guardian := policies.NewGuardian("/")
	toolRegistry := tools.NewRegistry("/", guardian, kvStore)

	eng := engine.NewEngine(
		vecStore.(*db.ChromemStore),
		kvStore.(*db.BBoltStore),
		router,
		toolRegistry,
		guardian,
		nil, // indexer not needed for MCP
	)

	return NewStdioServer(eng, logger)
}

// Start runs the MCP protocol server over stdin/stdout.
// Messages are newline-delimited JSON-RPC 2.0.
// This function blocks until context is cancelled or an error occurs.
func (s *StdioServer) Start(ctx context.Context) error {
	s.logger.Info("MCP Server started via stdio")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read a line from stdin (newline-delimited JSON).
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.logger.Info("MCP Server: stdin closed")
				return nil
			}
			s.logger.Error("MCP Server: read error", slog.Any("error", err))
			return err
		}

		// Parse JSON-RPC 2.0 request.
		var req struct {
			JSONRPC string          `json:"jsonrpc"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
			ID      any             `json:"id"`
		}

		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.writeError(-32700, "Parse error", nil)
			continue
		}

		// Validate JSON-RPC version.
		if req.JSONRPC != "2.0" {
			s.writeError(-32600, "Invalid Request (jsonrpc must be 2.0)", req.ID)
			continue
		}

		// Process the request.
		s.handleRequest(ctx, req.Method, req.Params, req.ID)
	}
}

// handleRequest processes a single JSON-RPC 2.0 request method.
func (s *StdioServer) handleRequest(ctx context.Context, method string, params json.RawMessage, id any) {
	var result any
	var err error

	switch method {
	case "initialize":
		// MCP initialization response.
		result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]string{
				"name":    "Vectora Core",
				"version": "0.1.0",
			},
		}

	case "tools/list":
		// List all available tools.
		result = s.listTools()

	case "tools/call":
		// Execute a tool.
		result, err = s.callTool(ctx, params)

	case "completion/complete":
		// Optional: text completion (not implemented yet).
		err = fmt.Errorf("method not implemented: %s", method)

	default:
		s.writeError(-32601, fmt.Sprintf("Method not found: %s", method), id)
		return
	}

	if err != nil {
		s.writeError(-32000, fmt.Sprintf("Server error: %v", err), id)
		return
	}

	s.writeResponse(result, id)
}

// listTools returns only embedding-related tools in MCP format.
// This prevents duplication of file system tools when Vectora is as a sub-agent.
func (s *StdioServer) listTools() map[string]any {
	allTools := s.Engine.Tools.GetAll()
	mcpTools := make([]map[string]any, 0)

	// Whitelist of embedding and deep analysis tools
	whitelist := map[string]bool{
		"embed":                    true,
		"search_database":          true,
		"web_search_and_embed":     true,
		"web_fetch_and_embed":      true,
		"plan_mode":                true,
		"refactor_with_context":    true,
		"analyze_code_patterns":    true,
		"knowledge_graph_analysis": true,
		"doc_coverage_analysis":    true,
		"test_generation":          true,
		"bug_pattern_detection":    true,
		"rag_plan":                 true, // Future tool
	}

	for _, t := range allTools {
		if whitelist[t.Name()] {
			mcpTools = append(mcpTools, map[string]any{
				"name":        t.Name(),
				"description": t.Description(),
				"inputSchema": json.RawMessage(t.Schema()),
			})
		}
	}

	return map[string]any{
		"tools": mcpTools,
	}
}

// callTool executes a tool by name with the given arguments.
func (s *StdioServer) callTool(ctx context.Context, params json.RawMessage) (any, error) {
	var req struct {
		Name  string                 `json:"name"`
		Input map[string]interface{} `json:"input"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid tool call params: %w", err)
	}

	// Execute tool via engine
	result, err := s.Engine.ExecuteTool(ctx, req.Name, req.Input)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return map[string]any{
		"name":    req.Name,
		"output":  result.Output,
		"isError": result.IsError,
	}, nil
}

// writeResponse writes a successful JSON-RPC 2.0 response to stdout.
func (s *StdioServer) writeResponse(result any, id any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := map[string]any{
		"jsonrpc": "2.0",
		"result":  result,
		"id":      id,
	}

	data, _ := json.Marshal(response)
	fmt.Fprintf(s.writer, "%s\n", data)
}

// writeError writes a JSON-RPC 2.0 error response to stdout.
func (s *StdioServer) writeError(code int, message string, id any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := map[string]any{
		"jsonrpc": "2.0",
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
		"id": id,
	}

	data, _ := json.Marshal(response)
	fmt.Fprintf(s.writer, "%s\n", data)
}
