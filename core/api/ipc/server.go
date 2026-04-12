package ipc

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Kaffyn/Vectora/core/manager" // Novo import
	vecos "github.com/Kaffyn/Vectora/core/os"
)

const ipcScannerBufSize = 4 * 1024 * 1024 // 4 MiB per connection

type RouterFunc func(ctx context.Context, payload json.RawMessage) (any, *IPCError)

type Server struct {
	addr          string
	listener      net.Listener
	handlers      map[string]RouterFunc
	clients       map[net.Conn]bool
	clientsLock   sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	token         string // IPC auth token; empty = no auth
	tenantManager *manager.TenantManager
	resourcePool  *manager.ResourcePool
}

func NewServer(tm *manager.TenantManager, rp *manager.ResourcePool) (*Server, error) {
	osMgr, err := vecos.NewManager()
	if err != nil {
		return nil, err
	}
	baseDir, _ := osMgr.GetAppDataDir()

	var addr string
	if runtime.GOOS == "windows" {
		addr = `\\.\pipe\vectora`
	} else {
		addr = filepath.Join(baseDir, "run", "vectora.sock")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		addr:          addr,
		handlers:      make(map[string]RouterFunc),
		clients:       make(map[net.Conn]bool),
		ctx:           ctx,
		cancel:        cancel,
		tenantManager: tm,
		resourcePool:  rp,
	}, nil
}

func (s *Server) Register(method string, handler RouterFunc) {
	s.handlers[method] = handler
}

func (s *Server) Start() error {
	l, err := listenIPC(s.addr)
	if err != nil {
		return err
	}

	s.listener = l

	// Generate token after the listener is open so we only write a token
	// that corresponds to a live server. Hard-fail if token generation fails
	// — running without auth is a security risk.
	if err := s.generateToken(); err != nil {
		s.listener.Close()
		return fmt.Errorf("IPC: cannot generate auth token: %w", err)
	}

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.ctx.Done():
					return
				default:
					continue
				}
			}

			s.clientsLock.Lock()
			s.clients[conn] = true
			s.clientsLock.Unlock()

			go s.handleConnection(conn)
		}
	}()

	return nil
}

// generateToken creates a random 32-byte hex token and writes it to
// %APPDATA%\Vectora\ipc.token (Windows) or ~/.Vectora/ipc.token (Unix).
func (s *Server) generateToken() error {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	s.token = hex.EncodeToString(b)

	osMgr, err := vecos.NewManager()
	if err != nil {
		return err
	}
	baseDir, err := osMgr.GetAppDataDir()
	if err != nil {
		return err
	}

	tokenPath := filepath.Join(baseDir, "ipc.token")
	return os.WriteFile(tokenPath, []byte(s.token), 0600)
}

func (s *Server) StartDevHTTP(port int) {
	mux := http.NewServeMux()

	// Register pprof endpoints for CPU/memory profiling (localhost only, dev mode)
	// Access via:
	// - http://localhost:{port}/debug/pprof/ (profile index)
	// - http://localhost:{port}/debug/pprof/heap (memory allocation)
	// - http://localhost:{port}/debug/pprof/goroutine (goroutines)
	// - http://localhost:{port}/debug/pprof/profile?seconds=30 (CPU profile)
	// Download: go tool pprof http://localhost:{port}/debug/pprof/heap
	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		method := strings.TrimPrefix(r.URL.Path, "/api/v1/")
		handler, exists := s.handlers[method]
		if !exists {
			http.Error(w, fmt.Sprintf("Method '%s' not found", method), http.StatusNotFound)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		resData, ipcErr := handler(s.ctx, json.RawMessage(body))

		w.Header().Set("Content-Type", "application/json")
		if ipcErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{"error": ipcErr})
			return
		}

		json.NewEncoder(w).Encode(resData)
	})

	log.Printf("IPC-HTTP Bridge Active at http://localhost:%d (Dev Mode Only)", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Printf("Failed to start Dev HTTP Bridge: %v", err)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.clientsLock.Lock()
		delete(s.clients, conn)
		s.clientsLock.Unlock()
		conn.Close()
	}()

	// Per-connection state
	authorized := s.token == ""
	var activeTenant *manager.Tenant

	scanner := bufio.NewScanner(conn)
	buf := make([]byte, ipcScannerBufSize)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		frame := scanner.Bytes()
		if len(frame) == 0 {
			continue
		}

		var msg IPCMessage
		if err := json.Unmarshal(frame, &msg); err != nil {
			s.sendError(conn, "", ErrIPCPayloadInvalid)
			continue
		}

		if msg.Type != MsgTypeRequest {
			continue
		}

		// Ensure JSONRPC field is set for compliance
		msg.JSONRPC = "2.0"

		// Use Params if Payload is empty
		payload := msg.Payload
		if len(payload) == 0 {
			payload = msg.Params
		}

		// 1. Auth handshake
		if msg.Method == "ipc.auth" {
			var req struct {
				Token string `json:"token"`
			}
			if err := json.Unmarshal(payload, &req); err != nil || req.Token != s.token {
				s.sendError(conn, msg.ID, ErrUnauthorized)
				conn.Close()
				return
			}
			authorized = true
			s.writeMessage(conn, IPCMessage{JSONRPC: "2.0", ID: msg.ID, Type: MsgTypeResponse, Payload: json.RawMessage(`{"ok":true}`), Result: json.RawMessage(`{"ok":true}`)})
			continue
		}

		if !authorized {
			s.sendError(conn, msg.ID, ErrUnauthorized)
			conn.Close()
			return
		}

		// 2. Multi-Tenancy Handshake: workspace.init
		// This must happen before any other data method is called.
		if msg.Method == "workspace.init" {
			var req WorkspaceInitRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				s.sendError(conn, msg.ID, ErrIPCPayloadInvalid)
				continue
			}

			wsID := GenerateWorkspaceID(req.WorkspaceRoot)

			// Load or create the tenant via manager
			tenant, err := s.tenantManager.GetOrCreateTenant(wsID, req.WorkspaceRoot, req.ProjectName)
			if err != nil {
				s.sendError(conn, msg.ID, errServer("mtp_load_failed", err.Error()))
				continue
			}

			activeTenant = tenant

			res, _ := json.Marshal(map[string]string{
				"workspace_id": wsID,
				"status":       "initialized",
			})
			s.writeMessage(conn, IPCMessage{JSONRPC: "2.0", ID: msg.ID, Type: MsgTypeResponse, Result: res, Payload: res})
			continue
		}

		// 3. Command Routing (requires initialized workspace)
		if activeTenant == nil {
			s.sendError(conn, msg.ID, &IPCError{
				Code:    -32001,
				Slug:    "workspace_not_initialized",
				Message: "You must call 'workspace.init' before any other command.",
			})
			continue
		}

		handler, exists := s.handlers[msg.Method]
		if !exists {
			s.sendError(conn, msg.ID, &IPCError{
				Code:    CodeMethodNotFound,
				Slug:    "method_not_found",
				Message: fmt.Sprintf("Method '%s' does not exist.", msg.Method),
			})
			continue
		}

		go func(m IPCMessage, p json.RawMessage, t *manager.Tenant) {
			// Create a per-request context that carries the tenant info
			ctx := manager.ContextWithTenant(s.ctx, t)

			resData, ipcErr := handler(ctx, p)

			resp := IPCMessage{
				JSONRPC: "2.0",
				ID:      m.ID,
				Type:    MsgTypeResponse,
			}

			if ipcErr != nil {
				resp.Error = ipcErr
			} else {
				payloadBytes, _ := json.Marshal(resData)
				resp.Result = payloadBytes
				resp.Payload = payloadBytes
			}

			s.writeMessage(conn, resp)
		}(msg, payload, activeTenant)
	}
}

func (s *Server) writeMessage(conn net.Conn, msg IPCMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	data = append(data, FrameDelimiter)
	conn.Write(data)
}

func (s *Server) sendError(conn net.Conn, id string, ipcErr *IPCError) {
	resp := IPCMessage{
		ID:    id,
		Type:  MsgTypeResponse,
		Error: ipcErr,
	}
	s.writeMessage(conn, resp)
}

func (s *Server) Broadcast(method string, payloadData any) {
	b, _ := json.Marshal(payloadData)
	eventMsg := IPCMessage{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:    MsgTypeEvent,
		Method:  method,
		Payload: b,
	}

	raw, _ := json.Marshal(eventMsg)
	raw = append(raw, FrameDelimiter)

	// Snapshot the connection set under read-lock so writes don't hold the lock.
	s.clientsLock.RLock()
	conns := make([]net.Conn, 0, len(s.clients))
	for conn := range s.clients {
		conns = append(conns, conn)
	}
	s.clientsLock.RUnlock()

	for _, conn := range conns {
		conn.Write(raw)
	}
}

func (s *Server) Shutdown() {
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}

	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()
	for conn := range s.clients {
		conn.Close()
	}
}
