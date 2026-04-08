package jsonrpc

import (
	"context"
	"fmt"
	"net"
	"net/rpc"

	"github.com/Kaffyn/Vectora/core/api/handlers"
	"github.com/Kaffyn/Vectora/core/engine"
)

type Server struct {
	Engine *engine.Engine
}

type RPCService struct {
	Engine *engine.Engine
}

func (s *RPCService) Initialize(req handlers.InitRequest, resp *handlers.InitResponse) error {
	*resp = handlers.HandleInitialize(req)
	return nil
}

func (s *RPCService) ToolsList(req interface{}, resp *handlers.ToolsListResponse) error {
	*resp = handlers.HandleToolsList()
	return nil
}

func (s *RPCService) ToolsCall(req handlers.ToolCallRequest, resp *handlers.ToolCallResponse) error {
	result, err := handlers.HandleToolsCall(context.Background(), s.Engine, req)
	if err != nil {
		return err
	}
	*resp = *result
	return nil
}

// StartStdioServer is a stub - use core/ipc/ for production MCP.
func StartStdioServer(engine *engine.Engine) {
	fmt.Println("JSON-RPC stdio server not implemented - use core/ipc/ instead")
}

// StartTCPServer starts a TCP-based JSON-RPC server for debugging.
func StartTCPServer(engine *engine.Engine, port int) {
	service := &RPCService{Engine: engine}
	rpc.Register(service)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("JSON-RPC TCP server listening on :%d\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}
