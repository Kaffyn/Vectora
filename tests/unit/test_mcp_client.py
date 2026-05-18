"""Testes para o módulo MCP Client."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.mcp.client import MCPClient, MCPToolCallResult


@pytest.fixture
def mcp_client() -> MCPClient:
    """Criar cliente MCP com parâmetros de teste."""
    return MCPClient(command="python", args=["-m", "mcp.server"], env={"TEST": "1"})


@pytest.fixture
def mock_session() -> MagicMock:
    """Criar mock de ClientSession."""
    session = MagicMock()
    session.initialize = AsyncMock(return_value=None)
    session.list_tools = AsyncMock()
    session.call_tool = AsyncMock()
    session.list_resources = AsyncMock()
    session.read_resource = AsyncMock()
    return session


class TestMCPClientInitialization:
    """Testes de inicialização do MCPClient."""

    def test_mcp_client_initializes_with_command(self) -> None:
        """Verificar que MCPClient inicializa com comando."""
        client = MCPClient(command="npx")
        assert client.server_params.command == "npx"
        assert client.server_params.args == []
        assert client.session is None

    def test_mcp_client_initializes_with_args(self) -> None:
        """Verificar que MCPClient inicializa com argumentos."""
        client = MCPClient(command="python", args=["-m", "server"])
        assert client.server_params.command == "python"
        assert client.server_params.args == ["-m", "server"]

    def test_mcp_client_initializes_with_env(self) -> None:
        """Verificar que MCPClient inicializa com variáveis de ambiente."""
        env = {"API_KEY": "test"}
        client = MCPClient(command="node", env=env)
        assert client.server_params.env == env


class TestMCPClientConnect:
    """Testes de conexão do MCPClient."""

    @pytest.mark.asyncio
    async def test_connect_success(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que connect() estabelece conexão com sucesso."""
        with patch("vectora.mcp.client.stdio_client") as mock_stdio:
            read_stream = MagicMock()
            write_stream = MagicMock()
            mock_stdio.return_value.__aenter__.return_value = (
                read_stream,
                write_stream,
            )
            mock_stdio.return_value.__aexit__.return_value = None

            with patch("vectora.mcp.client.ClientSession") as mock_client_session:
                mock_client_session.return_value.__aenter__.return_value = mock_session
                mock_client_session.return_value.__aexit__.return_value = None

                await mcp_client.connect()

                assert mcp_client.session is not None
                mock_session.initialize.assert_called_once()

    @pytest.mark.asyncio
    async def test_connect_failure(self, mcp_client: MCPClient) -> None:
        """Verificar que connect() falha gracefully em caso de erro."""
        with patch(
            "vectora.mcp.client.stdio_client",
            side_effect=RuntimeError("Connection failed"),
        ):
            with pytest.raises(ConnectionError, match="Could not connect"):
                await mcp_client.connect()

    @pytest.mark.asyncio
    async def test_disconnect(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que disconnect() fecha a conexão."""
        mcp_client.session = mock_session

        await mcp_client.disconnect()

        assert mcp_client.session is None


class TestMCPClientContextManager:
    """Testes de context manager do MCPClient."""

    @pytest.mark.asyncio
    async def test_aenter_aexitexit(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que __aenter__ e __aexit__ funcionam."""
        with patch("vectora.mcp.client.stdio_client") as mock_stdio:
            read_stream = MagicMock()
            write_stream = MagicMock()
            mock_stdio.return_value.__aenter__.return_value = (
                read_stream,
                write_stream,
            )
            mock_stdio.return_value.__aexit__.return_value = None

            with patch("vectora.mcp.client.ClientSession") as mock_client_session:
                mock_client_session.return_value.__aenter__.return_value = mock_session
                mock_client_session.return_value.__aexit__.return_value = None

                async with mcp_client as client:
                    assert client.session is not None

                # Após sair do contexto, sessão deve ser desconectada
                assert mcp_client.session is None


class TestMCPClientListTools:
    """Testes de listagem de ferramentas."""

    @pytest.mark.asyncio
    async def test_list_tools_success(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que list_tools() retorna ferramentas."""
        tools = [{"name": "tool1"}, {"name": "tool2"}]
        mock_result = MagicMock()
        mock_result.tools = tools
        mock_session.list_tools.return_value = mock_result

        mcp_client.session = mock_session
        result = await mcp_client.list_tools()

        assert result == tools
        mock_session.list_tools.assert_called_once()

    @pytest.mark.asyncio
    async def test_list_tools_not_connected(self, mcp_client: MCPClient) -> None:
        """Verificar que list_tools() falha se não conectado."""
        with pytest.raises(RuntimeError, match="not connected"):
            await mcp_client.list_tools()

    @pytest.mark.asyncio
    async def test_list_tools_error(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que list_tools() trata erros gracefully."""
        mock_session.list_tools.side_effect = RuntimeError("API error")
        mcp_client.session = mock_session

        with pytest.raises(RuntimeError, match="Failed to list tools"):
            await mcp_client.list_tools()


class TestMCPClientCallTool:
    """Testes de chamada de ferramentas."""

    @pytest.mark.asyncio
    async def test_call_tool_success(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que call_tool() executa ferramenta com sucesso."""
        content_item = MagicMock()
        content_item.model_dump.return_value = {"type": "text", "text": "result"}

        mock_result = MagicMock()
        mock_result.content = [content_item]
        mock_result.isError = False
        mock_session.call_tool.return_value = mock_result

        mcp_client.session = mock_session
        result = await mcp_client.call_tool("test_tool", {"arg": "value"})

        assert isinstance(result, MCPToolCallResult)
        assert result.is_error is False
        assert len(result.content) == 1
        mock_session.call_tool.assert_called_once_with("test_tool", {"arg": "value"})

    @pytest.mark.asyncio
    async def test_call_tool_without_args(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que call_tool() funciona sem argumentos."""
        content_item = MagicMock()
        content_item.model_dump.return_value = {"type": "text", "text": "result"}

        mock_result = MagicMock()
        mock_result.content = [content_item]
        mock_result.isError = False
        mock_session.call_tool.return_value = mock_result

        mcp_client.session = mock_session
        await mcp_client.call_tool("test_tool")

        mock_session.call_tool.assert_called_once_with("test_tool", {})

    @pytest.mark.asyncio
    async def test_call_tool_not_connected(self, mcp_client: MCPClient) -> None:
        """Verificar que call_tool() falha se não conectado."""
        with pytest.raises(RuntimeError, match="not connected"):
            await mcp_client.call_tool("test_tool")

    @pytest.mark.asyncio
    async def test_call_tool_error(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que call_tool() retorna erro em MCPToolCallResult."""
        mock_session.call_tool.side_effect = RuntimeError("Tool execution failed")
        mcp_client.session = mock_session

        result = await mcp_client.call_tool("test_tool")

        assert isinstance(result, MCPToolCallResult)
        assert result.is_error is True
        assert "Error:" in result.content[0]["text"]


class TestMCPClientListResources:
    """Testes de listagem de recursos."""

    @pytest.mark.asyncio
    async def test_list_resources_success(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que list_resources() retorna recursos."""
        resources = [{"uri": "resource1"}, {"uri": "resource2"}]
        mock_result = MagicMock()
        mock_result.resources = resources
        mock_session.list_resources.return_value = mock_result

        mcp_client.session = mock_session
        result = await mcp_client.list_resources()

        assert result == resources
        mock_session.list_resources.assert_called_once()

    @pytest.mark.asyncio
    async def test_list_resources_not_connected(self, mcp_client: MCPClient) -> None:
        """Verificar que list_resources() falha se não conectado."""
        with pytest.raises(RuntimeError, match="not connected"):
            await mcp_client.list_resources()

    @pytest.mark.asyncio
    async def test_list_resources_error(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que list_resources() trata erros gracefully."""
        mock_session.list_resources.side_effect = RuntimeError("API error")
        mcp_client.session = mock_session

        with pytest.raises(RuntimeError, match="Failed to list resources"):
            await mcp_client.list_resources()


class TestMCPClientReadResource:
    """Testes de leitura de recursos."""

    @pytest.mark.asyncio
    async def test_read_resource_success(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que read_resource() retorna conteúdo."""
        content_item = MagicMock()
        content_item.text = "resource content"

        mock_result = MagicMock()
        mock_result.contents = [content_item]
        mock_session.read_resource.return_value = mock_result

        mcp_client.session = mock_session
        result = await mcp_client.read_resource("resource://test")

        assert result == "resource content"
        mock_session.read_resource.assert_called_once_with("resource://test")

    @pytest.mark.asyncio
    async def test_read_resource_multiple_items(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que read_resource() junta múltiplos itens."""
        content_item1 = MagicMock()
        content_item1.text = "line1"
        content_item2 = MagicMock()
        content_item2.text = "line2"

        mock_result = MagicMock()
        mock_result.contents = [content_item1, content_item2]
        mock_session.read_resource.return_value = mock_result

        mcp_client.session = mock_session
        result = await mcp_client.read_resource("resource://test")

        assert result == "line1\nline2"

    @pytest.mark.asyncio
    async def test_read_resource_not_connected(self, mcp_client: MCPClient) -> None:
        """Verificar que read_resource() falha se não conectado."""
        with pytest.raises(RuntimeError, match="not connected"):
            await mcp_client.read_resource("resource://test")

    @pytest.mark.asyncio
    async def test_read_resource_error(
        self, mcp_client: MCPClient, mock_session: MagicMock
    ) -> None:
        """Verificar que read_resource() trata erros gracefully."""
        mock_session.read_resource.side_effect = RuntimeError("Read failed")
        mcp_client.session = mock_session

        with pytest.raises(RuntimeError, match="Failed to read resource"):
            await mcp_client.read_resource("resource://test")


class TestMCPToolCallResult:
    """Testes para MCPToolCallResult."""

    def test_mcp_tool_call_result_default(self) -> None:
        """Verificar que MCPToolCallResult inicializa com defaults."""
        result = MCPToolCallResult()
        assert result.content == []
        assert result.is_error is False

    def test_mcp_tool_call_result_with_values(self) -> None:
        """Verificar que MCPToolCallResult armazena valores."""
        content = [{"type": "text", "text": "output"}]
        result = MCPToolCallResult(content=content, is_error=True)
        assert result.content == content
        assert result.is_error is True
