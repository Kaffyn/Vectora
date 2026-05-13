import subprocess
import sys
import webbrowser

from langchain_core.language_models import BaseChatModel
from rich.console import Console
from rich.panel import Panel
from rich.prompt import Confirm, Prompt
from rich.table import Table

from config import Config
from utils import load_llm


class SetupWizard:
    """App Textual para setup interativo do Vectora."""

    def __init__(self) -> None:
        """Inicializa wizard de setup."""
        self.config = Config.instance()
        self.selected_provider: str | None = None
        self.api_key: str | None = None

    def detect_providers(self) -> list[str]:
        """Detecta provedores já configurados."""
        return self.config.get_available_providers()

    def get_provider_info(self, provider: str) -> dict[str, str]:
        """Retorna informações sobre um provedor."""
        providers: dict[str, dict[str, str]] = {
            "google-genai": {
                "name": "Google Gemini",
                "key_name": "GOOGLE_API_KEY",
                "get_key_url": "https://aistudio.google.com/app/apikeys",
                "model_env": "GOOGLE_MODEL",
                "default_model": "gemini-2.0-flash",
            },
            "openai": {
                "name": "OpenAI GPT-4",
                "key_name": "OPENAI_API_KEY",
                "get_key_url": "https://platform.openai.com/api-keys",
                "model_env": "OPENAI_MODEL",
                "default_model": "gpt-4o",
            },
            "anthropic": {
                "name": "Anthropic Claude",
                "key_name": "ANTHROPIC_API_KEY",
                "get_key_url": "https://console.anthropic.com/account/keys",
                "model_env": "ANTHROPIC_MODEL",
                "default_model": "claude-opus-4-1",
            },
            "ollama": {
                "name": "Ollama (Local)",
                "key_name": "OLLAMA_BASE_URL",
                "get_key_url": "https://ollama.ai/download",
                "model_env": "OLLAMA_MODEL",
                "default_model": "llama2",
            },
        }
        return providers.get(provider, {})

    def open_key_url(self, provider: str) -> None:
        """Abre URL para obter chave do provedor."""
        info = self.get_provider_info(provider)
        if "get_key_url" in info:
            webbrowser.open(info["get_key_url"])

    async def test_llm_connection(
        self, provider: str, api_key: str
    ) -> tuple[bool, str]:
        """Testa conexão com o LLM escolhido."""
        try:
            self.config.set(self.get_provider_info(provider)["key_name"], api_key)
            self.config.set("LLM_PROVIDER", provider)

            llm = load_llm()
            if not isinstance(llm, BaseChatModel):
                return False, "LLM carregado mas tipo inválido"

            response = await llm.ainvoke([("user", "teste")])
            if response:
                return True, "✓ Conexão testada com sucesso"
            return False, "LLM não retornou resposta"

        except Exception as e:
            return False, f"✗ Erro: {e!s}"

    def save_config(self, provider: str, api_key: str) -> None:
        """Salva configuração em .env."""
        info = self.get_provider_info(provider)
        data = {
            "LLM_PROVIDER": provider,
            info["key_name"]: api_key,
            info["model_env"]: info["default_model"],
            "LOG_LEVEL": "INFO",
        }
        self.config.save_to_env(data)

    def print_welcome_screen(self) -> None:
        """Imprime tela de boas-vindas no console."""
        console = Console()

        welcome_text = """
[bold cyan]🚀 Bem-vindo ao Vectora[/bold cyan]

Assistente de IA avançado com RAG, manipulação de código e integração de ferramentas.

Este wizard o guiará através da configuração inicial.
"""

        console.print(Panel(welcome_text, border_style="cyan"))

    def print_provider_selection(self) -> None:
        """Imprime opções de provedores."""
        console = Console()

        table = Table(title="Provedores Disponíveis")
        table.add_column("Provedor", style="cyan")
        table.add_column("Configurado", style="green")

        providers = [
            ("google-genai", "Google Gemini"),
            ("openai", "OpenAI GPT-4"),
            ("anthropic", "Anthropic Claude"),
            ("ollama", "Ollama (Local)"),
        ]

        configured = self.detect_providers()

        for provider, name in providers:
            is_configured = "✓" if provider in configured else "✗"
            table.add_row(name, is_configured)

        console.print(
            Panel(table, border_style="green", title="[bold]Selecione um Provedor")
        )

    def run_interactive_setup(self) -> None:
        """Executa setup interativo no console."""
        console = Console()

        self.print_welcome_screen()
        self.print_provider_selection()

        providers_map = {
            "1": "google-genai",
            "2": "openai",
            "3": "anthropic",
            "4": "ollama",
        }

        choice = Prompt.ask(
            "\n[bold]Escolha um provedor[/bold]",
            choices=["1", "2", "3", "4", "s"],
            default="s",
        )

        if choice == "s":
            choice = Prompt.ask("Ou digite o nome do provedor", default="google-genai")
            provider = choice
        else:
            provider = providers_map.get(choice, "google-genai")

        info = self.get_provider_info(provider)
        console.print(f"\n[bold]{info['name']}[/bold]")

        if Confirm.ask(f"Abrir [link]{info['get_key_url']}[/link] para obter a chave?"):
            self.open_key_url(provider)

        api_key = Prompt.ask(
            f"\nDigite sua {info['key_name']}",
            password=True,
        )

        console.print("\n[yellow]⏳ Testando conexão...[/yellow]")

        import asyncio

        success, message = asyncio.run(self.test_llm_connection(provider, api_key))

        if success:
            console.print(f"[green]{message}[/green]")
            self.save_config(provider, api_key)
            console.print(
                f"\n[green]✓ Configuração salva em[/green] [bold]{self.config.env_path}[/bold]"
            )
            console.print("\n[cyan]🎉 Setup concluído! Iniciando Vectora...[/cyan]\n")
            self.start_chat()
        else:
            console.print(f"[red]{message}[/red]")
            console.print(
                "[yellow]Verifique suas credenciais e tente novamente.[/yellow]"
            )

    def start_chat(self) -> None:
        """Inicia o chat Vectora após setup bem-sucedido."""
        try:
            subprocess.run(
                [sys.executable, "src/run_chat.py"],
                check=False,
                shell=False,
            )
        except Exception as e:
            Console().print(f"[red]Erro ao iniciar chat: {e}[/red]")


def main() -> None:
    """Executa o setup wizard."""
    wizard = SetupWizard()
    wizard.run_interactive_setup()


if __name__ == "__main__":
    main()
