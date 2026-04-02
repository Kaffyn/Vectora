package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"github.com/Kaffyn/vectora/internal/infra"
)

// IPCMessage represents the canonical JSON-ND message format.
type IPCMessage struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`   // "request", "response", "event"
	Method  string          `json:"method"`
	Payload json.RawMessage `json:"payload"`
}

type Server struct {
	listener net.Listener
	clients  map[net.Conn]bool
	mu       sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients: make(map[net.Conn]bool),
	}
}

// Start initiates the IPC listener.
// Note: In production this will bind to \\.\pipe\vectora (Windows) or UDS (Unix).
// Utilizing TCP localhost randomly for initial mockup.
func (s *Server) Start(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.listener = l
	
	if infra.Logger != nil {
		infra.Logger.Info(fmt.Sprintf("IPC Server listening on %s", address))
	}

	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		s.clients[conn] = true
		s.mu.Unlock()
		
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()
	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 0, 64*1024)
	// Supports max 4MB payloads (as defined in SSOT)
	scanner.Buffer(buf, 4*1024*1024)

	for scanner.Scan() {
		var msg IPCMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			if infra.Logger != nil {
				infra.Logger.Warn(fmt.Sprintf("Failed to decode IPC message: %v", err))
			}
			continue
		}
		
		if infra.Logger != nil {
			infra.Logger.Info(fmt.Sprintf("Received IPC Message: %s %s", msg.Type, msg.Method))
		}
		// Method routing will dynamically inject dependencies into internal/core
	}
}
