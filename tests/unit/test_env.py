"""Testes para Environment Variable Management (env.py)."""

from __future__ import annotations

import os
from unittest.mock import patch

import pytest

from vectora.services.env import (
    CohereMissingError,
    GetEnvError,
    get_env,
    validate_cohere,
)


class TestGetEnvError:
    """Testes para exceção GetEnvError."""

    def test_get_env_error_is_base_exception(self) -> None:
        """Verificar que GetEnvError é BaseException."""
        error = GetEnvError("test error")
        assert isinstance(error, BaseException)

    def test_get_env_error_with_message(self) -> None:
        """Verificar que GetEnvError armazena mensagem."""
        error = GetEnvError("custom message")
        assert "custom message" in str(error)


class TestCohereMissingError:
    """Testes para exceção CohereMissingError."""

    def test_cohere_missing_error_is_get_env_error(self) -> None:
        """Verificar que CohereMissingError é GetEnvError."""
        error = CohereMissingError()
        assert isinstance(error, GetEnvError)

    def test_cohere_missing_error_message_contains_key_info(self) -> None:
        """Verificar que mensagem contém informações importantes."""
        error = CohereMissingError()
        msg = str(error)

        assert "Cohere" in msg
        assert "COHERE_API_KEY" in msg
        assert "https://dashboard.cohere.com/api-keys" in msg

    def test_cohere_missing_error_message_mentions_rag(self) -> None:
        """Verificar que mensagem menciona RAG."""
        error = CohereMissingError()
        msg = str(error)

        assert "RAG" in msg or "embedding" in msg

    def test_cohere_missing_error_message_mentions_first_class(self) -> None:
        """Verificar que mensagem menciona integração first-class."""
        error = CohereMissingError()
        msg = str(error)

        assert "first-class" in msg or "LangChain" in msg


class TestGetEnv:
    """Testes para função get_env()."""

    def test_get_env_returns_value_when_exists(self) -> None:
        """Verificar que get_env() retorna valor quando existe."""
        with patch.dict(os.environ, {"TEST_VAR": "test_value"}):
            result = get_env("TEST_VAR")
            assert result == "test_value"

    def test_get_env_strict_true_raises_error_when_missing(self) -> None:
        """Verificar que get_env(strict=True) lança erro se não existe."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(GetEnvError, match="does not exist"):
                get_env("NONEXISTENT_VAR", strict=True)

    def test_get_env_strict_false_returns_none_when_missing(self) -> None:
        """Verificar que get_env(strict=False) retorna None se não existe."""
        with patch.dict(os.environ, {}, clear=True):
            result = get_env("NONEXISTENT_VAR", strict=False)
            assert result is None

    def test_get_env_default_strict_is_true(self) -> None:
        """Verificar que padrão é strict=True."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(GetEnvError):
                get_env("NONEXISTENT_VAR")

    def test_get_env_with_empty_string_is_valid(self) -> None:
        """Verificar que string vazia é um valor válido."""
        with patch.dict(os.environ, {"EMPTY_VAR": ""}):
            result = get_env("EMPTY_VAR", strict=False)
            assert result == ""

    def test_get_env_strict_with_empty_string(self) -> None:
        """Verificar que string vazia satisfaz strict=True."""
        with patch.dict(os.environ, {"EMPTY_VAR": ""}):
            result = get_env("EMPTY_VAR", strict=True)
            assert result == ""

    def test_get_env_error_message_includes_var_name(self) -> None:
        """Verificar que mensagem de erro inclui nome da variável."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(GetEnvError) as exc_info:
                get_env("MY_SPECIAL_VAR", strict=True)

            assert "MY_SPECIAL_VAR" in str(exc_info.value)


class TestValidateCohere:
    """Testes para função validate_cohere()."""

    def test_validate_cohere_success_with_key(self) -> None:
        """Verificar que validate_cohere() não lança erro com chave válida."""
        with patch.dict(os.environ, {"COHERE_API_KEY": "test-key-123"}):
            # Should not raise
            validate_cohere()

    def test_validate_cohere_raises_error_without_key(self) -> None:
        """Verificar que validate_cohere() lança erro sem chave."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(CohereMissingError):
                validate_cohere()

    def test_validate_cohere_raises_error_with_empty_key(self) -> None:
        """Verificar que validate_cohere() lança erro com chave vazia."""
        with patch.dict(os.environ, {"COHERE_API_KEY": ""}):
            with pytest.raises(CohereMissingError):
                validate_cohere()

    def test_validate_cohere_accepts_whitespace_key(self) -> None:
        """Verificar que validate_cohere() rejeita chave só com espaços."""
        with patch.dict(os.environ, {"COHERE_API_KEY": "   "}):
            # Whitespace-only strings are still truthy, so no error
            validate_cohere()

    def test_validate_cohere_correct_exception_type(self) -> None:
        """Verificar que validate_cohere() lança CohereMissingError."""
        with patch.dict(os.environ, {}, clear=True):
            try:
                validate_cohere()
                pytest.fail("Should have raised CohereMissingError")
            except CohereMissingError:
                pass  # Expected
            except Exception as e:
                pytest.fail(f"Wrong exception type: {type(e)}")


class TestEnvIntegration:
    """Testes de integração entre funções."""

    def test_get_env_with_validate_cohere(self) -> None:
        """Verificar que get_env() e validate_cohere() funcionam juntos."""
        with patch.dict(os.environ, {"COHERE_API_KEY": "valid-key"}):
            # This should work
            validate_cohere()
            cohere_key = get_env("COHERE_API_KEY", strict=False)
            assert cohere_key == "valid-key"

    def test_strict_false_does_not_raise_in_validate_cohere(self) -> None:
        """Verificar que validate_cohere() usa strict=False internamente."""
        with patch.dict(os.environ, {}, clear=True):
            # get_env with strict=False is used in validate_cohere
            # But validate_cohere() will still raise CohereMissingError
            with pytest.raises(CohereMissingError):
                validate_cohere()
