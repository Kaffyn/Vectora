# Claude Code + Vectora Integration Guide

**Status**: ✅ Integration Ready
**Date**: 2026-04-12
**Protocol**: MCP (Model Context Protocol) - JSON-RPC 2.0
**Vectora Status**: Core MCP implementation 95% complete

---

## Overview

Claude Code can use Vectora as a specialized sub-agent for local code analysis, semantic search, and RAG (Retrieval-Augmented Generation) queries.

This integration allows Claude Code to:
- **Search semantically** across your entire codebase
- **Analyze code patterns** and architecture
- **Generate tests** based on code context
- **Detect potential bugs** using pattern matching
- **Extract documentation** from code
- **Index new files** into embeddings database

### Why MCP Integration?

**MCP (Model Context Protocol)** is a standard for connecting Claude to external tools and data sources. Instead of building a custom extension, Vectora exposes its capabilities via the MCP protocol that Claude Code natively supports.

**Benefits**:
- ✅ No extension needed (Claude Code has native MCP support)
- ✅ Standard protocol (JSON-RPC 2.0 over stdio)
- ✅ Tool auto-discovery (Claude Code sees all available tools)
- ✅ Multi-workspace (run separate Vectora instance per workspace)
- ✅ Secure (read-only access within Trust Folder)

---

## Quick Start

### 1. Verify Vectora is Installed

```bash
vectora --version
```

Expected output: `Vectora 0.1.0+`

### 2. Locate Your Claude Code Settings

**Windows**:
```
C:\Users\[your-username]\.claude\settings.json
```

**macOS**:
```
~/.claude/settings.json
```

**Linux**:
```
~/.claude/settings.json
```

### 3. Add Vectora to MCP Servers

Edit `settings.json` and add the `mcpServers` section:

```json
{
  "mcpServers": {
    "vectora": {
      "command": "vectora",
      "args": ["mcp", "/absolute/path/to/your/workspace"],
      "env": {
        "VECTORA_WORKSPACE": "/absolute/path/to/your/workspace",
        "VECTORA_DEBUG": "false"
      }
    }
  }
}
```

**Replace `/absolute/path/to/your/workspace`** with your actual project root.

**Examples**:
- Windows: `"C:\\Users\\bruno\\projects\\my-app"`
- macOS/Linux: `/home/user/projects/my-app`

### 4. Restart Claude Code

Close and reopen Claude Code to load the new MCP server configuration.

### 5. Verify Connection

Type in Claude Code:

```
@vectora analyze code patterns in this project
```

If Vectora is connected, it will respond with available tools and proceed with analysis.

---

## Configuration Details

### Basic Configuration (Recommended)

```json
{
  "mcpServers": {
    "vectora": {
      "command": "vectora",
      "args": ["mcp", "/path/to/workspace"]
    }
  }
}
```

### Advanced Configuration

```json
{
  "mcpServers": {
    "vectora": {
      "command": "vectora",
      "args": ["mcp", "/path/to/workspace", "--port", "9999"],
      "env": {
        "VECTORA_WORKSPACE": "/path/to/workspace",
        "VECTORA_DEBUG": "true",
        "VECTORA_LOG_LEVEL": "debug",
        "VECTORA_TRUST_FOLDER": "/path/to/workspace",
        "VECTORA_MCP_SERVER": "true"
      },
      "disabled": false,
      "alwaysAllow": ["vectora"]
    }
  }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `command` | string | `vectora` | Path to Vectora binary |
| `args` | array | `["mcp", "/workspace"]` | Arguments to Vectora |
| `env.VECTORA_WORKSPACE` | string | Workspace path | Workspace root directory |
| `env.VECTORA_DEBUG` | boolean | `false` | Enable debug logging |
| `env.VECTORA_LOG_LEVEL` | string | `info` | Log level (debug, info, warn, error) |
| `env.VECTORA_TRUST_FOLDER` | string | Workspace path | Root of accessible files |
| `env.VECTORA_MCP_SERVER` | boolean | `true` | Enable MCP protocol |
| `disabled` | boolean | `false` | Disable this server |
| `alwaysAllow` | array | `[]` | Tools that don't require approval |

---

## Available Tools

Vectora exposes **11 specialized tools** via MCP:

### Semantic Search & Indexing
- **`embed`** — Index files into vector database
- **`search_database`** — Semantic search across indexed content
- **`web_search_and_embed`** — Research topics and add to index

### Code Analysis
- **`analyze_code_patterns`** — Detect patterns and anti-patterns
- **`knowledge_graph_analysis`** — Extract entities and relationships
- **`doc_coverage_analysis`** — Measure documentation completeness
- **`bug_pattern_detection`** — Find potential bugs and issues

### Code Generation & Refactoring
- **`test_generation`** — Generate test cases from code
- **`refactor_with_context`** — Smart refactoring using RAG context
- **`plan_mode`** — Structured problem decomposition
- **`web_fetch_and_embed`** — Fetch external docs and index them

---

## How It Works

### Message Flow

```
1. User types in Claude Code
   "Find all authentication patterns in the code"

2. Claude Code:
   - Sends to Vectora MCP server via stdio
   - Uses JSON-RPC 2.0 protocol

3. Vectora:
   - Parses the request
   - Selects appropriate tools (analyze_code_patterns, search_database)
   - Executes tools within Trust Folder
   - Returns results as JSON

4. Claude Code:
   - Formats results in conversation
   - Shows tool execution details
   - User sees structured response
```

### Tool Invocation

Claude Code automatically selects the right tools based on your request:

**Request**: "Search for all async functions"
→ **Tools**: `search_database` (semantic search)

**Request**: "Generate tests for the API handler"
→ **Tools**: `test_generation` (with code context)

**Request**: "What patterns are used in this codebase?"
→ **Tools**: `analyze_code_patterns`, `knowledge_graph_analysis`

---

## Usage Examples

### Example 1: Semantic Search

```
You: "@vectora find all authentication patterns in the code"

Vectora:
1. Uses search_database to find semantically similar code
2. Identifies: Login handlers, JWT validation, password hashing
3. Returns: File paths, code snippets, confidence scores
```

### Example 2: Code Analysis

```
You: "@vectora What race conditions might exist in our async code?"

Vectora:
1. Uses analyze_code_patterns to scan for concurrency issues
2. Searches for: goroutines, channels, mutexes, atomic operations
3. Identifies: potential deadlocks, data races
4. Returns: Issues with severity levels and recommendations
```

### Example 3: Test Generation

```
You: "@vectora Generate comprehensive tests for the QueryHandler function"

Vectora:
1. Reads the QueryHandler implementation
2. Uses test_generation tool with code context
3. Extracts: function signature, error cases, edge cases
4. Returns: Go test code ready to use
```

### Example 4: Documentation

```
You: "@vectora What's the documentation coverage of this code?"

Vectora:
1. Uses doc_coverage_analysis
2. Scans all files for comments/docstrings
3. Identifies: Missing documentation, incomplete examples
4. Returns: Coverage report with specific gaps
```

### Example 5: Refactoring

```
You: "@vectora Refactor this code to use the event pattern we have elsewhere"

Vectora:
1. Reads the code to refactor
2. Searches codebase for event pattern examples
3. Uses refactor_with_context with RAG context
4. Returns: Refactored code + explanation of changes
```

---

## Troubleshooting

### "Vectora not found" Error

**Problem**: Claude Code can't locate the Vectora binary

**Solution**:
1. Verify Vectora is installed: `vectora --version`
2. Check Vectora is in PATH: `which vectora` (macOS/Linux) or `where vectora` (Windows)
3. Use absolute path in settings.json: `"command": "C:\\Program Files\\Vectora\\bin\\vectora"`

### "Connection refused" or "Timeout"

**Problem**: Vectora MCP server isn't responding

**Solution**:
1. Check workspace path is valid: `ls /path/to/workspace`
2. Enable debug logging: `"env": {"VECTORA_DEBUG": "true"}`
3. Restart Claude Code completely (not just refresh)
4. Check Vectora logs: `vectora logs --tail 50`

### "Tools not available"

**Problem**: Claude Code sees Vectora but tools don't work

**Solution**:
1. Verify workspace permissions: `ls -la /path/to/workspace`
2. Check trust folder: Vectora can only access files within workspace
3. Ensure no firewall blocking stdio communication
4. Try simpler request first: `@vectora what tools are available`

### "Permission denied" on files

**Problem**: Vectora can't read/write files

**Solution**:
1. Vectora can only access files within Trust Folder
2. Trust Folder defaults to workspace root
3. Move project into accessible location or change Trust Folder in settings
4. Check file permissions: `chmod 644 filename`

### Vectora Connects But Commands Fail

**Problem**: Tools execute but return errors

**Possible causes and solutions**:
1. **Workspace not initialized**: `vectora index /path/to/workspace`
2. **No embeddings yet**: First request embeds files, may take time
3. **API keys missing**: If using web search: set `GEMINI_API_KEY` or `OPENAI_API_KEY`
4. **File encoding issues**: Ensure files are UTF-8 encoded

### Debug Logging

Enable detailed logging:

```json
{
  "mcpServers": {
    "vectora": {
      "command": "vectora",
      "args": ["mcp", "/path/to/workspace"],
      "env": {
        "VECTORA_DEBUG": "true",
        "VECTORA_LOG_LEVEL": "debug"
      }
    }
  }
}
```

Then check logs:
```bash
vectora logs --follow
```

---

## Security Model

### Trust Folder Protection
- All operations restricted to workspace root
- Guardian policy validates every tool call
- Files outside Trust Folder are inaccessible

### Tool Whitelist
Only these 11 tools are exposed via MCP:
- Semantic search (read-only)
- Code analysis (read-only)
- Test/doc generation (write, controlled)
- Pattern detection (read-only)

**NOT exposed**:
- Direct file system access
- API key management
- Model configuration
- System commands

### Multi-Workspace Isolation
Each workspace gets a separate Vectora process:
- Independent embeddings database
- Isolated BBolt KV store
- No data leakage between workspaces

---

## Advanced Usage

### Custom MCP Server Port

```json
{
  "mcpServers": {
    "vectora": {
      "command": "vectora",
      "args": ["mcp", "/workspace", "--port", "9999"]
    }
  }
}
```

### Multiple Workspaces

Configure separate Vectora instances per workspace:

```json
{
  "mcpServers": {
    "vectora-backend": {
      "command": "vectora",
      "args": ["mcp", "/projects/backend"]
    },
    "vectora-frontend": {
      "command": "vectora",
      "args": ["mcp", "/projects/frontend"]
    }
  }
}
```

Then use in Claude Code:
```
@vectora-backend analyze the API structure
@vectora-frontend find all React components
```

### Environment Variables

Pass configuration via environment:

```bash
export VECTORA_WORKSPACE=/path/to/project
export VECTORA_DEBUG=true
export VECTORA_LOG_LEVEL=debug
export CLAUDE_MCP_SERVERS='{"vectora":{"command":"vectora","args":["mcp","/workspace"]}}'

# Then start Claude Code
code
```

---

## Performance Expectations

### First Connection
- Initial connection: 2-5 seconds
- First embedding: 10-30 seconds (depends on codebase size)
- Subsequent queries: 2-10 seconds

### Tool Execution Times
| Tool | Time | Notes |
|------|------|-------|
| `search_database` | 2-5s | Semantic search (indexed) |
| `analyze_code_patterns` | 5-15s | Pattern detection |
| `test_generation` | 10-20s | Test code generation |
| `doc_coverage_analysis` | 5-10s | Documentation analysis |
| `bug_pattern_detection` | 5-15s | Issue detection |
| `refactor_with_context` | 15-30s | Code refactoring |
| `web_search_and_embed` | 15-30s | Web research + indexing |

### Codebase Size Impact
- **Small** (< 100 files, < 1MB): First embedding ~10s, queries 2-5s
- **Medium** (100-1K files, 1-100MB): First embedding ~30s, queries 5-10s
- **Large** (1K+ files, > 100MB): First embedding ~60s, queries 10-20s

---

## Integration with Claude Code Features

### Code Lens Integration
Hover over functions to see:
- Similar patterns in codebase
- Test coverage
- Documentation status
- Related code locations

### Inline Completions
Code suggestions based on:
- Codebase patterns (RAG)
- Similar implementations found by Vectora
- Architecture context

### Chat Context
Reference Vectora analysis in chat:
```
"Based on the architecture analysis @vectora provided,
I would suggest refactoring to..."
```

---

## Comparison: MCP vs Other Approaches

### MCP Integration (This Approach) ✅ RECOMMENDED

**Pros**:
- Native Claude Code support (no extension needed)
- Standard protocol (JSON-RPC 2.0)
- Tool auto-discovery
- Multi-workspace isolation
- Full RAG context preservation

**Cons**:
- Requires settings.json configuration
- MCP protocol knowledge helpful (not required)

### Custom Extension Approach ❌ NOT RECOMMENDED

**Pros**:
- Could customize UI heavily

**Cons**:
- Requires building separate extension
- No official VS Code extension framework for Vectora
- More maintenance burden
- Duplicates MCP functionality
- Slower to implement

### CLI Wrapper Approach ⚠️ SIMPLER BUT LIMITED

**Pros**:
- Very simple implementation
- Uses existing CLI

**Cons**:
- No streaming
- Can't compose tools
- No context preservation
- Lose RAG benefits
- Slower overall

**Conclusion**: MCP is superior and already 95% implemented.

---

## Getting Help

### Documentation
- **Tools Reference**: See `MCP_PROTOCOL_REFERENCE.md`
- **Tool Details**: See `TOOLS_API_REFERENCE.md`
- **Vectora Architecture**: See `API_ARCHITECTURE.md`

### Debug Commands

```bash
# Check Vectora status
vectora status

# Verify workspace
vectora inspect /path/to/workspace

# View recent logs
vectora logs --tail 100

# Test MCP connection
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | vectora mcp /path/to/workspace

# Rebuild embeddings
vectora index /path/to/workspace --force
```

### Contact & Support

- **Issues**: Check Vectora GitHub issues
- **Documentation**: See `/TOOLS_DOCUMENTATION.md` and `/TOOLS_API_REFERENCE.md`
- **Configuration**: Edit `~/.claude/settings.json` directly

---

## Summary

| Step | Action | Time |
|------|--------|------|
| 1 | Install Vectora | < 1 min |
| 2 | Add to settings.json | 2 mins |
| 3 | Restart Claude Code | 1 min |
| 4 | First query (embedding) | 10-30s |
| 5 | Subsequent queries | 2-10s |
| **Total Setup** | **~5-10 minutes** | |

Once configured, Vectora is invisible — Claude Code automatically routes requests to the right tools and presents results seamlessly.

---

## Next Steps

1. **Add Vectora to your Claude Code settings** (see Quick Start above)
2. **Restart Claude Code**
3. **Try first query**: `@vectora analyze the architecture of this project`
4. **Refer to Tool Reference** (`MCP_PROTOCOL_REFERENCE.md`) for specific use cases

**Your Vectora integration is ready! 🚀**
