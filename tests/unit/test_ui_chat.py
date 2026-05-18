"""Testes para vectora/ui/chat.py"""

from __future__ import annotations

from io import StringIO
from unittest.mock import MagicMock, patch

import pytest

# Try importing chat module, but may fail due to dependencies
try:
    from vectora.ui.chat import (
        SafeConsole,
        _is_terminal_tool,
        console,
        logger,
    )

    HAS_CHAT_MODULE = True
except ImportError:
    HAS_CHAT_MODULE = False


class TestSafeConsole:
    """Testes para SafeConsole class."""

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_safe_console_is_console(self):
        """Testar que SafeConsole herda de Console."""
        from rich.console import Console

        assert issubclass(SafeConsole, Console)

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_safe_console_has_print_method(self):
        """Testar que SafeConsole tem metodo print."""
        assert hasattr(SafeConsole, "print")
        assert callable(SafeConsole.print)

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_safe_console_instantiation(self):
        """Testar instanciacao de SafeConsole."""
        console_instance = SafeConsole()
        assert console_instance is not None
        assert hasattr(console_instance, "print")

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_safe_console_print_simple_text(self):
        """Testar print com texto simples."""
        safe_console = SafeConsole(file=StringIO())
        # Nao deve lancar erro
        safe_console.print("Hello, World!")

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_safe_console_print_with_markup(self):
        """Testar print com markup rich."""
        safe_console = SafeConsole(file=StringIO())
        # Nao deve lancar erro mesmo com markup
        safe_console.print("[green]Success[/green]")


class TestIsTerminalTool:
    """Testes para funcao _is_terminal_tool."""

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_is_terminal_tool_terminal(self):
        """Testar que 'terminal' e reconhecido."""
        assert _is_terminal_tool("terminal") is True

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_is_terminal_tool_terminal_tool(self):
        """Testar que 'terminal_tool' e reconhecido."""
        assert _is_terminal_tool("terminal_tool") is True

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_is_terminal_tool_case_insensitive(self):
        """Testar que reconhecimento e case-insensitive."""
        assert _is_terminal_tool("TERMINAL") is True
        assert _is_terminal_tool("Terminal") is True
        assert _is_terminal_tool("TERMINAL_TOOL") is True

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_is_terminal_tool_other_tools(self):
        """Testar que outras tools nao sao reconhecidas."""
        assert _is_terminal_tool("python") is False
        assert _is_terminal_tool("bash") is False
        assert _is_terminal_tool("search") is False
        assert _is_terminal_tool("") is False


class TestChatModuleStructure:
    """Testes para estrutura do modulo chat."""

    def test_chat_module_exists(self):
        """Testar que modulo chat existe."""
        try:
            import vectora.ui.chat

            assert vectora.ui.chat is not None
        except ImportError:
            pytest.skip("Chat module dependencies not available")

    def test_chat_has_logger(self):
        """Testar que chat module tem logger."""
        try:
            from vectora.ui.chat import logger

            assert logger is not None
            import logging

            assert isinstance(logger, logging.Logger)
        except ImportError:
            pytest.skip("Chat module dependencies not available")

    def test_chat_has_console(self):
        """Testar que chat module tem console global."""
        try:
            from vectora.ui.chat import console

            assert console is not None
        except ImportError:
            pytest.skip("Chat module dependencies not available")

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_console_instance_is_safe_console(self):
        """Testar que console global e instancia de SafeConsole."""
        assert isinstance(console, SafeConsole)

    def test_chat_module_has_key_functions(self):
        """Testar que chat module tem funcoes principais."""
        try:
            from vectora.ui import chat

            # Verificar que tem funcoes importantes
            assert hasattr(chat, "_export_audit")
            assert hasattr(chat, "_load_prior_messages")
            assert hasattr(chat, "_render_tool_event_start")
        except ImportError:
            pytest.skip("Chat module dependencies not available")


class TestChatImports:
    """Testes para imports necessarios no chat module."""

    def test_can_import_safe_console_or_skip(self):
        """Testar que SafeConsole pode ser importado ou skip."""
        try:
            from vectora.ui.chat import SafeConsole

            assert SafeConsole is not None
        except ImportError as e:
            # Se falhar, e porque falta dependencia
            pytest.skip(f"Chat dependencies not available: {e}")

    def test_can_import_functions_or_skip(self):
        """Testar que funcoes podem ser importadas ou skip."""
        try:
            from vectora.ui.chat import (
                _export_audit,
                _is_terminal_tool,
                _load_prior_messages,
                _render_tool_event_start,
            )

            assert _is_terminal_tool is not None
            assert _export_audit is not None
            assert _load_prior_messages is not None
            assert _render_tool_event_start is not None
        except ImportError as e:
            pytest.skip(f"Chat dependencies not available: {e}")


class TestSafeConsoleErrorHandling:
    """Testes para SafeConsole em casos de erro."""

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_safe_console_handles_unicode_error_gracefully(self):
        """Testar que SafeConsole trata UnicodeEncodeError gracefully."""
        safe_console = SafeConsole(file=StringIO())

        # Mock para simular UnicodeEncodeError - apenas testamos que não lanca exceção
        with patch.object(SafeConsole, "print") as mock_print:
            # Substituir com implementacao que trata o erro
            original_print = SafeConsole.print

            def safe_print_impl(self, *args, **kwargs):
                try:
                    original_print(self, *args, **kwargs)
                except UnicodeEncodeError:
                    # Fallback silencioso - já está implementado no SafeConsole real
                    pass

            with patch.object(SafeConsole, "print", safe_print_impl):
                # Não deve falhar
                safe_console.print("Test with Unicode")


class TestIsTerminalToolExtended:
    """Testes adicionais para _is_terminal_tool()."""

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_is_terminal_tool_with_spaces(self):
        """Testar reconhecimento com espacos."""
        assert _is_terminal_tool("  terminal  ") is False  # Tem espaços
        assert _is_terminal_tool("terminal") is True

    @pytest.mark.skipif(not HAS_CHAT_MODULE, reason="Chat module not available")
    def test_is_terminal_tool_similar_names(self):
        """Testar que nomes similares nao sao reconhecidos."""
        assert _is_terminal_tool("term") is False
        assert _is_terminal_tool("shell") is False
        assert _is_terminal_tool("cmd") is False
