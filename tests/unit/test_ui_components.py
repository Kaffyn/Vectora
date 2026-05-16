"""Tests for Rich UI components - "Rich Gorda" dashboard."""

from datetime import datetime
from pathlib import Path

import pytest
from rich.console import Console

from vectora.ui import (
    AuditPanel,
    ChatMessage,
    ErrorPanel,
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
