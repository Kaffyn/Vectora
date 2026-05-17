# Contributing to Vectora

Obrigado por contribuir! Este guia cobre tudo que você precisa para configurar o ambiente, entender os padrões e submeter código de qualidade.

---

## Setup do Ambiente

### Requisitos

- Python 3.13+
- [uv](https://github.com/astral-sh/uv) — gerenciador de dependências
- Git

### Instalação

```bash
# Clonar
git clone https://github.com/Kaffyn/vectora.git
cd vectora

# Instalar todas as dependências (incluindo dev e test)
uv sync

# Instalar pre-commit hooks
uv run pre-commit install

# Copiar e configurar .env
cp .env.example .env
# Editar .env com pelo menos GOOGLE_API_KEY e COHERE_API_KEY
```

### Verificar Setup

```bash
# Rodar testes básicos
uv run pytest tests/unit/ -v

# Verificar linting
uv run ruff check vectora/

# Verificar tipos
uv run mypy vectora/

# Iniciar o chat localmente
uv run vectora chat
```

---

## Rodando Testes

### Todos os testes

```bash
uv run pytest tests/ -v
```

### Com coverage

```bash
uv run pytest tests/ --cov=vectora --cov-report=html
# Abrir htmlcov/index.html para relatório detalhado
```

### Por categoria

```bash
uv run pytest tests/unit/        # Unitários (rápidos, sem IO externo)
uv run pytest tests/integration/ # Integração (RAG, grafo, A2A)
uv run pytest tests/e2e/         # End-to-end (chat completo, MCP)
uv run pytest tests/stress/      # Stress (concorrência)
```

### Debug de testes

```bash
uv run pytest tests/ -vv -s         # Output detalhado
uv run pytest tests/ --pdb          # Breakpoint em falhas
uv run pytest tests/ --lf           # Só o último teste que falhou
uv run pytest tests/ -k "rag"       # Apenas testes que contém "rag" no nome
```

### Coverage Target

- **Overall:** >80%
- **tools/**, **services/**: >85%
- **ui/**: >60% (Rich components são difíceis de testar)

---

## Padrões de Código

### Tipagem — Obrigatório

Todo código deve ter type hints completos. Python 3.13+ syntax:

```python
# ✅ Correto
async def search_docs(query: str, limit: int = 5) -> list[dict[str, str]]:
    """Busca documentos no LanceDB."""
    ...

# ❌ Errado — sem tipos
async def search_docs(query, limit=5):
    ...
```

Use `pydantic` para contratos de dados entre camadas:

```python
from pydantic import BaseModel

class SearchResult(BaseModel):
    content: str
    score: float
    source: str | None = None
```

### Async — Tudo I/O-bound deve ser async

```python
# ✅ Correto — async para database e rede
async def save_memory(key: str, content: str) -> str:
    async with aiosqlite.connect(db_path) as db:
        await db.execute("INSERT ...", (key, content))
        await db.commit()

# ❌ Errado — bloqueia o event loop
def save_memory(key: str, content: str) -> str:
    conn = sqlite3.connect(db_path)
    conn.execute("INSERT ...", (key, content))
```

### Ferramentas (Tools) — Tratamento de Erros

Toda ferramenta DEVE ter `try/except` para que falhas retornem ao LLM como observação, sem derrubar o grafo:

```python
@tool
async def minha_tool(input: str) -> str:
    """Descrição clara para o LLM entender quando usar."""
    try:
        result = await fazer_algo(input)
        logger.info("minha_tool completed", extra={"input": input[:50]})
        return result
    except Exception:
        logger.exception("minha_tool failed", extra={"input": input[:50]})
        return "Error: falha na tool. Verifique os logs."
```

### Imports — Ordem padrão (isort automático)

```python
# 1. stdlib
import asyncio
import logging
from pathlib import Path

# 2. third-party
from langchain.tools import tool
from pydantic import BaseModel
from rich.panel import Panel

# 3. local
from vectora.config.settings import settings
from vectora.services.security import is_safe_file_path
```

### Docstrings — Apenas o necessário

```python
# ✅ Bom — explica o "por quê" não óbvio
async def call_llm(state: State, config: RunnableConfig) -> dict:
    """Invoke LLM with sliding window history.

    Uses trim_messages with fallback to prevent 'contents are required'
    error when ToolMessage alone exceeds max_context_tokens.
    """

# ❌ Desnecessário — só repete o nome da função
async def call_llm(state: State, config: RunnableConfig) -> dict:
    """Call the LLM with the given state and config."""
```

---

## Estrutura do Projeto

```
vectora/
├── agent.py          # AgentManager — orquestrador principal
├── graph.py          # LangGraph builder
├── state.py          # TypedDict State
├── context.py        # Context schema
├── prompts.py        # System prompts
├── main.py           # CLI entry point
├── version.py        # Versão dinâmica via importlib.metadata
├── config/           # Settings (Pydantic), defaults.env
├── nodes/            # Nós do LangGraph (engine, debug)
├── tools/            # 14 ferramentas (fs, rag, web, memory, mcp)
├── mcp/              # MCP Server, Client, VectoraProxy
├── services/         # Serviços (embedding, memory, checkpoint, security...)
├── ui/               # TUI (chat, commands, rich components)
└── testing/          # Fixtures, mocks, message factories
```

---

## Git Workflow

### 1. Criar branch

```bash
git checkout main && git pull origin main
git checkout -b feat/minha-feature
# ou fix/bug-descricao, docs/atualizar-readme, etc.
```

### 2. Commits — Conventional Commits (obrigatório)

```bash
git add vectora/tools/nova_tool.py
git commit -m "feat: add nova_tool para X"
```

Tipos válidos:

- `feat:` — Nova funcionalidade
- `fix:` — Correção de bug
- `docs:` — Documentação apenas
- `refactor:` — Refatoração sem mudança de comportamento
- `test:` — Testes
- `chore:` — Deps, build, config

### 3. Pre-commit Hooks

Os hooks rodam automaticamente no `git commit`:

```
Ruff Lint          → formatação e linting Python
Ruff Format        → estilo consistente
Mypy               → type checking
Prettier           → markdown e YAML
Bandit             → security scan
```

Se um hook falhar, corrija o erro e faça `git add` novamente antes de re-commitar:

```bash
uv run ruff check vectora/ --fix
uv run ruff format vectora/
git add vectora/
git commit -m "feat: minha feature"
```

### 4. Pull Request

```bash
git push origin feat/minha-feature
gh pr create --title "feat: minha feature" --body "Descrição..."
```

---

## Adicionando uma Nova Ferramenta

1. Criar em `vectora/tools/<categoria>.py` com decorator `@tool`
2. Adicionar ao `vectora/tools/__init__.py` (imports + `__all__`)
3. Registrar no `vectora/mcp/server.py` como `@mcp.tool()` com timeout
4. Atualizar `vectora/config/settings.py` se precisar de feature flag
5. Escrever testes em `tests/unit/test_tools_core.py`
6. Atualizar `MVP_SCOPE.md` e `README.md`

---

## Adicionando um Novo Nó ao LangGraph

1. Implementar a função do nó em `vectora/nodes/engine.py`
2. Registrar no builder em `vectora/graph.py` com `builder.add_node()`
3. Adicionar edges (`add_edge` ou `add_conditional_edges`)
4. Atualizar `State` em `vectora/state.py` se o nó precisar de novo campo
5. Escrever testes em `tests/integration/test_graph_execution.py`

---

## Código de Conduta

- Respeito em issues, PRs e discussões
- Erros são oportunidades de aprendizado, não motivo de julgamento
- Pergunte antes de submeter grandes mudanças arquiteturais
- Abra uma issue primeiro para discutir features significativas

Dúvidas? Abra uma issue no GitHub.
