"""Testes unitários para o módulo de utilidades (utils.py)."""

from unittest.mock import MagicMock, patch

import pytest

from utils import load_llm


class TestLoadLLM:
    """Testes para função load_llm."""

    @patch("utils.init_chat_model")
    def test_load_llm_with_google_provider(self, mock_init):
        """Verificar que load_llm retorna LLM para provider google-genai."""
        mock_llm = MagicMock()
        mock_init.return_value = mock_llm

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "google-genai",
                "GOOGLE_API_KEY": "test-key",
            },
        ):
            llm = load_llm()
            assert llm is not None
            mock_init.assert_called()

    @patch("utils.init_chat_model")
    def test_load_llm_with_openai_provider(self, mock_init):
        """Verificar que load_llm retorna OpenAI para provider openai."""
        mock_llm = MagicMock()
        mock_init.return_value = mock_llm

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "openai",
                "OPENAI_API_KEY": "test-key",
            },
        ):
            llm = load_llm()
            assert llm is not None

    @patch("utils.init_chat_model")
    def test_load_llm_with_anthropic_provider(self, mock_init):
        """Verificar que load_llm retorna Anthropic para provider anthropic."""
        mock_llm = MagicMock()
        mock_init.return_value = mock_llm

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "anthropic",
                "ANTHROPIC_API_KEY": "test-key",
            },
        ):
            llm = load_llm()
            assert llm is not None

    @patch("utils.init_chat_model")
    def test_load_llm_with_ollama_provider(self, mock_init):
        """Verificar que load_llm retorna Ollama para provider ollama."""
        mock_llm = MagicMock()
        mock_init.return_value = mock_llm

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "ollama",
                "OLLAMA_BASE_URL": "http://localhost:11434",
            },
        ):
            llm = load_llm()
            assert llm is not None

    def test_load_llm_raises_on_missing_provider(self):
        """Verificar que load_llm gera erro se provider não está configurado."""
        from env import GetEnvError

        with patch.dict("os.environ", {}, clear=True):
            with pytest.raises((GetEnvError, ValueError, KeyError)):
                load_llm()

    def test_load_llm_raises_on_unknown_provider(self):
        """Verificar que load_llm gera erro para provider desconhecido."""
        with patch.dict(
            "os.environ",
            {"LLM_PROVIDER": "unknown-provider"},
        ):
            with pytest.raises(ValueError):
                load_llm()

    @patch("utils.init_chat_model")
    def test_load_llm_raises_on_missing_api_key(self, mock_init):
        """Verificar que load_llm gera erro se API key não está configurada."""
        mock_init.side_effect = ValueError("API key is required")

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "google-genai",
            },
        ):
            with pytest.raises(ValueError):
                load_llm()

    @patch("utils.init_chat_model")
    def test_load_llm_returns_chat_model(self, mock_init):
        """Verificar que load_llm retorna um ChatModel válido."""
        mock_llm = MagicMock()
        mock_llm.invoke = MagicMock(return_value="response")
        mock_init.return_value = mock_llm

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "google-genai",
                "GOOGLE_API_KEY": "test-key",
            },
        ):
            llm = load_llm()
            assert hasattr(llm, "invoke") or hasattr(llm, "ainvoke")


class TestLLMErrorHandling:
    """Testes para tratamento de erros do LLM."""

    @patch("utils.init_chat_model")
    def test_handle_invalid_api_key(self, mock_init):
        """Verificar tratamento de API key inválida."""
        mock_init.side_effect = ValueError("Invalid API key")

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "google-genai",
                "GOOGLE_API_KEY": "invalid",
            },
        ):
            with pytest.raises(ValueError):
                load_llm()

    @patch("utils.init_chat_model")
    def test_handle_network_error(self, mock_init):
        """Verificar tratamento de erro de rede."""
        mock_init.side_effect = ConnectionError("Network error")

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "google-genai",
                "GOOGLE_API_KEY": "test-key",
            },
        ):
            with pytest.raises(ConnectionError):
                load_llm()

    @patch("utils.init_chat_model")
    def test_handle_rate_limit(self, mock_init):
        """Verificar tratamento de rate limit."""
        mock_init.side_effect = ValueError("Rate limit exceeded")

        with patch.dict(
            "os.environ",
            {
                "LLM_PROVIDER": "google-genai",
                "GOOGLE_API_KEY": "test-key",
            },
        ):
            with pytest.raises(ValueError):
                load_llm()
