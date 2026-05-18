"""Testes unitários para o módulo de configuração (vectora.config.settings)."""

from vectora.config.settings import Settings, settings


class TestSettingsSingleton:
    """Testes para o singleton settings."""

    def test_settings_is_singleton_instance(self):
        """Verificar que settings é uma instância de Settings."""
        assert isinstance(settings, Settings)

    def test_settings_same_object_on_import(self):
        """Verificar que importar settings retorna o mesmo objeto."""
        from vectora.config.settings import settings as s2

        assert settings is s2


class TestSettingsFields:
    """Testes para campos essenciais de Settings."""

    def test_settings_has_enable_rag(self):
        """Verificar que settings tem campo enable_rag."""
        assert hasattr(settings, "enable_rag")
        assert isinstance(settings.enable_rag, bool)

    def test_settings_has_embedding_model(self):
        """Verificar que settings tem campo embedding_model."""
        assert hasattr(settings, "embedding_model")
        assert isinstance(settings.embedding_model, str)
        assert len(settings.embedding_model) > 0

    def test_settings_has_embedding_queue_enabled(self):
        """Verificar que settings tem embedding_queue_enabled."""
        assert hasattr(settings, "embedding_queue_enabled")
        assert isinstance(settings.embedding_queue_enabled, bool)

    def test_settings_has_llm_provider(self):
        """Verificar que settings tem campo llm_provider."""
        assert hasattr(settings, "llm_provider")
        assert settings.llm_provider in (
            "google-genai",
            "openai",
            "anthropic",
            "ollama",
        )

    def test_settings_has_enable_web_search(self):
        """Verificar que settings tem enable_web_search."""
        assert hasattr(settings, "enable_web_search")
        assert isinstance(settings.enable_web_search, bool)

    def test_settings_has_enable_file_operations(self):
        """Verificar que settings tem enable_file_operations."""
        assert hasattr(settings, "enable_file_operations")
        assert isinstance(settings.enable_file_operations, bool)

    def test_settings_has_cohere_api_key_field(self):
        """Verificar que settings tem campo cohere_api_key."""
        assert hasattr(settings, "cohere_api_key")
        # cohere_api_key pode ser None ou str

    def test_settings_has_search_min_score(self):
        """Verificar que settings tem search_min_score."""
        assert hasattr(settings, "search_min_score")
        assert isinstance(settings.search_min_score, float)

    def test_settings_has_chunk_size(self):
        """Verificar que settings tem chunk_size."""
        assert hasattr(settings, "chunk_size")
        assert isinstance(settings.chunk_size, int)
        assert settings.chunk_size > 0


class TestSettingsMethods:
    """Testes para métodos de Settings."""

    def test_settings_has_get_cohere_api_key(self):
        """Verificar que Settings tem get_cohere_api_key()."""
        assert hasattr(settings, "get_cohere_api_key")
        assert callable(settings.get_cohere_api_key)

    def test_settings_has_get_llm_provider(self):
        """Verificar que Settings tem get_llm_provider()."""
        assert hasattr(settings, "get_llm_provider")
        assert callable(settings.get_llm_provider)

    def test_get_llm_provider_returns_string(self):
        """Verificar que get_llm_provider() retorna string."""
        provider = settings.get_llm_provider()
        assert isinstance(provider, str)
        assert len(provider) > 0

    def test_settings_has_get_llm_model(self):
        """Verificar que Settings tem get_llm_model()."""
        assert hasattr(settings, "get_llm_model")
        assert callable(settings.get_llm_model)

    def test_get_llm_model_returns_string(self):
        """Verificar que get_llm_model() retorna string."""
        model = settings.get_llm_model()
        assert isinstance(model, str)
        assert len(model) > 0


class TestSettingsFromEnvironment:
    """Testes para carregamento de variáveis de ambiente."""

    def test_settings_instance_is_valid(self, monkeypatch):
        """Verificar que uma nova instância de Settings é válida."""
        s = Settings()
        assert s is not None

    def test_settings_enable_rag_default(self):
        """Verificar valor padrão de enable_rag."""
        s = Settings()
        # O valor padrão conforme definido no código
        assert isinstance(s.enable_rag, bool)

    def test_settings_embedding_model_default(self):
        """Verificar valor padrão de embedding_model."""
        s = Settings()
        assert s.embedding_model == "embed-multilingual-v3.0"

    def test_settings_reranker_model_default(self):
        """Verificar valor padrão de reranker_model."""
        s = Settings()
        assert "rerank" in s.reranker_model


class TestSettingsDefaults:
    """Testes para valores padrão razoáveis."""

    def test_chunk_size_reasonable(self):
        """Verificar que chunk_size tem valor razoável."""
        s = Settings()
        assert 100 <= s.chunk_size <= 10000

    def test_chunk_overlap_less_than_chunk_size(self):
        """Verificar que chunk_overlap é menor que chunk_size."""
        s = Settings()
        assert s.chunk_overlap < s.chunk_size

    def test_search_min_score_in_range(self):
        """Verificar que search_min_score está entre 0 e 1."""
        s = Settings()
        assert 0.0 <= s.search_min_score <= 1.0

    def test_default_search_top_k_positive(self):
        """Verificar que default_search_top_k é positivo."""
        s = Settings()
        assert s.default_search_top_k > 0

    def test_max_context_tokens_positive(self):
        """Verificar que max_context_tokens é positivo."""
        s = Settings()
        assert s.max_context_tokens > 0
