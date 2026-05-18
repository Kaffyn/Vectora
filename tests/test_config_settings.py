"""Tests for vectora/config/settings.py"""

from __future__ import annotations

from vectora.config.settings import Settings


def test_settings_loads():
    s = Settings()
    assert s is not None


def test_settings_derived_paths():
    s = Settings()
    assert s.lancedb_dir is not None
    assert str(s.lancedb_dir).endswith("lancedb")


def test_settings_embedding_queue_dsn():
    s = Settings()
    assert "sqlite" in s.embedding_queue_dsn


def test_settings_defaults():
    s = Settings()
    assert isinstance(s.enable_rag, bool)
    assert isinstance(s.debug_mode, bool)
    assert isinstance(s.max_context_tokens, int)
    assert s.max_context_tokens > 0


def test_settings_get_cohere_api_key_returns_none_or_str():
    s = Settings()
    key = s.get_cohere_api_key()
    assert key is None or isinstance(key, str)
