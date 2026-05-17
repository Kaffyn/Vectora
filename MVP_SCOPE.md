# MVP Scope — Vectora v0.1.0

Escopo definitivo para o lançamento oficial do Vectora v0.1.0. Tudo listado aqui está implementado ou é bloqueador para o release.

**Versão:** 0.1.0
**Status:** 🟡 Feature-complete, polimento final

---

## Core Features

### Agent & Reasoning (LangGraph)

O coração do Vectora é um grafo de estado com 4 nós, compilado pelo LangGraph:

- ✅ `call_llm` — Invoca o LLM com histórico deslizante (sliding window) e system prompt adaptativo
- ✅ `tools` — Executa ferramentas em paralelo com `DiagnosticToolNode` (logging detalhado)
- ✅ `process_retrieval` — Pós-processamento de resultados RAG (reranking, filtragem)
- ✅ `sub_node` — Workflows complexos em sub-instância isolada (delegação A2A)

Comportamentos do agente:

- ✅ Histórico deslizante com `trim_messages` e fallback anti-vazio (fix `ValueError: contents are required`)
- ✅ Auto-summarização de histórico longo
- ✅ Auto-retry inteligente (até 3x) em falhas de ferramentas
- ✅ `max_context_tokens: 8000` (configurável via settings)

### 14 Ferramentas Implementadas

Todas com `try/except` defensivo, logging estruturado e timeout individual:

**Web (2)**

- ✅ `web_search` — Busca real-time via Tavily
- ✅ `fetch_url` — Extração de conteúdo via Tavily Extract API

**RAG (3)**

- ✅ `vector_search` — Busca semântica em LanceDB com reranking Voyage AI
- ✅ `embedding` — Indexação assíncrona (fire-and-forget, queue com retry)
- ✅ `ingest_docs` — Ingestão em lote de pastas inteiras (glob pattern)

**Filesystem (6)**

- ✅ `file_read` — Leitura com validação de path
- ✅ `file_edit` — Substituição cirúrgica (`replace_all` para múltiplas ocorrências)
- ✅ `file_write` — Criar/sobrescrever arquivo completo
- ✅ `grep` — Busca regex com proteção anti-ReDoS
- ✅ `list_dir` — Listagem recursiva com filtros
- ✅ `terminal` — Execução shell async (`asyncio.create_subprocess_shell`), whitelist de segurança, timeout 30s

**Memory (3)**

- ✅ `save_memory` — Persistir memória cross-session em SQLite com TTL opcional
- ✅ `get_memory` — Recuperar memórias por chave ou listar todas
- ✅ `delete_memory` — Remoção de memórias específicas

**MCP (1)**

- ✅ `call_mcp_tool` — Invocar ferramenta de outro servidor MCP

### Persistência Local (File-Based)

Zero infraestrutura externa — tudo em `~/.vectora/`:

- ✅ **SQLite** (`aiosqlite`) — Histórico de conversas, checkpoints LangGraph, memórias persistentes
- ✅ **LanceDB** — Vector store file-based para RAG semântico de alta performance
- ✅ **Embedding Queue** — Fila assíncrona com SQLAlchemy (`sqlite+aiosqlite://`) e retry exponencial
- ✅ **Roaming Profile Pattern** — Dados em `~/.vectora/` independente do diretório de instalação

### MCP Server (Vectora como Sub-Agente)

Vectora roda como servidor MCP para que Claude Code, Claude Desktop e Paperclip consumam suas capacidades:

- ✅ **13 tools expostas** via `@mcp.tool()` com descrições otimizadas para LLM selection
- ✅ **4 resources expostos** — `thread/context`, `thread/history`, `status`, `collections`
- ✅ **Transport stdio** — Para uso local com Claude Code / Claude Desktop
- ✅ **Transport SSE** — Para uso multi-agent (Paperclip, múltiplas instâncias)
- ✅ **`delegate_task_to_vectora`** — Ferramenta A2A que roda o LangGraph completo como sub-agente
- ✅ **Rich feedback no stderr** — Panel colorido ao iniciar (`✓ Vectora MCP Server pronto`)
- ✅ **Timeouts por ferramenta** — Proteção em camadas (10s–120s por tool, 300s global para A2A)
- ✅ **Singleton AgentManager** — Evita reinicialização cara do LanceDB por chamada

### Integração Multi-Agent (Paperclip)

Protocolo formal de integração para múltiplos agentes Paperclip compartilharem um único Vectora:

- ✅ **Hub centralizado** — Um processo Vectora serve N agentes Paperclip via `thread_id`
- ✅ **Isolamento por thread_id** — Sessões isoladas via LangGraph Checkpointer (sem vazamento de contexto)
- ✅ **Transport SSE** — HTTP/SSE para comunicação multi-container
- ✅ **`VectoraProxy`** — Cliente async oficial em `vectora/mcp/proxy.py`
- ✅ **Protocolo documentado** — `integrations/paperclip/@AGENTS.md` (v1.0.0)
- ✅ **Modo `stdio` e `sse`** — Seleção via `MCP_TRANSPORT` env var
- ✅ **Docker Compose** — `docker-compose.yml` para subir hub multi-agent

### CLI & Interface

- ✅ **Terminal TUI** com Rich (Panels, Tables, Layout, Live)
- ✅ **Prompt multiline** — `Alt+Enter` / `Shift+Enter` para quebra de linha
- ✅ **Setup Wizard** — Seleção de provider, input seguro de API key, teste de conexão
- ✅ **Visual feedback colorido** — Tool calls (amarelo), tool responses (vermelho), terminal (verde)
- ✅ **Debug Mode** — `/debug` toggle, persiste em `~/.vectora/chat_config.json`
- ✅ **Session Management** — `/new`, `/sessions`, `/session <id>`
- ✅ **Model Switching** — `/model` em tempo real (sem restart)
- ✅ **Welcome Screen** com todos os comandos disponíveis

### Configuração & Observabilidade

- ✅ **Pydantic Settings** — Single Source of Truth em `vectora/config/settings.py`
- ✅ **3-level hierarchy** — `defaults.env` → `.env` local → `~/.vectora/.env` (precedência crescente)
- ✅ **Logs estruturados** em JSON Lines (`~/.vectora/logs/vectora.log`)
- ✅ **LangSmith integration** — Tracing opcional via `LANGSMITH_API_KEY`
- ✅ **System prompt multilíngue** — Auto-detecção de idioma do SO
- ✅ **Auto-detecção de LLM** — Detecta provider disponível pelas API keys presentes

### Testing

- ✅ **Unit tests** — Ferramentas, checkpointer, config, memória, prompts
- ✅ **Integration tests** — RAG pipeline, graph execution, A2A, message flow
- ✅ **E2E tests** — Fire-and-forget, MCP resources, run commands
- ✅ **Stress tests** — Concorrência e paralelismo
- ✅ **>80% coverage** (pytest-cov)
- ✅ **pytest-asyncio** com `asyncio_mode = "auto"`

### CI/CD & Deployment

- ✅ **GitHub Actions** (`runner.yml`) — Lint, type check, test, build
- ✅ **Dockerfile** + **docker-compose.yml** para desenvolvimento e produção
- ✅ **Pre-commit hooks** — Ruff, Mypy, Prettier (markdown), Bandit
- ✅ **Conventional Commits** enforced via AGENTS.md

---

## Melhorias de Qualidade de Vida (Pré-Lançamento)

Adicionadas ao escopo após feature-freeze para garantir uma experiência polida no launch:

- ✅ **`file_write` tool** — Criar/sobrescrever arquivos inteiros (antes só existia `file_edit`)
- ✅ **`file_edit` com `replace_all`** — Suporte a substituição de múltiplas ocorrências
- ✅ **Memory tools** — `save_memory`, `get_memory`, `delete_memory` para contexto cross-session
- ✅ **Terminal async** — Migrado de `subprocess.run` (síncrono, bloqueia UI) para `asyncio.create_subprocess_shell`
- ✅ **Rich panels para tools** — Feedback visual colorido durante execução de ferramentas
- ✅ **Fix `ValueError: contents are required`** — Fallback no `trim_messages` quando histórico fica vazio
- ✅ **Fix embedding_queue_dsn** — Formato SQLAlchemy correto (`sqlite+aiosqlite:///path`)
- ✅ **MCP SSE transport** — Suporte a múltiplos agentes concorrentes (modo Paperclip)
- ✅ **Protocolo Paperclip** — Documentação formal de integração multi-agent
- ✅ **Singleton AgentManager no MCP Server** — Evita re-inicialização em cada delegação A2A
- ✅ **Multiline input** — Sem dependência de `EditingMode.EMACS` (removida)

---

## Fora do MVP (Pós-Lançamento)

- ❌ PostgreSQL / Qdrant Cloud (infra escalável)
- ❌ Streaming de respostas (SSE token-by-token)
- ❌ `thread_id: str` nativo no LangGraph (workaround atual: `hash & 0xFFFFFFFF`)
- ❌ Human-in-the-loop (`interrupt_before` em ações destrutivas)
- ❌ SSE heartbeat para manter conexões longas
- ❌ Dashboard CLI multi-sessão
- ❌ Plugin oficial do Paperclip
- ❌ VSCode Extension / Gemini CLI Plugin
- ❌ Vectora Asset Library (buckets pré-treinados)
- ❌ LangSmith trace com `client_thread_id` nos metadados (observabilidade multi-agent)

---

## Estrutura de Arquivos

```
vectora/
├── agent.py               # AgentManager (orchestrator, DI hub)
├── graph.py               # LangGraph builder (3-node pattern)
├── state.py               # TypedDict State definition
├── context.py             # Context schema (injetado via configurable)
├── prompts.py             # System prompts, language detection
├── main.py                # CLI entry point
├── version.py             # Dynamic version from pyproject.toml
├── config/
│   ├── settings.py        # Pydantic Settings (single source of truth)
│   └── defaults.env       # Embedded defaults
├── nodes/
│   ├── engine.py          # call_llm, process_retrieval, handle_sub_node
│   └── debug.py           # DiagnosticToolNode, call_llm_debug
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
│   ├── background.py      # Embedding background worker
│   ├── checkpoint.py      # LangGraph SQLite checkpointer
│   ├── embedding.py       # EmbeddingService (Voyage AI)
│   ├── memory.py          # MemoryStore (SQLite cross-session)
│   ├── queue.py           # Embedding queue (SQLAlchemy async)
│   ├── security.py        # Whitelist, path validation, ReDoS protection
│   ├── session.py         # Session lifecycle
│   ├── setup_wizard.py    # Interactive first-run wizard
│   ├── telemetry.py       # LangSmith tracing
│   └── log_setup.py       # Structured logging setup
├── ui/
│   ├── chat.py            # Chat TUI (Rich Live, prompt-toolkit)
│   ├── commands.py        # Command dispatcher (/quit, /debug, /model, ...)
│   └── main.py            # Rich components (panels, layouts, widgets)
└── testing/
    ├── fixtures.py        # Pytest fixtures
    ├── mocks.py           # LLM/tool mocks
    └── message_factory.py # Test message builders
```

---

## Checklist de Release

### Core

- [x] 14 ferramentas implementadas e testadas
- [x] LangGraph 3-node compilado com checkpointer
- [x] Persistência SQLite + LanceDB funcional
- [x] MCP Server (stdio + SSE) operacional
- [x] Memory cross-session funcional
- [x] Protocolo Paperclip documentado e validado

### Qualidade

- [x] `ruff check vectora/` — 0 erros
- [x] `mypy vectora/` — tipo-correto
- [x] `pytest tests/ --cov=vectora` — >80% coverage
- [x] Pre-commit hooks passando
- [ ] Release notes escritas
- [ ] Git tag `v0.1.0` criada

### Publicação

- [ ] `uv build` — wheel gerado sem erros
- [ ] GHCR push — `ghcr.io/kaffyn/vectora:0.1.0`
- [ ] PyPI publish — `uv publish`
- [ ] GitHub Release criada com changelog
