"""Tests for vectora/nodes/engine.py"""

from __future__ import annotations

import json
from unittest.mock import AsyncMock, MagicMock

import pytest

from vectora.nodes.engine import _extract_tavily_results, _process_tavily_results


class TestExtractTavilyResults:
    def test_dict_with_results_key(self):
        data = {"results": [{"content": "a"}, {"content": "b"}]}
        assert _extract_tavily_results(data, "web_search") == data["results"]

    def test_list_input(self):
        data = [{"content": "x"}]
        assert _extract_tavily_results(data, "web_search") == data

    def test_empty_dict_returns_empty_list(self):
        result = _extract_tavily_results({}, "web_search")
        assert result == []

    def test_invalid_type_returns_none(self):
        result = _extract_tavily_results("invalid", "web_search")  # type: ignore[arg-type]
        assert result is None


class TestProcessTavilyResults:
    @pytest.mark.asyncio
    async def test_formats_and_enqueues(self):
        results = [{"content": "doc content", "title": "Title", "url": "https://a.com"}]
        mock_embedding = AsyncMock()
        mock_embedding.ainvoke.return_value = json.dumps({"queue_id": "qid-1"})

        docs, queue_ids = await _process_tavily_results(
            results, "web_search", mock_embedding
        )

        assert len(docs) == 1
        assert docs[0]["page_content"] == "doc content"
        assert docs[0]["metadata"]["url"] == "https://a.com"
        assert "qid-1" in queue_ids

    @pytest.mark.asyncio
    async def test_skips_empty_content(self):
        results = [{"content": "", "title": "Empty", "url": "https://a.com"}]
        mock_embedding = AsyncMock()

        docs, queue_ids = await _process_tavily_results(
            results, "web_search", mock_embedding
        )

        assert docs == []
        assert queue_ids == []
        mock_embedding.ainvoke.assert_not_called()

    @pytest.mark.asyncio
    async def test_embedding_failure_skips_queue_id(self):
        results = [{"content": "valid content", "title": "T", "url": "https://b.com"}]
        mock_embedding = AsyncMock()
        mock_embedding.ainvoke.side_effect = Exception("API error")

        docs, queue_ids = await _process_tavily_results(
            results, "web_search", mock_embedding
        )

        assert len(docs) == 1  # doc ainda formatado
        assert queue_ids == []  # mas sem queue_id

    @pytest.mark.asyncio
    async def test_multiple_results(self):
        results = [
            {"content": "doc 1", "title": "A", "url": "https://a.com"},
            {"content": "doc 2", "title": "B", "url": "https://b.com"},
        ]
        mock_embedding = AsyncMock()
        mock_embedding.ainvoke.return_value = json.dumps({"queue_id": "q1"})

        docs, queue_ids = await _process_tavily_results(
            results, "web_search", mock_embedding
        )

        assert len(docs) == 2
        assert len(queue_ids) == 2
