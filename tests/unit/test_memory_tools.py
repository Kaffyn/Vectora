"""Testes para as tools de memória persistente.

Nota: Os testes aqui testam principalmente o MemoryStore diretamente,
não as tools decoradas com @tool (que são StructuredTool objects).
As tools serão testadas como parte dos testes E2E do grafo.
"""

import tempfile
from pathlib import Path

import pytest

from vectora.services.memory import MemoryStore


@pytest.fixture
async def temp_memory_store():
    """Cria um memory store com banco de dados temporário."""
    with tempfile.TemporaryDirectory() as tmpdir:
        db_path = Path(tmpdir) / "test_memory_tools.db"
        store = MemoryStore(db_dsn=str(db_path))
        await store.initialize()
        yield store


@pytest.mark.asyncio
async def test_memory_tool_save_basic(temp_memory_store: MemoryStore):
    """Testa salvar memória básica."""
    memory_id = await temp_memory_store.save(
        user_id="user_test",
        key="test_key",
        content="Test content",
        metadata={"source": "test"},
    )

    assert memory_id == "user_test:test_key"
    retrieved = await temp_memory_store.get("user_test", "test_key")
    assert retrieved is not None
    assert retrieved["content"] == "Test content"


@pytest.mark.asyncio
async def test_memory_tool_save_with_ttl(temp_memory_store: MemoryStore):
    """Testa salvar memória com TTL."""
    await temp_memory_store.save(
        user_id="user_ttl",
        key="ttl_key",
        content="Temporary",
        ttl_days=7,
    )

    retrieved = await temp_memory_store.get("user_ttl", "ttl_key")
    assert retrieved is not None
    assert retrieved["expires_at"] is not None


@pytest.mark.asyncio
async def test_memory_tool_get_by_key(temp_memory_store: MemoryStore):
    """Testa recuperar memória específica."""
    await temp_memory_store.save(
        user_id="user_get",
        key="get_test",
        content="Content to retrieve",
        metadata={"type": "test"},
    )

    retrieved = await temp_memory_store.get("user_get", "get_test")
    assert retrieved is not None
    assert retrieved["content"] == "Content to retrieve"
    assert retrieved["metadata"]["type"] == "test"


@pytest.mark.asyncio
async def test_memory_tool_get_all(temp_memory_store: MemoryStore):
    """Testa recuperar todas as memórias."""
    user_id = "user_all"

    await temp_memory_store.save(user_id, "mem1", "Content 1")
    await temp_memory_store.save(user_id, "mem2", "Content 2")
    await temp_memory_store.save(user_id, "mem3", "Content 3")

    all_memories = await temp_memory_store.get_all(user_id)

    assert len(all_memories) == 3
    keys = {m["key"] for m in all_memories}
    assert keys == {"mem1", "mem2", "mem3"}


@pytest.mark.asyncio
async def test_memory_tool_get_not_found(temp_memory_store: MemoryStore):
    """Testa recuperar memória inexistente."""
    retrieved = await temp_memory_store.get("user_notfound", "nonexistent_key")
    assert retrieved is None


@pytest.mark.asyncio
async def test_memory_tool_delete(temp_memory_store: MemoryStore):
    """Testa deletar memória."""
    await temp_memory_store.save("user_delete", "to_delete", "Will be deleted")

    deleted = await temp_memory_store.delete("user_delete", "to_delete")
    assert deleted is True

    retrieved = await temp_memory_store.get("user_delete", "to_delete")
    assert retrieved is None


@pytest.mark.asyncio
async def test_memory_tool_delete_not_found(temp_memory_store: MemoryStore):
    """Testa deletar memória inexistente."""
    deleted = await temp_memory_store.delete("user_never", "never_existed")
    assert deleted is False


@pytest.mark.asyncio
async def test_memory_tool_special_characters(temp_memory_store: MemoryStore):
    """Testa memória com caracteres especiais."""
    special_content = "Conteúdo com acentos, émojis 🎉, e símbolos @#$%^&*()"
    metadata = {"type": "especial", "language": "português"}

    await temp_memory_store.save(
        user_id="user_special",
        key="special_chars",
        content=special_content,
        metadata=metadata,
    )

    retrieved = await temp_memory_store.get("user_special", "special_chars")
    assert retrieved is not None
    assert retrieved["content"] == special_content
    assert retrieved["metadata"]["language"] == "português"


@pytest.mark.asyncio
async def test_memory_tool_empty_content(temp_memory_store: MemoryStore):
    """Testa salvar memória com conteúdo vazio."""
    await temp_memory_store.save("user_empty", "empty", "")

    retrieved = await temp_memory_store.get("user_empty", "empty")
    assert retrieved is not None
    assert retrieved["content"] == ""


@pytest.mark.asyncio
async def test_memory_tool_overwrite(temp_memory_store: MemoryStore):
    """Testa sobrescrever memória existente."""
    user_id = "user_overwrite"
    key = "overwrite_test"

    await temp_memory_store.save(user_id, key, "Version 1")
    first = await temp_memory_store.get(user_id, key)
    assert first is not None

    await temp_memory_store.save(user_id, key, "Version 2 - Updated")
    second = await temp_memory_store.get(user_id, key)
    assert second is not None

    assert first["content"] == "Version 1"
    assert second["content"] == "Version 2 - Updated"
    assert second["updated_at"] >= first["updated_at"]
