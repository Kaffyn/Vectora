"""Tests for MCP (Model Context Protocol) integration in Vectora tools."""

from typing import Self
from unittest.mock import patch

import pytest

from src.tools import _get_mcp_client, _get_mcp_tools


class MockMCPTool:
    """Mock MCP tool for testing."""

    def __init__(self: Self, name: str, description: str = "Mock tool") -> None:
        self.name = name
        self.description = description


class MockMCPClient:
    """Mock MultiServerMCPClient for testing."""

    async def get_tools(self):
        """Return mock tools."""
        return [
            MockMCPTool("add", "Add two numbers"),
            MockMCPTool("multiply", "Multiply two numbers"),
            MockMCPTool("greet", "Greet a person"),
        ]

    async def call_tool(self, name: str, args: dict):
        """Execute mock tool."""
        if name == "add":
            return args.get("a", 0) + args.get("b", 0)
        if name == "multiply":
            return args.get("a", 0) * args.get("b", 0)
        if name == "greet":
            return f"Hello, {args.get('name', 'World')}!"
        msg = f"Unknown tool: {name}"
        raise ValueError(msg)


@pytest.fixture
def reset_mcp_cache():
    """Reset MCP client cache between tests."""
    import src.tools

    src.tools._mcp_client = None
    src.tools._mcp_tools_cache = None
    yield
    src.tools._mcp_client = None
    src.tools._mcp_tools_cache = None


class TestMCPClientConnection:
    """Test MCP client connection and caching."""

    @pytest.mark.asyncio
    async def test_mcp_client_disabled(self, reset_mcp_cache):
        """Test that client is None when MCP is disabled."""
        with patch("src.tools.get_tool_config") as mock_config:
            mock_config.return_value.enable_mcp = False
            client = await _get_mcp_client()
            assert client is None

    @pytest.mark.asyncio
    async def test_mcp_client_missing_config(self, reset_mcp_cache):
        """Test that client is None when server not configured."""
        with patch("src.tools.get_tool_config") as mock_config:
            mock_config.return_value.enable_mcp = True
            mock_config.return_value.mcp_server_url = None
            mock_config.return_value.mcp_command = None
            client = await _get_mcp_client()
            assert client is None

    @pytest.mark.asyncio
    async def test_mcp_client_http_transport(self, reset_mcp_cache):
        """Test HTTP client creation."""
        with (
            patch("src.tools.get_tool_config") as mock_config,
            patch("src.tools.MultiServerMCPClient") as mock_mcp_class,
        ):
            mock_config.return_value.enable_mcp = True
            mock_config.return_value.mcp_server_url = "http://localhost:8000"
            mock_config.return_value.mcp_transport_type = "http"

            mock_mcp_class.return_value = MockMCPClient()

            client = await _get_mcp_client()
            assert client is not None
            mock_mcp_class.assert_called_once()

    @pytest.mark.asyncio
    async def test_mcp_client_stdio_transport(self, reset_mcp_cache):
        """Test stdio transport configuration."""
        with (
            patch("src.tools.get_tool_config") as mock_config,
            patch("src.tools.MultiServerMCPClient") as mock_mcp_class,
        ):
            mock_config.return_value.enable_mcp = True
            mock_config.return_value.mcp_server_url = None
            mock_config.return_value.mcp_transport_type = "stdio"
            mock_config.return_value.mcp_command = "npx"
            mock_config.return_value.mcp_command_args = ["chrome-devtools-mcp@latest"]

            mock_mcp_class.return_value = MockMCPClient()

            client = await _get_mcp_client()
            assert client is not None

            # Verify the client was created with stdio config
            call_args = mock_mcp_class.call_args
            assert call_args is not None
            config = call_args[0][0]["mcp_server"]
            assert config["transport"] == "stdio"
            assert config["command"] == "npx"

    @pytest.mark.asyncio
    async def test_mcp_client_caching(self, reset_mcp_cache):
        """Test that MCP client is cached and reused."""
        with (
            patch("src.tools.get_tool_config") as mock_config,
            patch("src.tools.MultiServerMCPClient") as mock_mcp_class,
        ):
            mock_config.return_value.enable_mcp = True
            mock_config.return_value.mcp_server_url = "http://localhost:8000"
            mock_config.return_value.mcp_transport_type = "http"

            mock_mcp_class.return_value = MockMCPClient()

            # First call creates client
            client1 = await _get_mcp_client()
            # Second call should return cached client
            client2 = await _get_mcp_client()

            assert client1 is client2
            # Constructor called only once
            mock_mcp_class.assert_called_once()


class TestMCPToolsRetrieval:
    """Test MCP tools retrieval and caching."""

    @pytest.mark.asyncio
    async def test_get_mcp_tools_success(self, reset_mcp_cache):
        """Test successful retrieval of MCP tools."""
        with (
            patch("src.tools.get_tool_config") as mock_config,
            patch("src.tools.MultiServerMCPClient") as mock_mcp_class,
        ):
            mock_config.return_value.enable_mcp = True
            mock_config.return_value.mcp_server_url = "http://localhost:8000"
            mock_config.return_value.mcp_transport_type = "http"

            mock_mcp_class.return_value = MockMCPClient()

            tools = await _get_mcp_tools()
            assert tools is not None
            assert "add" in tools
            assert "multiply" in tools
            assert "greet" in tools

    @pytest.mark.asyncio
    async def test_get_mcp_tools_caching(self, reset_mcp_cache):
        """Test that MCP tools are cached."""
        with (
            patch("src.tools.get_tool_config") as mock_config,
            patch("src.tools.MultiServerMCPClient") as mock_mcp_class,
        ):
            mock_config.return_value.enable_mcp = True
            mock_config.return_value.mcp_server_url = "http://localhost:8000"
            mock_config.return_value.mcp_transport_type = "http"

            mock_mcp_class.return_value = MockMCPClient()

            # First call retrieves tools
            tools1 = await _get_mcp_tools()
            # Second call should return cached tools
            tools2 = await _get_mcp_tools()

            assert tools1 is tools2

    @pytest.mark.asyncio
    async def test_get_mcp_tools_connection_failed(self, reset_mcp_cache):
        """Test graceful failure when connection fails."""
        with (
            patch("src.tools.get_tool_config") as mock_config,
            patch("src.tools._get_mcp_client") as mock_get_client,
        ):
            mock_config.return_value.enable_mcp = True
            mock_get_client.return_value = None

            tools = await _get_mcp_tools()
            assert tools is None


class TestMCPIntegrationScenarios:
    """Integration scenarios for MCP."""

    @pytest.mark.asyncio
    async def test_mcp_workflow_http_transport(self, reset_mcp_cache):
        """Test complete MCP workflow with HTTP transport."""
        with (
            patch("src.tools.get_tool_config") as mock_config,
            patch("src.tools.MultiServerMCPClient") as mock_mcp_class,
        ):
            mock_config.return_value.enable_mcp = True
            mock_config.return_value.mcp_server_url = "http://localhost:8000"
            mock_config.return_value.mcp_transport_type = "http"

            mock_client = MockMCPClient()
            mock_mcp_class.return_value = mock_client

            # Get client
            client = await _get_mcp_client()
            assert client is not None

            # List tools
            tools = await _get_mcp_tools()
            assert len(tools) == 3

    @pytest.mark.asyncio
    async def test_mcp_workflow_stdio_transport(self, reset_mcp_cache):
        """Test complete MCP workflow with stdio transport."""
        with (
            patch("src.tools.get_tool_config") as mock_config,
            patch("src.tools.MultiServerMCPClient") as mock_mcp_class,
        ):
            mock_config.return_value.enable_mcp = True
            mock_config.return_value.mcp_server_url = None
            mock_config.return_value.mcp_transport_type = "stdio"
            mock_config.return_value.mcp_command = "npx"
            mock_config.return_value.mcp_command_args = ["chrome-devtools-mcp@latest"]

            mock_client = MockMCPClient()
            mock_mcp_class.return_value = mock_client

            # Get client
            client = await _get_mcp_client()
            assert client is not None

            # List tools
            tools = await _get_mcp_tools()
            assert "add" in tools
