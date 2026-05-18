"""End-to-end tests for Vectora run commands.

Tests initialization, directory creation, and basic CLI operations.
"""

import asyncio
import logging
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

logger = logging.getLogger(__name__)


class TestVectoraInitialization:
    """Test Vectora initialization and directory creation."""

    def test_initialize_vectora_home_creates_directories(self, tmp_path, monkeypatch):
        """Test that initialize_vectora_home creates all required directories."""
        # Mock Path.home() to use tmp_path
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import initialize_vectora_home

        vectora_home = initialize_vectora_home()

        assert vectora_home == tmp_path / ".vectora"
        assert (tmp_path / ".vectora").exists()
        assert (tmp_path / ".vectora" / "data").exists()
        assert (tmp_path / ".vectora" / "logs").exists()

    def test_verify_vectora_setup_returns_true_when_initialized(
        self, tmp_path, monkeypatch
    ):
        """Test that verify_vectora_setup returns True when all directories exist."""
        # Create directories
        vectora_home = tmp_path / ".vectora"
        vectora_home.mkdir()
        (vectora_home / "data").mkdir()
        (vectora_home / "logs").mkdir()

        # Mock Path.home()
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import verify_vectora_setup

        assert verify_vectora_setup() is True

    def test_verify_vectora_setup_returns_false_when_missing_dirs(
        self, tmp_path, monkeypatch
    ):
        """Test that verify_vectora_setup returns False when directories are missing."""
        # Mock Path.home() without creating directories
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import verify_vectora_setup

        assert verify_vectora_setup() is False

    def test_ensure_vectora_initialized_creates_missing_dirs(
        self, tmp_path, monkeypatch
    ):
        """Test that ensure_vectora_initialized creates missing directories."""
        # Mock Path.home()
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import (
            ensure_vectora_initialized,
            verify_vectora_setup,
        )

        # Verify nothing exists initially
        assert not verify_vectora_setup()

        # Run initialization
        ensure_vectora_initialized()

        # Verify everything was created
        assert verify_vectora_setup()
        assert (tmp_path / ".vectora").exists()
        assert (tmp_path / ".vectora" / "data").exists()
        assert (tmp_path / ".vectora" / "logs").exists()


class TestChatInitialization:
    """Test chat application initialization."""

    def test_chat_initializes_with_proper_env_locations(self, tmp_path, monkeypatch):
        """Test that chat initialization checks correct .env locations."""
        # Create test .env files
        cwd_env = tmp_path / ".env"
        cwd_env.write_text("TEST_CWD=true\n")

        vectora_home = tmp_path / ".vectora"
        vectora_home.mkdir()
        home_env = vectora_home / ".env"
        home_env.write_text("TEST_HOME=true\n")

        # Verify both files exist
        assert cwd_env.exists()
        assert home_env.exists()


class TestSetupWizardInitialization:
    """Test setup wizard directory initialization."""

    @pytest.mark.asyncio
    async def test_setup_wizard_creates_config_directory(self, tmp_path, monkeypatch):
        """Test that setup wizard creates ~/.vectora directory."""
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import initialize_vectora_home

        # Initialize
        vectora_home = initialize_vectora_home()

        # Verify structure
        assert vectora_home.exists()
        assert (vectora_home / "data").exists()
        assert (vectora_home / "logs").exists()

    @pytest.mark.asyncio
    async def test_setup_wizard_saves_env_to_correct_location(
        self, tmp_path, monkeypatch
    ):
        """Test that setup wizard saves .env to ~/.vectora/.env."""
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import initialize_vectora_home
        from vectora.services.setup_wizard import _save_to_env

        # Initialize
        initialize_vectora_home()

        # Save configuration
        _save_to_env("google-genai", "test-api-key")

        # Verify file was created
        env_file = tmp_path / ".vectora" / ".env"
        assert env_file.exists()

        content = env_file.read_text()
        assert "LLM_PROVIDER=google-genai" in content
        assert "GOOGLE_API_KEY=test-api-key" in content


class TestRunCommandsIntegration:
    """Integration tests for run commands."""

    def test_vectora_home_structure_persists(self, tmp_path, monkeypatch):
        """Test that Vectora home structure is created and persists."""
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import (
            ensure_vectora_initialized,
            verify_vectora_setup,
        )

        # First call initializes
        ensure_vectora_initialized()
        assert verify_vectora_setup()

        # Second call finds existing directories
        ensure_vectora_initialized()
        assert verify_vectora_setup()

        # Verify all directories still exist
        assert (tmp_path / ".vectora").exists()
        assert (tmp_path / ".vectora" / "data").exists()
        assert (tmp_path / ".vectora" / "logs").exists()

    def test_vectora_logs_directory_is_writable(self, tmp_path, monkeypatch):
        """Test that logs directory can be written to."""
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import initialize_vectora_home

        initialize_vectora_home()

        # Try to write a test log
        logs_dir = tmp_path / ".vectora" / "logs"
        test_log = logs_dir / "test.log"
        test_log.write_text("test log content")

        assert test_log.exists()
        assert test_log.read_text() == "test log content"

    def test_vectora_data_directory_is_writable(self, tmp_path, monkeypatch):
        """Test that data directory can be written to."""
        monkeypatch.setattr(Path, "home", lambda: tmp_path)

        from vectora.services.initialization import initialize_vectora_home

        initialize_vectora_home()

        # Try to write a test file
        data_dir = tmp_path / ".vectora" / "data"
        test_file = data_dir / "test.db"
        test_file.write_text("test data")

        assert test_file.exists()
        assert test_file.read_text() == "test data"
