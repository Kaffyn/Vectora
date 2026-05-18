"""Testes para vectora/nodes/rag_subgraph.py"""

from __future__ import annotations

import json
from unittest.mock import AsyncMock, MagicMock, patch

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

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_state(**kwargs: object) -> State:
    base: State = {
        "messages": [HumanMessage(content="como funciona o JWT?")],
        "session_metadata": {},
    }
    base.update(kwargs)
    return base


def _make_doc(score: float | None = 0.8, content: str = "doc content") -> Document:
    return Document(
        page_content=content,
        metadata={"source": "test"},
        relevance_score=score,
    )


# ---------------------------------------------------------------------------
# _extract_query
# ---------------------------------------------------------------------------


class TestExtractQuery:
    def test_extracts_last_human_message(self):
        state = _make_state(
            messages=[
                HumanMessage(content="primeira"),
                HumanMessage(content="segunda"),
            ]
        )
        assert _extract_query(state) == "segunda"

    def test_returns_empty_if_no_human_message(self):
        from langchain_core.messages import AIMessage

        state = _make_state(messages=[AIMessage(content="ai msg")])
        assert _extract_query(state) == ""

    def test_returns_empty_if_no_messages(self):
        state = _make_state(messages=[])
        assert _extract_query(state) == ""


# ---------------------------------------------------------------------------
# _best_score
# ---------------------------------------------------------------------------


class TestBestScore:
    def test_returns_max_score(self):
        docs = [_make_doc(0.3), _make_doc(0.9), _make_doc(0.5)]
        assert _best_score(docs) == pytest.approx(0.9)

    def test_returns_zero_for_empty(self):
        assert _best_score([]) == 0.0

    def test_ignores_none_scores(self):
        docs = [_make_doc(None), _make_doc(0.6)]
        assert _best_score(docs) == pytest.approx(0.6)

    def test_all_none_returns_zero(self):
        docs = [_make_doc(None), _make_doc(None)]
        assert _best_score(docs) == 0.0


# ---------------------------------------------------------------------------
# rag_decide
# ---------------------------------------------------------------------------


class TestRagDecide:
    def test_high_score_goes_to_rag_inject(self):
        state = _make_state(rag_docs=[_make_doc(0.85)])
        assert rag_decide(state) == "rag_inject"

    def test_medium_score_goes_to_rag_rerank(self):
        state = _make_state(rag_docs=[_make_doc(0.55)])
        assert rag_decide(state) == "rag_rerank"

    def test_low_score_goes_to_rag_websearch(self):
        state = _make_state(rag_docs=[_make_doc(0.2)])
        assert rag_decide(state) == "rag_websearch"

    def test_empty_docs_goes_to_rag_websearch(self):
        state = _make_state(rag_docs=[])
        assert rag_decide(state) == "rag_websearch"

    def test_none_docs_goes_to_rag_websearch(self):
        state = _make_state(rag_docs=None)
        assert rag_decide(state) == "rag_websearch"

    def test_exactly_threshold_high(self):
        state = _make_state(rag_docs=[_make_doc(0.7)])
        assert rag_decide(state) == "rag_inject"

    def test_exactly_threshold_low(self):
        state = _make_state(rag_docs=[_make_doc(0.4)])
        assert rag_decide(state) == "rag_rerank"


# ---------------------------------------------------------------------------
# rag_retrieve
# ---------------------------------------------------------------------------


class TestRagRetrieve:
    @pytest.mark.asyncio
    async def test_retrieve_returns_docs_on_success(self):
        mock_result = json.dumps(
            {
                "status": "success",
                "results": [
                    {"content": "doc1", "metadata": {}, "relevance_score": 0.9},
                    {"content": "doc2", "metadata": {}, "relevance_score": 0.7},
                ],
            }
        )

        with patch(
            "vectora.nodes.rag_subgraph._call_vector_search", new_callable=AsyncMock
        ) as mock_vs:
            mock_vs.return_value = [
                Document(page_content="doc1", metadata={}, relevance_score=0.9),
                Document(page_content="doc2", metadata={}, relevance_score=0.7),
            ]

            state = _make_state()
            result = await rag_retrieve(state)

            assert "rag_query" in result
            assert result["rag_query"] == "como funciona o JWT?"
            assert len(result["rag_docs"]) == 2

    @pytest.mark.asyncio
    async def test_retrieve_returns_empty_on_no_results(self):
        with patch(
            "vectora.nodes.rag_subgraph._call_vector_search", new_callable=AsyncMock
        ) as mock_vs:
            mock_vs.return_value = []

            state = _make_state()
            result = await rag_retrieve(state)

            assert result["rag_docs"] == []

    @pytest.mark.asyncio
    async def test_retrieve_empty_query_returns_early(self):
        state = _make_state(messages=[])
        result = await rag_retrieve(state)

        assert result["rag_query"] == ""
        assert result["rag_docs"] == []


# ---------------------------------------------------------------------------
# rag_websearch
# ---------------------------------------------------------------------------


class TestRagWebsearch:
    @pytest.mark.asyncio
    async def test_websearch_adds_web_docs(self):
        web_results = [
            {
                "content": "web content 1",
                "url": "https://example.com",
                "title": "Example",
            },
            {"content": "web content 2", "url": "https://other.com", "title": "Other"},
        ]

        with patch(
            "vectora.nodes.rag_subgraph._call_web_search", new_callable=AsyncMock
        ) as mock_ws:
            with patch(
                "vectora.nodes.rag_subgraph._enqueue_for_embedding",
                new_callable=AsyncMock,
            ) as mock_emb:
                mock_ws.return_value = web_results
                mock_emb.return_value = "queue-id-123"

                with patch("vectora.nodes.rag_subgraph.settings") as mock_settings:
                    mock_settings.enable_rag = True
                    mock_settings.embedding_queue_enabled = True

                    state = _make_state(rag_query="JWT autenticação", rag_docs=[])
                    result = await rag_websearch(state)

                    assert result["web_search_triggered"] is True
                    assert len(result["rag_docs"]) == 2
                    assert result["rag_docs"][0]["page_content"] == "web content 1"

    @pytest.mark.asyncio
    async def test_websearch_tracks_queue_ids(self):
        web_results = [{"content": "content", "url": "https://a.com", "title": "A"}]

        with patch(
            "vectora.nodes.rag_subgraph._call_web_search", new_callable=AsyncMock
        ) as mock_ws:
            with patch(
                "vectora.nodes.rag_subgraph._enqueue_for_embedding",
                new_callable=AsyncMock,
            ) as mock_emb:
                mock_ws.return_value = web_results
                mock_emb.return_value = "queue-abc"

                with patch("vectora.nodes.rag_subgraph.settings") as mock_settings:
                    mock_settings.enable_rag = True
                    mock_settings.embedding_queue_enabled = True

                    state = _make_state(rag_query="test")
                    result = await rag_websearch(state)

                    assert "queue-abc" in result["pending_embeds"]

    @pytest.mark.asyncio
    async def test_websearch_combines_with_existing_docs(self):
        existing = [_make_doc(0.3, "existing doc")]
        web_results = [{"content": "web doc", "url": "https://a.com", "title": "A"}]

        with patch(
            "vectora.nodes.rag_subgraph._call_web_search", new_callable=AsyncMock
        ) as mock_ws:
            with patch(
                "vectora.nodes.rag_subgraph._enqueue_for_embedding",
                new_callable=AsyncMock,
            ) as mock_emb:
                mock_ws.return_value = web_results
                mock_emb.return_value = None

                with patch("vectora.nodes.rag_subgraph.settings") as mock_settings:
                    mock_settings.enable_rag = True
                    mock_settings.embedding_queue_enabled = True

                    state = _make_state(rag_query="test", rag_docs=existing)
                    result = await rag_websearch(state)

                    # Existing + web = 2 docs
                    assert len(result["rag_docs"]) == 2

    @pytest.mark.asyncio
    async def test_websearch_skips_empty_content(self):
        web_results = [{"content": "", "url": "https://a.com", "title": "Empty"}]

        with patch(
            "vectora.nodes.rag_subgraph._call_web_search", new_callable=AsyncMock
        ) as mock_ws:
            with patch("vectora.nodes.rag_subgraph.settings") as mock_settings:
                mock_ws.return_value = web_results
                mock_settings.enable_rag = True
                mock_settings.embedding_queue_enabled = True

                state = _make_state(rag_query="test", rag_docs=[])
                result = await rag_websearch(state)

                assert result["rag_docs"] == []


# ---------------------------------------------------------------------------
# rag_inject
# ---------------------------------------------------------------------------


class TestRagInject:
    @pytest.mark.asyncio
    async def test_inject_adds_system_message(self):
        docs = [
            _make_doc(0.9, "conteúdo relevante sobre JWT"),
        ]
        state = _make_state(rag_docs=docs, rag_query="JWT")

        result = await rag_inject(state)

        assert "messages" in result
        assert len(result["messages"]) == 1
        msg = result["messages"][0]
        assert isinstance(msg, SystemMessage)
        assert "JWT" in msg.content
        assert "conteúdo relevante sobre JWT" in msg.content

    @pytest.mark.asyncio
    async def test_inject_returns_empty_if_no_docs(self):
        state = _make_state(rag_docs=[], rag_query="JWT")
        result = await rag_inject(state)

        assert result == {}

    @pytest.mark.asyncio
    async def test_inject_returns_empty_if_none_docs(self):
        state = _make_state(rag_docs=None)
        result = await rag_inject(state)

        assert result == {}

    @pytest.mark.asyncio
    async def test_inject_truncates_to_5_docs(self):
        docs = [_make_doc(0.8, f"doc {i}") for i in range(10)]
        state = _make_state(rag_docs=docs, rag_query="test")

        result = await rag_inject(state)

        # Verifica que máximo 5 docs são incluídos no contexto
        msg_content = result["messages"][0].content
        assert "doc 4" in msg_content  # 5 docs (0-4)
        assert "doc 5" not in msg_content  # 6º doc não incluído

    @pytest.mark.asyncio
    async def test_inject_includes_source_info(self):
        doc = Document(
            page_content="conteúdo",
            metadata={"source": "https://example.com", "title": "Exemplo"},
            relevance_score=0.85,
        )
        state = _make_state(rag_docs=[doc], rag_query="test")

        result = await rag_inject(state)

        msg_content = result["messages"][0].content
        assert "https://example.com" in msg_content
        assert "Exemplo" in msg_content
        assert "0.850" in msg_content  # score formatado
