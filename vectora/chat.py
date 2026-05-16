"""Rich CLI Chat Interface for Vectora - "Rich Gorda" Dashboard.

Professional agent dashboard using Rich components for real-time rendering.
Features advanced layout, status indicators, and audit trails.

Features:
    - Three-panel dashboard (header, body, footer)
    - Real-time status updates for background workers
    - Markdown rendering for AI responses
    - Live audit panel with conversation history
    - Conversation export to markdown files
    - Professional error and success panels
    - Debug Mode with real-time log monitoring (God-Mode dashboard)
"""

import asyncio
import logging
import os
import warnings
from pathlib import Path
from queue import Queue
from typing import TYPE_CHECKING

# Suppress UserWarnings from external libraries in quiet mode
if os.getenv("QUIET_MODE", "true").lower() == "true":
    warnings.filterwarnings("ignore", category=UserWarning)

from background_worker import get_background_worker
from checkpointer import Checkpointer
from commands import handle_command
from constants import DB_DSN
from context import Context
from graph import build_graph
from ui import (
    AuditPanel,
    ChatMessage,
    LogPanel,
    SeparatorLine,
    SuccessPanel,
    VectoraLayout,
    VectoraStatusPanel,
    WelcomeScreen,
)
from version import __version__

if TYPE_CHECKING:
    from collections.abc import Sequence
from langchain_core.messages import AIMessage, BaseMessage, HumanMessage
from langgraph.graph.state import CompiledStateGraph, RunnableConfig
from log_setup import setup_logging, setup_queue_handler
from rich.console import Console
from rich.live import Live
from rich.panel import Panel
from rich.prompt import Prompt
from state import State
from utils import async_lifespan

logger = logging.getLogger(__name__)


class SafeConsole(Console):
    """Console wrapper that gracefully handles Unicode encoding errors on Windows."""

    def print(self, *args, **kwargs):
        """Print with fallback to plain text if encoding fails."""
        try:
            super().print(*args, **kwargs)
        except UnicodeEncodeError:
            # Fallback: print plain text without rich formatting using UTF-8 binary output
            if args:
                import re
                import sys
                from io import StringIO

                # Try to render the renderable to plain text
                try:
                    # Create a temporary non-file console to extract text
                    temp_buffer = StringIO()
                    temp_console = Console(
                        file=temp_buffer,
                        force_terminal=False,
                        width=120,
                    )
                    temp_console.print(args[0], **kwargs)
                    text = temp_buffer.getvalue().rstrip("\n")
                except Exception:
                    # If that fails, just convert to string
                    text = str(args[0])

                # Remove rich markup tags for cleaner output
                text = re.sub(r"\[/?[^\]]*\]", "", text)
                # Write directly to stdout buffer with UTF-8 encoding
                # This bypasses the console's cp1252 encoding on Windows
                if text:
                    sys.stdout.buffer.write(
                        (text + "\n").encode("utf-8", errors="replace")
                    )


# Configure console for Windows compatibility
import sys

console = SafeConsole()


async def _export_audit(
    audit_panel: AuditPanel,
) -> None:
    """Display and save final message audit with rich formatting."""
    try:
        # Clear terminal
        import os

        os.system("cls" if os.name == "nt" else "clear")

        try:
            console.print("\n")
            console.print(SeparatorLine.render("[LIST] SESSION AUDIT"))

            # Display audit messages using chat format
            for msg in audit_panel.messages:
                console.print(msg.to_panel())

            # Save to file
            audit_file = audit_panel.save_to_file()
            console.print(
                f"\n[green][OK] Audit saved to[/green] [dim]{audit_file}[/dim]"
            )
        except UnicodeEncodeError:
            # Fallback for Windows encoding issues - print without rich formatting
            print("\n=== SESSION AUDIT ===\n")
            for msg in audit_panel.messages:
                print(f"[{msg.role}]")
                print(msg.content)
                print("-" * 40)
            # Save to file
            audit_file = audit_panel.save_to_file()
            print(f"\nAudit saved to {audit_file}")

    except Exception as e:
        logger.warning(f"Audit failed: {e}")


async def _load_prior_messages(
    graph: CompiledStateGraph[State, Context, State, State],
    context: Context,
    audit: AuditPanel,
) -> int:
    """Load prior messages from checkpointer into audit and display them."""
    config = RunnableConfig(
        configurable={
            "thread_id": context.thread_id,
            "context": context,
        }
    )
    try:
        state = await graph.aget_state(config)
        prior_messages = state.values.get("messages", [])
        for msg in prior_messages:
            role = "User" if isinstance(msg, HumanMessage) else "Vectora"
            audit.add_message(role, msg.content)
            # Display the message to the user
            console.print(ChatMessage(role, msg.content).to_panel())
        return len(prior_messages)
    except Exception as e:
        logger.warning(f"Could not load prior messages: {e}")
        return 0


async def _process_user_turn(
    user_input: str,
    graph: CompiledStateGraph[State, Context, State, State],
    config: RunnableConfig,
    audit: AuditPanel,
    status_panel: VectoraStatusPanel,
) -> str:
    """Process a single user turn and return AI response."""
    audit.add_message("User", user_input)
    console.print(ChatMessage("User", user_input).to_panel())

    response_content = ""
    with status_panel.thinking("Processing your message..."):
        async for event in graph.astream_events(
            {"messages": [HumanMessage(user_input)]},
            config=config,
            version="v2",
        ):
            if event.get("event") == "on_chat_model_stream":
                chunk = event.get("data", {}).get("chunk")
                if chunk and hasattr(chunk, "content") and chunk.content:
                    # Handle content as string or list of content blocks
                    content = chunk.content
                    if isinstance(content, list):
                        # Extract text from content blocks
                        for item in content:
                            if isinstance(item, dict) and "text" in item:
                                response_content += item["text"]
                            elif isinstance(item, str):
                                response_content += item
                            else:
                                response_content += str(item)
                    elif isinstance(content, dict) and "text" in content:
                        response_content += content["text"]
                    else:
                        response_content += str(content)

    if response_content:
        console.print(ChatMessage("Vectora", response_content).to_panel())
        audit.add_message("Vectora", response_content)

    return response_content


async def chat_loop(
    graph: CompiledStateGraph[State, Context, State, State],
    checkpointer: Checkpointer,
    context: Context,
    provider: str = "unset",
) -> None:
    """Rich Gorda chat loop with dashboard layout and live rendering."""
    # Initialize Debug Mode (load from persistent config)
    from commands import _load_debug_config

    debug_mode = _load_debug_config()
    log_queue: Queue | None = None
    log_panel_obj: LogPanel | None = None

    def _setup_debug_mode() -> None:
        """Set up debug mode components (queue and panel)."""
        nonlocal log_queue, log_panel_obj
        log_queue = Queue()
        setup_queue_handler(log_queue)
        log_panel_obj = LogPanel(log_queue, max_lines=15)
        logger.info("🔧 Debug Mode enabled - God-Mode dashboard active")

    def _teardown_debug_mode() -> None:
        """Clean up debug mode components."""
        nonlocal log_queue, log_panel_obj
        if log_queue is not None:
            # Drain the queue to prevent lingering handler
            try:
                while not log_queue.empty():
                    log_queue.get_nowait()
            except Exception:
                pass
        log_queue = None
        log_panel_obj = None
        logger.info("Debug Mode disabled")

    # Initialize dashboard
    layout = VectoraLayout()

    # Set up debug mode if enabled
    if debug_mode:
        _setup_debug_mode()
        layout.split_with_debug(log_queue)

    status_panel = VectoraStatusPanel(console)
    audit = AuditPanel(max_visible=3)
    # Injetar contexto no configurable para que os nós possam acessá-lo
    config = RunnableConfig(
        configurable={
            "thread_id": context.thread_id,
            "context": context,
        }
    )
    # Track current thread_id to detect session changes
    current_thread_id = context.thread_id

    # Load prior messages
    message_count = await _load_prior_messages(graph, context, audit)

    # Update header and body based on mode
    if debug_mode:
        main_layout = layout.get_main_layout()
        main_layout["header"].update(
            Panel(
                f"[bold cyan][ROCKET] Vectora v{__version__}[/bold cyan] | "
                f"[yellow]Provider: {provider}[/yellow] | "
                f"[magenta]Thread: {context.thread_id}[/magenta] | "
                f"[green]Messages: {message_count}[/green] | "
                f"[cyan]🔧 DEBUG MODE[/cyan]",
                style="blue",
                expand=False,
            )
        )
        main_layout["body"].update(WelcomeScreen.render(provider=provider))
        main_layout["footer"].update(
            Panel(
                f"[green]*[/green] Background Worker | "
                f"[cyan]Embedding Queue: 0[/cyan] | "
                f"[yellow]RAG: Ready[/yellow]",
                style="dim",
                expand=False,
            )
        )
    else:
        layout.update_header(provider=provider, message_count=message_count)
        layout.update_body(WelcomeScreen.render(provider=provider))
        layout.update_footer()

    console.print(layout.render())
    console.print()

    # Display prior messages on screen
    if message_count > 0:
        for msg in audit.messages:
            console.print(msg.to_panel())

    # Main chat loop
    prompt = Prompt()
    Prompt.prompt_suffix = ""

    while True:
        try:
            user_input = prompt.ask("[bold cyan]You: [/bold cyan]")

            # Handle system commands (/, /model, /help, etc)
            if user_input.startswith("/"):
                should_exit, context, debug_mode = await handle_command(
                    user_input, config, console, context, debug_mode
                )
                if should_exit:
                    console.print("\n[yellow][WAVE] Goodbye![/yellow]")
                    break

                # Handle debug mode changes
                old_debug_mode = log_queue is not None
                if debug_mode and not old_debug_mode:
                    # Debug mode was enabled
                    _setup_debug_mode()
                    layout.split_with_debug(log_queue)
                elif not debug_mode and old_debug_mode:
                    # Debug mode was disabled
                    _teardown_debug_mode()
                    layout = VectoraLayout()
                    layout.update_header(
                        provider=provider, message_count=len(audit.messages)
                    )
                    layout.update_body(audit.render())
                    layout.update_footer(embedding_queue=0, worker_active=True)
                    console.print(layout.render())

                # If context changed (new session), reset audit and update config
                if context.thread_id != current_thread_id:
                    old_thread_id = current_thread_id
                    current_thread_id = context.thread_id
                    audit = AuditPanel(max_visible=3)
                    config = RunnableConfig(
                        configurable={
                            "thread_id": context.thread_id,
                            "context": context,
                        }
                    )
                    console.print(
                        SuccessPanel.render(
                            f"Switched to session {context.thread_id} "
                            f"(from {old_thread_id})",
                            title="Session Switched",
                        )
                    )
                    logger.info(
                        f"Session switched: {old_thread_id} → {context.thread_id}"
                    )
                continue

            if not user_input.strip():
                continue

            # Process turn
            await _process_user_turn(user_input, graph, config, audit, status_panel)

            # Update display
            if debug_mode:
                main_layout = layout.get_main_layout()
                main_layout["header"].update(
                    Panel(
                        f"[bold cyan][ROCKET] Vectora v{__version__}[/bold cyan] | "
                        f"[yellow]Provider: {provider}[/yellow] | "
                        f"[magenta]Thread: {context.thread_id}[/magenta] | "
                        f"[green]Messages: {len(audit.messages)}[/green] | "
                        f"[cyan]🔧 DEBUG MODE[/cyan]",
                        style="blue",
                        expand=False,
                    )
                )
                main_layout["body"].update(audit.render())
                main_layout["footer"].update(
                    Panel(
                        f"[green]*[/green] Background Worker | "
                        f"[cyan]Embedding Queue: 0[/cyan] | "
                        f"[yellow]RAG: Ready[/yellow]",
                        style="dim",
                        expand=False,
                    )
                )
                # Update debug panel with latest logs
                if log_panel_obj:
                    layout.update_debug_panel(log_panel_obj.render())
                console.print(layout.render())
            else:
                layout.update_header(
                    provider=provider, message_count=len(audit.messages)
                )
                layout.update_footer(embedding_queue=0, worker_active=True)
                console.print(SeparatorLine.render())

        except KeyboardInterrupt:
            logger.info("Chat interrupted by user")
            console.print("\n[yellow][!] Chat interrupted[/yellow]")
            break
        except Exception as e:
            logger.exception("Chat error")
            console.print(
                Panel(
                    f"[red]{e!s}[/red]",
                    title="[bold red][X] Error[/bold red]",
                    style="red",
                )
            )

    # Export audit on exit
    await _export_audit(audit)


async def run_chat() -> None:
    """Initialize and run chat with Rich Gorda dashboard."""
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
        Checkpointer(DB_DSN) as checkpointer,
    ):
        # Build graph and start worker
        graph = build_graph(checkpointer)
        worker = await get_background_worker()
        await worker.start()

        console.print("[green][*][/green] System initialized successfully\n")

        # Get LLM provider from config
        provider = "unset"
        try:
            from config import Config

            config_instance = Config.instance()
            provider = config_instance.get_llm_provider() or "unset"
        except Exception as e:
            logger.warning(f"Could not load provider config: {e}")

        context = Context(user_type="default", thread_id=1)

        try:
            await chat_loop(graph, checkpointer, context, provider=provider)
        except Exception as e:
            error_panel = Panel(
                f"[red]Critical error: {e!s}[/red]",
                title="[bold red][X] Fatal Error[/bold red]",
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
