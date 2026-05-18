"""Testes para vectora/ui/__init__.py (package exports)."""

from __future__ import annotations


class TestUiModuleExports:
    """Testes para exports do módulo ui."""

    def test_ui_module_has_all_attribute(self) -> None:
        """Verificar que módulo define __all__."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "__all__")
        assert isinstance(ui_module.__all__, list)

    def test_ui_module_all_contains_expected_exports(self) -> None:
        """Verificar que __all__ contém os exports esperados."""
        import vectora.ui as ui_module

        expected_exports = [
            "AuditPanel",
            "ChatMessage",
            "ErrorPanel",
            "LogPanel",
            "ProgressIndicator",
            "SeparatorLine",
            "SuccessPanel",
            "TerminalPanel",
            "ToolCallPanel",
            "ToolMessagePanel",
            "VectoraLayout",
            "VectoraStatusPanel",
            "WelcomeScreen",
        ]
        for export in expected_exports:
            assert export in ui_module.__all__

    def test_audit_panel_import(self) -> None:
        """Verificar que AuditPanel é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "AuditPanel")

    def test_chat_message_import(self) -> None:
        """Verificar que ChatMessage é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "ChatMessage")

    def test_error_panel_import(self) -> None:
        """Verificar que ErrorPanel é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "ErrorPanel")

    def test_log_panel_import(self) -> None:
        """Verificar que LogPanel é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "LogPanel")

    def test_progress_indicator_import(self) -> None:
        """Verificar que ProgressIndicator é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "ProgressIndicator")

    def test_separator_line_import(self) -> None:
        """Verificar que SeparatorLine é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "SeparatorLine")

    def test_success_panel_import(self) -> None:
        """Verificar que SuccessPanel é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "SuccessPanel")

    def test_terminal_panel_import(self) -> None:
        """Verificar que TerminalPanel é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "TerminalPanel")

    def test_tool_call_panel_import(self) -> None:
        """Verificar que ToolCallPanel é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "ToolCallPanel")

    def test_tool_message_panel_import(self) -> None:
        """Verificar que ToolMessagePanel é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "ToolMessagePanel")

    def test_vectora_layout_import(self) -> None:
        """Verificar que VectoraLayout é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "VectoraLayout")

    def test_vectora_status_panel_import(self) -> None:
        """Verificar que VectoraStatusPanel é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "VectoraStatusPanel")

    def test_welcome_screen_import(self) -> None:
        """Verificar que WelcomeScreen é importado."""
        import vectora.ui as ui_module

        assert hasattr(ui_module, "WelcomeScreen")
