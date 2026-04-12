package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandResult holds the output of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Err      error
}

// ExecCommand executes a CLI command and captures output
func ExecCommand(ctx context.Context, name string, args ...string) *CommandResult {
	start := time.Now()
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
		Err:      err,
	}
}

// ExecCommandPipes runs a command and allows stdin input
func ExecCommandWithInput(ctx context.Context, input string, name string, args ...string) *CommandResult {
	start := time.Now()
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(input)

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
		Err:      err,
	}
}

// ProcessJSON parses JSON from output
func ProcessJSON(output string, v interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(output))
	return decoder.Decode(v)
}

// ContainsString checks if output contains substring
func ContainsString(output, substring string) bool {
	return strings.Contains(output, substring)
}

// TempDir creates a temporary directory for tests
func TempDir(prefix string) (string, error) {
	return os.MkdirTemp(os.TempDir(), prefix)
}

// CleanupDir removes a directory
func CleanupDir(path string) error {
	return os.RemoveAll(path)
}

// WriteFile creates a file with content
func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadFile reads file content
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

// FileExists checks if file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CreateTestFile creates a file for testing
func CreateTestFile(dir, filename, content string) (string, error) {
	filePath := filepath.Join(dir, filename)
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}
	err := WriteFile(filePath, content)
	return filePath, err
}

// CreateTestStructure creates a test directory structure
func CreateTestStructure(baseDir string) error {
	files := map[string]string{
		"main.go":      "package main\n\nfunc main() {}\n",
		"README.md":    "# Test Project\n",
		"test.txt":     "Test content\n",
		"sub/file.go":  "package sub\n",
		"sub/test.txt": "Nested test\n",
	}

	for filename, content := range files {
		if _, err := CreateTestFile(baseDir, filename, content); err != nil {
			return err
		}
	}
	return nil
}

// WaitForPort waits for a port to be ready
func WaitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("port %d not ready after %s", port, timeout)
			}
		}
	}
}

// GetLogLevel returns log level from environment
func GetLogLevel() string {
	level := os.Getenv("TEST_LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	return level
}

// GetCoreTimeout returns Core timeout from environment
func GetCoreTimeout() time.Duration {
	timeout := os.Getenv("TEST_CORE_TIMEOUT")
	if timeout == "" {
		timeout = "10s"
	}
	d, _ := time.ParseDuration(timeout)
	return d
}

// GetCoreBinaryPath returns path to Core binary
func GetCoreBinaryPath() string {
	path := os.Getenv("VECTORA_CORE_BIN")
	if path == "" {
		// Try common locations
		possiblePaths := []string{
			"./bin/core",
			"./bin/core.exe",
			"bin/core",
			"bin/core.exe",
		}
		for _, p := range possiblePaths {
			if FileExists(p) {
				return p
			}
		}
		path = "vectora"
	}
	return path
}

// GetWorkspaceDir returns test workspace directory
func GetWorkspaceDir() string {
	ws := os.Getenv("TEST_WORKSPACE")
	if ws == "" {
		ws = filepath.Join(os.TempDir(), "vectora-test-workspace")
	}
	return ws
}

// JSONRPCRequest creates a JSON-RPC 2.0 request
func JSONRPCRequest(id int, method string, params interface{}) map[string]interface{} {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}
	return req
}

// JSONRPCError creates a JSON-RPC 2.0 error response
func JSONRPCError(id int, code int, message string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}

// JSONRPCResult creates a JSON-RPC 2.0 result response
func JSONRPCResult(id int, result interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
}

// ReadJSONRPCResponse reads a JSON-RPC response from reader
func ReadJSONRPCResponse(r io.Reader) (map[string]interface{}, error) {
	decoder := json.NewDecoder(r)
	var resp map[string]interface{}
	err := decoder.Decode(&resp)
	return resp, err
}

// ValidateJSONRPCResponse validates JSON-RPC 2.0 response structure
func ValidateJSONRPCResponse(resp map[string]interface{}) error {
	if version, ok := resp["jsonrpc"]; !ok || version != "2.0" {
		return fmt.Errorf("invalid jsonrpc version: %v", version)
	}

	// Must have either result or error
	_, hasResult := resp["result"]
	_, hasError := resp["error"]

	if !hasResult && !hasError {
		return fmt.Errorf("response must have either result or error field")
	}

	if hasResult && hasError {
		return fmt.Errorf("response cannot have both result and error fields")
	}

	return nil
}

// ParseTestConfig reads .testenv file
func ParseTestConfig(path string) map[string]string {
	config := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return config
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return config
}

// AssertEqual checks if two values are equal
func AssertEqual(expected, actual interface{}) error {
	if expected != actual {
		return fmt.Errorf("expected %v, got %v", expected, actual)
	}
	return nil
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(expected, actual interface{}) error {
	if expected == actual {
		return fmt.Errorf("expected not %v, got %v", expected, actual)
	}
	return nil
}

// AssertContains checks if string contains substring
func AssertContains(s, substring string) error {
	if !strings.Contains(s, substring) {
		return fmt.Errorf("expected to contain %q, got %q", substring, s)
	}
	return nil
}

// AssertNotContains checks if string does not contain substring
func AssertNotContains(s, substring string) error {
	if strings.Contains(s, substring) {
		return fmt.Errorf("expected not to contain %q, got %q", substring, s)
	}
	return nil
}

// AssertNil checks if value is nil
func AssertNil(v interface{}) error {
	if v != nil {
		return fmt.Errorf("expected nil, got %v", v)
	}
	return nil
}

// AssertNotNil checks if value is not nil
func AssertNotNil(v interface{}) error {
	if v == nil {
		return fmt.Errorf("expected not nil")
	}
	return nil
}

// AssertError checks if error occurred
func AssertError(err error) error {
	if err == nil {
		return fmt.Errorf("expected error, got nil")
	}
	return nil
}

// AssertNoError checks if no error occurred
func AssertNoError(err error) error {
	if err != nil {
		return fmt.Errorf("expected no error, got %v", err)
	}
	return nil
}

// AssertExitCode checks exit code
func AssertExitCode(expected int, actual int) error {
	if expected != actual {
		return fmt.Errorf("expected exit code %d, got %d", expected, actual)
	}
	return nil
}

// FormatDuration formats duration in human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dμs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// PrintTestHeader prints a test header
func PrintTestHeader(name string) {
	fmt.Printf("\n%-50s ", name)
}

// PrintTestResult prints test result
func PrintTestResult(passed bool, duration time.Duration) {
	status := "✅ PASS"
	if !passed {
		status = "❌ FAIL"
	}
	fmt.Printf("%s (%.0fms)\n", status, duration.Seconds()*1000)
}

// CaptureOutput captures function output
func CaptureOutput(fn func()) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	fn()
	w.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	return buf.String(), err
}
