# Vectora MCP Workflows - Practical Examples

This guide shows you **real, working examples** of using Vectora's tools via Claude Code. Each workflow is a complete task from start to finish.

---

## Workflow 1: Semantic Code Search

### What This Does
Search your codebase semantically to understand how features are implemented. Perfect for:
- Understanding existing code before making changes
- Finding similar patterns across files
- Discovering code you forgot about
- Learning how the team implements features

### Setup
- Ensure Vectora is configured in Claude Code (see CLAUDE_CODE_INTEGRATION.md)
- Have a codebase indexed (or let Vectora index on first use)

### Time Required
2-3 minutes

### Skill Level
🟢 Beginner

---

### Example: Find Authentication Implementation

**Goal**: Understand how user authentication works

**Step 1: Initial Search**
```
In Claude Code, type:

@vectora Semantic search for "how does user authentication work in this codebase?"

Show me:
1. All files involved in login/authentication
2. The flow: from user input to session creation
3. Key functions and what they do
4. Security mechanisms (hashing, tokens, etc)
```

**Expected Output**:
```
Found authentication in:

1. src/auth/handler.go
   - LoginHandler: Validates username/password
   - CreateSession: Creates JWT token

2. src/database/models.go
   - User struct with password hash

3. src/middleware/auth.go
   - VerifyToken: Validates JWT on each request

The flow is:
User Input → LoginHandler → VerifyPassword → CreateSession → JWT Token
            → Middleware validates on each request

Security: bcrypt for passwords, HS256 for JWT
```

**Step 2: Deep Dive into Specific File**

```
@vectora Analyze src/auth/handler.go in detail.

Explain:
1. What each function does
2. Security considerations
3. Potential issues or improvements
4. How it integrates with the rest of the app
```

**Expected Output**:
```
LoginHandler (lines 12-45):
- Takes username, password
- Queries database for user
- Compares passwords with bcrypt
- Returns error if not found
- Security: Good (bcrypt, no timing attacks)
- Issue: No rate limiting (could be brute forced)

CreateSession (lines 47-60):
- Generates JWT token
- Sets expiration (24 hours)
- Stores session in cache
- Returns to client
- Note: Cache is in-memory (lost on restart)
```

### Tips & Tricks

✅ **What works well**:
- Searching for "how does X work"
- Finding all code related to a feature
- Understanding security implementations
- Discovering forgotten code

❌ **Won't work as well**:
- Searching for exact function names (use IDE search for that)
- Very new code not yet indexed
- External libraries (searches codebase only)

### Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| No results found | Index your codebase first: `vectora embed` |
| Results seem irrelevant | Phrase query more like human language |
| Too many results | Add more specific keywords |
| Response timeout | Large codebase - try more specific query |

---

## Workflow 2: Generate Documentation

### What This Does
Automatically generate professional documentation from your code. Great for:
- Creating API documentation (OpenAPI/Swagger)
- Documenting architecture decisions
- Finding undocumented functions
- Creating README sections

### Setup
- Vectora configured in Claude Code
- Codebase indexed
- Understand what kind of docs you want to create

### Time Required
5-15 minutes (depending on codebase size)

### Skill Level
🟡 Intermediate

---

### Example: Generate API Documentation

**Goal**: Create OpenAPI documentation for REST endpoints

**Step 1: Analyze Endpoints**

```
@vectora Analyze the REST API structure.

Extract all HTTP endpoints:
- Method (GET, POST, PUT, DELETE)
- Path
- Request body structure
- Response structure
- Authentication required?
- Error responses

Format as a structured list.
```

**Expected Output**:
```
REST API Endpoints:

POST /auth/login
  Request: {username: string, password: string}
  Response: {token: string, expires: number}
  Auth: None
  Errors: 401 (bad credentials), 400 (missing fields)

GET /users/:id
  Request: Path param 'id'
  Response: {id, username, email, createdAt}
  Auth: JWT required
  Errors: 401 (no auth), 404 (user not found)

[... more endpoints ...]
```

**Step 2: Generate Documentation**

```
@vectora Generate OpenAPI 3.0 documentation.

Use the endpoints analyzed in Step 1.

For each endpoint provide:
1. Full path and method
2. Summary and description
3. Parameters (path, query, body)
4. Response schema
5. Example request/response
6. Authentication and authorization

Output valid YAML that can be used with Swagger UI.
```

**Expected Output**:
```yaml
openapi: 3.0.0
info:
  title: My API
  version: 1.0.0

paths:
  /auth/login:
    post:
      summary: User Login
      description: Authenticate user with credentials
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                password:
                  type: string
      responses:
        '200':
          description: Login successful
          content:
            application/json:
              schema:
                type: object
                properties:
                  token:
                    type: string
                  expires:
                    type: number
        '401':
          description: Invalid credentials
```

**Step 3: Find Undocumented Features**

```
@vectora Compare the API documentation from Step 2
with the actual code implementation.

Find:
1. Functions/endpoints in code but not in docs
2. Documentation that doesn't match implementation
3. Missing error handling in docs
4. Deprecated code still documented

List each issue with:
- What's missing/wrong
- Where in code
- Suggested fix
```

**Expected Output**:
```
Undocumented/Wrong:

1. GET /admin/stats
   - In code but missing from docs
   - Requires admin auth
   - Returns usage statistics
   - Should be added to OpenAPI

2. POST /users response
   - Docs say returns full user object
   - Code only returns {id, username}
   - Update docs to match implementation

3. Error responses
   - Code can return 500 (server error)
   - Docs don't document this
   - Add error handling to OpenAPI
```

### Tips & Tricks

✅ **What works well**:
- Generating OpenAPI/Swagger specs
- Finding documentation gaps
- Creating architecture documentation
- API reference generation

❌ **Won't work well**:
- Markdown tutorials (needs human storytelling)
- Very complex business logic docs
- UI/UX documentation

### Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| Generated docs don't match code | Run Step 3 to find differences |
| Missing endpoints in output | May need to index more files |
| Schema looks wrong | Ask Claude to refine based on examples |
| YAML syntax errors | Ask Claude to validate against OpenAPI spec |

---

## Workflow 3: Detect Code Patterns & Issues

### What This Does
Automatically find potential bugs, code smells, and architectural issues. Useful for:
- Finding security vulnerabilities
- Detecting memory leaks (Go, Rust, C++)
- Finding race conditions
- Identifying anti-patterns
- Code quality assessment

### Setup
- Vectora configured
- Codebase indexed
- Knowledge of what issues to look for

### Time Required
5-20 minutes (depending on codebase size and issue complexity)

### Skill Level
🟡 Intermediate to 🔴 Advanced

---

### Example: Find Race Conditions in Go Code

**Goal**: Find potential race conditions in concurrent code

**Step 1: Analyze Concurrency**

```
@vectora Analyze all concurrent code in this Go codebase.

Find:
1. All goroutines (go statements)
2. All synchronization (mutex, channels, atomics)
3. All shared data accessed by multiple goroutines
4. Missing synchronization that could cause races

For each, show:
- Location (file:line)
- What data is shared
- How it's synchronized (or not)
- Risk level (high/medium/low)
```

**Expected Output**:
```
Concurrency Analysis:

Safe Patterns (properly synchronized):

1. Cache access (cache/cache.go:25)
   - Shared: map[string]Value
   - Synchronization: sync.RWMutex
   - Risk: Low ✅

2. Message queue (queue/queue.go:40)
   - Shared: Channel of messages
   - Synchronization: Go channel (built-in)
   - Risk: Low ✅

Potential Race Conditions:

1. Global config (config.go:5)
   - Shared: var config Config
   - Synchronization: None
   - Accessed by: HTTP handlers + background job
   - Risk: HIGH 🔴
   - Issue: Multiple goroutines read/write config

2. Statistics counter (stats.go:12)
   - Shared: var requestCount int
   - Synchronization: None
   - Accessed by: Middleware (multiple goroutines)
   - Risk: HIGH 🔴
   - Issue: Read/modify/write not atomic
```

**Step 2: Suggest Fixes**

```
@vectora For each HIGH risk race condition from Step 1,
suggest a fix.

For each fix show:
1. Current code (problematic)
2. Fixed code (safe)
3. Why the fix works
4. Performance impact
```

**Expected Output**:
```
Fix 1: Global config race condition

CURRENT (UNSAFE):
var config Config

func updateConfig(c Config) {
    config = c  // Race: another goroutine might read while writing
}

FIXED:
var configMu sync.RWMutex
var config Config

func updateConfig(c Config) {
    configMu.Lock()
    defer configMu.Unlock()
    config = c
}

Why: Lock ensures only one writer, readers wait
Performance: Minimal (lock duration is microseconds)
```

### Tips & Tricks

✅ **What works well**:
- Finding race conditions (Go concurrency)
- Finding memory leaks
- Finding nil pointer risks
- Identifying missing error handling
- Finding security issues (SQL injection, etc)

❌ **Won't work well**:
- Logic bugs (would need unit tests)
- Performance issues (needs profiling)
- Architectural flaws (subjective)

### Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| Too many false positives | Refine query to be more specific |
| Misses real issues | May be code pattern not in indexed files |
| Unclear what risk means | Ask Claude to explain the specific risk |
| Suggested fix doesn't work | Share error message with Claude |

---

## Workflow 4: Refactor Code Using Context

### What This Does
Refactor code intelligently using context from the rest of the codebase. Useful for:
- Standardizing error handling
- Modernizing old code
- Applying design patterns
- Consistency improvements

### Setup
- Vectora configured
- Good understanding of target pattern
- Already done: Understand current state (Workflow 3)

### Time Required
10-30 minutes

### Skill Level
🔴 Advanced

---

### Example: Standardize Error Handling

**Goal**: Standardize how errors are handled across codebase

**Step 1: Analyze Current Error Handling**

```
@vectora Analyze error handling patterns in this codebase.

Show:
1. How errors are currently handled in different places
2. Inconsistencies (some functions panic, some return errors, etc)
3. Error types used
4. What the ideal pattern should be
5. Which files need updating

Organize by:
- Current pattern (e.g., "panics on error")
- Number of files using it
- Example code
- Risk level (should we even be doing this?)
```

**Expected Output**:
```
Error Handling Analysis:

Current Patterns:

1. PANIC ON ERROR (5 files)
   - Files: cmd/main.go, cmd/server.go, cmd/migrate.go
   - Example: log.Fatal(err)
   - Risk: High - crashes entire app
   - Used for: One-time startup errors

2. SILENT IGNORE (3 files)
   - Files: handlers/cache.go, handlers/stats.go
   - Example: _ = cache.Set(key, value)
   - Risk: High - errors go unnoticed
   - Used for: Non-critical operations

3. LOG AND CONTINUE (8 files)
   - Files: handlers/*.go, middleware/*.go
   - Example: if err != nil { log.Error(err); return }
   - Risk: Medium - inconsistent return values
   - Used for: Request handling

4. RETURN ERROR (4 files)
   - Files: service/*.go, database/*.go
   - Example: return fmt.Errorf("failed to fetch: %w", err)
   - Risk: Low - clear error propagation
   - Used for: Core business logic

RECOMMENDATION:
Use pattern #4 (return error) everywhere.
Create custom error types for different failure modes.
```

**Step 2: Design Target Pattern**

```
@vectora Design a standard error handling pattern for this project.

Requirements:
1. Errors should bubble up to handlers
2. Handlers should decide what to do (respond to user, log, retry)
3. Errors should be categorized (input error, server error, etc)
4. Error messages should be safe to show to users

Create:
1. Custom error types
2. Error handling middleware
3. Example usage in 3 different contexts (handler, service, database)
4. Testing strategy

Show complete code for each.
```

**Expected Output**:
```go
// 1. Custom error types
package errors

type ErrorType int

const (
    ValidationError ErrorType = 400
    NotFoundError   ErrorType = 404
    ServerError     ErrorType = 500
)

type APIError struct {
    Type      ErrorType
    Code      string
    Message   string
    Internal  error // Only for logging
}

// 2. Middleware
func ErrorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Handle panic as 500
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// 3. Example usage in handler
func LoginHandler(w http.ResponseWriter, r *http.Request) error {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return &APIError{
            Type: ValidationError,
            Code: "INVALID_JSON",
            Message: "Request body is invalid",
            Internal: err,
        }
    }
    // ... more code
}

// 4. Example in service
func (s *Service) Login(username, password string) (*Session, error) {
    user, err := s.db.GetUser(username)
    if err != nil {
        // Wrap database errors
        return nil, fmt.Errorf("failed to fetch user: %w", err)
    }
    // ... more code
}
```

**Step 3: Refactor Priority Files**

```
@vectora From Step 1, we found errors handled 5 different ways.

Start with the riskiest files:
1. handlers/cache.go - silently ignoring errors

Show:
- Current code (before)
- Refactored code (after)
- What changed and why
- How to test the changes

Make it diff-like so I can review line-by-line.
```

**Expected Output**:
```diff
--- handlers/cache.go (original)
+++ handlers/cache.go (refactored)

-    value, err := cache.Get(key)
-    _ = err  // silently ignore
+    value, err := cache.Get(key)
+    if err != nil {
+        // Log for debugging, but don't fail the request
+        log.WithError(err).Warn("Cache miss", "key", key)
+        return &APIError{
+            Type: ServerError,
+            Code: "CACHE_FAILED",
+            Message: "Failed to retrieve cached data",
+            Internal: err,
+        }
+    }
```

### Tips & Tricks

✅ **What works well**:
- Error handling standardization
- Applying consistent patterns
- Removing anti-patterns
- Code consistency improvements

❌ **Won't work well**:
- Changing business logic
- Rewriting algorithms
- Major architectural changes
- Changing API contracts

### Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| Refactored code doesn't compile | Share compilation error with Claude |
| Tests fail after refactoring | Ask Claude to update tests alongside code |
| Unclear why change was made | Ask Claude to explain the rationale |
| Need to update multiple files | Do 1-2 files to understand pattern, then apply to others |

---

## General Tips for All Workflows

### ✅ What Works Well

1. **Semantic understanding** - "How does X work?"
2. **Pattern finding** - "Find all places where we do X"
3. **Documentation generation** - "Create docs for X"
4. **Code analysis** - "Analyze X for issues"
5. **Error detection** - "Find potential bugs in X"

### ❌ Won't Work Well

1. **Exact code matching** - Use IDE search for that
2. **Logic bugs** - Need unit tests to find
3. **Performance bottlenecks** - Need profiling data
4. **UI/UX decisions** - Subjective
5. **Business logic changes** - Need human review

### Pro Tips

💡 **Be specific**: Instead of "explain the code", ask "how does user authentication work"

💡 **Break into steps**: Multi-step workflows are more reliable than one big question

💡 **Show examples**: If Vectora misunderstands, provide code examples

💡 **Ask for structured output**: "Format as JSON" or "Format as table" helps

💡 **Iterate**: Start broad, then drill down to specific details

---

## Next Steps

1. ✅ Pick one workflow that applies to your codebase
2. ✅ Follow the step-by-step instructions
3. ✅ Adapt the prompts to your specific needs
4. ✅ Share your results with your team

For more advanced patterns and multi-step workflows, see the troubleshooting section or ask Vectora directly!

---

_Vectora MCP Workflows_
_Production-ready examples for Claude Code integration_
