# Vectora — Gemini CLI Bridge (MCP)

> **Sub-Agent Mode Only** — Connects Gemini CLI to Vectora's RAG and agentic tools via the Model Context Protocol (MCP).

[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/Kaffyn/Vectora)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## Overview

`vectora-geminicli` is a bridge that allows the **Gemini CLI** (from Google Cloud SDK) to communicate with the **Vectora Core**. By exposing Vectora's capabilities as MCP tools, it enables the Gemini CLI to:

1. **Search your codebase** semantically (RAG).
2. **Read and write files** safely within your Trust Folder.
3. **Execute verified commands** to test or build code.

In this architecture, the Gemini CLI acts as the **Main Agent**, while Vectora acts as a specialized **Sub-Agent** (Specialist) invoked whenever deep codebase context or tool execution is required.

---

## Installation

### 1. Build from Source

Navigate to the extension directory and build the TypeScript source:

```bash
cd extensions/geminicli
npm install
npm run build
```

### 2. Verify Binary

Make sure you have the **Vectora Core** binary in your `PATH`. You can check with:

```bash
vectora --version
```

---

## Configuration

The Gemini CLI uses a JSON configuration to discover MCP servers. You can generate this configuration automatically using the `config` command.

### Automatic Configuration

Run the following command to see the JSON snippet required for your setup:

```bash
node dist/index.js config
```

### Setup Steps

1. Open (or create) your Gemini CLI settings file:

   - **Linux/macOS**: `~/.config/gemini/settings.json`
   - **Windows**: `%APPDATA%\google-cloud-sdk\gemini\settings.json` (approximate path)

2. Add the `mcpServers` entry:

```json
{
  "mcpServers": {
    "vectora": {
      "command": "node",
      "args": ["/path/to/Vectora/extensions/geminicli/dist/index.js", "call-tool"]
    }
  }
}
```

_Note: Ensure the path to `index.js` is absolute._

---

## Usage

### Interactive REPL (Testing)

You can launch a standalone REPL to test if the MCP connection to the Core is working correctly:

```bash
# Set your workspace before starting
export VECTORA_WORKSPACE=/path/to/your/project
node dist/index.js
```

Inside the REPL, try:

- `/tools` — List all tools available from Vectora Core.
- `/new` — Start a new session in the current folder.
- `How does the authentication work in this project?` — Ask a research question.

### Environment Variables

| Variable            | Description                                      | Default         |
| ------------------- | ------------------------------------------------ | --------------- |
| `VECTORA_CORE_PATH` | Path to the `vectora` binary                     | `vectora`       |
| `VECTORA_WORKSPACE` | Folder to index and search                       | `process.cwd()` |
| `DEBUG`             | Set to `mcp:*` to see detailed protocol messages | _(off)_         |

---

## Development

- `npm run watch`: Recompile on changes.
- `npm run start`: Run the bridge directly.

---

## License

MIT — Part of the [Vectora](https://github.com/Kaffyn/Vectora) ecosystem.
