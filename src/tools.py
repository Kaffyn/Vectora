import json
import logging
from datetime import UTC, datetime
from typing import Any

from langchain.tools import BaseTool, tool
from langchain_community.document_loaders import DirectoryLoader, TextLoader, WebBaseLoader

try:
    from langchain_community.tools.duckduckgo_search import DuckDuckGoSearchResults  # type: ignore
except ImportError:
    DuckDuckGoSearchResults = None

try:
    from langchain_mcp_adapters import MultiServerMCPClient  # type: ignore
except ImportError:
    MultiServerMCPClient = None

from embedding_queue import get_embedding_queue
from tool_config import get_tool_config

logger = logging.getLogger(__name__)

# Global MCP client cache (reuse connection across calls)
_mcp_client: Any | None = None
_mcp_tools_cache: dict[str, Any] | None = None


@tool
def web_search(query: str) -> str:
    """Busca a web por informações atuais usando DuckDuckGo.

    Args:
        query: String da consulta de busca

    Returns:
        Resultados da busca como string formatada com URLs e snippets
    """
    if DuckDuckGoSearchResults is None:
        return "DuckDuckGo search module not available. Install: pip install duckduckgo-search"

    config = get_tool_config()

    if not config.enable_web_search:
        logger.warning("web_search tool called but disabled")
        return "Web search is disabled. Enable ENABLE_WEB_SEARCH=true to use this tool."

    logger.info("web_search tool called", extra={"query": query})

    try:
        searcher = DuckDuckGoSearchResults(max_results=5)
        results = searcher.run(query)

        logger.info(
            "web_search completed",
            extra={"query": query, "result_length": len(str(results))},
        )

        return results
    except Exception:
        logger.exception(
            "web_search failed",
            extra={"query": query},
        )
        return "Error occurred. Please check logs."


@tool
async def fetch_url(url: str) -> str:
    """Busca e extrai conteúdo de texto de uma URL específica.

    Args:
        url: URL para buscar (deve começar com http:// ou https://)

    Returns:
        Conteúdo de texto extraído da página
    """
    config = get_tool_config()

    if not config.enable_web_fetch:
        logger.warning("fetch_url tool called but disabled")
        return "Web fetch is disabled. Enable ENABLE_WEB_FETCH=true to use this tool."

    if not url.startswith(("http://", "https://")):
        logger.warning("fetch_url called with invalid URL", extra={"url": url})
        return "Error: URL must start with http:// or https://"

    if config.allowed_domains:
        from urllib.parse import urlparse

        domain = urlparse(url).netloc
        if domain not in config.allowed_domains:
            logger.warning(
                "fetch_url blocked by domain whitelist",
                extra={"url": url, "domain": domain},
            )
            return f"Error: Domain {domain} is not in whitelist"

    logger.info("fetch_url tool called", extra={"url": url})

    try:
        loader = WebBaseLoader(url)
        docs = loader.load()

        logger.info(
            "fetch_url completed",
            extra={"url": url, "docs_count": len(docs)},
        )

        return "\n".join(doc.page_content for doc in docs)

    except Exception:
        logger.exception(
            "fetch_url failed",
            extra={"url": url},
        )
        return "Error occurred fetching URL. Please check logs."


async def _get_mcp_client() -> Any | None:
    """Obtém ou cria instância global do cliente MCP."""
    global _mcp_client

    if _mcp_client is not None:
        return _mcp_client

    if MultiServerMCPClient is None:
        logger.warning("MultiServerMCPClient not available")
        return None

    config = get_tool_config()

    if not config.mcp_server_url and not config.mcp_command:
        logger.debug("No MCP servers configured")
        return None

    try:
        # Constrói dicionário de servidores com base nos campos reais do ToolConfig
        servers: dict[str, Any] = {}
        if config.mcp_server_url:
            servers["default"] = {
                "url": config.mcp_server_url,
                "transport": config.mcp_transport_type,
            }
        if config.mcp_command:
            servers["local"] = {
                "command": config.mcp_command,
                "args": config.mcp_command_args or [],
                "transport": "stdio",
            }

        _mcp_client = MultiServerMCPClient(servers=servers)
        await _mcp_client.__aenter__()
        logger.info("MCP client initialized", extra={"servers": list(servers.keys())})
        return _mcp_client
    except Exception:
        logger.exception("Failed to initialize MCP client")
        _mcp_client = None
        return None


async def _get_mcp_tools() -> dict[str, Any] | None:
    """Obtém ferramentas MCP disponíveis do cliente inicializado."""
    global _mcp_tools_cache

    if _mcp_tools_cache is not None:
        return _mcp_tools_cache

    client = await _get_mcp_client()
    if client is None:
        return None

    try:
        tools_response = await client.list_tools()
        _mcp_tools_cache = {tool.name: tool for tool in tools_response.tools}
        logger.info("MCP tools loaded", extra={"count": len(_mcp_tools_cache)})
        return _mcp_tools_cache
    except Exception:
        logger.exception("Failed to list MCP tools")
        _mcp_tools_cache = {}
        return _mcp_tools_cache


@tool
async def call_mcp_tool(tool_name: str, arguments: str) -> str:
    """Chama uma ferramenta MCP (Model Context Protocol).

    Args:
        tool_name: Nome da ferramenta MCP a chamar
        arguments: String JSON com argumentos da ferramenta

    Returns:
        Resultado da execução da ferramenta MCP
    """
    config = get_tool_config()

    if not config.enable_mcp:
        logger.warning("call_mcp_tool called but MCP disabled")
        return "MCP is disabled. Enable ENABLE_MCP=true to use this tool."

    client = await _get_mcp_client()
    if client is None:
        return "MCP client not available"

    available_tools = await _get_mcp_tools()
    if not available_tools or tool_name not in available_tools:
        return f"Tool '{tool_name}' not found in MCP tools"

    try:
        result = await client.call_tool(tool_name, json.loads(arguments))
        logger.info(
            "mcp_tool_called",
            extra={"tool_name": tool_name, "success": True},
        )
        return json.dumps(result)
    except Exception:
        logger.exception(
            "mcp_tool_call_failed",
            extra={"tool_name": tool_name},
        )
        return f"Error calling MCP tool '{tool_name}'"


@tool
async def embedding(
    text: str, collection: str = "articles", metadata: dict[str, Any] | None = None
) -> str:
    """Indexa um documento com embedding no banco de dados vetorial LanceDB.

    Gera embeddings usando Voyage AI e indexa na coleção (tabela) especificada.
    LanceDB é file-based — nenhum container ou servidor é necessário.

    Args:
        text: Texto do documento a indexar
        collection: Nome da coleção/tabela LanceDB (articles, wiki, api_docs, knowledge_base)
        metadata: Dicionário opcional de metadados (source, author, timestamp, etc)

    Returns:
        String JSON com status: indexed/failed com doc_id ou mensagem de erro
    """
    from uuid import uuid4

    config = get_tool_config()

    if not config.enable_rag:
        logger.warning("embedding tool called but RAG disabled")
        return "RAG is disabled. Enable ENABLE_RAG=true to use this tool."

    if not config.voyage_api_key:
        logger.error("embedding called but VOYAGE_API_KEY not configured")
        return json.dumps(
            {
                "status": "failed",
                "error": "VOYAGE_API_KEY not configured",
                "collection": collection,
            }
        )

    try:
        import lancedb
        import pyarrow as pa
        from langchain_voyageai import VoyageAIEmbeddings
        from pydantic import SecretStr

        embeddings_model = VoyageAIEmbeddings(
            api_key=SecretStr(config.voyage_api_key),
            model=config.embedding_model,
        )

        vector = embeddings_model.embed_query(text)
        doc_id = str(uuid4())

        # Conecta ao banco LanceDB local (cria diretório se não existir)
        db = await lancedb.connect_async(str(config.lancedb_path))

        schema = pa.schema(
            [
                pa.field("id", pa.string()),
                pa.field("vector", pa.list_(pa.float32(), len(vector))),
                pa.field("text", pa.string()),
                pa.field("metadata", pa.string()),  # JSON serializado
            ]
        )

        try:
            table = await db.open_table(collection)
        except Exception:
            table = await db.create_table(collection, schema=schema)
            logger.info("lancedb_table_created", extra={"collection": collection})

        await table.add(
            [
                {
                    "id": doc_id,
                    "vector": vector,
                    "text": text,
                    "metadata": json.dumps(metadata or {}),
                }
            ]
        )

        logger.info(
            "embedding_indexed",
            extra={
                "collection": collection,
                "doc_id": doc_id,
                "text_length": len(text),
            },
        )

        return json.dumps(
            {
                "status": "indexed",
                "doc_id": doc_id,
                "collection": collection,
            }
        )

    except Exception as e:
        logger.exception(
            "embedding_failed",
            extra={"collection": collection, "text_length": len(text)},
        )

        # Fallback para fila de processamento assíncrono se configurado
        if config.embedding_queue_enabled:
            try:
                queue = await get_embedding_queue(config.embedding_queue_db)
                queue_id = await queue.enqueue(text, collection, metadata)
                logger.info(
                    "embedding_enqueued",
                    extra={"queue_id": queue_id, "collection": collection},
                )
                return json.dumps(
                    {
                        "status": "enqueued",
                        "queue_id": queue_id,
                        "collection": collection,
                        "message": "API failed, document enqueued for retry",
                    }
                )
            except Exception as queue_err:
                logger.error(f"Failed to enqueue embedding: {queue_err}")

        return json.dumps(
            {
                "status": "failed",
                "error": str(e) or "Embedding indexing failed",
                "collection": collection,
            }
        )


@tool
async def vector_search(
    query: str, collection: str = "articles", limit: int = 5
) -> str:
    """Busca o banco de dados vetorial LanceDB por documentos similares.

    LanceDB é file-based — nenhum container ou servidor é necessário.

    Args:
        query: String da consulta de busca
        collection: Nome da tabela LanceDB
        limit: Número máximo de resultados a retornar

    Returns:
        Resultados da busca em formato JSON com documentos e scores
    """
    config = get_tool_config()

    if not config.enable_rag:
        logger.warning("vector_search tool called but RAG disabled")
        return "RAG is disabled. Enable ENABLE_RAG=true to use this tool."

    try:
        import lancedb
        from langchain_voyageai import VoyageAIEmbeddings
        from pydantic import SecretStr

        if not config.voyage_api_key:
            logger.error("vector_search called but VOYAGE_API_KEY not configured")
            return json.dumps(
                {
                    "status": "failed",
                    "error": "VOYAGE_API_KEY not configured",
                }
            )

        embeddings_model = VoyageAIEmbeddings(
            api_key=SecretStr(config.voyage_api_key),
            model=config.embedding_model,
        )

        query_vector = embeddings_model.embed_query(query)

        db = await lancedb.connect_async(str(config.lancedb_path))

        try:
            table = await db.open_table(collection)
        except Exception:
            logger.warning(
                "LanceDB table not found",
                extra={"collection": collection},
            )
            return json.dumps(
                {
                    "status": "no_results",
                    "message": f"Collection '{collection}' not found or empty",
                }
            )

        search_results = await (
            table.vector_search(query_vector).limit(limit).to_pandas()
        )

        results = [
            {
                "id": str(row["id"]),
                "score": float(row.get("_distance", 0.0)),
                "content": row["text"],
                "metadata": json.loads(row.get("metadata", "{}")),
            }
            for _, row in search_results.iterrows()
        ]

        # Reranking (opcional, melhora precisão)
        if results and config.reranker_type == "voyage":
            try:
                from langchain_voyageai import VoyageAIRerank

                reranker = VoyageAIRerank(
                    api_key=SecretStr(config.voyage_api_key),
                    model=config.reranker_model,
                    top_k=config.reranker_top_k,
                )

                # Formata para o reranker (LangChain espera Document objects)
                from langchain_core.documents import Document as LCDoc

                docs_to_rerank = [
                    LCDoc(page_content=str(r["content"]), metadata=r["metadata"])
                    for r in results
                ]

                reranked_docs = reranker.compress_documents(docs_to_rerank, query)

                # Reconstrói lista de resultados com base no rerank
                results = [
                    {
                        "content": doc.page_content,
                        "metadata": doc.metadata,
                        "relevance_score": getattr(doc, "relevance_score", None),
                    }
                    for doc in reranked_docs
                ]
                logger.info("vector_search_reranked", extra={"top_k": len(results)})
            except Exception as rerank_err:
                logger.warning(f"Reranking failed, using raw results: {rerank_err}")

        logger.info(
            "vector_search completed",
            extra={
                "query": query,
                "collection": collection,
                "result_count": len(results),
            },
        )

        return json.dumps(
            {
                "status": "success",
                "results": results,
                "query": query,
                "collection": collection,
            }
        )

    except Exception as e:
        logger.exception(
            "vector_search_failed",
            extra={"query": query, "collection": collection},
        )
        return json.dumps(
            {
                "status": "failed",
                "error": str(e) or "Vector search failed",
            }
        )


@tool
def file_read(file_path: str) -> str:
    """Lê conteúdo completo de um arquivo de texto.

    Args:
        file_path: Caminho relativo ou absoluto do arquivo

    Returns:
        Conteúdo do arquivo como string
    """
    from pathlib import Path

    from tool_safety import is_safe_file_path

    if not is_safe_file_path(file_path):
        return f"Access denied: {file_path} is outside allowed directory"

    path = Path(file_path)
    if not path.exists():
        return f"File not found: {file_path}"

    return path.read_text(encoding="utf-8")


@tool
async def ingest_docs(
    directory_path: str,
    collection: str = "articles",
    glob_pattern: str = "**/*.md",
) -> str:
    """Lê todos os arquivos de um diretório e os indexa no banco vetorial.

    Útil para popular o conhecimento do agente com documentação local.

    Args:
        directory_path: Caminho da pasta contendo os documentos
        collection: Nome da coleção LanceDB de destino
        glob_pattern: Padrão de busca de arquivos (ex: **/*.md, **/*.txt)

    Returns:
        Status da ingestão com contagem de sucessos/falhas
    """
    from pathlib import Path

    from langchain_text_splitters import RecursiveCharacterTextSplitter

    from tool_safety import is_safe_file_path

    if not is_safe_file_path(directory_path):
        return f"Access denied: {directory_path} is outside allowed directory"

    path = Path(directory_path)
    if not path.is_dir():
        return f"Not a directory: {directory_path}"

    loader = DirectoryLoader(
        str(path),
        glob=glob_pattern,
        loader_cls=TextLoader,
        loader_kwargs={"encoding": "utf-8"},
    )

    docs = loader.load()
    if not docs:
        return f"No documents found in {directory_path} matching {glob_pattern}"

    # Splitter para documentos grandes
    splitter = RecursiveCharacterTextSplitter(chunk_size=1000, chunk_overlap=100)
    chunks = splitter.split_documents(docs)

    success_count = 0
    fail_count = 0

    for chunk in chunks:
        # Metadados enriquecidos para rastreabilidade
        chunk_metadata = chunk.metadata or {}
        chunk_metadata.update(
            {"ingested_at": datetime.now(UTC).isoformat(), "source_dir": directory_path}
        )

        res = await embedding.ainvoke(
            {
                "text": chunk.page_content,
                "collection": collection,
                "metadata": chunk_metadata,
            }
        )
        if '"status": "indexed"' in res or '"status": "enqueued"' in res:
            success_count += 1
        else:
            fail_count += 1

    logger.info(
        "ingest_docs_completed",
        extra={
            "collection": collection,
            "success": success_count,
            "fail": fail_count,
            "total_chunks": len(chunks),
        },
    )

    return json.dumps(
        {
            "status": "completed",
            "total_files": len(docs),
            "total_chunks": len(chunks),
            "indexed": success_count,
            "failed": fail_count,
            "collection": collection,
        }
    )

    if not config.enable_file_operations:
        return "File operations are disabled. Enable ENABLE_FILE_OPERATIONS=true to use this tool."

    if not is_safe_file_path(file_path, allowed_dirs=["."]):
        logger.warning("file_read blocked by safety check", extra={"path": file_path})
        return f"Error: File path '{file_path}' is not allowed"

    try:
        content = Path(file_path).read_text(encoding="utf-8")
        logger.info(
            "file_read completed",
            extra={"path": file_path, "size": len(content)},
        )
        return content
    except FileNotFoundError:
        return f"Error: File '{file_path}' not found"
    except Exception:
        logger.exception("file_read failed", extra={"path": file_path})
        return "Error reading file. Check logs."


@tool
def file_edit(file_path: str, old_text: str, new_text: str) -> str:
    """Edita arquivo substituindo texto.

    Args:
        file_path: Caminho do arquivo
        old_text: Texto a encontrar
        new_text: Texto de substituição

    Returns:
        Confirmação da edição
    """
    from pathlib import Path

    from tool_safety import is_safe_file_path

    config = get_tool_config()

    if not config.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_file_path(file_path, allowed_dirs=["."]):
        logger.warning("file_edit blocked by safety check", extra={"path": file_path})
        return f"Error: File path '{file_path}' is not allowed"

    try:
        path = Path(file_path)
        content = path.read_text(encoding="utf-8")

        if old_text not in content:
            return "Error: Text not found in file"

        new_content = content.replace(old_text, new_text, 1)
        path.write_text(new_content, encoding="utf-8")

        logger.info(
            "file_edit completed",
            extra={
                "path": file_path,
                "old_len": len(old_text),
                "new_len": len(new_text),
            },
        )
        return "✓ File edited successfully"
    except Exception:
        logger.exception("file_edit failed", extra={"path": file_path})
        return "Error editing file. Check logs."


@tool
def grep(pattern: str, path: str = ".") -> str:
    """Busca padrão em arquivos usando regex.

    Args:
        pattern: Padrão regex para buscar
        path: Caminho da pasta ou arquivo

    Returns:
        Linhas que correspondem ao padrão
    """
    import re
    from pathlib import Path

    from tool_safety import is_safe_regex_pattern

    config = get_tool_config()

    if not config.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_regex_pattern(pattern):
        return "Error: Invalid or unsafe regex pattern"

    try:
        results = []
        search_path = Path(path)

        files = [search_path] if search_path.is_file() else list(search_path.rglob("*"))

        for file_path in files:
            if not file_path.is_file() or file_path.suffix in {".pyc", ".o", ".exe"}:
                continue

            try:
                content = file_path.read_text(encoding="utf-8", errors="ignore")
                for line_num, line in enumerate(content.split("\n"), 1):
                    if re.search(pattern, line):
                        results.append(f"{file_path}:{line_num}: {line}")
            except Exception:
                pass

        logger.info(
            "grep completed",
            extra={"pattern": pattern, "path": path, "matches": len(results)},
        )
        return "\n".join(results[:100]) if results else "No matches found"
    except Exception:
        logger.exception("grep failed", extra={"pattern": pattern, "path": path})
        return "Error during grep. Check logs."


@tool
def list_dir(path: str = ".", *, recursive: bool = False) -> str:
    """Lista arquivos em um diretório.

    Args:
        path: Caminho do diretório
        recursive: Se True, lista recursivamente

    Returns:
        Lista de arquivos e pastas
    """
    from pathlib import Path

    config = get_tool_config()

    if not config.enable_file_operations:
        return "File operations are disabled."

    try:
        dir_path = Path(path)

        if not dir_path.exists():
            return f"Error: Directory '{path}' not found"

        if not dir_path.is_dir():
            return f"Error: '{path}' is not a directory"

        items = []
        if recursive:
            for item in sorted(dir_path.rglob("*")):
                rel_path = item.relative_to(dir_path)
                prefix = "[DIR]" if item.is_dir() else "[FILE]"
                items.append(f"{prefix} {rel_path}")
        else:
            for item in sorted(dir_path.iterdir()):
                prefix = "[DIR]" if item.is_dir() else "[FILE]"
                items.append(f"{prefix} {item.name}")

        logger.info(
            "list_dir completed",
            extra={"path": path, "recursive": recursive, "count": len(items)},
        )
        return "\n".join(items[:500]) if items else "(empty directory)"
    except Exception:
        logger.exception("list_dir failed", extra={"path": path})
        return "Error listing directory. Check logs."


@tool
def terminal(command: str) -> str:
    """Executa um comando shell com whitelist de segurança.

    Permite execução de comandos seguros como python, git, npm, node, etc.
    Bloqueia comandos perigosos como rm, dd, mkfs, format, etc.

    Args:
        command: Comando shell para executar (ex: "python --version")

    Returns:
        Saída do comando (stdout + stderr) ou mensagem de erro se bloqueado
    """
    import subprocess as sp

    from tool_safety import is_safe_shell_command

    config = get_tool_config()

    if not config.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_shell_command(command):
        logger.warning(
            "terminal command blocked by safety check",
            extra={"command": command[:50]},
        )
        return f"Error: Command '{command}' is not allowed. Only safe commands like python, git, npm, node, ls, pwd, cat, grep, find, tail, head, wc, sort, uniq, cut are permitted."

    try:
        result = sp.run(
            command,
            shell=True,
            capture_output=True,
            text=True,
            timeout=30,
            check=False,
        )

        output = result.stdout
        if result.stderr:
            output += result.stderr

        logger.info(
            "terminal_command_executed",
            extra={
                "command": command[:50],
                "exit_code": result.returncode,
                "output_length": len(output),
            },
        )

        return output or f"Command executed with exit code {result.returncode}"

    except sp.TimeoutExpired:
        logger.warning(
            "terminal_command_timeout",
            extra={"command": command[:50]},
        )
        return "Error: Command timed out after 30 seconds"
    except Exception:
        logger.exception("terminal_command_failed", extra={"command": command[:50]})
        return "Error executing command. Check logs."


def _build_tools_list() -> list[BaseTool]:
    """Constrói lista de ferramentas disponíveis baseado na configuração.

    Returns:
        Lista de instâncias BaseTool
    """
    config = get_tool_config()
    tools: list[BaseTool] = []

    # Core tools (sempre disponíveis)
    tools.append(web_search)
    tools.append(fetch_url)
    tools.append(vector_search)

    # RAG tools
    if config.enable_rag:
        tools.append(embedding)

    # File operations
    if config.enable_file_operations:
        tools.append(file_read)
        tools.append(file_edit)
        tools.append(grep)
        tools.append(list_dir)
        tools.append(terminal)

    # MCP tool
    if config.enable_mcp:
        tools.append(call_mcp_tool)

    logger.info("Tools initialized", extra={"count": len(tools)})
    return tools


def get_tools() -> list[BaseTool]:
    """Obtém lista de ferramentas disponíveis.

    Returns:
        Lista de instâncias BaseTool
    """
    return _build_tools_list()


# Export tools list for graph construction
TOOLS = _build_tools_list()
