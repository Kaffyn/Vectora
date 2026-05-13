# MCP (Model Context Protocol) Integration Guide

Vectora integra com servidores MCP para expor ferramentas externas ao agente de IA.

## O que é MCP?

**Model Context Protocol** é um padrão aberto para conectar aplicações de IA a sistemas externos. Pense nele como um "USB-C para IA" — fornece uma forma padronizada de conectar modelos de linguagem a dados, ferramentas e workflows externos.

**Referências:**

- [MCP Specification](https://modelcontextprotocol.io/docs/getting-started/intro)
- [FastMCP Framework](https://gofastmcp.com)
- [LangChain MCP Adapters](https://github.com/langchain-ai/langchain-mcp-adapters)

---

## Como Funciona em Vectora

1. **Servidor MCP** — Um processo separado que expõe ferramentas (tools), recursos (resources) ou prompts
2. **Cliente MCP** — Vectora se conecta ao servidor usando HTTP/WebSocket ou stdio
3. **Tool Bridge** — `call_mcp_tool` executa ferramentas do servidor MCP

```
[MCP Server] <---> [Vectora MCP Client] <---> [LLM Agent]
   (expõe           (conecta e                 (usa
    ferramentas)     executa)                  ferramentas)
```

---

## Configuração

### 1. Habilitar MCP

No arquivo `.env`:

```env
ENABLE_MCP=true
```

### 2. Conectar via HTTP (Remoto)

Para um servidor MCP rodando em outro host/porta:

```env
ENABLE_MCP=true
MCP_TRANSPORT_TYPE=http
MCP_SERVER_URL=http://localhost:8000
```

### 3. Conectar via Stdio (Local)

Para um comando local que inicia um servidor MCP:

```env
ENABLE_MCP=true
MCP_TRANSPORT_TYPE=stdio
MCP_COMMAND=npx
MCP_COMMAND_ARGS=chrome-devtools-mcp@latest
```

Para Python:

```env
ENABLE_MCP=true
MCP_TRANSPORT_TYPE=stdio
MCP_COMMAND=python
MCP_COMMAND_ARGS=/path/to/mcp_server.py
```

---

## Usando a Tool `call_mcp_tool`

### Sintaxe

```python
# No agente/chat
result = await call_mcp_tool.ainvoke({
    "tool_name": "nome_da_ferramenta",
    "arguments": '{"param1": "value1", "param2": "value2"}',
    "runtime": runtime
})
```

### Exemplo: Chrome DevTools MCP

Chrome DevTools MCP expõe 26 ferramentas para controlar um navegador Chrome. Exemplos:

```python
# Tirar screenshot
result = await call_mcp_tool.ainvoke({
    "tool_name": "screenshot",
    "arguments": '{}',
    "runtime": runtime
})

# Executar JavaScript
result = await call_mcp_tool.ainvoke({
    "tool_name": "execute_javascript",
    "arguments": '{"script": "document.title"}',
    "runtime": runtime
})

# Clicar em um elemento
result = await call_mcp_tool.ainvoke({
    "tool_name": "click",
    "arguments": '{"selector": "button.submit"}',
    "runtime": runtime
})
```

---

## Instalando Servidores MCP

### Chrome DevTools MCP

```bash
# Via npm (global)
npm install -g chrome-devtools-mcp

# Via npx (sem instalação)
ENABLE_MCP=true
MCP_TRANSPORT_TYPE=stdio
MCP_COMMAND=npx
MCP_COMMAND_ARGS=chrome-devtools-mcp@latest
```

### Criando seu próprio Servidor MCP com FastMCP

```python
from fastmcp import FastMCP

mcp = FastMCP("Meu Servidor")

@mcp.tool
def add(a: int, b: int) -> int:
    """Adiciona dois números"""
    return a + b

@mcp.tool
def multiply(a: int, b: int) -> int:
    """Multiplica dois números"""
    return a * b

if __name__ == "__main__":
    mcp.run()
```

Configure no `.env`:

```env
ENABLE_MCP=true
MCP_TRANSPORT_TYPE=stdio
MCP_COMMAND=python
MCP_COMMAND_ARGS=/path/to/meu_servidor.py
```

---

## Implementação Interna

### Funções Auxiliares (internas)

`src/tools.py` fornece duas funções auxiliares para gerenciar conexões MCP:

```python
async def _get_mcp_client() -> Optional[Any]:
    """Obtém ou cria cliente MCP (cached)."""

async def _get_mcp_tools() -> Optional[dict[str, Any]]:
    """Obtém ferramentas disponíveis do servidor MCP (cached)."""
```

Ambas são **cacheadas** para evitar reconectar a cada chamada.

### Fluxo de Execução

1. **call_mcp_tool** é chamado com `tool_name` e `arguments`
2. Valida se MCP está habilitado e configurado
3. Chama `_get_mcp_tools()` para listar ferramentas disponíveis
4. Busca a ferramenta solicitada na lista
5. Executa via `client.call_tool(tool_name, args)`
6. Retorna resultado como JSON

```json
{
  "status": "success",
  "result": "conteúdo do resultado"
}
```

Em caso de erro:

```json
{
  "status": "error",
  "message": "descrição do erro",
  "available_tools": ["tool1", "tool2"]
}
```

---

## Troubleshooting

### "MCP server integration is not enabled"

Solução: Adicione `ENABLE_MCP=true` ao `.env`

### "Failed to connect to MCP server"

Para HTTP:

- Verifique se `MCP_SERVER_URL` está correto
- Verifique se o servidor está rodando

Para stdio:

- Verifique se `MCP_COMMAND` existe e é executável
- Verifique os argumentos em `MCP_COMMAND_ARGS`

### "Tool 'xyz' not found in MCP server"

Solução:

- Liste as ferramentas disponíveis do servidor
- Verifique se digitou o nome correto

### Erro de validação de argumentos JSON

Solução: Certifique-se de que `arguments` é uma string JSON válida

```python
# ✅ Correto
arguments = json.dumps({"a": 2, "b": 3})

# ❌ Errado
arguments = {"a": 2, "b": 3}  # Não é string
arguments = "{a: 2, b: 3}"    # JSON inválido
```

---

## Testes

Testes para MCP estão em `tests/test_mcp_integration.py`:

```bash
uv run pytest tests/test_mcp_integration.py -v
```

Cobre:

- Conexão HTTP e stdio
- Caching de cliente e ferramentas
- Cenários de erro
- Workflows completos

---

## Próximos Passos

- [ ] Testar com Chrome DevTools MCP em produção
- [ ] Implementar MCP servers customizados para seu domínio
- [ ] Adicionar auditoria/logging de chamadas MCP
- [ ] Integrar com o contexto do usuário para autorização por tool
- [ ] Implementar retry/fallback para falhas de conexão

---

## Referências

- **MCP Spec**: https://modelcontextprotocol.io/docs/getting-started/intro
- **FastMCP**: https://gofastmcp.com
- **Chrome DevTools MCP**: https://github.com/ChromeDevTools/chrome-devtools-mcp
- **LangChain MCP**: https://docs.langchain.com/oss/python/langchain/mcp
