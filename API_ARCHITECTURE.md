# Arquitetura de API do Vectora

**Data:** 2026-04-11
**Status:** Completo - Fase 7D finalizada
**Última atualização:** Commit b47b984 (Phase 7D: Full MCP Protocol over stdio)

---

## 1. Visão Geral dos Protocolos

Vectora implementa **três protocolos independentes** que compartilham uma **camada JSON-RPC 2.0** comum:

```
┌─────────────────────────────────────────────────────────┐
│                  JSON-RPC 2.0 (Foundation)              │
│   (core/api/jsonrpc/ - Minimal, dependency-free)        │
└─────────────────────────────────────────────────────────┘
  ↓                    ↓                    ↓
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│     ACP      │ │     IPC      │ │     MCP      │
│  (Stdio)     │ │  (Named Pipe)│ │  (Stdio)     │
└──────────────┘ └──────────────┘ └──────────────┘
```

### 1.1 JSON-RPC 2.0 - Camada Base Independente

**Localização:** `core/api/jsonrpc/server.go`

JSON-RPC é **totalmente separado e agnóstico** aos protocolos. É uma implementação mínima que:

- ✅ Não tem dependências externas
- ✅ Suporta newline-delimited JSON (frames)
- ✅ Implementa erro padrão JSON-RPC (-32700 a -32099)
- ✅ Serve como transporte para ACP, IPC e MCP

**Estrutura base:**

```go
type Request struct {
    JSONRPC string          // "2.0"
    ID      *int64          // Identificador da requisição (null = notificação)
    Method  string          // Nome do método
    Params  json.RawMessage // Parâmetros da requisição
}

type Response struct {
    JSONRPC string      // "2.0"
    ID      *int64      // Mesmo ID da requisição
    Result  interface{} // Resultado (null se erro)
    Error   *Error      // Erro (null se sucesso)
}
```

---

## 2. ACP (Agent Client Protocol)

**Localização:** `core/api/acp/`

### 2.1 O que é ACP?

ACP é o **protocolo para Vectora funcionar como agente autônomo invocado por IDEs**:

- **Cliente:** VS Code Extension, Claude Code, Antigravity, IDE qualquer
- **Servidor:** Vectora Core (como agente)
- **Transporte:** Stdio (JSON-RPC 2.0 over newline-delimited JSON)
- **SDK:** `github.com/coder/acp-go-sdk` (oficial Coder)

### 2.2 Implementação

**Arquivo principal:** `core/api/acp/agent.go`

A classe `VectoraAgent` implementa o ACP SDK:

```go
type VectoraAgent struct {
    name        string
    version     string
    kvStore     db.KVStore
    vecStore    db.VectorStore
    llmRouter   *llm.Router
    msgService  *llm.MessageService
    logger      *slog.Logger
    sessions    map[acp.SessionId]*AgentSession
}

// Interface ACP do SDK
func (a *VectoraAgent) Initialize(ctx context.Context, params acp.InitializeRequest) (acp.InitializeResponse, error)
func (a *VectoraAgent) Authenticate(ctx context.Context, params acp.AuthenticateRequest) (acp.AuthenticateResponse, error)
func (a *VectoraAgent) Cancel(ctx context.Context, params acp.CancelNotification) error
func (a *VectoraAgent) CreateSession(ctx context.Context, params acp.CreateSessionRequest) (acp.CreateSessionResponse, error)
func (a *VectoraAgent) Prompt(ctx context.Context, params acp.PromptRequest) (acp.PromptResponse, error)
// ... mais métodos
```

### 2.3 Fluxo ACP

```
IDE (VS Code, Claude Code, etc)
  ↓ (stdio JSON-RPC 2.0)
Vectora ACP Agent
  ├─ initialize() → negocia versão e capabilities
  ├─ session/new → cria nova sessão
  ├─ session/prompt → recebe prompt do usuário
  └─ session/update → envia respostas streaming
```

### 2.4 Tipos de ACP

**Em `core/api/acp/types.go`:**

- `InitializeRequest/Response` - Handshake inicial
- `SessionNewRequest/Response` - Criar sessão
- `SessionPromptRequest` - Enviar prompt
- `SessionUpdate` - Streaming de respostas
- `RequestPermissionRequest` - Pedir autorização do IDE
- `FSReadRequest/FSWriteRequest` - Operações de arquivo
- `TerminalCreateRequest` - Executar comando

### 2.5 Status de Implementação

✅ **Completo:**

- Integração com Coder ACP SDK (v0.10.8)
- Suporte a sessions, prompts, streaming
- Suporte a tool calls, file operations, terminal
- Testes unitários em `acp_test.go`

❌ **Não implementado:**

- Extension methods específicas do IDE

---

## 3. MCP (Model Context Protocol)

**Localização:** `core/api/mcp/`

### 3.1 O que é MCP?

MCP é o **protocolo para Vectora funcionar como SUB-AGENTE invocado por outro agente**:

- **Cliente:** Claude Code, Antigravity, qualquer agente MCP
- **Servidor:** Vectora Core (expõe tools como sub-agente)
- **Transporte:** Stdio (JSON-RPC 2.0)
- **Propósito:** Expor ferramentas de Vectora para agentes parent

### 3.2 Implementação

**Arquivo principal:** `core/api/mcp/stdio.go`

```go
type StdioServer struct {
    Engine *engine.Engine    // Motor Vectora com tools
    logger *slog.Logger
    reader *bufio.Reader     // Lê de stdin
    writer io.Writer         // Escreve em stdout
    mu     sync.Mutex
}

func NewStdioServerFromMCP(...) *StdioServer
func (s *StdioServer) Start(ctx context.Context) error
func (s *StdioServer) handleRequest(ctx context.Context, method string, params json.RawMessage, id any)
func (s *StdioServer) listTools() map[string]any
func (s *StdioServer) callTool(ctx context.Context, params json.RawMessage) (any, error)
```

### 3.3 Métodos MCP

| Método       | Descrição                                                     |
| ------------ | ------------------------------------------------------------- |
| `initialize` | Retorna protocol version 2024-11-05, capabilities, serverInfo |
| `tools/list` | Lista todas as tools disponíveis do Engine                    |
| `tools/call` | Executa uma tool específica com parâmetros                    |

### 3.4 Exemplo de Tool Retornado

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [
      {
        "name": "workspace.query",
        "description": "Query indexed files with semantic search",
        "inputSchema": {
          "type": "object",
          "properties": {
            "query": { "type": "string" },
            "workspace": { "type": "string" }
          }
        }
      }
    ]
  }
}
```

### 3.5 Fluxo MCP

```
Claude Code (parent agent)
  ↓ (stdio JSON-RPC 2.0)
Vectora MCP Server
  ├─ initialize() → valida protocolo
  ├─ tools/list → retorna catálogo de tools
  └─ tools/call → executa tool via Engine.ExecuteTool()
```

### 3.6 Status de Implementação

✅ **Completo:**

- Protocolo JSON-RPC 2.0 sobre stdio
- Suporte a initialize, tools/list, tools/call
- Integração com Engine e Tool Registry
- 6 testes unitários em `stdio_test.go`
- Factory `NewStdioServerFromMCP()` para startup

❌ **Não implementado:**

- Resources (read-only access)
- Prompts (templates)

---

## 4. IPC (Inter-Process Communication)

**Localização:** `core/api/ipc/`

### 4.1 O que é IPC?

IPC é o **protocolo para comunicação interna** entre:

- **Clientes:** VS Code Extension, CLI
- **Servidor:** Vectora Core (local, sempre rodando)
- **Transporte:** Named Pipes (Windows) ou Unix Sockets (Linux/Mac)
- **Base:** JSON-RPC 2.0 customizado

### 4.2 Implementação

**Arquivo principal:** `core/api/ipc/server.go`

```go
type Server struct {
    addr          string                    // Pipe path
    listener      net.Listener              // Socket/Pipe listener
    handlers      map[string]RouterFunc     // Métodos registrados
    tenantManager *manager.TenantManager    // Multi-tenancy
    resourcePool  *manager.ResourcePool     // Resource pooling
}

type RouterFunc func(ctx context.Context, payload json.RawMessage) (any, *IPCError)
```

**Protocolos suportados:**

- Windows: `\\.\pipe\vectora` (named pipe)
- Unix: `~/.vectora/run/vectora.sock` (Unix socket)

### 4.3 Métodos IPC Registrados

**Em `core/api/ipc/router.go`:**

#### Workspace & RAG

| Método                  | Descrição                             |
| ----------------------- | ------------------------------------- |
| `workspace.query`       | RAG query + LLM response com contexto |
| `workspace.embed.start` | Inicia job de embedding em background |

#### Chat Management

| Método         | Descrição                   |
| -------------- | --------------------------- |
| `chat.list`    | Listar conversas            |
| `chat.history` | Obter histórico de conversa |
| `chat.create`  | Criar nova conversa         |
| `chat.delete`  | Deletar conversa            |
| `chat.rename`  | Renomear conversa           |

#### Messages

| Método        | Descrição                     |
| ------------- | ----------------------------- |
| `message.add` | Adicionar mensagem a conversa |

#### Provider Management

| Método         | Descrição                    |
| -------------- | ---------------------------- |
| `provider.get` | Obter status do provider LLM |
| `provider.set` | Configurar provider global   |

#### Utilities

| Método              | Descrição                  |
| ------------------- | -------------------------- |
| `app.health`        | Health check do core       |
| `i18n.get`          | Obter idioma/locale        |
| `i18n.translations` | Obter arquivo de traduções |

### 4.4 Exemplo de Requisição IPC

```json
{
  "id": "req-1",
  "type": "request",
  "method": "workspace.query",
  "payload": {
    "query": "como usar embedding?",
    "conversation_id": "conv-123"
  }
}
```

**Resposta:**

```json
{
  "id": "req-1",
  "type": "response",
  "method": "workspace.query",
  "payload": {
    "answer": "Vectora uses ChromemDB for vector storage...",
    "sources": [
      {
        "filename": "README.md",
        "content": "Embedding creates vector representations..."
      }
    ]
  }
}
```

### 4.5 Status de Implementação

✅ **Completo:**

- 14 métodos registrados
- Multi-tenancy support (TenantManager, ResourcePool)
- Suporte a events/broadcast para embedding progress
- Autenticação via token
- Error handling com JSON-RPC codes

---

## 5. Comparação dos Três Protocolos

| Aspecto           | ACP                                      | MCP                     | IPC                              |
| ----------------- | ---------------------------------------- | ----------------------- | -------------------------------- |
| **Propósito**     | Vectora como agente                      | Vectora como sub-agente | Comunicação interna              |
| **Cliente**       | IDE (VS Code, etc)                       | Outro agente            | Extension, CLI                   |
| **Transporte**    | Stdio                                    | Stdio                   | Named Pipe/Unix Socket           |
| **Base**          | JSON-RPC 2.0 + ACP SDK                   | JSON-RPC 2.0            | JSON-RPC 2.0 customizado         |
| **Métodos**       | initialize, session/_, fs/_, terminal/\* | initialize, tools/\*    | workspace._, chat._, provider.\* |
| **Streaming**     | ✅ Sim                                   | ❌ Não (tool results)   | ❌ Não (events via broadcast)    |
| **Multi-tenant**  | ❌ Não                                   | ❌ Não                  | ✅ Sim (TenantManager)           |
| **Estado**        | Sessões                                  | Stateless               | Conversas persistidas            |
| **Implementação** | Coder ACP SDK                            | Homemade                | Homemade                         |
| **Status**        | ✅ Completo                              | ✅ Completo             | ✅ Completo                      |

---

## 6. JSON-RPC 2.0 - Detalhes Técnicos

### 6.1 Error Codes Padrão

```
-32700: Parse error (JSON inválido)
-32600: Invalid request (request inválida)
-32601: Method not found (método não existe)
-32602: Invalid params (parâmetros errados)
-32603: Internal error (erro do servidor)
-32000 a -32099: Server-defined errors (customizados)
```

### 6.2 Error Codes Customizados

**Em IPC (core/api/ipc/protocol.go):**

```go
CodeProviderNotConfigured = -32000  // Provider não está configurado
CodeWorkspaceNotFound = -32001      // Workspace não encontrado
CodeToolNotFound = -32002           // Tool não existe
CodeUnauthorized = -32003           // Token inválido
CodeServerError = -32099            // Erro genérico do servidor
```

### 6.3 Estrutura de Erro IPC

```go
type IPCError struct {
    Code    int    `json:"code"`      // Código JSON-RPC
    Slug    string `json:"slug"`      // ID legível (programmatic)
    Message string `json:"message"`   // Mensagem humanizada
    Detail  any    `json:"detail,omitempty"` // Dados extras
}
```

**Exemplo:**

```json
{
  "code": -32000,
  "slug": "provider_not_configured",
  "message": "No LLM provider has been configured.",
  "detail": null
}
```

---

## 7. Startup e Detecção de Protocolo

**Em `core/api/startup.go`:**

```go
type ProtocolMode string

const (
    ModeACP ProtocolMode = "acp"  // Agent mode
    ModeMCP ProtocolMode = "mcp"  // Sub-agent mode
    ModeIPC ProtocolMode = "ipc"  // Internal communication
)

func InitializeProtocol(ctx, mode, ...) error
func DetectProtocolMode() ProtocolMode
```

### Detecção Automática

1. Verificar `VECTORA_PROTOCOL` env var
2. Verificar `VECTORA_ACP_AGENT` → mode ACP
3. Verificar `VECTORA_MCP_SERVER` → mode MCP
4. Default: IPC (local operation)

### Inicialização por Modo

```
ACP:
  ├─ Cria VectoraAgent
  ├─ Chama StartACPAgent(ctx, agent, logger)
  └─ Aguarda mensagens ACP via stdio

MCP:
  ├─ Cria StdioServer via NewStdioServerFromMCP()
  ├─ Cria Engine com Guardian em "/" (todas tools)
  └─ Aguarda mensagens MCP via stdio

IPC:
  ├─ Cria IPC Server
  ├─ Registra 14 métodos
  └─ Aguarda conexões em named pipe/socket
```

---

## 8. Método Detalhado: workspace.query (IPC)

**Fluxo completo:**

```go
// 1. Recebe request IPC
{
  "method": "workspace.query",
  "payload": {
    "query": "como embeddar arquivos?",
    "conversation_id": "conv-123"
  }
}

// 2. Extrai tenant do context
tenant := manager.TenantFromContext(ctx)

// 3. Obtém provider LLM configurado
provider := getProvider()

// 4. Cria embedding da query
vector, err := provider.Embed(ctx, "como embeddar arquivos?", "")

// 5. Busca chunks similares no VectorStore tenant-specific
chunks, err := tenant.VectorStore.Query(ctx, "ws_<salted-id>", vector, 5)

// 6. Cria contexto para LLM
contextText := formatChunksAsContext(chunks)

// 7. Recupera histórico de conversa (se houver)
conv, err := msgService.GetConversation(ctx, "conv-123")

// 8. Monta mensagens para LLM
messages := []llm.Message{
  {Role: "system", Content: "You are Vectora. Use context:\n" + contextText},
  {Role: "user", Content: "como embeddar arquivos?"}
}

// 9. Chama provider.Complete()
resp, err := provider.Complete(ctx, llm.CompletionRequest{
  Messages: messages,
  MaxTokens: 1500,
  Temperature: 0.1,
})

// 10. Persiste na conversa
msgService.AddMessage(ctx, "conv-123", "assistant", resp.Content)

// 11. Retorna resposta + sources
return {
  "answer": "Vectora usa ChromemDB para armazenar...",
  "sources": chunks
}
```

---

## 9. Métodos Faltando/Futuros

### Em ACP

- [ ] Extension methods específicas do IDE
- [ ] Suporte a MCP servers dentro do ACP

### Em MCP

- [ ] Resources (read-only access)
- [ ] Prompts (templates)
- [ ] Streaming de resultados

### Em IPC

- [ ] workspace.list (listar workspaces)
- [ ] workspace.delete
- [ ] Suporte a custom tools via Tool Registry

---

## 10. Exemplo Completo: Claude Code → Vectora

```
Claude Code (parent agent)
    │
    └─ Detecta Vectora como MCP server (VECTORA_MCP_SERVER=1)
        │
        ├─ Spawn subprocess: vectora start --protocol mcp
        │
        ├─ Enviaa initialize request
        │   {
        │     "jsonrpc": "2.0",
        │     "id": 1,
        │     "method": "initialize",
        │     "params": {}
        │   }
        │
        ├─ Recebe initialize response
        │   {
        │     "jsonrpc": "2.0",
        │     "id": 1,
        │     "result": {
        │       "protocolVersion": "2024-11-05",
        │       "serverInfo": {"name": "Vectora Core", "version": "0.1.0"},
        │       "capabilities": {"tools": {}}
        │     }
        │   }
        │
        ├─ Solicita tools/list
        │   {
        │     "jsonrpc": "2.0",
        │     "id": 2,
        │     "method": "tools/list",
        │     "params": {}
        │   }
        │
        ├─ Recebe lista de tools
        │   {
        │     "jsonrpc": "2.0",
        │     "id": 2,
        │     "result": {
        │       "tools": [
        │         {
        │           "name": "workspace.query",
        │           "description": "Query indexed files...",
        │           "inputSchema": {...}
        │         },
        │         ...
        │       ]
        │     }
        │   }
        │
        ├─ Invoca tool: workspace.query
        │   {
        │     "jsonrpc": "2.0",
        │     "id": 3,
        │     "method": "tools/call",
        │     "params": {
        │       "name": "workspace.query",
        │       "input": {"query": "como usar Vectora?"}
        │     }
        │   }
        │
        ├─ Executa via Engine.ExecuteTool()
        │   └─ RAG query + LLM completion
        │
        └─ Retorna resultado
            {
              "jsonrpc": "2.0",
              "id": 3,
              "result": {
                "output": "Vectora é um sistema de RAG...",
                "isError": false
              }
            }
```

---

## 11. Resumo Executivo

### Implementação Atual

| Protocolo    | Status  | Métodos                                | SDK       | Testes         |
| ------------ | ------- | -------------------------------------- | --------- | -------------- |
| **JSON-RPC** | ✅ 100% | N/A                                    | Homemade  | ✅ 6           |
| **ACP**      | ✅ 100% | 12+                                    | Coder SDK | ✅ Tests exist |
| **MCP**      | ✅ 100% | 3 (initialize, tools/list, tools/call) | Homemade  | ✅ 6           |
| **IPC**      | ✅ 100% | 14                                     | Homemade  | ✅ Integrated  |

### JSON-RPC é Separado?

**SIM!** JSON-RPC 2.0 é:

- ✅ Totalmente independente (`core/api/jsonrpc/`)
- ✅ Sem dependências externas
- ✅ Usado como transporte por ACP, IPC e MCP
- ✅ Implementação mínima, agnóstica

### Próximas Fases

**Phase 8:** Implementação full do Coder ACP SDK na VS Code Extension
**Phase 9:** Adicionar Resources e Prompts ao MCP
**Phase 10:** Otimização de performance e streaming

---

**Implementado por:** Claude + Bruno
**Última atualização:** 2026-04-11
**Commit:** b47b984
