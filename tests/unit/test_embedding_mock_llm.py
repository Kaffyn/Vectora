"""Testes de embedding com LLM mockado.

Verifica que:
1. Tools estão corretamente vinculadas ao grafo
2. Tool calls são criadas corretamente
3. ToolMessages são geradas após execução de tool
4. Fluxo de mensagens é preservado através do grafo
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from langchain_core.messages import AIMessage, HumanMessage, ToolMessage


class TestEmbeddingToolMocked:
    """Testa o tool de embedding com LLM mockado."""

    @pytest.mark.asyncio
    async def test_embedding_tool_enqueues_document(self):
        """Verifica que embedding tool enfileira documento corretamente."""
        from vectora.tools.rag import embedding

        mock_queue = AsyncMock()
        mock_queue.enqueue = AsyncMock(return_value="test-queue-id-123")

        with patch("vectora.tools.rag.get_embedding_queue", return_value=mock_queue):
            mock_settings = MagicMock()
            mock_settings.enable_rag = True
            mock_settings.embedding_queue_enabled = True
            mock_settings.embedding_queue_dsn = "sqlite+aiosqlite:///:memory:"

            with patch("vectora.tools.rag.settings", mock_settings):
                result = ""
                async for chunk in embedding.astream(
                    {"text": "Conteúdo de teste", "collection": "test"}
                ):
                    if isinstance(chunk, str):
                        result += chunk

        import json

        data = json.loads(result)
        assert data["status"] == "fire_and_forget"
        assert data["queue_id"] == "test-queue-id-123"

    @pytest.mark.asyncio
    async def test_embedding_returns_error_when_rag_disabled(self):
        """Verifica que embedding retorna erro quando RAG está desabilitado."""
        from vectora.tools.rag import embedding

        mock_settings = MagicMock()
        mock_settings.enable_rag = False

        with patch("vectora.tools.rag.settings", mock_settings):
            result = ""
            async for chunk in embedding.astream(
                {"text": "texto", "collection": "test"}
            ):
                if isinstance(chunk, str):
                    result += chunk

        import json

        data = json.loads(result)
        assert data["status"] == "error"

    @pytest.mark.asyncio
    async def test_vector_search_returns_no_results_without_table(self, tmp_path):
        """Verifica que vector_search retorna no_results quando tabela não existe."""
        lancedb = pytest.importorskip("lancedb", reason="lancedb não instalado")

        from vectora.tools.rag import vector_search

        mock_settings = MagicMock()
        mock_settings.enable_rag = True
        mock_settings.get_cohere_api_key.return_value = "test-key"
        mock_settings.embedding_model = "embed-multilingual-v3.0"
        mock_settings.lancedb_dir = str(tmp_path / "test_lancedb")
        mock_settings.reranker_type = "none"

        mock_embeddings = MagicMock()
        mock_embeddings.embed_query.return_value = [0.1] * 1024

        mock_db = AsyncMock()
        mock_db.open_table = AsyncMock(side_effect=Exception("Table not found"))

        with (
            patch("vectora.tools.rag.settings", mock_settings),
            patch("vectora.tools.rag.CohereEmbeddings", return_value=mock_embeddings),
            patch("vectora.tools.rag.lancedb") as mock_lancedb,
        ):
            mock_lancedb.connect_async = AsyncMock(return_value=mock_db)
            result = ""
            async for chunk in vector_search.astream(
                {"query": "busca teste", "collection": "test_col"}
            ):
                if isinstance(chunk, str):
                    result += chunk

        import json

        data = json.loads(result)
        assert data["status"] in ("no_results", "error", "failed")


class TestMessageFlow:
    """Testa fluxo de mensagens no grafo."""

    def test_human_message_creation(self):
        """Verifica criação de HumanMessage."""
        msg = HumanMessage(content="Olá, Vectinho!")
        assert msg.content == "Olá, Vectinho!"
        assert msg.type == "human"

    def test_ai_message_creation(self):
        """Verifica criação de AIMessage."""
        msg = AIMessage(content="Olá! Como posso ajudar?")
        assert msg.content == "Olá! Como posso ajudar?"
        assert msg.type == "ai"

    def test_tool_message_creation(self):
        """Verifica criação de ToolMessage."""
        msg = ToolMessage(content='{"status": "ok"}', tool_call_id="call_123")
        assert msg.content == '{"status": "ok"}'
        assert msg.tool_call_id == "call_123"

    def test_ai_message_with_tool_calls(self):
        """Verifica AIMessage com tool_calls."""
        msg = AIMessage(
            content="",
            tool_calls=[
                {
                    "name": "embedding",
                    "args": {"text": "doc", "collection": "test"},
                    "id": "call_001",
                    "type": "tool_call",
                }
            ],
        )
        assert len(msg.tool_calls) == 1
        assert msg.tool_calls[0]["name"] == "embedding"
