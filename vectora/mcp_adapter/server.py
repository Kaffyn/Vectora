"""Vectora MCP Server: Exposes Vectora tools and resources via Model Context Protocol.

This is the production-ready MCP server that Claude Desktop / Claude Code connects to.
It bridges Vectora's internal tools to the MCP protocol using FastMCP.

Transport: stdio JSON-RPC (standard for local MCP servers)
Protocol: MCP (Model Context Protocol)

Entry point: vectora-mcp → vectora.mcp.server:run
"""

import json
import logging
import sys
from datetime import UTC, datetime
from pathlib import Path

# Adiciona vectora/ ao sys.path para resolver imports internos dos módulos.
# Os módulos em vectora/ usam imports diretos (ex: from settings import settings)
# que exigem que o diretório vectora/ esteja no sys.path quando importados via vectora.*.
_vectora_dir = str(Path(__file__).parent.parent)
if _vectora_dir not in sys.path:
    sys.path.insert(0, _vectora_dir)

# Logging to file only — never pollute stdout (JSON-RPC channel)
_log_dir = Path.home() / ".vectora" / "logs"
_log_dir.mkdir(parents=True, exist_ok=True)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler(_log_dir / "mcp.log", encoding="utf-8"),
    ],
)

logger = logging.getLogger("vectora.mcp_adapter.server")

try:
    from mcp.server.fastmcp import FastMCP

    from vectora.checkpointer import Checkpointer
    from vectora.core.agent import AgentManager
    from vectora.settings import settings
    from vectora.tools import (
        call_mcp_tool,
        embedding,
        fetch_url,
        file_edit,
        file_read,
        grep,
        ingest_docs,
        list_dir,
        terminal,
        vector_search,
        web_search,
    )
except ImportError:
    logger.exception("Failed to import Vectora dependencies")
    sys.exit(1)

# ============================================================================
# MCP SERVER INSTANCE
# ============================================================================

mcp = FastMCP(
    "Vectora",
    instructions=(
        "Vectora é um agente de IA avançado com RAG, busca vetorial, "
        "manipulação de arquivos e capacidades de embedding. "
        "Use as ferramentas disponíveis para ajudar com pesquisas, "
        "análise de código e gestão de conhecimento."
    ),
)

logger.info("Vectora MCP server initialized with FastMCP")


# ============================================================================
# TOOLS — Bridge between LangChain tools and MCP protocol
# ============================================================================
# FastMCP automatically converts tool descriptions to MCP tool definitions.
# We wrap each LangChain tool in a @mcp.tool() async function.


@mcp.tool()
async def web_search_tool(query: str) -> str:
    """Busca informações atuais na web.

    Use para obter notícias recentes, dados atualizados ou qualquer
    informação que não esteja no conhecimento base do modelo.

    Args:
        query: Consulta de busca em linguagem natural

    Returns:
        Resultados da busca formatados como texto
    """
    result = await web_search.ainvoke({"query": query})
    return str(result)


@mcp.tool()
async def fetch_url_tool(url: str) -> str:
    """Busca e extrai o conteúdo textual de uma URL.

    Use para ler artigos, documentações ou qualquer página web específica.

    Args:
        url: URL completa para buscar (https://...)

    Returns:
        Conteúdo textual extraído da página
    """
    result = await fetch_url.ainvoke({"url": url})
    return str(result)


@mcp.tool()
async def vector_search_tool(
    query: str,
    collection: str = "default",
    limit: int = 5,
) -> str:
    """Busca documentos semanticamente similares no banco vetorial LanceDB.

    Use para encontrar informações já indexadas via embedding no Vectora.

    Args:
        query: Consulta em linguagem natural
        collection: Nome da coleção (default: "default")
        limit: Número máximo de resultados (default: 5)

    Returns:
        Documentos similares com score de relevância
    """
    result = await vector_search.ainvoke(
        {"query": query, "collection": collection, "limit": limit}
    )
    return str(result)


@mcp.tool()
async def embedding_tool(
    text: str,
    collection: str = "default",
    metadata: dict | None = None,
) -> str:
    """Enfileira um documento para embedding assíncrono no LanceDB.

    Use para indexar novos documentos no banco vetorial do Vectora.
    O embedding é processado em background pelo worker.

    Args:
        text: Texto do documento para indexar
        collection: Coleção de destino (default: "default")
        metadata: Metadados adicionais (title, source, etc.)

    Returns:
        Confirmação de enfileiramento com ID do documento
    """
    args: dict = {"text": text, "collection": collection}
    if metadata:
        args["metadata"] = metadata
    result = await embedding.ainvoke(args)
    return str(result)


@mcp.tool()
async def ingest_docs_tool(
    docs_pattern: str,
    collection: str = "default",
    recursive: bool = True,
) -> str:
    """Ingere múltiplos arquivos em lote no banco vetorial.

    Use para indexar pastas inteiras de documentos de uma vez.

    Args:
        docs_pattern: Padrão glob dos arquivos (ex: "src/**/*.py")
        collection: Coleção de destino (default: "default")
        recursive: Se True, busca recursivamente (default: True)

    Returns:
        Relatório de ingestão com arquivos processados
    """
    result = await ingest_docs.ainvoke(
        {"docs_pattern": docs_pattern, "collection": collection, "recursive": recursive}
    )
    return str(result)


@mcp.tool()
async def file_read_tool(file_path: str) -> str:
    """Lê o conteúdo completo de um arquivo de texto.

    Valida o caminho contra .gitignore para evitar leitura de arquivos sensíveis.

    Args:
        file_path: Caminho absoluto ou relativo do arquivo

    Returns:
        Conteúdo completo do arquivo como string
    """
    result = await file_read.ainvoke({"file_path": file_path})
    return str(result)


@mcp.tool()
async def file_edit_tool(file_path: str, old_text: str, new_text: str) -> str:
    """Edita um arquivo substituindo um trecho de texto por outro.

    Usa correspondência exata de string para localizar e substituir.

    Args:
        file_path: Caminho do arquivo para editar
        old_text: Texto exato a ser substituído
        new_text: Novo texto que substituirá o antigo

    Returns:
        Confirmação da edição realizada
    """
    result = await file_edit.ainvoke(
        {"file_path": file_path, "old_text": old_text, "new_text": new_text}
    )
    return str(result)


@mcp.tool()
async def grep_tool(pattern: str, path: str = ".") -> str:
    """Busca um padrão regex em arquivos.

    Retorna as linhas que correspondem ao padrão com número de linha e arquivo.

    Args:
        pattern: Expressão regular para buscar
        path: Diretório ou arquivo onde buscar (default: ".")

    Returns:
        Linhas correspondentes com contexto (arquivo:linha:conteúdo)
    """
    result = await grep.ainvoke({"pattern": pattern, "path": path})
    return str(result)


@mcp.tool()
async def list_dir_tool(path: str = ".", recursive: bool = False) -> str:
    """Lista arquivos e diretórios em um caminho.

    Args:
        path: Caminho do diretório para listar (default: ".")
        recursive: Se True, lista recursivamente (default: False)

    Returns:
        Lista de arquivos e diretórios com metadados
    """
    result = await list_dir.ainvoke({"path": path, "recursive": recursive})
    return str(result)


@mcp.tool()
async def terminal_tool(command: str) -> str:
    """Executa um comando shell com whitelist de segurança.

    Apenas comandos permitidos pela política de segurança são executados.

    Args:
        command: Comando shell para executar

    Returns:
        Saída do comando (stdout + stderr)
    """
    result = await terminal.ainvoke({"command": command})
    return str(result)


@mcp.tool()
async def call_mcp_tool_tool(tool_name: str, arguments: str) -> str:
    """Invoca uma ferramenta de outro servidor MCP via protocolo MCP.

    Use para encadear chamadas a outros servidores MCP registrados.

    Args:
        tool_name: Nome da ferramenta MCP para invocar
        arguments: Argumentos em formato JSON string

    Returns:
        Resultado da ferramenta MCP invocada
    """
    result = await call_mcp_tool.ainvoke(
        {"tool_name": tool_name, "arguments": arguments}
    )
    return str(result)


logger.info(
    "11 tools registered: web_search, fetch_url, vector_search, embedding, ingest_docs, file_read, file_edit, grep, list_dir, terminal, call_mcp_tool"
)


# ============================================================================
# RESOURCES — Cognitive state of the Vectora agent
# ============================================================================
# Resources expose Vectora's internal state so Claude Code can read context
# before deciding which tool to call.
# Pattern: vectora://<resource>/<id>


@mcp.resource("vectora://thread/{thread_id}/context")
async def get_thread_context(thread_id: str) -> str:
    """Returns the current context and summary of a Vectora conversation thread.

    Allows Claude Code to read the cognitive state of Vectora before
    deciding which tool to call.

    Args:
        thread_id: Thread/conversation ID

    Returns:
        JSON string with context summary
    """
    logger.info("Resource: get_thread_context(%s)", thread_id)

    try:
        async with Checkpointer() as checkpointer:
            config = {"configurable": {"thread_id": str(thread_id)}}
            values = await checkpointer.aget(config)

            if not values:
                return json.dumps(
                    {
                        "thread_id": thread_id,
                        "status": "empty",
                        "message": "No conversation found for this thread",
                        "timestamp": datetime.now(UTC).isoformat(),
                    }
                )

            state = values.get("values", {})
            messages = state.get("messages", [])
            summary = state.get("summarized_history", "")

            return json.dumps(
                {
                    "thread_id": thread_id,
                    "status": "active",
                    "message_count": len(messages),
                    "summary": summary or f"Thread with {len(messages)} messages",
                    "timestamp": datetime.now(UTC).isoformat(),
                }
            )

    except Exception:
        logger.exception("Failed to get thread context: %s", thread_id)
        return json.dumps(
            {"thread_id": thread_id, "status": "error", "error": "Context unavailable"}
        )


@mcp.resource("vectora://thread/{thread_id}/history")
async def get_thread_history(thread_id: str) -> str:
    """Returns the recent message history of a Vectora conversation thread.

    Useful for understanding recent conversation context before calling tools.

    Args:
        thread_id: Thread/conversation ID

    Returns:
        JSON string with last 5 messages
    """
    logger.info("Resource: get_thread_history(%s)", thread_id)

    try:
        async with Checkpointer() as checkpointer:
            config = {"configurable": {"thread_id": str(thread_id)}}
            values = await checkpointer.aget(config)

            if not values:
                return json.dumps(
                    {"thread_id": thread_id, "status": "empty", "messages": []}
                )

            state = values.get("values", {})
            messages = state.get("messages", [])
            recent = messages[-5:] if len(messages) > 5 else messages

            return json.dumps(
                {
                    "thread_id": thread_id,
                    "status": "active",
                    "message_count": len(messages),
                    "recent_messages": [
                        {
                            "type": msg.__class__.__name__,
                            "content": str(msg.content)[:500],
                        }
                        for msg in recent
                    ],
                }
            )

    except Exception:
        logger.exception("Failed to get thread history: %s", thread_id)
        return json.dumps(
            {"thread_id": thread_id, "status": "error", "error": "History unavailable"}
        )


@mcp.resource("vectora://status")
async def get_server_status() -> str:
    """Returns the current status and capabilities of the Vectora MCP server.

    Returns:
        JSON string with server status, version, and active features
    """
    logger.info("Resource: get_server_status")

    return json.dumps(
        {
            "server": "Vectora",
            "version": settings.version,
            "status": "ready",
            "timestamp": datetime.now(UTC).isoformat(),
            "capabilities": {
                "rag_enabled": settings.enable_rag,
                "web_search_enabled": settings.enable_web_search,
                "file_operations_enabled": settings.enable_file_operations,
                "mcp_enabled": settings.enable_mcp,
                "embedding_queue_enabled": settings.embedding_queue_enabled,
            },
            "tools_count": 11,
            "resources_count": 3,
        }
    )


logger.info("3 resources registered: context, history, status")


# ============================================================================
# ENTRY POINT
# ============================================================================


def run() -> None:
    """Start Vectora as a stdio MCP server.

    Entry point: vectora-mcp → vectora.mcp.server:run

    Starts the FastMCP server listening on stdin/stdout using JSON-RPC 2.0.
    All logging is redirected to ~/.vectora/logs/mcp.log to avoid polluting
    the stdio channel.
    """
    logger.info("Starting Vectora MCP server (stdio JSON-RPC)")
    logger.info("Tools: 11 | Resources: 3 | Transport: stdio")

    try:
        mcp.run()
    except KeyboardInterrupt:
        logger.info("Vectora MCP server stopped by user")
        sys.exit(0)
    except Exception:
        logger.exception("Fatal error in Vectora MCP server")
        sys.exit(1)


if __name__ == "__main__":
    run()
