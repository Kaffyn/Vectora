# MVP Scope — Vectora v0.1.0

Escopo definitivo para o lançamento oficial do Vectora v0.1.0.

**Versão:** 0.1.0
**Status:** 🟡 Feature-complete, polimento final

---

## Core Features

### Agent Architecture (LangGraph)

O coração do Vectora é um grafo de supervisor + workers especializados compilado pelo LangGraph:

```
START
  └─► supervisor (classifica intenção via regex + LLM fallback)
        ├─► direct    ──► direct_tools (memory) ──► direct ──► END
        ├─► search    ──► search_tools ──► process_retrieval ──► search ──► END
        ├─► coder     ──► coder_tools (fs + memory) ──► coder ──► END
        └─► rag_subgraph ─────────────────────────────────── ► direct ──► END
```

**Supervisor** (`agents/supervisor.py`):

- ✅ `classify_intent()` — regex patterns para direct/coder/search/rag + fallback por palavra-chave
- ✅ Routing via `Command(goto=...)` do LangGraph
- ✅ Self-awareness via `VECTORA_IDENTITY` (`agents/_identity.py`)

**Direct Agent** (`agents/direct.py`):

- ✅ Respostas diretas, síntese pós-RAG, gerenciamento de memória
- ✅ Ferramentas: `save_memory`, `get_memory`, `delete_memory`

**Search Agent** (`agents/search.py`):

- ✅ Pesquisa web em tempo real + busca vetorial
- ✅ Ferramentas: `web_search`, `fetch_url`, `vector_search` (+ `embedding` se RAG habilitado)
- ✅ Cascading automático: resultados de busca são enfileirados para embedding no LanceDB

**Coder Agent** (`agents/coder.py`):

- ✅ Operações de filesystem, terminal, geração de código
- ✅ Ferramentas: `file_read`, `file_edit`, `file_write`, `grep`, `list_dir`, `terminal`, + memória

### RAG Subgraph (`nodes/rag_subgraph.py`)

Pipeline de recuperação com threshold adaptativo antes de cada síntese:

```
rag_retrieve (vector_search)
  └─► rag_decide (score threshold)
        ├─► rag_inject    (score ≥ 0.7 — alta confiança, injeta diretamente)
        ├─► rag_rerank    (score 0.4–0.7 — reranking Cohere antes de injetar)
        └─► rag_websearch (score < 0.4 — fallback web + auto-embed dos resultados)
```

- ✅ `rag_retrieve` — busca semântica via `vector_search`
- ✅ `rag_decide` — threshold adaptativo (0.4 / 0.7) para escolher o próximo nó
- ✅ `rag_rerank` — reranking com CohereRerank, top-3 docs
- ✅ `rag_websearch` — busca web + enfileira resultados para embedding (cascading)
- ✅ `rag_inject` — injeta docs como `SystemMessage` no contexto antes do `direct`

### Process Retrieval (`nodes/engine.py`)

- ✅ Detecta `ToolMessages` de `web_search` / `fetch_url` no histórico
- ✅ Enfileira automaticamente para embedding no LanceDB (fire-and-forget)
- ✅ Rastreia `pending_embeds` no `State` para observabilidade

### 14 Ferramentas Implementadas

Todas com `try/except` defensivo, logging estruturado e timeout individual:

**Web (2)**

- ✅ `web_search` — Busca real-time via Tavily
- ✅ `fetch_url` — Extração de conteúdo via Tavily Extract API

**RAG (3)**

- ✅ `vector_search` — Busca semântica em LanceDB com reranking Cohere
- ✅ `embedding` — Indexação assíncrona (fire-and-forget, queue com retry)
- ✅ `ingest_docs` — Ingestão em lote com glob pattern + respeita `.gitignore`

**Filesystem (6)**

- ✅ `file_read` — Leitura com validação de path
- ✅ `file_edit` — Substituição cirúrgica (`replace_all` para múltiplas ocorrências)
- ✅ `file_write` — Criar/sobrescrever arquivo completo
- ✅ `grep` — Busca regex com proteção anti-ReDoS
- ✅ `list_dir` — Listagem recursiva com filtros
- ✅ `terminal` — Execução shell async, whitelist de segurança, timeout 30s

**Memory (3)**

- ✅ `save_memory` — Persistir memória cross-session em SQLite com TTL opcional
- ✅ `get_memory` — Recuperar memórias por chave ou listar todas
- ✅ `delete_memory` — Remoção de memórias específicas

**MCP (1)**

- ✅ `call_mcp_tool` — Invocar ferramenta de outro servidor MCP

### Persistência Local (File-Based)

Zero infraestrutura externa — tudo em `~/.vectora/`:

- ✅ **SQLite** (`aiosqlite`) — Histórico de conversas, checkpoints LangGraph, memórias persistentes
- ✅ **LanceDB** — Vector store file-based para RAG semântico
- ✅ **Embedding Queue** — Fila assíncrona com SQLAlchemy + retry exponencial + DLQ
- ✅ **Reconciliation** — Background worker recupera jobs travados após crash

### MCP Server

- ✅ **13 tools expostas** via `@mcp.tool()` com descrições otimizadas para LLM selection
- ✅ **4 resources expostos** — `thread/context`, `thread/history`, `status`, `collections`
- ✅ **Transport stdio** — Para uso local com Claude Code / Claude Desktop
- ✅ **Transport SSE** — Para uso multi-agent (Paperclip, múltiplas instâncias)
- ✅ **`delegate_task_to_vectora`** — Ferramenta A2A que executa o grafo completo como sub-agente
- ✅ **Singleton AgentManager** — Evita reinicialização cara do LanceDB por chamada

### CLI & Interface

- ✅ **Terminal TUI** com Rich (Panels, Tables, Layout, Live)
- ✅ **Prompt multiline** — `Alt+Enter` / `Shift+Enter` para quebra de linha
- ✅ **Setup Wizard** — Seleção de provider, input seguro de API key, teste de conexão
- ✅ **Visual feedback colorido** — Tool calls (amarelo), responses (vermelho), terminal (verde)
- ✅ **Debug Mode** — `/debug` toggle, persiste em `~/.vectora/chat_config.json`
- ✅ **Session Management** — `/new`, `/sessions`, `/session <id>`
- ✅ **Model Switching** — `/model` em tempo real (sem restart)

### Configuração & Observabilidade

- ✅ **Pydantic Settings** — Single Source of Truth em `vectora/config/settings.py`
- ✅ **3-level hierarchy** — `defaults.env` → `.env` local → `~/.vectora/.env`
- ✅ **Logs estruturados** em JSON Lines (`~/.vectora/logs/vectora.jsonl`), rotating 10 MB
- ✅ **LangSmith integration** — Tracing opcional via `LANGSMITH_API_KEY`
- ✅ **Feature flags** — `ENABLE_RAG`, `ENABLE_FILE_OPERATIONS` por ambiente

### Testing

- ✅ **KISS 1:1 pattern** — 1 arquivo de teste por arquivo fonte
- ✅ **197 testes** passando (unit + integration)
- ✅ **≥70% coverage** em todos os arquivos com teste dedicado
- ✅ **pytest-asyncio** com `asyncio_mode = "auto"`
- ✅ **ruff** — 0 erros de lint
- ✅ **ty** — 81 erros restantes são incompatibilidade LangGraph/ty + stubs ausentes de terceiros (não acionáveis)

### CI/CD & Deployment

- ✅ **GitHub Actions** (`runner.yml`) — Lint, type check, test, build
- ✅ **Dockerfile** + **docker-compose.yml** para desenvolvimento e produção
- ✅ **Pre-commit hooks** — ruff lint, ruff format, Prettier (markdown), uv-lock

---

## Estrutura de Arquivos

```
vectora/
├── agent.py               # AgentManager (orchestrator, DI hub)
├── graph.py               # build_graph() — supervisor + workers + RAG subgraph
├── state.py               # TypedDict State (messages, routing_decision, rag_*, etc.)
├── context.py             # Context schema (user_type, thread_id)
├── main.py                # CLI entry point
├── version.py             # Dynamic version from pyproject.toml
├── agents/
│   ├── _identity.py       # VECTORA_IDENTITY constant (self-awareness)
│   ├── supervisor.py      # classify_intent() + supervisor node
│   ├── direct.py          # Direct agent (chat, synthesis, memory)
│   ├── search.py          # Search agent (web, RAG)
│   └── coder.py           # Coder agent (files, terminal)
├── config/
│   ├── settings.py        # Pydantic Settings (single source of truth)
│   └── defaults.env       # Embedded defaults
├── nodes/
│   ├── base.py            # invoke_llm(), sanitize_for_gemini(), build_messages()
│   ├── debug.py           # DiagnosticToolNode
│   ├── engine.py          # process_retrieval (cascading web→LanceDB)
│   ├── rag_subgraph.py    # build_rag_subgraph() — retrieve→decide→rerank/web→inject
│   ├── retrieval.py       # retrieval_node() + _rerank()
│   └── tools.py           # SEARCH_TOOLS, FS_TOOLS, MEMORY_TOOLS, RAG_TOOLS
├── tools/
│   ├── fs.py              # file_read, file_edit, file_write, grep, list_dir, terminal
│   ├── rag.py             # embedding, vector_search, ingest_docs
│   ├── web.py             # web_search, fetch_url
│   ├── memory.py          # save_memory, get_memory, delete_memory
│   └── mcp.py             # call_mcp_tool
├── mcp/
│   ├── server.py          # MCP Server (FastMCP, 13 tools, 4 resources)
│   ├── client.py          # MCP Client (consumir MCPs externos)
│   └── proxy.py           # VectoraProxy (cliente oficial para Paperclip)
├── services/
│   ├── background.py      # Embedding background worker (retry + DLQ)
│   ├── checkpoint.py      # LangGraph SQLite checkpointer
│   ├── embedding.py       # EmbeddingService (Cohere)
│   ├── memory.py          # MemoryStore (SQLite cross-session)
│   ├── queue.py           # Embedding queue (SQLAlchemy async + WAL mode)
│   ├── security.py        # Whitelist, path validation, ReDoS protection
│   ├── session.py         # Session lifecycle
│   ├── setup_wizard.py    # Interactive first-run wizard
│   ├── telemetry.py       # Structured logging, audit trails, debug dumps
│   └── text.py            # TextService (split, count_tokens)
├── ui/
│   ├── chat.py            # Chat TUI (Rich Live, prompt-toolkit)
│   ├── commands.py        # Command dispatcher (/quit, /debug, /model, ...)
│   └── main.py            # Rich components (panels, layouts, widgets)
└── testing/
    ├── fixtures.py        # Pytest fixtures (test_graph, checkpointer, etc.)
    ├── mocks.py           # MockLLM, MockToolNode
    └── message_factory.py # Test message builders
```

---

## Checklist de Release

### Core

- [x] 14 ferramentas implementadas e testadas
- [x] Supervisor + 3 workers especializados compilados no LangGraph
- [x] RAG subgraph com threshold adaptativo (inject/rerank/websearch)
- [x] Cascading embeddings (web → LanceDB fire-and-forget)
- [x] Persistência SQLite + LanceDB funcional
- [x] MCP Server (stdio + SSE) operacional
- [x] Memory cross-session funcional
- [x] Protocolo Paperclip documentado e validado

### Qualidade

- [x] `ruff check vectora/` — 0 erros
- [x] `ty check vectora/` — 81 erros restantes são não-acionáveis (stubs de terceiros)
- [x] `pytest tests/ --cov=vectora` — ≥70% coverage em arquivos testados, 197 testes
- [x] Pre-commit hooks passando
- [ ] Release notes escritas
- [ ] Git tag `v0.1.0` criada

### Publicação

- [ ] `uv build` — wheel gerado sem erros
- [ ] GHCR push — `ghcr.io/kaffyn/vectora:0.1.0`
- [ ] PyPI publish — `uv publish`
- [ ] GitHub Release criada com changelog
