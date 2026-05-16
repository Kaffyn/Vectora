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
from pathlib import Path
from typing import TYPE_CHECKING

from background_worker import get_background_worker

if TYPE_CHECKING:
    from collections.abc import Sequence
from checkpointer import Checkpointer
from context import Context
from graph import build_graph
from langchain_core.messages import AIMessage, BaseMessage, HumanMessage
from langgraph.graph.state import CompiledStateGraph, RunnableConfig
from log_setup import setup_logging
from rich.console import Console
from rich.markdown import Markdown
from rich.panel import Panel
from rich.prompt import Prompt
from rich.rule import Rule
from rich.table import Table
from state import State
from utils import async_lifespan

logger = logging.getLogger(__name__)
console = Console()


async def _export_audit(
    graph: CompiledStateGraph[State, Context, State, State],
    checkpointer: Checkpointer,
) -> None:
    """Display and save final message audit with rich formatting."""
    try:
        all_tuples = list(checkpointer._get_all_history())
        if not all_tuples:
            return

        latest_thread_id = all_tuples[-1][0]
        config = RunnableConfig(configurable={"thread_id": latest_thread_id})
        state_snapshot = await graph.aget_state(config)
        messages: Sequence[BaseMessage] = state_snapshot.values.get("messages", [])

        if not messages:
            return

        console.print("\n")
        console.print(Rule("[bold blue]📋 SESSION AUDIT[/bold blue]", style="blue"))

        # Create rich table for audit
        table = Table(
            show_header=True, header_style="bold cyan", title="Message History"
        )
        table.add_column("#", style="dim", width=4)
        table.add_column("Role", style="bold")
        table.add_column("Content", style="white")

        for i, msg in enumerate(messages, 1):
            role = "👤 User" if isinstance(msg, HumanMessage) else "🤖 Vectora"
            content = (
                msg.content[:100] + "..." if len(msg.content) > 100 else msg.content
            )
            table.add_row(str(i), role, content)

        console.print(table)

        # Summary stats
        human_count = sum(1 for m in messages if isinstance(m, HumanMessage))
        ai_count = sum(1 for m in messages if isinstance(m, AIMessage))

        stats_panel = Panel(
            f"[green]✓[/green] Total Messages: [bold]{len(messages)}[/bold]\n"
            f"[cyan]User Messages: {human_count}[/cyan] | "
            f"[magenta]AI Messages: {ai_count}[/magenta]",
            title="[bold]Statistics[/bold]",
            style="green",
        )
        console.print(stats_panel)

        # Save to file
        log_dir = Path.home() / ".vectora" / "logs"
        log_dir.mkdir(parents=True, exist_ok=True)
        audit_file = log_dir / "session_audit.md"

        with audit_file.open("w", encoding="utf-8") as f:
            f.write(f"# Session Audit - {len(messages)} messages\n\n")
            for i, msg in enumerate(messages, 1):
                role = "**User**" if isinstance(msg, HumanMessage) else "**Vectora**"
                f.write(f"## [{i}] {role}\n\n{msg.content}\n\n---\n\n")

        console.print(f"[dim]Audit saved to {audit_file}[/dim]")

    except Exception as e:
        logger.warning(f"Audit failed: {e}")


async def chat_loop(
    graph: CompiledStateGraph[State, Context, State, State],
    checkpointer: Checkpointer,
    context: Context,
) -> None:
    """Enhanced chat loop with Rich formatting and interactive features."""
    # Header with title and info
    header_panel = Panel(
        "[bold cyan]🚀 Vectora Chat[/bold cyan]\n"
        "[dim]Type /sair, /quit, or /q to exit | Ctrl+C to interrupt[/dim]",
        style="bold blue",
        expand=False,
    )
    console.print(header_panel)

    # Load and display prior history
    config = RunnableConfig(configurable={"thread_id": context.thread_id})
    message_count = 0
    try:
        state = await graph.aget_state(config)
        prior_messages = state.values.get("messages", [])
        if prior_messages:
            message_count = len(prior_messages)
            console.print(
                f"[bold green]✓[/bold green] Loaded [cyan]{message_count}[/cyan] prior messages from session\n"
            )
    except Exception:
        pass

    # Main chat loop
    prompt = Prompt()
    Prompt.prompt_suffix = ""
    turn_count = 0

    while True:
        try:
            turn_count += 1
            # User input with rich formatting
            user_input = prompt.ask("[bold cyan]You[/bold cyan]")

            # Exit commands
            if user_input.lower() in ["/sair", "/quit", "/q"]:
                logger.info("Chat ended by user")
                console.print("[yellow]Goodbye! 👋[/yellow]")
                break

            if not user_input.strip():
                continue

            # Display user message in a panel
            console.print(
                Panel(
                    user_input,
                    title="[bold cyan]📝 You[/bold cyan]",
                    style="cyan",
                    expand=False,
                )
            )

            # Stream response with rich formatting
            console.print("[bold magenta]🤖 Vectora[/bold magenta]", end=" ")
            response_content = ""

            async for event in graph.astream_events(
                {"messages": [HumanMessage(user_input)]},
                config=config,
                version="v2",
            ):
                if event.get("event") == "on_chat_model_stream":
                    chunk = event.get("data", {}).get("chunk")
                    if chunk and hasattr(chunk, "content") and chunk.content:
                        console.print(chunk.content, end="", highlight=False)
                        response_content += chunk.content

            console.print()  # newline after response
            console.print(Rule(style="dim"), end="")

        except KeyboardInterrupt:
            logger.info("Chat interrupted by user")
            console.print("\n[yellow]⚠️  Chat interrupted[/yellow]")
            break
        except Exception as e:
            error_panel = Panel(
                f"[red]{e!s}[/red]",
                title="[bold red]❌ Error[/bold red]",
                style="red",
                expand=False,
            )
            console.print(error_panel)
            logger.exception("Chat error")

    # Audit on exit
    await _export_audit(graph, checkpointer)


async def run_chat() -> None:
    """Initialize and run chat with rich startup display."""
    setup_logging()
    logger.info("Chat started")

    # Display startup info
    startup_panel = Panel(
        "[bold cyan]Initializing Vectora...[/bold cyan]\n"
        "[dim]Loading graph, checkpointer, and background worker[/dim]",
        style="blue",
        expand=False,
    )
    console.print(startup_panel)

    async with (
        async_lifespan(),
        Checkpointer("sqlite+aiosqlite:///~/.vectora/data/vectora.db") as checkpointer,
    ):
        graph = build_graph(checkpointer)
        worker = await get_background_worker()
        await worker.start()

        console.print("[green]✓[/green] System initialized successfully\n")

        context = Context(user_type="default", thread_id=1)

        try:
            await chat_loop(graph, checkpointer, context)
        except Exception as e:
            error_panel = Panel(
                f"[red]Critical error: {e!s}[/red]",
                title="[bold red]❌ Fatal Error[/bold red]",
                style="red",
            )
            console.print(error_panel)
            logger.exception("Critical chat error")
        finally:
            console.print("\n[dim]Shutting down background worker...[/dim]")
            await worker.stop(timeout_seconds=30)
            logger.info("Chat ended")


if __name__ == "__main__":
    asyncio.run(run_chat())
