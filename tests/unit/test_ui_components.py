"""Tests for Rich UI components - "Rich Gorda" dashboard."""

import logging
from datetime import datetime
from pathlib import Path
from queue import Queue

import pytest
from rich.console import Console

from vectora.ui import (
    AuditPanel,
    ChatMessage,
    ErrorPanel,
    LogPanel,
    ProgressIndicator,
    SeparatorLine,
    SuccessPanel,
    VectoraLayout,
    VectoraStatusPanel,
    WelcomeScreen,
)


class TestChatMessage:
    """Test ChatMessage component."""

    def test_user_message_creation(self):
        """Test creating a user message."""
        msg = ChatMessage("User", "Hello, world!")
        assert msg.role == "User"
        assert msg.content == "Hello, world!"
        assert msg.timestamp is not None

    def test_ai_message_creation(self):
        """Test creating an AI message."""
        msg = ChatMessage("Vectora", "Hi there!")
        assert msg.role == "Vectora"
        assert msg.content == "Hi there!"

    def test_message_to_panel_user(self):
        """Test converting user message to panel."""
        msg = ChatMessage("User", "Test message")
        panel = msg.to_panel()
        assert panel is not None
        # Panel object is created successfully
        assert hasattr(panel, "title")

    def test_message_to_panel_ai(self):
        """Test converting AI message to panel."""
        msg = ChatMessage("Vectora", "Test response")
        panel = msg.to_panel()
        assert panel is not None
        # Panel object is created successfully
        assert hasattr(panel, "title")

    def test_message_to_markdown_line(self):
        """Test converting message to markdown."""
        msg = ChatMessage("User", "Test message")
        markdown = msg.to_markdown_line()
        assert "**User**" in markdown
        assert "Test message" in markdown
        assert "---" in markdown


class TestAuditPanel:
    """Test AuditPanel for conversation tracking."""

    def test_audit_panel_creation(self):
        """Test creating an audit panel."""
        audit = AuditPanel(max_visible=5)
        assert audit.max_visible == 5
        assert len(audit.messages) == 0

    def test_add_message_to_audit(self):
        """Test adding messages to audit."""
        audit = AuditPanel()
        audit.add_message("User", "Hello")
        audit.add_message("Vectora", "Hi there!")

        assert len(audit.messages) == 2
        assert audit.messages[0].role == "User"
        assert audit.messages[1].role == "Vectora"

    def test_audit_render(self):
        """Test rendering audit panel."""
        audit = AuditPanel()
        audit.add_message("User", "Test")
        panel = audit.render()
        assert panel is not None

    def test_audit_to_markdown(self):
        """Test exporting audit to markdown."""
        audit = AuditPanel()
        audit.add_message("User", "Hello")
        audit.add_message("Vectora", "Hi!")

        markdown = audit.to_markdown()
        assert "# Session Audit" in markdown
        assert "**User**" in markdown
        assert "**Vectora**" in markdown
        assert "Hello" in markdown
        assert "Hi!" in markdown

    def test_audit_save_to_file(self, tmp_path):
        """Test saving audit to file."""
        audit = AuditPanel()
        audit.add_message("User", "Test message")

        filepath = tmp_path / "audit.md"
        saved_path = audit.save_to_file(filepath)

        assert saved_path.exists()
        content = saved_path.read_text()
        assert "Session Audit" in content
        assert "Test message" in content

    def test_audit_save_default_location(self, tmp_path, monkeypatch):
        """Test saving audit to default location."""
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        audit = AuditPanel()
        audit.add_message("User", "Default location test")

        saved_path = audit.save_to_file()

        assert saved_path.exists()
        assert ".vectora" in str(saved_path)
        assert "logs" in str(saved_path)


class TestVectoraLayout:
    """Test VectoraLayout dashboard."""

    def test_layout_creation(self):
        """Test creating a layout."""
        layout = VectoraLayout()
        assert layout.layout is not None

    def test_update_header(self):
        """Test updating header."""
        layout = VectoraLayout()
        layout.update_header(provider="google-genai", thread_id=1, message_count=5)
        # Just verify it doesn't crash
        assert layout.layout is not None

    def test_update_footer(self):
        """Test updating footer."""
        layout = VectoraLayout()
        layout.update_footer(embedding_queue=3, rag_status="Ready", worker_active=True)
        assert layout.layout is not None

    def test_render_layout(self):
        """Test rendering the layout."""
        layout = VectoraLayout()
        rendered = layout.render()
        assert rendered is not None


class TestVectoraStatusPanel:
    """Test VectoraStatusPanel."""

    def test_status_panel_creation(self):
        """Test creating a status panel."""
        console = Console()
        status = VectoraStatusPanel(console)
        assert status.console is not None

    def test_thinking_context(self):
        """Test thinking indicator."""
        console = Console()
        status = VectoraStatusPanel(console)
        ctx = status.thinking("Processing...")
        assert ctx is not None

    def test_processing_documents(self):
        """Test document processing panel."""
        console = Console()
        status = VectoraStatusPanel(console)
        panel = status.processing_documents(5)
        assert panel is not None

    def test_connection_test_context(self):
        """Test connection test indicator."""
        console = Console()
        status = VectoraStatusPanel(console)
        ctx = status.connection_test()
        assert ctx is not None


class TestWelcomeScreen:
    """Test WelcomeScreen component."""

    def test_welcome_screen_render(self):
        """Test rendering welcome screen."""
        panel = WelcomeScreen.render(provider="google-genai", model="gemini-2.0")
        assert panel is not None
        assert hasattr(panel, "title")

    def test_welcome_screen_with_defaults(self):
        """Test welcome screen with default values."""
        panel = WelcomeScreen.render()
        assert panel is not None
        assert hasattr(panel, "title")


class TestErrorPanel:
    """Test ErrorPanel component."""

    def test_error_panel_with_string(self):
        """Test error panel with string message."""
        panel = ErrorPanel.render("Something went wrong", title="Fatal Error")
        assert panel is not None
        assert hasattr(panel, "title")

    def test_error_panel_with_exception(self):
        """Test error panel with exception."""
        error_msg = "Test error"
        try:
            raise ValueError(error_msg)
        except ValueError as e:
            panel = ErrorPanel.render(e)
            assert panel is not None
            assert hasattr(panel, "title")


class TestSuccessPanel:
    """Test SuccessPanel component."""

    def test_success_panel(self):
        """Test success panel."""
        panel = SuccessPanel.render("Operation completed successfully")
        assert panel is not None
        assert hasattr(panel, "title")


class TestSeparatorLine:
    """Test SeparatorLine component."""

    def test_separator_with_title(self):
        """Test separator with title."""
        sep = SeparatorLine.render("Section Title")
        assert sep is not None

    def test_separator_without_title(self):
        """Test separator without title."""
        sep = SeparatorLine.render()
        assert sep is not None


class TestProgressIndicator:
    """Test ProgressIndicator component."""

    def test_progress_creation(self):
        """Test creating progress indicator."""
        console = Console()
        progress = ProgressIndicator(console)
        assert progress.console is not None

    def test_embedding_progress(self):
        """Test embedding progress bar."""
        console = Console()
        progress = ProgressIndicator(console)
        prog = progress.embedding_progress(100)
        assert prog is not None

    def test_stream_indicator(self):
        """Test stream indicator."""
        console = Console()
        progress = ProgressIndicator(console)
        indicator = progress.stream_indicator()
        assert indicator is not None
        assert "▌" in str(indicator)


class TestLogPanel:
    """Test LogPanel for Debug Mode."""

    def test_log_panel_creation(self):
        """Test creating a log panel."""
        log_queue = Queue()
        log_panel = LogPanel(log_queue, max_lines=10)
        assert log_panel.log_queue is not None
        assert log_panel.max_lines == 10
        assert len(log_panel.logs) == 0

    def test_log_panel_render_empty(self):
        """Test rendering empty log panel."""
        log_queue = Queue()
        log_panel = LogPanel(log_queue)
        panel = log_panel.render()
        assert panel is not None
        assert hasattr(panel, "title")

    def test_log_panel_with_log_record(self):
        """Test adding log records to panel."""
        log_queue = Queue()
        log_panel = LogPanel(log_queue, max_lines=5)

        # Create a log record and put it in the queue
        record = logging.LogRecord(
            name="test.logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Test log message",
            args=(),
            exc_info=None,
        )
        log_queue.put(record)

        # Update and render
        log_panel.update_logs()
        assert len(log_panel.logs) == 1
        assert log_panel.logs[0][0] == "INFO"
        assert log_panel.logs[0][1] == "Test log message"

    def test_log_panel_circular_buffer(self):
        """Test that log panel maintains max lines."""
        log_queue = Queue()
        log_panel = LogPanel(log_queue, max_lines=3)

        # Add 5 log records
        for i in range(5):
            record = logging.LogRecord(
                name="test.logger",
                level=logging.INFO,
                pathname="test.py",
                lineno=i,
                msg=f"Message {i}",
                args=(),
                exc_info=None,
            )
            log_queue.put(record)

        log_panel.update_logs()

        # Should only have 3 most recent logs
        assert len(log_panel.logs) == 3
        assert log_panel.logs[0][1] == "Message 2"
        assert log_panel.logs[2][1] == "Message 4"

    def test_log_panel_color_levels(self):
        """Test log level coloring."""
        log_queue = Queue()
        log_panel = LogPanel(log_queue)

        assert log_panel._get_log_color("DEBUG") == "cyan"
        assert log_panel._get_log_color("INFO") == "green"
        assert log_panel._get_log_color("WARNING") == "yellow"
        assert log_panel._get_log_color("ERROR") == "red"
        assert log_panel._get_log_color("CRITICAL") == "bold red"

    def test_log_panel_render_with_logs(self):
        """Test rendering panel with log entries."""
        log_queue = Queue()
        log_panel = LogPanel(log_queue)

        for level in ["DEBUG", "INFO", "WARNING", "ERROR"]:
            record = logging.LogRecord(
                name="test",
                level=getattr(logging, level),
                pathname="test.py",
                lineno=1,
                msg=f"Test {level}",
                args=(),
                exc_info=None,
            )
            log_queue.put(record)

        log_panel.update_logs()
        panel = log_panel.render()
        assert panel is not None


class TestVectoraLayoutDebugMode:
    """Test VectoraLayout debug mode functionality."""

    def test_split_with_debug(self):
        """Test creating debug split layout."""
        layout = VectoraLayout()
        log_queue = Queue()
        layout.split_with_debug(log_queue)

        # Verify debug layout was created by checking child names
        child_names = [child.name for child in layout.layout.children]
        assert "debug" in child_names
        assert "main" in child_names

    def test_get_main_layout(self):
        """Test getting main layout from split."""
        layout = VectoraLayout()
        log_queue = Queue()
        layout.split_with_debug(log_queue)

        main = layout.get_main_layout()
        assert main is not None
        # Verify the main layout has the expected structure
        main_child_names = [child.name for child in main.children]
        assert "header" in main_child_names
        assert "body" in main_child_names
        assert "footer" in main_child_names

    def test_update_debug_panel(self):
        """Test updating debug panel."""
        layout = VectoraLayout()
        log_queue = Queue()
        layout.split_with_debug(log_queue)

        from rich.panel import Panel

        debug_panel = Panel("Test content")
        layout.update_debug_panel(debug_panel)

        # Should not raise exception
        assert layout is not None
