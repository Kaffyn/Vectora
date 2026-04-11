# Vectora Issue Report - Implementation Plan

## Context

O projeto Vectora tem 9 bugs, 10 decisões arquiteturais pendentes de implementação e 3 requisitos de modernização. Os problemas centrais são: identificadores de modelo com sufixo `-preview` faltando (causando 404), falha no carregamento da webview, todos os providers LLM usando HTTP puro (`net/http`) em vez dos SDKs oficiais (google/genai, anthropic-sdk-go, voyageai), implementação manual de JSON-RPC, e falta de polimento na CLI. Este plano endereça todos os 22 itens do relatório em ordem de dependência.

---

## Phase 0: Critical Bug Fixes (Immediate)

### 0A. Fix Gemini Model Identifiers (Issue #9)

- **File:** `core/llm/gemini_provider.go:31-37`
- Current: `"gemini-3-flash"`, `"gemini-3.1-pro"`, `"gemini-embedding-2-preview"`
- Os modelos existem mas os IDs corretos da API incluem o sufixo `-preview`:
  - `"gemini-3-flash"` → `"gemini-3-flash-preview"`
  - `"gemini-3.1-pro"` → `"gemini-3.1-pro-preview"`
  - `"gemini-embedding-2-preview"` → **já está correto** (confirmado na doc oficial)
- Claude aliases em `core/llm/claude_provider.go:76-86`: os modelos Claude 4.6 existem. O SDK Go usa constantes como `anthropic.ModelClaudeOpus4_6`, `anthropic.ModelClaude4_6Sonnet`. Alinhar aliases com as constantes do SDK na Phase 4.

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

## Phase 1: CLI UX (Quick Wins)

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

## Phase 2: Singleton & Process Management (Decisions #10, Issues #3, #4)

### 2A. Hybrid File Lock + PID Validation

- **Files:** `core/os/linux/linux.go`, `core/os/macos/macos.go` (replace TCP port binding)
- New cross-platform: Write PID to `~/.vectora/vectora.pid` + `flock()` on Unix
- Keep Windows mutex as-is (already works), add PID file as supplementary

### 2B. Graceful Shutdown

- **File:** `cmd/core/main.go` - signal handlers to clean up PID file on SIGTERM/SIGINT

---

## Phase 3: JSON-RPC Library Migration (Decision #19)

### 3A. Go Core: Adopt `sourcegraph/jsonrpc2`

- **Files:** `core/api/jsonrpc/`, `core/api/ipc/server.go`, `core/api/ipc/router.go`
- Add dependency, rewrite handler registration to use library's handler interface
- Migrate method-by-method

### 3B. VS Code Extension: Adopt `vscode-jsonrpc`

- **File:** `extensions/vscode/src/client.ts`
- Replace manual framing with `createMessageConnection`

### 3C. IPC Security Handshake (Decision #16)

- Core: Generate token on startup → `~/.vectora/ipc.token`
- Extension: Read token, send in `initialize` request
- Core: Reject connections without valid token

---

## Phase 4: LLM SDK Migration (Decisions #11, #20, #21)

### 4A. Gemini → `google.golang.org/genai`

- **File:** `core/llm/gemini_provider.go` - full rewrite usando SDK oficial
- Substituir `net/http` manual por `genai.NewClient()` + `client.Models.GenerateContent()`
- Modelos confirmados (docs oficiais 2026-04):
  - Chat: `gemini-3-flash-preview`, `gemini-3.1-pro-preview`
  - Embedding: `gemini-embedding-2-preview` (3072 dims)
  - Também disponíveis: `gemini-2.5-flash`, `gemini-2.5-pro`
- Streaming via SDK nativo com callbacks

### 4B. Claude → `github.com/anthropics/anthropic-sdk-go` (v1.27.1+)

- **File:** `core/llm/claude_provider.go` - full rewrite usando SDK oficial
- Requer Go 1.23+
- Usar constantes do SDK: `anthropic.ModelClaudeOpus4_6`, `anthropic.ModelClaude4_6Sonnet`, etc.
- `client := anthropic.NewClient(option.WithAPIKey(apiKey))`
- Chat: `client.Messages.New(ctx, anthropic.MessageNewParams{...})`
- Streaming: `client.Messages.NewStreaming(ctx, params)` + loop `stream.Next()`/`stream.Current()`
- Tool calling nativo via `anthropic.ToolParam` + `anthropic.ToolUnionParam`
- Retries automáticos (2x default) para 429/5xx
- Remover structs manuais `claudeRequest`, `claudeResponse`, `claudeMessage`, `claudeTool`

### 4C. Voyage → `github.com/austinfhunter/voyageai`

- **File:** `core/llm/voyage_provider.go` - rewrite usando SDK oficial
- `vo := voyageai.NewClient(voyageai.VoyageClientOpts{Key: apiKey})`
- Embedding: `vo.Embed(texts, voyageai.ModelVoyageCode3, &EmbeddingRequestOpts{InputType: "document"})`
- Modelos confirmados: `ModelVoyageCode3`, `ModelVoyage3Large`, `ModelVoyage35`, etc.
- Também suporta: Reranking (`vo.Rerank`) e Multimodal embedding

### 4D. Streaming Error Handling (Decision #15)

- Gemini: SDK gerencia reconexão; capturar erros do iterator
- Claude: `stream.Err()` após loop; enviar partial content acumulado via `message.Accumulate(event)`
- Em ambos: notificação JSON-RPC de erro com conteúdo parcial + botão "Retry" na UI

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
Phase 2 ──┘
           │
Phase 3 ───── (gate for Phase 4: SDKs need proper error propagation)
           │
Phase 4 ───── (depends on Phase 3)
           │
Phase 5 ───── (depends on Phase 2 for pprof port)
Phase 6 ───── (depends on Phase 2 + Phase 5)
```

## Verification

- **Phase 0:** `go build ./...` succeeds; extension loads webview without error; `vectora ask "test"` doesn't 404
- **Phase 1:** `vectora workspace ls` shows paths; `vectora workspaces` works; `vectora config set INVALID x` warns
- **Phase 2:** Starting two instances shows "already running"; `vectora status` reports correct state
- **Phase 3:** Extension connects via `vscode-jsonrpc`; IPC token auth rejects unauthorized clients
- **Phase 4:** `go test ./core/llm/...` passes with each SDK; streaming works end-to-end
- **Phase 5:** `curl localhost:<debug-port>/debug/pprof/` works; logs show no API keys
- **Phase 6:** Binary update + rollback tested manually; workspace IDs differ across installations
