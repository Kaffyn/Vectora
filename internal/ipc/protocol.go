package ipc

import (
	"encoding/json"
)

// Constante absoluta que garante delimitação do Socket Stream. (JSON-ND)
const FrameDelimiter = '\n'

// Tipos Restritos Padrão (RPC)
const (
	MsgTypeRequest  = "request"
	MsgTypeResponse = "response"
	MsgTypeEvent    = "event"
)

// IPCMessage é a Unidade de Transporte Universal para todas as Sockets do Vectora.
type IPCMessage struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`             // "request", "response", "event"
	Method  string          `json:"method,omitempty"` // Requerido apenas em Type == "request" ou "event"
	Payload json.RawMessage `json:"payload"`          // Corpo arbitrário
	Error   *IPCError       `json:"error,omitempty"`  // Requerido apenas se houver falha e Type == "response"
}

// IPCError é a devolução padronizada de falhas pro Frontend (Node/React).
type IPCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

// Respostas Cânonicas de Erro (Conforme as Business Rules)
var (
	ErrWorkspaceNotFound   = &IPCError{Code: "workspace_not_found", Message: "O Workspace solicitado não existe no bbbolt."}
	ErrProviderNotConfig   = &IPCError{Code: "provider_not_configured", Message: "Nenhum provedor (Gemini/Qwen) foi configurado ainda."}
	ErrToolNotFound        = &IPCError{Code: "tool_not_found", Message: "A Tool Agentica fornecida não consta no Registry."}
	ErrSnapshotFailed      = &IPCError{Code: "snapshot_failed", Message: "Rollback snapshot corrompido antes da execução da tool."}
	ErrIPCMethodUnknown    = &IPCError{Code: "ipc_method_unknown", Message: "Este Endpoit / Método IPC não existe."}
	ErrIPCPayloadInvalid   = &IPCError{Code: "ipc_payload_invalid", Message: "Payload JSON mal formado."}
	ErrInternalError       = &IPCError{Code: "internal_error", Message: "Crash abrupto não tratado no Backend IPC."}
)
