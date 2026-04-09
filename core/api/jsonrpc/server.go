// Package jsonrpc implements a minimal, dependency-free JSON-RPC 2.0 server.
// It handles message framing (newline-delimited JSON) and method routing.
// This is used for MCP and ACP bridges over stdio and TCP.
package jsonrpc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"` // Null for notifications
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int64      `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error object.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// NewError creates a new JSON-RPC error.
func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// NewErrorWithData creates a new JSON-RPC error with data payload.
func NewErrorWithData(code int, message string, data interface{}) *Error {
	return &Error{Code: code, Message: message, Data: data}
}

// HandlerFunc defines the signature for method handlers.
// It receives the raw params and returns the result object or an error.
type HandlerFunc func(params json.RawMessage) (interface{}, error)

// Server implements a JSON-RPC 2.0 server.
type Server struct {
	handlers map[string]HandlerFunc
	mu       sync.RWMutex
}

// NewServer creates a new JSON-RPC server instance.
func NewServer() *Server {
	return &Server{
		handlers: make(map[string]HandlerFunc),
	}
}

// RegisterMethod registers a handler for a specific method name.
func (s *Server) RegisterMethod(method string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// HandleRequest processes a single JSON-RPC request and returns the response.
func (s *Server) HandleRequest(req Request) Response {
	if req.JSONRPC != "2.0" {
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   NewError(-32600, "Invalid Request"),
		}
	}

	s.mu.RLock()
	handler, ok := s.handlers[req.Method]
	s.mu.RUnlock()

	if !ok {
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   NewError(-32601, fmt.Sprintf("Method not found: %s", req.Method)),
		}
	}

	result, err := handler(req.Params)
	if err != nil {
		// Check if it's already a JSON-RPC error
		var rpcErr *Error
		if errors.As(err, &rpcErr) {
			return Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   rpcErr,
			}
		}
		// Generic internal error
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   NewError(-32000, err.Error()),
		}
	}

	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// ServeStdio starts the server loop using standard input and output.
// This is used for MCP and ACP bridges.
func (s *Server) ServeStdio() error {
	reader := bufio.NewReader(io.LimitReader(os.Stdin, 10*1024*1024)) // 10MB limit per line
	writer := os.Stdout

	scanner := bufio.NewScanner(reader)
	// Increase buffer size for large payloads (e.g., file contents)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			resp := Response{
				JSONRPC: "2.0",
				ID:      nil, // Parse errors usually don't have ID if we can't parse it
				Error:   NewError(-32700, "Parse error"),
			}
			writeResponse(writer, resp)
			continue
		}

		resp := s.HandleRequest(req)

		// Notifications (no ID) do not expect a response
		if req.ID != nil {
			if err := writeResponse(writer, resp); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

// ServeTCP starts the server loop on a TCP listener.
// Useful for debugging with tools like curl or postman.
func (s *Server) ServeTCP(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			resp := Response{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   NewError(-32700, "Parse error"),
			}
			writeResponse(conn, resp)
			continue
		}

		resp := s.HandleRequest(req)
		if req.ID != nil {
			writeResponse(conn, resp)
		}
	}
}

func writeResponse(w io.Writer, resp Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = w.Write(append(data, '\n'))
	return err
}
