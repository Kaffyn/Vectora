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

type IPCMessage struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Method  string          `json:"method,omitempty"`
	Payload json.RawMessage `json:"payload"`
	Error   *IPCError       `json:"error,omitempty"`
}

type IPCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

var (
	ErrWorkspaceNotFound = &IPCError{Code: "workspace_not_found", Message: "The requested Workspace does not exist."}
	ErrProviderNotConfig = &IPCError{Code: "provider_not_configured", Message: "No LLM provider has been configured."}
	ErrToolNotFound      = &IPCError{Code: "tool_not_found", Message: "The provided Agentic Tool is not in the Registry."}
	ErrIPCMethodUnknown  = &IPCError{Code: "ipc_method_unknown", Message: "This IPC Endpoint does not exist."}
	ErrIPCPayloadInvalid = &IPCError{Code: "ipc_payload_invalid", Message: "Malformed JSON payload."}
	ErrInternalError     = &IPCError{Code: "internal_error", Message: "Unhandled error in the IPC Backend."}
)
