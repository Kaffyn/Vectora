# Ferramentas da Comunidade - Vectora Phase 3

Este documento descreve as 4 ferramentas (tools) da comunidade LangChain integradas no Vectora Phase 3.

## Visão Geral

Vectora implementa 4 tools prontas da comunidade LangChain, todas validadas e mantidas pela comunidade:

1. **Web Search** (DuckDuckGo) - Busca na web em tempo real
2. **Web Fetch** (URL Reader) - Lê conteúdo de URLs específicas
3. **Database Query** (SQL) - Acessa bancos de dados via SQL
4. **MCP Server Call** - Integra com servidores MCP (Model Context Protocol)

Todas as tools seguem o padrão de implementação do LangChain com:

- Decorador `@tool` do `langchain.tools`
- Assinatura `ToolRuntime[Context, State]` para acesso a contexto
- Logging estruturado para observabilidade
- Validações de segurança

---

## 1. Web Search (DuckDuckGo)

### O que faz

Busca na web usando DuckDuckGo e retorna os top 5 resultados com título, URL e snippet.

### Ativação

```bash
# Por padrão, está ativada
ENABLE_WEB_SEARCH=true
```

### Exemplo de Uso

```
Usuário: "Pesquise sobre arquitetura de RAG em LLMs"
Assistente: [chama web_search com query]
Assistente: "Encontrei 5 resultados sobre RAG..."
```

### Vantagens

- ✓ Gratuito (sem API key necessária)
- ✓ Nenhuma configuração adicional
- ✓ Resultados em tempo real
- ✓ Community-maintained (LangChain Community)

### Limitações

- Máximo 5 resultados por busca
- DuckDuckGo pode ter diferentes políticas de acesso por região
- Conteúdo pode ser desatualizado se DuckDuckGo não atualizou

### Implementação

```python
from langchain_community.tools.duckduckgo_search import DuckDuckGoSearchResults

searcher = DuckDuckGoSearchResults(max_results=5)
results = searcher.run("query")
```

---

## 2. Web Fetch (URL Reader)

### O que faz

Extrai conteúdo de texto de uma URL específica. Suporta HTML, PDF e outros formatos.

### Ativação

```bash
ENABLE_WEB_FETCH=true
WEB_FETCH_MAX_SIZE=5000  # máximo de caracteres
WEB_FETCH_ALLOWED_DOMAINS=""  # whitelist (opcional)
```

### Exemplo de Uso

```
Usuário: "Leia o conteúdo de https://docs.langchain.com/langraph"
Assistante: [chama fetch_url]
Assistente: "Encontrei a seguinte documentação... [conteúdo truncado a 5000 chars]"
```

### Configuração de Segurança

#### Domain Whitelist (Opcional)

Para restringir quais domínios podem ser acessados:

```bash
# .env
WEB_FETCH_ALLOWED_DOMAINS="docs.langchain.com,github.com,wikipedia.org"
```

Quando configurado, o assistente só consegue acessar URLs desses domínios.

#### Max Size

Evita token explosion limitando o tamanho da resposta:

```bash
WEB_FETCH_MAX_SIZE=5000  # padrão
```

### Vantagens

- ✓ Extrai conteúdo de páginas
- ✓ Suporta múltiplos formatos
- ✓ Whitelist de domínios para segurança
- ✓ Truncamento automático

### Limitações

- Apenas GET requests (sem POST)
- Timeout se o site for muito lento
- JavaScript executado no navegador não é incluído (apenas HTML estático)
- Requer acess público (não pega conteúdo por trás de login)

### Implementação

```python
from langchain_community.document_loaders import WebBaseLoader

loader = WebBaseLoader(url)
docs = loader.load()
content = docs[0].page_content
```

---

## 3. Database Query (SQL)

### O que faz

Executa queries SQL SELECT contra um banco de dados configurado. Suporta SQLite, PostgreSQL, MySQL, etc.

### Ativação

```bash
ENABLE_DATABASE=false  # disabled por padrão (opt-in)
DATABASE_URL="sqlite:///vectora.db"  # ou postgresql://user:pass@host/db
DATABASE_ALLOWED_TABLES=""  # whitelist (opcional)
```

### Exemplo de Uso (quando habilitada)

```
Usuário: "Quantos usuários ativos temos?"
Assistente: [chama query_database com SELECT COUNT(*) FROM users]
Assistente: "Temos 1250 usuários ativos."
```

### Configuração de Segurança

#### Apenas SELECT Permitido

A tool **bloqueia automaticamente** INSERT, UPDATE, DELETE, DROP, ALTER, CREATE queries.

```python
# Bloqueado ❌
INSERT INTO users VALUES (...)
DELETE FROM users WHERE id=1
ALTER TABLE users ADD column

# Permitido ✓
SELECT * FROM users
SELECT COUNT(*) FROM orders
```

#### Database URL

Exemplos de configuração:

```bash
# SQLite (arquivo local)
DATABASE_URL="sqlite:///vectora.db"

# PostgreSQL
DATABASE_URL="postgresql://user:password@localhost:5432/vectora"

# MySQL
DATABASE_URL="mysql+pymysql://user:password@localhost:3306/vectora"
```

#### Table Whitelist (Opcional)

Restringir acesso a tabelas específicas:

```bash
DATABASE_ALLOWED_TABLES="users,products,orders"
```

### Vantagens

- ✓ Acesso a dados estruturados
- ✓ Suporta múltiplos bancos (SQL genericamente)
- ✓ Bloqueio automático de operações perigosas
- ✓ Whitelist de tabelas para segurança
- ✓ Community-maintained

### Limitações

- Apenas SELECT (por segurança)
- Requer DATABASE_URL configurado
- Performance depende do banco de dados
- Não suporta procedimentos stored

### Instalação

Para usar esta tool, instale as dependências opcionais:

```bash
uv sync --group database
```

### Implementação

```python
from langchain_community.utilities import SQLDatabase

db = SQLDatabase.from_uri(database_url)
result = db.run(sql_query, fetch="all")
```

---

## 4. MCP Server Call (Placeholder)

### O que faz

Prepara a arquitetura para integração com servidores MCP (Model Context Protocol).

**Nota**: Implementação completa do cliente MCP vem em release futura.

### Ativação

```bash
ENABLE_MCP=false  # placeholder, implementação futura
MCP_SERVER_URL=""  # ex: ws://localhost:5000
```

### Exemplo de Uso Futuro

```
Usuário: "Use a ferramenta 'weather' do MCP server para saber o clima em São Paulo"
Assistente: [chamaria call_mcp_tool no futuro]
Assistente: "O clima em São Paulo é..."
```

### Por que apenas Placeholder?

- MCP é um protocolo novo e em evolução
- Requer arquitetura WebSocket no MVP
- Vectora já está arquitetado como um MCP server (via FastAPI + LangGraph)
- Integração bidirecional (Vectora como cliente MCP) é Phase 4+

### Vantagens (quando implementado)

- ✓ Acesso a qualquer ferramenta registrada em MCP servers
- ✓ Composição de agentes
- ✓ Escalabilidade

### Limitações Atuais

- Não implementado no MVP
- Requer cliente MCP do lado de Vectora

---

## Configuração Completa

Exemplo de `.env` com todas as tools configuradas:

```bash
# Logging
LOG_LEVEL="DEBUG"
LOG_JSON="true"
LOG_FILE="logs/vectora.jsonl"

# Web Search (gratuito)
ENABLE_WEB_SEARCH=true

# Web Fetch (com whitelist)
ENABLE_WEB_FETCH=true
WEB_FETCH_MAX_SIZE=5000
WEB_FETCH_ALLOWED_DOMAINS="docs.langchain.com,github.com,wikipedia.org"

# Database (com whitelist)
ENABLE_DATABASE=true
DATABASE_URL="postgresql://user:pass@localhost:5432/vectora"
DATABASE_ALLOWED_TABLES="users,products,orders"

# MCP (futuro)
# ENABLE_MCP=true
# MCP_SERVER_URL="ws://mcp.example.com:5000"
```

---

## Segurança

### Princípios Implementados

1. **Least Privilege**: Todas as tools começam desabilitadas, exceto search/fetch (gratuitas)
2. **Whitelist sobre Blacklist**: Domain whitelist, table whitelist
3. **Input Validation**: URL validation, SQL query validation
4. **Logging Completo**: Todas as ações são registradas com contexto
5. **Resource Limits**: Max size para fetch, max results para search

### Recomendações

- ✓ **Produção**: Habilite DATABASE apenas se necessário, use whitelist
- ✓ **Produção**: Configure WEB_FETCH_ALLOWED_DOMAINS se possível
- ✓ **Produção**: Monitore logs para uso suspeito de tools
- ✓ **Desenvolvimento**: Use SQLite local, domínios conhecidos

---

## Troubleshooting

### Web Search não funciona

```bash
# Erro: "langchain_community not installed"
# Solução:
uv sync
```

### Database tool não aparece

```bash
# Erro: "Database tool is disabled"
# Solução:
ENABLE_DATABASE=true uv run src/main.py
```

### Fetch URL está bloqueado

```bash
# Erro: "Domain X is not in whitelist"
# Solução: Adicione o domínio à whitelist ou deixe vazio
WEB_FETCH_ALLOWED_DOMAINS=""  # permite qualquer domínio
```

### SQLAlchemy não instalado

```bash
# Erro quando tenta usar database
# Solução:
uv sync --group database
```

---

## Próximas Fases

- **Phase 4**: RAG Tools (Retriever, Vector Search)
- **Phase 5**: Semantic Caching com Valkey
- **Phase 6**: MCP Client Implementation
- **Phase 7**: Deep Agents

---

## Referências

- [LangChain Community Tools](https://docs.langchain.com/oss/python/integrations/tools)
- [DuckDuckGoSearchResults](https://python.langchain.com/docs/integrations/tools/duckduckgo)
- [WebBaseLoader](https://python.langchain.com/docs/integrations/document_loaders/web_base)
- [SQLDatabase Toolkit](https://python.langchain.com/docs/integrations/toolkits/sql_database)
- [Model Context Protocol](https://modelcontextprotocol.io)
