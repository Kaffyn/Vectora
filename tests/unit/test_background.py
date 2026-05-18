"""Testes para background.py (Background Embedding Worker)."""

from __future__ import annotations

import asyncio
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.services.background import BackgroundEmbeddingWorker


class TestBackgroundEmbeddingWorkerInit:
    """Testes para inicializacao."""

    def test_worker_init(self) -> None:
        """Verificar que worker e inicializado."""
        worker = BackgroundEmbeddingWorker()

        assert worker.running is False
        assert worker.task is None
        assert worker.processed_count == 0
        assert worker.failed_count == 0


class TestBackgroundEmbeddingWorkerStart:
    """Testes para metodo start()."""

    @pytest.mark.asyncio
    async def test_start_sets_running_flag(self) -> None:
        """Verificar que start() ativa a flag running."""
        worker = BackgroundEmbeddingWorker()

        with patch.object(worker, "_reconcile_startup", new_callable=AsyncMock):
            with patch.object(worker, "_run_loop", new_callable=AsyncMock):
                with patch("vectora.services.background.settings") as mock_settings:
                    mock_settings.enable_rag = True
                    await worker.start()

                    assert worker.running is True

    @pytest.mark.asyncio
    async def test_start_creates_task(self) -> None:
        """Verificar que start() cria asyncio.Task."""
        worker = BackgroundEmbeddingWorker()

        with patch.object(worker, "_reconcile_startup", new_callable=AsyncMock):
            with patch.object(worker, "_run_loop", new_callable=AsyncMock):
                with patch("vectora.services.background.settings") as mock_settings:
                    mock_settings.enable_rag = True
                    await worker.start()

                    assert worker.task is not None

    @pytest.mark.asyncio
    async def test_start_respects_rag_enabled_flag(self) -> None:
        """Verificar que start() respeita enable_rag."""
        worker = BackgroundEmbeddingWorker()

        with patch("vectora.services.background.settings") as mock_settings:
            mock_settings.enable_rag = False
            await worker.start()

            assert worker.running is False
            assert worker.task is None

    @pytest.mark.asyncio
    async def test_start_ignores_second_call(self) -> None:
        """Verificar que start() ignora chamada duplicada."""
        worker = BackgroundEmbeddingWorker()

        with patch.object(worker, "_reconcile_startup", new_callable=AsyncMock):
            with patch.object(worker, "_run_loop", new_callable=AsyncMock):
                with patch("vectora.services.background.settings") as mock_settings:
                    mock_settings.enable_rag = True
                    await worker.start()
                    first_task = worker.task

                    await worker.start()
                    second_task = worker.task

                    assert first_task is second_task


class TestBackgroundEmbeddingWorkerStop:
    """Testes para metodo stop()."""

    @pytest.mark.asyncio
    async def test_stop_when_not_running(self) -> None:
        """Verificar que stop() com worker parado e seguro."""
        worker = BackgroundEmbeddingWorker()

        # Nao deve lancar excecao
        await worker.stop()

        assert worker.running is False

    @pytest.mark.asyncio
    async def test_stop_sets_running_false(self) -> None:
        """Verificar que stop() desativa flag running."""
        worker = BackgroundEmbeddingWorker()

        with patch.object(worker, "_reconcile_startup", new_callable=AsyncMock):
            with patch.object(worker, "_run_loop", new_callable=AsyncMock):
                with patch("vectora.services.background.settings") as mock_settings:
                    mock_settings.enable_rag = True
                    await worker.start()

                    # Aguardar um pouco
                    await asyncio.sleep(0.01)

                    await worker.stop(timeout_seconds=1)

                    assert worker.running is False


class TestBackgroundEmbeddingWorkerGetQueue:
    """Testes para metodo _get_queue()."""

    @pytest.mark.asyncio
    async def test_get_queue_returns_queue(self) -> None:
        """Verificar que _get_queue() retorna queue."""
        worker = BackgroundEmbeddingWorker()

        with patch("vectora.services.background.get_embedding_queue") as mock_get_queue:
            mock_queue = AsyncMock()
            mock_get_queue.return_value = mock_queue

            result = await worker._get_queue()

            assert result is mock_queue


class TestBackgroundEmbeddingWorkerGenerateEmbedding:
    """Testes para metodo _generate_embedding()."""

    @pytest.mark.asyncio
    async def test_generate_embedding_returns_list(self) -> None:
        """Verificar que _generate_embedding retorna lista de floats."""
        worker = BackgroundEmbeddingWorker()

        with patch("vectora.services.background.CohereEmbeddings") as mock_cohere:
            mock_instance = MagicMock()
            mock_instance.embed_query = AsyncMock(return_value=[0.1, 0.2, 0.3])
            mock_cohere.return_value = mock_instance

            with patch("vectora.services.background.settings") as mock_settings:
                mock_settings.cohere_api_key = "test-key"
                mock_settings.embedding_model = "test-model"

                result = await worker._generate_embedding("test text")

                assert isinstance(result, list)
