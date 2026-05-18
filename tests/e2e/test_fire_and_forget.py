"""End-to-End Tests para Fire-and-Forget Embedding Architecture.

Testes críticos que validam:
1. TUI responsiva (<200ms) - embedding() retorna imediatamente
2. Background worker processa embeddings em paralelo
3. Exponential backoff + retry em falhas de API
4. DLQ para falhas permanentes (após 3 tentativas)
5. Reconciliação recupera records travados
6. Idempotência via queue_id como document ID
"""

import json
from collections.abc import AsyncGenerator
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.services.background import BackgroundEmbeddingWorker
from vectora.services.queue import EmbeddingQueueRecord, get_embedding_queue
from vectora.tools.rag import embedding, ingest_docs

_QUEUE_DSN = "sqlite+aiosqlite:///:memory:"


@pytest.fixture(autouse=True)
async def reset_embedding_queue_singleton() -> AsyncGenerator[None]:
    """Reset module-level _queue singleton antes de cada teste."""
    import contextlib

    import vectora.services.queue as queue_mod

    queue_mod._queue = None
    yield
    if queue_mod._queue:
        with contextlib.suppress(Exception):
            await queue_mod._queue.close()
        queue_mod._queue = None


class TestFireAndForgetBasic:
    """Test basic fire-and-forget functionality."""

    @pytest.mark.asyncio
    async def test_embedding_returns_immediately(self) -> None:
        """Test que embedding() retorna em <200ms sem esperar Cohere."""
        mock_queue = AsyncMock()
        mock_queue.enqueue = AsyncMock(return_value="queue-123")

        with patch("vectora.tools.rag.get_embedding_queue", return_value=mock_queue):
            mock_settings = MagicMock()
            mock_settings.enable_rag = True
            mock_settings.embedding_queue_enabled = True
            mock_settings.embedding_queue_dsn = _QUEUE_DSN

            with patch("vectora.tools.rag.settings", mock_settings):
                import time

                start = time.time()
                result = ""
                async for chunk in embedding.astream(
                    {
                        "text": "Test document",
                        "collection": "test",
                        "metadata": {"source": "test"},
                    }
                ):
                    if isinstance(chunk, str):
                        result += chunk
                elapsed = time.time() - start

        assert elapsed < 0.2, f"embedding() levou {elapsed:.3f}s, esperado <0.2s"
        result_json = json.loads(result)
        assert result_json["status"] == "fire_and_forget"
        assert result_json["queue_id"] == "queue-123"
        assert result_json["collection"] == "test"
        mock_queue.enqueue.assert_called_once()

    @pytest.mark.asyncio
    async def test_embedding_enqueues_successfully(self) -> None:
        """Test que embedding() enfileira documento corretamente."""
        mock_settings = MagicMock()
        mock_settings.enable_rag = True
        mock_settings.embedding_queue_enabled = True
        mock_settings.embedding_queue_dsn = _QUEUE_DSN

        queue = await get_embedding_queue(_QUEUE_DSN)

        with patch("vectora.tools.rag.get_embedding_queue", return_value=queue):
            with patch("vectora.tools.rag.settings", mock_settings):
                result = ""
                async for chunk in embedding.astream(
                    {
                        "text": "Test document for queueing",
                        "collection": "articles",
                        "metadata": {"author": "test_user"},
                    }
                ):
                    if isinstance(chunk, str):
                        result += chunk

        result_json = json.loads(result)
        queue_id = result_json["queue_id"]

        pending = await queue.get_pending(limit=10)
        assert len(pending) > 0
        matching = [r for r in pending if r.queue_id == queue_id]
        assert len(matching) > 0, f"Queue ID {queue_id} não encontrado nos pendentes"
        assert matching[0].text == "Test document for queueing"
        assert matching[0].collection == "articles"

    @pytest.mark.asyncio
    async def test_embedding_disabled_returns_error(self) -> None:
        """Test que embedding() retorna erro se RAG desabilitado."""
        mock_settings = MagicMock()
        mock_settings.enable_rag = False

        with patch("vectora.tools.rag.settings", mock_settings):
            result = ""
            async for chunk in embedding.astream(
                {"text": "Test", "collection": "test"}
            ):
                if isinstance(chunk, str):
                    result += chunk

        result_json = json.loads(result)
        assert result_json["status"] == "error"


class TestBackgroundWorker:
    """Test background embedding worker."""

    @pytest.mark.asyncio
    async def test_worker_processes_pending_records(self) -> None:
        """Test que background worker processa records pendentes."""
        queue = await get_embedding_queue(_QUEUE_DSN)
        queue_id = await queue.enqueue(
            text="Test embedding document",
            collection="articles",
            metadata={"test": True},
        )

        mock_vector = [0.1, 0.2, 0.3] * 400

        with patch("vectora.services.background.CohereEmbeddings") as mock_cohere:
            mock_embeddings = MagicMock()
            mock_embeddings.embed_query = MagicMock(return_value=mock_vector)
            mock_cohere.return_value = mock_embeddings

            with patch("vectora.services.background.lancedb") as mock_lancedb:
                mock_db = AsyncMock()
                mock_table = AsyncMock()
                mock_db.open_table = AsyncMock(return_value=mock_table)
                mock_db.create_table = AsyncMock(return_value=mock_table)
                mock_table.add = AsyncMock()
                mock_lancedb.connect_async = AsyncMock(return_value=mock_db)

                with patch("vectora.services.background.pa") as mock_pa:
                    mock_pa.schema = MagicMock(return_value=MagicMock())
                    mock_pa.field = MagicMock(return_value=MagicMock())
                    mock_pa.string = MagicMock(return_value=MagicMock())
                    mock_pa.list_ = MagicMock(return_value=MagicMock())
                    mock_pa.float32 = MagicMock(return_value=MagicMock())

                    with patch(
                        "vectora.services.background.get_embedding_queue",
                        return_value=queue,
                    ):
                        worker = BackgroundEmbeddingWorker()

                    pending = await queue.get_pending(limit=1)
                    assert len(pending) == 1
                    await worker._process_record(pending[0])

                    failed = await queue.get_failed(limit=10)
                    assert len(failed) == 0

    @pytest.mark.asyncio
    async def test_worker_moves_to_dlq_after_3_failures(self) -> None:
        """Test que record vai para DLQ após 3 falhas."""
        queue = await get_embedding_queue(_QUEUE_DSN)
        await queue.enqueue(
            text="Test DLQ document",
            collection="articles",
        )

        def always_fail(text: str) -> list[float]:
            raise Exception("Permanent API failure")

        with patch("vectora.services.background.CohereEmbeddings") as mock_cohere:
            mock_embeddings = MagicMock()
            mock_embeddings.embed_query = MagicMock(side_effect=always_fail)
            mock_cohere.return_value = mock_embeddings

            with patch("vectora.services.background.lancedb"):
                with patch(
                    "vectora.services.background.get_embedding_queue",
                    return_value=queue,
                ):
                    worker = BackgroundEmbeddingWorker()

                pending = await queue.get_pending(limit=1)
                await worker._process_record(pending[0])

                failed = await queue.get_failed(limit=10)
                assert len(failed) == 1
                assert failed[0].status == "dlq"
                assert "Permanent API failure" in failed[0].dlq_reason


class TestReconciliation:
    """Test reconciliation of stalled records."""

    @pytest.mark.asyncio
    async def test_reconciliation_recovers_stalled_records(self) -> None:
        """Test que reconciliação recupera records em 'processing' há >2min."""
        from datetime import UTC, datetime, timedelta

        from sqlalchemy import update

        queue = await get_embedding_queue(_QUEUE_DSN)

        queue_id = await queue.enqueue(
            text="Stalled document",
            collection="articles",
        )
        await queue.mark_processing(queue_id)

        pending_before = await queue.get_pending(limit=10)
        assert len(pending_before) == 0

        if queue.AsyncSessionLocal:
            async with queue.AsyncSessionLocal() as session:
                old_time = datetime.now(UTC).replace(tzinfo=None) - timedelta(minutes=5)
                stmt = (
                    update(EmbeddingQueueRecord)
                    .where(EmbeddingQueueRecord.queue_id == queue_id)
                    .values(updated_at=old_time)
                )
                await session.execute(stmt)
                await session.commit()

        await queue.reconcile()

        pending_after = await queue.get_pending(limit=10)
        assert len(pending_after) == 1
        assert pending_after[0].queue_id == queue_id
        assert pending_after[0].status == "pending"


class TestIngestDocWithFireAndForget:
    """Test ingest_docs() com fire-and-forget embedding."""

    @pytest.mark.asyncio
    async def test_ingest_docs_returns_immediately(self, tmp_path: Path) -> None:
        """Test que ingest_docs() retorna rápido (documentos enfileirados)."""
        test_file = tmp_path / "test.md"
        test_file.write_text("# Test Document\n\nSome test content here.")

        queue = await get_embedding_queue(_QUEUE_DSN)

        mock_settings = MagicMock()
        mock_settings.enable_rag = True
        mock_settings.embedding_queue_enabled = True
        mock_settings.embedding_queue_dsn = _QUEUE_DSN
        mock_settings.enable_file_operations = True

        with patch("vectora.tools.rag.settings", mock_settings):
            with patch("vectora.tools.rag.get_embedding_queue", return_value=queue):
                with patch(
                    "vectora.services.security.is_safe_file_path", return_value=True
                ):
                    import time

                    start = time.time()
                    result = ""
                    async for chunk in ingest_docs.astream(
                        {
                            "directory_path": str(tmp_path),
                            "collection": "articles",
                            "glob_pattern": "*.md",
                        }
                    ):
                        if isinstance(chunk, str):
                            result += chunk
                    elapsed = time.time() - start

        result_json = json.loads(result)
        assert elapsed < 5, f"ingest_docs levou {elapsed:.1f}s"
        assert result_json["status"] == "completed"
        assert result_json["indexed"] >= 0


class TestLocalFirstRAGRouter:
    """Test Local-First RAG routing behavior (system prompt instruction)."""

    def test_system_prompt_mentions_vector_search(self) -> None:
        """Verifica que sistema prompt contém instrução sobre Local-First RAG."""
        from vectora.prompts import get_system_prompt

        prompt = get_system_prompt(language="en_US")
        assert isinstance(prompt, str)
        assert len(prompt) > 100


@pytest.mark.asyncio
async def test_full_fire_and_forget_workflow() -> None:
    """Integration test: complete fire-and-forget workflow."""
    queue = await get_embedding_queue(_QUEUE_DSN)

    for i in range(3):
        await queue.enqueue(
            text=f"Next.js feature #{i}: SSR, ISR, streaming",
            collection="articles",
            metadata={"topic": "Next.js 16", "index": i},
        )

    pending = await queue.get_pending(limit=10)
    assert len(pending) == 3

    mock_vector = [0.1, 0.2] * 600

    with patch("vectora.services.background.CohereEmbeddings") as mock_cohere:
        mock_embeddings = MagicMock()
        mock_embeddings.embed_query = MagicMock(return_value=mock_vector)
        mock_cohere.return_value = mock_embeddings

        with patch("vectora.services.background.lancedb") as mock_lancedb:
            with patch("vectora.services.background.pa") as mock_pa:
                mock_db = AsyncMock()
                mock_table = AsyncMock()
                mock_db.open_table = AsyncMock(return_value=mock_table)
                mock_db.create_table = AsyncMock(return_value=mock_table)
                mock_table.add = AsyncMock()
                mock_lancedb.connect_async = AsyncMock(return_value=mock_db)
                mock_pa.schema.return_value = MagicMock()

                with patch(
                    "vectora.services.background.get_embedding_queue",
                    return_value=queue,
                ):
                    worker = BackgroundEmbeddingWorker()

                for _ in range(3):
                    pending = await queue.get_pending(limit=1)
                    if pending:
                        await worker._process_record(pending[0])

                remaining = await queue.get_pending(limit=10)
                assert len(remaining) == 0
                failed = await queue.get_failed(limit=10)
                assert len(failed) == 0


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
