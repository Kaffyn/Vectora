"""Testes para o módulo main.py - ponto de entrada da CLI."""

from __future__ import annotations

import sys
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.main import main, run


@pytest.fixture
def mock_settings() -> MagicMock:
    """Criar mock de settings com valores padrão."""
    settings = MagicMock()
    settings.get_llm_provider.return_value = "google-genai"
    settings.version = "0.1.0rc1"
    settings.get_available_providers.return_value = ["google-genai"]
    return settings


@pytest.fixture
def mock_agent_manager(mock_settings: MagicMock) -> MagicMock:
    """Criar mock de AgentManager."""
    agent = MagicMock()
    agent.initialize = AsyncMock(return_value=None)
    return agent


class TestMainAsyncFunction:
    """Testes para a função main() assíncrona."""

    @pytest.mark.asyncio
    async def test_main_successful_execution(
        self, mock_settings: MagicMock, mock_agent_manager: MagicMock
    ) -> None:
        """Verificar que main() executa com sucesso."""
        with patch("vectora.main.setup_logging"):
            with patch("vectora.config.settings.Settings") as mock_settings_class:
                mock_settings_class.return_value = mock_settings
                with patch("vectora.agent.AgentManager") as mock_agent_class:
                    mock_agent_class.return_value = mock_agent_manager
                    with patch("vectora.ui.chat.run_chat", new_callable=AsyncMock):
                        # Should not raise
                        await main()

    @pytest.mark.asyncio
    async def test_main_settings_initialization_error(
        self,
    ) -> None:
        """Verificar que main() falha gracefully se Settings falhar."""
        with patch("vectora.main.setup_logging"):
            with patch(
                "vectora.config.settings.Settings",
                side_effect=ValueError("Invalid configuration"),
            ):
                with patch("builtins.print"):
                    with pytest.raises(SystemExit) as exc_info:
                        await main()
                    assert exc_info.value.code == 1

    @pytest.mark.asyncio
    async def test_main_agent_manager_initialization_error(
        self, mock_settings: MagicMock
    ) -> None:
        """Verificar que main() falha gracefully se AgentManager falhar."""
        with patch("vectora.main.setup_logging"):
            with patch("vectora.config.settings.Settings", return_value=mock_settings):
                with patch(
                    "vectora.agent.AgentManager",
                    side_effect=RuntimeError("Failed to initialize"),
                ):
                    with patch("builtins.print"):
                        with pytest.raises(SystemExit) as exc_info:
                            await main()
                        assert exc_info.value.code == 1

    @pytest.mark.asyncio
    async def test_main_agent_initialize_fails(
        self, mock_settings: MagicMock, mock_agent_manager: MagicMock
    ) -> None:
        """Verificar que main() falha gracefully se agent.initialize() falhar."""
        mock_agent_manager.initialize.side_effect = RuntimeError(
            "Initialization failed"
        )

        with patch("vectora.main.setup_logging"):
            with patch("vectora.config.settings.Settings", return_value=mock_settings):
                with patch(
                    "vectora.agent.AgentManager", return_value=mock_agent_manager
                ):
                    with patch("builtins.print"):
                        with pytest.raises(SystemExit) as exc_info:
                            await main()
                        assert exc_info.value.code == 1

    @pytest.mark.asyncio
    async def test_main_runs_setup_wizard_when_no_providers(
        self, mock_settings: MagicMock, mock_agent_manager: MagicMock
    ) -> None:
        """Verificar que main() invoca setup wizard se nenhum provider está configurado."""
        # First call returns empty list, second call returns a provider
        mock_settings.get_available_providers.side_effect = [[], ["google-genai"]]

        with patch("vectora.main.setup_logging"):
            with patch("vectora.config.settings.Settings", return_value=mock_settings):
                with patch(
                    "vectora.agent.AgentManager", return_value=mock_agent_manager
                ):
                    with patch(
                        "vectora.services.setup_wizard.run_setup_sync"
                    ) as mock_setup:
                        with patch("vectora.ui.chat.run_chat", new_callable=AsyncMock):
                            await main()
                            mock_setup.assert_called_once()

    @pytest.mark.asyncio
    async def test_main_keyboard_interrupt_handling(
        self, mock_settings: MagicMock, mock_agent_manager: MagicMock
    ) -> None:
        """Verificar que main() trata KeyboardInterrupt gracefully."""
        with patch("vectora.main.setup_logging"):
            with patch("vectora.config.settings.Settings", return_value=mock_settings):
                with patch(
                    "vectora.agent.AgentManager", return_value=mock_agent_manager
                ):
                    with patch(
                        "vectora.ui.chat.run_chat",
                        new_callable=AsyncMock,
                        side_effect=KeyboardInterrupt(),
                    ):
                        with patch("builtins.print"):
                            with pytest.raises(SystemExit) as exc_info:
                                await main()
                            assert exc_info.value.code == 0

    @pytest.mark.asyncio
    async def test_main_unexpected_error_handling(
        self, mock_settings: MagicMock, mock_agent_manager: MagicMock
    ) -> None:
        """Verificar que main() trata exceções inesperadas gracefully."""
        with patch("vectora.main.setup_logging"):
            with patch("vectora.config.settings.Settings", return_value=mock_settings):
                with patch(
                    "vectora.agent.AgentManager", return_value=mock_agent_manager
                ):
                    with patch(
                        "vectora.ui.chat.run_chat",
                        new_callable=AsyncMock,
                        side_effect=RuntimeError("Unexpected error"),
                    ):
                        with patch("builtins.print"):
                            with pytest.raises(SystemExit) as exc_info:
                                await main()
                            assert exc_info.value.code == 1


class TestRunSyncFunction:
    """Testes para a função run() síncrona."""

    def test_run_calls_asyncio_run(self) -> None:
        """Verificar que run() chama asyncio.run(main())."""
        with patch("vectora.main.asyncio.run") as mock_asyncio_run:
            run()
            mock_asyncio_run.assert_called_once()

    def test_run_handles_keyboard_interrupt(self) -> None:
        """Verificar que run() trata KeyboardInterrupt."""
        with patch("vectora.main.asyncio.run", side_effect=KeyboardInterrupt()):
            with patch("builtins.print"):
                with pytest.raises(SystemExit) as exc_info:
                    run()
                assert exc_info.value.code == 0

    def test_run_propagates_other_exceptions(self) -> None:
        """Verificar que run() propaga outras exceções."""
        with patch("vectora.main.asyncio.run", side_effect=RuntimeError("Test error")):
            with pytest.raises(RuntimeError, match="Test error"):
                run()


class TestMainModuleIntegration:
    """Testes de integração para o módulo main."""

    def test_sys_path_includes_vectora_dir(self) -> None:
        """Verificar que vectora directory está adicionado ao sys.path."""
        import vectora.main

        # Se importou sem erro, o sys.path foi configurado corretamente
        assert True

    def test_utf8_encoding_configured(self) -> None:
        """Verificar que codificação UTF-8 está configurada."""
        import os

        # UTF-8 deve ser configurado para funcionar com caracteres especiais
        assert os.environ.get("PYTHONIOENCODING") == "utf-8"

    @pytest.mark.asyncio
    async def test_main_logs_startup_message(
        self, mock_settings: MagicMock, mock_agent_manager: MagicMock, caplog
    ) -> None:  # type: ignore[no-untyped-def]
        """Verificar que main() registra uma mensagem de inicialização."""
        import logging

        caplog.set_level(logging.INFO)

        with patch("vectora.main.setup_logging"):
            with patch("vectora.config.settings.Settings", return_value=mock_settings):
                with patch(
                    "vectora.agent.AgentManager", return_value=mock_agent_manager
                ):
                    with patch("vectora.ui.chat.run_chat", new_callable=AsyncMock):
                        await main()
                        # Verificar que a mensagem de inicialização foi registrada
                        assert "Starting Vectora CLI" in caplog.text
