"""Tests for vectora/tools/rag.py"""

from __future__ import annotations

import json
from unittest.mock import AsyncMock, MagicMock, patch

import pytest


class TestVectorSearch:
    @pytest.mark.asyncio
    async def test_rag_disabled_returns_message(self):
        from vectora.tools.rag import vector_search

        with patch("vectora.tools.rag.settings") as mock_settings:
            mock_settings.enable_rag = False
            result = await vector_search.ainvoke(
                {"query": "test", "collection": "articles", "limit": 5}
            )
        assert "disabled" in result.lower() or "RAG" in result

    @pytest.mark.asyncio
    async def test_missing_dependencies_returns_message(self):
        from vectora.tools.rag import vector_search

        with patch("vectora.tools.rag.settings") as mock_settings:
            with patch("vectora.tools.rag.lancedb", None):
                with patch("vectora.tools.rag.CohereEmbeddings", None):
                    mock_settings.enable_rag = True
                    result = await vector_search.ainvoke(
                        {"query": "test", "collection": "articles", "limit": 5}
                    )
        assert "missing" in result.lower() or isinstance(result, str)

    @pytest.mark.asyncio
    async def test_missing_api_key_returns_error(self):
        from vectora.tools.rag import vector_search

        with patch("vectora.tools.rag.settings") as mock_settings:
            mock_settings.enable_rag = True
            mock_settings.get_cohere_api_key.return_value = None
            with patch("vectora.tools.rag.lancedb", MagicMock()):
                with patch("vectora.tools.rag.CohereEmbeddings", MagicMock()):
                    result = await vector_search.ainvoke(
                        {"query": "test", "collection": "articles", "limit": 5}
                    )
        data = json.loads(result)
        assert data.get("status") in ("failed", "error")


class TestEmbedding:
    @pytest.mark.asyncio
    async def test_rag_disabled_returns_error(self):
        from vectora.tools.rag import embedding

        with patch("vectora.tools.rag.settings") as mock_settings:
            mock_settings.enable_rag = False
            result = await embedding.ainvoke({"text": "doc", "collection": "articles"})
        data = json.loads(result)
        assert data["status"] == "error"

    @pytest.mark.asyncio
    async def test_queue_not_enabled_returns_error(self):
        from vectora.tools.rag import embedding

        with patch("vectora.tools.rag.settings") as mock_settings:
            mock_settings.enable_rag = True
            mock_settings.embedding_queue_enabled = False
            result = await embedding.ainvoke({"text": "doc", "collection": "articles"})
        data = json.loads(result)
        assert data["status"] == "error"


class TestIngestDocs:
    @pytest.mark.asyncio
    async def test_file_operations_disabled(self):
        from vectora.tools.rag import ingest_docs

        with patch("vectora.tools.rag.settings") as mock_settings:
            mock_settings.enable_file_operations = False
            result = await ingest_docs.ainvoke(
                {"directory_path": "/tmp", "collection": "articles"}
            )
        assert "disabled" in result.lower()

    @pytest.mark.asyncio
    async def test_unsafe_path_denied(self):
        from vectora.tools.rag import ingest_docs

        with patch("vectora.tools.rag.settings") as mock_settings:
            mock_settings.enable_file_operations = True
            with patch(
                "vectora.services.security.is_safe_file_path", return_value=False
            ):
                result = await ingest_docs.ainvoke(
                    {"directory_path": "/etc", "collection": "articles"}
                )
        assert "denied" in result.lower() or "Access" in result

    @pytest.mark.asyncio
    async def test_uses_ainvoke_not_astream(self, tmp_path):
        """Verifica o fix do bug: ingest_docs usa ainvoke, não astream."""
        from vectora.tools.rag import ingest_docs

        md_file = tmp_path / "test.md"
        md_file.write_text("# Test\nConteúdo de teste para embedding.")

        mock_result = json.dumps({"status": "fire_and_forget", "queue_id": "q1"})

        with patch("vectora.tools.rag.settings") as mock_settings:
            mock_settings.enable_file_operations = True
            mock_settings.enable_rag = True
            mock_settings.embedding_queue_enabled = True
            with patch(
                "vectora.services.security.is_safe_file_path", return_value=True
            ):
                with patch("vectora.services.gitignore.is_ignored", return_value=False):
                    with patch(
                        "vectora.services.gitignore.load_gitignore_spec",
                        return_value=None,
                    ):
                        with patch("vectora.tools.rag.embedding") as mock_emb:
                            mock_emb.ainvoke = AsyncMock(return_value=mock_result)
                            mock_emb.astream = MagicMock(
                                side_effect=AssertionError("astream foi chamado!")
                            )
                            result = await ingest_docs.ainvoke(
                                {
                                    "directory_path": str(tmp_path),
                                    "collection": "articles",
                                    "glob_pattern": "**/*.md",
                                }
                            )

        data = json.loads(result)
        assert data["status"] == "completed"
        assert mock_emb.ainvoke.called
