package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Client represents an IPC client
type Client struct {
	conn    net.Conn
	scanner *bufio.Scanner
	writer  io.Writer
	mu      sync.Mutex
	timeout time.Duration
}

// NewClient creates a new IPC client
func NewClient(timeout time.Duration) (*Client, error) {
	return &Client{
		timeout: timeout,
	}, nil
}

// Connect establishes connection to the IPC server
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var addr string
	var network string

	if runtime.GOOS == "windows" {
		// Try to connect to TCP server
		network = "tcp"
		addr = "127.0.0.1:0"
		// Note: In production, this would connect to a proper named pipe
	} else {
		// Unix socket
		homeDir, _ := os.UserHomeDir()
		addr = filepath.Join(homeDir, ".Vectora/run/vectora.sock")
		network = "unix"
	}

	dialer := net.Dialer{Timeout: c.timeout}
	conn, err := dialer.Dial(network, addr)
	if err != nil {
		return fmt.Errorf("failed to connect to IPC server: %w", err)
	}

	c.conn = conn
	c.scanner = bufio.NewScanner(conn)
	c.writer = conn

	return nil
}

// Send sends a message to the server and waits for response
func (c *Client) Send(msg *Message) (*Message, error) {
	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("not connected")
	}
	c.mu.Unlock()

	// Send message
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	_, err = c.writer.Write(append(data, '\n'))
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}

	// Wait for response
	responseChan := make(chan *Message, 1)
	errChan := make(chan error, 1)
	timeout := time.After(c.timeout)

	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if !c.scanner.Scan() {
			errChan <- fmt.Errorf("connection closed")
			return
		}

		var response Message
		if err := json.Unmarshal(c.scanner.Bytes(), &response); err != nil {
			errChan <- err
			return
		}
		responseChan <- &response
	}()

	select {
	case resp := <-responseChan:
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-timeout:
		return nil, fmt.Errorf("request timeout")
	}
}

// Close closes the connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns true if client is connected
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}
