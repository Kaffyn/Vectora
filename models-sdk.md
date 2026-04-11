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

### 4D. Streaming Error Handling (Decision #15)

- Gemini: SDK manages reconnection; capture iterator errors
- Claude: `stream.Err()` after loop; send accumulated partial content via `message.Accumulate(event)`
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

---

## Dependency Graph

```
Phase 0 ──┐
Phase 1 ──┼── (all parallel, no deps)
Phase 2 ──┐
           │── Phase 2.5 (Depends on Phase 2 for lock file logic)
Phase 1 ──┘
           │
Phase 3 ───── (gate for Phase 4: SDKs need proper error propagation)
           │
Phase 4 ───── (depends on Phase 3)
           │
Phase 5 ───── (depends on Phase 2 for pprof port)
Phase 6 ───── (depends on Phase 2 + Phase 5 + Phase 2.5 for updater)
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
  logs show no API keys
- **Phase 6:** Binary update + rollback tested manually; workspace IDs differ across installations
