"""Rich CLI Chat Interface for Vectora.

Implements a simple command-line chat interface using Rich for formatted output.
No Textual TUI - pure CLI with streaming message rendering.

Features:
    - Real-time message display with Rich formatting
    - Conversation history loading and persistence
    - Message audit with markdown export
    - Background embedding worker integration
    - Simple command-based interface (/sair, /quit, /q to exit)
"""

import asyncio
import logging
from collections.abc import Sequence
from datetime import datetime
from pathlib import Path
from typing import Any

from background_worker import get_background_worker
from checkpointer import Checkpointer
from constants import LOGS_DIR
from context import Context
from graph import build_graph
from langchain_core.messages import AIMessage, BaseMessage, HumanMessage, ToolMessage
from langgraph.graph.state import CompiledStateGraph, RunnableConfig
from log_setup import setup_logging
from rich.console import Console
from rich.markdown import Markdown
from rich.panel import Panel
from state import State
from utils import async_lifespan

logger = logging.getLogger(__name__)


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


def _display_chat_header(console: Console, context: Context) -> None:
    """Display formatted header with session info."""
    title = f"📋 Vectora Chat Session | Thread: {context.thread_id}"
    console.print(Panel(title, style="bold blue"))


async def _load_chat_history(
    graph: CompiledStateGraph[State, Context, State, State],
    checkpointer: Checkpointer,
    context: Context,
) -> Sequence[BaseMessage]:
    """Load prior messages from checkpointer."""
    config = RunnableConfig(configurable={"thread_id": context.thread_id})
    try:
        state_snapshot = await graph.aget_state(config)
        return state_snapshot.values.get("messages", [])
    except Exception as e:
        logger.warning(f"Could not load chat history: {e}")
        return []


def _render_message_history(console: Console, messages: Sequence[BaseMessage]) -> None:
    """Render prior messages from database."""
    if not messages:
        console.print("[dim]No prior messages. Type /help for commands.[/dim]\n")
        return

    console.print(f"[dim]Loaded {len(messages)} prior message(s):[/dim]\n")
    for msg in messages:
        formatted = _format_message_for_audit(msg, 0)
        console.print(Markdown(formatted))


async def chat_loop(
    graph: CompiledStateGraph[State, Context, State, State],
    checkpointer: Checkpointer,
    context: Context,
) -> None:
    """Simple CLI chat loop with Rich display.

    Args:
        graph: Compiled LangGraph instance
        checkpointer: State checkpointer for persistence
        context: Execution context with thread_id
    """
    console = Console()

    # Display header
    _display_chat_header(console, context)

    # Load and render prior messages
    messages = await _load_chat_history(graph, checkpointer, context)
    _render_message_history(console, messages)

    # Main chat loop
    while True:
        try:
            # Read user input
            user_input = input("\n👤 You:\n> ").strip()

            # Handle exit commands
            if user_input.lower() in ["/sair", "/quit", "/q"]:
                logger.info("Chat ended by user")
                break

            # Skip empty input
            if not user_input:
                continue

            # Display user message
            console.print(f"[bold green]{user_input}[/bold green]")

            # Invoke graph and stream response
            config = RunnableConfig(configurable={"thread_id": context.thread_id})
            ai_response = ""

            console.print("[bold cyan]🤖 Vectora:[/bold cyan]")

            async for event in graph.astream_events(
                {"messages": [HumanMessage(user_input)]},
                config=config,
                version="v2",
            ):
                event_type = event.get("event")

                # Stream LLM tokens
                if event_type == "on_chat_model_stream":
                    chunk = event.get("data", {}).get("chunk")
                    if chunk and hasattr(chunk, "content") and chunk.content:
                        console.print(chunk.content, end="", highlight=False)
                        ai_response += chunk.content

            console.print()

            logger.info(
                "User input processed",
                extra={
                    "thread_id": context.thread_id,
                    "response_length": len(ai_response),
                },
            )

        except KeyboardInterrupt:
            logger.info("Chat interrupted by user (Ctrl+C)")
            break
        except Exception as e:
            console.print(f"[red]Error: {e}[/red]")
            logger.exception("Chat error", exc_info=True)

    # Export audit on exit
    await _export_message_audit(graph, checkpointer)


async def run_chat() -> None:
    """Initialize resources and run chat CLI.

    Sets up logging, checkpointer, graph, and background worker,
    then starts the chat loop.
    """
    setup_logging()
    logger.info("Vectora Chat started")

    async with (
        async_lifespan(),
        Checkpointer("sqlite+aiosqlite:///~/.vectora/data/vectora.db") as checkpointer,
    ):
        logger.info("Checkpointer initialized")

        # Build graph
        graph = build_graph(checkpointer)
        logger.info("Graph compiled")

        # Initialize and start background worker
        worker = await get_background_worker()
        await worker.start()
        logger.info("BackgroundEmbeddingWorker started")

        # Create context
        context = Context(user_type="default", thread_id=1)

        try:
            # Run chat loop
            await chat_loop(graph, checkpointer, context)
        finally:
            # Graceful shutdown of worker
            await worker.stop(timeout_seconds=30)
            logger.info("Vectora Chat ended")


if __name__ == "__main__":
    asyncio.run(run_chat())
