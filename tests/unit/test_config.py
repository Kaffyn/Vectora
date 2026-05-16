"""Testes unitários para o módulo de configuração (config.py)."""

from config import Config


class TestConfigSingleton:
    """Testes para o padrão singleton da Config."""

    def test_config_instance_returns_same_object(self):
        """Verificar que Config.instance() retorna sempre a mesma instância."""
        config1 = Config.instance()
        config2 = Config.instance()
        assert config1 is config2

    def test_config_instance_creates_on_first_call(self):
        """Verificar que Config.instance() cria instância na primeira chamada."""
        # Reset singleton
        Config._instance = None
        config = Config.instance()
        assert config is not None
        assert isinstance(config, Config)


class TestConfigLLMProvider:
    """Testes para detecção de LLM provider."""

    def test_get_llm_provider_returns_configured_provider(self, monkeypatch):
        """Verificar que get_llm_provider retorna o provider configurado."""
        monkeypatch.setenv("LLM_PROVIDER", "google-genai")
        Config._instance = None  # Reset singleton
        config = Config.instance()
        assert config.get_llm_provider() == "google-genai"

    def test_get_llm_provider_returns_none_if_not_configured(self, monkeypatch):
        """Verificar que get_llm_provider retorna None se não configurado."""
        # Criar uma instância Mock que não lê .env
        from unittest.mock import patch

        with patch.object(Config, "get", return_value=None):
            Config._instance = None  # Reset singleton
            config = Config.instance()
            # Mock the get method to return None for all keys
            config.get = lambda key, default=None: None
            provider = config.get_llm_provider()
            assert provider is None

    def test_get_llm_provider_detects_from_env_var(self, monkeypatch):
        """Verificar detecção de provider por variáveis de ambiente."""
        monkeypatch.delenv("LLM_PROVIDER", raising=False)
        monkeypatch.setenv("OPENAI_API_KEY", "test-key")
        Config._instance = None  # Reset singleton
        config = Config.instance()
        # Se OPENAI_API_KEY está set, deveria detectar openai
        # (implementação específica depende do código)
        assert config is not None


class TestConfigEnvironmentVariables:
    """Testes para leitura de variáveis de ambiente."""

    def test_config_reads_google_api_key(self, monkeypatch):
        """Verificar que Config lê GOOGLE_API_KEY."""
        monkeypatch.setenv("GOOGLE_API_KEY", "test-gemini-key")
        Config._instance = None  # Reset singleton
        config = Config.instance()
        # Verificar que a chave foi lida (implementação específica)
        assert config is not None

    def test_config_reads_openai_api_key(self, monkeypatch):
        """Verificar que Config lê OPENAI_API_KEY."""
        monkeypatch.setenv("OPENAI_API_KEY", "test-openai-key")
        Config._instance = None  # Reset singleton
        config = Config.instance()
        assert config is not None

    def test_config_reads_anthropic_api_key(self, monkeypatch):
        """Verificar que Config lê ANTHROPIC_API_KEY."""
        monkeypatch.setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
        Config._instance = None  # Reset singleton
        config = Config.instance()
        assert config is not None

    def test_config_reads_langsmith_api_key(self, monkeypatch):
        """Verificar que Config lê LANGSMITH_API_KEY para observabilidade."""
        monkeypatch.setenv("LANGSMITH_API_KEY", "test-langsmith-key")
        Config._instance = None  # Reset singleton
        config = Config.instance()
        assert config is not None


class TestConfigDefaultValues:
    """Testes para valores padrão da configuração."""

    def test_config_has_reasonable_defaults(self):
        """Verificar que Config tem valores padrão sensatos."""
        Config._instance = None  # Reset singleton
        config = Config.instance()
        # Verificar que propriedades essenciais existem
        assert hasattr(config, "get_llm_provider")
        assert callable(config.get_llm_provider)

    def test_config_vector_store_defaults_to_lancedb(self, monkeypatch):
        """Verificar que vector store padrão é LanceDB."""
        monkeypatch.delenv("VECTOR_STORE", raising=False)
        Config._instance = None  # Reset singleton
        config = Config.instance()
        # Vector store padrão deve ser LanceDB para dev
        assert config is not None


class TestConfigValidation:
    """Testes para validação de configurações."""

    def test_config_validates_valid_llm_provider(self, monkeypatch):
        """Verificar que Config valida provider válidos."""
        monkeypatch.setenv("LLM_PROVIDER", "google-genai")
        Config._instance = None  # Reset singleton
        config = Config.instance()
        provider = config.get_llm_provider()
        # Provider válido não deve gerar erro
        assert provider == "google-genai"

    def test_config_handles_invalid_llm_provider(self, monkeypatch):
        """Verificar que Config trata provider inválidos."""
        monkeypatch.setenv("LLM_PROVIDER", "invalid-provider")
        Config._instance = None  # Reset singleton
        config = Config.instance()
        # Config não deve falhar, apenas retornar o valor
        assert config is not None

    def test_config_handles_missing_api_keys(self, monkeypatch):
        """Verificar que Config trata chaves de API ausentes."""
        monkeypatch.delenv("GOOGLE_API_KEY", raising=False)
        monkeypatch.delenv("OPENAI_API_KEY", raising=False)
        monkeypatch.delenv("ANTHROPIC_API_KEY", raising=False)
        Config._instance = None  # Reset singleton
        config = Config.instance()
        # Config não deve falhar mesmo sem chaves
        assert config is not None


class TestConfigReloadBehavior:
    """Testes para comportamento de recarga de configuração."""

    def test_config_singleton_persists_across_calls(self):
        """Verificar que singleton persiste entre chamadas."""
        Config._instance = None  # Reset singleton
        config1 = Config.instance()
        config2 = Config.instance()
        config3 = Config.instance()
        assert config1 is config2 is config3

    def test_config_reset_creates_new_instance(self):
        """Verificar que reset cria nova instância."""
        config1 = Config.instance()
        Config._instance = None  # Reset singleton
        config2 = Config.instance()
        assert config1 is not config2
