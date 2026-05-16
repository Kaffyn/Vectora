"""Setup Wizard for Vectora Configuration.

Simple CLI wizard for configuring LLM provider, API keys, and testing connection.
No Textual TUI - pure linear CLI with Rich formatting.

Features:
    - Provider selection (Google Gemini, OpenAI, Anthropic, Ollama)
    - Secure API key input with getpass
    - Connection testing with async spinner
    - .env file creation
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
        "model": "claude-3-5-sonnet-20241022",
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
    elif provider_id == "openai":
        from langchain_openai import ChatOpenAI

        if not api_key:
            msg = "API key required for OpenAI"
            raise ValueError(msg)
        return ChatOpenAI(
            api_key=api_key,
            model="gpt-4-turbo",
        )
    elif provider_id == "anthropic":
        from langchain_anthropic import ChatAnthropic

        if not api_key:
            msg = "API key required for Anthropic"
            raise ValueError(msg)
        return ChatAnthropic(
            api_key=api_key,
            model="claude-3-5-sonnet-20241022",
        )
    elif provider_id == "ollama":
        from langchain_ollama import ChatOllama

        return ChatOllama(model="mistral")
    else:
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


async def run_setup() -> None:
    """Run the setup wizard flow."""
    console = Console()

    # Welcome
    console.print(Panel("🚀 Vectora Setup Wizard", style="bold cyan"))
    console.print("\nWelcome to Vectora! Let's configure your LLM provider.\n")

    # Provider selection
    console.print("Available providers:")
    for key, info in PROVIDERS.items():
        console.print(f"  {key}) {info['name']}")

    # Get provider choice
    provider_choice = None
    while provider_choice not in PROVIDERS:
        choice_input = input("\nSelect a provider (1-4): ").strip()
        if choice_input in PROVIDERS:
            provider_choice = choice_input
        else:
            console.print("[red]Invalid choice. Please select 1-4.[/red]")

    provider_info = PROVIDERS[provider_choice]
    provider_id = provider_info["provider_id"]
    provider_name = provider_info["name"]

    console.print(f"\n[green]✓ Selected:[/green] {provider_name}\n")

    # Get API key (if needed)
    api_key = None
    if provider_id != "ollama":
        console.print(f"Get your API key from: {provider_info['url']}\n")
        api_key = getpass.getpass(f"Enter {provider_name} API key (hidden): ").strip()

        if not api_key:
            console.print("[red]✗ API key is required for this provider.[/red]")
            sys.exit(1)

    # Test connection
    console.print("\n[bold]Testing connection...[/bold]")

    try:
        llm = _load_llm_for_test(provider_id, api_key)

        # Create a test task
        async def _test_connection() -> str:
            """Test LLM connection asynchronously."""
            return await llm.ainvoke("Say 'Connected!' in one word.")

        # Run test
        with console.status("[bold green]Connecting to LLM...", spinner="dots"):
            response = await _test_connection()

        console.print(f"[green]✓ Connected! Response:[/green] {response.content}\n")

    except Exception as e:
        console.print(f"[red]✗ Connection failed:[/red] {e}")
        logger.error(f"Connection test failed: {e}", exc_info=True)
        sys.exit(1)

    # Save configuration
    _save_to_env(provider_id, api_key)
    console.print("[green]✓ Configuration saved to ~/.vectora/.env[/green]\n")

    # Launch chat
    console.print("[bold]Launching Vectora Chat...[/bold]\n")

    # Import and run chat
    from chat import run_chat

    await run_chat()


def run_setup_sync() -> None:
    """Synchronous entry point for setup wizard."""
    asyncio.run(run_setup())


if __name__ == "__main__":
    run_setup_sync()
