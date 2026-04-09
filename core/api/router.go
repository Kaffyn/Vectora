// Package api provides the multi-protocol router for Vectora.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/Kaffyn/Vectora/core/api/handlers"
	"github.com/Kaffyn/Vectora/core/api/jsonrpc"
	"github.com/Kaffyn/Vectora/core/engine"
)

// Router manages the initialization of protocol servers.
type Router struct {
	Engine *engine.Engine
	Config *RouterConfig
}

// RouterConfig holds configuration for protocol servers.
type RouterConfig struct {
	EnableJSONRPC bool
	EnableIPC     bool
	TCPPort       int    // JSON-RPC over TCP (Dev Mode)
	IPCPath       string // Path for Named Pipe or Unix Socket
}

// NewRouter creates a new API router.
func NewRouter(engine *engine.Engine, cfg *RouterConfig) *Router {
	return &Router{Engine: engine, Config: cfg}
}

// StartAll starts listeners in separate goroutines.
func (r *Router) StartAll(ctx context.Context) error {
	if r.Config.EnableJSONRPC {
		go r.startJSONRPCServer(ctx)
	}
	if r.Config.EnableIPC {
		go r.startIPCServer(ctx)
	}
	return nil
}

// startJSONRPCServer starts the JSON-RPC 2.0 server over stdio.
func (r *Router) startJSONRPCServer(ctx context.Context) {
	server := jsonrpc.NewServer()

	// Register Initialize
	server.RegisterMethod("initialize", func(params json.RawMessage) (interface{}, error) {
		return handlers.HandleInitialize(params)
	})

	// Register Tools List
	server.RegisterMethod("tools/list", func(params json.RawMessage) (interface{}, error) {
		return handlers.HandleToolsList()
	})

	// Register Tools Call
	server.RegisterMethod("tools/call", func(params json.RawMessage) (interface{}, error) {
		return handlers.HandleToolsCall(r.Engine, params)
	})

	// Register Session Methods (ACP)
	server.RegisterMethod("session/new", func(params json.RawMessage) (interface{}, error) {
		return handlers.HandleNewSession(r.Engine, params)
	})

	server.RegisterMethod("session/prompt", func(params json.RawMessage) (interface{}, error) {
		return handlers.HandleSessionPrompt(r.Engine, params)
	})

	fmt.Println("Vectora Core JSON-RPC Server started (stdio)")
	if err := server.ServeStdio(); err != nil {
		fmt.Printf("JSON-RPC server error: %v\n", err)
	}
}

// StartTCPDebugServer starts a TCP-based JSON-RPC server for debugging.
func StartTCPDebugServer(engine *engine.Engine, port int) error {
	server := jsonrpc.NewServer()

	server.RegisterMethod("initialize", func(params json.RawMessage) (interface{}, error) {
		return handlers.HandleInitialize(params)
	})

	server.RegisterMethod("tools/call", func(params json.RawMessage) (interface{}, error) {
		return handlers.HandleToolsCall(engine, params)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Vectora Core JSON-RPC Debug Server listening on :%d\n", port)
	return server.ServeTCP(listener)
}

// startIPCServer starts the IPC server (stub — use core/api/ipc/ for production).
func (r *Router) startIPCServer(ctx context.Context) {
	// IPC is handled by core/api/ipc/ — this is a placeholder
	fmt.Println("IPC server not implemented in router — use core/api/ipc/ directly")
}
