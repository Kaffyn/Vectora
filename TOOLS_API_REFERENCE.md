# Vectora Tools API Reference

Complete technical reference for all 10 tools with signatures, parameters, and return types.

---

## 1. read_file

**File**: `/core/tools/read_file.go`

```go
type ReadFileTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
}

// Read file content
// Error: File outside Trust Folder, permissions denied
// Success: File contents as string
```

**Interface**:
```
Name: read_file
Description: Reads the content of a file within the Trust Folder
Schema: {"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}
```

**Parameters**:
```json
{
  "path": "main.go"  // Relative path within Trust Folder
}
```

**Returns**:
```
✓ Output: File contents (string)
✓ IsError: false
✗ IsError: true, Error message if file not found or outside Trust Folder
```

**Example**:
```
read_file(path: "src/main.go")
read_file(path: "config/settings.json")
```

---

## 2. write_file

**File**: `/core/tools/write_file.go`

```go
type WriteFileTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
}

// Create or overwrite file
// Error: Outside Trust Folder, no write permission
// Success: File created/overwritten
```

**Interface**:
```
Name: write_file
Description: Creates or overwrites a file within the Trust Folder
Schema: {"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"}},"required":["path","content"]}
```

**Parameters**:
```json
{
  "path": "output.txt",
  "content": "file contents here..."
}
```

**Returns**:
```
✓ Output: "File written successfully"
✗ IsError: true if outside Trust Folder or write fails
```

**Example**:
```
write_file(path: "config.json", content: "{\"version\": \"1.0\"}")
write_file(path: "README.md", content: "# My Project\n...")
```

---

## 3. read_folder

**File**: `/core/tools/read_folder.go`

```go
type ReadFolderTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
}

// List directory contents recursively
// Error: Path outside Trust Folder
// Success: JSON array of files/directories
```

**Interface**:
```
Name: read_folder
Description: Lists files and directories within a base directory
Schema: {"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}
```

**Parameters**:
```json
{
  "path": "src/"  // Relative path to list
}
```

**Returns**:
```
✓ Output: JSON array with file/directory info
   [
     {"name": "main.go", "type": "file", "size": 1234},
     {"name": "subdirectory", "type": "directory"}
   ]
✗ IsError: true if path not found or outside Trust Folder
```

**Example**:
```
read_folder(path: ".")
read_folder(path: "src/")
read_folder(path: "src/components/")
```

---

## 4. find_files

**File**: `/core/tools/grep_search.go` (FindFilesTool)

```go
type FindFilesTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
}

// Find files matching glob pattern
// Error: Invalid pattern, outside Trust Folder
// Success: List of matching file paths
```

**Interface**:
```
Name: find_files
Description: Searches for files matching a pattern recursively
Schema: {"type":"object","properties":{"pattern":{"type":"string"}},"required":["pattern"]}
```

**Parameters**:
```json
{
  "pattern": "**/*.go"  // Glob pattern (**, *, ?)
}
```

**Returns**:
```
✓ Output: List of matching file paths (one per line)
   src/main.go
   src/handlers/api.go
   test/main_test.go
✗ IsError: true if invalid pattern
```

**Supported Patterns**:
- `*.go` - All .go files in current directory
- `**/*.ts` - All .ts files recursively
- `src/**/test*.go` - All .go files under src/ starting with "test"
- `**/main.go` - All main.go files at any depth

**Example**:
```
find_files(pattern: "**/*.go")
find_files(pattern: "**/*.ts")
find_files(pattern: "**/*_test.go")
find_files(pattern: "src/**/*.tsx")
```

---

## 5. grep_search

**File**: `/core/tools/grep_search.go` (GrepSearchTool)

```go
type GrepSearchTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
}

// Search file contents using regex
// Error: Invalid regex, path not found
// Success: Matching lines with file:line:content
```

**Interface**:
```
Name: grep_search
Description: Searches file contents using regex
Schema: {"type":"object","properties":{"pattern":{"type":"string"},"path":{"type":"string"}},"required":["pattern","path"]}
```

**Parameters**:
```json
{
  "pattern": "func.*Query",  // Regex pattern
  "path": "."                // Search path (default: current)
}
```

**Returns**:
```
✓ Output: Matching lines with context
   src/engine.go:42: func (e *Engine) Query(ctx context.Context) error {
   src/api/handler.go:15: func handleQuery(w http.ResponseWriter, r *http.Request) {
✗ IsError: true if invalid regex or path error
```

**Regex Examples**:
- `func.*Name` - Function definitions with Name
- `^import` - Import statements
- `error\s*:=` - Error assignments
- `TODO|FIXME` - Comments for later
- `const\s+\w+` - Constant definitions

**Example**:
```
grep_search(pattern: "func.*Query", path: ".")
grep_search(pattern: "error.*:=", path: "src/")
grep_search(pattern: "TODO", path: ".")
```

---

## 6. edit

**File**: `/core/tools/edit.go`

```go
type EditTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
}

// Semantic edit to existing file
// Error: File not found, invalid instruction
// Success: File modified
```

**Interface**:
```
Name: edit
Description: Replaces a specific block of text in a file
Schema: {"type":"object","properties":{"path":{"type":"string"},"instruction":{"type":"string"},"content":{"type":"string"}},"required":["path","instruction","content"]}
```

**Parameters**:
```json
{
  "path": "main.go",
  "instruction": "Add error handling at the beginning of ProcessQuery function",
  "content": "if query == \"\" {\n    return nil, errors.New(\"empty query\")\n}"
}
```

**Returns**:
```
✓ Output: "File edited successfully"
✗ IsError: true if file not found or instruction ambiguous
```

**How It Works**:
- NOT a simple find-replace
- Semantic understanding of where to insert
- Preserves code formatting
- Works with instruction descriptions

**Example**:
```
edit(
  path: "handlers.go",
  instruction: "Add logging at the beginning of ServeHTTP",
  content: "log.Printf(\"Request: %s %s\", r.Method, r.Path)"
)

edit(
  path: "db.go",
  instruction: "Add null check for context",
  content: "if ctx == nil {\n    return nil, ErrNilContext\n}"
)
```

---

## 7. run_shell_command

**File**: `/core/tools/terminal_run.go`

```go
type ShellTool struct {
    TrustFolder string
}

// Execute shell command (within Trust Folder)
// Error: Command fails, invalid command
// Success: Command output
```

**Interface**:
```
Name: run_shell_command
Description: Executes a shell command on the host system
Schema: {"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}
```

**Parameters**:
```json
{
  "command": "go test ./..."  // Shell command
}
```

**Returns**:
```
✓ Output: Command stdout + stderr
   ok  github.com/Kaffyn/Vectora/core/engine 0.523s
✗ IsError: true if command fails (exit code != 0)
   Output: Error message and stderr
```

**Examples**:
- `go test ./...` - Run all tests
- `go build` - Build project
- `ls -la` - List files
- `git status` - Git status
- `npm run build` - Build npm project

**Example**:
```
run_shell_command(command: "go test ./...")
run_shell_command(command: "go build -o bin/vectora")
run_shell_command(command: "npm run lint")
```

---

## 8. google_search

**File**: `/core/tools/web.go` (GoogleSearchTool)

```go
type GoogleSearchTool struct{}

// Search web using DuckDuckGo
// Error: Network failure
// Success: Search results with links
```

**Interface**:
```
Name: google_search
Description: Searches the web via DuckDuckGo
Schema: {"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}
```

**Parameters**:
```json
{
  "query": "Go generics 2026"  // Search query
}
```

**Returns**:
```
✓ Output: Search results as HTML snippet
   [1] Go 1.23 Generics Guide (example.com/post/generics)
   [2] Advanced Go Patterns (blog.example.com/go-patterns)
   [3] ...
✗ IsError: true if network error or query fails
```

**Example**:
```
google_search(query: "Go concurrency patterns 2026")
google_search(query: "REST API best practices")
google_search(query: "Docker performance optimization")
```

---

## 9. web_fetch

**File**: `/core/tools/web.go` (WebFetchTool)

```go
type WebFetchTool struct{}

// Fetch and process URL content
// Error: URL unreachable, invalid URL
// Success: HTML converted to readable text
```

**Interface**:
```
Name: web_fetch
Description: Fetches the content of a URL
Schema: {"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}
```

**Parameters**:
```json
{
  "url": "https://example.com/api-docs"  // URL to fetch
}
```

**Returns**:
```
✓ Output: Processed content (HTML → Text)
   # API Documentation
   ## Getting Started
   ...
✗ IsError: true if URL unreachable or fetch fails
```

**Processing**:
- Fetches HTTP/HTTPS URLs
- Converts HTML to readable markdown/text
- Removes unnecessary markup
- Preserves important formatting

**Example**:
```
web_fetch(url: "https://go.dev/doc/effective_go")
web_fetch(url: "https://api.example.com/documentation")
```

---

## 10. save_memory

**File**: `/core/tools/save_memory.go`

```go
type SaveMemoryTool struct {
    TrustFolder string
    Guardian    *policies.Guardian
    KV          db.KVStore
}

// Save to persistent memory (BBolt)
// Error: Invalid key/value, DB error
// Success: Memory saved
```

**Interface**:
```
Name: save_memory
Description: Stores a key-value pair in persistent memory
Schema: {"type":"object","properties":{"key":{"type":"string"},"value":{"type":"string"}},"required":["key","value"]}
```

**Parameters**:
```json
{
  "key": "event_patterns",
  "value": "{\"patterns\": [...], \"discovered\": \"2026-04-12\"}"
}
```

**Returns**:
```
✓ Output: "Memory saved successfully"
✗ IsError: true if database error or invalid data
```

**Storage**:
- Stored in BBolt KV database
- Persists across sessions
- Key-value format
- Value can be JSON string

**Example**:
```
save_memory(
  key: "go_patterns",
  value: "{\"patterns\": [\"observer\", \"factory\"], \"date\": \"2026-04-12\"}"
)

save_memory(
  key: "api_schema",
  value: "{\"version\": \"2.0\", \"endpoints\": [...]}"
)
```

---

## Tool Registration

All tools are registered in `/core/tools/registry.go`:

```go
func NewRegistry(trustFolder string, guardian *policies.Guardian, kv db.KVStore) *Registry {
    r := &Registry{
        Tools:       make(map[string]Tool),
        Guardian:    guardian,
        TrustFolder: trustFolder,
        KV:          kv,
    }

    r.Register(&ReadFileTool{...})
    r.Register(&WriteFileTool{...})
    r.Register(&ReadFolderTool{...})
    r.Register(&EditTool{...})
    r.Register(&FindFilesTool{...})
    r.Register(&GrepSearchTool{...})
    r.Register(&ShellTool{...})
    r.Register(&SaveMemoryTool{...})
    r.Register(&GoogleSearchTool{})
    r.Register(&WebFetchTool{})

    return r
}
```

---

## Tool Execution Flow

```go
// From /core/tools/tool.go
type Tool interface {
    Name() string                                              // "read_file"
    Description() string                                       // "Reads file content..."
    Schema() json.RawMessage                                   // JSON schema
    Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}

type ToolResult struct {
    Output     string                 // Result content
    IsError    bool                    // true if failed
    SnapshotID string                 // Snapshot reference (if applicable)
    Metadata   map[string]interface{} // Extra data
}
```

---

## Error Handling

All tools follow consistent error handling:

1. **Trust Folder Violation**
   ```
   IsError: true
   Output: "File path outside Trust Folder"
   ```

2. **File Not Found**
   ```
   IsError: true
   Output: "File not found: path/to/file"
   ```

3. **Permission Denied**
   ```
   IsError: true
   Output: "Permission denied for path/to/file"
   ```

4. **Network Error**
   ```
   IsError: true
   Output: "Network timeout or connection refused"
   ```

5. **Invalid Input**
   ```
   IsError: true
   Output: "Invalid [tool_name] parameters: [details]"
   ```

---

## Tool Chains (Recommended Combinations)

### Code Analysis
```
1. find_files("**/*.go")      → Find all Go files
2. grep_search("func.*Query") → Find Query functions
3. read_file("handler.go")    → Read implementation
```

### Code Modification
```
1. read_file("main.go")    → Read to understand
2. grep_search("func main") → Find location
3. edit("main.go", ...)    → Make modification
4. run_shell_command("go build") → Verify
```

### Research
```
1. google_search("topic") → Search web
2. web_fetch(URL)         → Fetch article
3. save_memory(...)       → Store findings
```

### Documentation
```
1. find_files("**/*.md")   → Find docs
2. read_file("README.md")  → Read content
3. web_fetch(url)          → Get external refs
4. write_file("index.md")  → Create summary
```

---

## Performance Characteristics

```
Tool                  | Time    | Network | CPU | I/O
----------------------|---------|---------|-----|-----
read_file             | 1-2s    | No      | Low | High
write_file            | 1-2s    | No      | Low | High
read_folder           | 1-2s    | No      | Low | High
find_files            | 2-5s    | No      | Med | High
grep_search           | 2-5s    | No      | High| High
edit                  | 5-10s   | No      | High| High
run_shell_command     | 5-30s   | Depends | High| High
google_search         | 5-10s   | Yes     | Low | Med
web_fetch             | 3-8s    | Yes     | Low | Med
save_memory           | 1-2s    | No      | Low | Med
```

---

## Quick Reference Table

| Tool | Input Type | Output Type | Speed | Network | Trust Folder |
|------|-----------|------------|-------|---------|--------------|
| read_file | path | text | Fast | No | Required |
| write_file | path, content | status | Fast | No | Required |
| read_folder | path | list | Fast | No | Required |
| find_files | pattern | paths | Medium | No | Required |
| grep_search | pattern, path | lines | Medium | No | Required |
| edit | path, instruction | status | Slow | No | Required |
| run_shell_command | command | output | Medium-Slow | Maybe | Required |
| google_search | query | results | Medium | Yes | N/A |
| web_fetch | URL | text | Medium | Yes | N/A |
| save_memory | key, value | status | Fast | No | N/A |

---

## Version & Standards

- **Vectora Version**: 0.1.0
- **Tool Count**: 10
- **Last Updated**: 2026-04-12
- **API Standard**: JSON-RPC 2.0 (via IPC)
- **Schema Format**: JSON Schema draft 7

---

## See Also

- `/core/tools/` - Tool implementations
- `/core/instructions/instructions.md` - Model training
- `/TOOLS_DOCUMENTATION.md` - User documentation
- `/IMPLEMENTATION_SUMMARY.md` - What was implemented
