package ipc

import (
	"encoding/json"
)

// Absolute constant that ensures Socket Stream delimitation (JSON-ND).
const FrameDelimiter = '\n'

// Standard Restricted Types (RPC)
const (
	MsgTypeRequest  = "request"
	MsgTypeResponse = "response"
	MsgTypeEvent    = "event"
)

// IPCMessage is the Universal Transport Unit for all Vectora Sockets.
type IPCMessage struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`             // "request", "response", "event"
	Method  string          `json:"method,omitempty"` // Required only on Type == "request" or "event"
	Payload json.RawMessage `json:"payload"`          // Arbitrary body
	Error   *IPCError       `json:"error,omitempty"`  // Required only if failure exists and Type == "response"
}

// IPCError is the standardized failure return for the Frontend (Node/React).
type IPCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

// Canonical Error Responses (According to Business Rules)
var (
	ErrWorkspaceNotFound   = &IPCError{Code: "workspace_not_found", Message: "The requested Workspace does not exist in bbolt."}
	ErrProviderNotConfig   = &IPCError{Code: "provider_not_configured", Message: "No provider (Gemini/Qwen) has been configured yet."}
	ErrToolNotFound        = &IPCError{Code: "tool_not_found", Message: "The provided Agentic Tool is not in the Registry."}
	ErrSnapshotFailed      = &IPCError{Code: "snapshot_failed", Message: "Rollback snapshot corrupted before tool execution."}
	ErrIPCMethodUnknown    = &IPCError{Code: "ipc_method_unknown", Message: "This IPC Endpoint / Method does not exist."}
	ErrIPCPayloadInvalid   = &IPCError{Code: "ipc_payload_invalid", Message: "Malformed JSON payload."}
	ErrInternalError       = &IPCError{Code: "internal_error", Message: "Unhandled abrupt crash in the IPC Backend."}
)
