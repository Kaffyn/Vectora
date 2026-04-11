# Vectora Issue Report - Implementation Plan

## Context

The Vectora project has 9 bugs, 10 architectural decisions pending implementation, and 3 modernization requirements. The core issues are: missing model identifiers with the `-preview` suffix (causing 404s), webview loading failure, all LLM providers using raw HTTP (`net/http`) instead of the official SDKs (google/genai, anthropic-sdk-go, voyageai), manual JSON-RPC implementation, and a lack of CLI polish. This plan addresses all 22 items in the report in order of dependency.

---

## Phase 0: Critical Bug Fixes (Immediate)

### 0A. Fix Gemini Model Identifiers (Issue #9)

- **File:** `core/llm/gemini_provider.go:31-37`
- Current: `"gemini-3-flash"`, `"gemini-3.1-pro"`, `"gemini-embedding-2-preview"`
- The models exist but the correct API IDs include the `-preview` suffix:
  - `"gemini-3-flash"` → `"gemini-3-flash-preview"`
  - `"gemini-3.1-pro"` → `"gemini-3.1-pro-preview"`
  - `"gemini-embedding-2-preview"` → **already correct** (confirmed in official docs)
- Claude aliases in `core/llm/claude_provider.go:76-86`: Claude 4.6 models exist. The Go SDK uses constants like `anthropic.ModelClaudeOpus4_6`, `anthropic.ModelClaude4_6Sonnet`. Align aliases with SDK constants in Phase 4.

### 0B. Fix Webview Load Failure (Issue #1)

- **File:** `extensions/vscode/src/extension.ts:25`
- Problem: `new ChatViewProvider(undefined as any, context)` passes null client
- Fix: ChatViewProvider must handle null client gracefully - show "Connecting..." state instead of crashing
- **File:** `extensions/vscode/src/chat-panel.ts` - add null-client guard in message handlers

### 0C. Fix Binary Naming Mismatch (Issues #2, #4)

- **File:** `extensions/vscode/src/binary-manager.ts`
- Problem: Looks for `vectora.exe` but build produces `vectora-windows-amd64.exe`
- Fix: Standardize on `vectora.exe` in build output OR add both names to resolution chain
- Also clean up `stop` command to kill processes matching both names

---

## Phase 1: CLI UX (Quick Wins) [COMPLETED]

### 1A. Config Key Validation (Issue #5)

- **File:** `cmd/core/config.go`
- Add valid key whitelist + help text showing accepted keys

### 1B. Workspace Path Display (Issue #6)

- **File:** `cmd/core/workspace.go`
- Store path metadata alongside workspace collections, display `ID → /path` in `workspace ls`

### 1C. Command Aliases (Issue #7)

- **File:** `cmd/core/main.go` (command registration)
- Add Cobra `Aliases`: `workspaceCmd.Aliases = []string{"workspaces", "ws"}`

### 1D. Windows Defender Documentation (Issue #8)

- Document code signing process in CONTRIBUTING.md or release docs (not a code fix)

---

## Phase 2: Singleton & Process Management [COMPLETED]

### 2A. Hybrid File Lock + PID Validation

- **Files:** `core/os/linux/linux.go`, `core/os/macos/macos.go` (replace TCP port binding)
- New cross-platform: Write PID to `~/.vectora/vectora.pid` + `flock()` on Unix
- Keep Windows mutex as-is (already works), add PID file as supplementary

### 2B. Graceful Shutdown

- **File:** `cmd/core/main.go` - signal handlers to clean up PID file on SIGTERM/SIGINT

---

## Phase 2.5: Windows AppData Directory Restructuring [COMPLETED]

### 2.5A. Native Pathing Strategy

- **Installation (%LOCALAPPDATA%\Programs\Vectora)**:
  - `vectora.exe` (Main), `updater.exe` (Aux), `version.json` (Meta).
  - Aligned with per-user installs that don't roam between machines.
- **Data & Config (%APPDATA%\Vectora)**:
  - `.env`, `.lock`, `trust_paths.txt`, `workspaces.json`.
  - `data/` (BBolt, Chromem, Memory).
  - `logs/` (Rotation: log + log.old).
  - Aligned with user state that should follow them in corporate domains.

### 2.5B. Core Changes

- **File:** `core/os/manager_windows.go`
  - `GetInstallDir()`: return `%LOCALAPPDATA%\Programs\Vectora`
  - `GetAppDataDir()`: return `%APPDATA%\Vectora`
- **File:** `core/service/singleton/singleton.go`
  - `singleton.New()`: base lock on `%APPDATA%\Vectora\.lock`
- **Unix/macOS:** Remains in `~/.Vectora/` (correct native standard).

### 2.5C. Extension Changes

- **File:** `extensions/vscode/src/binary-manager.ts`
  - `VECTORA_BIN_DIR`: update to `%LOCALAPPDATA%\Programs\Vectora\vectora.exe`

---

## Phase 3: IPC & JSON-RPC Modernization [COMPLETED]

### 3A. Windows Named Pipes (`go-winio`)

- **File:** `core/api/ipc/server.go`
- Replaced Unix domain sockets with Named Pipes on Windows (`\\.\pipe\vectora`) for native stability.
- Implemented `winio.PipeConfig` with ACLs restricted to the current user (`CU`).
- Added TCP fallback mechanism (port 42781) for restricted environments.

### 3B. Standardized JSON-RPC 2.0 Errors

- **File:** `core/api/ipc/protocol.go`, `router.go`
- Adopted strict numeric codes (`-32603` for internal, etc.) and `slug` identifiers.
- Updated all IPC handlers to return structured `IPCError` objects.

### 3C. IPC Security Handshake

- **File:** `core/api/ipc/server.go`, `client.go`
- Core generates a cryptographically secure token at `%APPDATA%\Vectora\ipc.token` on startup.
- Clients must read this file and provide the token in the `initialize` handshake.
- Server rejects any connection with a missing or invalid token (`CodeUnauthorized`).

---

## Phase 4: LLM SDK Migration (Decisions #11, #20, #21)

### 4A. Gemini → `google.golang.org/genai`

- **File:** `core/llm/gemini_provider.go` - full rewrite using official SDK
- Replace manual `net/http` with `genai.NewClient()` + `client.Models.GenerateContent()`
- Confirmed Models (official docs 2026-04):
  - Chat: `gemini-3-flash-preview`, `gemini-3.1-pro-preview`
  - Embedding: `gemini-embedding-2-preview` (3072 dims)
  - Also available: `gemini-2.5-flash`, `gemini-2.5-pro`
- Native streaming via SDK with callbacks

### 4B. Claude → `github.com/anthropics/anthropic-sdk-go` (v1.27.1+)

- **File:** `core/llm/claude_provider.go` - full rewrite using official SDK
- Requires Go 1.23+
- Use SDK constants: `anthropic.ModelClaudeOpus4_6`, `anthropic.ModelClaude4_6Sonnet`, etc.
- `client := anthropic.NewClient(option.WithAPIKey(apiKey))`
- Chat: `client.Messages.New(ctx, anthropic.MessageNewParams{...})`
- Streaming: `client.Messages.NewStreaming(ctx, params)` + loop `stream.Next()`/`stream.Current()`
- Native tool calling via `anthropic.ToolParam` + `anthropic.ToolUnionParam`
- Automatic retries (2x default) for 429/5xx
- Remove manual structs `claudeRequest`, `claudeResponse`, `claudeMessage`, `claudeTool`

### 4C. Voyage → `github.com/austinfhunter/voyageai`

- **File:** `core/llm/voyage_provider.go` - rewrite using official SDK
- `vo := voyageai.NewClient(voyageai.VoyageClientOpts{Key: apiKey})`
- Embedding: `vo.Embed(texts, voyageai.ModelVoyageCode3, &EmbeddingRequestOpts{InputType: "document"})`
- Confirmed Models: `ModelVoyageCode3`, `ModelVoyage3Large`, `ModelVoyage35`, etc.
- Also supports: Reranking (`vo.Rerank`) and Multimodal embedding

### 4D. OpenAI / Qwen → `github.com/openai/openai-go`

- **File:** `core/llm/openai_provider.go` - implement using official SDK
- Support API base URL overrides for Qwen compatibility (`https://dashscope.aliyuncs.com/compatible-mode/v1`)
- Chat models: `gpt-4o`, `qwen-max`, `qwen-plus`, `qwen-turbo`
- Embeddings: `text-embedding-3-small`, `text-embedding-3-large`
- Full compatibility with OpenAI's format for structured responses and Tool Calling.

### 4E. Streaming Error Handling (Decision #15)

- Gemini: SDK manages reconnection; capture iterator errors
- Claude & OpenAI: `stream.Err()` after loop; send accumulated partial content via `message.Accumulate(event)`
- In both: JSON-RPC error notification with partial content + "Retry" button in UI

---

## Phase 5: Observability & Safety (Decisions #12, #17, #18)

### 5A. pprof Integration (Decision #12)

- **File:** `cmd/core/main.go` - add `net/http/pprof` on localhost debug port

### 5B. Log Sanitization (Decision #18)

- New middleware in `core/infra/` - redact API keys and PII from logs

### 5C. Vector DB Schema Versioning (Decision #17)

- Store schema version in bbolt. On mismatch → auto re-index with user notification

---

## Phase 6: Update System & Security (Decisions #13, #14, #22)

### 6A. Auto-Updater with Rollback (Decision #13)

- New package: `core/updater/` - check GitHub releases, download, swap binary, health check, rollback

### 6B. Workspace Salted Hashes (Decision #14)

- Per-installation salt in `~/.vectora/salt`, use `SHA256(salt + path)` for workspace IDs

### 6C. Security Audit (Decision #22)

- Review all changes: IPC auth, log sanitization, Guardian enforcement, path traversal checks

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

---

## Multi-Tenancy Protocol (MTP)

Este documento especifica a arquitetura e o protocolo para gerenciar **múltiplos projetos simultâneos (Multi-Tenancy)** em uma **única instância singleton** do Vectora Daemon.

O objetivo é manter o consumo de memória baixo (usando um único daemon rodando em background) enquanto se garante isolamento absoluto de estados de conversa, índices vetoriais de código, concorrência a provedores LLM, e rotinas de leitura de discos (Trust Folders).

---

## Phase 9: O Modelo de Tenant baseado em Conexão (Connection-Bound Tenancy)

No Vectora, um "Tenant" representa um Projeto ou Workspace aberto no editor (como uma janela do VS Code). Quando a extensão do VS Code se conecta ao Daemon via IPC (Named Pipes ou Unix Sockets), o protocolo MTP estabelece o escopo de atuação daquela conexão.

### A. Handshake de Autenticação e Contexto (IPC)

A mensagem `ipc.auth` ou uma nova mensagem `workspace.init` serve para estabelecer o contexto do Tenant de forma contínua durante essa conexão.

**Request (Client -> Daemon):**

```json
{
  "type": "request",
  "method": "workspace.init",
  "payload": {
    "workspace_root": "C:\\Users\\bruno\\Projects\\MyApp",
    "project_name": "MyApp"
  }
}
```

O Daemon gera um **WorkspaceID** previsível e consistente baseado no path absoluto (e.g. `sha256(workspace_root)`).
A partir desse momento, todas as trocas de mensagens na conexão daquele _socket_ estarão amarradas a este isolamento lógico.

---

## Phase 10: Abstração de Armazenamento e Estado (Storage & State Isolation)

Para garantir que o Projeto A nunca sobrescreva os dados do Projeto B, toda persistência e gerenciamento em memória é dinamicamente enclausurado.

### A. Estrutura de Pastas de Dados

Os dados no `%APPDATA%\Vectora\` (ou Unix equivalente) serão reestruturados:

```text
%APPDATA%\Vectora\
├── global.db                  # BBolt: Configs globais do Daemon e chaves API
├── ipc.token                  # Token de segurança local
└── workspaces\                # NÓS DO MULTI-TENANCY
    ├── <workspace_id_A>\      # Namespace fixo para o Projeto A
    │   ├── chromadb\          # Índice vetorial (Chromem-go) local isolado
    │   ├── chat_history.db    # Histórico de agentes e sessões
    │   └── guardian.json      # Regras de limite de acesso (Trust Folder)
    └── <workspace_id_B>\      # Namespace fixo para o Projeto B
        ├── chromadb\
        ├── chat_history.db
        └── guardian.json
```

### B. Gestor de Ciclo de Vida dos Workspaces em Memória

O Daemon manterá um registro Singleton de "Workspaces Ativos".

- Quando a primeira query `workspace.init` para `<id_A>` chega, o Daemon carega a coleção do `chromem-go` A em memória.
- Uma política de **Eviction** descarregará o Workspace B da memória RAM após 30 minutos sem nenhuma conexão IPC ativa solicitando-o. Isso economiza RAM.

---

## Phase 11: Segurança e Limites Absolutos (Trust Folders & Guardian)

O conceito mais perigoso em um Singleton é um agente num Projeto A ser comprometido e tentar ler as variáveis `.env` do Projeto B do usuário. Para evitar isso:

1. **Restrição por Contexto IPC:** A rotina do IPC nunca permite repassar `<workspace_id_B>` com a conexão originada e autorizada no `<workspace_id_A>`. O contexto `context.Context` passado aos Handlers já contém a Root de Segurança fixada.
2. **Guardian File Interceptor:** Qualquer acesso de File System feito pelos Handlers ou por chamadas da LLM vai passar por um `guardian.ValidatePath(requestedPath, tenantRoot)`. Se tentarem `../../../ProjetoSecreto/`, o daemon corta e retorna um `IPCError`.

---

## Phase 12: Paralelismo Seguro e Pool de Recursos (Resource Throttling)

Como usamos requisições externas para LLMs (ou locações de memória na GPU local), o Projeto A pode gerar bloqueios ou exceder os Rate Limits de uma Cloud (Anthropic/Voyage), afetando o Projeto B.

### A. CPU-Bound Priority Queue (Indexadores)

- Indexação e Embedding será delegada a Background Workers de baixa prioridade.
- A Janela Ativa tem uma requisição Foreground impulsionada na pool. A interface IPC enviará sinais passivos para sinalizar quem está no "Foco" do OS.

### B. IO-Bound Rate Limiting Semaphore (LLMs Calls)

- O daemon terá Semáforos por **Workspace**. O limite padrão pode ser 2 requisições paralelas ativas _por projeto_. Se houver pico num arquivo, o projeto enfileira suas próprias chamadas, enchendo a própria cota, enquanto o _Projeto B_ no painel ao lado tem a própria cota intocável e fluida.

---

## Phase 13: Fluxo da Arquitetura para Implementação

**Passo 1: `core/manager/tenant.go`**
Criar um construtor de tenants `GetOrCreateTenant(workspaceRoot)`. Este objeto encapsula suas próprias classes do DB, LLM Histories e Storage Engine (evitando passagem de parametros massiva nas rotinas).

**Passo 2: Injetar Tenant no Pipeline de Handlers IPC**
Modificar o Event Loop em `core/api/ipc/server.go`. Quando ler uma mensagem conectada num descritor, ele insere o respectivo `*Tenant` no Context usando `context.WithValue(ctx, TenantKey, activeTenant)`. Assim os Handlers só batem na DB daquele Context.

**Passo 3: Mapeamento Dinâmico de Database e Index**
Separar os diretórios de forma limpa. A inicialização do Chromem-Go deixará de ser Global para ser mapeada no `GetOrCreateTenant()`.

**Passo 4: Monitor de Auto-Desligamento (Eviction)**
Rotina em background (Ticker de x mins) que fecha e consolida/salva estados de um Tenant persistido cujo socket de interface (Janela do editor) não deu sinais de vida ou bateu em um timeout de socket.

---

## Dependency Graph

```
Phase 0 ──┐
Phase 1 ──┼── (all parallel, no deps)
Phase 2 ──┐
           │── Phase 2.5 (Depends on Phase 2 for lock file logic)
Phase 1 ──┘
           │
Phase 3 ───── (gate for Phase 4 & Phase 7)
           │
Phase 4 ──┼── Phase 5 ── Phase 6
           │
Phase 7 ──┼── Phase 8 (Depends on Phase 3 and Phase 4)
           │
Phase 9 ──┼── Phase 10 ── Phase 11 ── Phase 12 ── Phase 13 (MTP System)
```

## Verification

- **Phase 0:** `go build ./...` succeeds; extension loads webview without error; `vectora ask "test"` doesn't 404
- **Phase 1:** `vectora workspace ls` shows paths; `vectora workspaces` works; `vectora config set INVALID x` warns
- **Phase 2:** Starting two instances shows "already running"; `vectora status` reports correct state
- **Phase 2.5:** `vectora.exe` exists in `LocalAppData\Programs`; `.env` and `data/` exist in `Roaming\Vectora`
- **Phase 3:** Extension connects via `vscode-jsonrpc`; IPC token auth rejects unauthorized clients
- **Phase 4:** `go test ./core/llm/...` passes with each SDK; streaming works end-to-end
- **Phase 5:** `curl localhost:<debug-port>/debug/pprof/` works; logs show no API keys
- **Phase 6:** Binary update + rollback tested manually; workspace IDs differ across installations
- **Phase 7/8:** `vectora-agent` CLI + Extension can be consumed natively from IDE engines conforming to SDK-ACP specs. Core exposed via Model Context Protocol (MCP).
- **Phase 9-13:** Opening multiple IDE projects locally concurrently uses only a single background `vectora.exe` process that properly resolves paths per-project and securely compartmentalizes the vector search logic.
