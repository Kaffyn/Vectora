"""Testes para tools/mcp.py.

Cobre: _get_mcp_client, _get_mcp_tools, call_mcp_tool
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

import pytest


def _reset_mcp_globals() -> None:
    """Reseta o estado global do módulo mcp entre testes."""
    import vectora.tools.mcp as mcp_mod

    mcp_mod._mcp_client = None
    mcp_mod._mcp_tools_cache = None


@pytest.fixture(autouse=True)
def reset_mcp_state():
    _reset_mcp_globals()
    yield
    _reset_mcp_globals()


class TestGetMcpClient:
    """Testa _get_mcp_client."""

    @pytest.mark.asyncio
    async def test_returns_none_when_no_multiserver_client(self):
        """Verifica que retorna None quando MultiServerMCPClient não está disponível."""
        from vectora.tools.mcp import _get_mcp_client

        with patch("vectora.tools.mcp.MultiServerMCPClient", None):
            result = await _get_mcp_client()
        assert result is None

    @pytest.mark.asyncio
    async def test_returns_none_when_no_server_configured(self):
        """Verifica que retorna None quando não há servidor MCP configurado."""
        from vectora.tools.mcp import _get_mcp_client

        mock_client_class = MagicMock()
        with patch("vectora.tools.mcp.MultiServerMCPClient", mock_client_class):
            with patch("vectora.tools.mcp.settings") as mock_settings:
                mock_settings.mcp_server_url = None
                mock_settings.mcp_command = None
                result = await _get_mcp_client()
        assert result is None

    @pytest.mark.asyncio
    async def test_returns_cached_client(self):
        """Verifica que retorna cliente cacheado sem criar novo."""
        import vectora.tools.mcp as mcp_mod
        from vectora.tools.mcp import _get_mcp_client

        mock_client = MagicMock()
        mcp_mod._mcp_client = mock_client
        result = await _get_mcp_client()
        assert result is mock_client


class TestGetMcpTools:
    """Testa _get_mcp_tools."""

    @pytest.mark.asyncio
    async def test_returns_none_when_client_unavailable(self):
        """Verifica que retorna None quando cliente não disponível."""
        from vectora.tools.mcp import _get_mcp_tools

        with patch("vectora.tools.mcp.MultiServerMCPClient", None):
            result = await _get_mcp_tools()
        assert result is None

    @pytest.mark.asyncio
    async def test_returns_cached_tools(self):
        """Verifica que retorna ferramentas cacheadas."""
        import vectora.tools.mcp as mcp_mod
        from vectora.tools.mcp import _get_mcp_tools

        cached = {"tool_a": MagicMock()}
        mcp_mod._mcp_tools_cache = cached
        result = await _get_mcp_tools()
        assert result is cached


class TestCallMcpTool:
    """Testa call_mcp_tool."""

    @pytest.mark.asyncio
    async def test_returns_message_when_client_not_available(self):
        """Verifica que retorna mensagem quando MultiServerMCPClient não disponível."""
        from vectora.tools.mcp import call_mcp_tool

        with patch("vectora.tools.mcp.MultiServerMCPClient", None):
            result = await call_mcp_tool.ainvoke(
                {
                    "tool_name": "test_tool",
                    "arguments": "{}",
                }
            )
        assert "not available" in result.lower() or "MCP" in result

    @pytest.mark.asyncio
    async def test_returns_message_when_mcp_disabled(self):
        """Verifica que retorna mensagem quando MCP está desabilitado."""
        from vectora.tools.mcp import call_mcp_tool

        mock_client_class = MagicMock()
        with patch("vectora.tools.mcp.MultiServerMCPClient", mock_client_class):
            with patch("vectora.tools.mcp.settings") as mock_settings:
                mock_settings.enable_mcp = False
                mock_settings.mcp_server_url = None
                result = await call_mcp_tool.ainvoke(
                    {
                        "tool_name": "test_tool",
                        "arguments": "{}",
                    }
                )
        assert "disabled" in result.lower() or "MCP" in result

    @pytest.mark.asyncio
    async def test_returns_error_on_exception(self):
        """Verifica que retorna mensagem de erro em caso de exceção."""
        from vectora.tools.mcp import call_mcp_tool

        mock_client_class = MagicMock(side_effect=Exception("connection error"))
        with patch("vectora.tools.mcp.MultiServerMCPClient", mock_client_class):
            with patch("vectora.tools.mcp.settings") as mock_settings:
                mock_settings.enable_mcp = True
                mock_settings.mcp_server_url = "http://localhost:8000"
                mock_settings.mcp_transport_type = "http"
                mock_settings.mcp_timeout = 30
                result = await call_mcp_tool.ainvoke(
                    {
                        "tool_name": "some_tool",
                        "arguments": '{"key": "value"}',
                    }
                )
        assert "Error" in result
