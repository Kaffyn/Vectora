"""Testes unitários para o módulo de utilidades (utils.py)."""

import pytest
from unittest.mock import patch, MagicMock

from utils import load_llm


class TestLoadLLM:
    """Testes para função load_llm."""

    @patch("utils.ChatGoogleGenerativeAI")
    def test_load_llm_with_google_provider(self, mock_google):
        """Verificar que load_llm retorna Gemini para provider google-genai."""
        mock_llm = MagicMock()
        mock_google.return_value = mock_llm

        with patch.dict("os.environ", {"LLM_PROVIDER": "google-genai"}):
            llm = load_llm()
            # Deve retornar instância de LLM
            assert llm is not None

    @patch("utils.ChatOpenAI")
    def test_load_llm_with_openai_provider(self, mock_openai):
        """Verificar que load_llm retorna OpenAI para provider openai."""
        mock_llm = MagicMock()
        mock_openai.return_value = mock_llm

        with patch.dict("os.environ", {"LLM_PROVIDER": "openai"}):
            llm = load_llm()
            assert llm is not None

    @patch("utils.ChatAnthropic")
    def test_load_llm_with_anthropic_provider(self, mock_anthropic):
        """Verificar que load_llm retorna Anthropic para provider anthropic."""
        mock_llm = MagicMock()
        mock_anthropic.return_value = mock_llm

        with patch.dict("os.environ", {"LLM_PROVIDER": "anthropic"}):
            llm = load_llm()
            assert llm is not None

    @patch("utils.ChatOllama")
    def test_load_llm_with_ollama_provider(self, mock_ollama):
        """Verificar que load_llm retorna Ollama para provider ollama."""
        mock_llm = MagicMock()
        mock_ollama.return_value = mock_llm

        with patch.dict("os.environ", {"LLM_PROVIDER": "ollama"}):
            llm = load_llm()
            assert llm is not None

    def test_load_llm_raises_on_missing_provider(self):
        """Verificar que load_llm gera erro se provider não está configurado."""
        with patch.dict("os.environ", {}, clear=True):
            with pytest.raises(ValueError):
                load_llm()

    def test_load_llm_raises_on_unknown_provider(self):
        """Verificar que load_llm gera erro para provider desconhecido."""
        with patch.dict("os.environ", {"LLM_PROVIDER": "unknown-provider"}):
            with pytest.raises(ValueError):
                load_llm()

    @patch("utils.ChatGoogleGenerativeAI")
    def test_load_llm_raises_on_missing_api_key(self, mock_google):
        """Verificar que load_llm gera erro se API key não está configurada."""
        mock_google.side_effect = ValueError("API key is required")

        with patch.dict("os.environ", {"LLM_PROVIDER": "google-genai"}, clear=True):
            with pytest.raises(ValueError):
                load_llm()

    @patch("utils.ChatGoogleGenerativeAI")
    def test_load_llm_returns_chat_model(self, mock_google):
        """Verificar que load_llm retorna um ChatModel válido."""
        mock_llm = MagicMock()
        mock_llm.invoke = MagicMock(return_value="response")
        mock_google.return_value = mock_llm

        with patch.dict("os.environ", {"LLM_PROVIDER": "google-genai"}):
            llm = load_llm()
            # Deve ter método invoke
            assert hasattr(llm, "invoke") or hasattr(llm, "ainvoke")


class TestTrimMessages:
    """Testes para função trim_messages (se existir em utils)."""

    def test_trim_messages_removes_old_messages(self):
        """Verificar que trim_messages remove mensagens antigas."""
        from langchain_core.messages import HumanMessage, AIMessage

        messages = [
            HumanMessage(content="msg1"),
            AIMessage(content="resp1"),
            HumanMessage(content="msg2"),
            AIMessage(content="resp2"),
            HumanMessage(content="msg3"),
        ]
        # Implementação específica de trim_messages
        # Este teste é genérico e deve ser adaptado

    def test_trim_messages_preserves_message_order(self):
        """Verificar que trim_messages preserva ordem das mensagens."""
        from langchain_core.messages import HumanMessage, AIMessage

        messages = [
            HumanMessage(content="msg1"),
            AIMessage(content="resp1"),
            HumanMessage(content="msg2"),
        ]
        # Verificar que ordem é preservada


class TestLLMDetection:
    """Testes para detecção automática de LLM."""

    def test_detect_llm_from_google_api_key(self):
        """Verificar detecção de Gemini por GOOGLE_API_KEY."""
        with patch.dict("os.environ", {"GOOGLE_API_KEY": "test-key"}):
            # Função de detecção deveria identificar Gemini
            pass

    def test_detect_llm_from_openai_api_key(self):
        """Verificar detecção de OpenAI por OPENAI_API_KEY."""
        with patch.dict("os.environ", {"OPENAI_API_KEY": "test-key"}):
            # Função de detecção deveria identificar OpenAI
            pass

    def test_detect_llm_from_anthropic_api_key(self):
        """Verificar detecção de Anthropic por ANTHROPIC_API_KEY."""
        with patch.dict("os.environ", {"ANTHROPIC_API_KEY": "test-key"}):
            # Função de detecção deveria identificar Anthropic
            pass

    def test_detect_no_llm_if_no_api_keys(self):
        """Verificar que nenhum LLM é detectado se não há chaves."""
        with patch.dict("os.environ", {}, clear=True):
            # Nenhum LLM deveria ser detectado
            pass


class TestLLMErrorHandling:
    """Testes para tratamento de erros do LLM."""

    @patch("utils.ChatGoogleGenerativeAI")
    def test_handle_invalid_api_key(self, mock_google):
        """Verificar tratamento de API key inválida."""
        mock_google.side_effect = ValueError("Invalid API key")

        with patch.dict("os.environ", {"LLM_PROVIDER": "google-genai"}):
            with pytest.raises(ValueError):
                load_llm()

    @patch("utils.ChatGoogleGenerativeAI")
    def test_handle_network_error(self, mock_google):
        """Verificar tratamento de erro de rede."""
        mock_google.side_effect = ConnectionError("Network error")

        with patch.dict("os.environ", {"LLM_PROVIDER": "google-genai"}):
            with pytest.raises(ConnectionError):
                load_llm()

    @patch("utils.ChatGoogleGenerativeAI")
    def test_handle_rate_limit(self, mock_google):
        """Verificar tratamento de rate limit."""
        mock_google.side_effect = ValueError("Rate limit exceeded")

        with patch.dict("os.environ", {"LLM_PROVIDER": "google-genai"}):
            with pytest.raises(ValueError):
                load_llm()
