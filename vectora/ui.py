"""Rich UI Components for Vectora - "Rich Gorda" Dashboard.

Advanced visual components using Rich for a professional, real-time CLI experience.
Includes layouts, status indicators, panels, and live rendering capabilities.
"""

from datetime import datetime
from pathlib import Path
from typing import Any

from rich.console import Console
from rich.layout import Layout
from rich.markdown import Markdown
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.rule import Rule
from rich.table import Table
from rich.text import Text


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
            f"[bold cyan]🚀 Vectora v0.1.0[/bold cyan] | "
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
        worker_indicator = "[green]●[/green]" if worker_active else "[dim]○[/dim]"
        footer_text = (
            f"{worker_indicator} Background Worker | "
            f"[cyan]Embedding Queue: {embedding_queue}[/cyan] | "
            f"[yellow]RAG: {rag_status}[/yellow]"
        )
        self.layout["footer"].update(Panel(footer_text, style="dim", expand=False))

    def render(self) -> Layout:
        """Return the layout for rendering."""
        return self.layout


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
                title="[bold cyan]👤 You[/bold cyan]",
                style="cyan",
                expand=False,
                border_style="cyan",
            )
        # Render AI response as markdown for better formatting
        markdown = Markdown(self.content)
        return Panel(
            markdown,
            title="[bold magenta]🤖 Vectora[/bold magenta]",
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
            role_emoji = "👤" if msg.role.lower() == "user" else "🤖"
            table.add_row(str(i), f"{role_emoji} {msg.role}", content_preview)

        return Panel(
            table,
            title="[bold]📋 Recent Messages[/bold]",
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

[green]✓ Ready to chat![/green]
"""
        return Panel(welcome_text, title="[bold cyan]🚀 Vectora Chat[/bold cyan]")


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
            title=f"[bold red]❌ {title}[/bold red]",
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
            title=f"[bold green]✓ {title}[/bold green]",
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
