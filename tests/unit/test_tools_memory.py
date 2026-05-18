"""Testes para tools/memory.py.

Cobre: save_memory, get_memory, delete_memory
"""

from __future__ import annotations

import json
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

# Patch target: get_memory_store é importado dentro das funções como
# `from vectora.services.memory import get_memory_store`
_PATCH_TARGET = "vectora.services.memory.get_memory_store"


class TestSaveMemory:
    """Testa save_memory tool."""

    @pytest.mark.asyncio
    async def test_save_memory_returns_saved_status(self):
        """Verifica que save_memory retorna status 'saved' em sucesso."""
        from vectora.tools.memory import save_memory

        mock_store = AsyncMock()
        mock_store.save = AsyncMock(return_value="mem-123")

        with patch(_PATCH_TARGET, AsyncMock(return_value=mock_store)):
            result = await save_memory.ainvoke(
                {
                    "key": "test_key",
                    "content": "conteúdo de teste",
                }
            )

        data = json.loads(result)
        assert data["status"] == "saved"
        assert data["key"] == "test_key"
        assert data["memory_id"] == "mem-123"

    @pytest.mark.asyncio
    async def test_save_memory_with_ttl(self):
        """Verifica que save_memory aceita ttl_days."""
        from vectora.tools.memory import save_memory

        mock_store = AsyncMock()
        mock_store.save = AsyncMock(return_value="mem-456")

        with patch(_PATCH_TARGET, AsyncMock(return_value=mock_store)):
            result = await save_memory.ainvoke(
                {
                    "key": "temp_key",
                    "content": "conteúdo temporário",
                    "ttl_days": 7,
                }
            )

        data = json.loads(result)
        assert data["status"] == "saved"
        assert data["expires_in_days"] == 7

    @pytest.mark.asyncio
    async def test_save_memory_handles_exception(self):
        """Verifica que save_memory trata exceções graciosamente."""
        from vectora.tools.memory import save_memory

        with patch(_PATCH_TARGET, AsyncMock(side_effect=Exception("DB error"))):
            result = await save_memory.ainvoke(
                {
                    "key": "fail_key",
                    "content": "conteúdo",
                }
            )

        data = json.loads(result)
        assert data["status"] == "failed"
        assert "DB error" in data["error"]
        assert data["key"] == "fail_key"


class TestGetMemory:
    """Testa get_memory tool."""

    @pytest.mark.asyncio
    async def test_get_memory_specific_key_found(self):
        """Verifica que get_memory retorna memória pelo key."""
        from vectora.tools.memory import get_memory

        mock_store = AsyncMock()
        mock_store.get = AsyncMock(
            return_value={
                "content": "memória salva",
                "metadata": {},
                "updated_at": "2026-01-01T00:00:00",
            }
        )

        with patch(_PATCH_TARGET, AsyncMock(return_value=mock_store)):
            result = await get_memory.ainvoke({"key": "my_key"})

        data = json.loads(result)
        assert data["status"] == "found"
        assert data["key"] == "my_key"
        assert data["content"] == "memória salva"

    @pytest.mark.asyncio
    async def test_get_memory_key_not_found(self):
        """Verifica que get_memory retorna not_found quando key não existe."""
        from vectora.tools.memory import get_memory

        mock_store = AsyncMock()
        mock_store.get = AsyncMock(return_value=None)

        with patch(_PATCH_TARGET, AsyncMock(return_value=mock_store)):
            result = await get_memory.ainvoke({"key": "missing_key"})

        data = json.loads(result)
        assert data["status"] == "not_found"
        assert data["key"] == "missing_key"

    @pytest.mark.asyncio
    async def test_get_memory_all(self):
        """Verifica que get_memory sem key retorna todas as memórias."""
        from vectora.tools.memory import get_memory

        mock_store = AsyncMock()
        mock_store.get_all = AsyncMock(
            return_value=[
                {
                    "key": "k1",
                    "content": "c1",
                    "metadata": {},
                    "updated_at": "2026-01-01",
                },
                {
                    "key": "k2",
                    "content": "c2",
                    "metadata": {},
                    "updated_at": "2026-01-02",
                },
            ]
        )

        with patch(_PATCH_TARGET, AsyncMock(return_value=mock_store)):
            result = await get_memory.ainvoke({})

        data = json.loads(result)
        assert data["status"] == "success"
        assert data["count"] == 2
        assert len(data["memories"]) == 2

    @pytest.mark.asyncio
    async def test_get_memory_handles_exception(self):
        """Verifica que get_memory trata exceções."""
        from vectora.tools.memory import get_memory

        with patch(_PATCH_TARGET, AsyncMock(side_effect=RuntimeError("crash"))):
            result = await get_memory.ainvoke({"key": "any"})

        data = json.loads(result)
        assert data["status"] == "failed"
        assert "crash" in data["error"]


class TestDeleteMemory:
    """Testa delete_memory tool."""

    @pytest.mark.asyncio
    async def test_delete_memory_success(self):
        """Verifica que delete_memory retorna 'deleted' em sucesso."""
        from vectora.tools.memory import delete_memory

        mock_store = AsyncMock()
        mock_store.delete = AsyncMock(return_value=True)

        with patch(_PATCH_TARGET, AsyncMock(return_value=mock_store)):
            result = await delete_memory.ainvoke({"key": "to_delete"})

        data = json.loads(result)
        assert data["status"] == "deleted"
        assert data["key"] == "to_delete"

    @pytest.mark.asyncio
    async def test_delete_memory_not_found(self):
        """Verifica que delete_memory retorna 'not_found' quando key não existe."""
        from vectora.tools.memory import delete_memory

        mock_store = AsyncMock()
        mock_store.delete = AsyncMock(return_value=False)

        with patch(_PATCH_TARGET, AsyncMock(return_value=mock_store)):
            result = await delete_memory.ainvoke({"key": "missing"})

        data = json.loads(result)
        assert data["status"] == "not_found"

    @pytest.mark.asyncio
    async def test_delete_memory_handles_exception(self):
        """Verifica que delete_memory trata exceções."""
        from vectora.tools.memory import delete_memory

        with patch(_PATCH_TARGET, AsyncMock(side_effect=Exception("fail"))):
            result = await delete_memory.ainvoke({"key": "crash_key"})

        data = json.loads(result)
        assert data["status"] == "failed"
        assert data["key"] == "crash_key"
