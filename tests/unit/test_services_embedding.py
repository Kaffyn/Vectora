"""Testes para vectora/services/embedding.py"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.config.settings import Settings
from vectora.services.embedding import EmbeddingService


class TestEmbeddingServiceInit:
    """Testes de inicializacao do EmbeddingService."""

    def test_init_basic(self):
        """Verificar inicializacao basica."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)

        assert service.settings is settings
        assert service.worker_running is False
        assert service.worker_task is None
        assert service.db is None
        assert service.embeddings is None

    def test_init_configuration(self):
        """Verificar configuracao inicial."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)

        assert service.batch_size == 10
        assert service.max_parallel == 5
        assert service.polling_interval == 5
        assert service.retry_backoff == [1, 2, 4]
        assert service.max_retries == 3

    def test_init_semaphores(self):
        """Verificar inicializacao de semaphores."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)

        assert service.semaphore is not None
        assert service.lancedb_semaphore is not None

    def test_init_ignore_validator(self):
        """Verificar que ignore_validator e inicializado."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)

        assert service.ignore_validator is not None


class TestEmbeddingServiceStart:
    """Testes para metodo start()."""

    @pytest.mark.asyncio
    async def test_start_rag_disabled(self):
        """Testar que start nao executa quando RAG desabilitado."""
        settings = MagicMock(spec=Settings)
        settings.enable_rag = False
        service = EmbeddingService(settings)

        await service.start()

        assert service.worker_running is False
        assert service.worker_task is None

    @pytest.mark.asyncio
    async def test_start_already_running(self):
        """Testar que start nao executa se ja rodando."""
        settings = MagicMock(spec=Settings)
        settings.enable_rag = True
        service = EmbeddingService(settings)
        service.worker_running = True

        await service.start()

        # Nao deve mudar estado
        assert service.worker_running is True

    @pytest.mark.asyncio
    async def test_start_initializes_services(self):
        """Testar que start inicializa vector store e embeddings."""
        settings = MagicMock(spec=Settings)
        settings.enable_rag = True
        service = EmbeddingService(settings)

        with patch.object(service, "_initialize_vector_store", new_callable=AsyncMock):
            with patch.object(
                service, "_initialize_embeddings", new_callable=AsyncMock
            ):
                await service.start()

                assert service.worker_running is True


class TestEmbeddingServiceQueueDocument:
    """Testes para metodo queue_document()."""

    @pytest.mark.asyncio
    async def test_queue_document_worker_not_running(self):
        """Testar que documento nao e enfileirado se worker nao roda."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.worker_running = False

        result = await service.queue_document(
            doc_id="doc1",
            text="Test content",
        )

        assert result is False

    @pytest.mark.asyncio
    async def test_queue_document_success(self):
        """Testar enfileiramento bem-sucedido de documento."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.worker_running = True

        # Mock ignore_validator
        with patch.object(service, "ignore_validator") as mock_validator:
            mock_validator.should_ignore.return_value = False

            result = await service.queue_document(
                doc_id="doc1",
                text="Test content",
                collection="documents",
            )

            assert result is True

    @pytest.mark.asyncio
    async def test_queue_document_ignored_path(self):
        """Testar que documento e rejeitado se path matches ignore pattern."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.worker_running = True

        # Mock ignore_validator
        with patch.object(service, "ignore_validator") as mock_validator:
            mock_validator.should_ignore.return_value = True

            result = await service.queue_document(
                doc_id="doc1",
                text="Secret content",
                file_path="node_modules/secret.env",
            )

            assert result is False

    @pytest.mark.asyncio
    async def test_queue_document_without_file_path(self):
        """Testar enfileiramento sem validacao de path."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.worker_running = True

        result = await service.queue_document(
            doc_id="doc2",
            text="Plain text",
        )

        assert result is True


class TestEmbeddingServiceSearch:
    """Testes para metodo search()."""

    @pytest.mark.asyncio
    async def test_search_vector_store_not_initialized(self):
        """Testar que search retorna vazio se vector store nao inicializado."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.db = None
        service.embeddings = None

        results = await service.search("test query")

        assert results == []

    @pytest.mark.asyncio
    async def test_search_with_query(self):
        """Testar busca com query."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)

        # Mock vector store
        mock_table = MagicMock()
        mock_table.search.return_value.limit.return_value.to_list.return_value = [
            {"doc_id": "1", "text": "result", "_distance": 0.95}
        ]

        service.db = MagicMock()
        service.db.open_table.return_value = mock_table
        service.embeddings = MagicMock()

        with patch.object(
            service, "_get_embedding_with_retry", new_callable=AsyncMock
        ) as mock_get_emb:
            mock_get_emb.return_value = [0.1, 0.2, 0.3]

            results = await service.search("test", limit=5)

            assert len(results) > 0


class TestEmbeddingServiceQueueStatus:
    """Testes para metodo get_queue_status()."""

    @pytest.mark.asyncio
    async def test_get_queue_status(self):
        """Testar obtencao de status da fila."""
        settings = MagicMock(spec=Settings)
        settings.max_embedding_queue_size = 1000
        service = EmbeddingService(settings)
        service.worker_running = True

        status = await service.get_queue_status()

        assert "pending_count" in status
        assert "max_queue_size" in status
        assert "queue_usage_percent" in status
        assert "worker_running" in status
        assert status["worker_running"] is True
        assert status["max_queue_size"] == 1000

    @pytest.mark.asyncio
    async def test_get_queue_status_worker_stopped(self):
        """Testar status quando worker esta parado."""
        settings = MagicMock(spec=Settings)
        settings.max_embedding_queue_size = 500
        service = EmbeddingService(settings)
        service.worker_running = False

        status = await service.get_queue_status()

        assert status["worker_running"] is False


class TestEmbeddingServiceClearCollection:
    """Testes para metodo clear_collection()."""

    @pytest.mark.asyncio
    async def test_clear_collection_vector_store_not_init(self):
        """Testar que clear retorna 0 se vector store nao existe."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.db = None

        count = await service.clear_collection("documents")

        assert count == 0

    @pytest.mark.asyncio
    async def test_clear_collection_success(self):
        """Testar limpeza bem-sucedida de colecao."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.db = MagicMock()
        service.db.drop_table = MagicMock()

        with patch("asyncio.to_thread", new_callable=AsyncMock) as mock_to_thread:
            mock_to_thread.return_value = None
            count = await service.clear_collection("documents")

            assert count == 1


class TestEmbeddingServiceHealthCheck:
    """Testes para metodo health_check()."""

    @pytest.mark.asyncio
    async def test_health_check_worker_not_running(self):
        """Testar health check quando worker nao roda."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.worker_running = False

        is_healthy, message = await service.health_check()

        assert is_healthy is False
        assert "not running" in message

    @pytest.mark.asyncio
    async def test_health_check_vector_store_not_init(self):
        """Testar health check quando vector store nao inicializado."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.worker_running = True
        service.db = None
        service.embeddings = None

        is_healthy, message = await service.health_check()

        assert is_healthy is False
        assert "not initialized" in message

    @pytest.mark.asyncio
    async def test_health_check_healthy(self):
        """Testar health check quando tudo esta OK."""
        settings = MagicMock(spec=Settings)
        service = EmbeddingService(settings)
        service.worker_running = True
        service.db = MagicMock()
        service.embeddings = MagicMock()

        with patch.object(
            service, "_get_embedding_with_retry", new_callable=AsyncMock
        ) as mock_get:
            mock_get.return_value = [0.1, 0.2, 0.3]

            is_healthy, message = await service.health_check()

            assert is_healthy is True
            assert "healthy" in message
