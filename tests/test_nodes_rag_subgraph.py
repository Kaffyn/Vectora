"""Tests for vectora/nodes/rag_subgraph.py"""

from __future__ import annotations

from typing import Any
from unittest.mock import AsyncMock, patch

import pytest
from langchain_core.messages import HumanMessage, SystemMessage

from vectora.nodes.rag_subgraph import (
    _best_score,
    _extract_query,
    rag_decide,
    rag_inject,
    rag_retrieve,
    rag_websearch,
)
from vectora.state import Document, State

_SCORE_HIGH = 0.7
_SCORE_LOW = 0.4


def _state(**kw: Any) -> State:
    base: State = {
        "messages": [HumanMessage(content="como funciona o JWT?")],
        "session_metadata": {},
    }
    base.update(kw)
    return base


def _doc(score=0.8, content="doc") -> Document:
    return Document(
        page_content=content, metadata={"source": "test"}, relevance_score=score
    )


class TestExtractQuery:
    def test_extracts_last_human(self):
        s = _state(messages=[HumanMessage(content="a"), HumanMessage(content="b")])
        assert _extract_query(s) == "b"

    def test_empty_messages(self):
        assert _extract_query(_state(messages=[])) == ""

    def test_no_human_message(self):
        from langchain_core.messages import AIMessage

        assert _extract_query(_state(messages=[AIMessage(content="ai")])) == ""


class TestBestScore:
    def test_max_score(self):
        assert _best_score([_doc(0.3), _doc(0.9)]) == pytest.approx(0.9)

    def test_empty(self):
        assert _best_score([]) == 0.0

    def test_none_scores(self):
        assert _best_score([_doc(None), _doc(0.6)]) == pytest.approx(0.6)

    def test_all_none(self):
        assert _best_score([_doc(None)]) == 0.0


class TestRagDecide:
    def test_high_score_inject(self):
        assert rag_decide(_state(rag_docs=[_doc(_SCORE_HIGH)])) == "rag_inject"

    def test_medium_score_rerank(self):
        assert rag_decide(_state(rag_docs=[_doc(0.55)])) == "rag_rerank"

    def test_low_score_websearch(self):
        assert rag_decide(_state(rag_docs=[_doc(0.2)])) == "rag_websearch"

    def test_empty_docs_websearch(self):
        assert rag_decide(_state(rag_docs=[])) == "rag_websearch"

    def test_exactly_high_threshold(self):
        assert rag_decide(_state(rag_docs=[_doc(_SCORE_HIGH)])) == "rag_inject"

    def test_exactly_low_threshold(self):
        assert rag_decide(_state(rag_docs=[_doc(_SCORE_LOW)])) == "rag_rerank"


class TestRagRetrieve:
    @pytest.mark.asyncio
    async def test_returns_docs(self):
        docs = [_doc(0.9, "c1"), _doc(0.7, "c2")]
        with patch(
            "vectora.nodes.rag_subgraph._call_vector_search", new_callable=AsyncMock
        ) as m:
            m.return_value = docs
            result = await rag_retrieve(_state())
        assert result["rag_query"] == "como funciona o JWT?"
        assert len(result["rag_docs"]) == 2

    @pytest.mark.asyncio
    async def test_empty_query_returns_early(self):
        result = await rag_retrieve(_state(messages=[]))
        assert result["rag_docs"] == []
        assert result["rag_query"] == ""

    @pytest.mark.asyncio
    async def test_no_results(self):
        with patch(
            "vectora.nodes.rag_subgraph._call_vector_search", new_callable=AsyncMock
        ) as m:
            m.return_value = []
            result = await rag_retrieve(_state())
        assert result["rag_docs"] == []


class TestRagWebsearch:
    @pytest.mark.asyncio
    async def test_adds_web_docs(self):
        web = [{"content": "web content", "url": "https://a.com", "title": "A"}]
        with patch(
            "vectora.nodes.rag_subgraph._call_web_search", new_callable=AsyncMock
        ) as m:
            with patch(
                "vectora.nodes.rag_subgraph._enqueue_for_embedding",
                new_callable=AsyncMock,
            ) as me:
                with patch("vectora.nodes.rag_subgraph.settings") as ms:
                    m.return_value = web
                    me.return_value = "qid"
                    ms.enable_rag = True
                    ms.embedding_queue_enabled = True
                    result = await rag_websearch(_state(rag_query="JWT", rag_docs=[]))
        assert result["web_search_triggered"] is True
        assert len(result["rag_docs"]) == 1

    @pytest.mark.asyncio
    async def test_skips_empty_content(self):
        web = [{"content": "", "url": "https://a.com", "title": "Empty"}]
        with patch(
            "vectora.nodes.rag_subgraph._call_web_search", new_callable=AsyncMock
        ) as m:
            with patch("vectora.nodes.rag_subgraph.settings") as ms:
                m.return_value = web
                ms.enable_rag = True
                ms.embedding_queue_enabled = False
                result = await rag_websearch(_state(rag_query="test", rag_docs=[]))
        assert result["rag_docs"] == []

    @pytest.mark.asyncio
    async def test_combines_existing_docs(self):
        existing = [_doc(0.3, "existing")]
        web = [{"content": "web doc", "url": "https://a.com", "title": "A"}]
        with patch(
            "vectora.nodes.rag_subgraph._call_web_search", new_callable=AsyncMock
        ) as m:
            with patch(
                "vectora.nodes.rag_subgraph._enqueue_for_embedding",
                new_callable=AsyncMock,
            ) as me:
                with patch("vectora.nodes.rag_subgraph.settings") as ms:
                    m.return_value = web
                    me.return_value = None
                    ms.enable_rag = True
                    ms.embedding_queue_enabled = True
                    result = await rag_websearch(
                        _state(rag_query="test", rag_docs=existing)
                    )
        assert len(result["rag_docs"]) == 2


class TestRagInject:
    @pytest.mark.asyncio
    async def test_injects_system_message(self):
        docs = [_doc(0.9, "conteúdo JWT")]
        result = await rag_inject(_state(rag_docs=docs, rag_query="JWT"))
        assert "messages" in result
        assert isinstance(result["messages"][0], SystemMessage)
        assert "JWT" in result["messages"][0].content
        assert "conteúdo JWT" in result["messages"][0].content

    @pytest.mark.asyncio
    async def test_no_docs_returns_empty(self):
        assert await rag_inject(_state(rag_docs=[])) == {}

    @pytest.mark.asyncio
    async def test_truncates_to_5_docs(self):
        docs = [_doc(0.8, f"doc {i}") for i in range(10)]
        result = await rag_inject(_state(rag_docs=docs, rag_query="test"))
        assert "doc 4" in result["messages"][0].content
        assert "doc 5" not in result["messages"][0].content
