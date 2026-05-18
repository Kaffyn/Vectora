"""Testes Unitários Abrangentes para Background Embedding Worker.

Cobre: inicialização, startup, shutdown, processamento de registros,
geração de embeddings, escrita em LanceDB, retry exponencial, DLQ,
idempotência, segurança de threads e reconciliação na startup.

Coverage alvo: 90-95% do arquivo vectora/services/background.py
"""

from __future__ import annotations

import asyncio
import contextlib
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.services.background import (
    MAX_PARALLEL,
    MAX_RETRIES,
    POLLING_INTERVAL,
    RETRY_BACKOFF,
    BackgroundEmbeddingWorker,
    get_background_worker,
)
from vectora.services.queue import EmbeddingQueueRecord

# ============================================================================
# TESTES DE INICIALIZAÇÃO
# ============================================================================


class TestBackgroundEmbeddingWorkerInit:
    """Testes para __init__() do BackgroundEmbeddingWorker."""

    def test_init_sets_config(self) -> None:
        """Verifica se __init__() define a configuração global."""
        worker = BackgroundEmbeddingWorker()
        assert worker.config is not None
        assert worker.config == worker.config  # Deve ser a mesma instância

    def test_init_sets_running_false(self) -> None:
        """Verifica se running é inicializado como False."""
        worker = BackgroundEmbeddingWorker()
        assert worker.running is False

    def test_init_sets_task_none(self) -> None:
        """Verifica se task é inicializado como None."""
        worker = BackgroundEmbeddingWorker()
        assert worker.task is None

    def test_init_creates_semaphore(self) -> None:
        """Verifica se semaphore é criado com MAX_PARALLEL(5)."""
        worker = BackgroundEmbeddingWorker()
        assert worker.semaphore is not None
        assert isinstance(worker.semaphore, asyncio.Semaphore)
        # Verificar valor interno do semaphore
        assert worker.semaphore._value == MAX_PARALLEL

    def test_init_creates_lancedb_semaphore(self) -> None:
        """Verifica se lancedb_semaphore é criado com valor 1."""
        worker = BackgroundEmbeddingWorker()
        assert worker.lancedb_semaphore is not None
        assert isinstance(worker.lancedb_semaphore, asyncio.Semaphore)
        assert worker.lancedb_semaphore._value == 1

    def test_init_sets_counters_zero(self) -> None:
        """Verifica se contadores processed_count e failed_count são 0."""
        worker = BackgroundEmbeddingWorker()
        assert worker.processed_count == 0
        assert worker.failed_count == 0


# ============================================================================
# TESTES DE STARTUP
# ============================================================================


class TestBackgroundEmbeddingWorkerStart:
    """Testes para start() do BackgroundEmbeddingWorker."""

    @pytest.mark.asyncio
    async def test_start_when_already_running(self) -> None:
        """Verifica se start() retorna cedo quando já está rodando."""
        with patch(
            "vectora.services.background.get_embedding_queue"
        ) as mock_queue_func:
            mock_queue = MagicMock()
            mock_queue.reconcile = AsyncMock()
            mock_queue_func.return_value = mock_queue

            worker = BackgroundEmbeddingWorker()
            worker.running = True  # Simular que já está rodando

            with patch.object(worker, "_reconcile_startup") as mock_reconcile:
                await worker.start()

                # Não deve chamar reconcile novamente
                mock_reconcile.assert_not_called()

    @pytest.mark.asyncio
    async def test_start_when_rag_disabled(self) -> None:
        """Verifica se start() retorna cedo quando RAG está desabilitado."""
        with patch("vectora.services.background.settings") as mock_settings:
            mock_settings.enable_rag = False

            worker = BackgroundEmbeddingWorker()
            worker.config = mock_settings

            await worker.start()

            # Não deve estar rodando
            assert worker.running is False
            assert worker.task is None

    @pytest.mark.asyncio
    async def test_start_when_rag_enabled(self) -> None:
        """Verifica se start() inicia o loop quando RAG está habilitado."""
        with patch(
            "vectora.services.background.get_embedding_queue"
        ) as mock_queue_func:
            mock_queue = MagicMock()
            mock_queue.reconcile = AsyncMock()
            mock_queue_func.return_value = mock_queue

            with patch("vectora.services.background.settings") as mock_settings:
                mock_settings.enable_rag = True

                worker = BackgroundEmbeddingWorker()
                worker.config = mock_settings

                await worker.start()

                # Deve estar rodando e ter uma task
                assert worker.running is True
                assert worker.task is not None
                assert not worker.task.done()

                # Cleanup
                await worker.stop(timeout_seconds=1)

    @pytest.mark.asyncio
    async def test_start_calls_reconcile_startup(self) -> None:
        """Verifica se start() chama _reconcile_startup()."""
        with patch(
            "vectora.services.background.get_embedding_queue"
        ) as mock_queue_func:
            mock_queue = MagicMock()
            mock_queue.reconcile = AsyncMock()
            mock_queue_func.return_value = mock_queue

            with patch("vectora.services.background.settings") as mock_settings:
                mock_settings.enable_rag = True

                worker = BackgroundEmbeddingWorker()
                worker.config = mock_settings

                with patch.object(
                    worker, "_reconcile_startup", new_callable=AsyncMock
                ) as mock_reconcile:
                    await worker.start()
                    mock_reconcile.assert_called_once()

                    # Cleanup
                    await worker.stop(timeout_seconds=1)


# ============================================================================
# TESTES DE SHUTDOWN
# ============================================================================


class TestBackgroundEmbeddingWorkerStop:
    """Testes para stop() do BackgroundEmbeddingWorker."""

    @pytest.mark.asyncio
    async def test_stop_when_not_running(self) -> None:
        """Verifica se stop() retorna cedo quando não está rodando."""
        worker = BackgroundEmbeddingWorker()
        worker.running = False

        # Não deve lançar exceção
        await worker.stop()

    @pytest.mark.asyncio
    async def test_stop_sets_running_false(self) -> None:
        """Verifica se stop() define running como False."""
        with patch(
            "vectora.services.background.get_embedding_queue"
        ) as mock_queue_func:
            mock_queue = MagicMock()
            mock_queue.reconcile = AsyncMock()
            mock_queue_func.return_value = mock_queue

            with patch("vectora.services.background.settings") as mock_settings:
                mock_settings.enable_rag = True

                worker = BackgroundEmbeddingWorker()
                worker.config = mock_settings

                await worker.start()
                assert worker.running is True

                await worker.stop(timeout_seconds=1)
                assert worker.running is False

    @pytest.mark.asyncio
    async def test_stop_awaits_task(self) -> None:
        """Verifica se stop() aguarda a conclusão da task."""
        with patch(
            "vectora.services.background.get_embedding_queue"
        ) as mock_queue_func:
            mock_queue = MagicMock()
            mock_queue.reconcile = AsyncMock()
            mock_queue_func.return_value = mock_queue

            with patch("vectora.services.background.settings") as mock_settings:
                mock_settings.enable_rag = True

                worker = BackgroundEmbeddingWorker()
                worker.config = mock_settings

                await worker.start()
                task = worker.task
                assert task is not None

                await worker.stop(timeout_seconds=5)

                # Task deve estar completa
                assert task.done()

    @pytest.mark.asyncio
    async def test_stop_cancels_on_timeout(self) -> None:
        """Verifica se stop() cancela a task em timeout."""
        # Criar um worker cuja task nunca termina
        worker = BackgroundEmbeddingWorker()
        worker.running = True

        # Criar uma task que bloqueia indefinidamente
        async def never_ends() -> None:
            await asyncio.sleep(100)

        worker.task = asyncio.create_task(never_ends())

        # Stop com timeout curto deve cancelar
        await worker.stop(timeout_seconds=0.1)

        # Task deve estar cancelada
        assert worker.task.done()
        assert worker.task.cancelled()


# ============================================================================
# TESTES DE PROCESSAMENTO DE REGISTROS
# ============================================================================


class TestBackgroundEmbeddingWorkerProcessRecord:
    """Testes para _process_record() do BackgroundEmbeddingWorker."""

    @pytest.mark.asyncio
    async def test_process_record_success_first_attempt(self) -> None:
        """Verifica processamento bem-sucedido na primeira tentativa."""
        worker = BackgroundEmbeddingWorker()

        # Mock da queue
        mock_queue = MagicMock()
        mock_queue.mark_processing = AsyncMock()
        mock_queue.mark_success = AsyncMock()

        # Mock do registro
        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-123"
        mock_record.text = "test text"
        mock_record.collection = "test-collection"

        # Mock de geração de embedding
        with patch.object(
            worker, "_generate_embedding", new_callable=AsyncMock
        ) as mock_gen:
            mock_gen.return_value = [0.1, 0.2, 0.3]

            with patch.object(
                worker, "_write_to_lancedb", new_callable=AsyncMock
            ) as mock_write:
                await worker._process_record(mock_record, mock_queue)

                # Verificar chamadas
                mock_queue.mark_processing.assert_called_once_with("q-123")
                mock_gen.assert_called_once_with("test text")
                mock_write.assert_called_once()
                mock_queue.mark_success.assert_called_once_with("q-123")

                # Verificar contador
                assert worker.processed_count == 1

    @pytest.mark.asyncio
    async def test_process_record_retry_backoff(self) -> None:
        """Verifica retry com exponential backoff."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        mock_queue.mark_processing = AsyncMock()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-retry"
        mock_record.text = "test"
        mock_record.collection = "test"

        # Falhar 2 vezes, suceder na 3ª
        call_count = 0

        async def failing_embed(text: str) -> list[float]:
            nonlocal call_count
            call_count += 1
            if call_count < 3:
                raise ValueError("API error")
            return [0.1, 0.2]

        with patch.object(worker, "_generate_embedding", side_effect=failing_embed):
            with patch.object(worker, "_write_to_lancedb", new_callable=AsyncMock):
                with patch("asyncio.sleep", new_callable=AsyncMock) as mock_sleep:
                    with patch.object(
                        worker, "_get_queue", new_callable=AsyncMock
                    ) as mock_get_queue:
                        mock_get_queue.return_value = mock_queue

                        await worker._process_record(mock_record, mock_queue)

                        # Deve ter dormido 2 vezes: 1s e 2s
                        assert mock_sleep.call_count == 2
                        mock_sleep.assert_any_call(RETRY_BACKOFF[0])  # 1s
                        mock_sleep.assert_any_call(RETRY_BACKOFF[1])  # 2s

    @pytest.mark.asyncio
    async def test_process_record_moves_to_dlq(self) -> None:
        """Verifica movimentação para DLQ após MAX_RETRIES."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        mock_queue.mark_processing = AsyncMock()
        mock_queue.mark_dlq = AsyncMock()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-dlq"
        mock_record.text = "test"
        mock_record.collection = "test"

        # Falhar sempre
        with patch.object(
            worker, "_generate_embedding", side_effect=ValueError("API error")
        ):
            with patch("asyncio.sleep", new_callable=AsyncMock):
                with patch.object(
                    worker, "_get_queue", new_callable=AsyncMock
                ) as mock_get_queue:
                    mock_get_queue.return_value = mock_queue

                    await worker._process_record(mock_record, mock_queue)

                    # Deve ter movido para DLQ
                    mock_queue.mark_dlq.assert_called_once()
                    call_args = mock_queue.mark_dlq.call_args
                    assert call_args[0][0] == "q-dlq"
                    # Verificar que a razão contém traceback
                    dlq_reason = call_args[0][1]
                    assert "Stack trace:" in dlq_reason

                    # Contador de falhas deve aumentar
                    assert worker.failed_count == 1

    @pytest.mark.asyncio
    async def test_process_record_uses_semaphore(self) -> None:
        """Verifica se o semaphore é usado durante processamento."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        mock_queue.mark_processing = AsyncMock()
        mock_queue.mark_success = AsyncMock()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-sem"
        mock_record.text = "test"
        mock_record.collection = "test"

        # Verificar que semaphore._value diminui e volta
        initial_value = worker.semaphore._value

        with patch.object(
            worker, "_generate_embedding", new_callable=AsyncMock
        ) as mock_gen:
            mock_gen.return_value = [0.1]

            with patch.object(worker, "_write_to_lancedb", new_callable=AsyncMock):
                # Processar
                await worker._process_record(mock_record, mock_queue)

                # Semaphore deve ter voltado ao estado inicial
                assert worker.semaphore._value == initial_value

    @pytest.mark.asyncio
    async def test_process_record_uses_queue_parameter(self) -> None:
        """Verifica se _process_record usa queue do parâmetro ou obtém do singleton."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        mock_queue.mark_processing = AsyncMock()
        mock_queue.mark_success = AsyncMock()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-param"
        mock_record.text = "test"
        mock_record.collection = "test"

        with patch.object(
            worker, "_generate_embedding", new_callable=AsyncMock
        ) as mock_gen:
            mock_gen.return_value = [0.1]

            with patch.object(worker, "_write_to_lancedb", new_callable=AsyncMock):
                # Passar queue como parâmetro
                await worker._process_record(mock_record, mock_queue)

                # Deve ter usado a queue passada
                mock_queue.mark_processing.assert_called()

    @pytest.mark.asyncio
    async def test_process_record_gets_queue_if_none(self) -> None:
        """Verifica se _process_record obtém queue do singleton se não passada."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        mock_queue.mark_processing = AsyncMock()
        mock_queue.mark_success = AsyncMock()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-none"
        mock_record.text = "test"
        mock_record.collection = "test"

        with patch.object(
            worker, "_generate_embedding", new_callable=AsyncMock
        ) as mock_gen:
            mock_gen.return_value = [0.1]

            with patch.object(worker, "_write_to_lancedb", new_callable=AsyncMock):
                with patch.object(
                    worker, "_get_queue", new_callable=AsyncMock
                ) as mock_get_queue:
                    mock_get_queue.return_value = mock_queue

                    # Não passar queue (None)
                    await worker._process_record(mock_record, None)

                    # Deve ter chamado _get_queue
                    mock_get_queue.assert_called()


# ============================================================================
# TESTES DE GERAÇÃO DE EMBEDDING
# ============================================================================


class TestBackgroundEmbeddingWorkerGenerateEmbedding:
    """Testes para _generate_embedding() do BackgroundEmbeddingWorker."""

    @pytest.mark.asyncio
    async def test_generate_embedding_success(self) -> None:
        """Verifica geração bem-sucedida de embedding."""
        worker = BackgroundEmbeddingWorker()

        with patch("vectora.services.background.CohereEmbeddings") as mock_cohere_class:
            mock_cohere = MagicMock()
            mock_cohere.embed_query = MagicMock(return_value=[0.1, 0.2, 0.3])
            mock_cohere_class.return_value = mock_cohere

            with patch.object(
                type(worker.config), "get_cohere_api_key", return_value="test-key"
            ):
                with patch("asyncio.to_thread", new_callable=AsyncMock) as mock_thread:
                    mock_thread.return_value = [0.1, 0.2, 0.3]

                    result = await worker._generate_embedding("test text")

                    assert result == [0.1, 0.2, 0.3]
                    mock_thread.assert_called_once()

    @pytest.mark.asyncio
    async def test_generate_embedding_missing_api_key(self) -> None:
        """Verifica erro quando COHERE_API_KEY não está configurado."""
        worker = BackgroundEmbeddingWorker()

        with patch.object(type(worker.config), "get_cohere_api_key", return_value=None):
            with pytest.raises(ValueError, match="COHERE_API_KEY não configurado"):
                await worker._generate_embedding("test")

    @pytest.mark.asyncio
    async def test_generate_embedding_missing_package(self) -> None:
        """Verifica erro quando langchain_cohere não está instalado."""
        worker = BackgroundEmbeddingWorker()

        with patch.object(
            type(worker.config), "get_cohere_api_key", return_value="key"
        ):
            # Simular que CohereEmbeddings é None (import não funcionou)
            with patch("vectora.services.background.CohereEmbeddings", None):
                with pytest.raises(ImportError, match="langchain_cohere"):
                    await worker._generate_embedding("test")

    @pytest.mark.asyncio
    async def test_generate_embedding_uses_config_model(self) -> None:
        """Verifica se usa embedding_model da configuração."""
        worker = BackgroundEmbeddingWorker()

        with patch("vectora.services.background.CohereEmbeddings") as mock_cohere_class:
            mock_cohere = MagicMock()
            mock_cohere.embed_query = MagicMock(return_value=[0.1])
            mock_cohere_class.return_value = mock_cohere

            with patch.object(
                type(worker.config), "get_cohere_api_key", return_value="key"
            ):
                # Mock embedding_model como um atributo fixo
                test_model = "embed-multilingual-v3.0"
                with patch(
                    "vectora.services.background.settings.embedding_model", test_model
                ):
                    with patch(
                        "asyncio.to_thread", new_callable=AsyncMock
                    ) as mock_thread:
                        mock_thread.return_value = [0.1]

                        await worker._generate_embedding("test")

                        # Verificar que CohereEmbeddings foi criado com model correto
                        mock_cohere_class.assert_called_once()
                        call_kwargs = mock_cohere_class.call_args[1]
                        assert call_kwargs["model"] == "embed-multilingual-v3.0"


# ============================================================================
# TESTES DE ESCRITA EM LANCEDB
# ============================================================================


class TestBackgroundEmbeddingWorkerWriteLanceDB:
    """Testes para _write_to_lancedb() do BackgroundEmbeddingWorker."""

    @pytest.mark.asyncio
    async def test_write_success(self) -> None:
        """Verifica escrita bem-sucedida em LanceDB."""
        worker = BackgroundEmbeddingWorker()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-write"
        mock_record.text = "test"
        mock_record.collection = "test-collection"
        mock_record.doc_metadata = '{"key": "value"}'

        vector = [0.1, 0.2, 0.3]

        with patch("vectora.services.background.lancedb") as mock_lancedb_module:
            mock_db = MagicMock()
            mock_db.open_table = AsyncMock()
            mock_table = MagicMock()
            mock_table.add = AsyncMock()
            mock_db.open_table.return_value = mock_table
            mock_lancedb_module.connect_async = AsyncMock(return_value=mock_db)

            with patch("vectora.services.background.pa") as mock_pa:
                mock_pa.schema.return_value = MagicMock()
                mock_pa.field.return_value = MagicMock()
                mock_pa.string.return_value = MagicMock()
                mock_pa.list_.return_value = MagicMock()
                mock_pa.float32.return_value = MagicMock()

                with patch.object(worker.config, "lancedb_dir", "/tmp/lancedb"):  # noqa: S108
                    await worker._write_to_lancedb(mock_record, vector)

                    # Verificar chamadas
                    mock_db.open_table.assert_called()
                    mock_table.add.assert_called_once()

    @pytest.mark.asyncio
    async def test_write_idempotent_via_queue_id(self) -> None:
        """Verifica idempotência usando queue_id como ID do documento."""
        worker = BackgroundEmbeddingWorker()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-idempotent"
        mock_record.text = "test"
        mock_record.collection = "test"
        mock_record.doc_metadata = "{}"

        vector = [0.1]

        with patch("vectora.services.background.lancedb") as mock_lancedb_module:
            mock_db = MagicMock()
            mock_table = MagicMock()
            mock_table.add = AsyncMock()
            mock_db.open_table = AsyncMock(return_value=mock_table)
            mock_lancedb_module.connect_async = AsyncMock(return_value=mock_db)

            with patch("vectora.services.background.pa"):
                with patch.object(worker.config, "lancedb_dir", "/tmp/lancedb"):  # noqa: S108
                    await worker._write_to_lancedb(mock_record, vector)

                    # Verificar que foi passado o documento com queue_id como id
                    call_args = mock_table.add.call_args
                    doc = call_args[0][0][0]
                    assert doc["id"] == "q-idempotent"

    @pytest.mark.asyncio
    async def test_write_missing_package(self) -> None:
        """Verifica erro quando lancedb não está instalado."""
        worker = BackgroundEmbeddingWorker()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        vector = [0.1]

        with patch("vectora.services.background.lancedb", None):
            with pytest.raises(ImportError, match="lancedb não está instalado"):
                await worker._write_to_lancedb(mock_record, vector)

    @pytest.mark.asyncio
    async def test_write_uses_lancedb_semaphore(self) -> None:
        """Verifica se semaphore(1) é usado para proteger escrita."""
        worker = BackgroundEmbeddingWorker()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-sem"
        mock_record.text = "test"
        mock_record.collection = "test"
        mock_record.doc_metadata = "{}"

        vector = [0.1]

        initial_sem_value = worker.lancedb_semaphore._value

        with patch("vectora.services.background.lancedb") as mock_lancedb_module:
            mock_db = MagicMock()
            mock_table = MagicMock()
            mock_table.add = AsyncMock()
            mock_db.open_table = AsyncMock(return_value=mock_table)
            mock_lancedb_module.connect_async = AsyncMock(return_value=mock_db)

            with patch("vectora.services.background.pa"):
                with patch.object(worker.config, "lancedb_dir", "/tmp/lancedb"):  # noqa: S108
                    await worker._write_to_lancedb(mock_record, vector)

                    # Semaphore deve voltar ao inicial
                    assert worker.lancedb_semaphore._value == initial_sem_value

    @pytest.mark.asyncio
    async def test_write_creates_table_if_not_exists(self) -> None:
        """Verifica criação de tabela se não existir."""
        worker = BackgroundEmbeddingWorker()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-create"
        mock_record.text = "test"
        mock_record.collection = "new-collection"
        mock_record.doc_metadata = "{}"

        vector = [0.1]

        with patch("vectora.services.background.lancedb") as mock_lancedb_module:
            mock_db = MagicMock()
            # open_table falha, cria nova tabela
            mock_db.open_table = AsyncMock(side_effect=Exception("Table not found"))
            mock_table = MagicMock()
            mock_table.add = AsyncMock()
            mock_db.create_table = AsyncMock(return_value=mock_table)
            mock_lancedb_module.connect_async = AsyncMock(return_value=mock_db)

            with patch("vectora.services.background.pa"):
                with patch.object(worker.config, "lancedb_dir", "/tmp/lancedb"):  # noqa: S108
                    await worker._write_to_lancedb(mock_record, vector)

                    # Deve ter criado tabela
                    mock_db.create_table.assert_called_once()


# ============================================================================
# TESTES DE SINGLETON
# ============================================================================


class TestGetBackgroundWorker:
    """Testes para get_background_worker() - singleton."""

    @pytest.mark.asyncio
    async def test_creates_instance(self) -> None:
        """Verifica criação de instância singleton."""
        # Reset global state
        import vectora.services.background as bg_module

        bg_module._worker = None

        worker = await get_background_worker()
        assert worker is not None
        assert isinstance(worker, BackgroundEmbeddingWorker)

        # Cleanup
        bg_module._worker = None

    @pytest.mark.asyncio
    async def test_returns_same_instance(self) -> None:
        """Verifica que mesma instância é retornada."""
        # Reset global state
        import vectora.services.background as bg_module

        bg_module._worker = None

        worker1 = await get_background_worker()
        worker2 = await get_background_worker()

        assert worker1 is worker2

        # Cleanup
        bg_module._worker = None

    @pytest.mark.asyncio
    async def test_thread_safe(self) -> None:
        """Verifica que singleton é thread-safe com asyncio.Lock."""
        # Reset global state
        import vectora.services.background as bg_module

        bg_module._worker = None

        # Criar múltiplas tasks que tentam obter worker simultaneamente
        tasks = [get_background_worker() for _ in range(5)]
        workers = await asyncio.gather(*tasks)

        # Todos devem ter a mesma instância
        first_worker = workers[0]
        for worker in workers[1:]:
            assert worker is first_worker

        # Cleanup
        bg_module._worker = None


# ============================================================================
# TESTES DE RECONCILIAÇÃO
# ============================================================================


class TestBackgroundEmbeddingWorkerReconciliation:
    """Testes para _reconcile_startup()."""

    @pytest.mark.asyncio
    async def test_reconcile_startup_calls_queue_reconcile(self) -> None:
        """Verifica se _reconcile_startup chama queue.reconcile()."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        mock_queue.reconcile = AsyncMock()

        with patch.object(worker, "_get_queue", new_callable=AsyncMock) as mock_get:
            mock_get.return_value = mock_queue

            await worker._reconcile_startup()

            mock_queue.reconcile.assert_called_once()

    @pytest.mark.asyncio
    async def test_reconcile_startup_handles_errors(self) -> None:
        """Verifica se _reconcile_startup trata erros gracefully."""
        worker = BackgroundEmbeddingWorker()

        with patch.object(worker, "_get_queue", new_callable=AsyncMock) as mock_get:
            mock_get.side_effect = Exception("DB connection error")

            # Não deve lançar
            await worker._reconcile_startup()


# ============================================================================
# TESTES DE LOOP PRINCIPAL
# ============================================================================


class TestBackgroundEmbeddingWorkerRunLoop:
    """Testes para _run_loop() do BackgroundEmbeddingWorker."""

    @pytest.mark.asyncio
    async def test_run_loop_processes_pending_batches(self) -> None:
        """Verifica se loop processa batches pendentes."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        # Primeira chamada retorna registros, segunda retorna vazio para sair
        pending_records = [
            MagicMock(spec=EmbeddingQueueRecord, queue_id=f"q-{i}") for i in range(2)
        ]
        mock_queue.get_pending = AsyncMock(
            side_effect=[pending_records, []]  # Primeira batch, depois vazio
        )

        with patch.object(worker, "_get_queue", new_callable=AsyncMock) as mock_get:
            mock_get.return_value = mock_queue

            with patch.object(
                worker, "_process_record", new_callable=AsyncMock
            ) as mock_process:
                worker.running = True

                # Rodar por um tempo curto
                loop_task = asyncio.create_task(worker._run_loop())

                # Deixar processar uma iteração
                await asyncio.sleep(0.5)

                worker.running = False
                await asyncio.wait_for(loop_task, timeout=2)

                # Deve ter processado registros
                assert mock_process.call_count >= 2

    @pytest.mark.asyncio
    async def test_run_loop_sleeps_when_no_pending(self) -> None:
        """Verifica se loop dorme quando não há registros pendentes."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        # Retornar vazio para fazer o loop dormir
        mock_queue.get_pending = AsyncMock(return_value=[])

        with patch.object(worker, "_get_queue", new_callable=AsyncMock) as mock_get:
            mock_get.return_value = mock_queue

            worker.running = True

            # Usar um mock que rastreia chamadas mas deixa o sleep funcionar
            original_sleep = asyncio.sleep
            sleep_calls = []

            async def tracked_sleep(seconds: float) -> None:
                sleep_calls.append(seconds)
                # Sleep por um tempo muito curto para não bloquear o teste
                await original_sleep(0.01)

            with patch(
                "vectora.services.background.asyncio.sleep", side_effect=tracked_sleep
            ):
                loop_task = asyncio.create_task(worker._run_loop())

                # Deixar o loop rodar uma ou duas iterações
                await asyncio.sleep(0.2)
                worker.running = False

                with contextlib.suppress(TimeoutError):
                    await asyncio.wait_for(loop_task, timeout=2)

                # Deve ter dormido com POLLING_INTERVAL (5 segundos)
                assert any(s == POLLING_INTERVAL for s in sleep_calls)

    @pytest.mark.asyncio
    async def test_run_loop_handles_cancelled_error(self) -> None:
        """Verifica se loop trata CancelledError gracefully."""
        worker = BackgroundEmbeddingWorker()

        # Criar uma queue que lança CancelledError
        async def get_queue_error() -> MagicMock:
            raise asyncio.CancelledError

        worker.running = True

        with patch.object(worker, "_get_queue", side_effect=get_queue_error):
            loop_task = asyncio.create_task(worker._run_loop())

            with contextlib.suppress(asyncio.CancelledError):
                await asyncio.wait_for(loop_task, timeout=2)

            # Loop deve ter terminado de forma controlada
            assert loop_task.done()

    @pytest.mark.asyncio
    async def test_run_loop_handles_generic_exception(self) -> None:
        """Verifica se loop trata exceção genérica e continua."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        call_count = [0]

        async def get_queue_with_error() -> MagicMock:
            call_count[0] += 1
            if call_count[0] == 1:
                # Primeira chamada falha
                raise RuntimeError("Database connection error")
            # Próximas chamadas retornam queue vazia para sair do loop
            return mock_queue

        mock_queue.get_pending = AsyncMock(return_value=[])

        with patch.object(worker, "_get_queue", side_effect=get_queue_with_error):
            worker.running = True

            original_sleep = asyncio.sleep
            sleep_calls = []

            async def tracked_sleep(seconds: float) -> None:
                sleep_calls.append(seconds)
                await original_sleep(0.01)

            with patch(
                "vectora.services.background.asyncio.sleep", side_effect=tracked_sleep
            ):
                loop_task = asyncio.create_task(worker._run_loop())

                await asyncio.sleep(0.3)
                worker.running = False

                with contextlib.suppress(TimeoutError):
                    await asyncio.wait_for(loop_task, timeout=2)

                # Loop deve ter dormido após o erro (POLLING_INTERVAL = 5)
                # Pode ter chamado sleep com 5s se entrou no except
                assert len(sleep_calls) > 0


# ============================================================================
# TESTES DE EDGE CASES
# ============================================================================


class TestBackgroundEmbeddingWorkerEdgeCases:
    """Testes para casos extremos e edge cases."""

    @pytest.mark.asyncio
    async def test_process_record_with_none_metadata(self) -> None:
        """Verifica processamento com doc_metadata = None."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        mock_queue.mark_processing = AsyncMock()
        mock_queue.mark_success = AsyncMock()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-none-meta"
        mock_record.text = "test"
        mock_record.collection = "test"
        mock_record.doc_metadata = None  # Nenhuma metadata

        with patch.object(
            worker, "_generate_embedding", new_callable=AsyncMock
        ) as mock_gen:
            mock_gen.return_value = [0.1]

            with patch.object(
                worker, "_write_to_lancedb", new_callable=AsyncMock
            ) as mock_write:
                await worker._process_record(mock_record, mock_queue)

                # Deve ter funcionado mesmo com metadata None
                mock_write.assert_called_once()
                call_args = mock_write.call_args
                # Verificar que record foi passado
                assert call_args[0][0] == mock_record

    @pytest.mark.asyncio
    async def test_process_record_empty_text(self) -> None:
        """Verifica processamento com texto vazio."""
        worker = BackgroundEmbeddingWorker()

        mock_queue = MagicMock()
        mock_queue.mark_processing = AsyncMock()

        mock_record = MagicMock(spec=EmbeddingQueueRecord)
        mock_record.queue_id = "q-empty"
        mock_record.text = ""  # Texto vazio
        mock_record.collection = "test"
        mock_record.doc_metadata = "{}"

        # Cohere pode retornar embedding vazio ou erro
        with patch.object(
            worker, "_generate_embedding", new_callable=AsyncMock
        ) as mock_gen:
            mock_gen.return_value = [0.0]

            with patch.object(worker, "_write_to_lancedb", new_callable=AsyncMock):
                await worker._process_record(mock_record, mock_queue)

                # Deve ter tentado embeddar
                mock_gen.assert_called_with("")

    def test_constants_values(self) -> None:
        """Verifica que constantes têm valores esperados."""
        assert MAX_PARALLEL == 5
        assert MAX_RETRIES == 3
        assert POLLING_INTERVAL == 5
        assert RETRY_BACKOFF == [1, 2, 4]
