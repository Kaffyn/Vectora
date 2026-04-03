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
| **Low Footprint**       | Systray daemon ≤ 100MB RSS at idle. Full system ≤ 4GB RSS under load.                                                                                           |
| **Pure Go**             | No CGO, no C++ dependencies, no runtime interpreters unless no viable Go alternative exists. Exceptions: `llama-cli` (subprocess sidecar), `Fyne` (installer). |
| **No Shared State**     | Interface binaries hold zero application state. All state lives in the daemon.                                                                                |
| **Write Protection**    | No filesystem write or shell execution may occur without a prior GitBridge snapshot.                                                                          |
| **Workspace Isolation** | No workspace may read from another workspace's vector collection at the storage layer.                                                                        |

---

## 3. REPOSITORY STRUCTURE (ENFORCED)

```markdown
vectora/
├── cmd/                        # Entry points only. No business logic.
│   ├── vectora/                # Systray daemon (core orchestrator) & CLI
│   ├── vectora-web/            # Web UI binary (Wails)
│   └── vectora-installer/      # Installer binary (Fyne)
│
├── internal/                   # All business logic. Not importable externally.
│   ├── app/                    # Next.js frontend (Source code and Assets)
│   ├── core/                   # RAG pipeline, workspace management, session logic
│   ├── db/                     # chromem-go and bbolt wrappers
│   ├── llm/                    # langchaingo-based provider abstraction and implementations
│   ├── ipc/                    # IPC server (daemon) and client (interfaces)
│   ├── tray/                   # Systray UI and process lifecycle
│   ├── tools/                  # Agentic toolkit (filesystem, search, shell, web, memory)
│   ├── git/                    # GitBridge: snapshot and rollback
│   ├── engines/                # Asset management for sidecars (llama-cli, etc)
│   ├── mcp/                    # MCP server implementation
│   ├── acp/                    # ACP agent implementation
│   ├── index/                  # Vectora Index HTTP client
│   └── infra/                  # Cross-cutting: logging, config, error types
│
├── pkg/                        # Public packages. Add only after team discussion.
│   └── vectorakit/             # Future external SDK
│
├── assets/                     # Embedded static assets (icons, default configs)
├── index-server/               # Vectora Index HTTP server (Go, net/http)
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
- Exposes the IPC server (Named Pipes/Sockets) for all interface connections.
- Spawns and reaps interface processes on demand.
- Executes the GitBridge before any tool write operation.

### 4.2 Process Lifecycle

```markdown
System Login
    └── cmd/vectora starts
            └── IPC server binds to socket
            └── Workspaces hydrated from bbolt
            └── LLM provider initialized via langchaingo
            └── llama-cli subprocess started (if Qwen mode) via Pipes
            └── Systray icon rendered
                    └── User triggers CLI / Web UI
                            └── Daemon spawns interface process
                            └── Interface connects via IPC
                            └── Interface exits → Daemon remains
```

---

## 5. WEB UI: WAILS + NEXT.JS ARCHITECTURE

The Web UI (`cmd/vectora-web`) is a Wails application that embeds the Next.js frontend as a static export from `internal/app`.

### 5.1 Build Model

- Next.js (`internal/app`) must be configured with `output: 'export'`. This produces a fully static site with no Node.js server requirement.
- The static build output is embedded into the Wails binary via `//go:embed`.
- No Node.js runtime runs at any point during normal application operation.
- The frontend must be fully functional with no external CDN dependencies. All assets are self-contained.

### 5.2 Frontend ↔ Go Communication

The frontend communicates with the Go backend exclusively through **Wails bindings** which resolve to **IPC calls** to the daemon.

```go
// Go — App struct methods in cmd/vectora-web act as an IPC proxy
type App struct {
    ipcClient *ipc.Client
}

func (app *App) QueryWorkspace(workspaceID string, query string) (string, error) {
    // Delegates to IPC
    return app.ipcClient.Call("workspace.query", ...)
}
```

---

## 6. AI ENGINE: ZERO-PORT & PIPES

All AI capabilities — completion, embedding, and tool calling — are mediated through `internal/llm`.

### 6.1 Zero-Port Architecture (Local)

- **llama-cli Integration:** Local models run as subprocesses. Communication is handled via Standard I/O Pipes (Stdin/Stdout).
- No local network ports (TCP) are opened for inference, ensuring total process isolation and security.
- Streaming is supported natively via JSON-ND over the pipes.

### 6.2 langchaingo as the Provider Abstraction

`langchaingo` provides the unified interface for all cloud providers (like Gemini).

- **Gemini** is integrated via langchaingo's Google AI provider.
- `langchaingo` lives entirely within `internal/llm`. No other package imports it directly.

### 6.3 Provider Interface

```go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
    Embed(ctx context.Context, input string) ([]float32, error)
    Name() string
    IsConfigured() bool
}
```

---

## 7. IPC CONTRACT (NAMED PIPES / SOCKETS)

The IPC layer (`internal/ipc`) is the communication backbone.

### 7.1 Transport

- **Windows:** Named Pipes (`\\.\pipe\vectora`).
- **Unix:** Sockets (`~/.vectora/run/vectora.sock`).

### 7.2 Message Format

All messages are **newline-delimited JSON** (JSON-ND). Each message follows a standard request/response/event envelope.

### 7.3 IPC Methods

| Method | Payload | Description |
| --- | --- | --- |
| `workspace.query` | `{ query, context[] }` | RAG Answer |
| `workspace.index` | `{ path }` | Async indexing |
| `tool.execute` | `{ tool, args }` | Agentic action |

---

## 8. INTERNAL API CONTRACTS

### 8.1 `internal/db` — Storage Layer

**chromem-go (Vector Store):**
- Isolated collections per Workspace ID.
- Mandatory pure Go implementation.

**bbolt (Key-Value Store):**
- Persistence for logs, chat history, and metadata.
- Thread-safe access via bbolt transaction API.

### 8.2 `internal/tools` — Security

- **BR-TOOL-01:** Tools that alter the filesystem must trigger a Git snapshot (GitBridge) or local backup before execution.
- **BR-TOOL-02:** The system must provide a revert mechanism (Undo) for all agentic actions.

[End of Business Rules - Revision 2026.04.03]
