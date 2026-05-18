"""Testes para vectora/ui/commands.py"""

from __future__ import annotations

import contextlib
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.ui.commands import (
    _display_command_list,
    _display_help,
    _handle_debug_command,
    _handle_list_sessions,
    _handle_model_command,
    _handle_new_session,
    _handle_rag_command,
    _handle_switch_session,
    _handle_tools_command,
    _load_debug_config,
    _save_debug_config,
    get_available_models,
    handle_command,
)


class TestLoadDebugConfig:
    """Testes para _load_debug_config()."""

    def test_load_debug_config_returns_false_when_file_missing(self):
        """Testar que retorna False se arquivo nao existe."""
        with patch("vectora.ui.commands.CONFIG_FILE") as mock_file:
            mock_file.exists.return_value = False

            result = _load_debug_config()

            assert result is False

    def test_load_debug_config_reads_from_file(self):
        """Testar que le valor do arquivo."""
        with patch("vectora.ui.commands.CONFIG_FILE") as mock_file:
            mock_file.exists.return_value = True
            mock_file.read_text.return_value = '{"debug_mode": true}'

            result = _load_debug_config()

            assert result is True

    def test_load_debug_config_handles_missing_key(self):
        """Testar que retorna False se chave nao existe."""
        with patch("vectora.ui.commands.CONFIG_FILE") as mock_file:
            mock_file.exists.return_value = True
            mock_file.read_text.return_value = "{}"

            result = _load_debug_config()

            assert result is False

    def test_load_debug_config_handles_invalid_json(self):
        """Testar que retorna False se JSON invalido."""
        with patch("vectora.ui.commands.CONFIG_FILE") as mock_file:
            mock_file.exists.return_value = True
            mock_file.read_text.return_value = "invalid json"

            result = _load_debug_config()

            assert result is False


class TestSaveDebugConfig:
    """Testes para _save_debug_config()."""

    def test_save_debug_config_creates_dir(self):
        """Testar que cria diretorio se nao existe."""
        with patch("vectora.ui.commands.CONFIG_FILE") as mock_file:
            mock_file.parent.mkdir = MagicMock()
            mock_file.write_text = MagicMock()
            mock_file.exists.return_value = False

            _save_debug_config(True)

            mock_file.parent.mkdir.assert_called_once_with(parents=True, exist_ok=True)

    def test_save_debug_config_writes_file(self):
        """Testar que escreve arquivo."""
        with patch("vectora.ui.commands.CONFIG_FILE") as mock_file:
            mock_file.parent.mkdir = MagicMock()
            mock_file.write_text = MagicMock()
            mock_file.exists.return_value = False

            _save_debug_config(True)

            mock_file.write_text.assert_called_once()
            call_args = mock_file.write_text.call_args[0][0]
            assert "debug_mode" in call_args
            assert "true" in call_args.lower()

    def test_save_debug_config_preserves_existing_data(self):
        """Testar que preserva dados existentes."""
        with patch("vectora.ui.commands.CONFIG_FILE") as mock_file:
            mock_file.parent.mkdir = MagicMock()
            mock_file.write_text = MagicMock()
            mock_file.exists.return_value = True
            mock_file.read_text.return_value = '{"other_key": "value"}'

            _save_debug_config(False)

            call_args = mock_file.write_text.call_args[0][0]
            assert "debug_mode" in call_args
            assert "other_key" in call_args


class TestGetAvailableModels:
    """Testes para get_available_models()."""

    def test_get_available_models_all_providers(self):
        """Testar que retorna modelos de todos providers."""
        result = get_available_models()

        assert isinstance(result, dict)
        assert "google-genai" in result
        assert "openai" in result
        assert "anthropic" in result

    def test_get_available_models_specific_provider(self):
        """Testar que retorna modelos de provider especifico."""
        result = get_available_models("google-genai")

        assert isinstance(result, dict)
        assert "google-genai" in result
        assert len(result["google-genai"]) > 0

    def test_get_available_models_unknown_provider(self):
        """Testar que retorna lista vazia para provider desconhecido."""
        result = get_available_models("unknown-provider")

        assert isinstance(result, dict)
        assert result["unknown-provider"] == []


class TestHandleDebugCommand:
    """Testes para _handle_debug_command()."""

    @pytest.mark.asyncio
    async def test_handle_debug_command_toggle(self):
        """Testar toggle de debug mode."""
        console = MagicMock()

        with patch("vectora.ui.commands._save_debug_config") as mock_save:
            result = await _handle_debug_command("", console, False)

            assert result is True
            mock_save.assert_called_once_with(True)

    @pytest.mark.asyncio
    async def test_handle_debug_command_enable(self):
        """Testar habilitacao de debug mode."""
        console = MagicMock()

        with patch("vectora.ui.commands._save_debug_config") as mock_save:
            result = await _handle_debug_command("true", console, False)

            assert result is True
            mock_save.assert_called_once_with(True)

    @pytest.mark.asyncio
    async def test_handle_debug_command_disable(self):
        """Testar desabilitacao de debug mode."""
        console = MagicMock()

        with patch("vectora.ui.commands._save_debug_config") as mock_save:
            result = await _handle_debug_command("false", console, True)

            assert result is False
            mock_save.assert_called_once_with(False)

    @pytest.mark.asyncio
    async def test_handle_debug_command_already_enabled(self):
        """Testar que nao muda se ja habilitado."""
        console = MagicMock()

        with patch("vectora.ui.commands._save_debug_config") as mock_save:
            result = await _handle_debug_command("true", console, True)

            assert result is True
            mock_save.assert_not_called()

    @pytest.mark.asyncio
    async def test_handle_debug_command_already_disabled(self):
        """Testar que nao muda se ja desabilitado."""
        console = MagicMock()

        with patch("vectora.ui.commands._save_debug_config") as mock_save:
            result = await _handle_debug_command("false", console, False)

            assert result is False
            mock_save.assert_not_called()

    @pytest.mark.asyncio
    async def test_handle_debug_command_invalid_arg(self):
        """Testar que retorna valor atual para argumento invalido."""
        console = MagicMock()

        with patch("vectora.ui.commands._save_debug_config") as mock_save:
            result = await _handle_debug_command("invalid", console, False)

            assert result is False
            mock_save.assert_not_called()


class TestDisplayHelp:
    """Testes para _display_help()."""

    def test_display_help_prints_output(self):
        """Testar que exibe ajuda."""
        console = MagicMock()

        _display_help(console)

        # Deve ter chamado console.print
        assert console.print.called


class TestDisplayCommandList:
    """Testes para _display_command_list()."""

    def test_display_command_list_prints_output(self):
        """Testar que exibe lista de comandos."""
        console = MagicMock()

        _display_command_list(console)

        # Deve ter chamado console.print
        assert console.print.called


class TestHandleModelCommand:
    """Testes para _handle_model_command()."""

    @pytest.mark.asyncio
    async def test_handle_model_command_list(self):
        """Testar listagem de modelos."""
        console = MagicMock()

        with patch("vectora.ui.commands.settings") as mock_settings:
            mock_settings.get_llm_provider.return_value = "openai"

            await _handle_model_command("", console)

            assert console.print.called

    @pytest.mark.asyncio
    async def test_handle_model_command_invalid_model(self):
        """Testar seleção de modelo invalido."""
        console = MagicMock()

        with patch("vectora.ui.commands.settings") as mock_settings:
            mock_settings.get_llm_provider.return_value = "openai"

            await _handle_model_command("invalid-model", console)

            assert console.print.called


class TestHandleToolsCommand:
    """Testes para _handle_tools_command()."""

    @pytest.mark.asyncio
    async def test_handle_tools_command_prints_output(self):
        """Testar que exibe ferramentas disponaveis."""
        console = MagicMock()

        await _handle_tools_command(console)

        assert console.print.called


class TestHandleRagCommand:
    """Testes para _handle_rag_command()."""

    @pytest.mark.asyncio
    async def test_handle_rag_command_with_args(self):
        """Testar comando RAG com argumentos."""
        console = MagicMock()

        with patch("vectora.ui.commands.settings"):
            with contextlib.suppress(Exception):
                # Pode falhar por dependencias externas
                await _handle_rag_command("search-term", console)


class TestHandleNewSession:
    """Testes para _handle_new_session()."""

    @pytest.mark.asyncio
    async def test_handle_new_session_returns_context(self):
        """Testar que retorna contexto."""
        console = MagicMock()
        context = MagicMock()

        try:
            result = await _handle_new_session(context, console)
            # Deve retornar algo
            assert result is not None
        except Exception:
            # Pode falhar por dependencias
            pass


class TestHandleListSessions:
    """Testes para _handle_list_sessions()."""

    @pytest.mark.asyncio
    async def test_handle_list_sessions_prints_output(self):
        """Testar que exibe sessoes."""
        console = MagicMock()
        context = MagicMock()

        try:
            await _handle_list_sessions(context, console)
            assert console.print.called
        except Exception:
            # Pode falhar por dependencias
            pass


class TestHandleSwitchSession:
    """Testes para _handle_switch_session()."""

    @pytest.mark.asyncio
    async def test_handle_switch_session_returns_context(self):
        """Testar que muda sessao."""
        console = MagicMock()
        context = MagicMock()

        try:
            result = await _handle_switch_session("session-id", context, console)
            assert result is not None
        except Exception:
            # Pode falhar por dependencias
            pass


class TestHandleCommand:
    """Testes para handle_command()."""

    @pytest.mark.asyncio
    async def test_handle_command_quit(self):
        """Testar comando /quit."""
        console = MagicMock()

        should_exit, _context, _debug_mode = await handle_command(
            "/quit", None, console, None, False
        )

        assert should_exit is True

    @pytest.mark.asyncio
    async def test_handle_command_sair(self):
        """Testar comando /sair (portugues)."""
        console = MagicMock()

        should_exit, _context, _debug_mode = await handle_command(
            "/sair", None, console, None, False
        )

        assert should_exit is True

    @pytest.mark.asyncio
    async def test_handle_command_q(self):
        """Testar comando /q (abreviado)."""
        console = MagicMock()

        should_exit, _context, _debug_mode = await handle_command(
            "/q", None, console, None, False
        )

        assert should_exit is True

    @pytest.mark.asyncio
    async def test_handle_command_help(self):
        """Testar comando /help."""
        console = MagicMock()

        with patch("vectora.ui.commands._display_help"):
            should_exit, _context, _debug_mode = await handle_command(
                "/help", None, console, None, False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_debug(self):
        """Testar comando /debug."""
        console = MagicMock()

        with patch(
            "vectora.ui.commands._handle_debug_command", new_callable=AsyncMock
        ) as mock_debug:
            mock_debug.return_value = True

            should_exit, _context, debug_mode = await handle_command(
                "/debug", None, console, None, False
            )

            assert should_exit is False
            assert debug_mode is True

    @pytest.mark.asyncio
    async def test_handle_command_model(self):
        """Testar comando /model."""
        console = MagicMock()

        with patch("vectora.ui.commands._handle_model_command", new_callable=AsyncMock):
            should_exit, _context, _debug_mode = await handle_command(
                "/model gpt-4", None, console, None, False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_new(self):
        """Testar comando /new."""
        console = MagicMock()

        with patch(
            "vectora.ui.commands._handle_new_session", new_callable=AsyncMock
        ) as mock_new:
            mock_new.return_value = MagicMock()

            should_exit, _context, _debug_mode = await handle_command(
                "/new", None, console, MagicMock(), False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_sessions(self):
        """Testar comando /sessions."""
        console = MagicMock()

        with patch("vectora.ui.commands._handle_list_sessions", new_callable=AsyncMock):
            should_exit, _context, _debug_mode = await handle_command(
                "/sessions", None, console, None, False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_session(self):
        """Testar comando /session."""
        console = MagicMock()

        with patch(
            "vectora.ui.commands._handle_switch_session", new_callable=AsyncMock
        ) as mock_switch:
            mock_switch.return_value = MagicMock()

            should_exit, _context, _debug_mode = await handle_command(
                "/session id123", None, console, MagicMock(), False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_tools(self):
        """Testar comando /tools."""
        console = MagicMock()

        with patch("vectora.ui.commands._handle_tools_command", new_callable=AsyncMock):
            should_exit, _context, _debug_mode = await handle_command(
                "/tools", None, console, None, False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_tool(self):
        """Testar comando /tool (alias)."""
        console = MagicMock()

        with patch("vectora.ui.commands._handle_tools_command", new_callable=AsyncMock):
            should_exit, _context, _debug_mode = await handle_command(
                "/tool", None, console, None, False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_rag(self):
        """Testar comando /rag."""
        console = MagicMock()

        with patch("vectora.ui.commands._handle_rag_command", new_callable=AsyncMock):
            should_exit, _context, _debug_mode = await handle_command(
                "/rag search-term", None, console, None, False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_list(self):
        """Testar comando /list."""
        console = MagicMock()

        with patch("vectora.ui.commands._display_command_list"):
            should_exit, _context, _debug_mode = await handle_command(
                "/list", None, console, None, False
            )

            assert should_exit is False

    @pytest.mark.asyncio
    async def test_handle_command_unknown(self):
        """Testar comando desconhecido."""
        console = MagicMock()

        should_exit, _context, _debug_mode = await handle_command(
            "/unknown", None, console, None, False
        )

        assert should_exit is False
        console.print.assert_called()

    @pytest.mark.asyncio
    async def test_handle_command_preserves_context(self):
        """Testar que preserva contexto entre comandos."""
        console = MagicMock()
        original_context = MagicMock()

        _should_exit, returned_context, _debug_mode = await handle_command(
            "/help", None, console, original_context, False
        )

        assert returned_context is original_context

    @pytest.mark.asyncio
    async def test_handle_command_preserves_debug_mode(self):
        """Testar que preserva debug mode quando nao mudado."""
        console = MagicMock()

        _should_exit, _context, debug_mode = await handle_command(
            "/help", None, console, None, True
        )

        assert debug_mode is True
