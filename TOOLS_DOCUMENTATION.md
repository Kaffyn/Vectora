# Vectora Tools Documentation

## Overview

Vectora has **10 primary tools** organized into 5 categories. Each tool has a specific purpose and should be called only when needed.

**Critical Principle**: The model learns from instructions to decide WHEN to use each tool. Not all tools should be called for every query.

---

## Tool Inventory

### Category 1: File System Operations (4 tools)

#### 1. **read_file** - Read File Content
```
Name: read_file
Purpose: Read complete content of a file within the Trust Folder
When to use:
  - User asks "read X.go"
  - Need to understand specific file implementation
  - Already know which file to read
When NOT to use:
  - Don't know which files are in the project (use find_files first)
  - Searching for specific patterns (use grep_search instead)
Parameters:
  - path: Path to file (relative to Trust Folder)
Time: ~1-2 seconds
Example:
  Query: "Ler o arquivo main.go"
  → read_file(path: "main.go")
```

#### 2. **write_file** - Create/Overwrite File
```
Name: write_file
Purpose: Create new file or completely overwrite existing file
When to use:
  - Generate code from scratch
  - Create configuration files
  - Export results to file
When NOT to use:
  - Modifying existing code (use edit instead)
Parameters:
  - path: File path to create
  - content: Complete file content
Time: ~1-2 seconds
Example:
  Query: "Criar arquivo config.json com padrão"
  → write_file(path: "config.json", content: "...")
```

#### 3. **read_folder** - List Directory Contents
```
Name: read_folder
Purpose: List all files and subdirectories in a path
When to use:
  - Explore project structure
  - See what files exist at top level
  - Understand directory organization
When NOT to use:
  - Looking for specific files (use find_files with pattern)
Parameters:
  - path: Directory path to list
Time: ~1-2 seconds
Example:
  Query: "Mostre os arquivos em src/"
  → read_folder(path: "src/")
```

#### 4. **find_files** - Search Files by Pattern
```
Name: find_files
Purpose: Find files matching glob patterns recursively
When to use:
  - Find all .go files: find_files("**/*.go")
  - Find test files: find_files("**/*_test.go")
  - Find by name pattern: find_files("**/config*")
When NOT to use:
  - Searching content inside files (use grep_search)
Parameters:
  - pattern: Glob pattern (**, *, ?)
Time: ~2-5 seconds (depends on project size)
Example:
  Query: "Encontre todos os arquivos TypeScript"
  → find_files(pattern: "**/*.ts")
```

---

### Category 2: Code Search & Pattern Matching (1 tool)

#### 5. **grep_search** - Search Content with Regex
```
Name: grep_search
Purpose: Search file contents using regular expressions
When to use:
  - Find function definitions: grep_search("func.*Query")
  - Find imports: grep_search("^import")
  - Find error handling: grep_search("error.*:=")
  - Find specific strings: grep_search("TODO")
When NOT to use:
  - Just want to read a file (use read_file)
  - Don't know which files to search (use find_files first)
Parameters:
  - pattern: Regex pattern to match
  - path: Path to search (default: current folder)
Time: ~2-5 seconds
Example:
  Query: "Encontre todas as funções async"
  → grep_search(pattern: "async.*function", path: ".")
```

---

### Category 3: Code Editing (1 tool)

#### 6. **edit** - Edit Code Semantically
```
Name: edit
Purpose: Make semantic edits to existing code (not find-replace)
When to use:
  - "Adicione logging ao início de ProcessQuery"
  - "Adicione tratamento de erro aqui"
  - "Refatore esta função para usar context"
When NOT to use:
  - Simple find-replace (use your prompt for manual specification)
  - Creating new files (use write_file)
  - Just reading code (use read_file)
Parameters:
  - path: File to edit
  - instruction: What change to make (semantic description)
  - content: Code snippet or context for the change
Time: ~5-10 seconds
Example:
  Query: "Adicione logging ao ProcessQuery"
  → edit(path: "engine.go",
          instruction: "No início de ProcessQuery, log the input",
          content: "log.Printf(\"Processing: %s\", query)")
```

---

### Category 4: Execution (1 tool)

#### 7. **run_shell_command** - Execute Shell Commands
```
Name: run_shell_command
Purpose: Execute shell commands (only within Trust Folder)
When to use:
  - Run tests: "go test ./..."
  - Compile code: "go build"
  - Run scripts: "bash build.sh"
  - Check versions: "go version"
When NOT to use:
  - Need to read file (use read_file)
  - Need to find files (use find_files)
Parameters:
  - command: Shell command to execute
Time: ~5-30 seconds (depends on command)
Example:
  Query: "Execute all tests"
  → run_shell_command(command: "go test ./...")
```

---

### Category 5: Web Integration (2 tools)

#### 8. **google_search** - Web Search
```
Name: google_search
Purpose: Search the web via DuckDuckGo
When to use:
  - User asks "pesquise sobre X"
  - Need external documentation/articles
  - Looking for best practices beyond local codebase
When NOT to use:
  - Searching local files (use grep_search)
  - Questions answerable from knowledge
Parameters:
  - query: Search query string
Time: ~5-10 seconds
Example:
  Query: "Pesquise sobre Go generics 2026"
  → google_search(query: "Go generics 2026 best practices")
```

#### 9. **web_fetch** - Download & Process URLs
```
Name: web_fetch
Purpose: Fetch content from URL and convert HTML to readable text
When to use:
  - After google_search, fetch promising URLs
  - Get API documentation: "web_fetch https://api.example.com/docs"
  - Read blog posts, specs, tutorials
When NOT to use:
  - Don't know the URL (use google_search first)
Parameters:
  - url: URL to fetch
Time: ~3-8 seconds (depends on page size)
Example:
  Query: "Leia a documentação API de X"
  → web_fetch(url: "https://example.com/api-docs")
```

---

### Category 6: Persistence (1 tool)

#### 10. **save_memory** - Long-term Memory Storage
```
Name: save_memory
Purpose: Store key-value pairs in persistent BBolt database
When to use:
  - Save important patterns discovered
  - Store learnings for future queries
  - Persist findings: "save_memory key=patterns value=JSON"
When NOT to use:
  - Just during current query (no need to save)
Parameters:
  - key: Memory key (string)
  - value: Memory value (string, typically JSON)
Time: ~1-2 seconds
Example:
  Query: "Salve os padrões encontrados"
  → save_memory(key: "event_patterns",
                value: "{\"patterns\": [...], \"date\": \"...\"}")
```

---

## Tool Usage Decision Tree

```
User Query Arrives
    ↓
[Is it TRIVIAL?] (greeting, thanks, simple knowledge)
    ├─ YES → RESPOND ONLY WITH HISTORY (no tools)
    │        Time: 2-5 seconds
    │
    └─ NO → [What is the GOAL?]
            ├─ "Read code" → read_file, grep_search, find_files
            │                Time: 5-15 seconds
            │
            ├─ "Write/modify code" → read_file, edit, write_file
            │                        Time: 10-20 seconds
            │
            ├─ "Run tests/build" → run_shell_command
            │                      Time: 5-30 seconds
            │
            ├─ "Research external info" → google_search, web_fetch
            │                             Time: 15-30 seconds
            │
            └─ "Deep analysis" → grep_search, find_files, read_file
                                 (Maybe embedding+RAG if configured)
                                 Time: 30-60 seconds
```

---

## Performance Guidelines

| Tool | Typical Time | When Slow |
|------|-------------|-----------|
| read_file | 1-2s | Large files (>100KB) |
| write_file | 1-2s | Network storage |
| read_folder | 1-2s | Many files (1000+) |
| find_files | 2-5s | Large projects, complex patterns |
| grep_search | 2-5s | Large codebase, complex regex |
| edit | 5-10s | Understanding context takes time |
| run_shell_command | 5-30s | Depends on command complexity |
| google_search | 5-10s | Network latency |
| web_fetch | 3-8s | Page size, network latency |
| save_memory | 1-2s | BBolt write time |

---

## Critical Rules

### ✅ DO

1. **Check if tool is needed before calling it**
   - Read instructions before grep_search
   - Find files before reading them
   - Ask user before modifying code

2. **Use most specific tool for the task**
   - read_file for entire file content
   - grep_search for patterns
   - find_files for discovering files

3. **Chain tools efficiently**
   - find_files → read_file (find first, then read)
   - google_search → web_fetch (search first, then fetch)
   - read_file → edit (read context before editing)

4. **Respect Trust Folder**
   - All file operations stay within Trust Folder
   - Cannot read `~/.ssh`, `/etc`, Windows system files

### ❌ DON'T

1. **Don't use embedding/RAG for trivial queries**
   - No embedding for "oi", greetings, simple questions
   - No RAG for content already available locally

2. **Don't call multiple tools needlessly**
   - Don't read_file if find_files results would show what you need
   - Don't grep_search if simple read_file would work

3. **Don't modify code without confirming**
   - Show what will be changed first
   - Ask user permission before edit/write

4. **Don't access files outside Trust Folder**
   - No .env files, .ssh, system directories
   - Only files within configured Trust Folder

---

## Integration with Instructions

The file `/core/instructions/instructions.md` contains the official training for when to use each tool. It includes:

1. **Section 6: DECISÕES DE TOOL CALLING** - Explains each tool
2. **Section 6.3: REGRA CRÍTICA** - When NOT to use embedding/RAG
3. **Section 6.4: FLUXO DE DECISÃO** - Decision tree for tool selection

These instructions are loaded into the system prompt at runtime and guide the model's tool selection behavior.

---

## Examples by Scenario

### Scenario 1: User asks "oi"
```
Input: "oi"
Decision: TRIVIAL (greeting)
Tools used: NONE
Response: "Oi! Tudo bem?"
Time: 2-5s
```

### Scenario 2: User asks "leia main.go"
```
Input: "leia main.go"
Decision: SIMPLE CODE READING
Tools used: read_file("main.go")
Response: [file content with explanation]
Time: 5-10s
```

### Scenario 3: User asks "encontre async functions"
```
Input: "encontre async functions"
Decision: CODE SEARCH
Tools used:
  1. grep_search("async.*function")
  2. find_files("**/*.ts") [if needed]
Response: [matching functions with file locations]
Time: 10-15s
```

### Scenario 4: User asks "refatore este código"
```
Input: "refatore este código"
Decision: CODE MODIFICATION
Tools used:
  1. read_file("relevant_file.go")
  2. grep_search to understand context
  3. edit(...) to apply changes
Response: [explanation of changes]
Time: 15-25s
```

### Scenario 5: User asks "pesquise sobre Go generics"
```
Input: "pesquise sobre Go generics"
Decision: WEB RESEARCH
Tools used:
  1. google_search("Go generics 2026")
  2. web_fetch(URL from results)
Response: [summary with sources]
Time: 20-30s
```

---

## Troubleshooting

### Problem: Tools take too long
- Check if trivial query is being analyzed unnecessarily
- Verify tools are only called when needed
- Check internet connectivity for web tools

### Problem: Tool returns empty/error
- Verify file exists within Trust Folder
- Check glob patterns are valid
- Ensure regex patterns are correct
- Check command syntax is valid

### Problem: Wrong files found
- Check find_files pattern is specific enough
- Use grep_search to filter results
- Try more specific file extensions

---

## Version & Updates

- **Created**: 2026-04-12
- **Last Updated**: 2026-04-12
- **Vectora Version**: 0.1.0
- **Tool Count**: 10
- **Categories**: 6

---

## See Also

- `/core/instructions/instructions.md` - Model training instructions
- `/core/instructions/tool_examples.json` - Structured tool examples
- `/core/tools/` - Tool implementations
- `/core/llm/prompt_factory.go` - Where instructions are loaded
