# QA Guide

Guia de testes e validação de qualidade para Vectora. Cobre testes automatizados, checklist de release e cenários manuais críticos.

---

## Rodar Testes

### Setup

```bash
# Instalar dependências de teste
uv sync

# Verificar ambiente
uv run python -c "import pytest; import pytest_asyncio; print('✓ OK')"
```

### Testes Unitários

```bash
# Todos
uv run pytest tests/unit/ -v

# Com cobertura
uv run pytest tests/unit/ --cov=vectora --cov-report=html

# Arquivo específico
uv run pytest tests/unit/test_tools_core.py -v
```

### Testes de Integração

```bash
# Pipeline RAG completo
uv run pytest tests/integration/test_rag_persistence.py -v

# Execução do grafo LangGraph
uv run pytest tests/integration/test_graph_execution.py -v

# Agent-to-Agent (A2A / Paperclip)
uv run pytest tests/integration/test_a2a_integration.py -v

# Flow de mensagens
uv run pytest tests/integration/test_message_flow.py -v
```

### Testes E2E

```bash
# Fire-and-forget (embedding background worker)
uv run pytest tests/e2e/test_fire_and_forget.py -v

# MCP Resources
uv run pytest tests/e2e/test_mcp_resources.py -v

# Comandos CLI
uv run pytest tests/e2e/test_run_commands.py -v
```

### Stress Tests

```bash
# Concorrência: múltiplos threads simultâneos
uv run pytest tests/stress/test_concurrency.py -v
```

### Suite Completa com Coverage

```bash
uv run pytest tests/ --cov=vectora --cov-report=html --cov-report=term-missing
# Abrir relatório HTML
open htmlcov/index.html
```

---

## Checklist de Release

Execute antes de qualquer release. Todos os itens devem estar ✅.

### Code Quality

- [ ] `uv run ruff check vectora/` — 0 erros, 0 warnings
- [ ] `uv run ruff format vectora/ --check` — nenhum arquivo a formatar
- [ ] `uv run mypy vectora/` — sem erros de tipo
- [ ] `uv run pytest tests/ --cov=vectora` — >80% coverage
- [ ] `uv run bandit -r vectora/ -ll` — sem issues HIGH ou MEDIUM

### Functional (Manual Smoke Test)

```bash
# 1. Setup fresh (sem ~/.vectora existente)
rm -rf ~/.vectora && vectora setup

# 2. Chat básico
vectora chat
# Digitar: "olá, quem é você?"
# Esperado: resposta coerente em <5s

# 3. Tool call (web search)
# Digitar: "pesquise sobre LangGraph 0.2"
# Esperado: painel amarelo [TOOL CALL], resposta com fontes

# 4. RAG round-trip
# Digitar: "indexe o conteúdo: Python é uma linguagem interpretada de alto nível"
# Esperar: "Document enqueued for async embedding"
# Após 10s: "busque sobre linguagem Python"
# Esperado: resposta usando o documento indexado

# 5. Memory persistence
# Digitar: "lembre que meu projeto se chama Vectora"
# Reiniciar: Ctrl+C && vectora chat
# Digitar: "qual é o nome do meu projeto?"
# Esperado: "Vectora" (lido da memória SQLite)

# 6. MCP Server
vectora mcp-server
# Esperado: painel Rich no terminal, aguarda sem fechar
# Ctrl+C para encerrar
```

### Integration

- [ ] Multi-turn conversation — histórico mantém contexto entre >10 mensagens
- [ ] Session persistence — dados recuperados após reiniciar com mesmo `thread_id`
- [ ] Thread isolation — dois threads diferentes não vazam contexto entre si
- [ ] Error handling — resposta graciosa em queries inválidas (sem crash)
- [ ] Tool timeout — ferramentas lentas retornam erro após timeout sem travar a UI

### Deployment

- [ ] `docker compose build` — sucesso, sem erros
- [ ] `docker compose run --rm vectora vectora setup` — wizard funcional no container
- [ ] Imagem Docker < 500MB (`docker image inspect vectora | jq '.[0].Size'`)
- [ ] `uv build` — wheel gerado sem erros

---

## Debug & Troubleshooting

### Logs Estruturados

```bash
# Acompanhar logs em tempo real
tail -f ~/.vectora/logs/vectora.log | python -c "import sys,json; [print(json.dumps(json.loads(l), indent=2)) for l in sys.stdin]"

# Filtrar por componente
grep '"name": "vectora.mcp.server"' ~/.vectora/logs/vectora.log

# Filtrar por nível
grep '"levelname": "ERROR"' ~/.vectora/logs/vectora.log

# MCP server logs
tail -f ~/.vectora/logs/mcp.log
```

### LangSmith Tracing

```bash
export LANGSMITH_API_KEY=your_key
export LANGSMITH_TRACING=true
export LANGSMITH_PROJECT=vectora-debug

vectora chat
# Traces aparecem em https://smith.langchain.com
```

### Debug Mode no Chat

```
/debug
```

Ativa exibição de todos os tool calls (amarelo), tool responses (vermelho) e terminal output (verde) em tempo real.

### Dump de Estado

```python
# Inspecionar checkpoint de um thread
from vectora.services.checkpoint import Checkpointer
import asyncio

async def dump():
    async with Checkpointer() as cp:
        config = {"configurable": {"thread_id": "1"}}
        state = await cp.aget(config)
        print(state)

asyncio.run(dump())
```

---

## Cenários de Teste Manual

### Cenário 1 — RAG End-to-End

1. `vectora chat`
2. Prompt: `"indexe este texto: FastAPI é um framework web Python assíncrono"`
3. Esperado: confirmação de enfileiramento (`status: queued`)
4. Aguardar 15s (background worker processa)
5. Prompt: `"o que é FastAPI?"`
6. Esperado: resposta usando o documento indexado (não web search)

### Cenário 2 — Memory Cross-Session

1. `vectora chat` → `"guarde que prefiro Python 3.13 e uso Mac"`
2. `Ctrl+C` → `vectora chat`
3. `"qual é minha preferência de Python?"`
4. Esperado: Vectora responde com a preferência salva na sessão anterior

### Cenário 3 — Tool Timeout

1. `vectora chat` (com `/debug` ativo)
2. Prompt: `"execute: sleep 120"` (bloqueia por 120s)
3. Esperado: tool retorna mensagem de erro após ~30s (timeout do terminal)
4. TUI permanece responsiva durante e após o timeout

### Cenário 4 — Multi-Thread Isolation (A2A)

```bash
# Terminal 1 — Iniciar MCP server
MCP_TRANSPORT=sse vectora mcp-server

# Terminal 2 — Agente A
python -c "
import asyncio
from vectora.mcp.proxy import create_remote_proxy

async def main():
    async with create_remote_proxy('http://localhost:8000/sse') as v:
        r = await v.delegate('lembre que meu nome é Alice', thread_id='agent_alice')
        print('Alice:', r)
asyncio.run(main())
"

# Terminal 3 — Agente B
python -c "
import asyncio
from vectora.mcp.proxy import create_remote_proxy

async def main():
    async with create_remote_proxy('http://localhost:8000/sse') as v:
        r = await v.delegate('qual é o meu nome?', thread_id='agent_bob')
        print('Bob:', r)
asyncio.run(main())
"
# Esperado: Bob responde que não sabe seu nome (contexto isolado de Alice)
```

---

## Metas de Cobertura

| Componente                                  | Meta     | Status |
| ------------------------------------------- | -------- | ------ |
| `tools/` (fs, rag, web, memory)             | 90%+     | ✅     |
| `nodes/` (engine, debug)                    | 85%+     | ✅     |
| `services/` (checkpoint, embedding, memory) | 80%+     | ✅     |
| `mcp/` (server, client, proxy)              | 70%+     | ✅     |
| `ui/` (chat, commands)                      | 60%+     | ⚠️     |
| **Overall**                                 | **>80%** | ✅     |

A UI tem coverage menor por design — componentes Rich são difíceis de testar sem display. O foco é nos componentes de lógica.

---

## Aprovação de Release

Todos os itens abaixo devem estar completos antes de criar a tag `v0.1.0`:

- [ ] Checklist de release acima — 100% completo
- [ ] Test coverage >80% confirmado
- [ ] Nenhuma issue aberta com label `blocker`
- [ ] AGENTS.md, README.md, MVP_SCOPE.md revisados e atualizados
- [ ] `git tag -a v0.1.0 -m "MVP Release — Vectora v0.1.0"` criada
- [ ] `uv publish` executado com sucesso
- [ ] GitHub Release criada com changelog
