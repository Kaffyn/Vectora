# Vectora Protocol SDKs Implementation Plan

**Data:** 2026-04-11
**Objetivo:** Integrar ACP (Agent Client Protocol) e MCP (Model Context Protocol) no Vectora
**Status:** Planning Phase

---

## Contexto Arquitetural

Vectora opera em **dois modos distintos**:

### 1. Agent Mode

- **VS Code Extension** (Vectora VSCode Chat) → **ACP Agent**
- **Vectora CLI** → **ACP Agent**
- Comunicam com IDE (VS Code, Claude Code, Antigravity) via **ACP (JSON-RPC over stdio)**
- Recebem prompts do usuário e executam actions (edit, terminal, etc)

### 2. Sub-Agent Mode

- **Vectora Core** → **MCP Server**
- Chamado como sub-agent por agente pai (Claude Code, Antigravity Chat)
- Expõe tools/resources via **MCP (JSON-RPC over stdio)**
- Recebe requests de contexto, embedding, query, etc

### 3. Internal Communication (unchanged)

- **VS Code Extension** ↔ **Vectora Core** → **IPC (JSON-RPC over named pipes/Unix socket)**
- **CLI** ↔ **Vectora Core** → **IPC (JSON-RPC)**

---

## Phase 7: Protocol Integration

### 7A. ACP Agent Implementation (VS Code Extension)

**Goal:** Make VS Code Extension a proper ACP Agent

**Files to modify/create:**

- `extensions/vscode/src/acp-agent.ts` (new)
- `extensions/vscode/src/extension.ts` (refactor)

**Implementation Steps:**

1. **Install ACP SDK**

   ```bash
   npm install @anthropic-ai/sdk-acp
   ```

2. **Create ACP Agent Handler**

   ```typescript
   // extensions/vscode/src/acp-agent.ts
   import { Agent, TextBlock, ImageBlock } from "@anthropic-ai/sdk-acp";

   export class VectoraACPAgent implements Agent {
     name = "vectora-vscode";
     version = "0.1.0";

     async initialize(): Promise<void> {
       // Connect to Vectora Core IPC
     }

     async createSession(): Promise<string> {
       // Create new session with Core
     }

     async prompt(sessionId: string, messages: TextBlock[]): Promise<TextBlock> {
       // Forward to Core via IPC, return response
     }

     async fileEdit(sessionId: string, path: string, content: string): Promise<void> {
       // Execute file edit via Core's ACP handler
     }

     async terminal(sessionId: string, command: string): Promise<string> {
       // Execute terminal command via Core's ACP handler
     }
   }
   ```

3. **Create stdio connection in main thread**

   ```typescript
   // extensions/vscode/src/extension.ts
   import { newAgentSideConnection } from "@anthropic-ai/sdk-acp";

   const agent = new VectoraACPAgent();
   const connection = newAgentSideConnection(agent, process.stdout, process.stdin);
   await connection.start();
   ```

**Capabilities to expose:**

- `workspace.query` — RAG query against indexed files
- `file.edit` — Edit file in workspace
- `file.create` — Create new file
- `terminal.run` — Execute shell command in workspace
- `chat.history` — Retrieve conversation history

---

### 7B. ACP Agent Implementation (CLI)

**Goal:** Make CLI a proper ACP Agent (when invoked by IDEs)

**Files to create:**

- `cmd/agent/agent.go` (new)
- `cmd/agent/main.go` (new)

**Implementation Steps:**

1. **Create new binary: `vectora-agent`**

   ```bash
   # This binary runs the ACP agent protocol when invoked by IDE
   # Separate from `vectora` (which is CLI + core launcher)
   ```

2. **Install ACP Go SDK**

   ```bash
   go get github.com/coder/acp-go-sdk@v0.10.8
   ```

3. **Implement Agent struct**

   ```go
   // cmd/agent/agent.go
   package main

   import (
       "github.com/coder/acp-go-sdk"
   )

   type VectoraAgent struct {
       core *CoreClient // Connection to running Core instance
   }

   // Implement acp.Agent interface
   func (a *VectoraAgent) Initialize(ctx context.Context) error {
       // Connect to local Vectora Core via IPC
   }

   func (a *VectoraAgent) NewSession(ctx context.Context) (string, error) {
       // Create session in Core
   }

   func (a *VectoraAgent) Prompt(ctx context.Context, sessionID string, messages []acp.Message) (acp.Message, error) {
       // Forward to Core's query handler
   }
   ```

4. **Handle ACP messages**
   - `initialize` → Validate client, connect to Core
   - `session/new` → Create session with IPC
   - `session/prompt` → Forward text blocks to Core
   - `_file.edit`, `_terminal.run` → Special extension methods

---

### 7C. ACP Client in Core

**Goal:** Core can connect to client IDEs (VS Code, Claude Code, etc) to request permissions/actions

**Files to create:**

- `core/api/acp/client.go` (new)
- `core/api/acp/models.go` (update)

**Implementation Steps:**

1. **Create ACP Client wrapper**

   ```go
   // core/api/acp/client.go
   type ACPClient struct {
       connection *acp.ClientSideConnection
       sessionID  string
   }

   // For when Core needs IDE permission
   func (c *ACPClient) RequestFileEdit(ctx context.Context, path, content string) error {
       // Send file.edit request to IDE
   }

   func (c *ACPClient) RequestTerminalRun(ctx context.Context, cmd string) (string, error) {
       // Send terminal.run request to IDE
   }
   ```

2. **Update existing ACP server to also be a client**
   ```go
   // core/api/acp/server.go (existing) + client support
   // Can both handle agent requests AND make client requests
   ```

---

### 7D. MCP Server Implementation (Core as Sub-Agent)

**Goal:** Core exposes MCP server for when called by parent agent (Claude Code, Antigravity)

**Files to create:**

- `core/mcp/server.go` (new)
- `core/mcp/tools.go` (new)
- `core/mcp/resources.go` (new)

**Installation:**

```bash
go get github.com/modelcontextprotocol/go-sdk
```

**Implementation Steps:**

1. **Create MCP Server in Core startup**

   ```go
   // cmd/core/main.go - add MCP server alongside IPC

   import "github.com/modelcontextprotocol/go-sdk/mcp"

   func runCore() {
       // ... existing IPC setup ...

       // Start MCP server (stdio transport)
       mcpServer := mcp.NewServer(mcp.ServerOptions{
           Name:    "vectora",
           Version: "0.1.0",
       })

       // Register MCP tools
       registerMCPTools(mcpServer, vecStore, kvStore, router)

       go mcpServer.Run(context.Background(), os.Stdin, os.Stdout)
   }
   ```

2. **Define MCP Tools**

   ```go
   // core/mcp/tools.go

   type WorkspaceQuery struct {
       WorkspacePath string `json:"workspace_path"`
       Query         string `json:"query"`
       TopK          int    `json:"top_k"`
   }

   type WorkspaceQueryResult struct {
       Results    []ScoredChunk `json:"results"`
       Answer     string        `json:"answer,omitempty"`
   }

   func registerMCPTools(server *mcp.Server, ...) {
       server.Tool("workspace.query", &WorkspaceQuery{},
           func(ctx context.Context, req *WorkspaceQuery) (*WorkspaceQueryResult, error) {
               // Query workspace, return chunks + RAG answer
           },
       )

       server.Tool("workspace.embed", ...) // Embed file
       server.Tool("workspace.list", ...)  // List indexed workspaces
       server.Tool("chat.create", ...)     // Create conversation
       server.Tool("chat.append", ...)     // Append message to conversation
   }
   ```

3. **Define MCP Resources (optional)**

   ```go
   // core/mcp/resources.go

   func registerMCPResources(server *mcp.Server, ...) {
       // Resources expose read-only access to indexed content
       server.Resource("vectora://workspace/{id}/summary", ...)
       server.Resource("vectora://workspace/{id}/files", ...)
   }
   ```

---

### 7E. Wire ACP + MCP into Core

**Goal:** Core handles both ACP (as agent) and MCP (as sub-agent) requests

**Files to modify:**

- `cmd/core/main.go`
- `core/api/acp/server.go`

**Implementation:**

1. **Detect transport layer**

   ```go
   func runCore() {
       // At startup, detect if we're running as:
       // - Standalone (IPC only)
       // - ACP Agent (stdio with ACP protocol)
       // - MCP Server (stdio with MCP protocol)

       if isACPTransport() {
           startACPServer()
       } else if isMCPTransport() {
           startMCPServer()
       } else {
           startIPCServer() // Default: local IPC mode
       }
   }
   ```

2. **Create protocol detector**

   ```go
   // core/protocols/detector.go

   func DetectTransport() Transport {
       // Read first byte/line from stdin
       // If JSON-RPC with "jsonrpc" field → could be ACP or MCP
       // Send "initialize" or "initialize_request" to determine
       // Return TRANSPORT_ACP, TRANSPORT_MCP, or TRANSPORT_IPC
   }
   ```

---

### 7F. Update Go Dependencies

**Add to `go.mod`:**

```go
require (
    github.com/coder/acp-go-sdk v0.10.8
    github.com/modelcontextprotocol/go-sdk v0.4.0  // or latest
)
```

---

## Phase 8: TypeScript/JavaScript ACP Client (VS Code Extension)

**Files to create:**

- `extensions/vscode/src/acp-transport.ts` (new)
- `extensions/vscode/src/acp-handler.ts` (new)

**Tasks:**

1. Install `@anthropic-ai/sdk-acp` npm package
2. Implement `Agent` interface for Vectora VSCode chat
3. Handle `initialize`, `session/new`, `session/prompt` messages
4. Implement extension methods for file editing, terminal execution
5. Maintain session state and sync with Core via IPC

---

## Data Flow Examples

### Example 1: Claude Code → Vectora as MCP Sub-Agent

```
Claude Code (parent agent)
    ↓ (MCP server.start)
    Vectora Core (MCP server)
        ↓ (MCP tools)
        {workspace.query, workspace.embed, chat.create}
    ↓ (result)
Claude Code (uses tool result in reasoning)
```

### Example 2: VS Code IDE → Vectora VSCode Extension → Vectora Core

```
VS Code IDE
    ↓ (ACP client connection)
    Vectora VSCode Extension (ACP agent)
        ↓ (IPC)
        Vectora Core
            ↓ (query, embed, etc)
        ↓ (IPC response)
    ↓ (ACP response blocks)
VS Code IDE (renders response)
```

### Example 3: CLI as ACP Agent

```
IDE (VS Code, Claude Code, etc)
    ↓ (spawns subprocess)
    vectora-agent (ACP agent binary)
        ↓ (IPC)
        Vectora Core
            ↓ (query, edit, terminal)
        ↓ (IPC response)
    ↓ (ACP response)
IDE (receives result)
```

---

## Dependencies & SDKs

| Component                    | SDK                                      | Language   | Version |
| ---------------------------- | ---------------------------------------- | ---------- | ------- |
| VS Code Extension            | `@anthropic-ai/sdk-acp`                  | TypeScript | Latest  |
| Vectora Core (as Agent)      | `github.com/coder/acp-go-sdk`            | Go         | v0.10.8 |
| Vectora Core (as MCP Server) | `github.com/modelcontextprotocol/go-sdk` | Go         | Latest  |
| Vectora CLI (as Agent)       | `github.com/coder/acp-go-sdk`            | Go         | v0.10.8 |

---

## Testing Strategy

1. **Unit Tests**
   - ACP message marshaling/unmarshaling
   - MCP tool invocation
   - Protocol detection logic

2. **Integration Tests**
   - Start Core as MCP server, invoke tools via MCP client
   - Start VS Code Extension, send ACP requests
   - Full flow: IDE → Extension/CLI → Core → query result

3. **End-to-End Tests**
   - Claude Code invokes Vectora as MCP sub-agent
   - VS Code IDE connects to Vectora VSCode Extension (ACP)
   - CLI invoked by IDE sends queries via ACP

---

## Rollout Order

1. **Phase 7A:** ACP Agent in VS Code Extension (highest priority)
2. **Phase 7D:** MCP Server in Core (critical for Claude Code integration)
3. **Phase 7B:** CLI as ACP Agent (lower priority, can work alongside 7A)
4. **Phase 7C:** ACP Client in Core (internal, lower priority)
5. **Phase 7E:** Wire both protocols into Core startup
6. **Phase 7F:** TypeScript ACP implementation
7. **Phase 8:** Full end-to-end testing

---

## Success Criteria

- [ ] `vectora-agent` binary can be invoked as ACP agent by VS Code
- [ ] VS Code Extension communicates with Vectora Core via ACP (over stdio)
- [ ] Vectora Core exposes MCP server on stdio
- [ ] Claude Code can call Vectora as MCP sub-agent via `invoke_tool`
- [ ] All protocol transitions (detect → initialize → session → prompt) work correctly
- [ ] Error handling graceful (invalid protocol → proper error response)
- [ ] IPC internal communication unaffected by new protocols

---

## Notes

- **ACP uses JSON-RPC 2.0** (same as IPC, so familiar)
- **MCP also uses JSON-RPC 2.0** over stdio (standard for sub-agents)
- **stdin/stdout transport** is standard for both (no additional network binaries needed)
- **Backwards compatibility:** Old IPC mode still works for CLI users
- **Protocol auto-detection** at startup based on first message shape
