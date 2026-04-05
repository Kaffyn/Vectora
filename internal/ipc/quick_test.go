package ipc

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestQuickModelHandlers(t *testing.T) {
	client, _ := NewClient()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Connect()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Skipf("Daemon offline: %v", err)
		}
	case <-ctx.Done():
		t.Skip("Connection timeout")
	}

	// Test _test.ping
	var pingRes map[string]string
	err := client.Send(context.Background(), "_test.ping", json.RawMessage(`{}`), &pingRes)
	t.Logf("_test.ping: err=%v, res=%v", err, pingRes)

	// Test model.list
	var models []map[string]interface{}
	err = client.Send(context.Background(), "model.list", json.RawMessage(`{}`), &models)
	t.Logf("model.list: err=%v, count=%d", err, len(models))

	// Print all registered handlers (debug)
	t.Logf("Testing if handlers exist...")
}
