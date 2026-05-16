"""Rich UI Components for Vectora - "Rich Gorda" Dashboard.

Advanced visual components using Rich for a professional, real-time CLI experience.
Includes layouts, status indicators, panels, and live rendering capabilities.
Features Debug Mode with real-time log panel for production monitoring.
"""

from datetime import datetime
from pathlib import Path
from queue import Queue
from typing import Any

from rich.console import Console
from rich.layout import Layout
from rich.markdown import Markdown
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.rule import Rule
from rich.table import Table
from rich.text import Text
from version import __version__


class VectoraLayout:
    """Professional dashboard layout for Vectora chat."""

    def __init__(self) -> None:
        """Initialize the dashboard layout."""
        self.layout = Layout()
        self._build_layout()

    def _build_layout(self) -> None:
        """Build the three-section layout: header, body, footer."""
        self.layout.split_column(
            Layout(name="header", size=4),
            Layout(name="body"),
            Layout(name="footer", size=3),
        )

    def update_header(
        self,
        provider: str = "unset",
        thread_id: int = 1,
        message_count: int = 0,
    ) -> None:
        """Update header with session info."""
        header_text = (
            f"[bold cyan][ROCKET] Vectora v{__version__}[/bold cyan] | "
            f"[yellow]Provider: {provider}[/yellow] | "
            f"[magenta]Thread: {thread_id}[/magenta] | "
            f"[green]Messages: {message_count}[/green]"
        )
        self.layout["header"].update(Panel(header_text, style="blue", expand=False))

    def update_body(self, content: str | Panel | Table) -> None:
        """Update main body with chat or content."""
        self.layout["body"].update(content)

    def update_footer(
        self,
        embedding_queue: int = 0,
        rag_status: str = "Ready",
        *,
        worker_active: bool = False,
    ) -> None:
        """Update footer with background worker status."""
        worker_indicator = "[green]*[/green]" if worker_active else "[dim]o[/dim]"
        footer_text = (
            f"{worker_indicator} Background Worker | "
            f"[cyan]Embedding Queue: {embedding_queue}[/cyan] | "
            f"[yellow]RAG: {rag_status}[/yellow]"
        )
        self.layout["footer"].update(Panel(footer_text, style="dim", expand=False))

    def render(self) -> Layout:
        """Return the layout for rendering."""
        return self.layout

    def split_with_debug(self, log_queue: Queue) -> Layout:
        """Return layout with debug panel in right sidebar (Debug Mode)."""
        self.layout.split_column(
            Layout(name="main"),
            Layout(name="debug", size=40),
        )
        self.layout["main"].split_column(
            Layout(name="header", size=4),
            Layout(name="body"),
            Layout(name="footer", size=3),
        )
        return self.layout

    def get_main_layout(self) -> Layout:
        """Get the main layout area (for use in split mode)."""
        return self.layout["main"]

    def update_debug_panel(self, log_panel: Panel) -> None:
        """Update debug panel with log data."""
        if "debug" in self.layout.children:
            self.layout["debug"].update(log_panel)


class LogPanel:
    """Real-time log display panel for Debug Mode."""

    def __init__(self, log_queue: Queue, max_lines: int = 20) -> None:
        """Initialize log panel.

        Args:
            log_queue: Queue containing log records
            max_lines: Maximum lines to display (circular buffer)
        """
        self.log_queue = log_queue
        self.max_lines = max_lines
        self.logs: list[tuple[str, str]] = []  # (level, message) tuples

    def _get_log_color(self, level: str) -> str:
        """Get color for log level."""
        level_colors = {
            "DEBUG": "cyan",
            "INFO": "green",
            "WARNING": "yellow",
            "ERROR": "red",
            "CRITICAL": "bold red",
        }
        return level_colors.get(level, "white")

    def update_logs(self) -> None:
        """Pull new logs from queue (non-blocking)."""
        while not self.log_queue.empty():
            try:
                record = self.log_queue.get_nowait()
                level = record.levelname
                message = record.getMessage()
                self.logs.append((level, message))

                # Keep only the last N lines
                if len(self.logs) > self.max_lines:
                    self.logs.pop(0)
            except Exception:
                break

    def render(self) -> Panel:
        """Render log panel as styled table."""
        self.update_logs()

        table = Table(show_header=False, box=None, padding=(0, 1))

        if not self.logs:
            table.add_row("[dim]No logs yet...[/dim]")
        else:
            # Show most recent logs at the bottom
            for level, message in self.logs[-self.max_lines :]:
                color = self._get_log_color(level)
                level_badge = f"[{color}][{level:7}][/{color}]"
                table.add_row(f"{level_badge} {message}")

        return Panel(
            table,
            title="[bold cyan][GEAR] DEBUG LOGS[/bold cyan]",
            style="cyan",
            expand=True,
            border_style="cyan",
        )


class VectoraStatusPanel:
    """Real-time status panel for long-running operations."""

    def __init__(self, console: Console) -> None:
        """Initialize status panel."""
        self.console = console

    def thinking(self, message: str = "Thinking...") -> Any:
        """Show a spinner while the model thinks."""
        return self.console.status(
            f"[bold cyan]{message}[/bold cyan]",
            spinner="dots",
            spinner_style="cyan",
        )

    def processing_documents(self, count: int) -> Panel:
        """Show document processing status."""
        table = Table(show_header=False, box=None)
        table.add_row("[cyan]Documents Processed[/cyan]", f"[bold]{count}[/bold]")
        return Panel(table, title="[bold]RAG Pipeline[/bold]", style="green")

    def connection_test(self) -> Any:
        """Show spinner for connection testing."""
        return self.console.status(
            "[bold cyan]Testing connection to LLM...[/bold cyan]",
            spinner="dots",
        )


class ChatMessage:
    """Formatted chat message with styling."""

    def __init__(self, role: str, content: str) -> None:
        """Initialize a chat message."""
        self.role = role
        self.content = content
        self.timestamp = datetime.now()

    def to_panel(self) -> Panel:
        """Render as a styled panel."""
        if self.role.lower() == "user":
            return Panel(
                self.content,
                title="[bold cyan][USER] You: [/bold cyan]",
                style="cyan",
                expand=False,
                border_style="cyan",
            )
        # Render AI response as markdown for better formatting
        markdown = Markdown(self.content)
        return Panel(
            markdown,
            title="[bold magenta][Vectora][/bold magenta]",
            style="magenta",
            expand=False,
            border_style="magenta",
        )

    def to_markdown_line(self) -> str:
        """Render as markdown for file export."""
        role_text = "**User**" if self.role.lower() == "user" else "**Vectora**"
        return f"## {role_text}\n\n{self.content}\n\n---\n\n"


class AuditPanel:
    """Panel for displaying conversation audit and history."""

    def __init__(self, max_visible: int = 5) -> None:
        """Initialize audit panel."""
        self.messages: list[ChatMessage] = []
        self.max_visible = max_visible

    def add_message(self, role: str, content: str) -> None:
        """Add a message to the audit."""
        self.messages.append(ChatMessage(role, content))

    def render(self) -> Panel:
        """Render audit as a table."""
        table = Table(show_header=True, header_style="bold cyan")
        table.add_column("#", style="dim", width=4)
        table.add_column("Role", style="bold")
        table.add_column("Content", style="white")

        # Show only the last N messages
        for i, msg in enumerate(self.messages[-self.max_visible :], 1):
            content_preview = (
                msg.content[:50] + "..." if len(msg.content) > 50 else msg.content
            )
            role_emoji = "[USER]" if msg.role.lower() == "user" else "[Vectora]"
            table.add_row(str(i), f"{role_emoji} {msg.role}", content_preview)

        return Panel(
            table,
            title="[bold][LIST] Recent Messages[/bold]",
            style="green",
            expand=False,
        )

    def to_markdown(self) -> str:
        """Export entire audit as markdown."""
        lines = [
            f"# Session Audit - {len(self.messages)} messages\n\n",
            f"*Generated: {datetime.now().isoformat()}*\n\n",
        ]
        lines.extend(msg.to_markdown_line() for msg in self.messages)
        return "".join(lines)

    def save_to_file(self, filepath: Path | None = None) -> Path:
        """Save audit to file."""
        if filepath is None:
            log_dir = Path.home() / ".vectora" / "logs"
            log_dir.mkdir(parents=True, exist_ok=True)
            filepath = (
                log_dir / f"session_{datetime.now().strftime('%Y%m%d_%H%M%S')}.md"
            )

        filepath.write_text(self.to_markdown(), encoding="utf-8")
        return filepath


class WelcomeScreen:
    """Welcome and information screen."""

    @staticmethod
    def render(provider: str = "unset", model: str = "unknown") -> Panel:
        """Render welcome screen."""
        welcome_text = f"""
[bold cyan]Welcome to Vectora![/bold cyan]

[yellow]Provider:[/yellow] [bold]{provider}[/bold]
[yellow]Model:[/yellow] [bold]{model}[/bold]

[dim]Commands:[/dim]
  [cyan]/sair[/cyan]  - Exit the chat
  [cyan]/quit[/cyan]  - Exit the chat
  [cyan]/q[/cyan]     - Exit the chat
  [cyan]Ctrl+C[/cyan] - Interrupt current operation

[green][OK] Ready to chat![/green]
"""
        return Panel(welcome_text, title="[bold cyan][ROCKET] Vectora Chat[/bold cyan]")


class ProgressIndicator:
    """Beautiful progress indicators for long operations."""

    def __init__(self, console: Console) -> None:
        """Initialize progress indicator."""
        self.console = console

    def embedding_progress(self, total: int) -> Progress:
        """Show embedding progress bar."""
        return Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            transient=True,
            console=self.console,
        )

    def stream_indicator(self) -> Text:
        """Animated streaming indicator."""
        return Text("▌", style="bold cyan")


class ErrorPanel:
    """Professional error display."""

    @staticmethod
    def render(error: Exception | str, title: str = "Error") -> Panel:
        """Render error as styled panel."""
        error_text = str(error)
        return Panel(
            f"[red]{error_text}[/red]",
            title=f"[bold red][X] {title}[/bold red]",
            style="red",
            expand=False,
            border_style="red",
        )


class SuccessPanel:
    """Professional success display."""

    @staticmethod
    def render(message: str, title: str = "Success") -> Panel:
        """Render success as styled panel."""
        return Panel(
            f"[green]{message}[/green]",
            title=f"[bold green][OK] {title}[/bold green]",
            style="green",
            expand=False,
            border_style="green",
        )


class SeparatorLine:
    """Visual separator between sections."""

    @staticmethod
    def render(title: str = "") -> Rule:
        """Render a separator line."""
        return Rule(f"[bold]{title}[/bold]" if title else "", style="dim")
