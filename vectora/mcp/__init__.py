"""MCP Module: Model Context Protocol integration for Vectora.

Exposes Vectora capabilities via MCP protocol:
- server.py: FastMCP stdio server for Claude Desktop / Claude Code
- client.py: Client for consuming other MCP servers

Usage:
    # Start as MCP server (via pyproject.toml entry point):
    vectora-mcp  →  vectora.mcp.server:run

    # Connect to other MCP servers (via call_mcp_tool tool):
    from vectora.mcp.client import MCPClient
"""

from vectora.mcp.server import run

__all__ = ["run"]
