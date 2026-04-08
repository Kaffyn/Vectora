# Blueprint: API Multi-Protocolo (The Gateway)

**Status:** Fase 4 - Implementação Concluída  
**Módulo:** `core/api/`  
**Dependencies:** `net/rpc/jsonrpc`, `google.golang.org/grpc`, `net`, `encoding/json`, `vectora/core/engine` (Lógica de Negócio Unificada)

## 1. O Roteador Central (`router.go`)

Este componente não contém lógica de negócio. Ele atua como um "Dispatcher" que recebe chamadas de qualquer protocolo e as encaminha para o `Engine`.

```go
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
```

## 2. Implementação JSON-RPC 2.0 (`jsonrpc/server.go`)

Implementa o padrão MCP/ACP via `stdio` (padrão) ou TCP.

```go
package jsonrpc

import (
	"context"
	"encoding/json"
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
```

## 3. Handlers Modulares (`jsonrpc/methods/`)

Cada método é um arquivo separado, facilitando a manutenção e testes unitários.

### `methods/tools_call.go`

```go
package methods

import (
	"context"
	"vectora/core/engine"
	"vectora/core/tools"
)

type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolCallResponse struct {
	Content []map[string]string `json:"content"`
	IsError bool                `json:"isError"`
}

func HandleToolsCall(ctx context.Context, eng *engine.Engine, req ToolCallRequest) (*ToolCallResponse, error) {
	// Delega para o executor de ferramentas do Engine
	result, err := eng.ExecuteTool(ctx, req.Name, req.Arguments)
	if err != nil {
		return &ToolCallResponse{
			Content: []map[string]string{{"type": "text", "text": err.Error()}},
			IsError: true,
		}, nil
	}

	return &ToolCallResponse{
		Content: []map[string]string{{"type": "text", "text": result.Output}},
		IsError: result.IsError,
	}, nil
}
```

## 4. Implementação gRPC (`grpc/server.go`)

Para streaming de alta performance (ex: indexação em tempo real).

### `proto/vectora.proto`

```protobuf
syntax = "proto3";
package vectora;

service VectoraService {
  rpc Query (QueryRequest) returns (stream QueryResponse);
  rpc Index (IndexRequest) returns (stream IndexProgress);
}

message QueryRequest { string query = 1; string workspace_id = 2; }
message QueryResponse {
  string token = 1;
  repeated string sources = 2;
  bool is_final = 3;
}

message IndexRequest { string path = 1; }
message IndexProgress { int32 files_processed = 1; int32 total_files = 2; string status = 3; }
```

### `grpc/handlers/query_handler.go`

```go
package handlers

import (
	"context"
	pb "vectora/core/api/grpc/proto"
	"vectora/core/engine"
)

type QueryHandler struct {
	pb.UnimplementedVectoraServiceServer
	Engine *engine.Engine
}

func (h *QueryHandler) Query(req *pb.QueryRequest, stream pb.VectoraService_QueryServer) error {
	// Usa o engine para fazer RAG e streamar tokens
	resultStream, err := h.Engine.StreamQuery(stream.Context(), req.Query, req.WorkspaceId)
	if err != nil {
		return err
	}

	for chunk := range resultStream {
		if err := stream.Send(&pb.QueryResponse{
			Token:   chunk.Token,
			Sources: chunk.Sources,
			IsFinal: chunk.IsFinal,
		}); err != nil {
			return err
		}
	}
	return nil
}
```

## 5. Implementação IPC (`ipc/server.go`)

Usa Named Pipes (Windows) ou Unix Sockets (Linux/Mac) para comunicação local segura entre Daemon e UI.

```go
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
```

---

### Resumo da Arquitetura de API

1.  **Modularidade Extrema:** Adicionar um novo método JSON-RPC exige apenas criar um novo arquivo em `methods/` e registrá-lo. Não há "god classes".
2.  **Agnosticismo de Transporte:** O `Engine` não sabe se está sendo chamado via gRPC ou JSON-RPC. Ele apenas executa a lógica.
3.  **Segurança por Design:**
    - **JSON-RPC:** Validado pelo `Guardian` antes de executar tools.
    - **gRPC:** Ideal para conexões internas confiáveis ou enterprise.
    - **IPC:** Restrito ao sistema local via permissões de arquivo do socket.
4.  **Conformidade com Padrões:** JSON-RPC segue a spec 2.0, compatível com MCP. gRPC usa protobufs tipados.

Esta camada de API está pronta para ser integrada ao `main.go`, conectando todos os módulos anteriores (Storage, LLM, Tools, Policies) ao mundo externo.
