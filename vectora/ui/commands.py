"""Command Dispatcher for Vectora Chat.

Handles system commands (e.g., /quit, /model) separately from chat input.
Provides clean separation between user input and control flow.
"""

import json
import logging
from pathlib import Path
from typing import Any

from rich.panel import Panel
from rich.table import Table

from vectora.settings import settings
from vectora.ui.main import SuccessPanel

logger = logging.getLogger(__name__)

# Config file path for persistent settings
CONFIG_FILE = Path.home() / ".vectora" / "chat_config.json"


def _load_debug_config() -> bool:
    """Load debug mode setting from config file.

    Returns:
        Saved debug_mode value or False if not found
    """
    try:
        if CONFIG_FILE.exists():
            data = json.loads(CONFIG_FILE.read_text())
            return data.get("debug_mode", False)
    except Exception as e:
        logger.warning(f"Could not load debug config: {e}")
    return False


def _save_debug_config(debug_mode: bool) -> None:
    """Save debug mode setting to config file.

    Args:
        debug_mode: Debug mode state to persist
    """
    try:
        CONFIG_FILE.parent.mkdir(parents=True, exist_ok=True)
        data = {}
        if CONFIG_FILE.exists():
            data = json.loads(CONFIG_FILE.read_text())
        data["debug_mode"] = debug_mode
        CONFIG_FILE.write_text(json.dumps(data, indent=2))
    except Exception as e:
        logger.warning(f"Could not save debug config: {e}")


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
    config: Any,
    console: Any,
    context: Any = None,
    debug_mode: bool = False,
) -> tuple[bool, Any, bool]:
    """Process system commands.

    Args:
        user_input: Raw user input (should start with /)
        config: Config instance for reading/writing settings (kept for backward compatibility)
        console: Rich console for output
        context: Context object that may be modified by commands
        debug_mode: Current debug mode state

    Returns:
        Tuple of (should_exit, updated_context, debug_mode)
    """
    parts = user_input.split(maxsplit=1)
    cmd = parts[0].lower()
    args = parts[1] if len(parts) > 1 else ""

    if cmd in {"/quit", "/sair", "/q"}:
        logger.info(f"Chat ended by command: {cmd}")
        return True, context, debug_mode

    if cmd == "/model":
        await _handle_model_command(args, console)

    elif cmd == "/help":
        _display_help(console)

    elif cmd == "/debug":
        debug_mode = await _handle_debug_command(args, console, debug_mode)

    elif cmd == "/new":
        context = await _handle_new_session(context, console)

    elif cmd == "/sessions":
        await _handle_list_sessions(context, console)

    elif cmd == "/session":
        context = await _handle_switch_session(args, context, console)

    elif cmd in {"/tools", "/tool"}:
        await _handle_tools_command(console)

    elif cmd == "/list":
        _display_command_list(console)

    else:
        console.print(f"[dim][red]Unknown command:[/red] {cmd}[/dim]")

    return False, context, debug_mode


async def _handle_debug_command(
    args: str,
    console: Any,
    current_debug_mode: bool,
) -> bool:
    """Handle /debug command for toggling debug mode.

    Args:
        args: Arguments after /debug (empty to toggle, "true"/"false" to set)
        console: Rich console for output
        current_debug_mode: Current debug mode state

    Returns:
        New debug mode state
    """
    args = args.strip().lower()

    if not args:
        # Toggle mode
        new_debug_mode = not current_debug_mode
        console.print(
            SuccessPanel.render(
                f"Debug Mode toggled: {new_debug_mode}",
                title="Debug Mode",
            )
        )
        logger.info(f"Debug mode toggled to: {new_debug_mode}")
        _save_debug_config(new_debug_mode)
        return new_debug_mode

    if args in {"true", "on", "yes"}:
        if current_debug_mode:
            console.print("[yellow]Debug Mode is already enabled[/yellow]")
            return current_debug_mode
        console.print(
            SuccessPanel.render(
                "Debug Mode enabled",
                title="Debug Mode",
            )
        )
        logger.info("Debug mode enabled")
        _save_debug_config(True)
        return True

    if args in {"false", "off", "no"}:
        if not current_debug_mode:
            console.print("[yellow]Debug Mode is already disabled[/yellow]")
            return current_debug_mode
        console.print(
            SuccessPanel.render(
                "Debug Mode disabled",
                title="Debug Mode",
            )
        )
        logger.info("Debug mode disabled")
        _save_debug_config(False)
        return False

    console.print(
        f"[red]Invalid argument: {args}[/red]\n"
        "[dim]Usage: /debug [true|false] or /debug to toggle[/dim]"
    )
    return current_debug_mode


async def _handle_model_command(
    args: str,
    console: Any,
) -> None:
    """Handle /model command for listing and switching models.

    Args:
        args: Arguments after /model (empty for list, model name to switch)
        console: Rich console for output
    """
    current_provider = settings.get_llm_provider()
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
    """Set model for a specific provider in settings.

    Args:
        provider: LLM provider (google-genai, openai, anthropic, ollama)
        model: Model name to use

    Raises:
        ValueError: If provider is unknown
    """
    # Update settings directly using set_model method
    try:
        settings.set_model(provider, model)

        # Also update environment
        import os

        env_var_map = {
            "google-genai": "GOOGLE_MODEL",
            "openai": "OPENAI_MODEL",
            "anthropic": "ANTHROPIC_MODEL",
            "ollama": "OLLAMA_MODEL",
        }

        if provider in env_var_map:
            env_var = env_var_map[provider]
            os.environ[env_var] = model
            logger.info(f"Set {env_var}={model}")
    except ValueError as e:
        logger.exception(f"Failed to set model for provider {provider}: {e}")
        raise


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

    # Create new context preserving user_type from previous context
    new_context = Context(
        user_type=context.user_type or "default", thread_id=new_thread_id
    )

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

    try:
        async with Checkpointer(settings.db_dsn) as checkpointer:
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

        # Create new context preserving user_type from current context
        new_context = Context(
            user_type=context.user_type or "default", thread_id=new_thread_id
        )

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


async def _handle_tools_command(console: Any) -> None:
    """List all available tools in Vectora.

    Args:
        console: Rich console for output
    """
    try:
        from tools import TOOLS

        table = Table(title="Available Tools", style="cyan")
        table.add_column("Tool Name", style="bold green")
        table.add_column("Description", style="dim")

        for tool in TOOLS:
            # Get tool description
            description = ""
            if hasattr(tool, "description"):
                description = tool.description
            elif hasattr(tool, "docstring"):
                # Try to get first line of docstring
                doc = tool.docstring or ""
                description = doc.split("\n")[0] if doc else ""

            table.add_row(tool.name, description)

        console.print(Panel(table, style="cyan", expand=False))
        logger.info(f"Tools listed: {len(TOOLS)} available")
    except Exception as e:
        console.print(f"[red]Error listing tools: {e}[/red]")
        logger.exception("Failed to list tools")


def _display_command_list(console: Any) -> None:
    """Display all available commands in Vectora.

    Args:
        console: Rich console for output
    """
    commands_text = """
[bold cyan]Available Commands:[/bold cyan]

[bold]/model[/bold]
  List available models for current provider
  Usage: [dim]/model[/dim]

[bold]/model <model_name>[/bold]
  Switch to a different model
  Usage: [dim]/model gemini-2.5-flash[/dim]

[bold]/debug[/bold]
  Toggle debug mode (shows logs from all components)
  Usage: [dim]/debug[/dim]

[bold]/debug true|false[/bold]
  Enable or disable debug mode
  Usage: [dim]/debug true[/dim] or [dim]/debug false[/dim]

[bold]/tools[/bold] or [bold]/tool[/bold]
  List all available tools
  Usage: [dim]/tools[/dim]

[bold]/new[/bold]
  Create a new chat session
  Usage: [dim]/new[/dim]

[bold]/sessions[/bold]
  List all available sessions
  Usage: [dim]/sessions[/dim]

[bold]/session <id>[/bold]
  Switch to a specific session by ID
  Usage: [dim]/session 1[/dim]

[bold]/list[/bold]
  Show this list of all available commands
  Usage: [dim]/list[/dim]

[bold]/quit[/bold], [bold]/sair[/bold], [bold]/q[/bold]
  Exit the chat

[bold]/help[/bold]
  Show basic help message
"""

    console.print(Panel(commands_text, title="Commands", style="cyan", expand=False))


def _display_help(console: Any) -> None:
    """Display basic help message pointing to /list for full commands.

    Args:
        console: Rich console for output
    """
    help_text = """
[bold cyan]Welcome to Vectora Chat![/bold cyan]

Type your message to chat, or use these commands:

[bold]/list[/bold] - Show all available commands
[bold]/help[/bold] - Show this message
[bold]/quit[/bold] - Exit the chat
[bold]/new[/bold] - Create new session
[bold]/tools[/bold] - List available tools
[bold]/debug[/bold] - Toggle debug mode

For complete command reference, type [bold]/list[/bold]
"""

    console.print(Panel(help_text, title="Help", style="cyan", expand=False))
