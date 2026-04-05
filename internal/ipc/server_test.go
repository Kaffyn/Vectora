package ipc

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if len(server.handlers) != 0 {
		t.Errorf("Expected 0 handlers, got %d", len(server.handlers))
	}
}

func TestRegisterHandler(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	handler := func(msg *Message) (*Message, error) {
		return &Message{
			ID:   msg.ID,
			Type: TypeResponse,
		}, nil
	}

	server.Register("test.method", handler)

	if len(server.handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(server.handlers))
	}
}

func TestHandleMessage(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Register test handler
	server.Register("test.echo", func(msg *Message) (*Message, error) {
		payload := map[string]string{"echo": "hello"}
		data, _ := json.Marshal(payload)
		return &Message{
			ID:      msg.ID,
			Type:    TypeResponse,
			Payload: data,
		}, nil
	})

	// Create test message
	payload := map[string]string{"data": "test"}
	payloadData, _ := json.Marshal(payload)
	msg := &Message{
		ID:      "test-1",
		Type:    TypeRequest,
		Method:  "test.echo",
		Payload: payloadData,
	}

	// Handle message
	response := server.handleMessage(msg)

	if response.ID != msg.ID {
		t.Errorf("Response ID mismatch: expected %s, got %s", msg.ID, response.ID)
	}
	if response.Type != TypeResponse {
		t.Errorf("Response type mismatch: expected %s, got %s", TypeResponse, response.Type)
	}
}

func TestHandleMessageNotFound(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	msg := &Message{
		ID:     "test-1",
		Type:   TypeRequest,
		Method: "nonexistent.method",
	}

	response := server.handleMessage(msg)

	if response.Error == nil {
		t.Error("Expected error response")
	}
	if response.Error.Code != "method_not_found" {
		t.Errorf("Expected method_not_found, got %s", response.Error.Code)
	}
}

func TestNewMessage(t *testing.T) {
	payload := map[string]string{"test": "data"}
	msg, err := NewMessage("test-1", TypeRequest, "test.method", payload)

	if err != nil {
		t.Fatalf("NewMessage failed: %v", err)
	}
	if msg.ID != "test-1" {
		t.Errorf("Message ID mismatch")
	}
	if msg.Type != TypeRequest {
		t.Errorf("Message type mismatch")
	}
}

func TestUnmarshalPayload(t *testing.T) {
	type TestPayload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := TestPayload{Name: "John", Age: 30}
	data, _ := json.Marshal(original)
	msg := &Message{Payload: data}

	var result TestPayload
	err := msg.UnmarshalPayload(&result)

	if err != nil {
		t.Fatalf("UnmarshalPayload failed: %v", err)
	}
	if result.Name != original.Name || result.Age != original.Age {
		t.Errorf("Payload mismatch")
	}
}

func BenchmarkHandleMessage(b *testing.B) {
	logger := slog.Default()
	server := NewServer(logger)

	server.Register("bench.test", func(msg *Message) (*Message, error) {
		return &Message{
			ID:   msg.ID,
			Type: TypeResponse,
		}, nil
	})

	msg := &Message{
		ID:     "bench-1",
		Type:   TypeRequest,
		Method: "bench.test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.handleMessage(msg)
	}
}
