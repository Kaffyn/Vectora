"""Testes para AgentManager - o orquestrador central do Vectora."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.agent import AgentManager
from vectora.config.settings import Settings


@pytest.fixture
def mock_settings() -> MagicMock:
    """Criar mock de settings."""
    settings = MagicMock(spec=Settings)
    settings.get_llm_provider.return_value = "google-genai"
    settings.get_llm_model.return_value = "gemini-2.5-flash"
    return settings


@pytest.fixture
def agent_manager(mock_settings: MagicMock) -> AgentManager:
    """Criar AgentManager com mocks."""
    with (
        patch("vectora.services.telemetry.TelemetryService"),
        patch("vectora.services.session.SessionService"),
        patch("vectora.services.embedding.EmbeddingService"),
    ):
        return AgentManager(settings=mock_settings)


class TestAgentManagerInitialization:
    """Testes de inicialização do AgentManager."""

    def test_agent_manager_initializes_with_settings(self, mock_settings: MagicMock):
        """Verificar que AgentManager inicializa com settings."""
        with (
            patch("vectora.services.telemetry.TelemetryService"),
            patch("vectora.services.session.SessionService"),
            patch("vectora.services.embedding.EmbeddingService"),
        ):
            manager = AgentManager(settings=mock_settings)
            assert manager.settings == mock_settings
            assert manager.graph is None

    def test_agent_manager_loads_default_settings(self):
        """Verificar que AgentManager carrega settings padrão."""
        with (
            patch("vectora.services.telemetry.TelemetryService"),
            patch("vectora.services.session.SessionService"),
            patch("vectora.services.embedding.EmbeddingService"),
            patch("vectora.agent.AgentManager._load_settings") as mock_load,
        ):
            mock_settings = MagicMock(spec=Settings)
            mock_load.return_value = mock_settings
            manager = AgentManager(settings=None)
            assert manager.settings == mock_settings


class TestAgentManagerChat:
    """Testes de operações de chat."""

    @pytest.mark.asyncio
    async def test_chat_without_graph_raises_error(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que chat() lança erro se graph não inicializado."""
        with pytest.raises(RuntimeError, match="Graph not initialized"):
            await agent_manager.chat("Hello")

    @pytest.mark.asyncio
    async def test_chat_returns_placeholder_response(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que chat() lança erro se graph está None."""
        # graph is None by default in AgentManager.__init__
        with pytest.raises(RuntimeError, match="Graph not initialized"):
            await agent_manager.chat("Hello, Vectora!")


class TestAgentManagerModelOperations:
    """Testes de operações de modelo."""

    @pytest.mark.asyncio
    async def test_switch_model_success(
        self, agent_manager: AgentManager, mock_settings: MagicMock
    ) -> None:
        """Verificar que switch_model() atualiza settings."""
        result = await agent_manager.switch_model("google-genai", "gemini-3-flash")
        assert result is True
        mock_settings.set_model.assert_called_once_with(
            "google-genai", "gemini-3-flash"
        )

    @pytest.mark.asyncio
    async def test_switch_model_handles_error(
        self, agent_manager: AgentManager, mock_settings: MagicMock
    ) -> None:
        """Verificar que switch_model() retorna False em erro."""
        mock_settings.set_model.side_effect = ValueError("Invalid model")
        result = await agent_manager.switch_model("google-genai", "invalid")
        assert result is False

    def test_get_available_models_all_providers(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que get_available_models() retorna todos os providers."""
        models = agent_manager.get_available_models()
        assert "google-genai" in models
        assert "openai" in models
        assert "anthropic" in models
        assert len(models["google-genai"]) > 0

    def test_get_available_models_specific_provider(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que get_available_models() filtra por provider."""
        models = agent_manager.get_available_models(provider="google-genai")
        assert "google-genai" in models
        assert len(models) == 1
        assert len(models["google-genai"]) > 0

    def test_get_available_models_invalid_provider(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que get_available_models() retorna provider com lista vazia para provider inválido."""
        models = agent_manager.get_available_models(provider="nonexistent")
        assert "nonexistent" in models
        assert models["nonexistent"] == []


class TestAgentManagerSessionOperations:
    """Testes de operações de sessão."""

    @pytest.mark.asyncio
    async def test_create_session_default_user_type(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que create_session() usa user_type padrão."""
        with patch.object(
            agent_manager.session_service, "create", new_callable=AsyncMock
        ) as mock_create:
            mock_create.return_value = 42
            session_id = await agent_manager.create_session()
            assert session_id == 42
            mock_create.assert_called_once_with("default")

    @pytest.mark.asyncio
    async def test_create_session_custom_user_type(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que create_session() usa user_type customizado."""
        with patch.object(
            agent_manager.session_service, "create", new_callable=AsyncMock
        ) as mock_create:
            mock_create.return_value = 123
            session_id = await agent_manager.create_session(user_type="premium")
            assert session_id == 123
            mock_create.assert_called_once_with("premium")

    @pytest.mark.asyncio
    async def test_switch_session_success(self, agent_manager: AgentManager) -> None:
        """Verificar que switch_session() funciona."""
        with patch.object(
            agent_manager.session_service, "switch", new_callable=AsyncMock
        ) as mock_switch:
            mock_switch.return_value = True
            result = await agent_manager.switch_session(42)
            assert result is True
            mock_switch.assert_called_once_with(42)

    @pytest.mark.asyncio
    async def test_list_sessions(self, agent_manager: AgentManager) -> None:
        """Verificar que list_sessions() retorna lista de sessões."""
        expected_sessions = [
            {"id": 1, "user_type": "default", "created_at": "2026-05-17"},
            {"id": 2, "user_type": "premium", "created_at": "2026-05-17"},
        ]
        with patch.object(
            agent_manager.session_service, "list_all", new_callable=AsyncMock
        ) as mock_list:
            mock_list.return_value = expected_sessions
            sessions = await agent_manager.list_sessions()
            assert sessions == expected_sessions
            assert len(sessions) == 2


class TestAgentManagerVectorOperations:
    """Testes de operações de vetores."""

    @pytest.mark.asyncio
    async def test_search_vectors_basic(self, agent_manager: AgentManager) -> None:
        """Verificar que search_vectors() executa busca."""
        expected_results = [
            {"score": 0.95, "content": "Match 1"},
            {"score": 0.87, "content": "Match 2"},
        ]
        with patch.object(
            agent_manager.embedding_service, "search", new_callable=AsyncMock
        ) as mock_search:
            mock_search.return_value = expected_results
            results = await agent_manager.search_vectors("test query")
            assert len(results) == 2
            assert results[0]["score"] == 0.95

    @pytest.mark.asyncio
    async def test_queue_document_for_embedding(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que queue_document_for_embedding() funciona."""
        with patch.object(
            agent_manager.embedding_service,
            "queue_document",
            new_callable=AsyncMock,
        ) as mock_queue:
            mock_queue.return_value = None
            result = await agent_manager.queue_document_for_embedding(
                "doc-123", "Test document", "articles"
            )
            assert result is True


class TestAgentManagerSecurityOperations:
    """Testes de operações de segurança (desabilitados - SecurityService não implementado)."""

    def test_security_operations_not_available(
        self, agent_manager: AgentManager
    ) -> None:
        """Verificar que operações de segurança não estão disponíveis (SecurityService não existe)."""
        # SecurityService não é inicializada no AgentManager
        # Esses testes são placeholders para quando a SecurityService for implementada


class TestAgentManagerAuditAndDebug:
    """Testes de auditoria e debug."""

    @pytest.mark.asyncio
    async def test_export_session_audit(self, agent_manager: AgentManager) -> None:
        """Verificar que export_session_audit() funciona."""
        expected_audit = "Session audit data..."
        with patch.object(
            agent_manager.telemetry_service,
            "export_session_audit",
            new_callable=AsyncMock,
        ) as mock_export:
            mock_export.return_value = expected_audit
            result = await agent_manager.export_session_audit(1)
            assert result == expected_audit
            mock_export.assert_called_once_with(1)

    @pytest.mark.asyncio
    async def test_get_debug_dump(self, agent_manager: AgentManager) -> None:
        """Verificar que get_debug_dump() funciona."""
        expected_dump = "Debug info..."
        with patch.object(
            agent_manager.telemetry_service,
            "export_debug_dump",
            new_callable=AsyncMock,
        ) as mock_dump:
            mock_dump.return_value = expected_dump
            result = await agent_manager.get_debug_dump()
            assert result == expected_dump
            mock_dump.assert_called_once()


class TestAgentManagerLifecycle:
    """Testes de ciclo de vida."""

    @pytest.mark.asyncio
    async def test_initialize(self, agent_manager: AgentManager) -> None:
        """Verificar que initialize() funciona."""
        with patch.object(
            agent_manager.session_service, "initialize", new_callable=AsyncMock
        ) as mock_sess_init:
            with patch.object(
                agent_manager.embedding_service, "start", new_callable=AsyncMock
            ) as mock_emb_start:
                mock_sess_init.return_value = None
                mock_emb_start.return_value = None
                # Should not raise
                await agent_manager.initialize()

    @pytest.mark.asyncio
    async def test_shutdown(self, agent_manager: AgentManager) -> None:
        """Verificar que shutdown() funciona."""
        with patch.object(
            agent_manager.session_service, "shutdown", new_callable=AsyncMock
        ):
            with patch.object(
                agent_manager.embedding_service, "stop", new_callable=AsyncMock
            ):
                # Should not raise
                await agent_manager.shutdown()
