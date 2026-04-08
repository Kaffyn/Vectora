package jsonrpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	"vectora/core/api/handlers"
	"vectora/core/engine"
)

type Server struct {
	Engine *engine.Engine
}

// RPCService expõe os métodos via Reflection do pacote net/rpc
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

func StartStdioServer(engine *engine.Engine) {
	service := &RPCService{Engine: engine}
	rpc.Register(service)

	// Lê do Stdin e escreve no Stdout (Padrão MCP)
	conn := jsonrpc.NewConn(&struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout})

	rpc.ServeConn(conn)
}

func StartTCPServer(engine *engine.Engine, port int) {
	service := &RPCService{Engine: engine}
	rpc.Register(service)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
}
