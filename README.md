# Vectora

**Vectora** is an open-source AI assistant (Apache 2.0) built for developers — local-first, self-hosted, and designed to run as a powerful sub-agent inside any MCP-compatible orchestrator (Claude Code, Claude Desktop, Paperclip, VS Code extensions).

At its core, Vectora solves the **knowledge gap problem**: LLMs don't know your codebase, your docs, or the latest versions of your stack. Vectora bridges that gap with RAG (Retrieval-Augmented Generation) — ingest your docs once, and every AI interaction becomes contextually aware.

---

## Why Vectora?

- **RAG-native**: Every conversation is backed by a local vector store. Ingest docs, code, wikis — and your AI actually knows them.
- **14 tools built-in**: File operations, terminal, web search, vector search, memory, MCP bridging — ready to use.
- **Sub-agent architecture**: Designed to run as an MCP server. Claude Code delegates complex tasks to Vectora; Vectora reasons and responds.
- **Persistent memory**: Cross-session memory stored in SQLite. Vectora remembers your preferences, project context, and decisions.
- **Zero infra**: SQLite + LanceDB. No Docker required for local use. No Postgres, no Redis, no cloud required.
- **Multi-LLM**: Google Gemini (free tier), OpenAI, Anthropic, or Ollama (fully local).

---

## Prerequisites

### Cohere — Required

Vectora uses [Cohere](https://cohere.com/) for embeddings (`embed-multilingual-v3.0`) and reranking (`rerank-multilingual-v3.0`). It offers a **generous free tier** with first-class LangChain integration.

Get your key: https://dashboard.cohere.com/api-keys

### LLM Provider — Choose One

| Provider                         | Free Tier | Get Key                                                       |
| -------------------------------- | --------- | ------------------------------------------------------------- |
| **Google Gemini** ✅ Recommended | Yes       | [aistudio.google.com](https://aistudio.google.com/app/apikey) |
| Ollama (local)                   | No cost   | [ollama.ai](https://ollama.ai)                                |
| OpenAI                           | Paid      | [platform.openai.com](https://platform.openai.com/api-keys)   |
| Anthropic                        | Paid      | [console.anthropic.com](https://console.anthropic.com/)       |

---

## Installation

### Option 1: UV (Recommended)

```bash
# Install globally
uv tool install vectora

# First-time setup (interactive wizard)
vectora setup

# Start chatting
vectora chat
```

### Option 2: From Source

```bash
git clone https://github.com/Kaffyn/vectora.git
cd vectora

# Install with all dependencies
uv sync

# Configure your keys
cp .env.example .env
# Edit .env with your GOOGLE_API_KEY and COHERE_API_KEY

# Run
uv run vectora chat
```

### Option 3: Docker

```bash
# Copy and configure environment
cp .env.example .env
# Edit .env with your API keys

# Run the chat interface
docker compose run --rm vectora

# Or run as MCP server (multi-agent mode)
MCP_TRANSPORT=sse docker compose up -d
```

---

## Running Modes

### Chat Mode (Interactive TUI)

The primary interface — a terminal dashboard built with Rich.

```bash
vectora chat
```

Features: multi-turn conversation, session history, live tool feedback (colored panels), debug mode toggle, model switching.

### MCP Server — Local (stdio)

Run Vectora as an MCP sub-agent for Claude Code or Claude Desktop. Single client, local process.

```bash
vectora mcp-server
```

### MCP Server — Remote (SSE, Multi-Agent)

Run Vectora as a shared hub for multiple Paperclip agents or orchestrators connecting simultaneously.

```bash
MCP_TRANSPORT=sse MCP_PORT=8000 vectora mcp-server
```

Each client passes its own `thread_id` — sessions are fully isolated.

### Setup Wizard

Interactive configuration to set up API keys, choose LLM provider, and test connectivity.

```bash
vectora setup
```

---

## Connecting to Claude Code / Claude Desktop

Add Vectora to your `.mcp.json` (in your project root) so Claude Code uses it as a sub-agent:

```json
{
  "mcpServers": {
    "Vectora-MCP": {
      "command": "uv",
      "args": ["run", "--project", "/absolute/path/to/vectora", "vectora-mcp"]
    }
  }
}
```

For a globally installed Vectora:

```json
{
  "mcpServers": {
    "Vectora-MCP": {
      "command": "vectora-mcp"
    }
  }
}
```

For Docker (SSE mode, multiple agents):

```json
{
  "mcpServers": {
    "Vectora-MCP": {
      "url": "http://localhost:8000/sse"
    }
  }
}
```

---

## Chat Commands

| Command         | Description                          |
| --------------- | ------------------------------------ |
| `/help`         | Show quick help                      |
| `/list`         | Show all commands                    |
| `/tools`        | List available tools                 |
| `/model`        | List or switch models                |
| `/debug`        | Toggle debug mode (shows tool calls) |
| `/new`          | Start a new session                  |
| `/sessions`     | List all sessions                    |
| `/session <id>` | Switch to a specific session         |
| `/quit`         | Exit                                 |

**Input shortcuts:** `Enter` sends, `Alt+Enter` or `Shift+Enter` adds a line break.

---

## Tools Reference

Vectora exposes 14 tools to the LLM and to MCP clients:

| Category     | Tools                                                      |
| ------------ | ---------------------------------------------------------- |
| **Web**      | `web_search`, `fetch_url`                                  |
| **RAG**      | `vector_search`, `embedding`, `ingest_docs`                |
| **Files**    | `file_read`, `file_edit`, `file_write`, `grep`, `list_dir` |
| **Terminal** | `terminal`                                                 |
| **Memory**   | `save_memory`, `get_memory`, `delete_memory`               |
| **MCP**      | `call_mcp_tool`                                            |

---

## Data & Persistence

All data is stored locally in `~/.vectora/`:

```
~/.vectora/
├── .env                    # Your API keys
├── chat_config.json        # Persistent chat settings
├── data/
│   ├── vectora.db          # Sessions, memories, checkpoints (SQLite)
│   ├── embedding_queue.db  # Async embedding queue (SQLite)
│   └── lancedb/            # Vector store for RAG
├── logs/
│   ├── vectora.log         # Structured JSON logs
│   └── mcp.log             # MCP server logs
└── keys/                   # Encrypted API keys (optional)
```

---

## Tech Stack

| Layer            | Technology                                                                                             |
| ---------------- | ------------------------------------------------------------------------------------------------------ |
| Language         | Python 3.13+ managed by [uv](https://github.com/astral-sh/uv)                                          |
| Agent Framework  | [LangChain](https://langchain.com/) + [LangGraph](https://langchain-ai.github.io/langgraph/)           |
| Vector Store     | [LanceDB](https://lancedb.github.io/lancedb/) — file-based, zero-config                                |
| Embeddings       | [Cohere](https://cohere.com/) — `embed-multilingual-v3.0` + `rerank-multilingual-v3.0`                 |
| Persistence      | SQLite via `aiosqlite` + LangGraph Checkpointer                                                        |
| Context Protocol | [MCP](https://modelcontextprotocol.io/) via [FastMCP](https://github.com/jlowin/fastmcp)               |
| Terminal UI      | [Rich](https://rich.readthedocs.io/) + [prompt-toolkit](https://python-prompt-toolkit.readthedocs.io/) |
| Observability    | [LangSmith](https://smith.langchain.com/) (optional)                                                   |

---

## Configuration

All configuration goes in `~/.vectora/.env` or a project-local `.env`:

```env
# LLM Provider
LLM_PROVIDER=google-genai
GOOGLE_API_KEY=your_key_here

# Required for RAG
COHERE_API_KEY=your_key_here

# Optional: Web Search
TAVILY_API_KEY=your_key_here

# Optional: Tracing
LANGSMITH_TRACING=false
LANGSMITH_API_KEY=your_key_here
LANGSMITH_PROJECT=vectora

# Optional: Logging
LOG_LEVEL=INFO
```

---

## License

Apache 2.0. See [LICENSE](./LICENSE).
