package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Server handles IPC communication
type Server struct {
	handlers    map[string]Handler
	mu          sync.RWMutex
	logger      *slog.Logger
	listener    net.Listener
	activeConns sync.WaitGroup
	stopChan    chan struct{}
	connsMutex  sync.Mutex
	connections map[net.Conn]bool
	maxClients  int
}

// ServerConfig contains server configuration
type ServerConfig struct {
	MaxClients int
	Logger     *slog.Logger
}

// NewServer creates a new IPC server
func NewServer(logger *slog.Logger) *Server {
	return NewServerWithConfig(&ServerConfig{
		MaxClients: 100,
		Logger:     logger,
	})
}

// NewServerWithConfig creates a new server with custom config
func NewServerWithConfig(cfg *ServerConfig) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.MaxClients == 0 {
		cfg.MaxClients = 100
	}

	return &Server{
		handlers:    make(map[string]Handler),
		logger:      cfg.Logger,
		stopChan:    make(chan struct{}),
		connections: make(map[net.Conn]bool),
		maxClients:  cfg.MaxClients,
	}
}

// Register registers a handler for a specific method
func (s *Server) Register(method string, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = h
	s.logger.Debug("Handler registered", "method", method)
}

// Listen starts listening for IPC connections
func (s *Server) Listen(ctx context.Context) error {
	if runtime.GOOS == "windows" {
		// Windows named pipe (using TCP for compatibility)
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("failed to create listener: %w", err)
		}
		s.listener = listener
		s.logger.Info("IPC server listening on TCP", "addr", listener.Addr())
	} else {
		// Unix socket
		homeDir, _ := os.UserHomeDir()
		sockDir := filepath.Join(homeDir, ".Vectora/run")
		sockPath := filepath.Join(sockDir, "vectora.sock")

		// Clean up old socket
		os.RemoveAll(sockPath)
		os.MkdirAll(sockDir, 0755)

		listener, err := net.Listen("unix", sockPath)
		if err != nil {
			return fmt.Errorf("failed to create unix socket: %w", err)
		}
		s.listener = listener
		os.Chmod(sockPath, 0666)
		s.logger.Info("IPC server listening on Unix socket", "path", sockPath)
	}

	defer s.listener.Close()

	// Accept connections
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopChan:
				return
			default:
			}

			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-s.stopChan:
					return
				default:
					s.logger.Error("Failed to accept connection", "error", err)
					continue
				}
			}

			s.connsMutex.Lock()
			if len(s.connections) >= s.maxClients {
				s.connsMutex.Unlock()
				conn.Close()
				s.logger.Warn("Max clients reached, rejecting connection")
				continue
			}
			s.connections[conn] = true
			s.connsMutex.Unlock()

			s.activeConns.Add(1)
			go func() {
				defer s.activeConns.Done()
				s.handleConnection(conn)
			}()
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	s.Stop()
	return nil
}

// handleConnection processes messages from a single connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.connsMutex.Lock()
		delete(s.connections, conn)
		s.connsMutex.Unlock()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			s.logger.Debug("Failed to unmarshal message", "error", err)
			continue
		}

		response := s.handleMessage(&msg)
		if response != nil {
			if data, err := json.Marshal(response); err == nil {
				conn.Write(append(data, '\n'))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Debug("Scanner error", "error", err)
	}
}

// handleMessage processes a single message
func (s *Server) handleMessage(msg *Message) *Message {
	s.mu.RLock()
	handler, exists := s.handlers[msg.Method]
	s.mu.RUnlock()

	if !exists {
		s.logger.Warn("Handler not found", "method", msg.Method)
		return NewErrorMessage(msg.ID, "method_not_found",
			fmt.Sprintf("method %s not found", msg.Method))
	}

	s.logger.Debug("Handling message", "id", msg.ID, "method", msg.Method)

	response, err := handler(msg)
	if err != nil {
		s.logger.Error("Handler error", "method", msg.Method, "error", err)
		return NewErrorMessage(msg.ID, "handler_error", err.Error())
	}

	if response == nil {
		response = &Message{
			ID:        msg.ID,
			Type:      TypeResponse,
			Timestamp: time.Now(),
		}
	}

	return response
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	s.logger.Info("Stopping IPC server")
	close(s.stopChan)

	if s.listener != nil {
		s.listener.Close()
	}

	// Close all connections
	s.connsMutex.Lock()
	for conn := range s.connections {
		conn.Close()
	}
	s.connsMutex.Unlock()

	// Wait for active connections to close
	done := make(chan struct{})
	go func() {
		s.activeConns.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		s.logger.Warn("Server shutdown timeout")
	}
}

// GetConnectionCount returns the number of active connections
func (s *Server) GetConnectionCount() int {
	s.connsMutex.Lock()
	defer s.connsMutex.Unlock()
	return len(s.connections)
}
