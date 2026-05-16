# Vectora MCP Server Setup Guide

## Overview

Vectora is a complete MCP (Model Context Protocol) server built with FastMCP. This guide explains how to connect it to Claude Desktop or Claude Code.

## Architecture

```
┌─────────────────────┐
│  Claude Desktop     │
│  Claude Code        │
└──────────┬──────────┘
           │ (JSON-RPC via stdio)
           │
┌──────────▼──────────┐
│   Vectora MCP       │
│   (mcp_server.py)   │
│   - FastMCP         │
│   - 11 Tools        │
│   - 3 Resources     │
└─────────────────────┘
```

## Setup Instructions

### Step 1: Locate Your Vectora Project

First, find the **absolute path** to your Vectora installation:

```bash
# On macOS/Linux:
which vectora
# or
realpath ~/path/to/vectora

# On Windows:
where vectora
```

### Step 2: Update `claude_desktop_config.json`

Replace `/path/to/vectora` with your actual path:

**macOS/Linux:**

```json
{
  "mcpServers": {
    "Vectora-MCP": {
      "command": "uv",
      "args": [
        "run",
        "--project",
        "/Users/bruno/projects/vectora",
        "vectora-mcp"
      ]
    }
  }
}
```

**Windows (example):**

```json
{
  "mcpServers": {
    "Vectora-MCP": {
      "command": "uv",
      "args": [
        "run",
        "--project",
        "C:\\Users\\bruno\\Desktop\\vectora",
        "vectora-mcp"
      ]
    }
  }
}
```

### Step 3: Copy to Claude Desktop Config

Copy the `claude_desktop_config.json` file to:

**macOS:**

```
~/Library/Application Support/Claude/claude.json
```

**Windows:**

```
%APPDATA%\Claude\claude.json
```

**Linux:**

```
~/.config/Claude/claude.json
```

### Step 4: Restart Claude Desktop

Close and reopen Claude Desktop. You should see "Vectora-MCP" appear in the MCP servers list.

## Troubleshooting

### Error: "Command not found"

- Ensure the path to `vectora` project is correct (use absolute paths)
- Verify `uv` is installed: `uv --version`
- Check that `uv` is in PATH or provide full path to it

### Error: "Connection closed"

- Check logs: `tail -f logs/mcp.log`
- Ensure `vectora-mcp` entry point exists in `pyproject.toml`
- Verify `mcp_server.py` has `mcp.run()` call

### MCP Server Not Appearing

- Restart Claude Desktop completely
- Check that JSON is valid in `claude.json`
- Verify file is in correct location for your OS

## Available Tools

Vectora exposes these tools via MCP:

- `web_search` - Search the internet
- `fetch_url` - Fetch and extract content from URLs
- `vector_search` - Search vector database
- `embedding` - Enqueue documents for embedding
- `file_read` - Read file contents
- `file_edit` - Edit file contents
- `grep` - Search patterns in files
- `list_dir` - List directory contents
- `terminal` - Execute shell commands
- `save_memory` - Save persistent memories
- `get_memory` - Retrieve memories

## Available Resources

- `vectora://thread/{thread_id}/context` - Thread context
- `vectora://thread/{thread_id}/history` - Thread history
- `vectora://status` - Server status

## Next Steps

Once connected:

1. Ask Claude to use Vectora tools
2. Vectora executes them via its graph
3. Results are returned to Claude
4. Claude processes and responds

## Architecture Details

### Server (mcp_server.py)

- Uses FastMCP for protocol handling
- stdio JSON-RPC transport
- Automatic tool discovery from `tools.py`
- Async/await support

### Tools (tools.py)

- LangChain-based tool definitions
- Pydantic validation
- Error handling with actionable messages
- Supports both local and MCP execution

### Transport

- **Protocol**: Model Context Protocol (MCP)
- **Format**: JSON-RPC 2.0
- **Transport**: stdio (standard input/output)
- **Language**: Python 3.13+

## References

- [MCP Specification](https://modelcontextprotocol.io/)
- [FastMCP Documentation](https://github.com/modelcontextprotocol/python-sdk)
- [Vectora GitHub](https://github.com/bssnem/vectora)
