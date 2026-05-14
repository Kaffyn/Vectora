# Vectora como Sub-Agente MCP

## O que é um Sub-Agente?

Vectora não é apenas uma "ferramenta" ou "plugin". É um **agente colaborativo independente** com:

1. **Raciocínio próprio** (LangGraph interno)
2. **Conhecimento persistente** (SQLite + LanceDB)
3. **Estado cognitivo exposto** (3 Resources via MCP)
4. **Ferramentas internas** (11 tools usadas internamente pelo LangGraph)

**Diferença crítica:** Claude Code NÃO chama as tools do Vectora diretamente. Claude Code comunica-se COM o Vectora, e o Vectora internamente executa suas ferramentas com seu próprio raciocínio.

## Comunicação MCP: Stdio JSON-RPC (Agent-to-Agent)

```
Claude Code (Agente Host)
    ↓ (lê estado)
MCP Client: GET vectora://thread/123/context
    ↓
Vectora MCP Server ← (processa via LangGraph)
    ├─ MAIN_NODE
    ├─ TOOL_NODE (executa tools internas)
    ├─ SUMMARIZER_NODE
    └─ SUB_NODE
    ↓ (responde)
Resultado processado → Claude Code
```

**Ponto-chave:** O MCP **não expõe as 11 tools diretamente**. Expõe apenas:
- **3 Resources**: estado cognitivo do Vectora (context, history, status)
- **Protocolo JSON-RPC**: comunicação entre agentes via stdio

Não é HTTP/REST. É comunicação de processo (stdin/stdout) entre dois agentes LLM.

## O que é Exposto via MCP?

### Resources (Único expostos via MCP)

O que Claude Code consegue **ler** sobre o estado do Vectora:

**Obs:** As 11 ferramentas (web_search, vector_search, file_read, etc) são **internas do Vectora**. Claude Code não as chama diretamente. O Vectora as executa internamente quando processa requisições.

```python
@mcp.resource("vectora://thread/{thread_id}/context")
async def get_thread_context(thread_id: str) -> Resource:
    """Resumo do conhecimento coletado nesta thread.
    Permite ao Claude Code saber o que o Vectora já aprendeu."""
    summary = await get_summary_from_sqlite(thread_id)
    return Resource(
        uri=f"vectora://thread/{thread_id}/context",
        mime_type="text/plain",
        text=summary  # "Já coletei X docs sobre Python async..."
    )

@mcp.resource("vectora://thread/{thread_id}/history")
async def get_thread_history(thread_id: str) -> Resource:
    """Histórico das últimas 5 mensagens da thread."""
    history = await get_history_from_sqlite(thread_id)
    return Resource(
        uri=f"vectora://thread/{thread_id}/history",
        mime_type="application/json",
        text=json.dumps(history)
    )

@mcp.resource("vectora://status")
async def get_status() -> Resource:
    """Status do servidor (LLM provider, RAG, uptime)."""
    return Resource(
        uri="vectora://status",
        mime_type="application/json",
        text=json.dumps({
            "llm_provider": "google-genai",
            "rag_enabled": True,
            "vector_store": "lancedb",
            "uptime_seconds": 3600
        })
    )
```

## Fluxo de Decisão do Claude Code

```
Claude Code lê Resource: vectora://thread/123/context
↓
"Vectora já sabe sobre: RAG, LanceDB, vector_search"
↓
Claude Code pensa: "Preciso de mais contexto. O que Vectora sabe sobre Redis?"
↓
Claude Code COMUNICA COM Vectora (via MCP JSON-RPC):
"Preciso saber sobre Redis distributed cache"
↓
Vectora internamente:
  ├─ MAIN_NODE processa requisição
  ├─ TOOL_NODE: executa web_search() e vector_search() internamente
  ├─ SUMMARIZER_NODE: resume conhecimento
  └─ Retorna resposta processada ao Claude Code
↓
Claude Code recebe: "Conhecimento coletado + contexto enriquecido"
↓
Claude Code sintetiza resposta ao usuário
```

**Ponto crucial:** Claude Code comunica COM Vectora, não com as tools individuais.

## Configuração no Claude Code

### `claude_config.json`

```json
{
  "mcpServers": {
    "vectora": {
      "command": "python",
      "args": ["-m", "mcp.server", "src/mcp_server.py"],
      "env": {
        "GOOGLE_API_KEY": "sk-...",
        "VOYAGE_API_KEY": "pa-...",
        "LANGSMITH_API_KEY": "ls-..."
      }
    }
  }
}
```

### Inicialização

```bash
# Vectora como standalone MCP Server
python -m mcp.server src/mcp_server.py

# Ou via Docker
docker run -e GOOGLE_API_KEY=xxx vectora:0.1.0 mcp-server
```

## Exemplo: Comunicação entre Claude Code e Vectora

### Passo 1: Claude Code lê o estado do Vectora

```
Claude Code: GET vectora://thread/123/context (via MCP)
↓
Vectora responde (Resource):
{
  "status": "active",
  "summary": "Thread tem conhecimento sobre:
    - Arquitetura de RAG em LangGraph
    - Implementação de vector search em LanceDB
    - Padrões de error handling"
}
```

### Passo 2: Claude Code toma decisão informada

```
Claude Code pensa:
"Usuário perguntou sobre cache distribuído.
Vectora sabe sobre RAG + vector search,
mas não menciona cache. Devo comunicar com Vectora
para buscar info sobre Redis e integrar com seu conhecimento."
```

### Passo 3: Claude Code comunica COM Vectora (não com as tools)

```
Claude Code: "Vectora, preciso que você pesquise
e contextualize: como usar Redis com vector search?"

(Enviado via MCP JSON-RPC)
↓
Vectora recebe requisição
↓
Vectora internamente (LangGraph):
  1. MAIN_NODE: interpreta "pesquise sobre Redis"
  2. TOOL_NODE: executa internally
     ├─ web_search("Redis with vector search 2026")
     └─ vector_search("cache patterns") em seu LanceDB
  3. SUMMARIZER_NODE: compila conhecimento
  4. Retorna resposta ao Claude Code
↓
Vectora responde: "Baseado em pesquisa + conhecimento local:
  - Redis 7.2 suporta X
  - Integração com vector search recomenda Y
  - Exemplos: Z"
```

### Passo 4: Claude Code sintetiza

```
Claude Code recebe resposta processada do Vectora
↓
Claude Code integra com seu próprio conhecimento
↓
Resposta final ao usuário:
"Recomendo integrar Redis com vector search, porque:
- Vectora já usa LanceDB para embeddings
- Redis pode cacheizar resultados
- Melhora latência em X%"
```

**Obs importante:** Em nenhum momento Claude Code chamou `web_search()` ou `vector_search()` diretamente. Essas são **ferramentas internas do Vectora**. Claude Code apenas comunicou uma requisição de alto nível ao Vectora.

## Por que Resources são fundamentais

**Sem Resources** (modelo antigo):

```
Claude Code: "Vector_search por X"
↓
Vectora: "Aqui estão 5 resultados"
↓
Claude Code: Sem contexto, faz outra busca que poderia ser evitada
```

**Com Resources** (modelo novo):

```
Claude Code lê: vectora://thread/123/context
↓
Claude Code vê: "Já tenho conhecimento de X, devo buscar Y?"
↓
Claude Code: Decisão informada sobre qual tool chamar
↓
Mais eficiente, menos buscas redundantes
```

## Implantação

### Docker

```dockerfile
# MCP Server rodando em stdio
CMD ["python", "-m", "mcp.server", "src/mcp_server.py"]
```

### Claude Code Local

```bash
# Rodar Vectora como sub-agente
docker run -d \
  -e GOOGLE_API_KEY=sk-... \
  -v ~/.vectora:/root/.vectora \
  --name vectora-mcp \
  vectora:0.1.0 mcp-server

# Configurar claude_config.json apontando para este container
```

### Paperclip Integration

```json
{
  "agents": {
    "vectora": {
      "type": "mcp",
      "command": "docker",
      "args": ["exec", "vectora-mcp", "python", "-m", "mcp.server"],
      "capabilities": ["vector_search", "web_search", "file_operations"]
    }
  }
}
```

## Diferenciais vs REST API

| Aspecto        | REST API             | MCP Sub-Agent       |
| -------------- | -------------------- | ------------------- |
| Protocolo      | HTTP                 | stdio JSON-RPC      |
| Latência       | rede (10-100ms)      | processo (1-5ms)    |
| Resources      | Não                  | Sim (state sharing) |
| Raciocínio     | Delegação simples    | LangGraph interno   |
| Escalabilidade | Múltiplas instâncias | Single process      |
| Deployment     | Servidor separado    | Em-processo         |
| Ideal para     | Microserviços        | Agentes locais      |

## Próximas Melhorias (Post-MVP)

- [ ] Streaming de respostas (Resources muito grandes)
- [ ] Paginated Resources (thread history > 5 msgs)
- [ ] Custom Resource types (análise semântica, reranking)
- [ ] Prompt templates integrados (hints para o Claude Code)

## Referências

- [MCP Protocol Spec](https://spec.modelcontextprotocol.io/)
- [MCP Resources](https://spec.modelcontextprotocol.io/latest/basic/resources/)
- [FastMCP (Python impl)](https://docs.modelcontextprotocol.io/tutorials/build-a-fastmcp-server)
