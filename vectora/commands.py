"""Command Dispatcher for Vectora Chat.

Handles system commands (e.g., /quit, /model) separately from chat input.
Provides clean separation between user input and control flow.
"""

import logging
from typing import Any

from config import Config
from rich.panel import Panel
from rich.table import Table
from ui import SuccessPanel

logger = logging.getLogger(__name__)


# Available models by LLM provider
AVAILABLE_MODELS = {
    "google-genai": [
        "gemini-3.1-flash-lite",
        "gemini-3.1-flash-lite-preview",
        "gemini-3.1-flash-image-preview",
        "gemini-3.1-pro-preview",
        "gemini-3-flash-preview",
        "gemini-3-pro-image-preview",
    ],
    "openai": [
        "gpt-5.5",
        "gpt-5.4",
        "gpt-5.4-mini",
        "gpt-5.3-codex",
    ],
    "anthropic": [
        "claude-sonnet-4-6",
        "claude-opus-4-6",
        "claude-haiku-4-5-20251001",
    ],
}


def get_available_models(provider: str | None = None) -> dict[str, list[str]]:
    """Get available models for a provider or all providers.

    Args:
        provider: Specific provider to get models for (e.g., "google-genai")
                 If None, returns all providers' models.

    Returns:
        Dictionary mapping provider names to lists of model names.
    """
    if provider:
        return {provider: AVAILABLE_MODELS.get(provider, [])}
    return AVAILABLE_MODELS


async def handle_command(
    user_input: str,
    config: Config,
    console: Any,
    context: Any = None,
) -> tuple[bool, Any]:
    """Process system commands.

    Args:
        user_input: Raw user input (should start with /)
        config: Config instance for reading/writing settings
        console: Rich console for output
        context: Context object that may be modified by commands

    Returns:
        Tuple of (should_exit, updated_context)
    """
    parts = user_input.split(maxsplit=1)
    cmd = parts[0].lower()
    args = parts[1] if len(parts) > 1 else ""

    if cmd == "/quit" or cmd == "/sair" or cmd == "/q":
        logger.info(f"Chat ended by command: {cmd}")
        return True, context

    elif cmd == "/model":
        await _handle_model_command(args, config, console)

    elif cmd == "/help":
        _display_help(console)

    elif cmd == "/new":
        context = await _handle_new_session(context, console)

    elif cmd == "/sessions":
        await _handle_list_sessions(context, console)

    elif cmd == "/session":
        context = await _handle_switch_session(args, context, console)

    else:
        console.print(f"[dim][red]Unknown command:[/red] {cmd}[/dim]")

    return False, context


async def _handle_model_command(
    args: str,
    config: Config,
    console: Any,
) -> None:
    """Handle /model command for listing and switching models.

    Args:
        args: Arguments after /model (empty for list, model name to switch)
        config: Config instance
        console: Rich console for output
    """
    current_provider = config.get_llm_provider()
    available = get_available_models(current_provider)

    if not args.strip():
        # Display available models
        models_list = available.get(current_provider, [])
        table = Table(title=f"Available Models ({current_provider})", style="cyan")
        table.add_column("Model", style="bold green")

        for model in models_list:
            table.add_row(model)

        console.print(Panel(table, style="cyan", expand=False))

    else:
        # Switch to specified model
        new_model = args.strip()

        # Validate model exists
        available_models = available.get(current_provider, [])
        if new_model not in available_models:
            console.print(
                f"[red]Model '{new_model}' not available for {current_provider}[/red]"
            )
            console.print(f"[dim]Available: {', '.join(available_models)}[/dim]")
            return

        # Update config and persist
        try:
            _set_model_for_provider(current_provider, new_model)
            console.print(SuccessPanel.render(f"Model switched to {new_model}"))
            logger.info(f"Model changed to {new_model}")
        except Exception as e:
            console.print(f"[red]Error setting model: {e}[/red]")
            logger.exception("Failed to set model")


def _set_model_for_provider(provider: str, model: str) -> None:
    """Set model for a specific provider in config.

    Args:
        provider: LLM provider (google-genai, openai, anthropic, ollama)
        model: Model name to use

    Raises:
        ValueError: If provider is unknown
    """
    env_var_map = {
        "google-genai": "GOOGLE_MODEL",
        "openai": "OPENAI_MODEL",
        "anthropic": "ANTHROPIC_MODEL",
        "ollama": "OLLAMA_MODEL",
    }

    if provider not in env_var_map:
        raise ValueError(f"Unknown provider: {provider}")

    env_var = env_var_map[provider]

    # Update in-memory config
    config = Config.instance()
    config.set(env_var, model)

    # Also update environment
    import os

    os.environ[env_var] = model

    logger.info(f"Set {env_var}={model}")


async def _handle_new_session(context: Any, console: Any) -> Any:
    """Create a new chat session.

    Args:
        context: Current context object
        console: Rich console for output

    Returns:
        Updated context with new thread_id
    """
    from context import Context

    # Generate new thread_id by finding the max existing and adding 1
    new_thread_id = (context.thread_id or 1) + 1

    # Create new context with the new thread_id
    new_context = Context(user_type="default", thread_id=new_thread_id)

    console.print(
        SuccessPanel.render(
            f"New session created with ID: {new_thread_id}",
            title="New Session",
        )
    )
    logger.info(f"New session created: thread_id={new_thread_id}")

    return new_context


async def _handle_list_sessions(context: Any, console: Any) -> None:
    """List all available sessions.

    Args:
        context: Current context object
        console: Rich console for output
    """
    from checkpointer import Checkpointer
    from constants import DB_DSN

    try:
        async with Checkpointer(DB_DSN) as checkpointer:
            # Get list of all thread_ids that have data
            # For now, we'll create a simple list showing the current session
            # In the future, this could query the checkpointer for all threads
            table = Table(title="Available Sessions", style="cyan")
            table.add_column("Thread ID", style="bold green")
            table.add_column("Status", style="bold cyan")

            # Show current session
            current_marker = "← Current" if context.thread_id else ""
            table.add_row(str(context.thread_id or 1), current_marker)

            # You can add more sessions here by querying the database
            console.print(Panel(table, style="cyan", expand=False))

    except Exception as e:
        console.print(f"[red]Error listing sessions: {e}[/red]")
        logger.exception("Failed to list sessions")


async def _handle_switch_session(args: str, context: Any, console: Any) -> Any:
    """Switch to a different session.

    Args:
        args: Session ID to switch to
        context: Current context object
        console: Rich console for output

    Returns:
        Updated context with the new thread_id
    """
    from context import Context

    if not args.strip():
        console.print("[red]Usage: /session <thread_id>[/red]")
        return context

    try:
        new_thread_id = int(args.strip())
        old_thread_id = context.thread_id

        # Create new context with the specified thread_id
        new_context = Context(user_type="default", thread_id=new_thread_id)

        console.print(
            SuccessPanel.render(
                f"Switched to session {new_thread_id} (from {old_thread_id})",
                title="Session Switched",
            )
        )
        logger.info(f"Session switched: {old_thread_id} → {new_thread_id}")

        return new_context
    except ValueError:
        console.print(f"[red]Invalid session ID: {args.strip()}[/red]")
        return context


def _display_help(console: Any) -> None:
    """Display help for available commands.

    Args:
        console: Rich console for output
    """
    help_text = """
[bold cyan]Available Commands:[/bold cyan]

[bold]/model[/bold]
  List available models for current provider
  Usage: [dim]/model[/dim]

[bold]/model <model_name>[/bold]
  Switch to a different model
  Usage: [dim]/model gemini-2.5-flash[/dim]

[bold]/new[/bold]
  Create a new chat session
  Usage: [dim]/new[/dim]

[bold]/sessions[/bold]
  List all available sessions
  Usage: [dim]/sessions[/dim]

[bold]/session <id>[/bold]
  Switch to a specific session by ID
  Usage: [dim]/session 1[/dim]

[bold]/quit[/bold], [bold]/sair[/bold], [bold]/q[/bold]
  Exit the chat

[bold]/help[/bold]
  Show this help message
"""

    console.print(Panel(help_text, title="Help", style="cyan", expand=False))
