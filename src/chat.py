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
from datetime import datetime
from typing import Any, Self

from langchain_core.messages import AIMessage, BaseMessage, HumanMessage
from langgraph.graph.state import CompiledStateGraph, RunnableConfig
from rich.markdown import Markdown
from rich.panel import Panel
from rich.text import Text
from textual.app import App, ComposeResult, on
from textual.containers import Container
from textual.widgets import Footer, Header, Input, RichLog, Static

from checkpointer import Checkpointer
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
        """Handle input submission using the Pull Model.

        Fluxo:
        1. Invoca o grafo com a mensagem do usuário (escreve no checkpointer)
        2. Lê o estado atualizado do checkpointer (fonte da verdade)
        3. Re-renderiza o log completo a partir do banco
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

            config = RunnableConfig(
                configurable={"thread_id": self.thread_id},
            )

            logger.debug(
                "Processing user input",
                extra={"thread_id": self.thread_id, "input_length": len(user_input)},
            )

            human_message = HumanMessage(user_input)

            # Passo 1: Invocar o grafo — escreve o resultado no checkpointer (SQLite)
            await self.graph.ainvoke(
                {"messages": [human_message]},
                config=config,
                context=self.context,
            )

            # Passo 2: Pull Model — ler o estado oficial do banco de dados
            current_state = await self.graph.aget_state(config)
            updated_messages: Sequence[BaseMessage] = current_state.values.get(
                "messages", []
            )

            # Passo 3: Re-renderizar o log como uma view do banco
            self._render_messages(chat_log, updated_messages)

            logger.info(
                "User input processed",
                extra={
                    "thread_id": self.thread_id,
                    "total_messages": len(updated_messages),
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


async def run_chat() -> None:
    """Inicializa recursos assíncronos e executa o chat TUI.

    O checkpointer e o grafo são construídos aqui, fora do Textual,
    no mesmo event loop que o App irá usar via `run_async()`. Isso evita
    deadlocks causados por `async with` dentro de `on_mount`.

    Também inicializa o BackgroundEmbeddingWorker que processa a fila
    de embeddings de forma assíncrona e não-bloqueante.
    """
    from log_setup import setup_logging

    from background_worker import get_background_worker

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
            await worker.stop(timeout=30)
            logger.info("Vectora Chat TUI ended")


if __name__ == "__main__":
    asyncio.run(run_chat())
