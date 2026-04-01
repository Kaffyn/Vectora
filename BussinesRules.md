# BUSINESS RULES: VECTORA — GOVERNANCE CONTRACT

> [!TIP]
> Read this file in another language.
> English [BUSINESS_RULES.md] | Portuguese [BUSINESS_RULES.pt.md]

This document establishes the architectural boundaries, internal API contracts, and mandatory business rules of Vectora. Any implementation that violates these boundaries must be refactored immediately. This file is the Single Source of Truth (SSOT). No complex change may be implemented before its rule is documented here.

---

## 1. PHILOSOPHY AND ENGINEERING RIGOR

Vectora rejects **"Vibe-Coding"** — programming by intuition, guesswork, or convenience. Every line of business logic is treated as an industrial engineering commitment.

### 1.1 Pair Programming and Governance

- **Radical Code Detachment:** If the code fails, the failure is in communication or architecture. The fix is made through dialogue and documentation adjustment, never through manual patches.
- **SSOT (Single Source of Truth):** This file is the Iron Law. Before any complex change, the rule must be documented here first.
- **Language:** Code and technical documentation in **English**. Dialogue and pair programming tone in **Portuguese**.

### 1.2 TDD Protocol (Red-Green-Refactor)

No business logic exists without a test that justifies it.

1. **RED:** Write the failing test, defining the contract.
2. **GREEN:** Implement the minimum code to pass.
3. **REFACTOR:** Optimize while keeping tests passing.

### 1.3 The 300% Standard (Iron Law)

Every feature must be proven by at least **3 variations** in the same test suite:

1. **Happy Path:** Ideal base scenario.
2. **Negative:** Invalid input or expected failure.
3. **Edge Case:** Complex combinations, concurrency, boundary limits.

---

## 2. ARCHITECTURAL PILLARS

These are the non-negotiable constraints from which all rules derive.

| Pillar                  | Constraint                                                                                                                                                    |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Local First**         | Every core feature must work without internet. Cloud providers are opt-in, never dependencies.                                                                |
| **Low Footprint**       | Systray daemon ≤ 5MB RSS at idle. Full system ≤ 4GB RSS under load.                                                                                           |
| **Pure Go**             | No CGO, no C++ dependencies, no runtime interpreters unless no viable Go alternative exists. Exceptions: `llama.cpp` (inference sidecar), `Fyne` (installer). |
| **No Shared State**     | Interface binaries hold zero application state. All state lives in the daemon.                                                                                |
| **Write Protection**    | No filesystem write or shell execution may occur without a prior GitBridge snapshot.                                                                          |
| **Workspace Isolation** | No workspace may read from another workspace's vector collection at the storage layer.                                                                        |

---

## 3. REPOSITORY STRUCTURE (ENFORCED)

```markdown
vectora/
├── cmd/                        # Entry points only. No business logic.
│   ├── vectora/                # Systray daemon (core orchestrator)
│   ├── vectora-cli/            # CLI binary (Bubbletea)
│   ├── vectora-web/            # Web UI binary (Wails)
│   └── vectora-installer/      # Installer binary (Fyne)
│
├── internal/                   # All business logic. Not importable externally.
│   ├── core/                   # RAG pipeline, workspace management, session logic
│   ├── db/                     # chromem-go and bbolt wrappers
│   ├── llm/                    # langchaingo-based provider abstraction and implementations
│   ├── ipc/                    # IPC server (daemon) and client (interfaces)
│   ├── tray/                   # Systray UI and process lifecycle
│   ├── tools/                  # Agentic toolkit (filesystem, search, shell, web, memory)
│   ├── git/                    # GitBridge: snapshot and rollback
│   ├── mcp/                    # MCP server implementation
│   ├── acp/                    # ACP agent implementation
│   ├── index/                  # Vectora Index HTTP client
│   └── infra/                  # Cross-cutting: logging, config, error types
│
├── pkg/                        # Public packages. Add only after team discussion.
│   └── vectorakit/             # Future external SDK
│
├── web/                        # Next.js frontend (static export, embedded via go:embed)
├── index-server/               # Vectora Index HTTP server (Go, net/http)
├── assets/                     # Embedded static assets (icons, default configs)
├── scripts/                    # Build, release, setup scripts
├── tests/                      # Integration and end-to-end test suites
├── docs/                       # Developer documentation
├── go.mod
├── go.sum
└── Makefile
```

### 3.1 Package Boundary Rules

- `cmd/` contains only: flag parsing, dependency wiring, and the call to start the root service. No `if`, no business logic.
- `internal/` packages must not import each other circularly. The dependency graph flows: `core → db`, `core → llm`, `tools → git`, `mcp/acp → core`, `ipc → core`.
- `internal/infra/` may be imported by any `internal/` package. It must never import other `internal/` packages.
- `pkg/` is frozen until the first stable release.
- `index-server/` is a standalone Go module. It shares no `internal/` packages with the main module.

---

## 4. THE DAEMON: SYSTRAY AS CORE ORCHESTRATOR

The `cmd/vectora` binary is the single source of truth for all runtime state. It must be the first process started and the last to exit.

### 4.1 Responsibilities

- Owns and manages all workspace state.
- Owns and manages all active LLM provider connections via `internal/llm`.
- Exposes the IPC server for all interface connections.
- Spawns and reaps interface processes on demand.
- Executes the GitBridge before any tool write operation.

### 4.2 Process Lifecycle

```markdown
System Login
    └── cmd/vectora starts
            └── IPC server binds to socket
            └── Workspaces hydrated from bbolt
            └── LLM provider initialized via langchaingo
            └── llama.cpp sidecar started (if Qwen mode)
            └── Systray icon rendered
                    └── User triggers CLI / Web UI
                            └── Daemon spawns interface process
                            └── Interface connects via IPC
                            └── Interface exits → Daemon remains
```

### 4.3 Interface Spawn Rules

- Interfaces are spawned only when explicitly requested by the user via the systray menu.
- Only one instance of each interface binary may run at a time. Attempting to spawn a second instance must focus the existing one.
- If an interface binary crashes, the daemon logs the event and cleans up the process handle. No crash in an interface may affect daemon stability.

---

## 5. WEB UI: WAILS + NEXT.JS ARCHITECTURE

The Web UI (`cmd/vectora-web`) is a Wails application that embeds the Next.js frontend as a static export.

### 5.1 Build Model

- Next.js must be configured with `output: 'export'`. This produces a fully static site with no Node.js server requirement.
- The static build output is embedded into the Wails binary via `//go:embed`.
- No Node.js runtime runs at any point during normal application operation.
- The frontend must be fully functional with no external CDN dependencies. All assets are self-contained.

### 5.2 Frontend ↔ Go Communication

The frontend communicates with the Go backend exclusively through **Wails bindings**. There is no HTTP server, no fetch to localhost, no WebSocket.

```go
// Go — App struct methods are exposed to the frontend via wails.Bind
type App struct {
    ipcClient *ipc.Client
}

func (app *App) QueryWorkspace(workspaceID string, query string) (QueryResponse, error) {
    return app.ipcClient.Send("workspace.query", map[string]any{
        "workspace_id": workspaceID,
        "query":        query,
    })
}
```

```typescript
// TypeScript — Wails auto-generates typed bindings from Go methods
import { QueryWorkspace } from "../wailsjs/go/main/App";

const response = await QueryWorkspace("godot-4.2", "how do I use signals?");
```

Wails generates TypeScript bindings automatically from exported Go methods. The binding layer is type-safe and requires no manual maintenance.

### 5.3 Web UI Scope

The Web UI is the primary user-facing interface. Its scope includes: the main chat experience, workspace creation and management, provider configuration (API key entry, model selection), and Vectora Index browsing and dataset download. It is not a generic web application and must not be designed as one.

### 5.4 Web UI Rules

- **BR-WEB-01:** The App struct in `cmd/vectora-web` must not contain business logic. It is a thin binding layer that delegates all calls to the IPC client.
- **BR-WEB-02:** No direct database or LLM calls may originate from the Wails App struct. All data flows through IPC to the daemon.
- **BR-WEB-03:** The Next.js build must pass `next build` with `output: 'export'` without errors before any PR touching `web/` is merged.
- **BR-WEB-04:** Wails binding methods must follow the same naming conventions defined in Section 9.

---

## 6. AI ENGINE: LANGCHAINGO + LLAMA.CPP

All AI capabilities — completion, embedding, and tool calling — are mediated through `internal/llm`, which is built on top of `langchaingo`.

### 6.1 langchaingo as the Provider Abstraction

`langchaingo` provides the unified interface for all LLM and embedding providers. It is the Go equivalent of LangChain/LlamaIndex, abstracting away provider-specific SDKs behind common interfaces.

- **Gemini** is integrated via langchaingo's Google AI provider, using the user's API key.
- **Future providers** (Claude, OpenAI, Ollama, etc.) are added by implementing or configuring the corresponding langchaingo provider. No changes to `internal/core` are required.
- `langchaingo` lives entirely within `internal/llm`. No other package imports it directly.

### 6.2 llama.cpp as Local Inference Sidecar

`llama.cpp` handles offline model execution for Qwen. It runs as a separate process (sidecar) and exposes a local HTTP server that `internal/llm` communicates with.

```
cmd/vectora (daemon)
    └── internal/llm
            └── llama.cpp process (sidecar)
                    └── HTTP server on localhost (loopback only)
                    └── Qwen GGUF model loaded
```

- The sidecar is started by the daemon on startup when Qwen mode is active.
- The sidecar binds only to loopback (`127.0.0.1`). It must never be exposed on a network interface.
- If the sidecar crashes, the daemon attempts one automatic restart before surfacing an error to the user.
- The sidecar port is allocated dynamically and stored in daemon state. It is never hardcoded.

### 6.3 Provider Interface

Despite using langchaingo internally, `internal/llm` still exposes its own interface to the rest of the system. This shields `internal/core` from langchaingo's API changes.

```go
// Provider is the contract all LLM providers must satisfy.
type Provider interface {
    // Complete sends a prompt and returns the model's response.
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)

    // Embed returns the vector embedding for the given input string.
    Embed(ctx context.Context, input string) ([]float32, error)

    // Name returns the canonical provider identifier ("qwen" | "gemini").
    Name() string

    // IsConfigured reports whether the provider has valid credentials or binaries.
    IsConfigured() bool
}

type CompletionRequest struct {
    Messages     []Message
    SystemPrompt string
    MaxTokens    int
    Temperature  float32
    Tools        []ToolDefinition // nil if no tool use needed
}

type CompletionResponse struct {
    Content   string
    ToolCalls []ToolCall // nil if no tool use
    Usage     TokenUsage
}
```

### 6.4 Gemini API Key Management

- The Gemini API key is stored in `~/.vectora/config.json`, encrypted at rest using the OS keychain where available (macOS Keychain, Windows Credential Manager, Linux Secret Service).
- The key is loaded into memory once at provider initialization and never written to logs, IPC payloads, or error messages.
- The key is passed to langchaingo's Google AI provider at initialization time and not stored in any other `internal/` package.
- Rotation: the user may update the key via the Web UI settings panel. The provider is re-initialized immediately with the new key without requiring a daemon restart.

### 6.5 Provider Rules

- **BR-LLM-01:** No package outside `internal/llm` may import `langchaingo` or any provider SDK directly.
- **BR-LLM-02:** `Complete` must be context-cancellable. A cancelled context must abort the in-flight request and return `ctx.Err()`.
- **BR-LLM-03:** `Embed` must return a deterministic vector for the same input string given the same model.
- **BR-LLM-04:** The Gemini API key must never appear in logs, error messages, IPC payloads, or crash reports.
- **BR-LLM-05:** The llama.cpp sidecar port must be dynamically allocated. Port 8080 or any other fixed port must not be hardcoded.
- **BR-LLM-06:** Adding a new provider requires only a new implementation of `Provider` in `internal/llm`. No changes to `internal/core`, `internal/ipc`, or any other package are permitted as part of a provider addition.

---

## 7. IPC CONTRACT

The IPC layer (`internal/ipc`) is the communication backbone between the daemon and all interfaces. This section defines its protocol, message format, and error handling.

### 7.1 Transport

IPC uses **Unix Domain Sockets** on Linux/macOS and **Named Pipes** on Windows. The socket path is:

```markdown
~/.vectora/run/vectora.sock   (Linux/macOS)
\\.\pipe\vectora               (Windows)
```

The daemon binds on startup. Clients connect on spawn. The socket file is removed on clean daemon shutdown.

### 7.2 Message Format

All messages are **newline-delimited JSON** (`\n` as frame delimiter). Each message is a JSON object with the following envelope:

```json
{
  "id": "string (UUIDv4, unique per request)",
  "type": "string (request | response | event)",
  "method": "string (only for type=request)",
  "payload": "object (method-specific)",
  "error": "object | null (only for type=response)"
}
```

**Request** — sent by an interface to the daemon:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "request",
  "method": "workspace.query",
  "payload": { "workspace_id": "godot-4.2", "query": "how do I use signals?" }
}
```

**Response** — sent by the daemon back to the requesting interface:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "response",
  "payload": { "answer": "...", "sources": [] },
  "error": null
}
```

**Event** — sent proactively by the daemon to all connected interfaces:

```json
{
  "id": "new-uuid",
  "type": "event",
  "method": "workspace.indexed",
  "payload": { "workspace_id": "godot-4.2", "chunk_count": 1482 }
}
```

### 7.3 IPC Methods (Request/Response)

#### Workspace

| Method                 | Payload                                        | Response                      |
| ---------------------- | ---------------------------------------------- | ----------------------------- |
| `workspace.list`       | `{}`                                           | `{ workspaces: Workspace[] }` |
| `workspace.create`     | `{ name, source_path }`                        | `{ workspace_id }`            |
| `workspace.delete`     | `{ workspace_id }`                             | `{}`                          |
| `workspace.activate`   | `{ workspace_id }`                             | `{}`                          |
| `workspace.deactivate` | `{ workspace_id }`                             | `{}`                          |
| `workspace.query`      | `{ workspace_id, query, active_workspaces[] }` | `{ answer, sources[] }`       |
| `workspace.index`      | `{ workspace_id }`                             | `{ job_id }` (async)          |

#### Provider

| Method         | Payload                  | Response                                             |
| -------------- | ------------------------ | ---------------------------------------------------- |
| `provider.get` | `{}`                     | `{ provider: "qwen" \| "gemini", configured: bool }` |
| `provider.set` | `{ provider, api_key? }` | `{}`                                                 |

#### Tools

| Method         | Payload                       | Response                  |
| -------------- | ----------------------------- | ------------------------- |
| `tool.execute` | `{ tool_name, args: object }` | `{ result, snapshot_id }` |
| `tool.undo`    | `{ snapshot_id }`             | `{ restored: bool }`      |

#### Index

| Method           | Payload                | Response                  |
| ---------------- | ---------------------- | ------------------------- |
| `index.browse`   | `{ query?, filters? }` | `{ datasets: Dataset[] }` |
| `index.download` | `{ dataset_id }`       | `{ job_id }` (async)      |
| `index.publish`  | `{ path, metadata }`   | `{ submission_id }`       |

#### Session

| Method            | Payload      | Response                  |
| ----------------- | ------------ | ------------------------- |
| `session.history` | `{ limit? }` | `{ messages: Message[] }` |
| `session.clear`   | `{}`         | `{}`                      |

### 7.4 IPC Events (Daemon → Interface, Proactive)

| Event Method              | Payload                         | Description                |
| ------------------------- | ------------------------------- | -------------------------- |
| `workspace.indexed`       | `{ workspace_id, chunk_count }` | Indexing job completed     |
| `workspace.index_failed`  | `{ workspace_id, error }`       | Indexing job failed        |
| `index.download_progress` | `{ job_id, percent }`           | Download progress update   |
| `index.download_complete` | `{ job_id, workspace_id }`      | Download finished          |
| `tool.snapshot_created`   | `{ snapshot_id, tool_name }`    | GitBridge snapshot created |
| `daemon.status`           | `{ ram_mb, workspaces_loaded }` | Periodic health broadcast  |

### 7.5 Error Object

```json
{
  "code": "string (machine-readable, snake_case)",
  "message": "string (human-readable)",
  "detail": "object | null (optional structured context)"
}
```

**Canonical Error Codes:**

| Code                       | Meaning                                          |
| -------------------------- | ------------------------------------------------ |
| `workspace_not_found`      | The referenced workspace_id does not exist       |
| `workspace_already_active` | Workspace is already in the active set           |
| `provider_not_configured`  | No LLM provider has been set up                  |
| `tool_not_found`           | The requested tool_name does not exist           |
| `snapshot_failed`          | GitBridge could not create a snapshot            |
| `index_signature_invalid`  | Downloaded dataset failed signature verification |
| `ipc_method_unknown`       | The requested method is not registered           |
| `ipc_payload_invalid`      | Payload failed schema validation                 |
| `internal_error`           | Unhandled daemon error (always logged)           |

### 7.6 IPC Rules

- **BR-IPC-01:** Every request must receive exactly one response with the matching `id`. No request may go unanswered.
- **BR-IPC-02:** Events carry a new UUID and are broadcast to all connected clients. Clients must not assume event ordering.
- **BR-IPC-03:** The IPC server must handle client disconnection gracefully. A disconnected client must not affect daemon state or other connected clients.
- **BR-IPC-04:** Message size limit is 4MB per frame. Payloads exceeding this must use streaming (chunked event sequence).
- **BR-IPC-05:** The daemon must validate every incoming payload against its schema before processing. Invalid payloads return `ipc_payload_invalid` immediately.

---

## 8. INTERNAL API CONTRACTS

### 8.1 `internal/db` — Storage Layer

Two storage engines, each with a dedicated wrapper. They must not be used interchangeably.

**chromem-go (Vector Store):**

```go
type VectorStore interface {
    UpsertChunk(ctx context.Context, collection string, chunk Chunk) error
    Query(ctx context.Context, collection string, query string, topK int) ([]ScoredChunk, error)
    DeleteCollection(ctx context.Context, collection string) error
    CollectionExists(ctx context.Context, collection string) bool
}
```

**bbolt (Key-Value Store):**

```go
type KVStore interface {
    Set(ctx context.Context, bucket string, key string, value []byte) error
    Get(ctx context.Context, bucket string, key string) ([]byte, error)
    Delete(ctx context.Context, bucket string, key string) error
    List(ctx context.Context, bucket string, prefix string) ([]string, error)
}
```

**Rules:**

- Each workspace maps to exactly one chromem-go collection and one bbolt bucket, both named `ws:<workspace_id>`.
- `DeleteCollection` and bucket deletion must be called together when a workspace is destroyed.
- The `db` package must not contain any business logic. It is a pure storage adapter.

### 8.2 `internal/core` — RAG Pipeline

```go
type RAGPipeline interface {
    Query(ctx context.Context, req QueryRequest) (QueryResponse, error)
    IndexWorkspace(ctx context.Context, workspaceID string) error
    Acti
```
