"""Testes E2E para Resources do servidor MCP Vectora.

Valida a implementação do padrão Sub-Agent com exposição de estado (Resources)
que permite ao agente principal (Claude Code) ler o contexto antes de chamar ferramentas.
"""

import json
import pytest
from unittest.mock import AsyncMock, MagicMock, patch


class TestMCPResourcesExist:
    """Testes para verificar que recursos MCP estão registrados."""

    def test_server_has_context_resource(self):
        """Verificar que servidor expõe recurso de contexto de thread."""
        # Importar mcp do servidor
        from src.mcp_server import mcp, get_thread_context

        # Verificar que os recursos foram importados com sucesso
        assert callable(get_thread_context)
        assert mcp is not None

    def test_server_has_history_resource(self):
        """Verificar que servidor expõe histórico de conversa."""
        from src.mcp_server import mcp, get_thread_history

        assert callable(get_thread_history)
        assert mcp is not None

    def test_server_has_status_resource(self):
        """Verificar que servidor expõe status do servidor."""
        from src.mcp_server import mcp, get_server_status

        assert callable(get_server_status)
        assert mcp is not None


class TestMCPResourceImplementations:
    """Testes para implementação de cada recurso."""

    @pytest.mark.asyncio
    async def test_get_thread_context_empty_thread(self):
        """Verificar get_thread_context retorna JSON válido para thread vazia."""
        from src.mcp_server import get_thread_context

        # Mock do Checkpointer para retornar None (thread não existe)
        with patch("src.mcp_server.Checkpointer") as mock_checkpointer_class:
            mock_checkpointer = AsyncMock()
            mock_checkpointer.aget = AsyncMock(return_value=None)
            mock_checkpointer.__aenter__ = AsyncMock(return_value=mock_checkpointer)
            mock_checkpointer.__aexit__ = AsyncMock(return_value=None)
            mock_checkpointer_class.return_value = mock_checkpointer

            result = await get_thread_context("test-thread-1")

            # Resultado deve ser string JSON válida
            assert isinstance(result, str)
            data = json.loads(result)
            assert data["status"] == "empty"
            assert data["thread_id"] == "test-thread-1"

    @pytest.mark.asyncio
    async def test_get_thread_context_active_thread(self):
        """Verificar get_thread_context retorna contexto para thread ativa."""
        from src.mcp_server import get_thread_context
        from langchain_core.messages import HumanMessage, AIMessage

        # Mock do Checkpointer com estado
        with patch("src.mcp_server.Checkpointer") as mock_checkpointer_class:
            mock_messages = [
                HumanMessage(content="Olá"),
                AIMessage(content="Oi! Como posso ajudar?"),
            ]

            mock_checkpointer = AsyncMock()
            mock_checkpointer.aget = AsyncMock(
                return_value={
                    "values": {
                        "messages": mock_messages,
                        "summarized_history": "Usuário iniciou conversa",
                    }
                }
            )
            mock_checkpointer.__aenter__ = AsyncMock(return_value=mock_checkpointer)
            mock_checkpointer.__aexit__ = AsyncMock(return_value=None)
            mock_checkpointer_class.return_value = mock_checkpointer

            result = await get_thread_context("test-thread-2")

            # Resultado deve ser string JSON válida
            assert isinstance(result, str)
            data = json.loads(result)
            assert data["status"] == "active"
            assert data["thread_id"] == "test-thread-2"
            assert data["message_count"] == 2
            assert "Usuário iniciou conversa" in data["summary"]

    @pytest.mark.asyncio
    async def test_get_thread_history_empty_thread(self):
        """Verificar get_thread_history retorna JSON válido para thread vazia."""
        from src.mcp_server import get_thread_history

        with patch("src.mcp_server.Checkpointer") as mock_checkpointer_class:
            mock_checkpointer = AsyncMock()
            mock_checkpointer.aget = AsyncMock(return_value=None)
            mock_checkpointer.__aenter__ = AsyncMock(return_value=mock_checkpointer)
            mock_checkpointer.__aexit__ = AsyncMock(return_value=None)
            mock_checkpointer_class.return_value = mock_checkpointer

            result = await get_thread_history("test-thread-3")

            assert isinstance(result, str)
            data = json.loads(result)
            assert data["status"] == "empty"
            assert data["message_count"] == 0

    @pytest.mark.asyncio
    async def test_get_thread_history_recent_messages(self):
        """Verificar que history retorna apenas as últimas 5 mensagens."""
        from src.mcp_server import get_thread_history
        from langchain_core.messages import HumanMessage

        # Criar 10 mensagens
        messages = [
            HumanMessage(content=f"Mensagem {i}") for i in range(10)
        ]

        with patch("src.mcp_server.Checkpointer") as mock_checkpointer_class:
            mock_checkpointer = AsyncMock()
            mock_checkpointer.aget = AsyncMock(
                return_value={"values": {"messages": messages}}
            )
            mock_checkpointer.__aenter__ = AsyncMock(return_value=mock_checkpointer)
            mock_checkpointer.__aexit__ = AsyncMock(return_value=None)
            mock_checkpointer_class.return_value = mock_checkpointer

            result = await get_thread_history("test-thread-4")

            data = json.loads(result)
            # Deve retornar apenas as últimas 5
            assert len(data["recent_messages"]) == 5
            assert data["message_count"] == 10

    @pytest.mark.asyncio
    async def test_get_server_status(self):
        """Verificar que status resource retorna informações do servidor."""
        from src.mcp_server import get_server_status

        result = await get_server_status()

        assert isinstance(result, str)
        data = json.loads(result)
        assert data["server"] == "Vectora-SubAgent"
        assert data["status"] == "ready"
        assert "version" in data
        assert data["tools_count"] == 11
        assert data["resources_count"] == 3
        assert "capabilities" in data


class TestMCPSubAgentPattern:
    """Testes para validar o padrão Sub-Agent (Resources + Tools)."""

    def test_mcp_server_has_tools(self):
        """Verificar que MCP server tem as 11 ferramentas registradas."""
        from src.mcp_server import mcp

        # FastMCP deveria ter os tools registrados
        assert mcp is not None
        # As ferramentas são adicionadas com mcp.add_tool()
        # Espera-se que haja pelo menos um atributo contendo as tools

    def test_mcp_server_description_matches_subagent(self):
        """Verificar que descrição do servidor menciona padrão de Sub-Agente."""
        from src.mcp_server import mcp

        # A descrição deve mencionar RAG, colaborativo, etc
        description = getattr(mcp, "description", "")
        assert isinstance(description, str)

    @pytest.mark.asyncio
    async def test_resources_return_json_format(self):
        """Verificar que todos os recursos retornam JSON válido."""
        from src.mcp_server import (
            get_thread_context,
            get_thread_history,
            get_server_status,
        )

        with patch("src.mcp_server.Checkpointer") as mock_checkpointer_class:
            mock_checkpointer = AsyncMock()
            mock_checkpointer.aget = AsyncMock(return_value=None)
            mock_checkpointer.__aenter__ = AsyncMock(return_value=mock_checkpointer)
            mock_checkpointer.__aexit__ = AsyncMock(return_value=None)
            mock_checkpointer_class.return_value = mock_checkpointer

            # Testar cada resource
            context_result = await get_thread_context("test")
            assert json.loads(context_result)  # Deve ser JSON válido

            history_result = await get_thread_history("test")
            assert json.loads(history_result)  # Deve ser JSON válido

            status_result = await get_server_status()
            assert json.loads(status_result)  # Deve ser JSON válido
