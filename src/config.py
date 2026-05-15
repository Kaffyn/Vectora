"""Configuration Management for Vectora Application.

Singleton pattern for managing environment variables, API keys, and system settings.
Supports hierarchical config loading: defaults.env (public) → .env (secrets).
"""

import os
from pathlib import Path
from typing import Any

from dotenv import dotenv_values, find_dotenv, load_dotenv


class Config:
    """Gerencia configuração do Vectora com hierarquia: defaults.env → .env.

    Padrão:
    1. Carrega defaults.env (comportamento padrão, commitado no Git)
    2. Sobrescreve com .env (segredos e overrides locais, gitignored)

    Isso garante reprodutibilidade (QAs usam mesmos defaults) e segurança
    (secrets não vazam).
    """

    _instance: "Config | None" = None
    _env_path: Path
    _defaults_path: Path

    def __init__(self) -> None:
        """Inicializa gerenciador de configuração com hierarquia."""
        # 1. Localiza e carrega defaults.env (público)
        self._defaults_path = Path.cwd() / "defaults.env"
        if self._defaults_path.exists():
            load_dotenv(self._defaults_path)

        # 2. Localiza e carrega .env (segredos/overrides, sobrescreve defaults)
        env_file = find_dotenv()
        self._env_path = Path(env_file) if env_file else Path.cwd() / ".env"
        if self._env_path.exists():
            load_dotenv(self._env_path, override=True)

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
