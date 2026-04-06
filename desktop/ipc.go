package desktop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	maxRetries        = 3
	initialBackoff    = 100 * time.Millisecond
	maxBackoff        = 5 * time.Second
	connectionTimeout = 5 * time.Second
)

// IPCClient wraps the IPC client with reconnect logic and error handling
type IPCClient struct {
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
	mu      sync.Mutex

	reconnecting bool
	lastError    error
	lastErrorMu  sync.RWMutex
}

// NewIPCClient creates and initializes an IPC client
func NewIPCClient() (*IPCClient, error) {
	c := &IPCClient{}

	// Try to connect
	if err := c.connect(); err != nil {
		// Not fatal - will retry later
		c.lastError = err
	}

	return c, nil
}

// connect establishes a connection to the daemon
func (c *IPCClient) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Determine socket address based on platform
	var addr string
	if runtime.GOOS == "windows" {
		// Use named pipe on Windows
		addr = `\\.\pipe\vectora-ipc`
	} else {
		// Use Unix socket on Linux/macOS
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		socketDir := filepath.Join(home, ".vectora")
		addr = filepath.Join(socketDir, "daemon.sock")
	}

	// Create dialer with timeout
	dialer := net.Dialer{
		Timeout: connectionTimeout,
	}

	// Connect to daemon
	var conn net.Conn
	var err error

	if runtime.GOOS == "windows" {
		// Windows named pipe
		conn, err = dialer.Dial("tcp", "127.0.0.1:9999")
		if err != nil {
			return fmt.Errorf("failed to connect to daemon: %w", err)
		}
	} else {
		// Unix socket
		conn, err = dialer.Dial("unix", addr)
		if err != nil {
			return fmt.Errorf("failed to connect to daemon: %w", err)
		}
	}

	c.conn = conn
	c.encoder = json.NewEncoder(conn)
	c.decoder = json.NewDecoder(conn)

	return nil
}

// EnsureConnected ensures we have an active connection to the daemon
func (c *IPCClient) EnsureConnected() error {
	c.mu.Lock()

	// If already connected, check connection is still alive
	if c.conn != nil {
		// Try a simple ping
		c.mu.Unlock()
		if err := c.Ping(); err == nil {
			return nil
		}
		c.mu.Lock()
		// Connection died
		c.conn.Close()
		c.conn = nil
	}

	// Not connected, attempt reconnection with backoff
	backoff := initialBackoff
	for attempt := 0; attempt < maxRetries; attempt++ {
		c.mu.Unlock()

		if attempt > 0 {
			time.Sleep(backoff)
			backoff = backoff * 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}

		c.mu.Lock()
		if err := c.connect(); err == nil {
			c.lastError = nil
			c.mu.Unlock()
			return nil
		} else {
			c.lastError = err
		}
	}

	c.mu.Unlock()
	return fmt.Errorf("failed to reconnect after %d attempts", maxRetries)
}

// Send sends a request to the daemon and receives response
func (c *IPCClient) Send(ctx context.Context, route string, params any) (any, error) {
	if err := c.EnsureConnected(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil, errors.New("not connected to daemon")
	}

	// Send request
	request := map[string]any{
		"jsonrpc": "2.0",
		"method":  route,
		"params":  params,
		"id":      time.Now().UnixNano(),
	}

	if err := c.encoder.Encode(request); err != nil {
		c.conn.Close()
		c.conn = nil
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Receive response
	var response map[string]any
	if err := c.decoder.Decode(&response); err != nil {
		if err == io.EOF {
			c.conn.Close()
			c.conn = nil
		}
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	// Check for JSON-RPC error
	if errObj, ok := response["error"]; ok && errObj != nil {
		return nil, fmt.Errorf("daemon error: %v", errObj)
	}

	return response["result"], nil
}

// Stream opens a stream from the daemon
func (c *IPCClient) Stream(ctx context.Context, route string, params any) (<-chan any, error) {
	if err := c.EnsureConnected(); err != nil {
		return nil, err
	}

	ch := make(chan any, 10)

	go func() {
		defer close(ch)

		c.mu.Lock()
		if c.conn == nil {
			c.mu.Unlock()
			return
		}

		// Send stream request
		request := map[string]any{
			"jsonrpc": "2.0",
			"method":  route + ":stream",
			"params":  params,
		}

		if err := c.encoder.Encode(request); err != nil {
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()

		// Receive stream data
		for {
			select {
			case <-ctx.Done():
				return
			default:
				c.mu.Lock()
				if c.conn == nil {
					c.mu.Unlock()
					return
				}
				var data any
				err := c.decoder.Decode(&data)
				c.mu.Unlock()

				if err != nil {
					return
				}

				select {
				case ch <- data:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// Ping checks if the daemon is responding
func (c *IPCClient) Ping() error {
	_, err := c.Send(context.Background(), "system.health", nil)
	return err
}

// Close closes the connection
func (c *IPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetLastError returns the last error encountered
func (c *IPCClient) GetLastError() error {
	c.lastErrorMu.RLock()
	defer c.lastErrorMu.RUnlock()
	return c.lastError
}

// Request/Response types for common operations
type (
	// ChatRequest represents a chat message request
	ChatRequest struct {
		Message string `json:"message"`
		Stream  bool   `json:"stream"`
	}

	// ChatResponse represents a chat response
	ChatResponse struct {
		Response string `json:"response"`
		Model    string `json:"model"`
	}

	// ModelInfo represents a model
	ModelInfo struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Active   bool   `json:"active"`
		Size     string `json:"size"`
		Type     string `json:"type"`
	}

	// IndexInfo represents an index
	IndexInfo struct {
		ID            string    `json:"id"`
		Name          string    `json:"name"`
		DocumentCount int       `json:"document_count"`
		Size          int64     `json:"size"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
	}

	// SystemHealth represents system health info
	SystemHealth struct {
		Status     string `json:"status"`
		Uptime     int64  `json:"uptime"`
		ModelsPath string `json:"models_path"`
		Version    string `json:"version"`
	}
)
