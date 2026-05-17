# Vectora MVP (v0.1) - Escopo de Lançamento

**Data:** 2026-05-14  
**Status:** 🔵 Em Desenvolvimento (Fases 1-3 Prontas)  
**Objetivo:** Lançar Vectora v0.1 como agente RAG standalone com integração MCP

---

## 🎯 O que Entra no MVP (v0.1)

### Core Features

**Comportamento de Agente:**

- ✅ Agente RAG completo (busca, indexação, reranking automático)
- ✅ LangGraph com 4 nós (MAIN, SUMMARIZER, TOOL, SUB)
- ✅ Histórico deslizante com auto-summarização
- ✅ Auto-retry inteligente (3x) para tools

**Persistência Local (Development):**

- ✅ SQLite (sessions, threads, checkpoints)
- ✅ LanceDB (vector store local, sem dependências)
- ✅ Pasta `$HOME/.vectora/` (keys, database, logs, sessions)

**Integração MCP (Sub-Agente):**

- ✅ **MCP Server** (Vectora como Sub-Agente do Claude Code)
  - 3 Resources expostos (thread context, history, status)
  - 11 Ferramentas internas (vector_search, web_search, file_read, etc — usadas internamente pelo LangGraph)
  - Protocolo JSON-RPC via stdio (comunicação Agent-to-Agent, não HTTP)
  - Executa LangGraph internamente (raciocínio próprio)
- ✅ **MCP Client** (consome MCPs de servidores externos)
- ✅ Modo duplex simultâneo (servidor + cliente)
- ✅ Transporte stdio (integração com Claude Code, Paperclip, etc)

**Interface:**

- ✅ **CLI Rich:** Setup Wizard + Chat TUI
- ✅ Setup automático de LLM (Gemini, OpenAI, Anthropic, Ollama)
- ✅ Auto-detecção de chaves em variáveis de ambiente
- ✅ Links diretos para obter API keys

**11 Ferramentas Implementadas:**

1. ✅ `web_search` — DuckDuckGo real-time
2. ✅ `fetch_url` — Extração de conteúdo
3. ✅ `embedding` — Indexação Voyage AI
4. ✅ `vector_search` — Busca LanceDB
5. ✅ `file_read` — Leitura com whitelist
6. ✅ `file_edit` — Edição segura
7. ✅ `grep` — Pattern search
8. ✅ `terminal` — Execução shell com whitelist
9. ✅ `list_dir` — Listagem recursiva
10. ✅ `call_mcp_tool` — Invocar MCPs externos
11. ✅ `ingest_docs` — Ingestão em lote de documentos

**Testing:**

- ✅ Unit tests (ferramentas isoladas)
- ✅ Integration tests (fluxo do grafo)
- ✅ E2E tests (chat completo)

**CI/CD & Deployment:**

- ✅ GitHub Actions (testes automáticos)
- ✅ Docker (Dockerfile + docker-compose local)
- ✅ Deploy em VPS (simulando instalação produção)

**Publicação:**

- ✅ **GHCR** (imagem Docker: `ghcr.io/user/vectora:0.1.0`)
- ✅ **GitHub MCP Registry** (registrar como servidor MCP)
- ✅ **UV Package** (instalável: `uv pip install vectora`)

**Observabilidade:**

- ✅ Logs estruturados em `$HOME/.vectora/logs/`
- ✅ LangSmith integration (opcional via LANGSMITH_API_KEY)
- ✅ Métricas por ferramenta (latência, tokens, etc)

**Arquitetura de Sub-Agente (MCP):**

Vectora não é apenas uma "ferramenta" para o Claude Code, mas um **agente colaborativo independente** com seu próprio raciocínio. Claude Code **comunica com** Vectora, não chama suas ferramentas diretamente:

1. **Ferramentas Internas** (usadas pelo LangGraph do Vectora):

   - `vector_search` — Busca conhecimento técnico profundo em LanceDB
   - `web_search` — Contexto externo em tempo real via DuckDuckGo
   - `file_read`, `file_edit`, `terminal` — Manipulação de sistema
   - `grep`, `list_dir` — Exploração do codebase
   - `embedding`, `ingest_docs`, `call_mcp_tool` — Integração e indexação

2. **Resources Expostos via MCP** (lidos por Claude Code):

   - `vectora://thread/{id}/context` — Resumo do conhecimento coletado
   - `vectora://thread/{id}/history` — Histórico da conversa (últimas 5 msgs)
   - `vectora://status` — Status do servidor (LLM, RAG disponível, uptime)

3. **Fluxo de Comunicação (stdio JSON-RPC):**
   ```
   Claude Code
       ↓ (lê Resources)
   vectora://thread/123/context
       ↓ (toma decisão)
   Comunica com Vectora: "Preciso saber sobre X"
       ↓
   MCP Server Vectora
       ↓ (processa internamente)
   LangGraph: MAIN → TOOL → SUMMARIZER → SUB
       ↓ (executa ferramentas internas)
   web_search(), vector_search(), etc
       ↓ (responde ao Claude Code)
   Resultado processado
   ```

**Ponto crítico:** Claude Code lê o "estado cognitivo" do Vectora (Resources) antes de comunicar requisições de alto nível. Não chama ferramentas individuais.

**Integração com Paperclip:**

- ✅ Funcionar como MCP Server via stdio
- ✅ Expor 3 Resources (context, history, status) para orquestração do Paperclip
- ✅ Executar LangGraph próprio com 11 ferramentas internas (não apenas delegar funções)
- ✅ Suportar contexto reduzido (resumo + últimas 2 msgs)

---

## ❌ O que NÃO Entra no MVP (Pós-MVP)

### Fase Post-MVP 1: Infraestrutura Escalável (Semanas 1-4)

**Persistência Avançada:**

- ⏳ PostgreSQL (remoto, produção)
- ⏳ Qdrant Cloud (vector store escalável)
- ⏳ Valkey (cache distribuído)
- ⏳ Migrações automáticas SQLite → PostgreSQL

**Arquitetura:**

- ⏳ Deep Agents (sub-agentes especializados)
- ⏳ Agent Client Protocol (ACP) padrão
- ⏳ Balanceamento de carga (múltiplas instâncias)

**Observabilidade:**

- ⏳ Prometheus/Grafana (métricas em tempo real)
- ⏳ Jaeger (distributed tracing)
- ⏳ Alert system (Slack, Discord, PagerDuty)

---

### Fase Post-MVP 2: Ecossistema de Plugins (Semanas 5-8)

**Plugins Oficiais:**

- ⏳ **VSCode Extension** (editar código com Vectora inline)
- ⏳ **Gemini CLI Plugin** (usar Vectora como tool)
- ⏳ **Codex Extension** (integração nativa)
- ⏳ **Paperclip Deep Integration** (orquestração avançada)

**Vectora Asset Library:**

- ⏳ Buckets de dados pré-treinados
- ⏳ Contextos por linguagem (Python, Go, JS, etc)
- ⏳ Contextos por framework (FastAPI, React, etc)
- ⏳ Compartilhamento comunitário de assets

---

### Fase Post-MVP 3: CLI Oficial & Documentação (Semanas 9-12)

**CLI do Deep Agents:**

- ✅ CLI Rich robusto com Panels, Tables, Layouts
- ⏳ Suporte a múltiplas sessões em paralelo
- ⏳ Gerenciador de agentes especializados
- ⏳ Dashboard CLI (status, métricas, logs)

**Website & Documentação:**

- ⏳ Site oficial (docs.vectora.dev)
- ⏳ API Reference (MCP, REST, CLI)
- ⏳ Manual de instalação (Docker, UV, VPS)
- ⏳ Guias de integração (Paperclip, VSCode, etc)
- ⏳ Showcase de casos de uso

**Comunidade:**

- ⏳ GitHub Discussions (fórum)
- ⏳ Discord Server
- ⏳ Contribuição guidelines

---

## 📁 Estrutura de Diretórios MVP

```
$HOME/.vectora/
├── config.toml           # Configurações (não-sensível)
├── .env                  # Variáveis de ambiente (sensível, .gitignore)
├── keys/
│   ├── api_keys.json     # API keys encriptadas (opcional)
│   └── mcp_config.json   # Configuração de MCPs
├── data/
│   ├── sqlite.db         # Banco de dados local
│   ├── lancedb/          # Vector store
│   │   ├── articles/
│   │   ├── wiki/
│   │   ├── api_docs/
│   │   └── knowledge_base/
│   └── cache/            # Cache temporário
├── sessions/             # Histórico de sessões
│   ├── session_001.json
│   ├── session_002.json
│   └── ...
├── logs/                 # Logs estruturados
│   ├── vectora.log
│   ├── mcp_client.log
│   └── mcp_server.log
├── plugins/              # Plugins customizados (futuro)
└── .vectora_rc           # Shell rc file (opcional)
```

---

## 🐳 Deployment MVP

### Local (Desenvolvimento)

```bash
# Instalação via UV
uv pip install vectora

# Setup interativo
vectora setup

# Iniciar chat
vectora chat

# Iniciar como MCP Server
vectora mcp-server
```

### Docker (MVP)

```bash
# Build
docker build -t vectora:0.1.0 .

# Run local
docker run -it \
  -e GOOGLE_API_KEY=xxx \
  -v ~/.vectora:/root/.vectora \
  vectora:0.1.0

# Push to GHCR
docker push ghcr.io/user/vectora:0.1.0
```

### VPS (MVP)

```bash
# SSH into VPS
ssh root@vps

# Instalar Docker + Docker Compose
curl -fsSL https://get.docker.com | bash

# Deploy Vectora + Paperclip
docker-compose -f docker-compose.prod.yml up -d

# Verificar que MCP Server está respondendo
# MCP usa stdio JSON-RPC, não HTTP
docker logs vectora-mcp-server  # Ver logs do servidor
```

**Nota:** O MCP Server do Vectora usa **stdio JSON-RPC**, não HTTP/REST. Não há endpoint `/health`. A comunicação é via stdin/stdout (processo padrão).

---

## 🚀 Releases & Publicação MVP

### GitHub Releases

- ✅ **v0.1.0-alpha** (testes completos locais)
- ✅ **v0.1.0-beta** (deploy em VPS validado)
- ✅ **v0.1.0** (release oficial MVP)

### Package Registries

**1. GHCR (Container Registry)**

```bash
ghcr.io/user/vectora:0.1.0
ghcr.io/user/vectora:latest
```

**2. UV (Python Package)**

```bash
uv pip install vectora==0.1.0
uv pip install vectora  # Pega latest
```

**3. GitHub MCP Registry**

```
MCP Name: vectora
Repository: github.com/user/vectora
Type: server (stdio)
Capabilities: tools, prompts
```

---

## ✅ Checklist MVP

### Fase 1: Setup Wizard

- [x] Rich CLI com Setup Wizard e Chat Interface
- [x] Auto-detecção de LLMs
- [x] Teste de conexão
- [x] Salvamento em ~/.vectora/

### Fase 2: Core Graph

- [x] LangGraph 4-node pattern
- [x] Histórico deslizante
- [x] Auto-retry inteligente
- [x] LangSmith integration

### Fase 3: 11 Ferramentas

- [x] Todas as 11 tools implementadas
- [x] Whitelisting seguro
- [x] Logging estruturado
- [x] Type hints completos

### Fase 4: MCP Integration

- [x] MCP Server (expor tools)
- [x] MCP Client (consumir MCPs)
- [x] Transporte stdio
- [x] Integração Paperclip validada

### Fase 5: Testing & CI/CD

- [ ] Unit tests (10+)
- [ ] Integration tests (5+)
- [ ] E2E tests (3+)
- [ ] GitHub Actions workflow
- [ ] Docker build & push
- [ ] VPS deployment script

### Publicação

- [ ] GHCR push
- [ ] UV package publish
- [ ] GitHub MCP Registry
- [ ] Release notes em português

---

## 🔒 Segurança MVP

**O que está dentro:**

- ✅ Whitelist de comandos shell
- ✅ Validação de paths (no directory traversal)
- ✅ Proteção contra ReDoS em regex
- ✅ Encriptação de chaves (opcional)
- ✅ Logs sem expor secrets

**O que NÃO está no MVP:**

- ❌ Autenticação (não há usuários)
- ❌ TLS/SSL (local apenas)
- ❌ Rate limiting
- ❌ Audit log completo

---

## 📊 Métricas de Sucesso MVP

**Funcionalidade:**

- ✅ 11/11 ferramentas operacionais
- ✅ RAG end-to-end funcionando
- ✅ MCP Server/Client integrados
- ✅ CLI Rich responsiva com Components avançados

**Qualidade:**

- ✅ >80% test coverage
- ✅ Todos os tipos validados (mypy)
- ✅ Linting passa (ruff)
- ✅ Zero warnings em build

**Deployment:**

- ✅ Docker image <500MB
- ✅ Install via UV < 30s
- ✅ Setup wizard < 2 min
- ✅ MCP Server startup < 5s

**Integração MCP:**

- ✅ Funciona com Paperclip 100% (como Sub-Agente)
- ✅ Expõe 3 Resources + comunicação Agent-to-Agent via MCP (stdio JSON-RPC)
- ✅ Executa LangGraph interno com 11 ferramentas e raciocínio próprio
- ✅ Documentado para outros clients MCP-compatíveis

---

## 🗺️ Roadmap Pós-MVP

```
v0.1.0 (MVP) ────────────────┐
                              │
                              ├─→ v0.2.0: Deep Agents + PostgreSQL + Qdrant
                              │
                              ├─→ v0.3.0: Plugins (VSCode, Gemini CLI)
                              │
                              ├─→ v0.4.0: Vectora Asset Library
                              │
                              └─→ v1.0.0: CLI oficial + Website
```

---

## 📝 Notas Finais

**MVP é congelado em branch v0.1:**

```bash
git tag -a v0.1.0 -m "MVP Release"
git branch release/0.1 main
```

**Pós-MVP desenvolvimento em `main` com Deep Agents**

**Todos os requirements MVP com ✅ acima devem estar 100% funcionais antes de tag v0.1.0**
