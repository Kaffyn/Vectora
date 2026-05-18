"""Tests for vectora/services/memory.py"""

from __future__ import annotations

import pytest

from vectora.services.memory import MemoryStore


@pytest.fixture
async def store():
    s = MemoryStore()
    await s.initialize()
    return s


class TestMemoryStore:
    @pytest.mark.asyncio
    async def test_save_and_get_all(self, store):
        await store.save("user1", "nome", "Bruno")
        memories = await store.get_all("user1")
        keys = [m["key"] for m in memories]
        assert "nome" in keys

    @pytest.mark.asyncio
    async def test_get_all_unknown_user_empty(self, store):
        memories = await store.get_all("user_xyz_desconhecido_99")
        assert memories == []

    @pytest.mark.asyncio
    async def test_delete_removes_key(self, store):
        await store.save("user2", "chave", "valor")
        await store.delete("user2", "chave")
        memories = await store.get_all("user2")
        keys = [m["key"] for m in memories]
        assert "chave" not in keys

    @pytest.mark.asyncio
    async def test_save_overwrites_existing(self, store):
        await store.save("user3", "pref", "A")
        await store.save("user3", "pref", "B")
        memories = await store.get_all("user3")
        pref = next((m for m in memories if m["key"] == "pref"), None)
        assert pref is not None
        assert pref["content"] == "B"
