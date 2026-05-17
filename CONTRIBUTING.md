# Contributing to Vectora

Obrigado por contribuir! Este guia explica como configurar o ambiente de desenvolvimento e rodar os testes.

## 📋 Tabela de Conteúdos

- [Setup Inicial](#setup-inicial)
- [Instalação de Dependências](#instalação-de-dependências)
- [Rodando Testes](#rodando-testes)
- [Rodando Vectora Localmente](#rodando-vectora-localmente)
- [Padrões de Código](#padrões-de-código)
- [Git Workflow](#git-workflow)

---

## 🚀 Setup Inicial

### Requisitos

- Python 3.13+
- uv (Python package manager)
- Git

### Passo 1: Clonar o Repositório

```bash
git clone https://github.com/seu-user/vectora.git
cd vectora
```

### Passo 2: Instalar Dependências

```bash
# Instalar com todas as dependências (incluindo testes e dev tools)
uv sync --all-extras

# Apenas produção
uv sync
```

---

## 📦 Instalação de Dependências

### SQLite (Database)

✅ **Incluído automaticamente** via `aiosqlite` no `pyproject.toml`

SQLite é um banco de dados embutido que não requer instalação separada. Perfeito para testes e desenvolvimento.

```python
# Configuração automática em testes
DB_DSN="sqlite:///./test.db"
```

### LanceDB (Vector Store)

LanceDB é uma biblioteca Python para armazenar e buscar vetores localmente. É usado nos testes em vez de Qdrant (que requer Docker/servidor separado).

#### Instalação

```bash
# Automático via uv sync
uv sync --all-extras

# Ou instalação manual
uv pip install lancedb>=0.1.0
```

#### Verificar Instalação

```bash
python -c "import lancedb; print(f'LanceDB {lancedb.__version__} instalado')"
```

#### Configuração em Testes

```python
# Automático em conftest.py
VECTOR_STORE_TYPE=lancedb
LANCEDB_DIR=./data/lancedb
```

#### Documentação

- [LanceDB Docs](https://docs.lancedb.com)
- [Python API](https://docs.lancedb.com/python/)

---

## 🧪 Rodando Testes

### Setup de Testes

```bash
# Instalar com dependências de teste
uv sync --all-extras

# Verificar que tudo está pronto
uv run python -c "import pytest; import pytest_asyncio; print('OK')"
```

### Executar Todos os Testes

```bash
# Todos
uv run pytest tests/ -v

# Com coverage
uv run pytest tests/ -v --cov=vectora --cov-report=html
```

### Executar Testes Específicos

```bash
# Apenas testes de MCP (estão passando ✅)
uv run pytest tests/test_mcp_integration.py -v

# Apenas um arquivo
uv run pytest tests/test_community_tools.py -v

# Um teste específico
uv run pytest tests/test_mcp_integration.py::TestMCPClientConnection::test_mcp_client_disabled -v
```

### Debug de Testes

```bash
# Com output detalhado
uv run pytest tests/ -vv -s

# Com breakpoint
uv run pytest tests/ --pdb

# Apenas último teste que falhou
uv run pytest tests/ --lf
```

### Status Atual dos Testes

| Componente      | Status                               | Notas                          |
| --------------- | ------------------------------------ | ------------------------------ |
| MCP Integration | ✅ 10/10 passando                    | Funcional                      |
| Community Tools | ❌ Fixture async em testes síncronos | Precisa refactor               |
| RAG Tools       | ❌ Falta VOYAGE_API_KEY              | Faz mock quando key não existe |
| Graph Flow      | ❌ Mesmos problemas de fixture       | Precisa refactor               |

**Próximo Passo:** Refatorar `conftest.py` para resolver conflitos async/sync (ver issue #XX).

---

## 🚀 Rodando Vectora Localmente

### Opção 1: CLI Interativa (Recomendado)

```bash
# Setup
cp .env.example .env
nano .env  # Adicionar GOOGLE_API_KEY (de https://ai.google.dev)

# Rodar
uv run python vectora/run_main.py

# Ou com alias instalado
uv sync  # Instala scripts em pyproject.toml
vectora-cli
```

**O que você verá:**

```
Você:
[digite sua mensagem]

RESPOSTA (gemini-2.0-flash):
[bot responde...]

---

Você:
_
```

### Opção 2: CLI Rich (Interface Terminal)

```bash
uv run python vectora/run_chat.py

# Ou com alias
vectora-tui
```

### Opção 3: Docker Completo

```bash
# Build e iniciar
docker compose build
docker compose up -d

# Testar API
curl http://localhost:8000/health

# Ver logs
docker compose logs -f vectora
```

---

## 📝 Padrões de Código

### Tipagem e Type Hints

Todo código Python deve ter type hints completos:

```python
# ✅ Bom
async def process_message(text: str) -> dict[str, Any]:
    """Process a user message."""
    result: dict[str, str] = {"processed": text}
    return result

# ❌ Ruim
async def process_message(text):
    result = {"processed": text}
    return result
```

### Imports

Use `isort` para ordenar (automático via pre-commit):

```python
# Ordem: stdlib, third-party, local
import os
from pathlib import Path

from langchain_core.messages import HumanMessage
from rich import print

from context import Context
```

### Docstrings

Use docstrings simples, apenas para o "por quê" não-óbvio:

```python
# ✅ Bom
async def call_graph(state: State) -> dict:
    """Execute the LangGraph workflow."""
    return await graph.astream_events(state)

# ❌ Desnecessário
async def call_graph(state: State) -> dict:
    """Call the graph with the given state.

    This function takes a state object and invokes the graph.
    Returns the result from the graph invocation.
    """
```

### Testing

- Use `pytest` + `pytest-asyncio`
- Fixtures em `vectora/testing/fixtures.py`
- Mocks em `vectora/testing/mocks.py`
- Testes em `tests/` com estrutura espelhando `vectora/`

```python
# tests/test_my_feature.py
import pytest

@pytest.mark.asyncio
async def test_something(test_graph, test_context):
    """Test something."""
    # Arrange
    state = State(messages=[...])

    # Act
    result = await test_graph.astream_events(state)

    # Assert
    assert result["messages"][-1].content
```

---

## 🔄 Git Workflow

### 1. Criar uma Branch

```bash
# A partir de main/develop
git checkout main
git pull origin main

# Criar branch com padrão: feature/xyz, fix/xyz, docs/xyz
git checkout -b feature/my-feature
```

### 2. Fazer Commits

Siga [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git add vectora/my_feature.py
git commit -m "feat: Add my awesome feature"
```

**Tipos válidos:**

- `feat:` - Nova funcionalidade
- `fix:` - Correção de bug
- `docs:` - Documentação
- `refactor:` - Refatoração
- `test:` - Testes
- `chore:` - Build, deps, etc

### 3. Push e Pull Request

```bash
# Push
git push origin feature/my-feature

# Criar PR no GitHub (via web ou CLI)
gh pr create --title "Add my feature" --body "Descrição..."
```

### 4. Pre-commit Hooks

Ruff, isort, mypy, prettier são executados automaticamente:

```bash
# Se falhar, o commit é rejeitado
# Corrija os erros:
uv run ruff check vectora/ --fix
uv run ruff format vectora/
uv run isort vectora/

# Tente novamente
git add vectora/
git commit -m "feat: Add feature"
```

---

## 🤝 Código de Conduta

- Seja respeitoso com outros contribuidores
- Erros e bugs são naturais, vamos aprender juntos
- Pergunte se não tiver certeza antes de submeter

---

## ❓ Dúvidas?

- Abra uma issue no GitHub
- Veja a [Documentação do Projeto](./docs/)
- Leia o [README](./README.md)

---

**Obrigado por contribuir! 🚀**
