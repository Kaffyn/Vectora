"""Setup Wizard for Vectora Configuration.

Enhanced CLI wizard for configuring LLM provider, API keys, and testing connection.
Rich formatting with provider selection, secure API key input, and connection testing.

Features:
    - Provider selection with table display (Google Gemini, OpenAI, Anthropic, Ollama)
    - Secure API key input with getpass
    - Connection testing with async spinner
    - .env file creation with summary
    - Automatic chat launch after setup
"""

import asyncio
import getpass
import logging
import sys
from pathlib import Path
from typing import Any

from rich.console import Console
from rich.panel import Panel
from rich.rule import Rule
from rich.table import Table

logger = logging.getLogger(__name__)

# Provider configuration
PROVIDERS = {
    "1": {
        "name": "Google Gemini",
        "provider_id": "google-genai",
        "env_var": "GOOGLE_API_KEY",
        "url": "https://ai.google.dev/apikey",
        "model": "gemini-2.0-flash",
    },
    "2": {
        "name": "OpenAI GPT-4",
        "provider_id": "openai",
        "env_var": "OPENAI_API_KEY",
        "url": "https://platform.openai.com/api-keys",
        "model": "gpt-4-turbo",
    },
    "3": {
        "name": "Anthropic Claude",
        "provider_id": "anthropic",
        "env_var": "ANTHROPIC_API_KEY",
        "url": "https://console.anthropic.com/",
        "model": "claude-3-5-sonnet",
    },
    "4": {
        "name": "Ollama (Local)",
        "provider_id": "ollama",
        "env_var": "",
        "url": "https://ollama.ai",
        "model": "mistral",
    },
}


def _load_llm_for_test(provider_id: str, api_key: str | None = None) -> Any:
    """Load LLM instance for connection test.

    Args:
        provider_id: Provider identifier (google-genai, openai, anthropic, ollama)
        api_key: API key for the provider (if applicable)

    Returns:
        LLM instance ready for testing

    Raises:
        ValueError: If provider_id is unknown
        ImportError: If required LangChain package is not installed
    """
    if provider_id == "google-genai":
        from langchain_google_genai import ChatGoogleGenerativeAI

        if not api_key:
            msg = "API key required for Google Gemini"
            raise ValueError(msg)
        return ChatGoogleGenerativeAI(
            api_key=api_key,
            model="gemini-2.0-flash",
        )
    if provider_id == "openai":
        from langchain_openai import ChatOpenAI

        if not api_key:
            msg = "API key required for OpenAI"
            raise ValueError(msg)
        return ChatOpenAI(
            api_key=api_key,
            model="gpt-4-turbo",
        )
    if provider_id == "anthropic":
        from langchain_anthropic import ChatAnthropic

        if not api_key:
            msg = "API key required for Anthropic"
            raise ValueError(msg)
        return ChatAnthropic(
            api_key=api_key,
            model="claude-3-5-sonnet",
        )
    if provider_id == "ollama":
        from langchain_ollama import ChatOllama

        return ChatOllama(model="mistral")
    msg = f"Unknown provider: {provider_id}"
    raise ValueError(msg)


def _save_to_env(provider_id: str, api_key: str | None = None) -> None:
    """Save configuration to ~/.vectora/.env file.

    Args:
        provider_id: Provider identifier
        api_key: API key to save (if applicable)
    """
    env_file = Path.home() / ".vectora" / ".env"
    env_file.parent.mkdir(parents=True, exist_ok=True)

    content = f"LLM_PROVIDER={provider_id}\n"
    if api_key and provider_id != "ollama":
        # Find the provider info to get env_var
        provider_key = next(
            (k for k, v in PROVIDERS.items() if v["provider_id"] == provider_id),
            None,
        )
        if provider_key:
            provider_info = PROVIDERS[provider_key]
            content += f"{provider_info['env_var']}={api_key}\n"

    env_file.write_text(content, encoding="utf-8")
    logger.info("Configuration saved to .env", extra={"file": str(env_file)})


async def _display_welcome() -> None:
    """Display welcome banner and provider table."""
    console = Console()

    welcome_panel = Panel(
        "[bold cyan][ROCKET] Vectora Setup Wizard[/bold cyan]\n"
        "[dim]Configure your LLM provider and test the connection[/dim]",
        style="bold blue",
        expand=False,
    )
    console.print(welcome_panel)
    console.print(
        "\n[cyan]Welcome to Vectora![/cyan] Let's set up your AI assistant.\n"
    )

    console.print("[bold]Available LLM Providers:[/bold]\n")
    table = Table(show_header=True, header_style="bold cyan")
    table.add_column("#", style="dim")
    table.add_column("Provider", style="bold")
    table.add_column("Model", style="yellow")

    for key, info in PROVIDERS.items():
        table.add_row(key, info["name"], info["model"])

    console.print(table)
    console.print()


async def _select_provider() -> tuple[str, dict[str, str]]:
    """Get user's provider selection and return provider info."""
    console = Console()
    provider_choice = None

    while provider_choice not in PROVIDERS:
        choice_input = (
            await asyncio.to_thread(
                input, "\n[bold cyan]Select a provider (1-4):[/bold cyan] "
            )
        ).strip()
        if choice_input in PROVIDERS:
            provider_choice = choice_input
        else:
            console.print("[red][X] Invalid choice. Please select 1-4.[/red]")

    provider_info = PROVIDERS[provider_choice]
    selection_panel = Panel(
        f"[bold cyan]{provider_info['name']}[/bold cyan]\n"
        f"[dim]Model: {provider_info['model']}[/dim]",
        title="[green][OK] Selected[/green]",
        style="green",
        expand=False,
    )
    console.print(selection_panel)
    console.print()

    return provider_choice, provider_info


async def _get_api_key(provider_info: dict[str, str]) -> str:
    """Get and validate API key for the selected provider."""
    console = Console()
    provider_id = provider_info["provider_id"]
    provider_name = provider_info["name"]

    if provider_id == "ollama":
        return ""

    info_panel = Panel(
        f"[cyan]{provider_info['url']}[/cyan]",
        title="[bold]API Key Location[/bold]",
        style="blue",
        expand=False,
    )
    console.print(info_panel)
    console.print()

    api_key = getpass.getpass(
        f"[bold cyan]Enter {provider_name} API key (hidden):[/bold cyan] "
    ).strip()

    if not api_key:
        error_panel = Panel(
            "[red]API key is required for this provider.[/red]",
            title="[bold red][X] Error[/bold red]",
            style="red",
            expand=False,
        )
        console.print(error_panel)
        sys.exit(1)

    return api_key


async def _test_connection(provider_id: str, api_key: str | None) -> None:
    """Test LLM connection and save configuration on success."""
    console = Console()
    console.print("\n[bold]Testing connection...[/bold]\n")

    try:
        llm = _load_llm_for_test(provider_id, api_key)

        async def _test_llm() -> str:
            """Test LLM connection asynchronously."""
            return await llm.ainvoke("Say 'Connected!' in one word.")

        with console.status(
            "[bold cyan]Connecting to LLM... This may take a moment[/bold cyan]",
            spinner="dots",
        ):
            response = await _test_llm()

        success_panel = Panel(
            f"[green][OK] Connection successful![/green]\n"
            f"[cyan]Response: {response.content}[/cyan]",
            title="[bold green]Connection Test[/bold green]",
            style="green",
            expand=False,
        )
        console.print(success_panel)
        console.print()

    except Exception as e:
        error_panel = Panel(
            f"[red]{e!s}[/red]",
            title="[bold red][X] Connection Failed[/bold red]",
            style="red",
        )
        console.print(error_panel)
        logger.exception(f"Connection test failed: {e}")
        sys.exit(1)


async def _finalize_setup(provider_id: str, api_key: str | None) -> None:
    """Save configuration and launch chat."""
    console = Console()
    _save_to_env(provider_id, api_key)

    save_panel = Panel(
        "[green][OK] Configuration saved[/green]\n[dim]Location: ~/.vectora/.env[/dim]",
        title="[bold]Setup Complete[/bold]",
        style="green",
        expand=False,
    )
    console.print(save_panel)
    console.print()

    console.print(Rule("[bold cyan]Launching Vectora Chat[/bold cyan]", style="cyan"))
    console.print()

    from chat import run_chat

    await run_chat()


async def run_setup() -> None:
    """Run the setup wizard flow with enhanced Rich formatting."""
    await _display_welcome()

    _, provider_info = await _select_provider()
    provider_id = provider_info["provider_id"]
    api_key = await _get_api_key(provider_info)
    await _test_connection(provider_id, api_key)
    await _finalize_setup(provider_id, api_key)


def run_setup_sync() -> None:
    """Synchronous entry point for setup wizard."""
    asyncio.run(run_setup())


if __name__ == "__main__":
    run_setup_sync()
