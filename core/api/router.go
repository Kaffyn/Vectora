package api

import (
	"context"
	"vectora/core/engine" // Interface unificada do Core
)

// Router gerencia a inicialização dos servidores de protocolo.
type Router struct {
	Engine *engine.Engine
	Config *RouterConfig
}

type RouterConfig struct {
	EnableJSONRPC bool
	EnablegRPC    bool
	EnableIPC     bool
	TCPPort       int    // Para JSON-RPC over TCP (Dev Mode)
	IPCPath       string // Path para Named Pipe ou Unix Socket
}

func NewRouter(engine *engine.Engine, cfg *RouterConfig) *Router {
	return &Router{Engine: engine, Config: cfg}
}

// StartAll inicia os listeners em goroutines separadas.
func (r *Router) StartAll(ctx context.Context) error {
	if r.Config.EnableJSONRPC {
		go r.startJSONRPCServer(ctx)
	}
	if r.Config.EnablegRPC {
		go r.startgRPCServer(ctx)
	}
	if r.Config.EnableIPC {
		go r.startIPCServer(ctx)
	}
	return nil
}

// --- Stubs para implementação nos módulos específicos ---
func (r *Router) startJSONRPCServer(ctx context.Context) { /* ... */ }
func (r *Router) startgRPCServer(ctx context.Context)    { /* ... */ }
func (r *Router) startIPCServer(ctx context.Context)     { /* ... */ }
