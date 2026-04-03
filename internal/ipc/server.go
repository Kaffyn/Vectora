package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	vecos "github.com/Kaffyn/vectora/internal/os"
)

type RouterFunc func(ctx context.Context, payload json.RawMessage) (any, *IPCError)

type Server struct {
	addr        string
	listener    net.Listener
	handlers    map[string]RouterFunc
	clients     map[net.Conn]bool
	clientsLock sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewServer() (*Server, error) {
	// Absolute determination of the Socket location (.Vectora)
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
		os.MkdirAll(filepath.Dir(addr), 0755)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		addr:     addr,
		handlers: make(map[string]RouterFunc),
		clients:  make(map[net.Conn]bool),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (s *Server) Register(method string, handler RouterFunc) {
	s.handlers[method] = handler
}

func (s *Server) Start() error {
	var l net.Listener
	var err error

	if runtime.GOOS == "windows" {
		// Named Pipes often require specific handling in Go.
		// Older versions used winio, but we now use AF_UNIX for Windows 10/11 or TCP loopback fallback.
		
		// Smart fallback for modern Windows (Windows 10 Build 17063+ supports AF_UNIX)
		addr := `\\.\pipe\vectora`
		l, err = net.Listen("unix", addr) 
		if err != nil {
			// Fallback (Hard TCP loopback) if the OS kernel does not support AF_UNIX pipes.
			l, err = net.Listen("tcp", "127.0.0.1:42780")
		}
	} else {
		// Clean old socket if it crashed
		os.Remove(s.addr)
		l, err = net.Listen("unix", s.addr)
	}

	if err != nil {
		return err
	}

	s.listener = l

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.ctx.Done():
					return // Server encerrou gracioso
				default:
					log.Println("Silent failure accepting socket:", err)
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

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.clientsLock.Lock()
		delete(s.clients, conn)
		s.clientsLock.Unlock()
		conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	// Custom limit (RN-IPC-04: Size Limit ~4MB)
	buf := make([]byte, 4*1024*1024)
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
			continue // IPC server only consumes "request" and reacts. It doesn't listen to responses because it is the master.
		}

		handler, exists := s.handlers[msg.Method]
		if !exists {
			s.sendError(conn, msg.ID, &IPCError{
				Code: "ipc_method_unknown", 
				Message: fmt.Sprintf("Method '%s' does not exist in the registry.", msg.Method),
			})
			continue
		}

		// Execute the endpoint and serialize
		go func(m IPCMessage) {
			resData, ipcErr := handler(s.ctx, m.Payload)
			
			resp := IPCMessage{
				ID:   m.ID,
				Type: MsgTypeResponse,
			}

			if ipcErr != nil {
				resp.Error = ipcErr
			} else {
				payloadBytes, _ := json.Marshal(resData)
				resp.Payload = payloadBytes
			}

			s.writeMessage(conn, resp)
		}(msg)
	}
}

func (s *Server) writeMessage(conn net.Conn, msg IPCMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	data = append(data, FrameDelimiter) // \n
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

// Broadcast emits alerts to ALL connected clients (e.g., for progress bars)
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

	s.clientsLock.RLock()
	defer s.clientsLock.RUnlock()
	for conn := range s.clients {
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
