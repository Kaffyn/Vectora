package ipc

import (
	"encoding/json"
	"net"
	"os"
	"runtime"
	"vectora/core/engine"
)

type IPCServer struct {
	Engine *engine.Engine
	Path   string
}

type IPCMessage struct {
	Type    string      `json:"type"` // "status_update", "permission_req"
	Payload interface{} `json:"payload"`
}

func (s *IPCServer) Start() error {
	// Remove socket antigo se existir
	os.Remove(s.Path)

	listener, err := net.Listen("unix", s.Path)
	if err != nil {
		if runtime.GOOS == "windows" {
			// Lógica específica para Named Pipes no Windows
			// listener, err = winio.ListenPipe(`\\.\pipe\vectora`, nil)
		}
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *IPCServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	var msg IPCMessage
	for decoder.Decode(&msg) == nil {
		switch msg.Type {
		case "start_index":
			s.Engine.StartIndexation()
		case "get_status":
			status := s.Engine.GetStatus()
			json.NewEncoder(conn).Encode(IPCMessage{Type: "status_update", Payload: status})
		}
	}
}
