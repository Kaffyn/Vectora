# Plano de Embedding Tools para Vectora MCP

**Data:** 2026-04-11
**Objetivo:** Criar tools de embedding únicas que exploram a força do Vectora como RAG engine
**Status:** Planning Phase

---

## Contexto Estratégico

### Problema
- Gemini CLI, VS Code Extension, Antigravity já têm suas próprias tools genéricas (read_file, write_file, grep_search, etc)
- Expor as mesmas 10 tools via MCP é redundante
- Não aproveita as forças únicas do Vectora (RAG, embedding, vector search)

### Solução
- **ACP Mode (Vectora Standalone):** Usar as 10 tools genéricas já implementadas
- **MCP Mode (Vectora Sub-Agent):** Expor apenas **tools de embedding/RAG únicas** que agregam valor
- Tools de embedding invocam o **LLM do Vectora** para processar dados

### Fluxo MCP Embedding
```
Gemini CLI (parent agent)
    ↓ (envia requisição JSON-RPC via MCP stdio)
Vectora MCP Server
    ├─ Recebe: {tool: "embed", input: {text: "código"}}
    ├─ Processa com LLM do Vectora (Gemini/Claude/Qwen)
    ├─ Gera embedding
    ├─ Armazena em ChromemDB
    └─ Retorna resultado
    ↓ (resposta JSON-RPC)
Gemini CLI (usa resultado em raciocínio)
```

---

## 1. Core Embedding Tools (4 tools)

### 1.1 `embed` — Embeddar texto/arquivo

**Propósito:** Converter texto em embedding usando o provider LLM do Vectora

**Schema:**
```json
{
  "name": "embed",
  "description": "Convert text/file content into embeddings using Vectora's configured LLM provider.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "content": {
        "type": "string",
        "description": "Text content to embed"
      },
      "metadata": {
        "type": "object",
        "description": "Optional metadata (source, filename, type, etc)",
        "properties": {
          "source": {"type": "string"},
          "filename": {"type": "string"},
          "type": {"type": "string"}
        }
      },
      "workspace_id": {
        "type": "string",
        "description": "Workspace to store embedding (default: global)"
      }
    },
    "required": ["content"]
  }
}
```

**Request Example:**
```json
{
  "tool": "embed",
  "input": {
    "content": "func (e *Engine) Query(ctx context.Context, query string) { ... }",
    "metadata": {
      "source": "engine.go",
      "filename": "engine.go",
      "type": "golang"
    },
    "workspace_id": "vectora-core"
  }
}
```

**Response:**
```json
{
  "output": {
    "chunk_id": "chunk-abc123",
    "embedding": [0.123, 0.456, ...],  // 768 ou 1536 dims
    "metadata": {
      "source": "engine.go",
      "filename": "engine.go",
      "type": "golang"
    },
    "tokens_used": 145,
    "provider": "gemini",
    "stored": true
  },
  "isError": false
}
```

**Implementação:**
```go
type EmbedTool struct {
    Router    *llm.Router
    VecStore  db.VectorStore
    Guardian  *policies.Guardian
}

func (t *EmbedTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
    // 1. Parse arguments
    var input struct {
        Content     string                 `json:"content"`
        Metadata    map[string]interface{} `json:"metadata,omitempty"`
        WorkspaceID string                 `json:"workspace_id,omitempty"`
    }

    // 2. Get LLM provider
    provider := t.Router.GetDefault()

    // 3. Generate embedding
    embedding, err := provider.Embed(ctx, input.Content, "")

    // 4. Store in ChromemDB
    chunkID := uuid.New().String()
    t.VecStore.Store(ctx, collectionID, chunkID, input.Content, embedding, input.Metadata)

    // 5. Return result
    return &ToolResult{
        Output: map[string]interface{}{
            "chunk_id": chunkID,
            "embedding": embedding,
            "metadata": input.Metadata,
            "provider": provider.Name(),
            "stored": true,
        },
    }, nil
}
```

**Guardian Checks:**
- ✅ Metadata validation (no PII exposure)
- ✅ Workspace isolation (tenant-specific storage)
- ✅ Token limits (max 50k tokens per request)

---

### 1.2 `search_database` — Buscar no banco vetorial

**Propósito:** Fazer busca semântica no banco de dados do Vectora (ChromemDB + BBolt)

**Schema:**
```json
{
  "name": "search_database",
  "description": "Search Vectora's vector database using semantic similarity.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Query text to search for"
      },
      "top_k": {
        "type": "integer",
        "description": "Number of results to return (default: 5, max: 20)"
      },
      "workspace_id": {
        "type": "string",
        "description": "Specific workspace to search (default: all)"
      },
      "filters": {
        "type": "object",
        "description": "Metadata filters (e.g., {type: 'golang', source: 'engine.go'})"
      }
    },
    "required": ["query"]
  }
}
```

**Request Example:**
```json
{
  "tool": "search_database",
  "input": {
    "query": "como fazer embedding de um arquivo inteiro",
    "top_k": 10,
    "workspace_id": "vectora-core",
    "filters": {
      "type": "golang",
      "source": "*.go"
    }
  }
}
```

**Response:**
```json
{
  "output": {
    "query": "como fazer embedding de um arquivo inteiro",
    "results": [
      {
        "chunk_id": "chunk-abc123",
        "content": "func (e *Engine) Embed(ctx context.Context, text string, model string) ([]float32, error) { ... }",
        "similarity_score": 0.892,
        "metadata": {
          "source": "engine.go",
          "filename": "engine.go",
          "line_start": 150,
          "line_end": 160
        }
      },
      {
        "chunk_id": "chunk-def456",
        "content": "// EmbedJob represents a background embedding job...",
        "similarity_score": 0.745,
        "metadata": {
          "source": "job.go",
          "filename": "job.go"
        }
      }
    ],
    "count": 2,
    "search_time_ms": 45
  },
  "isError": false
}
```

**Implementação:**
```go
type SearchDatabaseTool struct {
    VecStore db.VectorStore
    Router   *llm.Router
}

func (t *SearchDatabaseTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
    // 1. Parse query
    var input struct {
        Query       string                 `json:"query"`
        TopK        int                    `json:"top_k,omitempty"`
        WorkspaceID string                 `json:"workspace_id,omitempty"`
        Filters     map[string]interface{} `json:"filters,omitempty"`
    }

    // 2. Embed query using LLM
    provider := t.Router.GetDefault()
    queryVec, _ := provider.Embed(ctx, input.Query, "")

    // 3. Search ChromemDB
    chunks := t.VecStore.Query(ctx, collectionID, queryVec, input.TopK)

    // 4. Apply metadata filters if provided
    if input.Filters != nil {
        chunks = filterByMetadata(chunks, input.Filters)
    }

    // 5. Return results
    return &ToolResult{
        Output: map[string]interface{}{
            "query": input.Query,
            "results": chunks,
            "count": len(chunks),
        },
    }, nil
}
```

**Guardian Checks:**
- ✅ Workspace isolation (tenant-specific queries)
- ✅ Result limiting (top_k max 20)
- ✅ Metadata filter validation

---

### 1.3 `web_search_and_embed` — Buscar web e vetorizar end-to-end

**Propósito:** Buscar na web E imediatamente vetorizar resultados

**Schema:**
```json
{
  "name": "web_search_and_embed",
  "description": "Search the web and embed results into Vectora's database in one operation.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Web search query"
      },
      "top_results": {
        "type": "integer",
        "description": "Number of web results to process (default: 5, max: 20)"
      },
      "include_content": {
        "type": "boolean",
        "description": "Fetch full content from pages (slower, default: true)"
      },
      "workspace_id": {
        "type": "string",
        "description": "Store embeddings in this workspace"
      },
      "auto_embed": {
        "type": "boolean",
        "description": "Automatically embed results (default: true)"
      }
    },
    "required": ["query"]
  }
}
```

**Request Example:**
```json
{
  "tool": "web_search_and_embed",
  "input": {
    "query": "golang chromemdb vector store implementation 2026",
    "top_results": 10,
    "include_content": true,
    "workspace_id": "research",
    "auto_embed": true
  }
}
```

**Response:**
```json
{
  "output": {
    "query": "golang chromemdb vector store implementation 2026",
    "results_found": 10,
    "results_embedded": 8,
    "chunks_created": 24,
    "results": [
      {
        "url": "https://example.com/chromemdb-guide",
        "title": "ChromemDB Vector Store Guide",
        "status": "success",
        "chunks_embedded": 5,
        "summary": "Guide on using ChromemDB for vector storage...",
        "chunk_ids": ["chunk-web-1", "chunk-web-2", ...]
      },
      {
        "url": "https://example.com/golang-vectors",
        "title": "Working with Vectors in Go",
        "status": "success",
        "chunks_embedded": 3,
        "summary": "Best practices for vector operations in Go...",
        "chunk_ids": ["chunk-web-4", "chunk-web-5", ...]
      }
    ],
    "errors": [
      {
        "url": "https://blocked-site.com",
        "reason": "Access denied"
      }
    ],
    "total_time_ms": 5420
  },
  "isError": false
}
```

**Implementação (Pseudocódigo):**
```go
type WebSearchAndEmbedTool struct {
    Router      *llm.Router
    VecStore    db.VectorStore
    WebSearcher *web.Searcher  // DuckDuckGo
    WebFetcher  *web.Fetcher
}

func (t *WebSearchAndEmbedTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
    // 1. Web search
    results := t.WebSearcher.Search(input.Query, input.TopResults)

    // 2. Para cada resultado:
    for _, result := range results {
        // 3. Fetch full content
        if input.IncludeContent {
            content := t.WebFetcher.Fetch(result.URL)
        } else {
            content = result.Snippet
        }

        // 4. Chunk content (split into smaller pieces)
        chunks := chunkText(content, maxChunkSize)

        // 5. Para cada chunk:
        for _, chunk := range chunks {
            // 6. Embed usando LLM
            provider := t.Router.GetDefault()
            vec, _ := provider.Embed(ctx, chunk, "")

            // 7. Store em ChromemDB
            metadata := map[string]interface{}{
                "url": result.URL,
                "title": result.Title,
                "source": "web_search",
            }
            t.VecStore.Store(ctx, collectionID, chunkID, chunk, vec, metadata)
        }
    }

    return &ToolResult{Output: results}, nil
}
```

**Guardian Checks:**
- ✅ URL whitelist/blacklist (no scraping private sites)
- ✅ Rate limiting (max 20 results)
- ✅ Content validation (no malicious scripts)
- ✅ Data retention policy (mark as web_search origin)

---

### 1.4 `web_fetch_and_embed` — Navegar links internos e vetorizar

**Propósito:** Buscar uma URL, navegar links INTERNOS se necessário, vetorizar tudo end-to-end

**Schema:**
```json
{
  "name": "web_fetch_and_embed",
  "description": "Fetch a URL, optionally navigate internal links, and embed all content.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "url": {
        "type": "string",
        "description": "Starting URL to fetch"
      },
      "follow_internal_links": {
        "type": "boolean",
        "description": "Follow internal links up to max_depth (default: true)"
      },
      "max_depth": {
        "type": "integer",
        "description": "Maximum depth for internal link following (default: 2, max: 5)"
      },
      "max_pages": {
        "type": "integer",
        "description": "Maximum pages to fetch (default: 10, max: 50)"
      },
      "selector": {
        "type": "string",
        "description": "CSS selector for content extraction (default: 'main, article, .content')"
      },
      "workspace_id": {
        "type": "string",
        "description": "Store embeddings in this workspace"
      }
    },
    "required": ["url"]
  }
}
```

**Request Example:**
```json
{
  "tool": "web_fetch_and_embed",
  "input": {
    "url": "https://docs.example.com/guides/",
    "follow_internal_links": true,
    "max_depth": 3,
    "max_pages": 25,
    "selector": "main, article, .documentation",
    "workspace_id": "documentation"
  }
}
```

**Response:**
```json
{
  "output": {
    "start_url": "https://docs.example.com/guides/",
    "pages_fetched": 15,
    "pages_failed": 2,
    "chunks_embedded": 58,
    "pages": [
      {
        "url": "https://docs.example.com/guides/intro",
        "title": "Introduction to Vectora",
        "status": "success",
        "chunks": 4,
        "chunk_ids": ["chunk-doc-1", "chunk-doc-2", ...],
        "content_length": 2450,
        "extraction_method": "selector"
      },
      {
        "url": "https://docs.example.com/guides/embedding",
        "title": "Embedding Guide",
        "status": "success",
        "chunks": 5,
        "chunk_ids": ["chunk-doc-6", ...],
        "content_length": 3120
      }
    ],
    "failed_pages": [
      {
        "url": "https://docs.example.com/api-reference",
        "error": "Access denied (robots.txt)"
      }
    ],
    "sitemap_respect": true,
    "robots_txt_respected": true,
    "total_content_embedded": "145 KB",
    "total_time_ms": 12450
  },
  "isError": false
}
```

**Implementação (Pseudocódigo):**
```go
type WebFetchAndEmbedTool struct {
    WebFetcher *web.Fetcher
    Router     *llm.Router
    VecStore   db.VectorStore
}

func (t *WebFetchAndEmbedTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
    // 1. BFS crawl (respeitando depth e max_pages)
    crawler := web.NewCrawler(input.URL, input.MaxDepth, input.MaxPages)
    urls := crawler.Crawl()  // Respeita robots.txt, sitemap

    // 2. Fetch cada página
    for _, url := range urls {
        content := t.WebFetcher.Fetch(url)

        // 3. Extract com selector
        extracted := t.WebFetcher.ExtractBySelector(content, input.Selector)

        // 4. Chunk e embed
        for _, chunk := range chunkText(extracted) {
            vec, _ := t.Router.GetDefault().Embed(ctx, chunk, "")
            t.VecStore.Store(ctx, collectionID, chunkID, chunk, vec, metadata)
        }
    }

    return &ToolResult{Output: results}, nil
}
```

**Guardian Checks:**
- ✅ robots.txt compliance
- ✅ sitemap.xml respect
- ✅ Rate limiting (max 50 pages)
- ✅ Domain validation (same domain only)
- ✅ Content size limits (max 1MB per page)

---

## 2. Planning & Implementation Tools (2 tools)

### 2.1 `plan_mode` — Criar/atualizar plano de implementação

**Propósito:** Criar ou atualizar um plano de implementação usando embedding para maior precisão contextual

**Schema:**
```json
{
  "name": "plan_mode",
  "description": "Create or update implementation plans using Vectora's RAG for context precision.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "action": {
        "type": "string",
        "enum": ["create", "update", "refine"],
        "description": "create=new plan, update=modify existing, refine=improve with search"
      },
      "title": {
        "type": "string",
        "description": "Plan title/objective"
      },
      "description": {
        "type": "string",
        "description": "Detailed description of what needs to be implemented"
      },
      "context_queries": {
        "type": "array",
        "description": "Search queries for context gathering",
        "items": {"type": "string"}
      },
      "plan_id": {
        "type": "string",
        "description": "Existing plan ID (for update/refine)"
      },
      "workspace_id": {
        "type": "string",
        "description": "Store plan in workspace"
      }
    },
    "required": ["action", "title"]
  }
}
```

**Request Example:**
```json
{
  "tool": "plan_mode",
  "input": {
    "action": "create",
    "title": "Implementar GraphQL API para Vectora",
    "description": "Criar um GraphQL endpoint que permita queries complexas no banco de dados...",
    "context_queries": [
      "GraphQL implementação em Go",
      "Vector search GraphQL schema",
      "Como integrar GraphQL com REST API"
    ],
    "workspace_id": "vectora-core"
  }
}
```

**Response:**
```json
{
  "output": {
    "plan_id": "plan-graphql-001",
    "title": "Implementar GraphQL API para Vectora",
    "status": "created",
    "phases": [
      {
        "phase": 1,
        "title": "Setup GraphQL Schema",
        "tasks": [
          "Define Query types for vector search",
          "Create Mutation types for data management",
          "Setup GraphQL validation"
        ],
        "estimated_hours": 8,
        "dependencies": []
      },
      {
        "phase": 2,
        "title": "Implement Resolvers",
        "tasks": [
          "Implement search_database resolver",
          "Implement embed resolver",
          "Add error handling"
        ],
        "estimated_hours": 12,
        "dependencies": [1]
      }
    ],
    "context_gathered": 3,
    "relevant_chunks": [
      "chunk-graphql-best-practices",
      "chunk-vector-search-patterns"
    ],
    "created_at": "2026-04-11T15:30:00Z"
  },
  "isError": false
}
```

**Implementação:**
```go
type PlanModeTool struct {
    Router      *llm.Router
    VecStore    db.VectorStore
    KVStore     db.KVStore
    MessageSvc  *llm.MessageService
}

func (t *PlanModeTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
    // 1. Gather context via search_database
    for _, query := range input.ContextQueries {
        // Search usando embedding
        chunks := t.VecStore.Query(ctx, collectionID,
            t.Router.GetDefault().Embed(ctx, query, ""), 10)
        contextChunks = append(contextChunks, chunks...)
    }

    // 2. Montar prompt com contexto
    prompt := fmt.Sprintf(`
        Crie um plano detalhado para: %s

        Contexto relevante da base:
        %s

        Estruture em fases, tarefas, estimativas...
    `, input.Description, contextChunks)

    // 3. Chamar LLM com contexto
    resp, _ := t.Router.GetDefault().Complete(ctx, llm.CompletionRequest{
        Messages: []llm.Message{
            {Role: "user", Content: prompt},
        },
    })

    // 4. Parse resposta e armazenar
    plan := parsePlan(resp.Content)
    t.KVStore.Set(ctx, "plans", plan.ID, plan)

    return &ToolResult{Output: plan}, nil
}
```

---

### 2.2 `refactor_with_context` — Refatorar código usando contexto

**Propósito:** Refatorar código usando similiaridade semântica para encontrar padrões

**Schema:**
```json
{
  "name": "refactor_with_context",
  "description": "Refactor code using semantic similarity to find patterns and best practices.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "code": {
        "type": "string",
        "description": "Code to refactor"
      },
      "language": {
        "type": "string",
        "description": "Programming language (go, python, typescript, etc)"
      },
      "focus": {
        "type": "string",
        "description": "Refactoring focus (performance, readability, security, testing)"
      },
      "find_similar": {
        "type": "boolean",
        "description": "Search database for similar patterns (default: true)"
      },
      "workspace_id": {
        "type": "string"
      }
    },
    "required": ["code", "language"]
  }
}
```

**Request Example:**
```json
{
  "tool": "refactor_with_context",
  "input": {
    "code": "func (e *Engine) Query(ctx context.Context, q string) (string, error) {\n  // 50 lines of code...\n}",
    "language": "go",
    "focus": "performance",
    "find_similar": true,
    "workspace_id": "vectora-core"
  }
}
```

**Response:**
```json
{
  "output": {
    "status": "success",
    "original_code_length": 450,
    "refactored_code": "func (e *Engine) Query(ctx context.Context, q string) (string, error) {\n  // Refactored version...\n}",
    "refactored_code_length": 380,
    "improvements": [
      "Extracted embedding logic to separate function",
      "Added caching for frequent queries",
      "Improved error handling"
    ],
    "similar_patterns_found": 3,
    "pattern_examples": [
      {
        "source": "chunk-engine-caching",
        "pattern": "Query result caching with TTL"
      }
    ]
  },
  "isError": false
}
```

---

## 3. Analysis & Insights Tools (3 tools)

### 3.1 `analyze_code_patterns` — Analisar padrões no código

**Propósito:** Analisar padrões de código em toda codebase usando embedding

**Schema:**
```json
{
  "name": "analyze_code_patterns",
  "description": "Analyze code patterns across the codebase using semantic similarity.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "pattern_type": {
        "type": "string",
        "enum": ["antipatterns", "best_practices", "security_issues", "performance_issues"],
        "description": "Type of pattern to analyze"
      },
      "language": {
        "type": "string",
        "description": "Programming language to focus on"
      },
      "top_results": {
        "type": "integer",
        "description": "Top patterns to return (default: 10)"
      },
      "workspace_id": {
        "type": "string"
      }
    },
    "required": ["pattern_type"]
  }
}
```

**Response:**
```json
{
  "output": {
    "pattern_type": "security_issues",
    "patterns_found": 5,
    "patterns": [
      {
        "name": "SQL Injection Risk",
        "occurrences": 3,
        "severity": "high",
        "examples": [
          {"file": "database.go:123", "code": "SELECT * FROM users WHERE id = " + userInput}
        ],
        "recommendation": "Use parameterized queries"
      }
    ]
  }
}
```

---

### 3.2 `knowledge_graph_analysis` — Construir grafo de conhecimento

**Propósito:** Construir um grafo de conhecimento a partir dos embeddings armazenados

**Schema:**
```json
{
  "name": "knowledge_graph_analysis",
  "description": "Build and analyze a knowledge graph from embedded documents.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "topic": {
        "type": "string",
        "description": "Topic to build knowledge graph for"
      },
      "depth": {
        "type": "integer",
        "description": "Graph depth (1-3, default: 2)"
      },
      "workspace_id": {
        "type": "string"
      }
    },
    "required": ["topic"]
  }
}
```

**Response:**
```json
{
  "output": {
    "topic": "Vector Databases",
    "nodes": [
      {"id": "chromemdb", "label": "ChromemDB", "type": "technology"},
      {"id": "embedding", "label": "Embeddings", "type": "concept"}
    ],
    "edges": [
      {"from": "chromemdb", "to": "embedding", "relation": "uses"}
    ],
    "graph_json": "..."
  }
}
```

---

### 3.3 `doc_coverage_analysis` — Analisar cobertura de documentação

**Propósito:** Analisar quais partes da codebase têm documentação

**Schema:**
```json
{
  "name": "doc_coverage_analysis",
  "description": "Analyze documentation coverage of the codebase.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "language": {
        "type": "string",
        "description": "Programming language"
      },
      "min_complexity": {
        "type": "integer",
        "description": "Minimum function complexity to check (default: 1)"
      }
    },
    "required": ["language"]
  }
}
```

---

## 4. Quality & Testing Tools (2 tools)

### 4.1 `test_generation` — Gerar testes baseado em contexto

### 4.2 `bug_pattern_detection` — Detectar padrões de bugs

---

## Resumo: 11 Embedding Tools Propostas

| # | Tool | Categoria | Propósito | Impacto |
|---|------|-----------|----------|--------|
| 1 | `embed` | Core | Converter texto em embedding | Alto |
| 2 | `search_database` | Core | Buscar semântica no banco | Alto |
| 3 | `web_search_and_embed` | Core | Web search + vetorização end-to-end | Alto |
| 4 | `web_fetch_and_embed` | Core | Navegar links + vetorização | Alto |
| 5 | `plan_mode` | Planning | Criar planos com contexto | Médio |
| 6 | `refactor_with_context` | Planning | Refatorar com padrões | Médio |
| 7 | `analyze_code_patterns` | Analysis | Encontrar antipatterns | Médio |
| 8 | `knowledge_graph_analysis` | Analysis | Construir grafo de conhecimento | Médio |
| 9 | `doc_coverage_analysis` | Analysis | Cobertura de docs | Baixo |
| 10 | `test_generation` | Quality | Gerar testes | Médio |
| 11 | `bug_pattern_detection` | Quality | Detectar bugs | Médio |

---

## Fase de Implementação

### Phase 4G: Core Embedding Tools (2026-04-15)
- [ ] `embed` - embedding básico
- [ ] `search_database` - busca vetorial
- [ ] `web_search_and_embed` - busca web integrada
- [ ] `web_fetch_and_embed` - crawling com embedding

### Phase 4H: Planning Tools (2026-04-18)
- [ ] `plan_mode` - criação de planos
- [ ] `refactor_with_context` - refatoração

### Phase 4I: Analysis Tools (2026-04-22)
- [ ] `analyze_code_patterns`
- [ ] `knowledge_graph_analysis`
- [ ] `doc_coverage_analysis`

### Phase 4J: Quality Tools (2026-04-25)
- [ ] `test_generation`
- [ ] `bug_pattern_detection`

---

## Integração com Router

**Em `core/api/ipc/router.go` e novo `core/api/mcp/tools.go`:**

```go
// Registrar embedding tools
func RegisterEmbeddingTools(mcpServer *mcp.StdioServer, engine *engine.Engine) {
    mcpServer.RegisterTool("embed", NewEmbedTool(engine.LLM, engine.Storage))
    mcpServer.RegisterTool("search_database", NewSearchDatabaseTool(engine.Storage))
    mcpServer.RegisterTool("web_search_and_embed", NewWebSearchAndEmbedTool(...))
    mcpServer.RegisterTool("web_fetch_and_embed", NewWebFetchAndEmbedTool(...))
    mcpServer.RegisterTool("plan_mode", NewPlanModeTool(...))
    // ... etc
}
```

---

## Diferenças Estratégicas

| Scenario | Tools | Transporte |
|----------|-------|-----------|
| **Vectora Standalone (ACP Mode)** | 10 tools genéricas (read_file, write_file, etc) | stdio (ACP SDK) |
| **Vectora como MCP (Sub-Agent)** | 11 embedding tools únicos | stdio (MCP JSON-RPC) |
| **Internal Comm (CLI ↔ Core)** | 14 métodos IPC (workspace.query, chat.*, etc) | named pipes |

**Resultado:** Gemini CLI não precisa expor tools duplicadas, só aproveita as únicas capacidades do Vectora (RAG + embedding).

---

**Status:** Ready for Phase 4G implementation
**Owner:** Claude + Bruno
**Updated:** 2026-04-11
