# Vectora Models SDK Implementation Plan

Esse documento outline as fases referentes a integração de SDKs oficiais de modelos (Anthropic, Gemini, Voyage, OpenAI).

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
- **OpenAI Models (April 2026):**
  - Flagship: `gpt-5.4`, `gpt-5.4-pro`
  - Efficient/Agentic: `gpt-5.4-mini`, `gpt-5.4-nano`
  - *Note: GPT-4o was retired April 3, 2026.*
- **Qwen Models (April 2026):**
  - Frontier: `qwen3.6-plus` (1M context), `qwen-max`
  - Stable: `qwen-plus`, `qwen-turbo`, `qwen-flash`
- **Embeddings:** `text-embedding-3-small`, `text-embedding-3-large`

### 4E. Streaming Error Handling (Decision #15)

- Gemini: SDK manages reconnection; capture iterator errors
- Claude & OpenAI: `stream.Err()` after loop; send accumulated partial content via `message.Accumulate(event)`
- In both: JSON-RPC error notification with partial content + "Retry" button in UI

---

## Verification

- `vectora ask "test"` doesn't 404 (Phase 0)
- `go test ./core/llm/...` passes with each SDK; streaming works end-to-end (Phase 4)
