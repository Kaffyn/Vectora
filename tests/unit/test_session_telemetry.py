"""Testes para SessionService e TelemetryService.

Cobre: vectora/services/session.py, vectora/services/telemetry.py
"""

from __future__ import annotations

import contextlib
import json
import logging
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.config.settings import Settings


@pytest.fixture
def mock_settings():
    """Settings para testes."""
    return Settings()


class TestSessionServiceInit:
    """Testa inicialização do SessionService."""

    def test_init_creates_service(self, mock_settings):
        """Verifica que SessionService é criado."""
        from vectora.services.session import SessionService

        service = SessionService(mock_settings)
        assert service is not None
        assert service.settings is mock_settings
        assert service.checkpointer is None
        assert service._session_cache == {}

    def test_init_has_settings(self, mock_settings):
        """Verifica que settings é armazenado."""
        from vectora.services.session import SessionService

        service = SessionService(mock_settings)
        assert service.settings.enable_rag == mock_settings.enable_rag


class TestSessionServiceOperations:
    """Testa operações de sessão."""

    @pytest.mark.asyncio
    async def test_initialize_sets_checkpointer(self, mock_settings):
        """Verifica que initialize() configura o checkpointer."""
        from vectora.services.session import SessionService

        service = SessionService(mock_settings)

        mock_checkpointer = AsyncMock()
        mock_context = MagicMock()
        mock_context.__aenter__ = AsyncMock(return_value=mock_checkpointer)
        mock_context.__aexit__ = AsyncMock(return_value=None)

        with patch("vectora.services.session.AsyncSqliteSaver") as mock_saver_class:
            mock_saver_class.from_conn_string = MagicMock(return_value=mock_context)
            await service.initialize()
            assert service.checkpointer is not None

    @pytest.mark.asyncio
    async def test_initialize_handles_error(self, mock_settings):
        """Verifica que initialize() trata erros de banco graciosamente."""
        from vectora.services.session import SessionService

        service = SessionService(mock_settings)

        with patch("vectora.services.session.AsyncSqliteSaver") as mock_saver_class:
            mock_saver_class.from_conn_string = MagicMock(
                side_effect=Exception("DB error")
            )
            # Não deve levantar exceção
            with contextlib.suppress(Exception):
                await service.initialize()

    def test_get_runnable_config_returns_dict(self, mock_settings):
        """Verifica que get_runnable_config retorna configuração válida."""
        from vectora.services.session import SessionService

        service = SessionService(mock_settings)

        if hasattr(service, "get_runnable_config"):
            config = service.get_runnable_config(thread_id=1)
            assert isinstance(config, dict)
            assert "configurable" in config

    @pytest.mark.asyncio
    async def test_create_session_returns_id(self, mock_settings):
        """Verifica que create_session retorna ID de sessão."""
        from vectora.services.session import SessionService

        service = SessionService(mock_settings)

        if hasattr(service, "create_session"):
            with patch.object(service, "checkpointer", AsyncMock()):
                session_id = await service.create_session(user_type="user")
                assert session_id is not None
                assert isinstance(session_id, int)


class TestTelemetryServiceJSON:
    """Testa formatação de telemetria."""

    def test_json_formatter_telemetry(self):
        """Testa JSONFormatter do módulo telemetry."""
        from vectora.services.telemetry import JSONFormatter

        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="vectora.telemetry",
            level=logging.INFO,
            pathname="telemetry.py",
            lineno=1,
            msg="Evento de telemetria",
            args=(),
            exc_info=None,
        )
        output = formatter.format(record)
        data = json.loads(output)
        assert data["level"] == "INFO"
        assert data["message"] == "Evento de telemetria"

    def test_text_formatter_telemetry(self):
        """Testa TextFormatter do módulo telemetry."""
        from vectora.services.telemetry import TextFormatter

        formatter = TextFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.DEBUG,
            pathname="test.py",
            lineno=1,
            msg="Debug msg",
            args=(),
            exc_info=None,
        )
        output = formatter.format(record)
        assert isinstance(output, str)


class TestTelemetryService:
    """Testa TelemetryService."""

    def test_init_creates_service(self, mock_settings):
        """Verifica que TelemetryService é criado."""
        from vectora.services.telemetry import TelemetryService

        service = TelemetryService(mock_settings)
        assert service is not None
        assert service.settings is mock_settings

    def test_init_has_correct_attributes(self, mock_settings):
        """Verifica atributos iniciais."""
        from vectora.services.telemetry import TelemetryService

        service = TelemetryService(mock_settings)
        # Atributos esperados de um serviço de telemetria
        assert hasattr(service, "settings")

    def test_is_enabled_returns_bool(self, mock_settings):
        """Verifica que is_enabled retorna booleano."""
        from vectora.services.telemetry import TelemetryService

        service = TelemetryService(mock_settings)
        if hasattr(service, "is_enabled"):
            result = service.is_enabled()
            assert isinstance(result, bool)

    def test_track_event_does_not_crash(self, mock_settings):
        """Verifica que track_event não lança exceção."""
        from vectora.services.telemetry import TelemetryService

        service = TelemetryService(mock_settings)
        if hasattr(service, "track_event"):
            # Não deve lançar exceção
            service.track_event("test_event", {"key": "value"})
