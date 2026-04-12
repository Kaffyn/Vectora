package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
)

// AdapterHandler converts old-style HandlerFunc to sourcegraph/jsonrpc2 compatible handler
// This maintains backward compatibility during migration from custom JSON-RPC to sourcegraph/jsonrpc2
func AdapterHandler(oldHandler HandlerFunc) jsonrpc2.Handler {
	return func(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
		result, err := oldHandler(req.Params)

		// If error is already a JSON-RPC error, return it directly
		var rpcErr *Error
		if err != nil && !json.As(err, &rpcErr) {
			// Wrap non-JSON-RPC errors as internal errors
			return nil, &jsonrpc2.Error{
				Code:    -32000,
				Message: err.Error(),
			}
		}

		if rpcErr != nil {
			return nil, &jsonrpc2.Error{
				Code:    int64(rpcErr.Code),
				Message: rpcErr.Message,
				Data:    rpcErr.Data,
			}
		}

		return result, nil
	}
}

// ConvertError converts our Error type to jsonrpc2.Error
func ConvertError(err *Error) *jsonrpc2.Error {
	if err == nil {
		return nil
	}
	return &jsonrpc2.Error{
		Code:    int64(err.Code),
		Message: err.Message,
		Data:    err.Data,
	}
}

// ConvertJSONRPC2Error converts jsonrpc2.Error back to our Error type
func ConvertJSONRPC2Error(err *jsonrpc2.Error) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    int(err.Code),
		Message: err.Message,
		Data:    err.Data,
	}
}

// HandlerRegistry wraps jsonrpc2 handler map for compatibility
type HandlerRegistry struct {
	handlers map[string]jsonrpc2.Handler
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]jsonrpc2.Handler),
	}
}

// Register adds a handler using old HandlerFunc signature
func (r *HandlerRegistry) Register(method string, handler HandlerFunc) {
	r.handlers[method] = AdapterHandler(handler)
}

// RegisterNew adds a handler using new jsonrpc2.Handler signature
func (r *HandlerRegistry) RegisterNew(method string, handler jsonrpc2.Handler) {
	r.handlers[method] = handler
}

// Get retrieves a handler
func (r *HandlerRegistry) Get(method string) (jsonrpc2.Handler, bool) {
	h, ok := r.handlers[method]
	return h, ok
}

// GetAll returns all handlers
func (r *HandlerRegistry) GetAll() map[string]jsonrpc2.Handler {
	return r.handlers
}

// LegacyServerBridge provides compatibility layer between custom server and jsonrpc2
// Used for gradual migration without breaking existing code
type LegacyServerBridge struct {
	*Server
	registry *HandlerRegistry
}

// NewLegacyServerBridge creates a bridge that supports both old and new handlers
func NewLegacyServerBridge() *LegacyServerBridge {
	return &LegacyServerBridge{
		Server:   NewServer(),
		registry: NewHandlerRegistry(),
	}
}

// RegisterMethod maintains backward compatibility by registering in both places
func (b *LegacyServerBridge) RegisterMethod(method string, handler HandlerFunc) {
	b.Server.RegisterMethod(method, handler)
	b.registry.Register(method, handler)
}

// RegisterMethodNew registers a handler using new jsonrpc2 signature
func (b *LegacyServerBridge) RegisterMethodNew(method string, handler jsonrpc2.Handler) {
	// For new-style handlers, wrap to old signature for compatibility
	oldHandler := func(params json.RawMessage) (interface{}, error) {
		req := &jsonrpc2.Request{
			Params: params,
			Method: method,
		}
		return handler(context.Background(), req)
	}
	b.Server.RegisterMethod(method, oldHandler)
	b.registry.RegisterNew(method, handler)
}

// GetHandler retrieves from new registry first, falls back to old server
func (b *LegacyServerBridge) GetHandler(method string) (jsonrpc2.Handler, bool) {
	if h, ok := b.registry.Get(method); ok {
		return h, true
	}

	// Fallback to checking old server (for debugging)
	if _, ok := b.Server.handlers[method]; ok {
		// Create adapter for the old handler
		oldHandler := b.Server.handlers[method]
		return AdapterHandler(oldHandler), true
	}

	return nil, false
}

// MetadataAnnotation provides metadata for handlers
type MetadataAnnotation struct {
	Method      string
	Description string
	IsStreaming bool
	IsPrivate   bool
}

// MethodMetadata stores metadata for all methods
var MethodMetadata = map[string]MetadataAnnotation{
	"workspace.query":        {Method: "workspace.query", Description: "Vector search with RAG", IsStreaming: true},
	"chat.history":           {Method: "chat.history", Description: "Retrieve conversation history"},
	"provider.get":           {Method: "provider.get", Description: "Get LLM provider status"},
	"app.health":             {Method: "app.health", Description: "Health check"},
	"workspace.embed.start":  {Method: "workspace.embed.start", Description: "Start file embedding", IsStreaming: true},
	"ipc.auth":               {Method: "ipc.auth", Description: "Authenticate with token", IsPrivate: true},
	"workspace.init":         {Method: "workspace.init", Description: "Initialize workspace context"},
}

// GetMethodMetadata returns metadata for a method
func GetMethodMetadata(method string) (MetadataAnnotation, bool) {
	m, ok := MethodMetadata[method]
	return m, ok
}

// ValidationMiddleware validates requests before processing
func ValidationMiddleware(next jsonrpc2.Handler) jsonrpc2.Handler {
	return func(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
		// Validate method is not empty
		if req.Method == "" {
			return nil, &jsonrpc2.Error{
				Code:    -32600,
				Message: "Invalid Request: method cannot be empty",
			}
		}

		return next(ctx, req)
	}
}

// ErrorRecoveryMiddleware wraps handlers with panic recovery
func ErrorRecoveryMiddleware(next jsonrpc2.Handler) jsonrpc2.Handler {
	return func(ctx context.Context, req *jsonrpc2.Request) (result interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = &jsonrpc2.Error{
					Code:    -32603,
					Message: fmt.Sprintf("Internal error: %v", r),
				}
			}
		}()

		return next(ctx, req)
	}
}
