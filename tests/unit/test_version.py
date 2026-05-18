"""Testes para version.py (Vectora version management)."""

from __future__ import annotations

from importlib.metadata import PackageNotFoundError
from unittest.mock import patch

import pytest

from vectora.version import get_vectora_version


class TestGetVectoraVersion:
    """Testes para função get_vectora_version()."""

    def test_get_vectora_version_returns_string(self) -> None:
        """Verificar que get_vectora_version() retorna string."""
        result = get_vectora_version()
        assert isinstance(result, str)

    def test_get_vectora_version_returns_valid_version_format(self) -> None:
        """Verificar que versão tem formato válido."""
        result = get_vectora_version()
        # Deve ter pelo menos 1 caractere
        assert len(result) > 0

    def test_get_vectora_version_when_package_found(self) -> None:
        """Verificar que retorna versão quando pacote existe."""
        with patch("vectora.version.get_version", return_value="0.5.0"):
            result = get_vectora_version()
            assert result == "0.5.0"

    def test_get_vectora_version_when_package_not_found(self) -> None:
        """Verificar que retorna dev version quando pacote não existe."""
        with patch(
            "vectora.version.get_version",
            side_effect=PackageNotFoundError("vectora"),
        ):
            result = get_vectora_version()
            assert result == "0.1.0-dev"

    def test_get_vectora_version_catches_package_not_found_error(self) -> None:
        """Verificar que função trata PackageNotFoundError corretamente."""
        with patch(
            "vectora.version.get_version",
            side_effect=PackageNotFoundError("vectora"),
        ):
            # Não deve lançar exceção
            result = get_vectora_version()
            assert isinstance(result, str)
            assert "dev" in result.lower()

    def test_get_vectora_version_dev_fallback_is_valid_version(self) -> None:
        """Verificar que fallback dev é versão válida."""
        with patch(
            "vectora.version.get_version",
            side_effect=PackageNotFoundError("vectora"),
        ):
            result = get_vectora_version()
            # Deve ser "0.1.0-dev"
            assert "0.1.0" in result
            assert "dev" in result
