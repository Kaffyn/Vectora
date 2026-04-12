# Vectora MCP Protocol Reference

Complete technical reference for all 11 tools exposed via the MCP (Model Context Protocol) interface.

**Protocol**: JSON-RPC 2.0 over stdio
**Version**: 1.0
**Date**: 2026-04-12

---

## Overview

These 11 tools are available when Vectora runs as an MCP server (`vectora mcp /workspace`):

### Category: Semantic Search & Indexing (3 tools)
1. **`embed`** — Index files into vector database
2. **`search_database`** — Semantic search across indexed content
3. **`web_search_and_embed`** — Research + index from web

### Category: Code Analysis (4 tools)
4. **`analyze_code_patterns`** — Detect patterns and anti-patterns
5. **`knowledge_graph_analysis`** — Entity extraction & relationships
6. **`doc_coverage_analysis`** — Documentation completeness
7. **`bug_pattern_detection`** — Find potential bugs

### Category: Generation & Refactoring (4 tools)
8. **`test_generation`** — Generate test cases
9. **`refactor_with_context`** — Smart refactoring with RAG
10. **`plan_mode`** — Structured decomposition
11. **`web_fetch_and_embed`** — Fetch & index from URL

---

## 1. embed

Index files into the vector database for semantic search.

### Purpose
Creates embeddings for code files, making them searchable via semantic queries.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Relative path to file (within workspace)"
    },
    "content": {
      "type": "string",
      "description": "File content to embed (optional, reads file if omitted)"
    },
    "force_reindex": {
      "type": "boolean",
      "description": "Force reindexing even if already indexed",
      "default": false
    }
  },
  "required": ["path"]
}
```

### Example Usage

**Simple**: Embed a file by path
```
Tool Call: embed
Input: {
  "path": "src/handlers/auth.go"
}
Output: {
  "vector_id": "vec_123abc",
  "tokens": 245,
  "dimensions": 768,
  "indexed_at": "2026-04-12T10:30:00Z"
}
```

**Advanced**: Embed content directly
```
Tool Call: embed
Input: {
  "path": "src/new_feature.go",
  "content": "package main\n\nfunc NewFeature() {...}",
  "force_reindex": true
}
```

### When to Use
- Indexing new files added to project
- Re-indexing after major changes
- Preparing files for semantic search
- Initial codebase indexing

### When NOT to Use
- Don't embed binary files (use text files only)
- Don't embed generated code (cache files, build output)
- Don't re-index constantly (use after significant changes)

### Performance
- **Time**: 2-5 seconds per file (depends on file size)
- **Throughput**: ~100-200 files per minute

---

## 2. search_database

Semantic search across indexed content.

### Purpose
Find code semantically similar to your query, across entire indexed codebase.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Natural language search query"
    },
    "limit": {
      "type": "integer",
      "description": "Max results to return",
      "default": 10,
      "minimum": 1,
      "maximum": 50
    },
    "threshold": {
      "type": "number",
      "description": "Relevance threshold (0.0-1.0)",
      "default": 0.5,
      "minimum": 0,
      "maximum": 1
    }
  },
  "required": ["query"]
}
```

### Example Usage

**Search for authentication patterns**
```
Tool Call: search_database
Input: {
  "query": "user authentication and JWT validation",
  "limit": 10,
  "threshold": 0.6
}
Output: {
  "results": [
    {
      "file": "src/auth/jwt.go",
      "line": 42,
      "relevance": 0.94,
      "snippet": "func ValidateToken(token string) (*User, error) {...}",
      "context": "Package jwt implements token validation"
    },
    {
      "file": "src/handlers/login.go",
      "line": 18,
      "relevance": 0.87,
      "snippet": "user, err := auth.Login(email, password)",
      "context": "Handle POST /login requests"
    }
  ],
  "total_matches": 23,
  "search_time_ms": 1240
}
```

**Search for error handling patterns**
```
Tool Call: search_database
Input: {
  "query": "error handling and recovery strategies",
  "limit": 5
}
```

### When to Use
- Find similar implementations to learn from
- Locate patterns you want to reuse
- Search for specific functionality
- Understand how things are done in codebase

### When NOT to Use
- Use `grep_search` for exact string matching
- Don't use for simple name searching (use `find_files`)
- For syntax-specific searches use `analyze_code_patterns`

### Performance
- **Time**: 2-5 seconds (indexed search)
- **Accuracy**: Based on relevance threshold
- **Precision**: Improves with more specific queries

---

## 3. web_search_and_embed

Search the web for information and add results to searchable index.

### Purpose
Research topics online and make them searchable alongside your codebase.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Search query (e.g., 'Go concurrency patterns')"
    },
    "num_results": {
      "type": "integer",
      "description": "Number of results to fetch",
      "default": 5,
      "minimum": 1,
      "maximum": 20
    },
    "include_fetch": {
      "type": "boolean",
      "description": "Also fetch and embed full content",
      "default": true
    }
  },
  "required": ["query"]
}
```

### Example Usage

**Research architecture patterns**
```
Tool Call: web_search_and_embed
Input: {
  "query": "event-driven architecture Go implementation 2026",
  "num_results": 10,
  "include_fetch": true
}
Output: {
  "search_results": [
    {
      "title": "Event-Driven Architecture with Go",
      "url": "https://example.com/go-events",
      "snippet": "Event-driven systems allow...",
      "relevance": 0.92
    }
  ],
  "fetched_documents": 8,
  "indexed_documents": 8,
  "total_tokens": 45000,
  "search_time_ms": 3200
}
```

### When to Use
- Learning best practices for your domain
- Researching specific patterns or technologies
- Gathering reference implementations
- Building knowledge base from external sources

### When NOT to Use
- Don't use for common queries (results vary widely)
- Not ideal for fetching specific documentation (use `web_fetch_and_embed`)
- Skip if offline (requires internet)

### Performance
- **Time**: 10-20 seconds (includes web fetch)
- **Coverage**: Depends on search result relevance
- **Storage**: Adds to vector database

---

## 4. analyze_code_patterns

Detect patterns, anti-patterns, and architectural structures.

### Purpose
Analyze code to identify design patterns, anti-patterns, and architectural decisions.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "File or directory to analyze (within workspace)"
    },
    "pattern_type": {
      "type": "string",
      "enum": [
        "concurrency",
        "error_handling",
        "performance",
        "security",
        "testing",
        "design_patterns",
        "code_quality",
        "all"
      ],
      "description": "Type of patterns to detect",
      "default": "all"
    },
    "severity": {
      "type": "string",
      "enum": ["low", "medium", "high", "critical"],
      "description": "Minimum severity to report",
      "default": "low"
    }
  },
  "required": ["path"]
}
```

### Example Usage

**Find concurrency issues**
```
Tool Call: analyze_code_patterns
Input: {
  "path": "src/",
  "pattern_type": "concurrency"
}
Output: {
  "patterns_found": [
    {
      "type": "goroutine_without_cleanup",
      "severity": "high",
      "file": "src/worker.go",
      "line": 42,
      "description": "Goroutine launched but never cancelled",
      "code": "go doWork(data)",
      "recommendation": "Use context for cancellation"
    },
    {
      "type": "shared_memory_access",
      "severity": "high",
      "file": "src/cache.go",
      "line": 18,
      "description": "Concurrent map access without mutex",
      "recommendation": "Use sync.Map or protect with mutex"
    }
  ],
  "total_patterns": 7,
  "analysis_time_ms": 3500
}
```

**Analyze all patterns in a file**
```
Tool Call: analyze_code_patterns
Input: {
  "path": "src/handlers/api.go",
  "pattern_type": "all",
  "severity": "medium"
}
```

### When to Use
- Review code before submitting PR
- Identify architectural patterns
- Find potential race conditions
- Detect performance issues
- Security code review

### When NOT to Use
- For simple linting (use language-specific linters)
- On-the-fly as you type (too slow)
- For formatting issues (use formatters)

### Performance
- **Time**: 5-15 seconds (depends on code size)
- **Accuracy**: High for common patterns
- **Coverage**: All pattern types scanned

---

## 5. knowledge_graph_analysis

Extract entities, relationships, and semantic structure.

### Purpose
Build understanding of how concepts and components relate to each other.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Code path to analyze"
    },
    "entity_types": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Types to extract: 'functions', 'classes', 'interfaces', 'types', 'all'",
      "default": ["all"]
    },
    "include_relationships": {
      "type": "boolean",
      "description": "Extract how entities relate to each other",
      "default": true
    },
    "depth": {
      "type": "integer",
      "description": "Relationship depth (1-5)",
      "default": 2,
      "minimum": 1,
      "maximum": 5
    }
  },
  "required": ["path"]
}
```

### Example Usage

**Extract API structure**
```
Tool Call: knowledge_graph_analysis
Input: {
  "path": "src/api/",
  "entity_types": ["functions", "types"],
  "include_relationships": true,
  "depth": 2
}
Output: {
  "entities": [
    {
      "name": "HandleQuery",
      "type": "function",
      "file": "handler.go",
      "signature": "func HandleQuery(ctx Context, q Query) (Result, error)"
    },
    {
      "name": "Query",
      "type": "interface",
      "file": "types.go",
      "methods": ["Execute", "Validate"]
    }
  ],
  "relationships": [
    {
      "source": "HandleQuery",
      "target": "Query",
      "type": "uses",
      "description": "HandleQuery accepts Query interface"
    },
    {
      "source": "Query",
      "target": "Executor",
      "type": "depends_on",
      "description": "Query depends on Executor"
    }
  ],
  "total_entities": 42,
  "total_relationships": 63,
  "analysis_time_ms": 2800
}
```

### When to Use
- Understanding large codebases
- Documenting architecture
- Identifying component boundaries
- Finding circular dependencies
- Planning refactoring

### When NOT to Use
- For simple file listing (use `find_files`)
- Performance-critical tasks (can be slow)
- On small functions/classes (overkill)

### Performance
- **Time**: 5-20 seconds (depends on code complexity)
- **Accuracy**: High for syntactic structures
- **Depth Impact**: Deeper = slower but more complete

---

## 6. doc_coverage_analysis

Measure and improve documentation completeness.

### Purpose
Identify what's documented, what's missing, and suggest improvements.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Code path to analyze"
    },
    "languages": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Languages: 'go', 'typescript', 'python', 'all'",
      "default": ["all"]
    },
    "check_examples": {
      "type": "boolean",
      "description": "Check for code examples",
      "default": true
    },
    "report_format": {
      "type": "string",
      "enum": ["summary", "detailed", "csv"],
      "description": "Output format",
      "default": "detailed"
    }
  },
  "required": ["path"]
}
```

### Example Usage

**Analyze documentation coverage**
```
Tool Call: doc_coverage_analysis
Input: {
  "path": "src/",
  "languages": ["go"],
  "check_examples": true,
  "report_format": "detailed"
}
Output: {
  "summary": {
    "total_entities": 156,
    "documented": 89,
    "coverage_percent": 57,
    "examples_count": 12
  },
  "gaps": [
    {
      "entity": "QueryEngine.Execute",
      "type": "function",
      "severity": "high",
      "file": "engine.go:42",
      "suggestion": "Add docstring explaining parameters and return values"
    },
    {
      "entity": "CacheStore",
      "type": "interface",
      "severity": "medium",
      "file": "cache.go:18",
      "suggestion": "Add usage examples"
    }
  ],
  "quality_issues": [
    {
      "issue": "Outdated docstring",
      "file": "handler.go:100",
      "description": "References removed parameter 'timeout'"
    }
  ],
  "recommendations": [
    "Document all public APIs (currently 43% undocumented)",
    "Add examples for 8 complex functions",
    "Update 5 outdated docstrings"
  ]
}
```

### When to Use
- Before releasing new versions
- Planning documentation improvements
- Code review (checking quality)
- Onboarding new developers
- Audit code quality

### When NOT to Use
- Realtime checking (too slow)
- On documentation files themselves (focus on code)

### Performance
- **Time**: 5-10 seconds
- **Accuracy**: High for docstrings, moderate for examples
- **Scope**: Entire codebase or directory

---

## 7. bug_pattern_detection

Find potential bugs and issues using pattern matching.

### Purpose
Identify common bugs, security issues, and potential problems before they cause failures.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Code path to scan"
    },
    "check_types": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Check types: 'security', 'performance', 'correctness', 'all'",
      "default": ["all"]
    },
    "severity_threshold": {
      "type": "string",
      "enum": ["low", "medium", "high", "critical"],
      "description": "Only report this severity and above",
      "default": "medium"
    }
  },
  "required": ["path"]
}
```

### Example Usage

**Find security issues**
```
Tool Call: bug_pattern_detection
Input: {
  "path": "src/",
  "check_types": ["security"],
  "severity_threshold": "high"
}
Output: {
  "issues": [
    {
      "type": "sql_injection",
      "severity": "critical",
      "file": "src/db/query.go",
      "line": 34,
      "code": "query := \"SELECT * FROM users WHERE id=\" + userID",
      "description": "User input concatenated to SQL query",
      "fix": "Use prepared statements: db.Query(\"SELECT * FROM users WHERE id=?\", userID)"
    },
    {
      "type": "hardcoded_secret",
      "severity": "high",
      "file": ".env.example",
      "line": 2,
      "code": "API_KEY=sk-1234567890abcdef",
      "description": "Hardcoded API key in file"
    },
    {
      "type": "missing_input_validation",
      "severity": "high",
      "file": "src/handlers/api.go",
      "line": 15,
      "description": "User input used without validation"
    }
  ],
  "total_issues": 12,
  "critical_count": 2,
  "scan_time_ms": 4200
}
```

**Find performance issues**
```
Tool Call: bug_pattern_detection
Input: {
  "path": "src/",
  "check_types": ["performance"],
  "severity_threshold": "high"
}
```

### When to Use
- Code review before deployment
- Security audit
- Refactoring large modules
- Finding bugs to fix

### When NOT to Use
- During development (false positives)
- On auto-generated code
- Performance-critical operations

### Performance
- **Time**: 5-15 seconds
- **Accuracy**: High for common patterns
- **False Positives**: Expected in complex code

---

## 8. test_generation

Generate test cases from code automatically.

### Purpose
Create comprehensive test suites based on code structure and logic.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Function or file to generate tests for"
    },
    "test_type": {
      "type": "string",
      "enum": ["unit", "integration", "edge_cases", "all"],
      "description": "Type of tests to generate",
      "default": "unit"
    },
    "include_mocks": {
      "type": "boolean",
      "description": "Generate mock implementations",
      "default": true
    },
    "framework": {
      "type": "string",
      "enum": ["auto", "jest", "pytest", "go_test", "custom"],
      "description": "Test framework",
      "default": "auto"
    }
  },
  "required": ["path"]
}
```

### Example Usage

**Generate unit tests for function**
```
Tool Call: test_generation
Input: {
  "path": "src/auth/validate.go:ValidateToken",
  "test_type": "unit",
  "include_mocks": true,
  "framework": "go_test"
}
Output: {
  "tests": [
    {
      "test_name": "TestValidateToken_ValidToken",
      "code": "func TestValidateToken_ValidToken(t *testing.T) {\n  token := \"valid.jwt.token\"\n  user, err := ValidateToken(token)\n  if err != nil {\n    t.Fatalf(\"expected no error, got %v\", err)\n  }\n  if user.ID != \"123\" {\n    t.Errorf(\"expected user ID 123, got %s\", user.ID)\n  }\n}",
      "coverage": ["success_path"]
    },
    {
      "test_name": "TestValidateToken_InvalidToken",
      "code": "func TestValidateToken_InvalidToken(t *testing.T) {...}",
      "coverage": ["error_path"]
    },
    {
      "test_name": "TestValidateToken_ExpiredToken",
      "code": "func TestValidateToken_ExpiredToken(t *testing.T) {...}",
      "coverage": ["edge_case"]
    }
  ],
  "test_count": 5,
  "coverage_paths": ["success", "error", "edge_case", "boundary"],
  "mocks_generated": 2
}
```

### When to Use
- Generating test stubs
- Improving test coverage
- Testing complex functions
- Test-driven development

### When NOT to Use
- For integration tests (use tools that handle setup)
- Test-critical code (verify manually)
- Auto-trusting generated code

### Performance
- **Time**: 10-20 seconds
- **Test Count**: 3-10 per function (depends on complexity)
- **Coverage**: Best effort

---

## 9. refactor_with_context

Intelligently refactor code using codebase context via RAG.

### Purpose
Propose and apply refactorings that align with your codebase's patterns and style.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Code to refactor"
    },
    "goal": {
      "type": "string",
      "description": "What to refactor (e.g., 'use event pattern', 'improve error handling')"
    },
    "constraints": {
      "type": "string",
      "description": "Constraints (e.g., 'must remain backward compatible')",
      "default": ""
    },
    "apply_changes": {
      "type": "boolean",
      "description": "Apply changes automatically (vs. preview)",
      "default": false
    }
  },
  "required": ["path", "goal"]
}
```

### Example Usage

**Refactor to use pattern from codebase**
```
Tool Call: refactor_with_context
Input: {
  "path": "src/handler/process.go",
  "goal": "Use the event pattern we use elsewhere in the codebase",
  "constraints": "Must remain backward compatible",
  "apply_changes": false
}
Output: {
  "analysis": "Found 3 similar event-based implementations in codebase",
  "proposed_changes": [
    {
      "type": "refactor_to_event_driven",
      "description": "Convert synchronous call to event pub/sub",
      "before": "result := processor.Process(data)",
      "after": "eventBus.Publish(\"process:start\", data)",
      "rationale": "Consistent with event pattern in src/events/"
    },
    {
      "type": "extract_event_handler",
      "description": "Extract processing logic to separate handler",
      "impact": "Better separation of concerns"
    }
  ],
  "files_affected": ["src/handler/process.go", "src/events/listener.go"],
  "breaking_changes": false,
  "migration_effort": "medium"
}
```

### When to Use
- Large refactoring projects
- Aligning new code with patterns
- Modernizing old code
- Improving architecture incrementally

### When NOT to Use
- Simple changes (direct edit faster)
- Critical production code (review carefully)
- Time-sensitive changes

### Performance
- **Time**: 15-30 seconds (includes RAG search)
- **Accuracy**: Depends on available examples
- **Review**: Always review proposals first

---

## 10. plan_mode

Structured problem decomposition and planning.

### Purpose
Break down complex problems into manageable steps and create implementation plans.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "goal": {
      "type": "string",
      "description": "The goal or problem to solve"
    },
    "context_path": {
      "type": "string",
      "description": "Code context (optional)"
    },
    "constraints": {
      "type": "string",
      "description": "Constraints or requirements"
    },
    "output_format": {
      "type": "string",
      "enum": ["markdown", "json", "outline"],
      "description": "Output format",
      "default": "markdown"
    }
  },
  "required": ["goal"]
}
```

### Example Usage

**Plan refactoring**
```
Tool Call: plan_mode
Input: {
  "goal": "Refactor event system to support async subscriptions",
  "context_path": "src/events/",
  "constraints": "Backward compatible, no external dependencies",
  "output_format": "markdown"
}
Output: {
  "plan": "# Event System Async Refactoring Plan\n\n## Phase 1: Design (2-3 hours)\n1. Analyze current architecture\n2. Design async subscription interface\n3. Plan migration strategy\n\n## Phase 2: Implementation (1-2 days)\n1. Add async subscriber support\n2. Update EventBus.Subscribe() signature\n3. Create compatibility layer\n...",
  "phases": [
    {
      "phase": "Design",
      "duration": "2-3 hours",
      "tasks": ["Analyze architecture", "Design async interface", "Plan migration"],
      "dependencies": []
    },
    {
      "phase": "Implementation",
      "duration": "1-2 days",
      "tasks": ["Add async support", "Update signatures", "Create compatibility"],
      "dependencies": ["Design"]
    }
  ],
  "estimated_total": "3-5 days",
  "risks": [
    "Breaking existing code if not careful with compatibility",
    "Performance impact of async calls"
  ]
}
```

### When to Use
- Planning major refactoring
- Breaking down complex features
- Scoping work for team
- Understanding large changes

### When NOT to Use
- Trivial tasks (overkill)
- Fast decisions needed
- Real-time problem-solving

### Performance
- **Time**: 5-15 seconds
- **Accuracy**: Good for structured thinking
- **Detail Level**: Medium (can add more depth)

---

## 11. web_fetch_and_embed

Fetch content from URL and add to searchable index.

### Purpose
Get external documentation/articles and make them searchable alongside your code.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "url": {
      "type": "string",
      "description": "URL to fetch (must be HTTPS)"
    },
    "title": {
      "type": "string",
      "description": "Custom title for indexed content",
      "default": ""
    },
    "tags": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Tags for categorization"
    }
  },
  "required": ["url"]
}
```

### Example Usage

**Fetch and index API documentation**
```
Tool Call: web_fetch_and_embed
Input: {
  "url": "https://pkg.go.dev/context",
  "title": "Go context package",
  "tags": ["documentation", "go", "stdlib"]
}
Output: {
  "url": "https://pkg.go.dev/context",
  "title": "Go context package",
  "content_length": 45320,
  "indexed_tokens": 3420,
  "vector_id": "web_abc123",
  "chunks": 12,
  "indexed_at": "2026-04-12T10:35:00Z"
}
```

**Fetch article about patterns**
```
Tool Call: web_fetch_and_embed
Input: {
  "url": "https://example.com/go-patterns-guide",
  "tags": ["patterns", "go", "architecture"],
  "title": "Go Patterns Guide 2026"
}
```

### When to Use
- Adding reference documentation
- Researching specific topics
- Storing articles for later search
- Building knowledge base

### When NOT to Use
- For quick reference (directly search instead)
- Large documents (can be slow)
- Paywalled content (may fail)

### Performance
- **Time**: 3-8 seconds (includes fetch + embedding)
- **Size Limit**: Up to ~100KB per document
- **Timeout**: 10 seconds

---

## Error Handling

All tools follow consistent error format:

### Success Response
```json
{
  "status": "success",
  "data": {...},
  "metadata": {
    "execution_time_ms": 1234,
    "timestamp": "2026-04-12T10:30:00Z"
  }
}
```

### Error Response
```json
{
  "status": "error",
  "error": {
    "code": "WORKSPACE_NOT_FOUND",
    "message": "Workspace path /invalid/path does not exist",
    "details": {
      "path": "/invalid/path",
      "type": "validation_error"
    }
  },
  "metadata": {
    "execution_time_ms": 45
  }
}
```

### Common Errors

| Error Code | Meaning | Solution |
|------------|---------|----------|
| `WORKSPACE_NOT_FOUND` | Workspace path invalid | Verify path exists |
| `PERMISSION_DENIED` | Can't read file | Check permissions |
| `TRUST_FOLDER_VIOLATION` | File outside workspace | Files must be within workspace |
| `INVALID_INPUT` | Bad parameters | Check input schema |
| `NETWORK_ERROR` | Network issue (web tools) | Check internet, retry |
| `TIMEOUT` | Tool execution timeout | Reduce scope, retry |
| `API_ERROR` | LLM/API error | Check API keys, rate limits |

---

## Tool Chains (Recommended Sequences)

### Pattern 1: Code Discovery & Analysis
```
1. find_files("**/*.go") → Get all files
2. analyze_code_patterns("src/") → Find patterns
3. knowledge_graph_analysis("src/") → Understand relationships
4. doc_coverage_analysis("src/") → Check documentation
```

### Pattern 2: Feature Implementation
```
1. search_database("query about feature") → Find similar
2. plan_mode("goal of feature", "src/") → Plan approach
3. test_generation("handler.go:NewFeature") → Generate tests
4. refactor_with_context("handler.go", "use pattern") → Align with codebase
```

### Pattern 3: Learning & Research
```
1. web_search_and_embed("topic in Go") → Research
2. web_fetch_and_embed("documentation URL") → Get detailed docs
3. search_database("topic") → Compare with codebase
4. analyze_code_patterns("src/", "all") → Learn patterns used
```

### Pattern 4: Security Review
```
1. bug_pattern_detection("src/", "security", "high") → Find issues
2. analyze_code_patterns("src/", "security") → Review patterns
3. doc_coverage_analysis("src/") → Check what's documented
4. refactor_with_context("path", "improve security") → Fix
```

---

## Usage Tips

### Getting Better Results

1. **Be Specific**: "Find authentication handlers" beats "Find things"
2. **Use Keywords**: Include domain terms the code uses
3. **Set Limits**: Use `limit` parameter to reduce noise
4. **Iterate**: Refine queries based on results

### Performance Optimization

1. **Embed incrementally**: Don't re-embed entire codebase
2. **Use thresholds**: Higher threshold = fewer, better results
3. **Limit results**: Reduce `limit` for faster responses
4. **Cache results**: Save findings locally

### Combining Tools

1. **Search → Analyze**: Find code, then analyze it
2. **Plan → Generate**: Plan approach, then generate tests
3. **Research → Compare**: Learn from web, find in codebase
4. **Detect → Fix**: Find issues, then refactor

---

## Version & Standards

- **MCP Version**: 2024-11-05
- **Protocol**: JSON-RPC 2.0
- **Transport**: stdio
- **Authentication**: Via workspace path
- **Rate Limiting**: None (local execution)

---

## Troubleshooting

### Tool Not Found
**Problem**: "Tool 'search_database' not found"
**Solution**: Verify Vectora is running in MCP mode: `vectora mcp /workspace`

### Workspace Not Accessible
**Problem**: "Trust folder violation"
**Solution**: Tool can only access files within workspace root. Move files or update workspace path.

### Slow Performance
**Problem**: Tools are taking > 30 seconds
**Solution**:
- Initial embedding takes time
- Try with smaller `limit` parameter
- Check workspace size

### No Results
**Problem**: Search returns empty
**Solution**:
- Files must be indexed first (use `embed` tool)
- Try broader search terms
- Lower `threshold` parameter

---

## See Also

- **Configuration**: See `CLAUDE_CODE_INTEGRATION.md`
- **Architecture**: See `API_ARCHITECTURE.md`
- **Tools Documentation**: See `TOOLS_DOCUMENTATION.md`
- **API Reference**: See `TOOLS_API_REFERENCE.md`
