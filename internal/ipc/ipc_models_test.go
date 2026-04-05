package ipc

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestModelListEndpoint testa endpoint model.list
func TestModelListEndpoint(t *testing.T) {
	// Nota: Este teste requer que o daemon esteja rodando em background
	// Para executar: go test -v -run TestModelListEndpoint ./internal/ipc/...

	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create IPC client: %v", err)
	}
	defer client.Close()

	// Tentar conectar com timeout curto
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Connect()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Skipf("Daemon not running: %v", err)
		}
	case <-ctx.Done():
		t.Skip("Connection timeout: daemon not responding")
	}

	// Testar model.list
	var models []map[string]interface{}
	err = client.Send(context.Background(), "model.list", json.RawMessage(`{}`), &models)
	if err != nil {
		t.Errorf("model.list failed: %v", err)
		return
	}

	if len(models) == 0 {
		t.Error("Expected models in catalog, got none")
		return
	}

	t.Logf("✓ model.list returned %d models", len(models))
}

// TestModelDetectEndpoint testa endpoint model.detect
func TestModelDetectEndpoint(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create IPC client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Connect()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Skipf("Daemon not running: %v", err)
		}
	case <-ctx.Done():
		t.Skip("Connection timeout: daemon not responding")
	}

	// Testar model.detect
	var hw map[string]interface{}
	err = client.Send(context.Background(), "model.detect", json.RawMessage(`{}`), &hw)
	if err != nil {
		t.Errorf("model.detect failed: %v", err)
		return
	}

	if os, ok := hw["os"].(string); !ok || os == "" {
		t.Error("Hardware detection: OS not found")
		return
	}

	t.Logf("✓ model.detect returned valid hardware info")
}

// TestModelRecommendEndpoint testa endpoint model.recommend
func TestModelRecommendEndpoint(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create IPC client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Connect()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Skipf("Daemon not running: %v", err)
		}
	case <-ctx.Done():
		t.Skip("Connection timeout: daemon not responding")
	}

	// Testar model.recommend
	var model map[string]interface{}
	err = client.Send(context.Background(), "model.recommend", json.RawMessage(`{}`), &model)
	if err != nil {
		t.Errorf("model.recommend failed: %v", err)
		return
	}

	if id, ok := model["id"].(string); !ok || id == "" {
		t.Error("Recommendation: model ID not found")
		return
	}

	t.Logf("✓ model.recommend returned valid model recommendation")
}

// TestIPCProtocol testa a integridade do protocolo JSON-ND
func TestIPCProtocol(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create IPC client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Connect()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Skipf("Daemon not running: %v", err)
		}
	case <-ctx.Done():
		t.Skip("Connection timeout: daemon not responding")
	}

	// Testar app.health (simples)
	var health map[string]interface{}
	err = client.Send(context.Background(), "app.health", json.RawMessage(`{}`), &health)
	if err != nil {
		t.Errorf("app.health failed: %v", err)
		return
	}

	status, ok := health["status"].(string)
	if !ok || status != "ok" {
		t.Errorf("Invalid health response: %v", health)
		return
	}

	t.Logf("✓ IPC protocol working correctly: %+v", health)
}
