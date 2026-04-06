package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kaffyn/Vectora/test/mocks"
)

// StartTestDaemon starts a test daemon instance with mocked behavior
func StartTestDaemon(t *testing.T) *mocks.MockDaemon {
	daemon := mocks.NewMockDaemon()
	if err := daemon.Start(); err != nil {
		t.Fatalf("Failed to start test daemon: %v", err)
	}
	return daemon
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "vectora-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// CreateTestFile creates a test file with content
func CreateTestFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// AssertJSONValid checks if a string is valid JSON
func AssertJSONValid(t *testing.T, data string) {
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Errorf("Invalid JSON: %v\n%s", err, data)
	}
}

// WaitFor waits for a condition to be true or timeout
func WaitFor(timeout time.Duration, fn func() bool) error {
	deadline := time.Now().Add(timeout)
	for {
		if fn() {
			return nil
		}
		if time.Now().After(deadline) {
			return ErrTimeout
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNil checks if a value is nil
func AssertNil(t *testing.T, value interface{}, message string) {
	if value != nil {
		t.Errorf("%s: expected nil, got %v", message, value)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, value interface{}, message string) {
	if value == nil {
		t.Errorf("%s: expected non-nil value", message)
	}
}

// AssertError checks if an error occurred
func AssertError(t *testing.T, err error, message string) {
	if err == nil {
		t.Errorf("%s: expected error", message)
	}
}

// AssertNoError checks if no error occurred
func AssertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf("%s: unexpected error: %v", message, err)
	}
}
