from datetime import datetime
from typing import Any, Self

from langchain_core.messages import AIMessage, BaseMessage, HumanMessage
from langgraph.graph.state import CompiledStateGraph, RunnableConfig
from langgraph.pregel.main import BaseCheckpointSaver
from rich.markdown import Markdown
from rich.panel import Panel
from rich.text import Text
from textual.app import App, ComposeResult, on
from textual.containers import Container
from textual.widgets import Footer, Header, Input, RichLog, Static

import logging
from checkpointer import build_checkpointer_sqlite
from constants import DB_DSN
from context import Context
from graph import build_graph
from state import State
from utils import async_lifespan

logger = logging.getLogger(__name__)


class ChatHeader(Static):
    """Header widget for the chat application."""

    def render(self) -> Panel:
        """Render the header."""
        title = Text("⚡ Vectora Chat", style="bold cyan")
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
    """Container for chat components."""

    messages: list[BaseMessage] = []
    graph: CompiledStateGraph[State, Context, State, State] | None = None
    checkpointer: BaseCheckpointSaver | None = None
    thread_id: int = 1
    context: Context

    def __init__(self) -> None:
        """Initialize the chat container."""
        super().__init__()
        self.context = Context(user_type="plus", thread_id=self.thread_id)

    async def on_mount(self) -> None:
        """Initialize the application when mounted."""
        chat_log = self.query_one(ChatMessages)
        try:
            async with async_lifespan():
                async with build_checkpointer_sqlite(DB_DSN) as checkpointer:
                    self.checkpointer = checkpointer
                    self.graph = build_graph(checkpointer)
                    chat_log.add_system_message(
                        "Bem-vindo ao Vectora! 🤖 Digite suas perguntas abaixo."
                    )
                    self.query_one(ChatInput).focus()
        except Exception:
            logger.exception("Failed to initialize chat", exc_info=True)
            chat_log.add_system_message("Erro ao inicializar: <error>")

    @on(Input.Submitted)
    async def on_submit(self, event: Input.Submitted) -> None:
        """Handle input submission."""
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
            chat_log.add_human_message(user_input)
            chat_input.value = ""

            if self.graph is None or self.checkpointer is None:
                chat_log.add_system_message("Chat não inicializado corretamente")
                return

            config = RunnableConfig(
                configurable={"thread_id": self.thread_id},
            )

            logger.debug(
                "Processing user input",
                extra={"thread_id": self.thread_id, "input_length": len(user_input)},
            )

            human_message = HumanMessage(user_input)
            result = await self.graph.ainvoke(
                {"messages": [human_message]},
                config=config,
                context=self.context,
            )

            last_message = result["messages"][-1]
            model_name = ""

            if isinstance(last_message, AIMessage):
                model_name = last_message.response_metadata.get("model", "")

            chat_log.add_ai_message(last_message.text, model_name)
            self.messages = result["messages"]

            logger.info(
                "User input processed",
                extra={"thread_id": self.thread_id, "model": model_name},
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

    def compose(self) -> ComposeResult:
        """Compose the main layout."""
        yield Header()
        yield ChatHeader()
        yield ChatContainer()
        yield Footer()


def run_chat() -> None:
    """Run the chat application."""
    from log_setup import setup_logging

    setup_logging()
    logger.info("Vectora Chat TUI started")

    app = VectoraChatApp()
    try:
        app.run()
    finally:
        logger.info("Vectora Chat TUI ended")


if __name__ == "__main__":
    run_chat()
