"""Textual TUI Chat Interface for Vectora.

Implements the interactive chat terminal UI with real-time message rendering,
conversation history, and status display. Uses Textual framework for responsive,
cross-platform CLI experience.

Components:
    - MessageDisplay: Rich formatted message rendering
    - InputBox: User input with history and autocomplete
    - StatusBar: Real-time status and metrics display
    - ChatApp: Main Textual application class
"""

import asyncio
import logging
from collections.abc import Sequence
from contextlib import nullcontext
from datetime import datetime
from pathlib import Path
from typing import Any, Self

from background_worker import get_background_worker
from checkpointer import Checkpointer
from constants import DB_DSN, LOGS_DIR
from context import Context
from graph import build_graph
from langchain_core.messages import AIMessage, BaseMessage, HumanMessage, ToolMessage
from langgraph.graph.state import CompiledStateGraph, RunnableConfig
from log_setup import setup_logging
from rich.console import Console
from rich.markdown import Markdown
from rich.panel import Panel
from rich.text import Text
from state import State
from textual.app import App, ComposeResult, on
from textual.containers import Container
from textual.widgets import Footer, Header, Input, RichLog, Static
from utils import async_lifespan

logger = logging.getLogger(__name__)


class ChatHeader(Static):
    """Header widget for the chat application."""

    def __init__(self, context: Context | None = None) -> None:
        """Initialize header with optional context for displaying correlation_id."""
        super().__init__()
        self.context = context

    def render(self) -> Panel:
        """Render the header with version and optional correlation_id."""
        title_text = "⚡ Vectora Chat v0.0.1dev1"
        title = Text(title_text, style="bold cyan")

        if self.context:
            # Display correlation_id for debugging (useful for QA/support)
            correlation_info = Text(
                f"Thread: {self.context.thread_id} | Trace: {self.context.correlation_id}",
                style="dim white",
            )
            return Panel(
                title,
                subtitle=correlation_info,
                style="blue",
                height=4,
            )

        return Panel(
            title,
            style="blue",
            height=3,
        )


class ChatMessages(RichLog):
    """Widget to display chat messages."""

    def add_human_message(self: Self, text: str) -> None:
        """Add a human message to the log."""
        timestamp = datetime.now().strftime("%H:%M:%S")
        header = Text(f"[{timestamp}] Você:", style="bold green")
        self.write(header)
        self.write(Markdown(text))
        self.write(Text(""))

    def add_ai_message(self: Self, text: str, model: str = "") -> None:
        """Add an AI message to the log."""
        timestamp = datetime.now().strftime("%H:%M:%S")
        model_info = f" ({model})" if model else ""
        header = Text(f"[{timestamp}] Vectora{model_info}:", style="bold cyan")
        self.write(header)
        self.write(Markdown(text))
        self.write(Text(""))

    def add_system_message(self: Self, text: str) -> None:
        """Add a system message to the log."""
        msg = Text(f"ℹ️  {text}", style="dim yellow")
        self.write(msg)
        self.write(Text(""))


class ChatInput(Input):
    """Input widget for the chat."""

    def __init__(self) -> None:
        """Initialize the input widget."""
        super().__init__(
            placeholder="Digite sua mensagem (ou 'sair' para encerrar)...",
            id="chat-input",
        )


class ChatContainer(Static):
    """Container for chat components.

    Recebe o grafo já compilado via injeção de dependência no construtor.
    O log é uma view do estado do checkpointer (Pull Model) — não há estado
    paralelo de mensagens nesta classe.
    """

    thread_id: int = 1
    context: Context

    def __init__(self, graph: CompiledStateGraph[State, Context, State, State]) -> None:
        """Initialize the chat container with an injected, pre-built graph."""
        super().__init__()
        self.graph = graph
        self.context = Context(user_type="plus", thread_id=self.thread_id)

    async def on_mount(self) -> None:
        """Display welcome message and focus input on mount.

        A inicialização de recursos pesados (checkpointer, grafo) é feita
        externamente em `run_chat()`, antes de instanciar este widget.
        O `on_mount` apenas configura a UI inicial.
        """
        chat_log = self.query_one(ChatMessages)

        # Update header with correlation info for debugging/support
        try:
            header = self.app.query_one(ChatHeader)
            header.context = self.context
            header.refresh()
        except Exception:
            pass  # Header might not be available, continue anyway

        config = RunnableConfig(configurable={"thread_id": self.thread_id})
        try:
            # Pull Model: verifica se há histórico anterior no checkpointer
            existing_state = await self.graph.aget_state(config)
            prior_messages: Sequence[BaseMessage] = existing_state.values.get(
                "messages", []
            )

            if prior_messages:
                chat_log.add_system_message(
                    f"Sessão retomada — {len(prior_messages)} mensagem(s) carregada(s) do banco."
                )
                self._render_messages(chat_log, prior_messages)
            else:
                chat_log.add_system_message(
                    "Bem-vindo ao Vectora! 🤖 Digite suas perguntas abaixo."
                )
        except Exception:
            logger.exception("Falha ao verificar estado anterior")
            chat_log.add_system_message(
                "Bem-vindo ao Vectora! 🤖 Digite suas perguntas abaixo."
            )

        self.query_one(ChatInput).focus()

    def _render_messages(
        self, chat_log: ChatMessages, messages: Sequence[BaseMessage]
    ) -> None:
        """Re-renderiza o log completo a partir de uma sequência de mensagens.

        O log é limpo e reconstruído a partir do estado lido do checkpointer,
        garantindo que a view seja sempre um reflexo fiel do banco de dados.
        """
        chat_log.clear()
        for msg in messages:
            content = msg.content if hasattr(msg, "content") else str(msg)
            if not isinstance(content, str):
                content = str(content)

            if isinstance(msg, HumanMessage):
                chat_log.add_human_message(content)
            elif isinstance(msg, AIMessage):
                model_name = (
                    msg.response_metadata.get("model", "")
                    if hasattr(msg, "response_metadata")
                    else ""
                )
                chat_log.add_ai_message(content, model_name)
            # ToolMessages e outros tipos são silenciados na UI (detalhes de implementação)

    @on(Input.Submitted)
    async def on_submit(self, event: Input.Submitted) -> None:
        """Handle input submission using reactive streaming (astream_events).

        Fluxo reativo:
        1. Dispara astream_events (Fire-and-forget, sem bloquear)
        2. Consome eventos em tempo real (chunks, node start/end)
        3. Atualiza UI progressivamente (streaming de resposta)
        4. Persiste no checkpointer automaticamente
        """
        event.stop()
        user_input = event.value
        chat_log = self.query_one(ChatMessages)
        chat_input = self.query_one(ChatInput)

        if not user_input.strip():
            return

        if user_input.lower() in ["sair", "quit", "q"]:
            logger.info("Chat encerrado pelo usuário")
            self.app.exit()
            return

        try:
            chat_input.value = ""
            chat_log.add_human_message(user_input)

            config = RunnableConfig(
                configurable={"thread_id": self.thread_id},
            )

            logger.debug(
                "Processing user input",
                extra={"thread_id": self.thread_id, "input_length": len(user_input)},
            )

            human_message = HumanMessage(user_input)

            # LangSmith tracing é auto-injetado via env vars (LANGSMITH_API_KEY, etc)
            # Streaming reativo com astream_events
            ai_response = ""
            with nullcontext():
                async for graph_event in self.graph.astream_events(
                    {"messages": [human_message]},
                    config=config,
                    version="v2",
                ):
                    event_type = graph_event.get("event")
                    data = graph_event.get("data", {})

                    # Streaming de tokens do LLM
                    if event_type == "on_chat_model_stream":
                        chunk = data.get("chunk")
                        if chunk and hasattr(chunk, "content"):
                            content = chunk.content
                            if content:
                                ai_response += content
                                chat_log.write(content)

                    # Nó iniciando
                    elif event_type == "on_node_start":
                        node_name = graph_event.get("name", "")
                        if node_name:
                            logger.debug(f"Node started: {node_name}")

                    # Nó terminando
                    elif event_type == "on_node_end":
                        node_name = graph_event.get("name", "")
                        if node_name:
                            logger.debug(f"Node ended: {node_name}")

            # Garantir que mensagem de IA foi adicionada ao log
            if ai_response:
                chat_log.write("")

            logger.info(
                "User input processed",
                extra={
                    "thread_id": self.thread_id,
                    "response_length": len(ai_response),
                },
            )

        except Exception:
            logger.exception("Error processing input", exc_info=True)
            chat_log.add_system_message("Erro ao processar: <error>")

    def compose(self) -> ComposeResult:
        """Compose the layout."""
        yield ChatMessages(highlight=True, markup=True, id="chat-log")
        yield Container(ChatInput())


class VectoraChatApp(App[Any]):
    """Main Vectora Chat TUI application."""

    CSS = """
    Screen {
        layout: vertical;
    }

    #chat-log {
        width: 1fr;
        height: 1fr;
        border: solid $primary;
    }

    #chat-input {
        width: 1fr;
        height: auto;
        border: solid $accent;
    }
    """

    BINDINGS = [
        ("ctrl+c", "quit", "Sair"),
    ]

    def __init__(self, graph: CompiledStateGraph[State, Context, State, State]) -> None:
        """Initialize with a pre-built graph (dependency injection)."""
        super().__init__()
        self.graph = graph

    def compose(self) -> ComposeResult:
        """Compose the main layout."""
        yield Header()
        yield ChatHeader()
        yield ChatContainer(self.graph)
        yield Footer()


def _format_message_for_audit(msg: BaseMessage, index: int) -> str:
    """Format a single message for markdown audit output.

    Args:
        msg: BaseMessage to format
        index: Message index (1-based)

    Returns:
        Markdown formatted string for the message
    """
    msg_type = type(msg).__name__
    content = msg.content if isinstance(msg.content, str) else str(msg.content)

    # Truncate long content for display
    display_content = (
        content if len(content) <= 500 else content[:500] + "\n... (truncated)"
    )

    if isinstance(msg, HumanMessage):
        return f"### [{index}] 👤 Human Message\n\n```\n{display_content}\n```\n"
    elif isinstance(msg, AIMessage):
        tool_calls = getattr(msg, "tool_calls", None)
        tools_info = ""
        if tool_calls:
            tool_names = [t.get("name", "unknown") for t in tool_calls]
            tools_info = f"\n**Tools Called:** {', '.join(tool_names)}\n"

        return (
            f"### [{index}] 🤖 Vectora AI Message\n"
            f"{tools_info}\n"
            f"```\n{display_content}\n```\n"
        )
    elif isinstance(msg, ToolMessage):
        tool_name = getattr(msg, "name", "unknown")
        return (
            f"### [{index}] 🔧 Tool Result ({tool_name})\n\n"
            f"```\n{display_content}\n```\n"
        )
    else:
        # Fallback for SystemMessage or other types
        return f"### [{index}] ⚙️ {msg_type}\n\n```\n{display_content}\n```\n"


async def _export_message_audit(
    graph: CompiledStateGraph[State, Context, State, State],
    checkpointer: Checkpointer,
) -> None:
    """Export and display full message audit for verification.

    Extracts all messages from the latest checkpoint and formats them
    with Rich markdown for both terminal display and file persistence.
    Saves complete audit to ~/.vectora/logs/session_audit.md
    """
    try:
        # Get all thread IDs from checkpointer
        all_tuples = list(checkpointer._get_all_history())

        if not all_tuples:
            logger.info("No messages to audit")
            return

        # Get latest checkpoint
        latest_thread_id = all_tuples[-1][0]
        config = RunnableConfig(configurable={"thread_id": latest_thread_id})

        # Extract state
        state_snapshot = await graph.aget_state(config)
        messages: Sequence[BaseMessage] = state_snapshot.values.get("messages", [])

        if not messages:
            logger.info("No messages in final state")
            return

        # Create console for rich output
        console = Console()

        # Format title
        session_timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        title_text = (
            f"📋 MESSAGE AUDIT - Vectora Chat Session\n"
            f"Thread ID: {latest_thread_id} | Timestamp: {session_timestamp}"
        )
        title_panel = Panel(
            title_text,
            style="bold blue",
            expand=False,
        )

        # Display title in terminal
        console.print(title_panel)
        console.print()

        # Build markdown content for both display and file storage
        markdown_content = f"""# 📋 MESSAGE AUDIT - Vectora Chat Session

**Session Details:**
- Thread ID: {latest_thread_id}
- Timestamp: {session_timestamp}
- Total Messages: {len(messages)}

---

## Conversation History

"""

        # Format and display each message
        for i, msg in enumerate(messages, 1):
            formatted = _format_message_for_audit(msg, i)
            markdown_content += formatted + "\n"

            # Display in terminal with Rich
            msg_markdown = Markdown(formatted)
            console.print(msg_markdown)

        # Add footer with statistics
        footer_text = f"\n✓ **Total Messages Audited:** {len(messages)}\n"
        markdown_content += f"\n---\n\n## Summary\n\n{footer_text}"

        # Display footer
        footer_panel = Panel(
            f"✓ Total Messages: {len(messages)}",
            style="bold green",
            expand=False,
        )
        console.print(footer_panel)
        console.print()

        # Ensure logs directory exists
        LOGS_DIR.mkdir(parents=True, exist_ok=True)

        # Save to persistent audit file
        audit_file = LOGS_DIR / "session_audit.md"
        audit_file.write_text(markdown_content, encoding="utf-8")

        logger.info(
            "Message audit exported",
            extra={
                "audit_file": str(audit_file),
                "message_count": len(messages),
                "thread_id": latest_thread_id,
            },
        )

        # Inform user where audit was saved
        audit_info = (
            f"\n📄 Full audit saved to: {audit_file}\n"
            f"   (Use `cat {audit_file}` to view offline)\n"
        )
        console.print(audit_info, style="dim yellow")

    except Exception as e:
        logger.warning(
            f"Could not export message audit: {e}",
            exc_info=True,
        )


async def run_chat() -> None:
    """Inicializa recursos assíncronos e executa o chat TUI.

    O checkpointer e o grafo são construídos aqui, fora do Textual,
    no mesmo event loop que o App irá usar via `run_async()`. Isso evita
    deadlocks causados por `async with` dentro de `on_mount`.

    Também inicializa o BackgroundEmbeddingWorker que processa a fila
    de embeddings de forma assíncrona e não-bloqueante.
    """
    setup_logging()
    logger.info("Vectora Chat TUI started")

    async with async_lifespan(), Checkpointer(DB_DSN) as checkpointer:
        logger.info("Checkpointer SQLite initialized")
        graph = build_graph(checkpointer)
        logger.info("Graph compiled, starting TUI")

        # Inicializar e iniciar background worker para embeddings fire-and-forget
        worker = await get_background_worker()
        await worker.start()
        logger.info("BackgroundEmbeddingWorker iniciado")

        app = VectoraChatApp(graph)
        try:
            await app.run_async()
        finally:
            # Graceful shutdown do worker
            await worker.stop(timeout_seconds=30)

            # Audit: Extract and display full message history
            await _export_message_audit(graph, checkpointer)

            logger.info("Vectora Chat TUI ended")


if __name__ == "__main__":
    asyncio.run(run_chat())
