# Vectora CLI — Command Reference

The Vectora Command-Line Interface (CLI) is a Cobra-based tool for managing the Vectora Core service, configuring API keys, indexing codebases, and querying knowledge bases.

## Quick Start

```bash
# Start the background service with system tray
vectora start

# Check service status
vectora status

# Index your project (generate embeddings)
vectora embed

# Ask a question about your codebase
vectora ask "How does authentication work in this project?"

# Stop the service
vectora stop
```

---

## Core Commands

### `vectora start` — Start the Background Service

Launches the Vectora Core as a background process with a system tray icon.

```bash
vectora start
```

**Aliases:** `vectora core`

**Behavior:**
- Spawns a background process running the ACP (Agent Client Protocol) server
- Displays system tray icon for quick access and status monitoring
- Loads configuration from `~/.vectora/config.json`
- Initializes vector database and metadata store

**Environment Variables:**
- `VECTORA_DETACHED=1` — Run in fully detached mode (internal use)

---

### `vectora status` — Check Service Status

Returns the current state of the Vectora Core service.

```bash
vectora status
```

**Output Examples:**
```
✓ Vectora Core is running (PID: 12345)
  Location: /home/user/.vectora/data
  Embeddings: 2,450 vectors indexed
  Last updated: 2 hours ago
```

```
✗ Vectora Core is not running
  Start with: vectora start
```

---

### `vectora stop` — Stop the Service

Gracefully shuts down the running Vectora Core service.

```bash
vectora stop
```

**Behavior:**
- Sends shutdown signal to the background process
- Closes IPC connections gracefully
- Saves state to persistent storage
- Waits up to 5 seconds for graceful shutdown

---

### `vectora restart` — Restart the Service

Stops and restarts the Vectora Core service in one command.

```bash
vectora restart
```

**Useful for:**
- Applying configuration changes
- Recovering from connection issues
- Refreshing the ACP server

---

### `vectora reset --hard` — Wipe All Data

⚠️ **DESTRUCTIVE** — Permanently deletes all indexed data and chat histories.

```bash
vectora reset --hard
```

**What Gets Deleted:**
- All vector embeddings
- All chat history
- Metadata database
- Configuration (user settings preserved)

**Requires:**
- `--hard` flag (prevents accidental data loss)

**Will:**
1. Stop the Core service
2. Wipe the data directory
3. Automatically restart with a fresh instance

---

## Embedding & Indexing

### `vectora embed [path]` — Generate Embeddings

Indexes your codebase by creating semantic vectors for all source files.

```bash
# Index current directory
vectora embed

# Index specific directory
vectora embed /path/to/project

# Index with progress output
vectora embed --verbose
```

**Flags:**
- `--verbose` — Show progress and per-file status
- `--force` — Re-index all files (ignores cache)
- `--exclude-dirs` — Comma-separated list of directories to skip (e.g., `node_modules,build,.git`)
- `--extensions` — Comma-separated file extensions to index (e.g., `.go,.ts,.py`)

**What Gets Indexed:**
- Source code files (`.go`, `.ts`, `.js`, `.py`, `.rs`, etc.)
- Documentation (`.md`, `.txt`, `.rst`)
- Configuration files (`.json`, `.yaml`, `.toml`)

**What Gets Skipped (by Guardian):**
- Hidden files (`.env`, `.key`, `.pem`)
- Binary files (`.exe`, `.dll`, `.so`, `.dylib`)
- Database files (`.db`, `.sqlite`)
- Package directories (`node_modules`, `vendor`, `.venv`)

**Output:**
```
Indexing /home/user/project...
✓ Scanned 1,240 files
✓ Created 2,450 embeddings
✓ Indexed 15 MB of code
✓ Completed in 23 seconds
```

---

## Querying & Chat

### `vectora ask [query]` — Ask About Your Code

Queries the knowledge base using RAG (Retrieval-Augmented Generation).

```bash
# Simple question
vectora ask "How does the API handle authentication?"

# Complex multi-file question
vectora ask "What are all the places where database connections are created?"

# With verbose output (shows retrieved files)
vectora ask "Why might the checkout flow be slow?" --verbose
```

**Flags:**
- `--verbose` — Show source files and relevance scores
- `--model` — Override the default model (e.g., `--model gpt-4`)
- `--temperature` — Control creativity (0.0–1.0, default: 0.7)
- `--context-limit` — Maximum context to include (default: 200k tokens)

**Output:**
```
Searching codebase for relevant context...
✓ Found 5 relevant files

Based on your codebase, here's how authentication works:

1. User Login Flow
   - Request arrives at /api/auth/login
   - Credentials validated against bcrypt hash
   - JWT token generated with 24-hour expiry

2. Protected Routes
   - JWT middleware intercepts requests
   - Token validated against secret key
   - User context injected into request handler

[Sources: auth.go, middleware.go, handlers.go]
```

---

## Configuration & Models

### `vectora config` — Manage Configuration

View and update Vectora configuration.

```bash
# Show current configuration
vectora config --list

# Set API key for Gemini
vectora config --gemini YOUR_API_KEY

# Set API key for Claude (Anthropic)
vectora config --anthropic YOUR_API_KEY

# Set API key for OpenAI
vectora config --openai YOUR_API_KEY

# Set default provider
vectora config --set-provider gemini

# Set default model
vectora config --set-model gemini-1.5-pro

# View configuration file location
vectora config --path
```

**Supported Providers:**
- `gemini` — Google Gemini (Gemini 3.1 Pro, Flash, Embedding 2.0)
- `anthropic` — Anthropic Claude (Claude 4.6 Sonnet, Claude 4.5 Haiku)
- `openai` — OpenAI GPT (GPT-5.4 Pro, GPT-5-o1)
- `voyageai` — Voyage AI Embeddings (Voyage-3 Large, Voyage-3 Code)

**Configuration File:**
```
~/.vectora/config.json
```

Example:
```json
{
  "providers": {
    "gemini": {
      "api_key": "AIza...",
      "embedding_model": "text-embedding-004"
    },
    "anthropic": {
      "api_key": "sk-ant-..."
    }
  },
  "default_provider": "gemini",
  "default_model": "gemini-1.5-pro",
  "trust_folder": "/home/user/my-project",
  "log_level": "info"
}
```

---

### `vectora models` — List Available Models

Displays all models available from configured providers.

```bash
vectora models

# Show only a specific provider
vectora models --provider gemini

# Show with detailed information
vectora models --verbose
```

**Output:**
```
Available Models:

GEMINI
  ✓ gemini-1.5-pro (Recommended)
    Context: 1M tokens | Embedding: Gemini Embedding 2.0

  ✓ gemini-1.5-flash (Fast)
    Context: 1M tokens | Cost: ~10x cheaper

  ✓ gemini-2.0-flash-exp
    Context: 1M tokens | Experimental

ANTHROPIC
  ✓ claude-4-6-sonnet (Recommended)
    Context: 200k tokens

  ✓ claude-4-5-haiku
    Context: 200k tokens | Fast & cheap

OPENAI
  ✓ gpt-5-4-pro
    Context: 128k tokens
```

---

## Protocol & Integration

### `vectora acp` — ACP Protocol Server

Starts the ACP (Agent Client Protocol) server for programmatic access.

```bash
# Start ACP server on default port
vectora acp

# Start on custom port
vectora acp --port 9000

# Enable debug logging
vectora acp --debug
```

**Protocol:**
- **JSON-RPC 2.0** over stdio/named pipes
- **Transport:** IPC (Named Pipes on Windows, Unix sockets on Linux/macOS)
- **Used by:** VS Code Extension, Gemini CLI Bridge, external clients

**Methods Available:**
- `session/new` — Create a new chat session
- `session/prompt` — Send a user message
- `session/list` — List active sessions
- `workspace/query` — RAG search
- `tools/list` — List available tools

---

## Advanced Usage

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VECTORA_HOME` | Configuration & data directory | `~/.vectora` |
| `VECTORA_LOG_LEVEL` | Logging verbosity (debug, info, warn, error) | `info` |
| `VECTORA_MODE` | Execution mode (normal, test, dev) | `normal` |
| `VECTORA_TRUST_FOLDER` | Path to index and operate within | Current directory |
| `GEMINI_API_KEY` | Google Gemini API key | _(from config)_ |
| `ANTHROPIC_API_KEY` | Anthropic Claude API key | _(from config)_ |
| `OPENAI_API_KEY` | OpenAI API key | _(from config)_ |

**Example:**
```bash
# Run with debug logging
VECTORA_LOG_LEVEL=debug vectora start

# Index a specific project
VECTORA_TRUST_FOLDER=/home/user/my-project vectora embed

# Test mode (minimal output)
VECTORA_MODE=test vectora embed --verbose
```

---

### Running the Core Directly (Development)

For development, you can run the Core without the system tray:

```bash
# Start the ACP server in foreground (shows all logs)
vectora acp --debug

# In another terminal:
vectora status
vectora ask "Test query"
```

---

## Troubleshooting

### "Vectora Core is not running"

```bash
# Start it
vectora start

# Verify
vectora status
```

### "API key not configured"

```bash
# Set your key
vectora config --gemini YOUR_API_KEY

# Verify
vectora config --list
```

### "No embeddings found"

```bash
# Index your project
vectora embed

# Check progress
vectora embed --verbose
```

### Connection errors from IDE extensions

```bash
# Restart the service
vectora restart

# Check logs
cat ~/.vectora/logs/core.log

# Run with debug output
VECTORA_LOG_LEVEL=debug vectora start
```

---

## Use Cases

### Onboarding to a New Project

```bash
# 1. Navigate to project root
cd /path/to/project

# 2. Start Vectora
vectora start

# 3. Index the codebase
vectora embed --verbose

# 4. Ask questions
vectora ask "What's the architecture of this project?"
vectora ask "How are database migrations handled?"
vectora ask "Where would I add a new API endpoint?"
```

### Code Review Assistance

```bash
# Ask about specific patterns
vectora ask "Where are all the places we modify user data?"

# Understand complexity
vectora ask "Why might this codebase have performance issues?"

# Check consistency
vectora ask "How are errors handled across the codebase?"
```

### Integration with IDEs

The CLI is typically used directly by:
- **VS Code Extension** — via `vectora acp`
- **Gemini CLI Bridge** — via MCP protocol
- **Custom integrations** — via JSON-RPC

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid argument |
| 3 | Service not running |
| 4 | Configuration error |

---

## Performance Notes

**Typical Execution Times:**
- `vectora start` — ~1-2 seconds
- `vectora status` — ~200ms
- `vectora embed` (10k files) — ~45-60 seconds
- `vectora ask` (simple query) — ~2-8 seconds

**Memory Usage:**
- Idle (background) — ~50-80 MB
- During indexing — ~200-300 MB
- During query — ~150-200 MB

---

## License

MIT — Part of the [Vectora](https://github.com/Kaffyn/Vectora) ecosystem.

---

## Next Steps

- [Full Documentation](../README.md)
- [VS Code Extension Guide](../extensions/vscode/README.md)
- [Gemini CLI Bridge](../extensions/geminicli/README.md)
- [Testing Guide](../test/README.md)
