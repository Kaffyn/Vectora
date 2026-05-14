# Vectora como Sub-Agente MCP

## O que é um Sub-Agente?

Vectora não é apenas uma "ferramenta" ou "plugin". É um **agente colaborativo** com:

1. **Raciocínio próprio** (LangGraph internal)
2. **Conhecimento persistente** (SQLite + LanceDB)
3. **Recursos expostos** (estado cognitivo via MCP Resources)
4. **Ferramentas internas** (tools + integration capabilities)

## Comunicação MCP: Stdio JSON-RPC

```
Claude Code (Host Agent)
    ↓
MCP Client (stdin/stdout)
    ↓
Vectora MCP Server (stdio JSON-RPC)
    ↓
LangGraph (raciocínio interno)
    ↓
Tools (vector_search, web_search, file_read, etc)
    ↓
Response (resultado sintético)
```

**Importante:** Não é HTTP/REST. É comunicação de processo via stdin/stdout (JSON-RPC padrão).

## Arquitetura de Tools + Resources

### Tools (Capacidades)

O que o Vectora consegue **fazer**:

```python
@mcp.tool()
async def search_knowledge_base(query: str) -> str:
    """Busca conhecimento técnico profundo no banco de dados local.
    Use para: arquitetura, padrões de código, decisões técnicas anteriores.
    ESSENCIAL para: entender projetos legados, revisar código."""
    return await vector_search(query)

@mcp.tool()
async def search_web_real_time(query: str) -> str:
    """Busca em tempo real na internet (DuckDuckGo).
    Use para: bibliotecas novas, status de projetos, informações atuais."""
    return await web_search(query)

@mcp.tool()
async def read_local_file(path: str) -> str:
    """Lê arquivo do disco (com validação de path).
    Use para: ler código fonte, documentação local, configurações."""
    return await file_read(path)

# ... e 7 outras tools
```

### Resources (Estado Cognitivo)

O que o Claude Code consegue **ler** sobre o Vectora:

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
Claude Code lê Context Resource
↓
"Vectora já tem conhecimento sobre X"
↓
Claude Code decide: Chamar vector_search ou fazer web_search?
↓
Chama tool apropriado
↓
Integra resultado com conhecimento anterior do Vectora
↓
Resposta enriquecida
```

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

## Exemplo: Conversa entre Claude Code e Vectora

### Passo 1: Claude Code lê o estado

```
Claude Code: "Qual é o contexto atual do Vectora?"
↓
Lê: vectora://thread/123/context
↓
Vectora responde: "Thread tem conhecimento sobre:
- Arquitetura de RAG em LangGraph
- Implementação de vector search em LanceDB
- Padrões de error handling"
```

### Passo 2: Claude Code decide

```
Claude Code pensa:
"O usuário perguntou sobre cache distribuído.
Vectora já sabe sobre RAG + vector search,
mas não menciona cache. Devo:
1. Fazer web_search para info atualizada sobre Redis
2. Depois perguntar ao Vectora como integrar com cache"
```

### Passo 3: Claude Code executa

```
Claude Code: chamar search_web_real_time("Redis distributed cache 2026")
↓
Vectora executa: web_search("Redis distributed cache 2026")
↓
Retorna: "Redis 7.2 suporta X, custo é Y, alternativas são Z"
```

### Passo 4: Claude Code sintetiza

```
Claude Code integra:
- Conhecimento anterior do Vectora (RAG, LanceDB)
- Novo conhecimento (Redis features)
- Contexto da pergunta
↓
Resposta ao usuário:
"Baseado no conhecimento do Vectora + pesquisa realizada,
recomendo: use Redis para cache, porque..."
```

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
