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
# Os módulos em vectora/ usam imports diretos (ex: from vectora.config.settings import settings)
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

logger = logging.getLogger("vectora.mcp.server")

try:
    from mcp.server.fastmcp import FastMCP

    from vectora.agent import AgentManager
    from vectora.config.settings import settings
    from vectora.services.checkpoint import Checkpointer
    from vectora.tools import (
        call_mcp_tool,
        embedding,
        fetch_url,
        file_edit,
        file_read,
        file_write,
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
        "manipulação de arquivos e capacidades de embedding.\n\n"
        "QUANDO USAR CADA FERRAMENTA:\n\n"
        "🌐 WEB SEARCH (web_search_tool):\n"
        "- Buscar informações ATUAIS e em tempo real da internet\n"
        "- Quando o usuário pergunta sobre notícias, eventos recentes, dados de hoje\n"
        "- NÃO usar para informações já indexadas no Vectora\n\n"
        "📚 VECTOR SEARCH (vector_search_tool):\n"
        "- Buscar documentos/conhecimento JÁ INDEXADO no Vectora\n"
        "- Busca semântica em base de conhecimento persistente\n"
        "- Usa reranking automático para melhor relevância\n"
        "- Preferir sobre web_search quando informação já foi ingerida\n\n"
        "📄 FILE TOOLS (file_read, file_edit, file_write):\n"
        "- file_read: Ler arquivos do disco (projetos, docs, código)\n"
        "- file_edit: Editar trecho específico com replace_all para múltiplas ocorrências\n"
        "- file_write: Criar novo arquivo ou sobrescrever completo\n\n"
        "⚙️ EMBEDDING (embedding_tool):\n"
        "- Indexar novo documento no Vectora para busca semântica futura\n"
        "- Fire-and-forget: retorna queue_id, processa em background\n"
        "- Usar ANTES de vector_search se novo conteúdo foi adicionado\n\n"
        "🔍 INGEST (ingest_docs_tool):\n"
        "- Indexar pasta inteira de documentos de uma vez\n"
        "- Ideal para popular conhecimento base (docs, wiki, code)\n"
        "- Processa em batch com splitting automático\n\n"
        "🌐 FETCH URL (fetch_url_tool):\n"
        "- Extrair conteúdo textual de UMA URL específica\n"
        "- Use quando precisa ler um artigo/doc específico\n"
        "- Melhor que web_search quando você já sabe a URL\n\n"
        "FLUXO RECOMENDADO:\n"
        "1. Entender a pergunta do usuário\n"
        "2. SE é sobre algo em tempo real → web_search_tool\n"
        "3. SE é sobre conhecimento já indexado → vector_search_tool\n"
        "4. SE precisa ler um arquivo → file_read\n"
        "5. SE quer indexar novo conteúdo → embedding_tool ou ingest_docs_tool\n"
        "6. Sempre verificar resources primeiro: /vectora/thread/{id}/history"
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
async def file_write_tool(file_path: str, content: str) -> str:
    """Cria ou sobrescreve completamente um arquivo com o conteúdo fornecido.

    Use para criar novos arquivos ou substituir o conteúdo completo de um existente.
    Para edições cirúrgicas (substituir apenas um trecho), prefira file_edit_tool.

    Args:
        file_path: Caminho do arquivo (absoluto ou relativo)
        content: Conteúdo completo a escrever no arquivo

    Returns:
        Confirmação com caminho e tamanho em bytes
    """
    result = await file_write.ainvoke({"file_path": file_path, "content": content})
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


@mcp.tool()
async def delegate_task_to_vectora(
    task_prompt: str,
    thread_id: int = 1,
) -> str:
    """Delega uma tarefa complexa para o motor de raciocínio do Vectora.

    Use esta ferramenta quando a tarefa exigir múltiplas etapas de RAG,
    análise de arquivos, busca web ou raciocínio que o Agent Principal
    não deve gerenciar sozinho.

    O Vectora executará seu LangGraph interno e retornará apenas o resultado final.
    Esta é a ferramenta de "delegação de sub-agente" - o Vectora processa
    tudo internamente (RAG, file ops, web search) e devolve resultado processado.

    Args:
        task_prompt: Descrição completa da tarefa/pergunta para o Vectora processar
        thread_id: ID da sessão/conversa para manter contexto (default: 1)

    Returns:
        Resultado final após processar a tarefa completamente no LangGraph do Vectora
    """
    logger.info(
        "Delegação recebida do Agent Principal",
        extra={"thread_id": thread_id, "prompt_length": len(task_prompt)},
    )

    try:
        from vectora.agent import AgentManager
        from vectora.config.settings import settings as mcp_settings

        # Inicializar AgentManager com as settings
        agent = AgentManager(mcp_settings)
        await agent.initialize()

        # Executar a tarefa via o LangGraph interno do Vectora
        # Isso dispara todo o pipeline: RAG, tools, reasoning, etc.
        resultado = await agent.chat(
            user_input=task_prompt,
            session_id=thread_id,
        )

        logger.info(
            "Delegação processada com sucesso",
            extra={
                "thread_id": thread_id,
                "result_length": len(str(resultado)),
            },
        )

        return resultado

    except Exception as e:
        logger.exception(
            "Falha na delegação de sub-agente",
            extra={"thread_id": thread_id},
        )
        return (
            f"Erro ao processar tarefa delegada no Vectora: {str(e)}\n\n"
            f"Por favor, tente novamente ou quebre a tarefa em partes menores."
        )


logger.info(
    "13 tools registered: web_search, fetch_url, vector_search, embedding, ingest_docs, "
    "file_read, file_edit, file_write, grep, list_dir, terminal, call_mcp_tool, "
    "delegate_task_to_vectora"
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
            "tools_count": 13,
            "resources_count": 4,
        }
    )


@mcp.resource("vectora://collections")
async def list_vector_collections() -> str:
    """Returns available vector search collections in LanceDB.

    Useful for understanding what knowledge bases are indexed and ready for search.

    Returns:
        JSON string with list of collections and their status
    """
    logger.info("Resource: list_vector_collections")

    try:
        import lancedb

        if lancedb is None or settings.lancedb_dir is None:
            return json.dumps(
                {"status": "unavailable", "reason": "LanceDB not configured"}
            )

        db = await lancedb.connect_async(str(settings.lancedb_dir))
        table_names = await db.table_names()

        collections = []
        for table_name in table_names:
            try:
                table = await db.open_table(table_name)
                count = await table.count_rows()
                collections.append(
                    {"name": table_name, "documents": count, "status": "ready"}
                )
            except Exception as e:
                logger.warning(f"Error reading collection {table_name}: {e}")
                collections.append(
                    {"name": table_name, "documents": 0, "status": "error"}
                )

        return json.dumps(
            {
                "status": "success",
                "collections_count": len(collections),
                "collections": collections,
                "timestamp": datetime.now(UTC).isoformat(),
            }
        )

    except Exception:
        logger.exception("Failed to list vector collections")
        return json.dumps({"status": "error", "error": "Unable to list collections"})


logger.info("4 resources registered: context, history, status, collections")


# ============================================================================
# ENTRY POINT
# ============================================================================


def run() -> None:
    """Start Vectora as a stdio MCP server.

    Entry point: vectora-mcp → vectora.mcp.server:run

    Starts the FastMCP server listening on stdin/stdout using JSON-RPC 2.0.
    All logging is redirected to ~/.vectora/logs/mcp.log to avoid polluting
    the stdio channel. Status feedback goes to stderr (safe for MCP clients).
    """
    from rich.console import Console
    from rich.panel import Panel

    # stderr é seguro — o protocolo MCP usa apenas stdout para JSON-RPC
    err_console = Console(stderr=True)
    err_console.print(
        Panel(
            "[bold green]✓ Vectora MCP Server pronto[/bold green]\n"
            "[dim]Transport:[/dim] stdio JSON-RPC  "
            "[dim]Tools:[/dim] 13  [dim]Resources:[/dim] 4\n"
            f"[dim]Logs:[/dim] {_log_dir / 'mcp.log'}",
            title="[bold cyan]Vectora MCP[/bold cyan]",
            border_style="cyan",
        )
    )

    logger.info("Starting Vectora MCP server (stdio JSON-RPC)")
    logger.info("Tools: 13 | Resources: 4 | Transport: stdio")

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
