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
	// Determinação Absoluta do Local do Socket (.Vectora)
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
		// Named Pipes não costumam ser suportados perfeitamente via net padrão se não preparadas.
		// Em Go, `winio` era usado antes do 1.20, mas agora usaremos uma porta TCP de loopback fixa/segura
		// OU manteremos Unix Sockets experimentais no Windows (no r1.22 AF_UNIX é suportado no Windows 10/11)
		
		// Fallback inteligente para Windows moderno (Windows 10 Build 17063+ tem AF_UNIX)
		addr := `\\.\pipe\vectora`
		l, err = net.Listen("unix", addr) 
		if err != nil {
			// Alternativa fallback (Hard TCP loopback) se o kernel OS não suportar AF_UNIX pipes.
			l, err = net.Listen("tcp", "127.0.0.1:42780")
		}
	} else {
		// Clean antigo sock se crashou 
		os.Remove(s.addr)
		l, err = net.Listen("unix", s.addr)
	}

	if err != nil {
		return err
	}

	s.listener = l
	fmt.Printf("Servidor IPC atrelado na via do Kernel: %s\n", s.addr)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.ctx.Done():
					return // Server encerrou gracioso
				default:
					log.Println("Falha silenciosa ao aceitar socket:", err)
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
	// Limite Customizado (RN-IPC-04: Size Limit ~4MB)
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
			continue // Servidor IPC só engole "request" puro e reage. Ele não ouve respostas porque ele é o master.
		}

		handler, exists := s.handlers[msg.Method]
		if !exists {
			s.sendError(conn, msg.ID, ErrIPCMethodUnknown)
			continue
		}

		// Roda O Endpoitn e Serializa
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

// Broadcast Emite alertas para TUDO o Que Estiver Connectado na Tela (Para Barras de Progresso!)
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
