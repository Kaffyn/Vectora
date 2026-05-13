import json
import logging
from typing import Any

from langchain.tools import BaseTool, tool
from langchain_community.document_loaders import WebBaseLoader
from langchain_community.tools.duckduckgo_search import DuckDuckGoSearchResults
from langgraph.prebuilt.tool_node import ToolRuntime

from context import Context
from state import State
from tool_config import get_tool_config

logger = logging.getLogger(__name__)


@tool
def multiply(a: float, b: float, runtime: ToolRuntime[Context, State]) -> float:  # noqa: ARG001
    """Multiply a * b and returns the result

    Args:
        a: float multiplicand
        b: float multiplier

    Returns:
        the resulting float of the equation a * b
    """
    result = a * b
    logger.info(
        "multiply tool executed",
        extra={"a": a, "b": b, "result": result},
    )
    return result


@tool
def web_search(query: str, runtime: ToolRuntime[Context, State]) -> str:  # noqa: ARG001
    """Search the web for current information using DuckDuckGo.

    Args:
        query: Search query string

    Returns:
        Search results as formatted string with URLs and snippets
    """
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
    except Exception as e:
        logger.error(
            "web_search failed",
            extra={"query": query, "error": str(e)},
        )
        return f"Error searching web: {e}"


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

    # Validate URL
    if not url.startswith(("http://", "https://")):
        logger.warning("fetch_url called with invalid URL", extra={"url": url})
        return "Error: URL must start with http:// or https://"

    # Check domain whitelist if configured
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

        # Combine document content and truncate
        content = "\n".join(doc.page_content for doc in docs)
        truncated_content = content[: config.max_fetch_size]

        if len(content) > config.max_fetch_size:
            truncated_content += f"\n... (truncated, max {config.max_fetch_size} chars)"

        logger.info(
            "fetch_url completed",
            extra={"url": url, "content_length": len(content)},
        )

        return truncated_content
    except Exception as e:
        logger.error("fetch_url failed", extra={"url": url, "error": str(e)})
        return f"Error fetching URL: {e}"


@tool
def query_database(sql: str, runtime: ToolRuntime[Context, State]) -> str:  # noqa: ARG001
    """Execute a SQL SELECT query against the configured database.

    Args:
        sql: SQL SELECT query (no INSERT/UPDATE/DELETE for safety)

    Returns:
        Query results as formatted table string
    """
    config = get_tool_config()

    if not config.enable_database:
        logger.warning("query_database tool called but disabled")
        return (
            "Database tool is disabled. Enable ENABLE_DATABASE=true to use this tool."
        )

    if not config.database_url:
        logger.warning("query_database called but DATABASE_URL not configured")
        return "Error: DATABASE_URL environment variable not configured"

    # Security: Only allow SELECT queries
    sql_upper = sql.strip().upper()
    if not sql_upper.startswith("SELECT"):
        logger.warning("query_database blocked non-SELECT query", extra={"query": sql})
        return "Error: Only SELECT queries are allowed for security reasons"

    # Block dangerous SQL keywords
    dangerous_keywords = ["INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE"]
    for keyword in dangerous_keywords:
        if keyword in sql_upper:
            logger.warning(
                "query_database blocked query with dangerous keyword",
                extra={"query": sql, "keyword": keyword},
            )
            return f"Error: {keyword} operations are not allowed"

    logger.info("query_database tool called", extra={"query": sql})

    try:
        from langchain_community.utilities import SQLDatabase

        db = SQLDatabase.from_uri(config.database_url)

        # Validate table names if whitelist configured
        if config.allowed_tables:
            for table in db.get_usable_table_names():
                if table not in config.allowed_tables:
                    logger.warning(
                        "query_database accessed unauthorized table",
                        extra={"table": table},
                    )
                    return f"Error: Table {table} is not in allowed tables"

        result = db.run(sql, fetch="all")

        logger.info(
            "query_database completed",
            extra={"query": sql, "result_length": len(str(result))},
        )

        return str(result)
    except ImportError:
        logger.error("query_database failed: sqlalchemy not installed")
        return "Error: SQLAlchemy is not installed. Run 'uv sync --group database' to enable this tool."
    except Exception as e:
        logger.error("query_database failed", extra={"query": sql, "error": str(e)})
        return f"Error executing query: {e}"


@tool
async def call_mcp_tool(
    tool_name: str, arguments: str, runtime: ToolRuntime[Context, State]
) -> str:
    """Call a tool registered in a connected MCP (Model Context Protocol) server.

    Args:
        tool_name: Name of the tool in the MCP server
        arguments: JSON string of tool arguments

    Returns:
        Tool execution result
    """
    config = get_tool_config()

    if not config.enable_mcp:
        logger.warning("call_mcp_tool called but disabled")
        return (
            "MCP server integration is not enabled. "
            "Set ENABLE_MCP=true and MCP_SERVER_URL to use this tool."
        )

    if not config.mcp_server_url:
        logger.warning("call_mcp_tool called but MCP_SERVER_URL not configured")
        return (
            "Error: MCP_SERVER_URL environment variable not configured. "
            "Set it to the WebSocket URL of your MCP server (e.g., ws://localhost:5000)"
        )

    logger.info(
        "call_mcp_tool called",
        extra={"tool_name": tool_name, "mcp_server": config.mcp_server_url},
    )

    # Parse arguments
    try:
        args_dict: dict[str, Any] = json.loads(arguments)
    except json.JSONDecodeError:
        logger.error(
            "call_mcp_tool failed to parse arguments", extra={"arguments": arguments}
        )
        return f"Error: Invalid JSON arguments: {arguments}"

    try:
        # Future: Implement actual MCP client connection
        # For MVP: Return informative message
        logger.info(
            "call_mcp_tool - MCP server support coming in future release",
            extra={"tool_name": tool_name},
        )

        return (
            f"MCP server support is coming in a future release. "
            f"Tool '{tool_name}' is registered in the MCP server at {config.mcp_server_url} "
            f"but integration is not yet implemented."
        )
    except Exception as e:
        logger.error(
            "call_mcp_tool failed",
            extra={"tool_name": tool_name, "error": str(e)},
        )
        return f"Error calling MCP tool: {e}"


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

    from langchain_voyageai import VoyageAIEmbeddings
    from qdrant_client import QdrantClient
    from qdrant_client.models import Distance, PointStruct, VectorParams

    from embedding_queue import get_embedding_queue

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
        # 1. Generate embeddings using Voyage AI
        embeddings_model = VoyageAIEmbeddings(
            api_key=config.voyage_api_key,
            model=config.embedding_model,
        )

        vector = embeddings_model.embed_query(text)

        # 2. Connect to Qdrant
        qdrant_client = QdrantClient(
            url=config.qdrant_url,
            api_key=config.qdrant_api_key,
        )

        # Ensure collection exists
        try:
            qdrant_client.get_collection(collection)
        except Exception:
            # Collection doesn't exist, create it
            qdrant_client.create_collection(
                collection_name=collection,
                vectors_config=VectorParams(
                    size=len(vector),
                    distance=Distance.COSINE,
                ),
            )
            logger.info("qdrant_collection_created", extra={"collection": collection})

        # 3. Index the document
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

    except Exception as e:
        logger.error(
            "embedding_failed",
            extra={"error": str(e), "collection": collection},
        )

        # Fallback: Enqueue for later processing
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

            except Exception as queue_error:
                logger.error(
                    "embedding_queue_failed",
                    extra={"error": str(queue_error)},
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
                    "error": str(e),
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

    # Extract documents for reranking
    documents = [r.get("page_content", "") for r in raw_results]

    try:
        if config.reranker_type == "bm25":
            # Local BM25 reranking (MVP)
            return _rerank_bm25(query, raw_results, documents, top_k)
        else:
            # Voyage AI Reranker (production)
            return await _rerank_voyage(query, raw_results, documents, top_k, config)

    except Exception as e:
        logger.error(
            "reranker_failed",
            extra={"error": str(e), "reranker_type": config.reranker_type},
        )
        # Fallback to BM25 if Voyage fails
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

        # VoyageAIRerank returns list of CompressedDocument
        reranked_docs = reranker.compress_documents(
            documents=[{"page_content": doc, "metadata": {}} for doc in documents],
            query=query,
        )

        # Map back to original results with scores
        reranked = []
        for idx, doc in enumerate(reranked_docs):
            original = raw_results[documents.index(doc.page_content)].copy()
            # Voyage reranker provides relevance_score
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

    except Exception as e:
        logger.error(
            "reranker_voyage_failed",
            extra={"error": str(e)},
        )
        # Fallback to BM25
        return _rerank_bm25(query, raw_results, documents, top_k)


def _rerank_bm25(
    query: str,
    raw_results: list[dict[str, Any]],
    documents: list[str],
    top_k: int,
) -> list[dict[str, Any]]:
    """Rerank using local BM25 algorithm (fallback or MVP)."""
    from rank_bm25 import BM25Okapi

    # Tokenize documents
    tokenized_docs = [doc.lower().split() for doc in documents]
    bm25 = BM25Okapi(tokenized_docs)

    # Score query against documents
    query_tokens = query.lower().split()
    scores = bm25.get_scores(query_tokens)

    # Create scored results
    scored_results = []
    for idx, score in enumerate(scores):
        result = raw_results[idx].copy()
        result["relevance_score"] = float(score)
        scored_results.append(result)

    # Sort by relevance score (descending)
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
        # 1. Get embeddings for query
        embeddings_model = VoyageAIEmbeddings(
            api_key=config.voyage_api_key,
            model=config.embedding_model,
        )

        query_vector = embeddings_model.embed_query(query)

        # 2. Determine which collections to search
        search_collections = collections or config.qdrant_collections
        if not search_collections:
            search_collections = ["articles", "wiki", "api_docs", "knowledge_base"]

        # 3. Search in all collections
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

            except Exception as collection_error:
                logger.warning(
                    "vector_search_collection_failed",
                    extra={
                        "collection": collection,
                        "error": str(collection_error),
                    },
                )
                continue

        # 4. If no results, return immediately
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

        # 5. Apply reranker internally (LLM doesn't see this)
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

        # 6. Return to LLM (already reranked)
        return json.dumps(
            {
                "status": "found",
                "count": len(reranked_results),
                "results": reranked_results,
                "collections_searched": search_collections,
            }
        )

    except Exception as e:
        logger.error(
            "vector_search_failed",
            extra={"query": query[:100], "error": str(e)},
        )
        return json.dumps(
            {
                "status": "failed",
                "error": str(e),
            }
        )


# Build tools list dynamically based on configuration
def _build_tools_list() -> list[BaseTool]:
    """Build the list of available tools based on configuration."""
    tools: list[BaseTool] = [
        multiply,
        web_search,
        fetch_url,
        embedding,
        vector_search,
        call_mcp_tool,
    ]

    config = get_tool_config()
    if config.enable_database:
        tools.append(query_database)

    return tools


TOOLS: list[BaseTool] = _build_tools_list()
TOOLS_BY_NAME: dict[str, BaseTool] = {tool.name: tool for tool in TOOLS}
