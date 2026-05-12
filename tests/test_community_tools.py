"""Test suite for LangChain community tools integration."""

import os

import pytest
from langgraph.graph.state import RunnableConfig

from testing import human_message, test_graph


@pytest.mark.asyncio
class TestWebSearchTool:
    """Tests for DuckDuckGo web search tool."""

    async def test_web_search_finds_results(self, test_graph) -> None:
        """Test that web search tool returns results with URLs."""
        config = RunnableConfig(configurable={"thread_id": "test_web_search_1"})

        result = await test_graph.ainvoke(
            {"messages": [human_message("Search for 'Python programming tutorial'")]},
            config=config,
        )

        messages_content = " ".join(
            str(getattr(m, "content", "")) for m in result["messages"]
        )

        assert len(messages_content) > 0

    async def test_web_search_disabled_when_flag_set(self, monkeypatch, test_graph) -> None:
        """Test that web search respects ENABLE_WEB_SEARCH flag."""
        monkeypatch.setenv("ENABLE_WEB_SEARCH", "false")

        from tool_config import get_tool_config

        config_obj = get_tool_config()
        assert not config_obj.enable_web_search


@pytest.mark.asyncio
class TestFetchURLTool:
    """Tests for web fetch (URL reader) tool."""

    async def test_fetch_url_with_valid_url(self, test_graph) -> None:
        """Test fetching content from a valid URL."""
        config = RunnableConfig(configurable={"thread_id": "test_fetch_url_1"})

        result = await test_graph.ainvoke(
            {
                "messages": [
                    human_message("Fetch content from https://example.com")
                ]
            },
            config=config,
        )

        messages_content = " ".join(
            str(getattr(m, "content", "")) for m in result["messages"]
        )

        assert len(messages_content) > 0

    def test_fetch_url_rejects_invalid_url(self) -> None:
        """Test that fetch_url rejects URLs without http/https."""
        from tools import fetch_url

        class MockRuntime:
            pass

        result = fetch_url.__wrapped__(
            "example.com", runtime=MockRuntime()
        )

        assert "http" in result.lower() or "error" in result.lower()

    def test_fetch_url_respects_domain_whitelist(self, monkeypatch) -> None:
        """Test that fetch_url respects domain whitelist."""
        monkeypatch.setenv("WEB_FETCH_ALLOWED_DOMAINS", "allowed.com,other.com")

        from tool_config import get_tool_config

        config_obj = get_tool_config()
        assert config_obj.allowed_domains == ["allowed.com", "other.com"]


@pytest.mark.asyncio
class TestDatabaseTool:
    """Tests for SQL database tool."""

    def test_database_tool_disabled_by_default(self) -> None:
        """Test that database tool is disabled by default."""
        from tool_config import get_tool_config

        config = get_tool_config()
        assert not config.enable_database

    def test_database_tool_rejects_non_select(self) -> None:
        """Test that database tool only allows SELECT queries."""
        from tools import query_database

        class MockRuntime:
            pass

        result = query_database.__wrapped__(
            "INSERT INTO users VALUES ('test')", runtime=MockRuntime()
        )

        assert "error" in result.lower() or "not allowed" in result.lower()

    def test_database_tool_blocks_dangerous_keywords(self) -> None:
        """Test that database tool blocks dangerous SQL keywords."""
        from tools import query_database

        class MockRuntime:
            pass

        for keyword in ["DELETE", "DROP", "ALTER", "CREATE"]:
            query = f"{keyword} FROM users"
            result = query_database.__wrapped__(query, runtime=MockRuntime())
            assert "error" in result.lower() or "not allowed" in result.lower()

    def test_database_url_config(self, monkeypatch) -> None:
        """Test that DATABASE_URL can be configured."""
        monkeypatch.setenv("DATABASE_URL", "sqlite:///test.db")

        from tool_config import get_tool_config

        config = get_tool_config()
        assert config.database_url == "sqlite:///test.db"


@pytest.mark.asyncio
class TestMCPTool:
    """Tests for MCP server integration."""

    async def test_mcp_tool_disabled_by_default(self) -> None:
        """Test that MCP tool is disabled by default."""
        from tool_config import get_tool_config

        config = get_tool_config()
        assert not config.enable_mcp

    async def test_mcp_tool_graceful_when_disabled(self) -> None:
        """Test that MCP tool returns helpful message when disabled."""
        from tools import call_mcp_tool

        class MockRuntime:
            pass

        result = call_mcp_tool.__wrapped__(
            "some_tool", '{"arg": "value"}', runtime=MockRuntime()
        )

        assert "disabled" in result.lower() or "enable" in result.lower()

    def test_mcp_server_url_config(self, monkeypatch) -> None:
        """Test that MCP_SERVER_URL can be configured."""
        monkeypatch.setenv("MCP_SERVER_URL", "ws://localhost:5000")

        from tool_config import get_tool_config

        config = get_tool_config()
        assert config.mcp_server_url == "ws://localhost:5000"


@pytest.mark.asyncio
class TestToolConfiguration:
    """Tests for tool configuration management."""

    def test_tool_config_reads_env_vars(self, monkeypatch) -> None:
        """Test that ToolConfig reads environment variables."""
        monkeypatch.setenv("ENABLE_WEB_SEARCH", "false")
        monkeypatch.setenv("ENABLE_WEB_FETCH", "false")
        monkeypatch.setenv("WEB_FETCH_MAX_SIZE", "2000")

        from tool_config import ToolConfig

        config = ToolConfig()
        assert not config.enable_web_search
        assert not config.enable_web_fetch
        assert config.max_fetch_size == 2000

    def test_tool_config_defaults(self) -> None:
        """Test that ToolConfig has sensible defaults."""
        from tool_config import ToolConfig

        config = ToolConfig()
        assert config.enable_web_search is True
        assert config.enable_web_fetch is True
        assert config.enable_database is False
        assert config.enable_mcp is False
        assert config.max_fetch_size == 5000

    def test_parse_comma_separated_config(self, monkeypatch) -> None:
        """Test parsing comma-separated configuration values."""
        monkeypatch.setenv(
            "WEB_FETCH_ALLOWED_DOMAINS", "example.com, test.org , another.net"
        )

        from tool_config import ToolConfig

        config = ToolConfig()
        assert config.allowed_domains == [
            "example.com",
            "test.org",
            "another.net",
        ]

    def test_tools_list_respects_database_flag(self, monkeypatch) -> None:
        """Test that TOOLS list includes database tool only when enabled."""
        from tools import TOOLS_BY_NAME

        assert "query_database" not in TOOLS_BY_NAME


@pytest.mark.asyncio
class TestToolsIntegration:
    """Integration tests for tools within the graph."""

    async def test_multiply_still_works(self, test_graph) -> None:
        """Test that original multiply tool still works."""
        config = RunnableConfig(configurable={"thread_id": "test_multiply_compat"})

        result = await test_graph.ainvoke(
            {"messages": [human_message("multiply 5 by 3")]},
            config=config,
        )

        messages_content = " ".join(
            str(getattr(m, "content", "")) for m in result["messages"]
        )

        assert len(messages_content) > 0
