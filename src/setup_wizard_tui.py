"""Setup Wizard TUI usando Textual para interface profissional."""

import asyncio
import subprocess
import sys

from langchain_core.language_models import BaseChatModel
from textual.app import App, ComposeResult, on
from textual.screen import Screen
from textual.widgets import Button, Input, Label, RadioButton, RadioSet

from config import Config
from utils import load_llm


class WelcomeScreen(Screen):
    """Tela de boas-vindas do Vectora."""

    BINDINGS = [("q", "quit", "Sair")]

    def compose(self) -> ComposeResult:
        """Renderiza a tela de boas-vindas."""
        yield Label("🚀 Bem-vindo ao Vectora")
        yield Label("")
        yield Label("Assistente de IA avançado com RAG,")
        yield Label("manipulação de código e integração de ferramentas.")
        yield Label("")
        yield Button("Iniciar Setup", id="start_btn", variant="primary")
        yield Button("Sair", id="quit_btn", variant="error")

    @on(Button.Pressed, "#start_btn")
    def start_setup(self) -> None:
        """Inicia o setup."""
        self.app.push_screen(ProviderSelectScreen())

    @on(Button.Pressed, "#quit_btn")
    def quit_app(self) -> None:
        """Encerra o app."""
        self.app.exit()


class ProviderSelectScreen(Screen):
    """Tela de seleção de provedor LLM."""

    BINDINGS = [("q", "quit", "Sair")]

    def compose(self) -> ComposeResult:
        """Renderiza seleção de provedores."""
        yield Label("Selecione um provedor LLM:")
        yield Label("")

        config = Config.instance()
        with RadioSet(id="provider_selection"):
            providers = [
                ("google-genai", "🔵 Google Gemini"),
                ("openai", "🟢 OpenAI GPT-4"),
                ("anthropic", "🔴 Anthropic Claude"),
                ("ollama", "🟠 Ollama (Local)"),
            ]

            configured = config.get_available_providers()

            for provider_id, label in providers:
                status = " (✓ Configurado)" if provider_id in configured else ""
                yield RadioButton(f"{label}{status}", id=provider_id, value=provider_id)

        yield Label("")
        yield Button("Continuar", id="continue_btn", variant="primary")
        yield Button("Voltar", id="back_btn")

    @on(Button.Pressed, "#continue_btn")
    def continue_to_api_key(self) -> None:
        """Vai para entrada de API key."""
        radio_set = self.query_one(RadioSet)
        provider = radio_set.pressed_button.value if radio_set.pressed_button else None

        if provider:
            self.app.push_screen(
                ApiKeyScreen(provider),
            )

    @on(Button.Pressed, "#back_btn")
    def go_back(self) -> None:
        """Volta para tela anterior."""
        self.app.pop_screen()


class ApiKeyScreen(Screen):
    """Tela para entrada de API key."""

    BINDINGS = [("q", "quit", "Sair")]

    def __init__(self, provider: str) -> None:
        """Inicializa tela com provedor selecionado."""
        super().__init__()
        self.provider = provider
        self.config = Config.instance()
        self.provider_info = self._get_provider_info(provider)

    def _get_provider_info(self, provider: str) -> dict[str, str]:
        """Retorna informações do provedor."""
        providers = {
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

    def compose(self) -> ComposeResult:
        """Renderiza entrada de API key."""
        yield Label(f"Configurar {self.provider_info['name']}")
        yield Label("")
        yield Label(f"Digite sua {self.provider_info['key_name']}:")

        yield Input(
            id="api_key_input",
            password=True,
            type="password",
        )

        yield Label("")
        yield Label(f"🔗 Obter chave: {self.provider_info['get_key_url']}")
        yield Label("")
        yield Button("Testar e Salvar", id="test_btn", variant="primary")
        yield Button("Voltar", id="back_btn")
        yield Label("", id="status")

    @on(Button.Pressed, "#test_btn")
    def test_and_save(self) -> None:
        """Testa conexão e salva config."""
        api_key_input = self.query_one(Input)
        api_key = api_key_input.value

        if not api_key:
            status_label = self.query_one("#status", Label)
            status_label.update("❌ API key não pode estar vazia")
            return

        status_label = self.query_one("#status", Label)
        status_label.update("⏳ Testando conexão...")

        task = asyncio.ensure_future(self._async_test_and_save(api_key, status_label))
        # task is tracked by event loop

    async def _async_test_and_save(self, api_key: str, status_label: Label) -> None:
        """Testa e salva de forma assíncrona."""
        try:
            self.config.set(self.provider_info["key_name"], api_key)
            self.config.set("LLM_PROVIDER", self.provider)

            llm = load_llm()
            if not isinstance(llm, BaseChatModel):
                status_label.update("❌ LLM tipo inválido")
                return

            response = await llm.ainvoke([("user", "teste")])
            if response:
                self.config.save_to_env(
                    {
                        "LLM_PROVIDER": self.provider,
                        self.provider_info["key_name"]: api_key,
                        self.provider_info["model_env"]: self.provider_info[
                            "default_model"
                        ],
                        "LOG_LEVEL": "INFO",
                    }
                )
                status_label.update("✅ Configuração salva com sucesso!")

                await asyncio.sleep(1)
                self.app.push_screen(CompleteScreen())
            else:
                status_label.update("❌ LLM não retornou resposta")

        except Exception as e:
            status_label.update(f"❌ Erro: {str(e)[:50]}")

    @on(Button.Pressed, "#back_btn")
    def go_back(self) -> None:
        """Volta para tela anterior."""
        self.app.pop_screen()


class CompleteScreen(Screen):
    """Tela de conclusão do setup."""

    BINDINGS: list[tuple[str, str, str]] = []

    def compose(self) -> ComposeResult:
        """Renderiza tela de conclusão."""
        yield Label("🎉 Setup Concluído!")
        yield Label("")
        yield Label("Seu Vectora está pronto para usar.")
        yield Label("")
        yield Button("Iniciar Chat", id="chat_btn", variant="primary")
        yield Button("Sair", id="quit_btn")

    @on(Button.Pressed, "#chat_btn")
    def start_chat(self) -> None:
        """Inicia o chat após setup concluído."""
        self.app.exit(return_code=0)
        subprocess.run(
            [sys.executable, "src/run_chat.py"],
            check=False,
            shell=False,
        )

    @on(Button.Pressed, "#quit_btn")
    def quit_app(self) -> None:
        """Encerra."""
        self.app.exit()


class SetupWizardApp(App):
    """App principal do setup wizard."""

    BINDINGS = [("q", "quit", "Sair")]
    CSS = """
    Screen {
        align: center middle;
    }

    Label {
        margin-bottom: 1;
    }

    Input {
        margin-bottom: 1;
        width: 50;
    }

    Button {
        margin-right: 2;
    }
    """

    def on_mount(self) -> None:
        """Inicia na tela de boas-vindas."""
        self.push_screen(WelcomeScreen())


def run_setup() -> None:
    """Executa o setup wizard Textual."""
    app = SetupWizardApp()
    app.run()


if __name__ == "__main__":
    run_setup()
