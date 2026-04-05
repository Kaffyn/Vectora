package ipc

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of IPC message
type MessageType string

const (
	TypeRequest  MessageType = "request"
	TypeResponse MessageType = "response"
	TypeEvent    MessageType = "event"
)

// Message represents an IPC message
type Message struct {
	ID        string          `json:"id"`
	Type      MessageType     `json:"type"`
	Method    string          `json:"method"`
	Payload   json.RawMessage `json:"payload"`
	Error     *ErrorInfo      `json:"error,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

// Handler is a function that processes IPC messages
type Handler func(msg *Message) (*Message, error)

// RequestPayload defines common request structure
type RequestPayload struct {
	SessionID string                 `json:"sessionId,omitempty"`
	Workspace string                 `json:"workspace,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// ResponsePayload defines common response structure
type ResponsePayload struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// NewMessage creates a new IPC message
func NewMessage(id string, msgType MessageType, method string, payload interface{}) (*Message, error) {
	var jsonPayload json.RawMessage
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		jsonPayload = data
	}

	return &Message{
		ID:        id,
		Type:      msgType,
		Method:    method,
		Payload:   jsonPayload,
		Timestamp: time.Now(),
	}, nil
}

// NewErrorMessage creates an error response message
func NewErrorMessage(id string, code, message string) *Message {
	return &Message{
		ID:        id,
		Type:      TypeResponse,
		Timestamp: time.Now(),
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// UnmarshalPayload unmarshals the message payload into a target struct
func (m *Message) UnmarshalPayload(target interface{}) error {
	return json.Unmarshal(m.Payload, target)
}

// MarshalPayload creates a raw message from a payload struct
func (m *Message) MarshalPayload(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	m.Payload = data
	return nil
}
