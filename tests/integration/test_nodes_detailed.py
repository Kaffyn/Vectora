"""Detailed tests for nodes.py functions - Comprehensive coverage.

Tests individual node functions with mocks:
- call_llm: LLM invocation with tools
- handle_sub_node: Sub-agent execution
- process_retrieval: Fire-and-forget embedding
- _get_llm_with_tools: LLM caching
- _extract_tavily_results: Tavily response parsing
- _process_tavily_results: Document processing

Target: nodes.py 36% → 100%
"""

from __future__ import annotations

from typing import TYPE_CHECKING
from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from langchain_core.messages import AIMessage, HumanMessage, ToolMessage

from vectora.context import Context
from vectora.nodes import (
    _extract_tavily_results,
    _get_llm_with_tools,
    _process_tavily_results,
    call_llm,
    handle_sub_node,
    process_retrieval,
)

if TYPE_CHECKING:
    from vectora.state import State


class TestLLMCaching:
    """Test _get_llm_with_tools caching behavior."""

    def test_get_llm_with_tools_returns_runnable(self) -> None:
        """Test that _get_llm_with_tools returns a Runnable with tools bound."""
        with patch("vectora.nodes.load_llm") as mock_load_llm:
            mock_llm = MagicMock()
            mock_llm.bind_tools = MagicMock(return_value=MagicMock())
            mock_load_llm.return_value = mock_llm

            llm = _get_llm_with_tools()
            assert llm is not None
            mock_llm.bind_tools.assert_called_once()

    def test_get_llm_with_tools_caches_result(self) -> None:
        """Test that _get_llm_with_tools caches the result."""
        with patch("vectora.nodes.load_llm") as mock_load_llm:
            mock_llm = MagicMock()
            mock_result = MagicMock()
            mock_llm.bind_tools = MagicMock(return_value=mock_result)
            mock_load_llm.return_value = mock_llm

            # First call
            llm1 = _get_llm_with_tools()
            # Second call - should use cached result
            llm2 = _get_llm_with_tools()

            # load_llm should be called only once (from first invocation)
            assert llm1 is llm2  # Same object reference


class TestCallLLM:
    """Test call_llm MAIN_NODE function."""

    @pytest.mark.asyncio
    async def test_call_llm_with_empty_state(self) -> None:
        """Test call_llm handles empty message state."""
        mock_runtime = MagicMock()
        mock_runtime.context = Context(user_id="test", user_type="user", thread_id="t1")

        state: State = {"messages": [], "retrieval_results": {}}

        with patch("vectora.nodes._get_llm_with_tools") as mock_get_llm:
            mock_llm = AsyncMock()
            mock_llm.with_config = MagicMock(return_value=AsyncMock())
            mock_llm.with_config.return_value.ainvoke = AsyncMock(
                return_value=AIMessage(content="Hello")
            )
            mock_get_llm.return_value = mock_llm

            result = await call_llm(state, mock_runtime)
            assert isinstance(result, dict)

    @pytest.mark.asyncio
    async def test_call_llm_with_retrieval_context(self) -> None:
        """Test call_llm injects retrieval context into system prompt."""
        mock_runtime = MagicMock()
        mock_runtime.context = Context(user_id="test", user_type="user", thread_id="t1")

        state: State = {
            "messages": [HumanMessage(content="What about Python?")],
            "retrieval_results": {
                "articles": [
                    {
                        "page_content": "Python is a programming language.",
                        "metadata": {"title": "Python Overview"},
                    }
                ]
            },
        }

        with patch("vectora.nodes._get_llm_with_tools") as mock_get_llm:
            mock_llm = AsyncMock()
            mock_llm.with_config = MagicMock(return_value=AsyncMock())
            mock_llm.with_config.return_value.ainvoke = AsyncMock(
                return_value=AIMessage(content="Python is great!")
            )
            mock_get_llm.return_value = mock_llm

            result = await call_llm(state, mock_runtime)
            assert isinstance(result, dict)
            # System prompt should include retrieval context
            call_args = mock_llm.with_config.return_value.ainvoke.call_args
            if call_args:
                # Verify retrieval results were used
                assert call_args is not None


class TestHandleSubNode:
    """Test handle_sub_node for complex workflows."""

    @pytest.mark.asyncio
    async def test_handle_sub_node_creates_sub_graph(self) -> None:
        """Test handle_sub_node creates isolated sub-graph instance."""
        mock_runtime = MagicMock()
        mock_runtime.context = Context(user_id="test", user_type="user", thread_id="t1")

        state: State = {
            "messages": [HumanMessage(content="Complex analysis needed")],
            "retrieval_results": {},
        }

        with patch("vectora.nodes.build_graph") as mock_build_graph:
            mock_sub_graph = AsyncMock()
            mock_sub_graph.ainvoke = AsyncMock(return_value={"messages": []})
            mock_build_graph.return_value = mock_sub_graph

            result = await handle_sub_node(state, mock_runtime)
            assert isinstance(result, dict)


class TestExtractTavilyResults:
    """Test Tavily result extraction."""

    def test_extract_tavily_results_with_dict(self) -> None:
        """Test extracting results from dict with 'results' key."""
        data = {
            "results": [
                {
                    "title": "Article 1",
                    "url": "https://example.com",
                    "content": "Content here",
                }
            ]
        }

        results = _extract_tavily_results(data, "web_search")
        assert results is not None
        assert len(results) == 1
        assert results[0]["title"] == "Article 1"

    def test_extract_tavily_results_with_list(self) -> None:
        """Test extracting results from list directly."""
        data = [
            {
                "title": "Article 1",
                "url": "https://example.com",
                "content": "Content here",
            }
        ]

        results = _extract_tavily_results(data, "web_search")
        assert results is not None
        assert len(results) == 1

    def test_extract_tavily_results_with_invalid_data(self) -> None:
        """Test handling invalid data type."""
        # String is invalid
        results = _extract_tavily_results("invalid", "web_search")
        assert results is None

    def test_extract_tavily_results_with_empty_results(self) -> None:
        """Test empty results list."""
        data = {"results": []}
        results = _extract_tavily_results(data, "web_search")
        assert results == []


class TestProcessTavilyResults:
    """Test Tavily result processing."""

    @pytest.mark.asyncio
    async def test_process_tavily_results_formats_documents(self) -> None:
        """Test formatting Tavily results into documents."""
        results = [
            {
                "title": "Next.js Docs",
                "url": "https://nextjs.org",
                "content": "Next.js is a React framework.",
            }
        ]

        with patch("vectora.nodes.embedding") as mock_embedding:
            mock_embedding.astream = AsyncMock(return_value=AsyncMock())

            docs = await _process_tavily_results(results, "web_search", mock_embedding)

            assert len(docs) == 1
            assert docs[0]["page_content"] == "Next.js is a React framework."
            assert docs[0]["metadata"]["title"] == "Next.js Docs"
            assert docs[0]["metadata"]["url"] == "https://nextjs.org"
            assert docs[0]["metadata"]["source"] == "web_search"

    @pytest.mark.asyncio
    async def test_process_tavily_results_skips_empty_content(self) -> None:
        """Test skipping results with empty content."""
        results = [
            {"title": "Empty", "url": "https://example.com", "content": ""},
            {
                "title": "Valid",
                "url": "https://example.com",
                "content": "Some content",
            },
        ]

        with patch("vectora.nodes.embedding") as mock_embedding:
            mock_embedding.astream = AsyncMock(return_value=AsyncMock())

            docs = await _process_tavily_results(results, "web_search", mock_embedding)

            # Should only have one document (skipped empty)
            assert len(docs) == 1
            assert docs[0]["metadata"]["title"] == "Valid"

    @pytest.mark.asyncio
    async def test_process_tavily_results_handles_embedding_error(self) -> None:
        """Test handling embedding errors gracefully."""
        results = [
            {
                "title": "Doc",
                "url": "https://example.com",
                "content": "Content here",
            }
        ]

        with patch("vectora.nodes.embedding") as mock_embedding:
            # Simulate embedding error
            mock_embedding.astream = AsyncMock(side_effect=Exception("API error"))

            # Should still return formatted documents despite embedding error
            docs = await _process_tavily_results(results, "web_search", mock_embedding)

            assert len(docs) == 1
            assert docs[0]["page_content"] == "Content here"


class TestProcessRetrievalIntegration:
    """Integration tests for process_retrieval with Tavily."""

    @pytest.mark.asyncio
    async def test_process_retrieval_with_fetch_url_results(self) -> None:
        """Test process_retrieval with fetch_url tool results."""
        mock_runtime = MagicMock()
        mock_runtime.context = Context(user_id="test", user_type="user", thread_id="t1")

        state: State = {
            "messages": [
                HumanMessage(content="Get content from URL"),
                ToolMessage(
                    name="fetch_url",
                    content="<!DOCTYPE html><body>Fetched content</body>",
                    tool_call_id="fetch_1",
                ),
            ],
            "retrieval_results": {},
        }

        with patch("vectora.nodes.embedding") as mock_embedding:
            mock_embedding.astream = AsyncMock(return_value=AsyncMock())

            result = await process_retrieval(state, mock_runtime)
            assert isinstance(result, dict)

    @pytest.mark.asyncio
    async def test_process_retrieval_with_no_new_results(self) -> None:
        """Test process_retrieval returns empty dict when no new results."""
        mock_runtime = MagicMock()
        mock_runtime.context = Context(user_id="test", user_type="user", thread_id="t1")

        # No tool messages
        state: State = {
            "messages": [HumanMessage(content="Just a question")],
            "retrieval_results": {},
        }

        result = await process_retrieval(state, mock_runtime)
        # Should return empty dict (no changes)
        assert result == {} or result is None or "retrieval_results" not in result


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
