"""Configuration Management for Vectora Application.

Singleton pattern for managing environment variables, API keys, and system settings.
Supports hierarchical config loading: defaults.env (public) → .env (secrets).

defaults.env is embedded in the package via hatchling, ensuring it's always
available even when installed globally via `uv tool install vectora`.
"""

import os
import sys
from pathlib import Path
from typing import Any

from dotenv import dotenv_values, find_dotenv, load_dotenv


def _load_defaults_env() -> Path | None:
    """Load defaults.env from package or project root.

    Tries multiple locations in order:
    1. importlib.resources (packaged installation)
    2. Project root (development/local installation)

    Returns path if found, None otherwise.
    """
    # Try importlib.resources first (packaged installation)
    try:
        from importlib.resources import files

        # Try to get defaults.env from package root
        try:
            defaults_data = files("vectora").joinpath("defaults.env")
            # Check if it exists by reading a bit
            defaults_data.read_text()
            return None  # Will load from package below
        except (FileNotFoundError, TypeError):
            pass
    except ImportError:
        pass

    # Fallback: look in project root (development mode)
    project_root = Path.cwd()
    defaults_path = project_root / "defaults.env"
    if defaults_path.exists():
        return defaults_path

    # Also check parent directory (common in monorepos)
    parent_defaults = project_root.parent / "defaults.env"
    if parent_defaults.exists():
        return parent_defaults

    return None


class Config:
    """Gerencia configuração do Vectora com hierarquia: defaults.env → .env.

    Padrão:
    1. Carrega defaults.env (comportamento padrão, commitado no Git)
       - Primeiro tenta importlib.resources (packaged install)
       - Depois tenta project root (development mode)
    2. Sobrescreve com .env (segredos e overrides locais, gitignored)

    Isso garante reprodutibilidade (QAs usam mesmos defaults) e segurança
    (secrets não vazam). Funciona em ambos os modos: desenvolvimento e
    instalação global via `uv tool install`.
    """

    _instance: "Config | None" = None
    _env_path: Path
    _defaults_path: Path | None

    def __init__(self) -> None:
        """Inicializa gerenciador de configuração com hierarquia."""
        # 1. Localiza e carrega defaults.env (público)
        self._defaults_path = _load_defaults_env()
        if self._defaults_path:
            load_dotenv(self._defaults_path)
        else:
            # Tenta carregar from package (importlib.resources style)
            try:
                from importlib.resources import files

                defaults_text = files("vectora").joinpath("defaults.env").read_text()
                # Parse and load manually
                for raw_line in defaults_text.split("\n"):
                    line = raw_line.strip()
                    if line and not line.startswith("#"):
                        if "=" in line:
                            key, value = line.split("=", 1)
                            os.environ[key.strip()] = value.strip()
            except (ImportError, FileNotFoundError, TypeError, AttributeError):
                # Silently continue if defaults.env not found (might be in .env)
                pass

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
