"""Tests for vectora/services/queue.py"""

from __future__ import annotations

import pytest

from vectora.services.queue import get_embedding_queue

_DSN = "sqlite+aiosqlite:///:memory:"


@pytest.fixture
async def queue():
    return await get_embedding_queue(_DSN)


class TestEmbeddingQueue:
    @pytest.mark.asyncio
    async def test_enqueue_returns_id(self, queue):
        qid = await queue.enqueue("texto", "articles", {})
        assert isinstance(qid, str)
        assert len(qid) > 0

    @pytest.mark.asyncio
    async def test_get_pending_after_enqueue(self, queue):
        await queue.enqueue("texto pendente", "articles", {})
        items = await queue.get_pending(limit=10)
        assert len(items) >= 1
        texts = [i.text for i in items]
        assert "texto pendente" in texts

    @pytest.mark.asyncio
    async def test_mark_processing_then_success(self, queue):
        qid = await queue.enqueue("texto done", "articles", {})
        await queue.mark_processing(qid)
        await queue.mark_success(qid)
        pending = await queue.get_pending(limit=10)
        ids = [i.id for i in pending]
        assert qid not in ids

    @pytest.mark.asyncio
    async def test_mark_failed(self, queue):
        qid = await queue.enqueue("texto failed", "articles", {})
        await queue.mark_processing(qid)
        await queue.mark_failed(qid, "test error")
        pending = await queue.get_pending(limit=10)
        ids = [i.id for i in pending]
        assert qid not in ids

    @pytest.mark.asyncio
    async def test_get_stats(self, queue):
        stats = await queue.get_stats()
        assert isinstance(stats, dict)

    @pytest.mark.asyncio
    async def test_count_pending(self, queue):
        count = await queue.count_pending()
        assert isinstance(count, int)
        assert count >= 0
