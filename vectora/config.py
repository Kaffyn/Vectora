"""Configuration Management for Vectora Application.

Singleton pattern for managing environment variables, API keys, and system settings.
Supports 3-level hierarchical config loading:
1. defaults.env (embedded in package, public)
2. .env (project-local, gitignored)
3. ~/.vectora/.env (user-global, gitignored)

defaults.env is embedded via hatchling, ensuring it's always available even
when installed globally via `uv tool install vectora`.
"""

import os
from importlib import resources
from pathlib import Path
from typing import Any

from dotenv import dotenv_values, find_dotenv, load_dotenv


class Config:
    """Gerencia configuração do Vectora com 3 níveis hierárquicos.

    Níveis de Configuração (em ordem de precedência):
    1. **Level 1 (Defaults - Embarcado):** defaults.env dentro do pacote Python
       - Comportamento padrão, reproducível, commitado no Git
       - Sempre disponível, independente de onde Vectora é instalado

    2. **Level 2 (Local):** .env na raiz do projeto/diretório atual
       - Configurações específicas do projeto
       - Sobrescreve os defaults

    3. **Level 3 (Global/User):** ~/.vectora/.env
       - Preferências globais do usuário
       - Sobrescreve project-local e defaults

    Padrão de instalação profissional (como aws-cli, kubectl):
    - Usuário rodar `uv tool install vectora`
    - Comando `vectora chat` encontra defaults.env (embutido)
    - Usuário customiza via .env local ou ~/.vectora/.env
    """

    _instance: "Config | None" = None
    _env_path: Path
    _defaults_loaded: bool = False

    def __init__(self) -> None:
        """Inicializa gerenciador de configuração com hierarquia 3-níveis."""
        # 1. Carrega defaults.env (embutido no pacote Python)
        self._load_package_defaults()

        # 2. Carrega .env local (projeto atual) - sobrescreve Level 1
        self._load_local_env()

        # 3. Carrega ~/.vectora/.env (preferências globais) - sobrescreve Level 1+2
        self._load_user_env()

        self._env_path = Path.cwd() / ".env"

    def _load_package_defaults(self) -> None:
        """Carrega defaults.env embutido no pacote Python.

        Usa importlib.resources para garantir que funciona em qualquer
        instalação (desenvolvimento, pip install, uv tool install).
        """
        try:
            # importlib.resources: lê arquivo de dentro do pacote instalado
            defaults_env = resources.files("vectora").joinpath("defaults.env")
            defaults_text = defaults_env.read_text(encoding="utf-8")

            # Parse manualmente (sem criar arquivo temporário)
            for raw_line in defaults_text.split("\n"):
                line = raw_line.strip()
                # Ignora comentários e linhas vazias
                if line and not line.startswith("#"):
                    if "=" in line:
                        key, value = line.split("=", 1)
                        # Só define se não estiver já em os.environ
                        os.environ.setdefault(key.strip(), value.strip())

            self._defaults_loaded = True
        except (FileNotFoundError, TypeError, ModuleNotFoundError, AttributeError):
            # Silenciosamente continua se defaults.env não encontrado
            # (pode estar em .env ao invés)
            pass

    def _load_local_env(self) -> None:
        """Carrega .env local (raiz do projeto)."""
        env_path = Path.cwd() / ".env"
        if env_path.exists():
            load_dotenv(env_path, override=True)

    def _load_user_env(self) -> None:
        """Carrega ~/.vectora/.env (preferências globais do usuário)."""
        user_env_path = Path.home() / ".vectora" / ".env"
        if user_env_path.exists():
            load_dotenv(user_env_path, override=True)

    @classmethod
    def instance(cls: type["Config"]) -> "Config":
        """Retorna instância singleton do gerenciador de config."""
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def get(self, key: str, default: Any = None) -> Any:
        """Obtém valor de variável de ambiente ou .env."""
        return os.getenv(key, default)

    def set(self, key: str, value: str) -> None:
        """Define variável de ambiente (mantém em cache até save)."""
        os.environ[key] = value

    def save_to_env(self, data: dict[str, str]) -> None:
        """Salva múltiplas variáveis em arquivo .env."""
        env_dict = (
            dict(dotenv_values(self._env_path)) if self._env_path.exists() else {}
        )
        env_dict.update(data)

        with self._env_path.open("w") as f:
            for key, value in env_dict.items():
                if value is not None:
                    f.write(f"{key}={value}\n")

        load_dotenv(self._env_path, override=True)

    def get_llm_provider(self) -> str | None:
        """Detecta qual provedor LLM está configurado."""
        if self.get("GOOGLE_API_KEY"):
            return "google-genai"
        if self.get("OPENAI_API_KEY"):
            return "openai"
        if self.get("ANTHROPIC_API_KEY"):
            return "anthropic"
        if self.get("OLLAMA_BASE_URL"):
            return "ollama"
        return None

    def get_llm_model(self) -> str | None:
        """Get the configured LLM model for the current provider."""
        provider = self.get_llm_provider()
        if not provider:
            return None

        model_env_map = {
            "google-genai": "GOOGLE_MODEL",
            "openai": "OPENAI_MODEL",
            "anthropic": "ANTHROPIC_MODEL",
            "ollama": "OLLAMA_MODEL",
        }

        env_var = model_env_map.get(provider)
        if env_var:
            return self.get(env_var)
        return None

    def get_available_providers(self) -> list[str]:
        """Retorna lista de provedores com API keys configuradas."""
        available = []
        if self.get("GOOGLE_API_KEY"):
            available.append("google-genai")
        if self.get("OPENAI_API_KEY"):
            available.append("openai")
        if self.get("ANTHROPIC_API_KEY"):
            available.append("anthropic")
        if self.get("OLLAMA_BASE_URL"):
            available.append("ollama")
        return available

    @property
    def env_path(self) -> Path:
        """Retorna caminho do arquivo .env."""
        return self._env_path
