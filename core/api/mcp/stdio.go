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
	"time"

	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/policies"
	"github.com/Kaffyn/Vectora/core/tools"
)

// StdioServer implements the Model Context Protocol (MCP) over stdin/stdout.
// This allows Claude Code and other agents to invoke Vectora as a sub-agent.
// Communicates via newline-delimited JSON-RPC 2.0 messages.
type StdioServer struct {
	Engine *engine.Engine
	Logger *slog.Logger

	Reader *bufio.Reader
	Writer io.Writer
	Mu     sync.Mutex
}

// NewStdioServer creates a new MCP server that communicates via stdin/stdout.
func NewStdioServer(eng *engine.Engine, logger *slog.Logger) *StdioServer {
	return &StdioServer{
		Engine: eng,
		Logger: logger,
		Reader: bufio.NewReader(os.Stdin),
		Writer: os.Stdout,
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
	s.Logger.Info("MCP Server started via stdio",
		slog.String("protocol", "JSON-RPC 2.0"),
		slog.String("transport", "stdio"),
	)

	for {
		select {
		case <-ctx.Done():
			s.Logger.Info("MCP Server: context cancelled, shutting down gracefully")
			return ctx.Err()
		default:
		}

		// Read a line from stdin (newline-delimited JSON).
		line, err := s.Reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.Logger.Info("MCP Server: stdin closed, shutting down")
				return nil
			}
			s.Logger.Error("MCP Server: read error",
				slog.Any("error", err),
				slog.String("error_type", fmt.Sprintf("%T", err)),
			)
			return err
		}

		// Debug logging for request
		if len(line) > 500 {
			s.Logger.Debug("MCP Server: received request",
				slog.String("request", line[:100]+"..."),
				slog.Int("total_size", len(line)),
			)
		} else {
			s.Logger.Debug("MCP Server: received request", slog.String("request", line))
		}

		// Parse JSON-RPC 2.0 request.
		var req struct {
			JSONRPC string          `json:"jsonrpc"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
			ID      any             `json:"id"`
		}

		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.Logger.Warn("MCP Server: JSON parse error",
				slog.String("error", err.Error()),
				slog.String("input", line),
			)
			s.WriteError(-32700, fmt.Sprintf("Parse error: invalid JSON - %v", err), nil)
			continue
		}

		// Validate JSON-RPC version.
		if req.JSONRPC != "2.0" {
			s.Logger.Warn("MCP Server: invalid jsonrpc version",
				slog.String("version", req.JSONRPC),
				slog.Any("id", req.ID),
			)
			s.WriteError(-32600, fmt.Sprintf("Invalid Request: jsonrpc must be '2.0', got '%s'", req.JSONRPC), req.ID)
			continue
		}

		// Validate method is present
		if req.Method == "" {
			s.Logger.Warn("MCP Server: missing method in request", slog.Any("id", req.ID))
			s.WriteError(-32600, "Invalid Request: method is required", req.ID)
			continue
		}

		s.Logger.Debug("MCP Server: processing method",
			slog.String("method", req.Method),
			slog.Any("id", req.ID),
		)

		// Process the request.
		s.HandleRequest(ctx, req.Method, req.Params, req.ID)
	}
}

// HandleRequest processes a single JSON-RPC 2.0 request method.
func (s *StdioServer) HandleRequest(ctx context.Context, method string, params json.RawMessage, id any) {
	var result any
	var err error

	switch method {
	case "initialize":
		// MCP initialization response.
		s.Logger.Info("MCP: initialize request", slog.Any("id", id))
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
		s.Logger.Info("MCP: tools/list request", slog.Any("id", id))
		result = s.ListTools()

	case "tools/call":
		// Execute a tool.
		s.Logger.Info("MCP: tools/call request", slog.Any("id", id))
		result, err = s.CallTool(ctx, params)

	case "completion/complete":
		// Optional: text completion (not implemented yet).
		s.Logger.Warn("MCP: unimplemented method requested",
			slog.String("method", method),
			slog.Any("id", id),
		)
		err = fmt.Errorf("method not implemented: completion/complete (Vectora focuses on tools)")

	default:
		s.Logger.Warn("MCP: unknown method",
			slog.String("method", method),
			slog.Any("id", id),
		)
		s.WriteError(-32601, fmt.Sprintf("Method not found: '%s' is not a valid MCP method. Valid methods: initialize, tools/list, tools/call", method), id)
		return
	}

	if err != nil {
		s.Logger.Error("MCP: request failed",
			slog.String("method", method),
			slog.Any("id", id),
			slog.String("error", err.Error()),
		)
		s.WriteError(-32000, fmt.Sprintf("Server error in %s: %v", method, err), id)
		return
	}

	s.Logger.Debug("MCP: request succeeded",
		slog.String("method", method),
		slog.Any("id", id),
	)
	s.WriteResponse(result, id)
}

// ListTools returns only embedding-related tools in MCP format.
// This prevents duplication of file system tools when Vectora is as a sub-agent.
func (s *StdioServer) ListTools() map[string]any {
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

// CallTool executes a tool by name with the given arguments.
// Returns detailed error information to help with debugging.
func (s *StdioServer) CallTool(ctx context.Context, params json.RawMessage) (any, error) {
	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"` // MCP spec usa "arguments"
		Input     map[string]interface{} `json:"input"`     // fallback legado
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid tool call params: cannot parse JSON - %w", err)
	}

	// Validate required fields
	if req.Name == "" {
		return nil, fmt.Errorf("invalid tool call: name is required")
	}

	// MCP spec usa "arguments"; fallback para "input" (legado)
	toolArgs := req.Arguments
	if len(toolArgs) == 0 {
		toolArgs = req.Input
	}

	s.Logger.Debug("MCP: executing tool",
		slog.String("tool_name", req.Name),
		slog.Any("input_keys", getMapKeys(toolArgs)),
	)

	// Check if tool exists before attempting execution
	allTools := s.Engine.Tools.GetAll()
	toolExists := false
	for _, t := range allTools {
		if t.Name() == req.Name {
			toolExists = true
			break
		}
	}

	if !toolExists {
		availableTools := make([]string, 0)
		for _, t := range allTools {
			availableTools = append(availableTools, t.Name())
		}
		return nil, fmt.Errorf("tool not found: '%s'. Available tools: %v", req.Name, availableTools)
	}

	// Execute tool via engine with timeout context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := s.Engine.ExecuteTool(ctxWithTimeout, req.Name, toolArgs)
	if err != nil {
		s.Logger.Error("MCP: tool execution error",
			slog.String("tool_name", req.Name),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("tool '%s' execution failed: %w", req.Name, err)
	}

	s.Logger.Debug("MCP: tool execution succeeded",
		slog.String("tool_name", req.Name),
		slog.Bool("is_error", result.IsError),
	)

	return map[string]any{
		"name":    req.Name,
		"output":  result.Output,
		"isError": result.IsError,
	}, nil
}

// getMapKeys returns the keys of a map as a slice of strings
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// WriteResponse writes a successful JSON-RPC 2.0 response to stdout.
// Ensures thread-safe output and logs response details for debugging.
func (s *StdioServer) WriteResponse(result any, id any) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	response := map[string]any{
		"jsonrpc": "2.0",
		"result":  result,
		"id":      id,
	}

	data, err := json.Marshal(response)
	if err != nil {
		s.Logger.Error("MCP: failed to marshal response",
			slog.Any("id", id),
			slog.String("error", err.Error()),
		)
		return
	}

	s.Logger.Debug("MCP: writing response",
		slog.Any("id", id),
		slog.Int("response_size", len(data)),
	)

	_, writeErr := fmt.Fprintf(s.Writer, "%s\n", data)
	if writeErr != nil {
		s.Logger.Error("MCP: failed to write response to stdout",
			slog.Any("id", id),
			slog.String("error", writeErr.Error()),
		)
	}
}

// WriteError writes a JSON-RPC 2.0 error response to stdout.
// Includes error code, message, and request ID for client correlation.
// Ensures thread-safe output and logs error details for debugging.
func (s *StdioServer) WriteError(code int, message string, id any) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	response := map[string]any{
		"jsonrpc": "2.0",
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
		"id": id,
	}

	data, err := json.Marshal(response)
	if err != nil {
		s.Logger.Error("MCP: failed to marshal error response",
			slog.Any("id", id),
			slog.Int("error_code", code),
			slog.String("error", err.Error()),
		)
		return
	}

	s.Logger.Debug("MCP: writing error response",
		slog.Any("id", id),
		slog.Int("error_code", code),
		slog.String("error_message", message),
	)

	_, writeErr := fmt.Fprintf(s.Writer, "%s\n", data)
	if writeErr != nil {
		s.Logger.Error("MCP: failed to write error response to stdout",
			slog.Any("id", id),
			slog.String("error", writeErr.Error()),
		)
	}
}
