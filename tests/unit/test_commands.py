"""Tests for command dispatcher module."""

import pytest
from rich.console import Console

from vectora.commands import (
    AVAILABLE_MODELS,
    get_available_models,
    handle_command,
)


class TestAvailableModels:
    """Test model registry."""

    def test_models_exist_for_all_providers(self):
        """Test that models are defined for all providers."""
        assert "google-genai" in AVAILABLE_MODELS
        assert "openai" in AVAILABLE_MODELS
        assert "anthropic" in AVAILABLE_MODELS
        assert "ollama" in AVAILABLE_MODELS

    def test_each_provider_has_models(self):
        """Test that each provider has at least one model."""
        for provider, models in AVAILABLE_MODELS.items():
            assert len(models) > 0, f"No models defined for {provider}"

    def test_google_models_exist(self):
        """Test Google GenAI models."""
        google_models = AVAILABLE_MODELS["google-genai"]
        assert "gemini-2.5-flash" in google_models
        assert "gemini-2.0-flash" in google_models

    def test_openai_models_exist(self):
        """Test OpenAI models."""
        openai_models = AVAILABLE_MODELS["openai"]
        assert "gpt-5.5" in openai_models
        assert "gpt-5.4" in openai_models
        assert "gpt-4o" in openai_models

    def test_anthropic_models_exist(self):
        """Test Anthropic models."""
        anthropic_models = AVAILABLE_MODELS["anthropic"]
        assert "claude-opus-4-1" in anthropic_models


class TestGetAvailableModels:
    """Test get_available_models function."""

    def test_get_all_models(self):
        """Test getting all models returns all providers."""
        all_models = get_available_models()
        assert len(all_models) == 4
        assert all(
            p in all_models for p in ["google-genai", "openai", "anthropic", "ollama"]
        )

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
        """Test /quit command returns True."""
        console = Console()
        from vectora.config import Config

        config = Config.instance()
        result = await handle_command("/quit", config, console)
        assert result is True

    @pytest.mark.asyncio
    async def test_sair_command(self):
        """Test /sair command returns True."""
        console = Console()
        from vectora.config import Config

        config = Config.instance()
        result = await handle_command("/sair", config, console)
        assert result is True

    @pytest.mark.asyncio
    async def test_q_command(self):
        """Test /q command returns True."""
        console = Console()
        from vectora.config import Config

        config = Config.instance()
        result = await handle_command("/q", config, console)
        assert result is True

    @pytest.mark.asyncio
    async def test_model_command_no_args(self):
        """Test /model command without args."""
        console = Console()
        from vectora.config import Config

        config = Config.instance()
        result = await handle_command("/model", config, console)
        assert result is False

    @pytest.mark.asyncio
    async def test_unknown_command(self):
        """Test unknown command."""
        console = Console()
        from vectora.config import Config

        config = Config.instance()
        result = await handle_command("/unknown", config, console)
        assert result is False

    @pytest.mark.asyncio
    async def test_help_command(self):
        """Test /help command."""
        console = Console()
        from vectora.config import Config

        config = Config.instance()
        result = await handle_command("/help", config, console)
        assert result is False
