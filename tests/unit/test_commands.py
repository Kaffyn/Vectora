"""Tests for command dispatcher module."""

from unittest.mock import MagicMock

import pytest
from rich.console import Console

from vectora.ui.commands import (
    AVAILABLE_MODELS,
    get_available_models,
    handle_command,
)


class TestAvailableModels:
    """Test model registry."""

    def test_models_exist_for_core_providers(self):
        """Test that models are defined for core providers."""
        assert "google-genai" in AVAILABLE_MODELS
        assert "openai" in AVAILABLE_MODELS
        assert "anthropic" in AVAILABLE_MODELS

    def test_each_provider_has_models(self):
        """Test that each provider has at least one model."""
        for provider, models in AVAILABLE_MODELS.items():
            assert len(models) > 0, f"No models defined for {provider}"

    def test_google_models_exist(self):
        """Test Google GenAI models."""
        google_models = AVAILABLE_MODELS["google-genai"]
        assert len(google_models) > 0

    def test_openai_models_exist(self):
        """Test OpenAI models."""
        openai_models = AVAILABLE_MODELS["openai"]
        assert len(openai_models) > 0

    def test_anthropic_models_exist(self):
        """Test Anthropic models."""
        anthropic_models = AVAILABLE_MODELS["anthropic"]
        assert len(anthropic_models) > 0


class TestGetAvailableModels:
    """Test get_available_models function."""

    def test_get_all_models(self):
        """Test getting all models returns all providers."""
        all_models = get_available_models()
        assert len(all_models) >= 3
        assert all(p in all_models for p in ["google-genai", "openai", "anthropic"])

    def test_get_specific_provider_models(self):
        """Test getting models for specific provider."""
        google_models = get_available_models("google-genai")
        assert "google-genai" in google_models
        assert len(google_models) == 1

    def test_get_nonexistent_provider(self):
        """Test getting models for nonexistent provider."""
        result = get_available_models("nonexistent")
        assert "nonexistent" in result
        assert result["nonexistent"] == []


class TestHandleCommand:
    """Test command dispatcher."""

    @pytest.mark.asyncio
    async def test_quit_command(self):
        """Test /quit command returns True as exit signal."""
        console = Console(quiet=True)
        config = MagicMock()
        result = await handle_command("/quit", config, console)
        # handle_command retorna tuple[bool, Any, bool]
        assert isinstance(result, tuple)
        should_exit = result[0]
        assert should_exit is True

    @pytest.mark.asyncio
    async def test_sair_command(self):
        """Test /sair command returns exit signal True."""
        console = MagicMock()
        config = MagicMock()
        result = await handle_command("/sair", config, console)
        assert isinstance(result, tuple)
        assert result[0] is True

    @pytest.mark.asyncio
    async def test_q_command(self):
        """Test /q command returns exit signal True."""
        console = MagicMock()
        config = MagicMock()
        result = await handle_command("/q", config, console)
        assert isinstance(result, tuple)
        assert result[0] is True

    @pytest.mark.asyncio
    async def test_model_command_no_args(self):
        """Test /model command without args."""
        console = MagicMock()
        config = MagicMock()
        result = await handle_command("/model", config, console)
        assert isinstance(result, tuple)
        assert result[0] is False

    @pytest.mark.asyncio
    async def test_unknown_command(self):
        """Test unknown command does not exit."""
        console = MagicMock()
        config = MagicMock()
        result = await handle_command("/unknown_xyz", config, console)
        assert isinstance(result, tuple)
        assert result[0] is False

    @pytest.mark.asyncio
    async def test_help_command(self):
        """Test /help command does not exit."""
        console = MagicMock()
        config = MagicMock()
        result = await handle_command("/help", config, console)
        assert isinstance(result, tuple)
        assert result[0] is False
