import json
import logging
from typing import Any

from langchain.tools import BaseTool, tool
from langchain_community.document_loaders import WebBaseLoader

try:
    from langchain_community.tools.duckduckgo_search import DuckDuckGoSearchResults
except ImportError:
    DuckDuckGoSearchResults = None

try:
    from langchain_mcp_adapters import MultiServerMCPClient
except ImportError:
    MultiServerMCPClient = None

from tool_config import get_tool_config

logger = logging.getLogger(__name__)

# Global MCP client cache (reuse connection across calls)
_mcp_client: Any | None = None
_mcp_tools_cache: dict[str, Any] | None = None


@tool
def web_search(query: str) -> str:
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
async def fetch_url(url: str) -> str:
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
    """Get or create global MCP client instance."""
    global _mcp_client

    if _mcp_client is not None:
        return _mcp_client

    if MultiServerMCPClient is None:
        logger.warning("MultiServerMCPClient not available")
        return None

    config = get_tool_config()

    if not config.mcp_servers:
        logger.debug("No MCP servers configured")
        return None

    try:
        _mcp_client = MultiServerMCPClient(servers=config.mcp_servers)
        await _mcp_client.__aenter__()
        logger.info(
            "MCP client initialized", extra={"servers": list(config.mcp_servers.keys())}
        )
        return _mcp_client
    except Exception:
        logger.exception("Failed to initialize MCP client")
        _mcp_client = None
        return None


async def _get_mcp_tools() -> dict[str, Any] | None:
    """Get available MCP tools from initialized client."""
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
    """Call an MCP (Model Context Protocol) tool.

    Args:
        tool_name: Name of the MCP tool to call
        arguments: JSON string of tool arguments

    Returns:
        Result from the MCP tool execution
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
async def vector_search(
    query: str, collection: str = "articles", limit: int = 5
) -> str:
    """Search the vector database for similar documents.

    Args:
        query: Search query string
        collection: Qdrant collection name
        limit: Maximum number of results to return

    Returns:
        JSON formatted search results with documents and scores
    """
    config = get_tool_config()

    if not config.enable_rag:
        logger.warning("vector_search tool called but RAG disabled")
        return "RAG is disabled. Enable ENABLE_RAG=true to use this tool."

    try:
        from langchain_voyageai import VoyageAIEmbeddings
        from qdrant_client import QdrantClient

        if not config.voyage_api_key:
            logger.error("vector_search called but VOYAGE_API_KEY not configured")
            return json.dumps(
                {
                    "status": "failed",
                    "error": "VOYAGE_API_KEY not configured",
                }
            )

        embeddings_model = VoyageAIEmbeddings(
            api_key=config.voyage_api_key,
            model=config.embedding_model,
        )

        query_vector = embeddings_model.embed_query(query)

        qdrant_client = QdrantClient(
            url=config.qdrant_url,
            api_key=config.qdrant_api_key,
        )

        try:
            search_results = qdrant_client.search(
                collection_name=collection,
                query_vector=query_vector,
                limit=limit,
            )
        except Exception:
            logger.warning(
                "Collection not found or search failed",
                extra={"collection": collection},
            )
            return json.dumps(
                {
                    "status": "no_results",
                    "message": f"Collection '{collection}' not found or empty",
                }
            )

        results = [
            {
                "id": point.id,
                "score": point.score,
                "content": point.payload.get("page_content", ""),
                "metadata": point.payload.get("metadata", {}),
            }
            for point in search_results
        ]

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

    except Exception:
        logger.exception(
            "vector_search_failed",
            extra={"query": query, "collection": collection},
        )
        return json.dumps(
            {
                "status": "failed",
                "error": "Vector search failed",
            }
        )


def _build_tools_list() -> list[BaseTool]:
    """Build list of available tools based on configuration.

    Returns:
        List of BaseTool instances
    """
    config = get_tool_config()
    tools: list[BaseTool] = []

    # Core tools
    tools.append(web_search)
    tools.append(fetch_url)
    tools.append(vector_search)

    # MCP tool
    if config.enable_mcp:
        tools.append(call_mcp_tool)

    logger.info("Tools initialized", extra={"count": len(tools)})
    return tools


def get_tools() -> list[BaseTool]:
    """Get list of available tools.

    Returns:
        List of BaseTool instances
    """
    return _build_tools_list()
