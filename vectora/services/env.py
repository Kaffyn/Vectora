"""Environment Variable Management with Strict Validation.

Provides typed access to environment variables with optional strict mode.
Includes custom exceptions for missing required configuration.
"""

import os
from typing import Literal, overload


class GetEnvError(BaseException): ...


class VoyageAIMissingError(GetEnvError):
    """Erro quando Voyage AI API key não está configurada."""

    def __init__(self) -> None:
        msg = (
            "\n"
            "[X] ERRO: Vectora depende 100% do Voyage AI para funcionar\n"
            "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"
            "Vectora utiliza Voyage AI para:\n"
            "  • Embedding (gerar vetores dos documentos)\n"
            "  • Reranking (ordenar resultados por relevância)\n\n"
            "Sem Voyage AI, Vectora não consegue executar RAG (Retrieval-Augmented Generation).\n\n"
            "SOLUÇÃO:\n"
            "  1. Obtenha uma API key GRATUITA em: https://www.voyageai.com/\n"
            "  2. Configure a variável VOYAGE_API_KEY em seu .env:\n\n"
            "     VOYAGE_API_KEY=pa-...\n\n"
            "  3. Reinicie a aplicação\n\n"
            "O Voyage AI oferece um free tier com excelentes limites para uso.\n"
            "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
        )
        super().__init__(msg)


@overload
def get_env(name: str) -> str: ...


@overload
def get_env(name: str, *, strict: Literal[True]) -> str: ...


@overload
def get_env(name: str, *, strict: Literal[False]) -> str | None: ...


def get_env(name: str, *, strict: bool = True) -> str | None:
    """Get environment variable with optional strict validation."""
    value = os.getenv(name)

    if value is None and strict:
        msg = f"Env variable {name!r} does not exist"
        raise GetEnvError(msg)

    return value


def validate_voyage_ai() -> None:
    """Valida que Voyage AI API key está configurada.

    Vectora depende 100% do Voyage AI para embedding e reranking.
    Sem ele, a aplicação não funciona.

    Raises:
        VoyageAIMissingError: Se VOYAGE_API_KEY não está configurada
    """
    voyage_key = get_env("VOYAGE_API_KEY", strict=False)

    if not voyage_key:
        raise VoyageAIMissingError
