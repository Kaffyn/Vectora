import json
import logging
import re
from typing import Any

from langchain.tools import BaseTool, tool
from langchain_community.tools.duckduckgo_search import DuckDuckGoSearchResults
from langchain_community.document_loaders import WebBaseLoader
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
        return "Database tool is disabled. Enable ENABLE_DATABASE=true to use this tool."

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
) -> str:  # noqa: ARG001
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
        logger.error("call_mcp_tool failed to parse arguments", extra={"arguments": arguments})
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


# Build tools list dynamically based on configuration
def _build_tools_list() -> list[BaseTool]:
    """Build the list of available tools based on configuration."""
    tools: list[BaseTool] = [multiply, web_search, fetch_url, call_mcp_tool]

    config = get_tool_config()
    if config.enable_database:
        tools.append(query_database)

    return tools


TOOLS: list[BaseTool] = _build_tools_list()
TOOLS_BY_NAME: dict[str, BaseTool] = {tool.name: tool for tool in TOOLS}
