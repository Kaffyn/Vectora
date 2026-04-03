# Contributing to Vectora

> [!TIP]
> Read this file in another language.
> English [CONTRIBUTING.md] | Portuguese [CONTRIBUTING.pt.md]

This document is intended for developers who want to contribute to Vectora. It covers the project philosophy, repository structure, architecture decisions, and the rules that keep the codebase consistent.

---

## Philosophy

Vectora is built around three non-negotiable principles:

**Local first.** Every core feature must work without internet access. Cloud providers (Gemini) are opt-in extensions, never dependencies.

**Low footprint.** The systray daemon must remain under 100MB RAM RSS in idle. The full system must operate under 4GB RAM on modest hardware. Every dependency added must justify its memory and binary size cost.

**Pure Go where possible.** Avoid CGO bindings, heavy C++ dependencies, and runtime interpreters unless there is no viable Go alternative. Exceptions are `Fyne` (installer, CGO/OpenGL) and `llama-cli` (used as a sidecar via pipes).

---

## Repository Structure

Vectora follows standard Go project layout conventions.

```markdown
vectora/
├── cmd/                        # Binary entry points (one per executable)
│   ├── vectora/                # Main orchestrator (systray daemon) & CLI
│   │   └── main.go
│   ├── vectora-app/            # Web UI binary (Wails)
│   │   └── main.go
│   └── vectora-installer/      # Installer binary (Fyne)
│       └── main.go
│
├── internal/                   # Private packages (not importable externally)
│   ├── app/                    # Next.js frontend (Source code and Assets)
│   │   ├── app/                # Next.js App Router
│   │   ├── components/         # React Components
│   │   └── public/             # Static Assets
│   ├── core/                   # Business logic: RAG pipeline, workspace management
│   │   ├── rag.go
│   │   ├── workspace.go
│   │   └── indexer.go
│   ├── db/                     # Database layer
│   │   ├── vector.go           # chromem-go wrapper
│   │   └── store.go            # bbolt wrapper
│   ├── llm/                    # LLM provider abstraction
│   │   ├── provider.go         # Interface definition
│   │   ├── gemini.go           # Gemini implementation
│   │   └── qwen.go             # Qwen/llama-cli implementation
│   ├── ipc/                    # IPC server and client (systray ↔ interfaces)
│   │   ├── server.go
│   │   └── client.go
│   ├── tray/                   # Systray UI and lifecycle management
│   │   └── tray.go
│   ├── tools/                  # Agentic toolkit (shared across MCP, ACP, CLI)
│   │   ├── filesystem.go       # read_file, write_file, read_folder, edit
│   │   ├── search.go           # find_files, grep_search
│   │   ├── web.go              # google_search, web_fetch
│   │   ├── shell.go            # run_shell_command
│   │   └── memory.go           # save_memory, enter_plan_mode
│   ├── git/                    # GitBridge: snapshot and rollback
│   │   └── bridge.go
│   ├── engines/                # Binary management for sidecars
│   │   └── manager.go
│   ├── mcp/                    # MCP server implementation
│   │   └── server.go
│   ├── acp/                    # ACP agent implementation
│   │   └── agent.go
│   └── index/                  # Vectora Index client (catalog, download, publish)
│       └── client.go
│
├── pkg/                        # Public packages (safe to import externally)
│   └── vectorakit/             # SDK for external integrations (future)
│
├── assets/                     # Embedded static assets (icons, default configs)
├── scripts/                    # Build, release, and setup scripts
├── docs/                       # Developer documentation
├── tests/                      # Integration and end-to-end tests
│   └── suite/
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── CONTRIBUTING.md
```

### Key Rules

- `cmd/` contains only entry points. No business logic lives here — only initialization, flag parsing, and wiring.
- `internal/` is where all logic lives. Packages here cannot be imported by external projects.
- `pkg/` is reserved for code intentionally exposed to external consumers. Do not add to `pkg/` without discussion.
- `app/` is the Next.js frontend. It is embedded into the Wails binary at build time and must not depend on any external CDN at runtime.
- Cross-cutting concerns (logging, config, errors) belong in `internal/infra/`, not scattered across features.

---

## Architecture

### The Systray as Core Daemon

The `cmd/vectora` binary is the single source of truth for application state. It runs as a background process from login and exposes an IPC server that all other interfaces connect to.

```markdown
cmd/vectora (systray daemon)
    └── IPC Server (Named Pipes / Unix Sockets)
            ├── cmd/vectora-app   (spawned on demand)
            ├── vectora chat / cli (integrated into daemon binary)
            └── MCP / ACP clients (external)
```

No interface binary holds state. They are stateless clients. If an interface crashes, the daemon and its state are unaffected.

### IPC Contract

All communication between the daemon and interfaces goes through the IPC layer in `internal/ipc`. Adding a new interface means implementing the IPC client, not duplicating core logic. The protocol is JSON-ND for streaming efficiency.

### LLM Provider Interface

All LLM interactions go through the `internal/llm.Provider` interface. New providers must implement this interface. No provider-specific code leaks into `internal/core`. Local inference (Qwen) is handled via zero-port process pipes (`internal/llm/protocol_llama.go`).

```markdowngo
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
    Embed(ctx context.Context, input string) ([]float32, error)
    Name() string
}
```

### Tool Execution and GitBridge

Every tool in `internal/tools` that performs a write or shell operation must call `internal/git.Bridge.Snapshot()` before execution. This is mandatory, not optional. The snapshot enables the `undo` command across all interfaces, providing industrial-grade rollback safety.

### Workspace Isolation

Each workspace maps to an isolated chromem-go collection and a dedicated bbolt bucket. Workspaces must never share a collection or cross-query without explicit user intent. The `internal/core` package enforces this boundary at the ingestion and retrieval layers.

---

## Business Rules

These rules define constraints that must be respected across the entire codebase.

**BR-01 — RAM Budget**
The systray daemon must not exceed 100MB RSS at idle. The full system (daemon + one active interface + one loaded workspace) must not exceed 4GB RSS. Any PR that measurably degrades this must include justification and profiling data.

**BR-02 — No Network in Core**
`internal/core`, `internal/db`, `internal/tools`, and `internal/git` must not make outbound network calls. Network access is restricted to `internal/llm` (provider calls) and `internal/index` (Index client).

**BR-03 — Workspace Isolation**
A workspace may only read from its own vector collection. Cross-workspace queries are not permitted at the storage layer. Aggregation, if needed, happens at the `internal/core` RAG layer, never at the `internal/db` layer.

**BR-04 — Write Protection**
No tool may write to the filesystem or execute shell commands without a prior Git snapshot via `internal/git.Bridge`. Tests must verify snapshot creation before write operations.

**BR-05 — Provider Abstraction**
No file outside `internal/llm` may import a provider SDK directly (Gemini SDK, llama-cli pipes). Provider details are fully encapsulated.

**BR-06 — Interface Statelessness**
Interface binaries (`vectora-app`, CLI modes) hold no application state. All state lives in the daemon. Interfaces are disposable and should reconnect to the IPC on restart.

**BR-07 — Embedded Frontend**
The Next.js frontend must be fully buildable and functional without any external CDN dependency. All assets must be embeddable via `go:embed` from `internal/app/out`.

**BR-08 — Index Curation**
The Vectora Index client (`internal/index`) may only download datasets that carry a valid Kaffyn review signature. The client must reject unsigned or tampered datasets at download time.

---

## Development Setup

### Requirements

- Go 1.22+
- Node.js 20+ (for the web frontend)
- Bun (for frontend builds)
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- `llama-cli` binary (for local inference, managed via `internal/engines`)

### Building

```bash
# Build all binaries via powershell script
./build.ps1

# Build a specific binary via Makefile
make build-tray
make build-web
make build-installer

# Run the daemon in development mode
go run ./cmd/vectora daemon
```

### Testing

```bash
# Run all tests
go test ./...

# Run the integration suite using the daemon flag
go run ./cmd/vectora --tests
```

All PRs must pass the full test suite including the race detector before review.

---

## Commit Convention

Vectora uses [Conventional Commits](https://www.conventionalcommits.org/).

```markdown
<type>(scope): <description>

Types: feat, fix, docs, refactor, test, chore, perf
Scope: core, cli, web, tray, installer, mcp, acp, ipc, db, tools, index, git
```

**Examples:**

```markdown
feat(core): add multi-workspace query aggregation
fix(tools): ensure git snapshot precedes shell execution
perf(db): reduce chromem-go collection load time
docs(contributing): add workspace isolation rules
```

Commits that skip the convention will not be merged.

---

## Pull Request Rules

- Every PR must reference an open issue.
- PRs touching `internal/core` or `internal/db` require two approvals.
- PRs must not introduce new CGO dependencies without prior discussion.
- All new public functions in `internal/` must have Go doc comments.
- Breaking changes to the IPC contract require a migration path documented in the PR.

---

## What Not to Do

- Do not add state to interface binaries.
- Do not call provider SDKs outside of `internal/llm`.
- Do not bypass the GitBridge for write operations.
- Do not add packages to `pkg/` without team discussion.
- Do not embed external CDN URLs in the frontend.
- Do not introduce dependencies that require a running server process (no Postgres, no Redis, no external vector DBs).

---

_Part of the [Kaffyn](https://github.com/Kaffyn) open source organization._
