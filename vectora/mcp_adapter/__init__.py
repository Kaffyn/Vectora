"""MCP Module: Model Context Protocol integration for Vectora.

Exposes Vectora capabilities via MCP protocol:
- server.py: FastMCP stdio server for Claude Desktop / Claude Code
- client.py: Client for consuming other MCP servers

Usage:
    # Start as MCP server (via pyproject.toml entry point):
    vectora-mcp  →  vectora.mcp_adapter.server:run

    # Connect to other MCP servers (via call_mcp_tool tool):
    from vectora.mcp_adapter.client import MCPClient
"""

from vectora.mcp_adapter.client import MCPClient
from vectora.mcp_adapter.server import run

__all__ = ["MCPClient", "run"]
