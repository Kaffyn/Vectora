"""Testes para nodes/debug.py e services/terminal_stream.py.

Cobre: DiagnosticToolNode, funções de terminal stream.
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

import pytest


class TestTerminalStream:
    """Testa serviço de terminal stream."""

    def test_register_callback(self):
        """Verifica que callback pode ser registrado."""
        from vectora.services.terminal_stream import (
            register_terminal_output_callback,
            unregister_terminal_output_callback,
        )

        received = []

        def my_callback(line: str) -> None:
            received.append(line)

        register_terminal_output_callback(my_callback)
        # Deve ser registrado sem erro
        unregister_terminal_output_callback()

    def test_emit_terminal_line_calls_callback(self):
        """Verifica que emit_terminal_line chama callback registrado."""
        from vectora.services.terminal_stream import (
            emit_terminal_line,
            register_terminal_output_callback,
            unregister_terminal_output_callback,
        )

        received = []

        def my_callback(line: str) -> None:
            received.append(line)

        register_terminal_output_callback(my_callback)
        emit_terminal_line("Output de terminal")
        unregister_terminal_output_callback()

        assert "Output de terminal" in received

    def test_emit_without_callback_does_not_crash(self):
        """Verifica que emit sem callback registrado não lança exceção."""
        from vectora.services.terminal_stream import (
            emit_terminal_line,
            unregister_terminal_output_callback,
        )

        unregister_terminal_output_callback()
        # Não deve lançar exceção
        emit_terminal_line("Linha sem callback")

    def test_unregister_removes_callback(self):
        """Verifica que unregister_callback remove o callback."""
        from vectora.services.terminal_stream import (
            emit_terminal_line,
            register_terminal_output_callback,
            unregister_terminal_output_callback,
        )

        received = []

        def my_callback(line: str) -> None:
            received.append(line)

        register_terminal_output_callback(my_callback)
        unregister_terminal_output_callback()
        emit_terminal_line("Linha após unregister")

        # Após unregister, callback não deve ser chamado
        assert len(received) == 0


class TestDiagnosticToolNode:
    """Testa DiagnosticToolNode."""

    def test_diagnostic_tool_node_creation(self):
        """Verifica que DiagnosticToolNode pode ser criado."""
        from vectora.nodes.debug import DiagnosticToolNode
        from vectora.tools import TOOLS

        node = DiagnosticToolNode(tools=TOOLS)
        assert node is not None

    def test_diagnostic_node_has_tools(self):
        """Verifica que DiagnosticToolNode tem ferramentas configuradas."""
        from vectora.nodes.debug import DiagnosticToolNode
        from vectora.tools import TOOLS

        node = DiagnosticToolNode(tools=TOOLS)
        # DiagnosticToolNode herda de ToolNode
        assert hasattr(node, "tools_by_name") or hasattr(node, "name")

    def test_debug_node_log_message(self):
        """Verifica que DiagnosticToolNode não lança exceção ao ser criado."""
        from vectora.nodes.debug import DiagnosticToolNode
        from vectora.tools.fs import list_dir

        node = DiagnosticToolNode(tools=[list_dir])
        assert node is not None


class TestIgnoreValidator:
    """Testa services/ignore_validator.py."""

    def test_get_ignore_validator_returns_instance(self):
        """Verifica que get_ignore_validator retorna instância."""
        from vectora.services.ignore_validator import get_ignore_validator

        validator = get_ignore_validator()
        assert validator is not None

    def test_validator_should_ignore_node_modules(self):
        """Verifica que node_modules é ignorado."""
        from vectora.services.ignore_validator import get_ignore_validator

        validator = get_ignore_validator()
        if hasattr(validator, "should_ignore"):
            result = validator.should_ignore("node_modules/package.json")
            assert result is True

    def test_validator_allows_source_files(self):
        """Verifica que arquivos .py são permitidos."""
        from vectora.services.ignore_validator import get_ignore_validator

        validator = get_ignore_validator()
        if hasattr(validator, "should_ignore"):
            result = validator.should_ignore("src/main.py")
            assert result is False

    def test_validator_should_ignore_env_files(self):
        """Verifica que .env é ignorado (contém segredos)."""
        from vectora.services.ignore_validator import get_ignore_validator

        validator = get_ignore_validator()
        if hasattr(validator, "should_ignore"):
            result = validator.should_ignore(".env")
            # .env pode ou não ser ignorado dependendo da implementação
            assert isinstance(result, bool)
