# 🔄 Guia de Migração - Eliminação de Legacy Config

**Status**: PRONTO PARA EXECUTAR
**Data**: 2026-05-16
**Objetivo**: Consolidar `constants.py`, `tool_config.py`, `config.py` em um único `settings.py`

---

## 📋 Resumo Executivo

O novo `settings.py` com Pydantic **substitui completamente** os três arquivos antigos:

| Arquivo Antigo   | O que Fazia                 | Substituído Por                                   |
| ---------------- | --------------------------- | ------------------------------------------------- |
| `constants.py`   | Paths, versão, DB DSN       | `Settings.{data_dir, db_dsn, version}`            |
| `tool_config.py` | RAG, embeddings, MCP config | `Settings.{voyage_api_key, embedding_model, etc}` |
| `config.py`      | 3-level env hierarchy       | `Settings._load_environment_hierarchy()`          |

---

## 🎯 Ordem de Migração (SEM QUEBRAR NADA)

### PASSO 1: Validar que Settings funciona

```bash
# Tente rodar a app com o novo Settings
uv run vectora

# Deve inicializar com sucesso (confirmado ✅)
```

### PASSO 2: Migrar imports (um arquivo por vez)

Para cada arquivo na lista abaixo, **siga o padrão**:

1. **Encontre** todas as linhas que importam do arquivo antigo
2. **Substitua** pela importação do novo Settings
3. **Atualize** o código para usar `settings.CAMPO`
4. **Teste** se roda sem erros
5. **Commite** com mensagem clara

---

## 📋 Lista de Arquivos para Migrar

### ✅ 1. `background_worker.py`

**ANTES**:

```python
from tool_config import ToolConfig, get_tool_config

config = get_tool_config()
db_url = config.embedding_queue_url
lancedb_path = config.lancedb_path
```

**DEPOIS**:

```python
from settings import settings

db_url = settings.embedding_queue_dsn  # Já é um SQLite path
lancedb_path = Path(settings.lancedb_dir)
```

**Mudanças**:

- `get_tool_config()` → `settings` (já instantiado e validado)
- `config.embedding_queue_url` → `settings.embedding_queue_dsn`
- `config.lancedb_path` → `Path(settings.lancedb_dir)`
- `config.voyage_api_key` → `settings.voyage_api_key`
- `config.embedding_model` → `settings.embedding_model`

**Comando para encontrar imports**:

```bash
grep -n "from tool_config\|from config" vectora/background_worker.py
```

---

### ✅ 2. `chat.py`

**ANTES**:

```python
from constants import DB_DSN

checkpointer = Checkpointer(DB_DSN)
```

**DEPOIS**:

```python
from settings import settings

checkpointer = Checkpointer(settings.db_dsn)
```

**Mudanças**:

- `DB_DSN` → `settings.db_dsn`
- `LOG_LEVEL` → `settings.log_level`
- `DEBUG_MODE` → `settings.debug_mode`

---

### ✅ 3. `checkpointer.py`

**ANTES**:

```python
from constants import DB_DSN

async def __aenter__(self):
    self.conn = await aiosqlite.connect(DB_DSN)
```

**DEPOIS**:

```python
from settings import settings

async def __aenter__(self):
    self.conn = await aiosqlite.connect(settings.db_dsn)
```

---

### ✅ 4. `commands.py`

**ANTES**:

```python
from config import Config
from constants import DB_DSN

config = Config.instance()
```

**DEPOIS**:

```python
from settings import settings

# Remova a instanciação, use settings diretamente
provider = settings.llm_provider
model = settings.get_llm_model()
```

---

### ✅ 5. `debug_dump.py`

**ANTES**:

```python
from tool_config import get_tool_config

config = get_tool_config()
lancedb_path = config.lancedb_path
```

**DEPOIS**:

```python
from settings import settings
from pathlib import Path

lancedb_path = Path(settings.lancedb_dir)
```

---

### ✅ 6. `mcp_server.py`

**ANTES**:

```python
from constants import VERSION
from tool_config import get_tool_config

version = VERSION
config = get_tool_config()
mcp_enabled = config.enable_mcp
```

**DEPOIS**:

```python
from settings import settings

version = settings.version
mcp_enabled = settings.enable_mcp
```

---

### ✅ 7. `memory_store.py`

**ANTES**:

```python
from constants import DB_DSN

db_url = DB_DSN
```

**DEPOIS**:

```python
from settings import settings

db_url = settings.db_dsn
```

---

### ✅ 8. `tools.py`

**ANTES**:

```python
from tool_config import get_tool_config

config = get_tool_config()
tavily_key = config.tavily_api_key
voyage_key = config.voyage_api_key
```

**DEPOIS**:

```python
from settings import settings

tavily_key = settings.tavily_api_key
voyage_key = settings.voyage_api_key
```

---

## 🔄 Mapeamento Completo de Campos

| Campo Antigo                    | Arquivo        | Campo Novo                            |
| ------------------------------- | -------------- | ------------------------------------- |
| `constants.DB_DSN`              | constants.py   | `settings.db_dsn`                     |
| `constants.EMBEDDING_QUEUE_DSN` | constants.py   | `settings.embedding_queue_dsn`        |
| `constants.LANCEDB_DIR`         | constants.py   | `settings.lancedb_dir`                |
| `constants.VERSION`             | constants.py   | `settings.version`                    |
| `constants.VECTORA_HOME`        | constants.py   | `settings.vectora_home`               |
| `tool_config.voyage_api_key`    | tool_config.py | `settings.voyage_api_key`             |
| `tool_config.embedding_model`   | tool_config.py | `settings.embedding_model`            |
| `tool_config.enable_rag`        | tool_config.py | `settings.enable_rag`                 |
| `tool_config.enable_mcp`        | tool_config.py | `settings.enable_mcp`                 |
| `tool_config.tavily_api_key`    | tool_config.py | `settings.tavily_api_key`             |
| `Config.instance()`             | config.py      | `settings` (singleton já instanciado) |

---

## ⚠️ Cuidados Importantes

### 1. `get_tool_config()` é uma função, `settings` é um objeto

**ANTES** (função que retorna singleton):

```python
config = get_tool_config()
value = config.enable_rag
```

**DEPOIS** (objeto importado direto):

```python
from settings import settings

value = settings.enable_rag
```

### 2. Alguns campos têm nomes diferentes

- `embedding_queue.embedding_queue_url` → `embedding_queue_dsn` (Path, não URL)
- `config.lancedb_path` (property) → `Path(settings.lancedb_dir)` (string que vira Path)

---

## 🧪 Teste Após Cada Migração

```bash
# Depois de migrar cada arquivo:
uv run vectora

# Verificar se inicializa sem erros
# Tentar digitar uma mensagem
# Verificar se responde
```

---

## 🗑️ Limpeza Final (Depois que TUDO rodar)

Depois que todos os arquivos forem migrados e testados:

```bash
# DELETE:
rm vectora/constants.py
rm vectora/config.py
rm vectora/tool_config.py

# VERIFY:
grep -r "from constants\|from config\|from tool_config" vectora/

# Se retornar nada, você eliminou com sucesso!
```

---

## 📝 Checklist de Migração

- [ ] `background_worker.py` → testado
- [ ] `chat.py` → testado
- [ ] `checkpointer.py` → testado
- [ ] `commands.py` → testado
- [ ] `debug_dump.py` → testado
- [ ] `mcp_server.py` → testado
- [ ] `memory_store.py` → testado
- [ ] `tools.py` → testado
- [ ] Grep confirma zero imports legados
- [ ] Deletar `constants.py`
- [ ] Deletar `config.py`
- [ ] Deletar `tool_config.py`
- [ ] Final test: `uv run vectora` e chat funciona

---

## 🎯 Benefícios Esperados Após Migração

✅ **Zero `NoneType` errors** - Pydantic valida tudo na inicialização
✅ **Configuração centralizada** - Um único arquivo `settings.py`
✅ **IDE autocompletion** - Type hints em todos os campos
✅ **Código mais limpo** - Menos imports "mágicos"
✅ **Easier testing** - Mock `settings` ao invés de 3 arquivos diferentes

---

## 🚀 Próximo Passo

**Comece pelo `background_worker.py`** - é o primeiro na cadeia de dependências.

Depois que estiver feito, os outros são em cascata.

---

**Status**: PRONTO ✅
**Tempo Estimado**: 1-2 horas (8 arquivos × 15 min cada)
**Risco**: BAIXO (Settings já validado e rodando)
