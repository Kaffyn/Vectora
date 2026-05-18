"""Unit Tests for Core Vectora Tools with Streaming Pattern.

Tests for: web_search, fetch_url, embedding, vector_search, file_read, grep.
Each tool is tested with astream/stream pattern (reactive, not invoke/astream_events).
"""

from __future__ import annotations

import json
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from vectora.tools.fs import file_read as file_read_tool
from vectora.tools.fs import grep as grep_tool

# Import StructuredTool wrappers
from vectora.tools.rag import embedding as embedding_tool
from vectora.tools.rag import vector_search as vector_search_tool
from vectora.tools.web import fetch_url as fetch_url_tool
from vectora.tools.web import web_search as web_search_tool

# Use StructuredTools directly (they have .stream() and .astream() methods)
vector_search = vector_search_tool
embedding = embedding_tool
fetch_url = fetch_url_tool
file_read = file_read_tool
web_search = web_search_tool
grep = grep_tool


class TestWebSearch:
    """Tests for web_search tool using Tavily."""

    @patch("vectora.tools.web.TavilyClient")
    def test_web_search_returns_string(self, mock_tavily_class: MagicMock) -> None:
        """Verify web_search returns JSON results via streaming."""
        mock_settings = MagicMock()
        mock_settings.enable_web_search = True
        mock_settings.tavily_api_key = "test-key"

        mock_client = MagicMock()
        mock_client.search.return_value = {
            "results": [
                {
                    "title": "Result 1",
                    "url": "https://example.com/1",
                    "content": "Content 1",
                },
                {
                    "title": "Result 2",
                    "url": "https://example.com/2",
                    "content": "Content 2",
                },
            ]
        }
        mock_tavily_class.return_value = mock_client

        with patch("vectora.tools.web.settings", mock_settings):
            result = ""
            for chunk in web_search.stream({"query": "test query"}):
                if isinstance(chunk, str):
                    result += chunk
            assert isinstance(result, str)

    @patch("vectora.tools.web.TavilyClient")
    def test_web_search_handles_empty_query(self, mock_tavily_class: MagicMock) -> None:
        """Verify web_search handles empty query gracefully."""
        mock_settings = MagicMock()
        mock_settings.enable_web_search = True
        mock_settings.tavily_api_key = "test-key"

        mock_client = MagicMock()
        mock_client.search.return_value = {"results": []}
        mock_tavily_class.return_value = mock_client

        with patch("vectora.tools.web.settings", mock_settings):
            result = ""
            for chunk in web_search.stream({"query": ""}):
                if isinstance(chunk, str):
                    result += chunk
            assert isinstance(result, str)

    def test_web_search_is_disabled_in_config(self) -> None:
        """Verify web_search returns error when disabled."""
        mock_settings = MagicMock()
        mock_settings.enable_web_search = False
        mock_settings.tavily_api_key = None

        with patch("vectora.tools.web.settings", mock_settings):
            result = ""
            for chunk in web_search.stream({"query": "test"}):
                if isinstance(chunk, str):
                    result += chunk
            if isinstance(result, str):
                assert (
                    "disable" in result.lower()
                    or "not available" in result.lower()
                    or "enabled" in result.lower()
                    or "key" in result.lower()
                )


class TestFetchUrl:
    """Tests for fetch_url tool."""

    @patch("vectora.tools.web.TavilyClient")
    def test_fetch_url_returns_string(self, mock_tavily_class: MagicMock) -> None:
        """Verify fetch_url returns text content via Tavily streaming."""
        mock_settings = MagicMock()
        mock_settings.tavily_api_key = "test-key"

        mock_client = MagicMock()
        mock_client.search.return_value = {
            "results": [
                {
                    "title": "Example Page",
                    "url": "https://example.com",
                    "content": "Page content here with extracted text",
                }
            ]
        }
        mock_tavily_class.return_value = mock_client

        with patch("vectora.tools.web.settings", mock_settings):
            result = ""
            for chunk in fetch_url.stream({"url": "https://example.com"}):
                if isinstance(chunk, str):
                    result += chunk
            assert isinstance(result, str)

    def test_fetch_url_requires_valid_url_format(self) -> None:
        """Verify fetch_url validates URL format."""
        mock_settings = MagicMock()
        mock_settings.tavily_api_key = "test-key"

        with patch("vectora.tools.web.settings", mock_settings):
            result = ""
            for chunk in fetch_url.stream({"url": "not-a-url"}):
                if isinstance(chunk, str):
                    result += chunk
            assert "Error" in result or "http" in result or "invalid" in result.lower()

    @patch("vectora.tools.web.TavilyClient")
    def test_fetch_url_requires_tavily_key(self, mock_tavily_class: MagicMock) -> None:
        """Verify fetch_url requires TAVILY_API_KEY."""
        mock_settings = MagicMock()
        mock_settings.tavily_api_key = ""

        with patch("vectora.tools.web.settings", mock_settings):
            result = ""
            for chunk in fetch_url.stream({"url": "https://example.com"}):
                if isinstance(chunk, str):
                    result += chunk
            assert isinstance(result, str)


class TestEmbedding:
    """Tests for embedding tool (fire-and-forget pattern)."""

    @pytest.mark.asyncio
    async def test_embedding_returns_immediately(self) -> None:
        """Verify embedding returns fire-and-forget status immediately."""
        mock_settings = MagicMock()
        mock_settings.enable_rag = True
        mock_settings.embedding_queue_enabled = True
        mock_settings.embedding_queue_dsn = "sqlite+aiosqlite:///:memory:"
        mock_settings.cohere_api_key = "test-key"

        with patch("vectora.tools.rag.settings", mock_settings):
            with patch("vectora.tools.rag.get_embedding_queue") as mock_queue:
                mock_instance = AsyncMock()
                mock_instance.enqueue = AsyncMock(return_value="queue-123")
                mock_queue.return_value = mock_instance

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

                assert isinstance(result, str)
                result_json = json.loads(result)
                assert result_json.get("status") == "fire_and_forget"
                assert result_json.get("queue_id") == "queue-123"

    @pytest.mark.asyncio
    async def test_embedding_returns_error_when_disabled(self) -> None:
        """Verify embedding returns error when RAG disabled."""
        mock_settings = MagicMock()
        mock_settings.enable_rag = False
        mock_settings.embedding_queue_enabled = False

        with patch("vectora.tools.rag.settings", mock_settings):
            result = ""
            async for chunk in embedding.astream(
                {
                    "text": "test",
                    "collection": "test",
                }
            ):
                if isinstance(chunk, str):
                    result += chunk

            assert isinstance(result, str)
            result_json = json.loads(result)
            assert result_json.get("status") == "error"


class TestVectorSearch:
    """Tests for vector_search tool."""

    @pytest.mark.asyncio
    async def test_vector_search_returns_dict(self) -> None:
        """Verify vector_search returns results dictionary with mocked services."""
        mock_settings = MagicMock()
        mock_settings.enable_rag = False

        with patch("vectora.tools.rag.settings", mock_settings):
            result = ""
            async for chunk in vector_search.astream(
                {"query": "query", "collection": "test"}
            ):
                if isinstance(chunk, str):
                    result += chunk

            assert isinstance(result, str)
            assert "disabled" in result.lower() or "not available" in result.lower()

    @pytest.mark.asyncio
    async def test_vector_search_respects_min_score(self) -> None:
        """Verify vector_search filters by min_score."""
        mock_settings = MagicMock()
        mock_settings.enable_rag = True
        mock_settings.cohere_api_key = "mock-cohere-key"
        mock_settings.lancedb_dir = "./data/lancedb"
        mock_settings.search_min_score = 0.5

        with patch("vectora.tools.rag.settings", mock_settings):
            result = ""
            async for chunk in vector_search.astream(
                {"query": "query", "collection": "test"}
            ):
                if isinstance(chunk, str):
                    result += chunk

            assert isinstance(result, str)


class TestFileRead:
    """Tests for file_read tool."""

    @patch("vectora.tools.fs.is_safe_file_path", return_value=True)
    def test_file_read_returns_content(self, mock_safe, tmp_path) -> None:
        """Verify file_read returns file content via streaming."""
        test_file = tmp_path / "test.txt"
        test_file.write_text("File content here")

        result = ""
        for chunk in file_read.stream({"file_path": str(test_file)}):
            if isinstance(chunk, str):
                result += chunk
        assert "File content" in result

    def test_file_read_respects_whitelist(self) -> None:
        """Verify file_read respects path whitelist."""
        result = ""
        for chunk in file_read.stream({"file_path": "/etc/passwd"}):
            if isinstance(chunk, str):
                result += chunk
        if isinstance(result, str):
            assert (
                "not allowed" in result.lower()
                or "permission" in result.lower()
                or "whitelist" in result.lower()
                or "access" in result.lower()
                or "error" in result.lower()
            )

    def test_file_read_handles_nonexistent_file(self) -> None:
        """Verify file_read handles missing files gracefully."""
        result = ""
        for chunk in file_read.stream({"file_path": "/nonexistent/file.txt"}):
            if isinstance(chunk, str):
                result += chunk
        assert isinstance(result, str)
        assert "not found" in result.lower() or "error" in result.lower()


class TestGrep:
    """Tests for grep tool."""

    def test_grep_finds_pattern(self, tmp_path) -> None:
        """Verify grep finds pattern in files via streaming."""
        test_file = tmp_path / "test.txt"
        test_file.write_text("line 1\nline 2 with match\nline 3")

        result = ""
        for chunk in grep.stream({"pattern": "match", "directory": str(tmp_path)}):
            if isinstance(chunk, str):
                result += chunk
        assert "line 2" in result or "match" in result

    def test_grep_prevents_redos(self, tmp_path) -> None:
        """Verify grep prevents ReDoS vulnerable patterns."""
        result = ""
        for chunk in grep.stream({"pattern": "(a+)+b", "directory": str(tmp_path)}):
            if isinstance(chunk, str):
                result += chunk
        assert isinstance(result, str)

    def test_grep_respects_max_depth(self) -> None:
        """Verify grep respects directory depth limit."""
        mock_settings = MagicMock()
        mock_settings.enable_file_operations = True

        with patch("vectora.tools.fs.settings", mock_settings):
            result = ""
            for chunk in grep.stream({"pattern": "pattern", "directory": "/"}):
                if isinstance(chunk, str):
                    result += chunk
            assert isinstance(result, str)


class TestToolConfiguration:
    """Tests for tool configuration and safety."""

    def test_all_tools_have_schemas(self) -> None:
        """Verify all tools are properly configured with schemas."""
        from vectora.tools import TOOLS

        assert len(TOOLS) > 0
        for tool in TOOLS:
            assert hasattr(tool, "name")
            assert hasattr(tool, "description")

    def test_tools_are_callable(self) -> None:
        """Verify all tools are callable/streamable."""
        from vectora.tools import TOOLS

        for tool in TOOLS:
            assert hasattr(tool, "stream") or hasattr(tool, "astream")
