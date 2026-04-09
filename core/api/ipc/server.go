package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	vecos "github.com/Kaffyn/Vectora/core/os"
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
		addr := `\\.\pipe\vectora`
		l, err = net.Listen("unix", addr)
		if err != nil {
			l, err = net.Listen("tcp", "127.0.0.1:42781")
		}
	} else {
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

func (s *Server) StartDevHTTP(port int) {
	mux := http.NewServeMux()

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

	scanner := bufio.NewScanner(conn)
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
			continue
		}

		handler, exists := s.handlers[msg.Method]
		if !exists {
			s.sendError(conn, msg.ID, &IPCError{
				Code:    "ipc_method_unknown",
				Message: fmt.Sprintf("Method '%s' does not exist.", msg.Method),
			})
			continue
		}

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
