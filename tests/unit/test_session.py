"""Testes para session.py (Session Management Service)."""

from __future__ import annotations

from datetime import UTC, datetime
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.services.session import SessionService


class TestSessionServiceInitialization:
    """Testes para inicializacao de SessionService."""

    def test_session_service_init(self) -> None:
        """Verificar que SessionService e inicializado corretamente."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        assert service.settings is settings
        assert service.checkpointer is None
        assert service._session_cache == {}


class TestSessionServiceCreate:
    """Testes para metodo create()."""

    @pytest.mark.asyncio
    async def test_create_new_session(self) -> None:
        """Verificar que nova sessao e criada com ID incremental."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        thread_id = await service.create(user_type="default")

        assert thread_id == 1
        assert thread_id in service._session_cache

    @pytest.mark.asyncio
    async def test_create_session_increments_id(self) -> None:
        """Verificar que IDs incrementam corretamente."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        id1 = await service.create()
        id2 = await service.create()
        id3 = await service.create()

        assert id1 == 1
        assert id2 == 2
        assert id3 == 3

    @pytest.mark.asyncio
    async def test_create_session_stores_metadata(self) -> None:
        """Verificar que metadados sao armazenados."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        thread_id = await service.create(user_type="premium")

        metadata = service._session_cache[thread_id]
        assert metadata["user_type"] == "premium"
        assert "created_at" in metadata
        assert "last_activity" in metadata
        assert metadata["message_count"] == 0


class TestSessionServiceSwitch:
    """Testes para metodo switch()."""

    @pytest.mark.asyncio
    async def test_switch_to_existing_session(self) -> None:
        """Verificar que pode fazer switch para sessao existente."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        thread_id = await service.create()
        result = await service.switch(thread_id)

        assert result is True

    @pytest.mark.asyncio
    async def test_switch_to_nonexistent_session(self) -> None:
        """Verificar que switch falha para sessao inexistente."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        result = await service.switch(999)

        assert result is False


class TestSessionServiceListAll:
    """Testes para metodo list_all()."""

    @pytest.mark.asyncio
    async def test_list_all_empty(self) -> None:
        """Verificar que list_all retorna lista vazia."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        sessions = await service.list_all()

        assert sessions == []

    @pytest.mark.asyncio
    async def test_list_all_with_sessions(self) -> None:
        """Verificar que list_all retorna todas as sessoes."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        await service.create()
        await service.create()
        await service.create()

        sessions = await service.list_all()

        assert len(sessions) == 3


class TestSessionServiceGetRunnableConfig:
    """Testes para metodo get_runnable_config()."""

    def test_get_runnable_config_returns_config(self) -> None:
        """Verificar que get_runnable_config retorna RunnableConfig."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        config = service.get_runnable_config(thread_id=1)

        assert config is not None

    def test_get_runnable_config_includes_thread_id(self) -> None:
        """Verificar que config inclui thread_id."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        config = service.get_runnable_config(thread_id=42)

        assert config["configurable"]["thread_id"] == 42


class TestSessionServiceDelete:
    """Testes para metodo delete()."""

    @pytest.mark.asyncio
    async def test_delete_existing_session(self) -> None:
        """Verificar que sessao existente pode ser deletada."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        thread_id = await service.create()
        result = await service.delete(thread_id)

        assert result is True
        assert thread_id not in service._session_cache

    @pytest.mark.asyncio
    async def test_delete_nonexistent_session(self) -> None:
        """Verificar que delete falha para sessao inexistente."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        result = await service.delete(999)

        assert result is False


class TestSessionServiceGetHistory:
    """Testes para metodo get_history()."""

    @pytest.mark.asyncio
    async def test_get_history_without_checkpointer(self) -> None:
        """Verificar que get_history retorna lista vazia sem checkpointer."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        history = await service.get_history(thread_id=1)

        assert history == []


class TestSessionServiceUpdateActivity:
    """Testes para metodo update_activity()."""

    @pytest.mark.asyncio
    async def test_update_activity_for_existing_session(self) -> None:
        """Verificar que update_activity atualiza timestamp."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        thread_id = await service.create()

        await service.update_activity(thread_id)

        assert service._session_cache[thread_id]["message_count"] > 0

    @pytest.mark.asyncio
    async def test_update_activity_increments_message_count(self) -> None:
        """Verificar que update_activity incrementa message_count."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        thread_id = await service.create()
        initial_count = service._session_cache[thread_id]["message_count"]

        await service.update_activity(thread_id)

        new_count = service._session_cache[thread_id]["message_count"]
        assert new_count == initial_count + 1


class TestSessionServiceShutdown:
    """Testes para metodo shutdown()."""

    @pytest.mark.asyncio
    async def test_shutdown_without_context(self) -> None:
        """Verificar que shutdown funciona sem contexto ativo."""
        settings = MagicMock()
        service = SessionService(settings=settings)

        await service.shutdown()

        assert service.checkpointer is None
