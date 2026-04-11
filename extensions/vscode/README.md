# Vectora AI — VS Code Extension

> **Local AI coding assistant** — RAG-powered chat + agentic tools via ACP (Agent Client Protocol).

Vectora operates in two distinct modes tailored for different development needs:

- **Agent Mode:** A high-fidelity, interactive chat panel for direct developer interaction.
- **Sub-Agent Mode:** Integrated background Specialist that can be invoked by other agents (like **Antigravity**) to perform repo-wide research and execution.

[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/Kaffyn/Vectora)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## Why Vectora?

Unlike generic AI assistants that only see your current editor window, Vectora **understands your entire codebase**. It combines RAG (Retrieval-Augmented Generation) with agentic tools, all within a secure, locally-scoped Trust Folder.

| Feature      |    Copilot / Generic AI    |               Vectora               |
| :----------- | :------------------------: | :---------------------------------: |
| **Context**  |    Current window only     |        Full codebase via RAG        |
| **Privacy**  | Data sent to cloud servers |     Local index + Trust Folder      |
| **Actions**  |   Text suggestions only    |   Read / Write / Execute securely   |
| **Protocol** |        Proprietary         |      Open ACP (interoperable)       |
| **Index**    |            None            | chromem-go vectors + BBolt metadata |

---

## Architecture

```
┌─────────────────────┐     ACP (JSON-RPC 2.0)     ┌──────────────────────┐
│  VS Code Extension  │ ◄──────────────────────────► │  Vectora Core        │
│  (TypeScript)       │     stdio · newline JSON    │  vectora acp (Go)    │
│                     │                              └──────────┬───────────┘
│  ┌───────────────┐  │                                         │
│  │  ACPClient    │  │                              ┌──────────▼───────────┐
│  │  ChatPanel    │  │                              │  Engine              │
│  │  BinaryMgr    │  │                              │  RAG · Tools · LLM   │
│  │  InlineComp   │  │                              └──────────┬───────────┘
│  └───────────────┘  │                                         │
└─────────────────────┘                              ┌──────────▼───────────┐
                                                     │  Gemini / Claude API │
                                                     │  (HTTPS, cloud)      │
                                                     └──────────────────────┘
```

- **ACP** — Agent Client Protocol over stdio
- **JSON-RPC 2.0** — message framing (newline-delimited)
- **chromem-go** — local vector index (no external DB)
- **BBolt** — local metadata persistence

**Security principle:** All network traffic goes Core → Cloud API only.
The extension ↔ Core channel is 100% local via stdio pipes.

Vectora uses **official SDKs** (`google-golang.org/genai`, `anthropic-sdk-go`, `voyageai`) to communicate with AI providers, ensuring maximum stability and security.

---

## Features

### RAG-Powered Chat

Ask questions about your codebase. Vectora semantically searches your indexed files, reads the relevant code, and provides context-aware answers — streamed token-by-token.

### Agentic Tools (Core-side)

Vectora's core can autonomously invoke tools inside the Trust Folder:

| Tool           | What it does                                    |
| -------------- | ----------------------------------------------- |
| `read_file`    | Reads file content (max 50 KB, with truncation) |
| `grep_search`  | Regex search across all project files           |
| `terminal_run` | Executes shell commands (30 s timeout sandbox)  |

All tool calls are gated by the **Guardian** — a compile-time policy layer that blocks access to secrets, keys, databases, and binaries.

### Streaming Responses

Watch answers appear token-by-token. Tool calls, agent plans, and reasoning are visible in real-time inside the chat panel.

### Inline Completion Provider (Stub)

A `VectoraInlineProvider` is registered as a VS Code `InlineCompletionItemProvider`. It respects VS Code's cancellation token and is ready to be wired to any fast completion backend.

> **Note:** The inline completion backend call is a planned feature. The provider is structurally complete; connect it to your preferred fast LLM endpoint when ready.

### Workspace Isolation — Trust Folder

All operations are scoped to your project folder. The Guardian enforces file-level policies — blocked extensions (`.env`, `.key`, `.db`, `.exe`, etc.) are silently skipped.

### Automatic Binary Management

The `BinaryManager` class handles binary resolution in priority order:

1. User-configured `vectora.corePath` setting
2. Managed install directory (`~/.vectora/bin/`)
3. System `PATH` (`where vectora` / `which vectora`)
4. Auto-download from GitHub Releases (with progress notification)

---

## Prerequisites

### 1. Vectora Core Binary

The extension spawns `vectora acp <workspace>` over stdio. Get the binary one of these ways:

**Option A — Auto-Download (recommended)**

On first use, the extension detects if the binary is missing and offers to download it automatically to `~/.vectora/bin/`.

**Option B — Download from Releases**

```
https://github.com/Kaffyn/Vectora/releases
```

Place `vectora.exe` (Windows) or `vectora` (Linux/macOS) in your `PATH`, or configure `vectora.corePath`.

**Option C — Build from Source**

```bash
git clone https://github.com/Kaffyn/Vectora.git
cd Vectora
go build -o vectora ./cmd/core/
# Move to PATH or set vectora.corePath in VS Code settings
```

**Verify:**

```bash
vectora --version
# Expected: vectora version 0.1.0
```

### 2. API Key

Vectora uses cloud LLM/embedding APIs. Configure your key:

```
# Windows: %USERPROFILE%\.Vectora\.env
# Linux/macOS: ~/.Vectora/.env

GEMINI_API_KEY=your_gemini_api_key_here
# or
CLAUDE_API_KEY=your_claude_api_key_here
# or
VOYAGE_API_KEY=your_voyage_api_key_here
```

**Supported providers:**

| Provider           | Chat | Embeddings             |
| ------------------ | ---- | ---------------------- |
| Gemini (Google)    | ✅   | ✅ Native              |
| Claude (Anthropic) | ✅   | ⚠️ Fallback embeddings |
| Voyage AI          | ❌   | ✅ High-Fidelity       |

### 3. Open Workspace

The extension requires an open VS Code workspace folder. All tool operations are restricted to this folder.

---

## Installation

### From VSIX File

1. Download `vectora-vscode-0.1.0.vsix` from [Releases](https://github.com/Kaffyn/Vectora/releases)
2. Open VS Code → **Extensions** (`Ctrl+Shift+X`)
3. Click **`···`** → **"Install from VSIX..."**
4. Select the `.vsix` file
5. **Restart VS Code** when prompted

### From Command Line

```bash
code --install-extension vectora-vscode-0.1.0.vsix
```

### First Launch

1. Open a project folder in VS Code
2. The extension activates automatically
3. If the binary is missing, accept the download prompt
4. Ensure `%USERPROFILE%\.Vectora\.env` contains a valid API key
5. Click the **Vectora icon** in the Activity Bar or use the Command Palette

---

## Configuration

Open VS Code Settings (`Ctrl+,`) and search for `vectora`:

| Setting             | Type    | Default          | Description                                           |
| ------------------- | ------- | ---------------- | ----------------------------------------------------- |
| `vectora.corePath`  | string  | `"vectora"`      | Path to the Vectora binary. Full path if not in PATH. |
| `vectora.workspace` | string  | _(first folder)_ | Override workspace root for RAG indexing.             |
| `vectora.streaming` | boolean | `true`           | Enable streaming responses (token-by-token).          |

**`settings.json` example:**

```json
{
  "vectora.corePath": "C:\\Users\\you\\.vectora\\bin\\vectora.exe",
  "vectora.workspace": "C:\\Projects\\my-app",
  "vectora.streaming": true
}
```

---

## Usage

### Chat Panel

1. Click the **Vectora icon** in the Activity Bar
2. Type your question and press **Enter** (or `Shift+Enter` for new line)
3. Watch the response stream in — tool calls appear as they execute
4. Press **Stop** to cancel a running request

### Context Menu Commands

Select code in the editor, then **right-click**:

| Command                    | Action                             |
| -------------------------- | ---------------------------------- |
| **Vectora: Explain Code**  | Opens chat with an explain request |
| **Vectora: Refactor Code** | Opens chat with a refactor request |

If no code is selected, the entire file content is used.

### Command Palette

Open with `Ctrl+Shift+P` and type `Vectora`:

| Command                  | Description                   |
| ------------------------ | ----------------------------- |
| `Vectora: New Session`   | Clear history and start fresh |
| `Vectora: Explain Code`  | Explain the selected code     |
| `Vectora: Refactor Code` | Refactor the selected code    |

### Status Bar

The status bar item (bottom-right) shows the current state:

| Icon                     | Meaning                                |
| ------------------------ | -------------------------------------- |
| `⟳ Vectora starting...`  | Binary spawning, handshake in progress |
| `✓ Vectora: <model>`     | Connected and ready                    |
| `○ Vectora Disconnected` | Core not running                       |
| `✗ Vectora Error`        | Error — check tooltip for details      |

---

## Security

### Trust Folder

All file system operations are scoped to the open workspace folder. The extension cannot instruct the core to access files outside the project root.

### Guardian — Hard-Coded Policies

These file types are blocked at the engine level, regardless of what the LLM requests:

| Category  | Patterns                        |
| --------- | ------------------------------- |
| Secrets   | `.env`, `.env.*`                |
| Keys      | `.key`, `.pem`, `.p12`, `.pfx`  |
| Databases | `.db`, `.sqlite`                |
| Binaries  | `.exe`, `.dll`, `.so`, `.dylib` |
| SSH Keys  | `id_rsa`, `id_ed25519`, `*.pub` |

### Permission Requests

For write operations, the core emits a `session/request_permission` notification. The extension surfaces this as a VS Code modal dialog:

- **Allow** — permit this specific operation
- **Deny** — block the operation

---

## ACP Protocol Reference

The extension communicates with the core using the **Agent Client Protocol** (JSON-RPC 2.0 over stdio).

### Client → Core (Requests)

| Method               | Description                              |
| -------------------- | ---------------------------------------- |
| `initialize`         | Handshake — declares client capabilities |
| `session/new`        | Creates a new agent session              |
| `session/prompt`     | Sends user prompt, returns `stopReason`  |
| `session/cancel`     | Cancels an in-progress prompt (notify)   |
| `fs/read_text_file`  | Reads a file via the core                |
| `fs/write_text_file` | Writes a file via the core               |

### Core → Client (Notifications)

| Method                       | Description                                     |
| ---------------------------- | ----------------------------------------------- |
| `session/update`             | Streaming chunk, tool call event, or plan entry |
| `session/request_permission` | Permission gate for write/execute tools         |

### `session/update` Variants

| `sessionUpdate` field | Payload                                     |
| --------------------- | ------------------------------------------- |
| `agent_message_chunk` | Streamed text token from the model          |
| `tool_call`           | Tool invocation started (id, title, status) |
| `tool_call_update`    | Tool status change (pending → completed)    |
| `plan`                | Agent plan entries (⏳ / ✅)                |

---

## Troubleshooting

### "Vectora starting..." never resolves

**Check binary:**

```powershell
# Windows
where vectora

# Linux/macOS
which vectora
```

**Check API key:**

```powershell
# Windows
type $env:USERPROFILE\.Vectora\.env

# Linux/macOS
cat ~/.Vectora/.env
```

**Check logs:**

```
%USERPROFILE%\.Vectora\logs\core.log
```

### "Vectora failed to start"

1. Set `vectora.corePath` to the full binary path
2. Ensure the workspace folder is open
3. Verify `GEMINI_API_KEY` is present in `.env`

### Responses are slow on the first query

The first query triggers codebase indexing (embedding generation). Subsequent queries use the cached chromem-go index and are significantly faster.

To pre-index manually:

```bash
vectora embed --workspace /path/to/project
```

---

## Development

### Prerequisites

- Node.js 20+
- npm 10+
- VS Code 1.90+
- TypeScript 5+

### Build

```bash
cd extensions/vscode
npm install
npm run compile          # webpack production build
npm run watch            # rebuild on file changes
npx vsce package         # generate .vsix
```

### Debug

1. Open `extensions/vscode` in VS Code
2. Press `F5` → Extension Development Host opens
3. Set breakpoints in `src/*.ts`
4. Test in the new VS Code window

### Source Structure

```
extensions/vscode/
├── src/
│   ├── extension.ts         # Activation, commands, status bar
│   ├── client.ts        # ACP JSON-RPC client (stdio transport)
│   ├── chat-panel.ts        # WebviewPanel with streaming UI
│   ├── binary-manager.ts    # Binary resolution + auto-download
│   ├── inline-completion.ts # InlineCompletionItemProvider (stub)
│   └── types/
│       └── acp.d.ts         # ACP protocol type definitions
├── dist/
│   └── extension.js         # Bundled output (webpack)
├── media/
│   ├── icon.svg
│   └── icon.ico
├── package.json
├── tsconfig.json
└── webpack.config.js
```

---

## Roadmap

| Feature                          | Status     |
| -------------------------------- | ---------- |
| RAG chat via ACP                 | ✅ Shipped |
| Streaming responses              | ✅ Shipped |
| Tool call visualization          | ✅ Shipped |
| Plan visualization               | ✅ Shipped |
| Auto binary download             | ✅ Shipped |
| Inline completion provider       | ⚙️ Stub    |
| Diff provider (write edits)      | Planned    |
| Terminal integration             | Planned    |
| Slash commands (`/embed`)        | Planned    |
| Keyboard shortcut `Ctrl+Shift+V` | Planned    |

---

## License

MIT — see [LICENSE](LICENSE)

---

**Vectora** — Part of the [Kaffyn](https://github.com/Kaffyn) open source organization.
