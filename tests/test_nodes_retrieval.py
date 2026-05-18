"""Tests for vectora/nodes/retrieval.py"""

from __future__ import annotations

from unittest.mock import AsyncMock, patch

import pytest
from langchain_core.messages import HumanMessage

from vectora.nodes.retrieval import retrieval_node
from vectora.state import Document, State


def _state(query="test query") -> State:
    return {
        "messages": [HumanMessage(content=query)],
        "session_metadata": {},
    }


class TestRetrievalNode:
    @pytest.mark.asyncio
    async def test_rag_disabled_returns_empty(self):
        # settings importado localmente em retrieval_node → patch no módulo de origem
        with patch("vectora.config.settings.settings") as mock_settings:
            mock_settings.enable_rag = False
            # re-importa dentro do contexto do patch não funciona facilmente;
            # melhor testar via resultado com settings real desabilitado
            from vectora.config.settings import settings as real_settings

            original = real_settings.enable_rag
            try:
                real_settings.enable_rag = False  # type: ignore[misc]
                result = await retrieval_node(_state())
            finally:
                real_settings.enable_rag = original  # type: ignore[misc]
        assert result == {}

    @pytest.mark.asyncio
    async def test_empty_query_returns_empty(self):
        state: State = {"messages": [], "session_metadata": {}}
        result = await retrieval_node(state)
        # sem mensagem humana → sem query → retorna {}
        assert result == {}

    @pytest.mark.asyncio
    async def test_no_results_returns_empty(self):
        with patch(
            "vectora.nodes.rag_subgraph._call_vector_search", new_callable=AsyncMock
        ) as mock_vs:
            mock_vs.return_value = []
            result = await retrieval_node(_state())
        assert result == {}

    @pytest.mark.asyncio
    async def test_returns_rag_docs_on_success(self):
        docs = [
            Document(page_content="doc1", metadata={}, relevance_score=0.9),
            Document(page_content="doc2", metadata={}, relevance_score=0.7),
        ]
        with patch(
            "vectora.nodes.rag_subgraph._call_vector_search", new_callable=AsyncMock
        ) as mock_vs:
            with patch(
                "vectora.nodes.retrieval._rerank", new_callable=AsyncMock
            ) as mock_rerank:
                mock_vs.return_value = docs
                mock_rerank.return_value = docs
                result = await retrieval_node(_state("JWT token"))

        assert "rag_docs" in result
        assert result["rag_query"] == "JWT token"
        assert len(result["rag_docs"]) == 2
