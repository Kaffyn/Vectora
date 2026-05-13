import json
from typing import Any

from langchain.tools import BaseTool, tool
from langchain_community.document_loaders import WebBaseLoader
from langgraph.prebuilt.tool_node import ToolRuntime

import logging

try:
    from langchain_community.tools.duckduckgo_search import DuckDuckGoSearchResults
except ImportError:
    DuckDuckGoSearchResults = None

try:
    from langchain_mcp_adapters import MultiServerMCPClient
except ImportError:
    MultiServerMCPClient = None

from context import Context
from state import State
from tool_config import get_tool_config

logger = logging.getLogger(__name__)

# Global MCP client cache (reuse connection across calls)
_mcp_client: Any | None = None
_mcp_tools_cache: dict[str, Any] | None = None


@tool
def web_search(query: str, runtime: ToolRuntime[Context, State]) -> str:  # noqa: ARG001
    """Search the web for current information using DuckDuckGo.

    Args:
        query: Search query string

    Returns:
        Search results as formatted string with URLs and snippets
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
async def fetch_url(url: str, runtime: ToolRuntime[Context, State]) -> str:  # noqa: ARG001
    """Fetch and extract text content from a specific URL.

    Args:
        url: URL to fetch (must start with http:// or https://)

    Returns:
        Extracted text content from the page
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

        if not docs:
            logger.warning("fetch_url returned no content", extra={"url": url})
            return "No content found at URL"

        content = "\n".join(doc.page_content for doc in docs)
        truncated_content = content[: config.max_fetch_size]

        if len(content) > config.max_fetch_size:
            truncated_content += f"\n... (truncated, max {config.max_fetch_size} chars)"

        logger.info(
            "fetch_url completed",
            extra={"url": url, "content_length": len(content)},
        )

        return truncated_content
    except Exception:
        logger.exception("fetch_url failed", extra={"url": url})
        return "Error occurred. Please check logs."


async def _get_mcp_client() -> Any | None:
    """Get or create MCP client connection (cached)."""
    global _mcp_client
    config = get_tool_config()

    if not config.enable_mcp:
        return None

    # Validate transport-specific config
    if config.mcp_transport_type == "stdio":
        if not config.mcp_command:
            return None
    elif not config.mcp_server_url:
        return None

    if MultiServerMCPClient is None:
        logger.error("langchain_mcp_adapters not installed")
        return None

    if _mcp_client is not None:
        return _mcp_client

    try:
        server_config: dict[str, Any] = {}

        if config.mcp_transport_type == "stdio":
            server_config["transport"] = "stdio"
            server_config["command"] = config.mcp_command
            if config.mcp_command_args:
                server_config["args"] = config.mcp_command_args
        else:
            # HTTP/WebSocket transport (default)
            server_config["transport"] = "http"
            server_config["url"] = config.mcp_server_url

        _mcp_client = MultiServerMCPClient({"mcp_server": server_config})

        logger.info(
            "mcp_client_connected",
            extra={
                "transport": config.mcp_transport_type,
                "server": config.mcp_server_url or config.mcp_command,
            },
        )

        return _mcp_client

    except Exception:
        logger.exception("mcp_client_connection_failed", extra={"error": str(e)})
        return None


async def _get_mcp_tools() -> dict[str, Any] | None:
    """Retrieve available tools from MCP server (cached)."""
    global _mcp_tools_cache

    if _mcp_tools_cache is not None:
        return _mcp_tools_cache

    client = await _get_mcp_client()
    if not client:
        return None

    try:
        tools = await client.get_tools()
        _mcp_tools_cache = {tool.name: tool for tool in tools}

        logger.info(
            "mcp_tools_loaded",
            extra={
                "count": len(_mcp_tools_cache),
                "names": list(_mcp_tools_cache.keys()),
            },
        )

        return _mcp_tools_cache

    except Exception:
        logger.exception("mcp_tools_retrieval_failed", extra={"error": str(e)})
        return None


@tool
async def call_mcp_tool(
    tool_name: str, arguments: str, runtime: ToolRuntime[Context, State]
) -> str:
    """Call a tool registered in a connected MCP (Model Context Protocol) server.

    Args:
        tool_name: Name of the tool in the MCP server
        arguments: JSON string of tool arguments

    Returns:
        Tool execution result as JSON
    """
    config = get_tool_config()

    if not config.enable_mcp:
        logger.warning("call_mcp_tool called but disabled")
        return (
            "MCP server integration is not enabled. "
            "Set ENABLE_MCP=true in environment to use this tool."
        )

    if not config.mcp_server_url and not config.mcp_command:
        logger.warning(
            "call_mcp_tool called but neither MCP_SERVER_URL nor MCP_COMMAND configured"
        )
        return (
            "Error: MCP server not configured. "
            "Set either MCP_SERVER_URL (for HTTP) or MCP_COMMAND (for stdio) environment variables."
        )

    try:
        args_dict: dict[str, Any] = json.loads(arguments)
    except json.JSONDecodeError:
        logger.error(
            "call_mcp_tool failed to parse arguments", extra={"arguments": arguments}
        )
        return f"Error: Invalid JSON arguments: {arguments}"

    logger.info(
        "call_mcp_tool invoked",
        extra={
            "tool_name": tool_name,
            "mcp_server": config.mcp_server_url or config.mcp_command,
        },
    )

    try:
        # Get available tools from MCP server
        tools = await _get_mcp_tools()
        if not tools:
            return (
                "Error: Failed to retrieve tools from MCP server. "
                "Ensure MCP_SERVER_URL or MCP_COMMAND is properly configured and the server is running."
            )

        # Find the requested tool
        if tool_name not in tools:
            available = list(tools.keys())
            logger.warning(
                "mcp_tool_not_found",
                extra={"requested": tool_name, "available": available},
            )
            return json.dumps(
                {
                    "status": "error",
                    "message": f"Tool '{tool_name}' not found in MCP server",
                    "available_tools": available,
                }
            )

        mcp_tool = tools[tool_name]

        # Execute the tool with provided arguments
        client = await _get_mcp_client()
        if not client:
            return "Error: Failed to connect to MCP server"

        result = await client.call_tool(tool_name, args_dict)

        logger.info(
            "mcp_tool_executed_success",
            extra={"tool_name": tool_name, "result_type": type(result).__name__},
        )

        # Return result as JSON
        if isinstance(result, str):
            return json.dumps({"status": "success", "result": result})
        else:
            return json.dumps({"status": "success", "result": result})

    except Exception:
        logger.exception(
            "call_mcp_tool execution failed",
            extra={"tool_name": tool_name},
        )
        return json.dumps(
            {
                "status": "error",
                "message": f"Failed to execute MCP tool '{tool_name}': {e!s}",
            }
        )


# ==============================================================================
# RAG TOOLS: Embedding, Vector Search, and Internal Reranking
# ==============================================================================


@tool
async def embedding(
    text: str,
    collection: str = "articles",
    metadata: dict[str, Any] | None = None,
    runtime: ToolRuntime[Context, State] | None = None,
) -> str:
    """Index document with embedding in Qdrant vector store.

    Generates embeddings using Voyage AI and indexes in the specified collection.
    Falls back to embedding queue if API fails.

    Args:
        text: Document text to embed
        collection: Qdrant collection (articles, wiki, api_docs, knowledge_base)
        metadata: Optional metadata dict (source, author, timestamp, etc)
        runtime: Tool runtime (unused, for LangGraph compatibility)

    Returns:
        JSON status: indexed/queued_for_indexing/failed with point_id or queue_id
    """
    from uuid import uuid4

    from embedding_queue import get_embedding_queue
    from langchain_voyageai import VoyageAIEmbeddings
    from qdrant_client import QdrantClient
    from qdrant_client.models import Distance, PointStruct, VectorParams

    config = get_tool_config()

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
        embeddings_model = VoyageAIEmbeddings(
            api_key=config.voyage_api_key,
            model=config.embedding_model,
        )

        vector = embeddings_model.embed_query(text)

        qdrant_client = QdrantClient(
            url=config.qdrant_url,
            api_key=config.qdrant_api_key,
        )

        try:
            qdrant_client.get_collection(collection)
        except Exception:
            qdrant_client.create_collection(
                collection_name=collection,
                vectors_config=VectorParams(
                    size=len(vector),
                    distance=Distance.COSINE,
                ),
            )
            logger.info("qdrant_collection_created", extra={"collection": collection})

        point_id = str(uuid4())
        qdrant_client.upsert(
            collection_name=collection,
            points=[
                PointStruct(
                    id=point_id,
                    vector=vector,
                    payload={
                        "page_content": text,
                        "metadata": metadata or {},
                    },
                )
            ],
        )

        logger.info(
            "embedding_indexed_success",
            extra={
                "collection": collection,
                "point_id": point_id,
                "text_length": len(text),
                "vector_dims": len(vector),
            },
        )

        return json.dumps(
            {
                "status": "indexed",
                "collection": collection,
                "point_id": point_id,
                "text_length": len(text),
            }
        )

    except Exception:
        logger.exception(
            "embedding_failed",
            extra={"error": str(e), "collection": collection},
        )

        if config.embedding_queue_enabled:
            try:
                queue = await get_embedding_queue(config.embedding_queue_db)
                queue_id = await queue.enqueue(text, collection, metadata)

                logger.warning(
                    "embedding_queued_for_retry",
                    extra={"queue_id": queue_id, "collection": collection},
                )

                return json.dumps(
                    {
                        "status": "queued_for_indexing",
                        "queue_id": queue_id,
                        "message": f"Documento enfileirado para indexação. ID: {queue_id}. Será processado em background.",
                        "collection": collection,
                    }
                )

            except Exception:
                logger.error(
                    "embedding_queue_failed",
                    extra={"error": str(queue_err)},
                )

                return json.dumps(
                    {
                        "status": "failed",
                        "error": f"Both embedding and queue failed: {e!s}",
                        "collection": collection,
                    }
                )
        else:
            return json.dumps(
                {
                    "status": "failed",
                    "collection": collection,
                }
            )


async def _internal_reranker(
    query: str,
    raw_results: list[dict[str, Any]],
    top_k: int = 5,
) -> list[dict[str, Any]]:
    """Internal reranker (not exposed as tool to LLM).

    Reranks raw vector search results using BM25 (local) or Voyage Reranker API.

    Args:
        query: Original search query
        raw_results: Raw results from Qdrant
        top_k: Number of top results to return

    Returns:
        Reranked results with updated relevance_score
    """
    if not raw_results:
        return []

    config = get_tool_config()
    documents = [r.get("page_content", "") for r in raw_results]

    try:
        if config.reranker_type == "bm25":
            return _rerank_bm25(query, raw_results, documents, top_k)
        else:
            return await _rerank_voyage(query, raw_results, documents, top_k, config)

    except Exception:
        logger.exception(
            "reranker_failed",
            extra={"error": str(e), "reranker_type": config.reranker_type},
        )
        return _rerank_bm25(query, raw_results, documents, top_k)


async def _rerank_voyage(
    query: str,
    raw_results: list[dict[str, Any]],
    documents: list[str],
    top_k: int,
    config: Any,
) -> list[dict[str, Any]]:
    """Rerank using Voyage AI Reranker API."""
    from langchain_voyageai import VoyageAIRerank

    try:
        reranker = VoyageAIRerank(
            api_key=config.voyage_api_key,
            model=config.reranker_model,
            top_k=top_k,
        )

        reranked_docs = reranker.compress_documents(
            documents=[{"page_content": doc, "metadata": {}} for doc in documents],
            query=query,
        )

        reranked = []
        for idx, doc in enumerate(reranked_docs):
            original = raw_results[documents.index(doc.page_content)].copy()
            original["relevance_score"] = getattr(
                doc, "relevance_score", 1.0 - (idx / len(reranked_docs))
            )
            reranked.append(original)

        logger.debug(
            "reranker_voyage_completed",
            extra={
                "raw_count": len(raw_results),
                "reranked_count": len(reranked),
            },
        )

        return reranked

    except Exception:
        logger.exception(
            "reranker_voyage_failed",
            extra={"error": str(e)},
        )
        return _rerank_bm25(query, raw_results, documents, top_k)


def _rerank_bm25(
    query: str,
    raw_results: list[dict[str, Any]],
    documents: list[str],
    top_k: int,
) -> list[dict[str, Any]]:
    """Rerank using local BM25 algorithm (fallback or MVP)."""
    from rank_bm25 import BM25Okapi

    tokenized_docs = [doc.lower().split() for doc in documents]
    bm25 = BM25Okapi(tokenized_docs)

    query_tokens = query.lower().split()
    scores = bm25.get_scores(query_tokens)

    scored_results = []
    for idx, score in enumerate(scores):
        result = raw_results[idx].copy()
        result["relevance_score"] = float(score)
        scored_results.append(result)

    scored_results.sort(key=lambda x: x.get("relevance_score", 0), reverse=True)

    logger.debug(
        "reranker_bm25_completed",
        extra={
            "raw_count": len(raw_results),
            "reranked_count": min(top_k, len(scored_results)),
        },
    )

    return scored_results[:top_k]


@tool
async def vector_search(
    query: str,
    collections: list[str] | None = None,
    top_k: int = 10,
    min_score: float = 0.5,
    runtime: ToolRuntime[Context, State] | None = None,
) -> str:
    """Search Qdrant vector store and return reranked results.

    Automatically applies reranker to results (internal, not exposed to LLM).

    Args:
        query: Search query string
        collections: Which collections to search (default: all)
        top_k: Number of raw results before reranking
        min_score: Minimum similarity threshold (0-1)
        runtime: Tool runtime (unused, for LangGraph compatibility)

    Returns:
        JSON: {status: found/no_results, results: [...], total_count: N}
    """
    from langchain_voyageai import VoyageAIEmbeddings
    from qdrant_client import QdrantClient

    config = get_tool_config()

    if not config.voyage_api_key:
        logger.error("vector_search called but VOYAGE_API_KEY not configured")
        return json.dumps(
            {
                "status": "failed",
                "error": "VOYAGE_API_KEY not configured",
            }
        )

    try:
        embeddings_model = VoyageAIEmbeddings(
            api_key=config.voyage_api_key,
            model=config.embedding_model,
        )

        query_vector = embeddings_model.embed_query(query)

        search_collections = collections or config.qdrant_collections
        if not search_collections:
            search_collections = ["articles", "wiki", "api_docs", "knowledge_base"]

        qdrant_client = QdrantClient(
            url=config.qdrant_url,
            api_key=config.qdrant_api_key,
        )

        all_results: list[dict[str, Any]] = []

        for collection in search_collections:
            try:
                hits = qdrant_client.search(
                    collection_name=collection,
                    query_vector=query_vector,
                    limit=top_k,
                    score_threshold=min_score,
                )

                for hit in hits:
                    result = {
                        "page_content": hit.payload.get("page_content", ""),
                        "metadata": hit.payload.get("metadata", {}),
                        "relevance_score": hit.score,
                        "collection": collection,
                        "point_id": hit.id,
                    }
                    all_results.append(result)

            except Exception:
                logger.warning(
                    "vector_search_collection_failed",
                    extra={
                        "collection": collection,
                        "error": str(collection_err),
                    },
                )
                continue

        if not all_results:
            logger.info(
                "vector_search_no_results",
                extra={"query": query[:100], "collections": search_collections},
            )
            return json.dumps(
                {
                    "status": "no_results",
                    "count": 0,
                    "message": "No matching documents found. Try using embedding() to index new documents.",
                }
            )

        reranked_results = await _internal_reranker(
            query, all_results, config.reranker_top_k
        )

        logger.info(
            "vector_search_completed",
            extra={
                "query": query[:100],
                "collections": search_collections,
                "raw_results": len(all_results),
                "reranked_results": len(reranked_results),
                "reranker_type": config.reranker_type,
            },
        )

        return json.dumps(
            {
                "status": "found",
                "count": len(reranked_results),
                "results": reranked_results,
                "collections_searched": search_collections,
            }
        )

    except Exception:
        logger.exception(
            "vector_search_failed",
            extra={"query": query[:100]},
        )
        return json.dumps(
            {
                "status": "failed",
            }
        )


def _build_tools_list() -> list[BaseTool]:
    """Build the list of available tools based on configuration."""
    return [
        web_search,
        fetch_url,
        embedding,
        vector_search,
        call_mcp_tool,
    ]


TOOLS: list[BaseTool] = _build_tools_list()
TOOLS_BY_NAME: dict[str, BaseTool] = {tool.name: tool for tool in TOOLS}
