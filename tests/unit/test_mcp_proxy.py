"""Testes para VectoraProxy - cliente helper para múltiplos agentes via MCP."""

from __future__ import annotations

import json
from contextlib import AsyncExitStack
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.mcp.proxy import (
    VectoraProxy,
    VectoraProxyError,
    create_local_proxy,
    create_remote_proxy,
)


class TestVectoraProxyInitialization:
    """Testes de inicialização do VectoraProxy."""

    def test_proxy_initializes_stdio_with_command(self) -> None:
        """Verificar que proxy inicializa para transport stdio."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        assert proxy.transport == "stdio"
        assert proxy.command == "uv"
        assert proxy.args == []
        assert proxy.timeout == 300.0
        assert proxy._session is None

    def test_proxy_initializes_stdio_with_args(self) -> None:
        """Verificar que proxy armazena args corretamente."""
        args = ["run", "vectora-mcp"]
        proxy = VectoraProxy(transport="stdio", command="uv", args=args)
        assert proxy.args == args

    def test_proxy_initializes_stdio_with_env(self) -> None:
        """Verificar que proxy armazena env vars corretamente."""
        env = {"API_KEY": "test"}
        proxy = VectoraProxy(transport="stdio", command="uv", env=env)
        assert proxy.env == env

    def test_proxy_initializes_sse_with_url(self) -> None:
        """Verificar que proxy inicializa para transport SSE."""
        proxy = VectoraProxy(transport="sse", url="http://localhost:8000/sse")
        assert proxy.transport == "sse"
        assert proxy.url == "http://localhost:8000/sse"

    def test_proxy_initializes_with_custom_timeout(self) -> None:
        """Verificar que proxy aceita timeout customizado."""
        proxy = VectoraProxy(transport="stdio", command="uv", timeout=600.0)
        assert proxy.timeout == 600.0

    def test_proxy_raises_error_sse_without_url(self) -> None:
        """Verificar que SSE requer URL."""
        with pytest.raises(ValueError, match="url é obrigatório"):
            VectoraProxy(transport="sse")

    def test_proxy_raises_error_stdio_without_command(self) -> None:
        """Verificar que stdio requer command."""
        with pytest.raises(ValueError, match="command é obrigatório"):
            VectoraProxy(transport="stdio")

    def test_proxy_initializes_exit_stack_as_none(self) -> None:
        """Verificar que _exit_stack começa como None."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        assert proxy._exit_stack is None


class TestVectoraProxyContextManager:
    """Testes de context manager assíncrono do VectoraProxy."""

    @pytest.mark.asyncio
    async def test_aenter_calls_connect(self) -> None:
        """Verificar que __aenter__ chama connect()."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        with patch.object(proxy, "connect", new_callable=AsyncMock):
            await proxy.__aenter__()
            proxy.connect.assert_called_once()

    @pytest.mark.asyncio
    async def test_aexit_calls_disconnect(self) -> None:
        """Verificar que __aexit__ chama disconnect()."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        with patch.object(proxy, "disconnect", new_callable=AsyncMock):
            await proxy.__aexit__(None, None, None)
            proxy.disconnect.assert_called_once()

    @pytest.mark.asyncio
    async def test_context_manager_flow(self) -> None:
        """Verificar que async with proxy funciona."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        with patch.object(proxy, "connect", new_callable=AsyncMock):
            with patch.object(proxy, "disconnect", new_callable=AsyncMock):
                async with proxy as p:
                    assert p is proxy
                proxy.disconnect.assert_called_once()


class TestVectoraProxyConnect:
    """Testes de conexão do VectoraProxy."""

    @pytest.mark.asyncio
    async def test_connect_stdio_success(self) -> None:
        """Verificar que connect() estabelece conexão stdio."""
        proxy = VectoraProxy(transport="stdio", command="uv", args=["run", "test"])

        with patch("mcp.client.stdio.stdio_client") as mock_stdio:
            read_stream = MagicMock()
            write_stream = MagicMock()
            mock_stdio.return_value.__aenter__.return_value = (
                read_stream,
                write_stream,
            )
            mock_stdio.return_value.__aexit__.return_value = None

            with patch("mcp.ClientSession") as mock_session_class:
                mock_session = MagicMock()
                mock_session.initialize = AsyncMock()
                mock_session_class.return_value.__aenter__.return_value = mock_session
                mock_session_class.return_value.__aexit__.return_value = None

                await proxy.connect()

                assert proxy._session is not None
                assert proxy._exit_stack is not None
                mock_session.initialize.assert_called_once()

    @pytest.mark.asyncio
    async def test_connect_sse_success(self) -> None:
        """Verificar que connect() estabelece conexão SSE."""
        proxy = VectoraProxy(transport="sse", url="http://localhost:8000/sse")

        with patch("mcp.client.sse.sse_client") as mock_sse:
            read_stream = MagicMock()
            write_stream = MagicMock()
            mock_sse.return_value.__aenter__.return_value = (
                read_stream,
                write_stream,
            )
            mock_sse.return_value.__aexit__.return_value = None

            with patch("mcp.ClientSession") as mock_session_class:
                mock_session = MagicMock()
                mock_session.initialize = AsyncMock()
                mock_session_class.return_value.__aenter__.return_value = mock_session
                mock_session_class.return_value.__aexit__.return_value = None

                await proxy.connect()

                assert proxy._session is not None
                mock_sse.assert_called_once_with(url="http://localhost:8000/sse")

    @pytest.mark.asyncio
    async def test_connect_failure_raises_error(self) -> None:
        """Verificar que erro na conexão lança VectoraProxyError."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        with patch(
            "mcp.client.stdio.stdio_client",
            side_effect=RuntimeError("Connection failed"),
        ):
            with pytest.raises(VectoraProxyError, match="Falha ao conectar"):
                await proxy.connect()

    @pytest.mark.asyncio
    async def test_disconnect_success(self) -> None:
        """Verificar que disconnect() fecha conexão."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        mock_exit_stack = AsyncMock(spec=AsyncExitStack)
        proxy._exit_stack = mock_exit_stack
        proxy._session = MagicMock()

        await proxy.disconnect()

        mock_exit_stack.aclose.assert_called_once()
        assert proxy._exit_stack is None
        assert proxy._session is None

    @pytest.mark.asyncio
    async def test_disconnect_with_error_logs_warning(self) -> None:
        """Verificar que erro no disconnect loga warning."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        proxy._exit_stack = MagicMock(spec=AsyncExitStack)
        proxy._exit_stack.aclose = AsyncMock(side_effect=RuntimeError("Close failed"))
        proxy._session = MagicMock()

        with patch("vectora.mcp.proxy.logger") as mock_logger:
            await proxy.disconnect()
            mock_logger.warning.assert_called_once()
            assert proxy._exit_stack is None

    @pytest.mark.asyncio
    async def test_disconnect_when_not_connected(self) -> None:
        """Verificar que disconnect() sem conexão não falha."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        # _exit_stack é None, não deve raise
        await proxy.disconnect()
        assert proxy._exit_stack is None


class TestVectoraProxyDelegate:
    """Testes do método delegate()."""

    @pytest.mark.asyncio
    async def test_delegate_success_with_string_thread_id(self) -> None:
        """Verificar que delegate() executa com thread_id string numérica."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_content = MagicMock()
        mock_content.text = "result"
        mock_result = MagicMock()
        mock_result.content = [mock_content]
        mock_session.call_tool = AsyncMock(return_value=mock_result)
        proxy._session = mock_session

        result = await proxy.delegate("Analise código", thread_id="123")

        assert "result" in result
        mock_session.call_tool.assert_called_once_with(
            "delegate_task_to_vectora",
            arguments={"task_prompt": "Analise código", "thread_id": 123},
        )

    @pytest.mark.asyncio
    async def test_delegate_success_with_int_thread_id(self) -> None:
        """Verificar que delegate() executa com thread_id int."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_content = MagicMock()
        mock_content.text = "result"
        mock_result = MagicMock()
        mock_result.content = [mock_content]
        mock_session.call_tool = AsyncMock(return_value=mock_result)
        proxy._session = mock_session

        result = await proxy.delegate("Tarefa", thread_id=42)

        assert result == "result"
        mock_session.call_tool.assert_called_once()

    @pytest.mark.asyncio
    async def test_delegate_not_connected_raises_error(self) -> None:
        """Verificar que delegate() sem conexão lança erro."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        proxy._session = None

        with pytest.raises(VectoraProxyError, match="não conectado"):
            await proxy.delegate("Tarefa")

    @pytest.mark.asyncio
    async def test_delegate_call_failure_raises_error(self) -> None:
        """Verificar que erro na call_tool lança VectoraProxyError."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_session.call_tool = AsyncMock(side_effect=RuntimeError("API error"))
        proxy._session = mock_session

        with pytest.raises(VectoraProxyError, match="Falha na delegação"):
            await proxy.delegate("Tarefa")


class TestVectoraProxyCallTool:
    """Testes do método call_tool()."""

    @pytest.mark.asyncio
    async def test_call_tool_success_with_arguments(self) -> None:
        """Verificar que call_tool() executa com argumentos."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_content = MagicMock()
        mock_content.text = "search result"
        mock_result = MagicMock()
        mock_result.content = [mock_content]
        mock_session.call_tool = AsyncMock(return_value=mock_result)
        proxy._session = mock_session

        result = await proxy.call_tool("vector_search_tool", {"query": "RAG"})

        assert "search result" in result
        mock_session.call_tool.assert_called_once_with(
            "vector_search_tool",
            arguments={"query": "RAG"},
        )

    @pytest.mark.asyncio
    async def test_call_tool_success_without_arguments(self) -> None:
        """Verificar que call_tool() funciona sem argumentos."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_content = MagicMock()
        mock_content.text = "result"
        mock_result = MagicMock()
        mock_result.content = [mock_content]
        mock_session.call_tool = AsyncMock(return_value=mock_result)
        proxy._session = mock_session

        await proxy.call_tool("some_tool")

        mock_session.call_tool.assert_called_once_with(
            "some_tool",
            arguments={},
        )

    @pytest.mark.asyncio
    async def test_call_tool_not_connected_raises_error(self) -> None:
        """Verificar que call_tool() sem conexão lança erro."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        proxy._session = None

        with pytest.raises(VectoraProxyError, match="não conectado"):
            await proxy.call_tool("tool_name")

    @pytest.mark.asyncio
    async def test_call_tool_failure_raises_error(self) -> None:
        """Verificar que erro na ferramenta lança VectoraProxyError."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_session.call_tool = AsyncMock(side_effect=RuntimeError("Tool error"))
        proxy._session = mock_session

        with pytest.raises(VectoraProxyError, match="Falha na ferramenta"):
            await proxy.call_tool("tool_name")


class TestVectoraProxyGetResource:
    """Testes do método get_resource()."""

    @pytest.mark.asyncio
    async def test_get_resource_success(self) -> None:
        """Verificar que get_resource() lê resource."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_content = MagicMock()
        mock_content.text = "resource data"
        mock_result = MagicMock()
        mock_result.contents = [mock_content]
        mock_session.read_resource = AsyncMock(return_value=mock_result)
        proxy._session = mock_session

        result = await proxy.get_resource("vectora://thread/1/context")

        assert result == "resource data"
        mock_session.read_resource.assert_called_once_with("vectora://thread/1/context")

    @pytest.mark.asyncio
    async def test_get_resource_fallback_to_str(self) -> None:
        """Verificar que get_resource() faz fallback para str()."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_result = MagicMock()
        mock_result.contents = None  # sem contents
        mock_session.read_resource = AsyncMock(return_value=mock_result)
        proxy._session = mock_session

        result = await proxy.get_resource("vectora://status")

        assert isinstance(result, str)

    @pytest.mark.asyncio
    async def test_get_resource_not_connected_raises_error(self) -> None:
        """Verificar que get_resource() sem conexão lança erro."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        proxy._session = None

        with pytest.raises(VectoraProxyError, match="não conectado"):
            await proxy.get_resource("vectora://status")

    @pytest.mark.asyncio
    async def test_get_resource_failure_raises_error(self) -> None:
        """Verificar que erro na leitura lança VectoraProxyError."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_session = MagicMock()
        mock_session.read_resource = AsyncMock(side_effect=RuntimeError("Read failed"))
        proxy._session = mock_session

        with pytest.raises(VectoraProxyError, match="Falha ao ler resource"):
            await proxy.get_resource("vectora://status")


class TestVectoraProxyListTools:
    """Testes do método list_tools()."""

    @pytest.mark.asyncio
    async def test_list_tools_success(self) -> None:
        """Verificar que list_tools() retorna lista de ferramentas."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        mock_tool1 = MagicMock()
        mock_tool1.name = "tool1"
        mock_tool1.description = "Description 1"
        mock_tool1.inputSchema = {"type": "object"}

        mock_tool2 = MagicMock()
        mock_tool2.name = "tool2"
        mock_tool2.description = "Description 2"
        mock_tool2.inputSchema = {"type": "object"}

        mock_session = MagicMock()
        mock_result = MagicMock()
        mock_result.tools = [mock_tool1, mock_tool2]
        mock_session.list_tools = AsyncMock(return_value=mock_result)
        proxy._session = mock_session

        result = await proxy.list_tools()

        assert len(result) == 2
        assert result[0]["name"] == "tool1"
        assert result[1]["name"] == "tool2"

    @pytest.mark.asyncio
    async def test_list_tools_not_connected_raises_error(self) -> None:
        """Verificar que list_tools() sem conexão lança erro."""
        proxy = VectoraProxy(transport="stdio", command="uv")
        proxy._session = None

        with pytest.raises(VectoraProxyError, match="não conectado"):
            await proxy.list_tools()


class TestVectoraProxyThreadContext:
    """Testes dos métodos get_thread_context() e get_thread_history()."""

    @pytest.mark.asyncio
    async def test_get_thread_context_valid_json(self) -> None:
        """Verificar que get_thread_context() parseia JSON."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        context_data = {"message_count": 5, "status": "active"}
        with patch.object(
            proxy,
            "get_resource",
            new_callable=AsyncMock,
            return_value=json.dumps(context_data),
        ):
            result = await proxy.get_thread_context("thread_1")

            assert result["message_count"] == 5
            assert result["status"] == "active"

    @pytest.mark.asyncio
    async def test_get_thread_context_invalid_json(self) -> None:
        """Verificar que get_thread_context() faz fallback para raw."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        with patch.object(
            proxy,
            "get_resource",
            new_callable=AsyncMock,
            return_value="invalid json {",
        ):
            result = await proxy.get_thread_context("thread_1")

            assert "raw" in result
            assert result["raw"] == "invalid json {"

    @pytest.mark.asyncio
    async def test_get_thread_history_valid_json(self) -> None:
        """Verificar que get_thread_history() parseia JSON."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        history_data = {"recent_messages": [{"role": "user", "content": "Hi"}]}
        with patch.object(
            proxy,
            "get_resource",
            new_callable=AsyncMock,
            return_value=json.dumps(history_data),
        ):
            result = await proxy.get_thread_history("thread_1")

            assert len(result["recent_messages"]) == 1

    @pytest.mark.asyncio
    async def test_get_thread_history_invalid_json(self) -> None:
        """Verificar que get_thread_history() faz fallback para raw."""
        proxy = VectoraProxy(transport="stdio", command="uv")

        with patch.object(
            proxy,
            "get_resource",
            new_callable=AsyncMock,
            return_value="not json [",
        ):
            result = await proxy.get_thread_history("thread_1")

            assert "raw" in result


class TestVectoraProxyExtractText:
    """Testes do método estático _extract_text()."""

    def test_extract_text_with_content(self) -> None:
        """Verificar que _extract_text() extrai de content."""
        mock_item = MagicMock()
        mock_item.text = "extracted text"

        mock_result = MagicMock()
        mock_result.content = [mock_item]

        result = VectoraProxy._extract_text(mock_result)

        assert result == "extracted text"

    def test_extract_text_multiple_items(self) -> None:
        """Verificar que _extract_text() junta múltiplos items."""
        mock_item1 = MagicMock()
        mock_item1.text = "text1"
        mock_item2 = MagicMock()
        mock_item2.text = "text2"

        mock_result = MagicMock()
        mock_result.content = [mock_item1, mock_item2]

        result = VectoraProxy._extract_text(mock_result)

        assert "text1" in result
        assert "text2" in result
        assert "\n" in result

    def test_extract_text_without_content(self) -> None:
        """Verificar que _extract_text() faz fallback para str()."""
        mock_result = MagicMock()
        mock_result.content = None
        mock_result.__str__ = MagicMock(return_value="fallback")

        result = VectoraProxy._extract_text(mock_result)

        assert result == "fallback"

    def test_extract_text_with_non_text_items(self) -> None:
        """Verificar que _extract_text() converte items sem text."""
        mock_item1 = MagicMock()
        mock_item1.text = "text1"
        mock_item2 = MagicMock(spec=[])  # sem text

        mock_result = MagicMock()
        mock_result.content = [mock_item1, mock_item2]

        result = VectoraProxy._extract_text(mock_result)

        assert "text1" in result


class TestVectoraProxyFactories:
    """Testes das factory functions."""

    def test_create_local_proxy(self) -> None:
        """Verificar que create_local_proxy() cria proxy stdio."""
        proxy = create_local_proxy()

        assert proxy.transport == "stdio"
        assert proxy.command == "uv"
        assert proxy.args == ["run", "vectora-mcp"]
        assert proxy.timeout == 300.0

    def test_create_local_proxy_custom_timeout(self) -> None:
        """Verificar que create_local_proxy() aceita timeout customizado."""
        proxy = create_local_proxy(timeout=600.0)

        assert proxy.timeout == 600.0

    def test_create_remote_proxy(self) -> None:
        """Verificar que create_remote_proxy() cria proxy SSE."""
        proxy = create_remote_proxy()

        assert proxy.transport == "sse"
        assert proxy.url == "http://localhost:8000/sse"
        assert proxy.timeout == 300.0

    def test_create_remote_proxy_custom_url(self) -> None:
        """Verificar que create_remote_proxy() aceita URL customizada."""
        proxy = create_remote_proxy(url="http://vectora.prod:8000/sse")

        assert proxy.url == "http://vectora.prod:8000/sse"

    def test_create_remote_proxy_custom_timeout(self) -> None:
        """Verificar que create_remote_proxy() aceita timeout customizado."""
        proxy = create_remote_proxy(timeout=900.0)

        assert proxy.timeout == 900.0


class TestVectoraProxyError:
    """Testes para VectoraProxyError."""

    def test_proxy_error_is_exception(self) -> None:
        """Verificar que VectoraProxyError é uma Exception."""
        error = VectoraProxyError("test error")
        assert isinstance(error, Exception)
        assert str(error) == "test error"
