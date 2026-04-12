package ipc

import (
	"encoding/json"
)

const FrameDelimiter = '\n'

const (
	MsgTypeRequest  = "request"
	MsgTypeResponse = "response"
	MsgTypeEvent    = "event"
)

// JSON-RPC 2.0 standard error codes (https://www.jsonrpc.org/specification#error_object)
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	// -32000 to -32099 are reserved for server-defined errors
	CodeProviderNotConfigured = -32000
	CodeWorkspaceNotFound     = -32001
	CodeToolNotFound          = -32002
	CodeUnauthorized          = -32003
	CodeServerError           = -32099 // generic server-defined error
)

type IPCMessage struct {
	JSONRPC string          `json:"jsonrpc"` // Always "2.0"
	ID      string          `json:"id"`
	Type    string          `json:"type"`             // Vectora-specific internal use: "request", "response", "event"
	Method  string          `json:"method,omitempty"` // For requests and events
	Params  json.RawMessage `json:"params,omitempty"` // Standard JSON-RPC params
	Result  json.RawMessage `json:"result,omitempty"` // Standard JSON-RPC result
	Payload json.RawMessage `json:"payload,omitempty"` // Vectora legacy/internal payload
	Error   *IPCError       `json:"error,omitempty"`  // Standard JSON-RPC error
}

// IPCError maps to JSON-RPC 2.0 error object. Code is a numeric JSON-RPC
// error code; Slug is a human-readable identifier for programmatic matching.
type IPCError struct {
	Code    int    `json:"code"`
	Slug    string `json:"slug"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

func (e *IPCError) Error() string { return e.Message }

// errServer constructs a server-defined error in the -32000 range with a slug and message.
// Use this for application-level errors (provider not configured, workspace not found, etc.).
// For protocol-level errors use the pre-defined Err* vars directly.
func errServer(slug, message string) *IPCError {
	return &IPCError{Code: CodeServerError, Slug: slug, Message: message}
}

var (
	ErrWorkspaceNotFound = &IPCError{Code: CodeWorkspaceNotFound, Slug: "workspace_not_found", Message: "The requested Workspace does not exist."}
	ErrProviderNotConfig = &IPCError{Code: CodeProviderNotConfigured, Slug: "provider_not_configured", Message: "No LLM provider has been configured."}
	ErrToolNotFound      = &IPCError{Code: CodeToolNotFound, Slug: "tool_not_found", Message: "The provided Agentic Tool is not in the Registry."}
	ErrIPCMethodUnknown  = &IPCError{Code: CodeMethodNotFound, Slug: "method_not_found", Message: "This IPC Endpoint does not exist."}
	ErrIPCPayloadInvalid = &IPCError{Code: CodeInvalidParams, Slug: "invalid_params", Message: "Malformed JSON payload."}
	ErrInternalError     = &IPCError{Code: CodeInternalError, Slug: "internal_error", Message: "Unhandled error in the IPC Backend."}
	ErrUnauthorized      = &IPCError{Code: CodeUnauthorized, Slug: "unauthorized", Message: "Invalid or missing IPC token."}
)
