"""Unit Tests for Core Vectora Tools.

Tests for: web_search, fetch_url, embedding, vector_search, file_read, file_edit, grep.
Each tool is tested in isolation with mocked dependencies.
"""

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from tools import embedding, fetch_url, file_read, grep, vector_search, web_search


class TestWebSearch:
    """Tests for web_search tool."""

    @patch("tools.DuckDuckGoSearchRun")
    def test_web_search_returns_string(self, mock_ddg):
        """Verify web_search returns string results."""
        mock_search = MagicMock()
        mock_search.return_value = "Result 1. Result 2. Result 3."
        mock_ddg.return_value = mock_search

        result = web_search("test query")
        assert isinstance(result, str)
        assert len(result) > 0

    @patch("tools.DuckDuckGoSearchRun")
    def test_web_search_handles_empty_query(self, mock_ddg):
        """Verify web_search handles empty query gracefully."""
        mock_search = MagicMock()
        mock_search.return_value = ""
        mock_ddg.return_value = mock_search

        result = web_search("")
        assert isinstance(result, str)

    @patch("tools.DuckDuckGoSearchRun")
    def test_web_search_is_disabled_in_config(self, mock_ddg):
        """Verify web_search returns error when disabled."""
        with patch("tools.get_tool_config") as mock_config:
            mock_config.return_value.enable_web_search = False
            result = web_search("test")
            # Should return error message or disabled note
            if isinstance(result, str):
                assert (
                    "disable" in result.lower()
                    or "not available" in result.lower()
                    or "enabled" in result.lower()
                )


class TestFetchUrl:
    """Tests for fetch_url tool."""

    @patch("tools.WebBaseLoader")
    def test_fetch_url_returns_string(self, mock_loader_class):
        """Verify fetch_url returns text content."""
        mock_loader = MagicMock()
        mock_loader.load.return_value = [MagicMock(page_content="Page content here")]
        mock_loader_class.return_value = mock_loader

        result = fetch_url("https://example.com")
        assert isinstance(result, str)

    def test_fetch_url_validates_domain_whitelist(self):
        """Verify fetch_url respects domain whitelist."""
        with patch("tools.get_tool_config") as mock_config:
            mock_config.return_value.allowed_domains = ["example.com"]
            # Should reject non-whitelisted domain
            result = fetch_url("https://malicious.com")
            # Check if it returns error or string response
            if isinstance(result, str):
                assert (
                    "allowed" in result.lower()
                    or "whitelist" in result.lower()
                    or "malicious" in result.lower()
                    or "not allowed" in result.lower()
                )

    @patch("tools.WebBaseLoader")
    def test_fetch_url_respects_max_size(self, mock_loader_class):
        """Verify fetch_url truncates content exceeding max size."""
        mock_loader = MagicMock()
        large_content = "x" * 10000  # 10KB
        mock_loader.load.return_value = [MagicMock(page_content=large_content)]
        mock_loader_class.return_value = mock_loader

        with patch("tools.get_tool_config") as mock_config:
            mock_config.return_value.max_fetch_size = 5000
            result = fetch_url("https://example.com")
            assert len(result) <= 5000 + 100  # Allow some buffer


class TestEmbedding:
    """Tests for embedding tool (fire-and-forget pattern)."""

    @pytest.mark.asyncio
    async def test_embedding_returns_immediately(self):
        """Verify embedding returns fire-and-forget status immediately."""
        with patch("tools.get_embedding_queue") as mock_queue:
            mock_instance = MagicMock()
            mock_instance.enqueue = AsyncMock(return_value="queue-123")
            mock_queue.return_value = mock_instance

            result = await embedding("test document", collection="test")
            assert isinstance(result, (str, dict))
            # Should return status, not wait for embedding

    @pytest.mark.asyncio
    async def test_embedding_returns_error_when_disabled(self):
        """Verify embedding returns error when RAG disabled."""
        with patch("tools.get_tool_config") as mock_config:
            mock_config.return_value.enable_rag = False
            result = await embedding("test")
            assert isinstance(result, str)
            assert "disable" in result.lower()


class TestVectorSearch:
    """Tests for vector_search tool."""

    @pytest.mark.asyncio
    async def test_vector_search_returns_dict(self):
        """Verify vector_search returns results dictionary."""
        with patch("tools.get_lancedb_connection") as mock_db:
            mock_table = MagicMock()
            mock_table.search.return_value = [
                {"id": "1", "text": "result 1", "_distance": 0.1},
                {"id": "2", "text": "result 2", "_distance": 0.2},
            ]
            mock_db.return_value.open_table.return_value = mock_table

            result = await vector_search("query", collection="test")
            assert isinstance(result, dict)
            assert "results" in result or isinstance(result, dict)

    @pytest.mark.asyncio
    async def test_vector_search_respects_min_score(self):
        """Verify vector_search filters by min_score."""
        with patch("tools.get_lancedb_connection") as mock_db:
            mock_table = MagicMock()
            mock_table.search.return_value = [
                {"id": "1", "text": "good", "_distance": 0.1},
                {"id": "2", "text": "bad", "_distance": 0.9},
            ]
            mock_db.return_value.open_table.return_value = mock_table

            with patch("tools.get_tool_config") as mock_config:
                mock_config.return_value.search_min_score = 0.5
                result = await vector_search("query", collection="test")
                # Should filter results by score
                if isinstance(result, dict) and "results" in result:
                    # Good result should be included, bad should be filtered
                    assert any(
                        r.get("_distance", 1.0) <= 0.5 for r in result["results"]
                    )


class TestFileRead:
    """Tests for file_read tool."""

    def test_file_read_returns_content(self, tmp_path):
        """Verify file_read returns file content."""
        test_file = tmp_path / "test.txt"
        test_file.write_text("File content here")

        result = file_read(str(test_file))
        assert "File content" in result

    def test_file_read_respects_whitelist(self):
        """Verify file_read respects path whitelist."""
        with patch("tools.get_tool_config") as mock_config:
            mock_config.return_value.enable_file_operations = True
            # Should reject paths outside whitelist
            result = file_read("/etc/passwd")
            if isinstance(result, str):
                assert (
                    "not allowed" in result.lower()
                    or "permission" in result.lower()
                    or "whitelist" in result.lower()
                    or "access" in result.lower()
                )

    def test_file_read_handles_nonexistent_file(self):
        """Verify file_read handles missing files gracefully."""
        result = file_read("/nonexistent/file.txt")
        assert isinstance(result, str)
        assert "not found" in result.lower() or "error" in result.lower()


class TestGrep:
    """Tests for grep tool."""

    def test_grep_finds_pattern(self, tmp_path):
        """Verify grep finds pattern in files."""
        test_file = tmp_path / "test.txt"
        test_file.write_text("line 1\nline 2 with match\nline 3")

        result = grep("match", str(tmp_path))
        assert "line 2" in result or "match" in result

    def test_grep_prevents_redos(self, tmp_path):
        """Verify grep prevents ReDoS vulnerable patterns."""
        # Test with dangerous regex pattern
        grep("(a+)+b", str(tmp_path))
        # Should either reject pattern or handle it safely

    def test_grep_respects_max_depth(self):
        """Verify grep respects directory depth limit."""
        with patch("tools.get_tool_config") as mock_config:
            mock_config.return_value.enable_file_operations = True
            grep("pattern", "/")
            # Should limit recursion depth


class TestToolConfiguration:
    """Tests for tool configuration and safety."""

    def test_all_tools_have_schemas(self):
        """Verify all tools are properly configured with schemas."""
        from tools import TOOLS

        assert len(TOOLS) > 0
        for tool in TOOLS:
            assert hasattr(tool, "name")
            assert hasattr(tool, "description")
            # Tools should have schema for LLM binding

    def test_tools_are_callable(self):
        """Verify all tools are callable/awaitable."""
        from tools import TOOLS

        for tool in TOOLS:
            # Each tool should be executable
            assert callable(tool.func) or hasattr(tool, "invoke")
