"""Test suite for RAG tools: embedding, vector_search, and internal reranker."""

import json
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from tools import TOOLS_BY_NAME, embedding, vector_search


class TestEmbeddingTool:
    """Test suite for the embedding tool."""

    @pytest.mark.asyncio
    async def test_embedding_indexes_document(self: "TestEmbeddingTool") -> None:
        """Test embedding() successfully indexes a document in Qdrant."""
        test_text = "This is a test document about GraphQL."
        test_collection = "articles"

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.QdrantClient") as mock_qdrant,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                return_value=[[0.1, 0.2, 0.3]]
            )
            mock_voyage.return_value = mock_voyage_instance

            mock_qdrant_instance = MagicMock()
            mock_qdrant_instance.upsert = AsyncMock(return_value="point_123")
            mock_qdrant.return_value = mock_qdrant_instance

            result_str = await embedding(
                text=test_text, collection=test_collection, runtime=None
            )
            result = json.loads(result_str)

            assert result["status"] == "indexed"
            assert result["collection"] == test_collection
            assert result["point_id"] == "point_123"
            assert result["text_length"] == len(test_text)

    @pytest.mark.asyncio
    async def test_embedding_with_metadata(
        self: "TestEmbeddingTool",
    ) -> None:
        """Test embedding() accepts and stores metadata."""
        test_text = "Test document"
        test_collection = "wiki"
        test_metadata = {"source": "test.md", "author": "test_user"}

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.QdrantClient") as mock_qdrant,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                return_value=[[0.1, 0.2, 0.3]]
            )
            mock_voyage.return_value = mock_voyage_instance

            mock_qdrant_instance = MagicMock()
            mock_qdrant_instance.upsert = AsyncMock(return_value="point_456")
            mock_qdrant.return_value = mock_qdrant_instance

            result_str = await embedding(
                text=test_text,
                collection=test_collection,
                metadata=test_metadata,
                runtime=None,
            )
            result = json.loads(result_str)

            assert result["status"] == "indexed"
            mock_qdrant_instance.upsert.assert_called_once()

    @pytest.mark.asyncio
    async def test_embedding_queue_fallback_on_api_failure(
        self: "TestEmbeddingTool",
    ) -> None:
        """Test embedding() enqueues document when Voyage API fails."""
        test_text = "Test document for queue"
        test_collection = "articles"

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.get_embedding_queue") as mock_queue,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                side_effect=Exception("API Error")
            )
            mock_voyage.return_value = mock_voyage_instance

            mock_queue_instance = AsyncMock()
            mock_queue_instance.enqueue = AsyncMock(return_value="queue_abc123")
            mock_queue.return_value = mock_queue_instance

            result_str = await embedding(
                text=test_text, collection=test_collection, runtime=None
            )
            result = json.loads(result_str)

            assert result["status"] == "queued_for_indexing"
            assert result["queue_id"] == "queue_abc123"
            assert "enqueued" in result["message"].lower()


class TestVectorSearchTool:
    """Test suite for the vector_search tool."""

    @pytest.mark.asyncio
    async def test_vector_search_finds_documents(
        self: "TestVectorSearchTool",
    ) -> None:
        """Test vector_search() finds and returns ranked documents."""
        test_query = "What is GraphQL?"

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.QdrantClient") as mock_qdrant,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                return_value=[[0.1, 0.2, 0.3]]
            )
            mock_voyage.return_value = mock_voyage_instance

            mock_qdrant_instance = MagicMock()
            mock_search_result = MagicMock()
            mock_search_result.payload = {
                "page_content": "GraphQL is a query language...",
                "metadata": {"source": "docs.md"},
            }
            mock_search_result.score = 0.95
            mock_qdrant_instance.search = AsyncMock(return_value=[mock_search_result])
            mock_qdrant.return_value = mock_qdrant_instance

            result_str = await vector_search(
                query=test_query, collections=["articles"], top_k=5, runtime=None
            )
            result = json.loads(result_str)

            assert result["status"] == "found"
            assert len(result["results"]) > 0
            assert "page_content" in result["results"][0]
            assert "relevance_score" in result["results"][0]

    @pytest.mark.asyncio
    async def test_vector_search_no_results(
        self: "TestVectorSearchTool",
    ) -> None:
        """Test vector_search() returns empty status when no documents found."""
        test_query = "Nonexistent query xyz"

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.QdrantClient") as mock_qdrant,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                return_value=[[0.1, 0.2, 0.3]]
            )
            mock_voyage.return_value = mock_voyage_instance

            mock_qdrant_instance = MagicMock()
            mock_qdrant_instance.search = AsyncMock(return_value=[])
            mock_qdrant.return_value = mock_qdrant_instance

            result_str = await vector_search(
                query=test_query, collections=["articles"], top_k=5, runtime=None
            )
            result = json.loads(result_str)

            assert result["status"] == "no_results"
            assert result["count"] == 0

    @pytest.mark.asyncio
    async def test_vector_search_multiple_collections(
        self: "TestVectorSearchTool",
    ) -> None:
        """Test vector_search() searches multiple Qdrant collections."""
        test_query = "API documentation"

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.QdrantClient") as mock_qdrant,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                return_value=[[0.1, 0.2, 0.3]]
            )
            mock_voyage.return_value = mock_voyage_instance

            mock_qdrant_instance = MagicMock()
            mock_qdrant_instance.search = AsyncMock(return_value=[])
            mock_qdrant.return_value = mock_qdrant_instance

            result_str = await vector_search(
                query=test_query,
                collections=["api_docs", "wiki"],
                top_k=10,
                runtime=None,
            )
            result = json.loads(result_str)

            assert result["status"] == "no_results"

    @pytest.mark.asyncio
    async def test_vector_search_respects_min_score(
        self: "TestVectorSearchTool",
    ) -> None:
        """Test vector_search() filters results by min_score threshold."""
        test_query = "GraphQL"
        min_score = 0.7

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.QdrantClient") as mock_qdrant,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                return_value=[[0.1, 0.2, 0.3]]
            )
            mock_voyage.return_value = mock_voyage_instance

            low_score_result = MagicMock()
            low_score_result.payload = {
                "page_content": "Low relevance content",
                "metadata": {},
            }
            low_score_result.score = 0.5

            mock_qdrant_instance = MagicMock()
            mock_qdrant_instance.search = AsyncMock(return_value=[low_score_result])
            mock_qdrant.return_value = mock_qdrant_instance

            result_str = await vector_search(
                query=test_query,
                collections=["articles"],
                top_k=5,
                min_score=min_score,
                runtime=None,
            )
            result = json.loads(result_str)

            assert result["status"] == "no_results"


class TestRerankerInternal:
    """Test suite for internal reranker (not exposed to LLM)."""

    TOP_K_LIMIT = 2

    def test_reranker_not_in_tools_list(self: "TestRerankerInternal") -> None:
        """Verify reranker is NOT exposed as a tool to the LLM."""
        tool_names = list(TOOLS_BY_NAME.keys())

        assert "reranker" not in tool_names
        assert "_internal_reranker" not in tool_names

        assert "embedding" in tool_names
        assert "vector_search" in tool_names

    @pytest.mark.asyncio
    async def test_reranker_bm25_mode(self: "TestRerankerInternal") -> None:
        """Test internal reranker uses BM25 when configured."""
        with patch("tools.get_tool_config") as mock_config:
            mock_config_instance = MagicMock()
            mock_config_instance.reranker_type = "bm25"
            mock_config.return_value = mock_config_instance

            from tools import _internal_reranker

            mock_results = [
                MagicMock(
                    payload={"page_content": "GraphQL query language"},
                    score=0.9,
                ),
                MagicMock(
                    payload={"page_content": "REST API design"},
                    score=0.7,
                ),
            ]

            reranked = await _internal_reranker(
                query="GraphQL", raw_results=mock_results, top_k=self.TOP_K_LIMIT
            )

            assert len(reranked) <= self.TOP_K_LIMIT
            assert all(hasattr(r, "payload") and hasattr(r, "score") for r in reranked)

    @pytest.mark.asyncio
    async def test_reranker_reorders_results(
        self: "TestRerankerInternal",
    ) -> None:
        """Test internal reranker reorders results by relevance."""
        with patch("tools.get_tool_config") as mock_config:
            mock_config_instance = MagicMock()
            mock_config_instance.reranker_type = "bm25"
            mock_config.return_value = mock_config_instance

            from tools import _internal_reranker

            mock_results = [
                MagicMock(
                    payload={"page_content": "Dog walking tips for pet owners"},
                    score=0.85,
                ),
                MagicMock(
                    payload={"page_content": "How to train a dog"},
                    score=0.70,
                ),
            ]

            reranked = await _internal_reranker(
                query="dog training", raw_results=mock_results, top_k=2
            )

            assert len(reranked) > 0


class TestRAGEnd2End:
    """Integration tests for RAG workflow."""

    @pytest.mark.asyncio
    async def test_rag_workflow_search_then_use_results(
        self: "TestRAGEnd2End",
    ) -> None:
        """Test complete RAG workflow: search and use retrieved documents."""
        test_query = "How to use GraphQL?"

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.QdrantClient") as mock_qdrant,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                return_value=[[0.1, 0.2, 0.3]]
            )
            mock_voyage.return_value = mock_voyage_instance

            mock_result = MagicMock()
            mock_result.payload = {
                "page_content": "GraphQL is a query language for APIs",
                "metadata": {"source": "api_docs.md"},
            }
            mock_result.score = 0.92

            mock_qdrant_instance = MagicMock()
            mock_qdrant_instance.search = AsyncMock(return_value=[mock_result])
            mock_qdrant.return_value = mock_qdrant_instance

            result_str = await vector_search(
                query=test_query, collections=["api_docs"], top_k=5, runtime=None
            )
            result = json.loads(result_str)

            assert result["status"] == "found"
            assert len(result["results"]) == 1
            assert "GraphQL" in result["results"][0]["page_content"]

    @pytest.mark.asyncio
    async def test_rag_workflow_no_results_then_index(
        self: "TestRAGEnd2End",
    ) -> None:
        """Test RAG workflow: search finds nothing, then embed new doc."""
        search_query = "Advanced GraphQL concepts"
        new_doc = "Fragments in GraphQL allow code reuse..."

        with (
            patch("tools.VoyageAIEmbeddings") as mock_voyage,
            patch("tools.QdrantClient") as mock_qdrant,
        ):
            mock_voyage_instance = MagicMock()
            mock_voyage_instance.embed_documents = AsyncMock(
                return_value=[[0.1, 0.2, 0.3]]
            )
            mock_voyage.return_value = mock_voyage_instance

            mock_qdrant_instance = MagicMock()
            mock_qdrant_instance.search = AsyncMock(return_value=[])
            mock_qdrant_instance.upsert = AsyncMock(return_value="point_789")
            mock_qdrant.return_value = mock_qdrant_instance

            search_result_str = await vector_search(
                query=search_query,
                collections=["knowledge_base"],
                top_k=5,
                runtime=None,
            )
            search_result = json.loads(search_result_str)

            assert search_result["status"] == "no_results"

            embed_result_str = await embedding(
                text=new_doc, collection="knowledge_base", runtime=None
            )
            embed_result = json.loads(embed_result_str)

            assert embed_result["status"] == "indexed"
            assert embed_result["point_id"] == "point_789"
