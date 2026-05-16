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
        "claude-3-5-sonnet-20241022",
        "claude-3-opus-20240229",
        "claude-3-haiku-20240307",
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
) -> bool:
    """Process system commands.

    Args:
        user_input: Raw user input (should start with /)
        config: Config instance for reading/writing settings
        console: Rich console for output

    Returns:
        True if chat should exit, False otherwise
    """
    parts = user_input.split(maxsplit=1)
    cmd = parts[0].lower()
    args = parts[1] if len(parts) > 1 else ""

    if cmd == "/quit" or cmd == "/sair" or cmd == "/q":
        logger.info(f"Chat ended by command: {cmd}")
        return True

    elif cmd == "/model":
        await _handle_model_command(args, config, console)

    elif cmd == "/help":
        _display_help(console)

    else:
        console.print(f"[dim][red]Unknown command:[/red] {cmd}[/dim]")

    return False


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

[bold]/quit[/bold], [bold]/sair[/bold], [bold]/q[/bold]
  Exit the chat

[bold]/help[/bold]
  Show this help message
"""

    console.print(Panel(help_text, title="Help", style="cyan", expand=False))
