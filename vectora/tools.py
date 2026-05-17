"""Core Vectora Tools Implementation (10-tool Suite).

Implements: web_search, fetch_url, embedding, vector_search, file_read, file_edit,
grep, terminal, list_dir, call_mcp_tool. All tools include error handling and logging.
"""

import asyncio
import json
import logging
from datetime import UTC, datetime
from typing import Any

from langchain.tools import BaseTool, tool
from langchain_community.document_loaders import DirectoryLoader, TextLoader
from langchain_core.documents import Document as LCDoc
from tavily import TavilyClient

try:
    from langchain_mcp_adapters import MultiServerMCPClient  # type: ignore
except ImportError:
    MultiServerMCPClient = None

from embedding_queue import get_embedding_queue
from settings import settings

# Heavy RAG imports moved to top for better initialization
try:
    import lancedb
    import pyarrow as pa
    from langchain_voyageai import VoyageAIEmbeddings, VoyageAIRerank
    from pydantic import SecretStr
except ImportError:
    lancedb = None
    pa = None
    VoyageAIEmbeddings = None
    VoyageAIRerank = None
    SecretStr = None

logger = logging.getLogger(__name__)

# Global MCP client cache (reuse connection across calls)
_mcp_client: Any | None = None
_mcp_tools_cache: dict[str, Any] | None = None


@tool
def web_search(query: str) -> str:
    """Busca a web por informações atuais usando Tavily (otimizado para agentes).

    Tavily retorna resultados estruturados, prontos para RAG, com conteúdo já extraído.

    Args:
        query: String da consulta de busca

    Returns:
        JSON com resultados estruturados (url, title, content) prontos para embedding
    """
    if not settings.enable_web_search:
        logger.warning("web_search tool called but disabled")
        return "Web search is disabled. Enable ENABLE_WEB_SEARCH=true to use this tool."

    if not settings.tavily_api_key:
        logger.error("TAVILY_API_KEY not configured")
        return json.dumps(
            {
                "status": "error",
                "error": "TAVILY_API_KEY not configured. Set TAVILY_API_KEY environment variable.",
            }
        )

    logger.info("web_search tool called", extra={"query": query})

    try:
        client = TavilyClient(api_key=settings.tavily_api_key)
        # search_depth="advanced" garante que o Tavily extraia e limpe o conteúdo
        response = client.search(query=query, search_depth="advanced", max_results=5)

        logger.info(
            "web_search completed",
            extra={"query": query, "num_results": len(response.get("results", []))},
        )

        # Retorna JSON estruturado pronto para RAG
        return json.dumps(response["results"])
    except Exception:
        logger.exception(
            "web_search failed",
            extra={"query": query},
        )
        return json.dumps(
            {
                "status": "error",
                "error": "Web search failed. Please try again.",
            }
        )


@tool
def fetch_url(url: str) -> str:
    """Busca e extrai conteúdo de texto de uma URL específica usando Tavily.

    Args:
        url: URL para buscar (deve começar com http:// ou https://)

    Returns:
        Conteúdo de texto extraído da página
    """
    if not url.startswith(("http://", "https://")):
        logger.warning("fetch_url called with invalid URL", extra={"url": url})
        return "Error: URL must start with http:// or https://"

    if not settings.tavily_api_key:
        logger.error("TAVILY_API_KEY not configured")
        return "Error: TAVILY_API_KEY not configured. Cannot fetch URL."

    logger.info("fetch_url tool called", extra={"url": url})

    try:
        client = TavilyClient(api_key=settings.tavily_api_key)

        # Usa extract() dedicado para extração de conteúdo de URL
        # (search() usa a URL como query de busca — comportamento incorreto)
        response = client.extract(urls=[url])

        results = response.get("results", [])
        if not results:
            logger.warning("fetch_url returned no content", extra={"url": url})
            return f"No content found at {url}"

        content = results[0].get("raw_content", "") or results[0].get("content", "")

        logger.info(
            "fetch_url completed",
            extra={"url": url, "content_length": len(content)},
        )

        return content

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

    if not settings.mcp_server_url and not settings.mcp_command:
        logger.debug("No MCP servers configured")
        return None

    try:
        # Constrói dicionário de servidores com base nos campos reais do Settings
        servers: dict[str, Any] = {}
        if settings.mcp_server_url:
            servers["default"] = {
                "url": settings.mcp_server_url,
                "transport": settings.mcp_transport_type,
            }
        if settings.mcp_command:
            servers["local"] = {
                "command": settings.mcp_command,
                "args": settings.mcp_command_args or [],
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
    """Invoca ferramentas via MCP Protocol (Vector Context Protocol).

    Args:
        tool_name: Nome da ferramenta no servidor MCP
        arguments: Argumentos em formato JSON string

    Returns:
        Resposta da ferramenta MCP
    """
    global _mcp_client

    if MultiServerMCPClient is None:
        return "MCP client not available. Install: pip install langchain-mcp-adapters"

    # Using global settings singleton instead of get_tool_config()
    if not settings.enable_mcp or not settings.mcp_server_url:
        return "MCP is disabled or server URL not configured."

    try:
        if _mcp_client is None:
            _mcp_client = MultiServerMCPClient()
            if settings.mcp_transport_type == "http":
                await _mcp_client.connect_sse(settings.mcp_server_url)
            else:
                pass

        logger.info("call_mcp_tool", extra={"tool": tool_name, "arguments": arguments})

        args_dict = json.loads(arguments)

        async with asyncio.timeout(settings.mcp_timeout):
            result_text = ""
            async for event in _mcp_client.astream_events(
                {"tool_name": tool_name, "arguments": args_dict}, version="v1"
            ):
                if event.get("event") == "on_tool_stream":
                    chunk = event.get("data", {}).get("chunk")
                    if chunk:
                        result_text += str(chunk)

        return result_text or str(args_dict)

    except TimeoutError:
        logger.exception(
            f"MCP tool {tool_name} timed out after {settings.mcp_timeout}s"
        )
        return f"Error: MCP tool {tool_name} timed out."
    except Exception as e:
        logger.exception("call_mcp_tool_failed", extra={"tool": tool_name})
        return f"Error invoking MCP tool: {e!s}"


@tool
async def embedding(
    text: str, collection: str = "articles", metadata: dict[str, Any] | None = None
) -> str:
    """Enfileira documento para embedding assíncrono (fire-and-forget).

    Em vez de bloquear esperando Voyage AI (5+ segundos), este método
    enfileira o documento imediatamente. Um background worker processa
    a fila e indexa em LanceDB de forma não-bloqueante.

    Args:
        text: Texto do documento a indexar
        collection: Nome da coleção/tabela LanceDB (articles, wiki, api_docs, knowledge_base)
        metadata: Dicionário opcional de metadados (source, author, timestamp, etc)

    Returns:
        String JSON com status: fire_and_forget + queue_id ou error
    """

    # Using global settings singleton instead of get_tool_config()

    if not settings.enable_rag:
        logger.warning("embedding tool called but RAG disabled")
        return json.dumps(
            {
                "status": "error",
                "error": "RAG is disabled. Enable ENABLE_RAG=true to use this tool.",
            }
        )

    if not settings.embedding_queue_enabled:
        logger.error("embedding called but queue not enabled")
        return json.dumps(
            {
                "status": "error",
                "error": "Embedding queue not configured.",
            }
        )

    try:
        queue = await get_embedding_queue(settings.embedding_queue_url)
        queue_id = await queue.enqueue(text, collection, metadata)

        logger.info(
            "embedding_enqueued",
            extra={
                "queue_id": queue_id,
                "collection": collection,
                "text_length": len(text),
            },
        )

        return json.dumps(
            {
                "status": "fire_and_forget",
                "queue_id": queue_id,
                "collection": collection,
                "message": "Document enqueued for async embedding and indexing.",
            }
        )

    except Exception as e:
        logger.exception(
            "embedding_enqueue_failed",
            extra={"collection": collection, "text_length": len(text)},
        )
        return json.dumps(
            {
                "status": "error",
                "error": str(e) or "Failed to enqueue embedding",
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
    # Using global settings singleton instead of get_tool_config()

    if not settings.enable_rag:
        logger.warning("vector_search tool called but RAG disabled")
        return "RAG is disabled. Enable ENABLE_RAG=true to use this tool."

    try:
        if lancedb is None or VoyageAIEmbeddings is None:
            return "LanceDB or Voyage AI dependencies missing."

        if not settings.voyage_api_key:
            logger.error("vector_search called but VOYAGE_API_KEY not configured")
            return json.dumps(
                {
                    "status": "failed",
                    "error": "VOYAGE_API_KEY not configured",
                }
            )

        embeddings_model = VoyageAIEmbeddings(
            api_key=SecretStr(settings.voyage_api_key),
            model=settings.embedding_model,
        )

        # embed_query is synchronous (no timeout needed)
        query_vector = embeddings_model.embed_query(query)

        db = await lancedb.connect_async(str(settings.lancedb_path))

        try:
            async with asyncio.timeout(10):  # 10s timeout para LanceDB
                table = await db.open_table(collection)
        except TimeoutError:
            logger.exception(f"LanceDB open_table timed out for {collection}")
            return json.dumps(
                {
                    "status": "error",
                    "error": "Vector store access timed out",
                }
            )
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

        try:
            async with asyncio.timeout(10):  # 10s timeout para search
                search_results = await (
                    table.vector_search(query_vector).limit(limit).to_pandas()
                )
        except TimeoutError:
            logger.exception(f"vector_search timed out for collection {collection}")
            return json.dumps(
                {
                    "status": "error",
                    "error": "Search timed out after 10s",
                }
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
        if results and settings.reranker_type == "voyage" and VoyageAIRerank:
            try:
                reranker = VoyageAIRerank(
                    api_key=SecretStr(settings.voyage_api_key),
                    model=settings.reranker_model,
                    top_k=settings.reranker_top_k,
                )

                docs_to_rerank = [
                    LCDoc(page_content=str(r["content"]), metadata=r["metadata"])
                    for r in results
                ]

                reranked_docs = reranker.compress_documents(docs_to_rerank, query)

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
                # Sinaliza ao LLM que os resultados não foram reordenados por relevância
                for r in results:
                    r["reranking_status"] = "unavailable"

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
    # Using global settings singleton instead of get_tool_config()
    if not settings.enable_file_operations:
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

    # Using global settings singleton instead of get_tool_config()
    if not settings.enable_file_operations:
        return "File operations are disabled. Enable ENABLE_FILE_OPERATIONS=true to use this tool."

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

    try:
        docs = loader.load()
    except Exception as e:
        logger.exception(f"Error loading documents from {directory_path}: {e}")
        return f"Error loading documents: {e!s}"

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

        res = ""
        async for stream_chunk in embedding.astream(
            {
                "text": chunk.page_content,
                "collection": collection,
                "metadata": chunk_metadata,
            }
        ):
            if isinstance(stream_chunk, str):
                res += stream_chunk

        # Sucesso: documento enfileirado para processamento assíncrono
        if '"status": "fire_and_forget"' in res:
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


@tool
def file_edit(
    file_path: str, old_text: str, new_text: str, replace_all: bool = False
) -> str:
    """Edita arquivo substituindo texto.

    Args:
        file_path: Caminho do arquivo
        old_text: Texto a encontrar (use "" para criar arquivo se não existir)
        new_text: Texto de substituição
        replace_all: Se True, substitui todas as ocorrências; padrão substitui apenas a 1ª

    Returns:
        Confirmação da edição
    """
    from pathlib import Path

    from tool_safety import is_safe_file_path

    if not settings.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_file_path(file_path, allowed_dirs=["."]):
        logger.warning("file_edit blocked by safety check", extra={"path": file_path})
        return f"Error: File path '{file_path}' is not allowed"

    try:
        path = Path(file_path)

        # Cria arquivo novo quando old_text="" e arquivo não existe
        if old_text == "" and not path.exists():
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(new_text, encoding="utf-8")
            logger.info("file_edit created new file", extra={"path": file_path})
            return f"[OK] File created: {file_path}"

        content = path.read_text(encoding="utf-8")

        if old_text and old_text not in content:
            return "Error: Text not found in file"

        new_content = (
            content.replace(old_text, new_text)
            if replace_all
            else content.replace(old_text, new_text, 1)
        )
        path.write_text(new_content, encoding="utf-8")

        count = content.count(old_text) if replace_all else 1
        logger.info(
            "file_edit completed",
            extra={"path": file_path, "occurrences": count, "replace_all": replace_all},
        )
        return f"[OK] File edited successfully ({count} occurrence{'s' if count != 1 else ''} replaced)"
    except Exception:
        logger.exception("file_edit failed", extra={"path": file_path})
        return "Error editing file. Check logs."


@tool
def file_write(file_path: str, content: str) -> str:
    """Cria ou sobrescreve completamente um arquivo com o conteúdo fornecido.

    Use para criar novos arquivos ou substituir o conteúdo completo de um existente.
    Para edições cirúrgicas (substituir trechos), prefira file_edit.

    Args:
        file_path: Caminho do arquivo (absoluto ou relativo)
        content: Conteúdo completo a escrever no arquivo

    Returns:
        Confirmação com caminho e tamanho em bytes
    """
    from pathlib import Path

    from tool_safety import is_safe_file_path

    if not settings.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_file_path(file_path, allowed_dirs=["."]):
        logger.warning("file_write blocked by safety check", extra={"path": file_path})
        return f"Error: File path '{file_path}' is not allowed"

    try:
        path = Path(file_path)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")

        size = path.stat().st_size
        logger.info(
            "file_write completed", extra={"path": file_path, "size_bytes": size}
        )
        return f"[OK] File written: {file_path} ({size} bytes)"
    except Exception:
        logger.exception("file_write failed", extra={"path": file_path})
        return "Error writing file. Check logs."


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

    # Using global settings singleton instead of get_tool_config()

    if not settings.enable_file_operations:
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

    # Using global settings singleton instead of get_tool_config()

    if not settings.enable_file_operations:
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

    # Using global settings singleton instead of get_tool_config()

    if not settings.enable_file_operations:
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


@tool
async def save_memory(
    key: str,
    content: str,
    metadata: dict[str, Any] | None = None,
    ttl_days: int | None = None,
) -> str:
    """Salva uma memória persistente global para uso em futuras conversas.

    As memórias são armazenadas no SQLite e recuperadas automaticamente
    ao iniciar novas sessões com o mesmo usuário.

    Args:
        key: Chave única da memória (ex: 'user_preferences', 'project_context', 'api_keys')
        content: Conteúdo da memória (string com informações a recordar)
        metadata: Metadados adicionais (dict com informações sobre a memória)
        ttl_days: Dias até expiração automática (None = nunca expira)

    Returns:
        String JSON com status: saved/failed com memory_id ou mensagem de erro
    """
    try:
        from memory_store import get_memory_store

        # Usa user_id padrão para armazenamento de memória
        user_id = "default_user"

        memory_store = await get_memory_store()
        memory_id = await memory_store.save(
            user_id=user_id,
            key=key,
            content=content,
            metadata=metadata,
            ttl_days=ttl_days,
        )

        logger.info(
            "memory_saved",
            extra={"key": key, "memory_id": memory_id, "ttl_days": ttl_days},
        )

        return json.dumps(
            {
                "status": "saved",
                "memory_id": memory_id,
                "key": key,
                "expires_in_days": ttl_days,
            }
        )

    except Exception as e:
        logger.exception("save_memory_failed", extra={"key": key})
        return json.dumps({"status": "failed", "error": str(e), "key": key})


@tool
async def get_memory(key: str | None = None) -> str:
    """Recupera memórias persistentes salvas anteriormente.

    Se key for None, recupera todas as memórias ativas do usuário.
    Memórias expiradas são automaticamente filtradas.

    Args:
        key: Chave da memória específica. Se None, retorna todas as memórias.

    Returns:
        String JSON com conteúdo da memória ou lista de memórias
    """
    try:
        from memory_store import get_memory_store

        user_id = "default_user"

        memory_store = await get_memory_store()

        if key is not None:
            # Recupera memória específica
            memory = await memory_store.get(user_id, key)
            if memory is None:
                logger.warning("memory_not_found", extra={"key": key})
                return json.dumps({"status": "not_found", "key": key})

            logger.debug("memory_retrieved", extra={"key": key})
            return json.dumps(
                {
                    "status": "found",
                    "key": key,
                    "content": memory["content"],
                    "metadata": memory["metadata"],
                    "updated_at": memory["updated_at"],
                }
            )
        # Recupera todas as memórias
        all_memories = await memory_store.get_all(user_id)
        logger.debug("all_memories_retrieved", extra={"count": len(all_memories)})
        return json.dumps(
            {
                "status": "success",
                "count": len(all_memories),
                "memories": [
                    {
                        "key": m["key"],
                        "content": m["content"],
                        "metadata": m["metadata"],
                        "updated_at": m["updated_at"],
                    }
                    for m in all_memories
                ],
            }
        )

    except Exception as e:
        logger.exception("get_memory_failed", extra={"key": key})
        return json.dumps({"status": "failed", "error": str(e)})


@tool
async def delete_memory(key: str) -> str:
    """Deleta uma memória persistente.

    Args:
        key: Chave da memória a deletar

    Returns:
        String JSON com status: deleted/not_found/failed
    """
    try:
        from memory_store import get_memory_store

        user_id = "default_user"

        memory_store = await get_memory_store()
        deleted = await memory_store.delete(user_id, key)

        if deleted:
            logger.info("memory_deleted", extra={"key": key})
            return json.dumps({"status": "deleted", "key": key})
        logger.warning("memory_not_found_for_deletion", extra={"key": key})
        return json.dumps({"status": "not_found", "key": key})

    except Exception as e:
        logger.exception("delete_memory_failed", extra={"key": key})
        return json.dumps({"status": "failed", "error": str(e), "key": key})


def _build_tools_list() -> list[BaseTool]:
    """Constrói lista de ferramentas disponíveis baseado na configuração.

    Returns:
        Lista de instâncias BaseTool
    """
    # Using global settings singleton instead of get_tool_config()
    tools: list[BaseTool] = []

    # Core tools (sempre disponíveis)
    tools.append(web_search)
    tools.append(fetch_url)
    tools.append(vector_search)
    tools.append(save_memory)
    tools.append(get_memory)
    tools.append(delete_memory)

    # RAG tools
    if settings.enable_rag:
        tools.append(embedding)

    # File operations
    if settings.enable_file_operations:
        tools.append(file_read)
        tools.append(file_edit)
        tools.append(file_write)
        tools.append(grep)
        tools.append(list_dir)
        tools.append(terminal)

    # MCP tool
    if settings.enable_mcp:
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

# Export tools by name for test discovery
TOOLS_BY_NAME = {tool.name: tool for tool in TOOLS}
