"""Testes para o gerenciador de memórias persistentes."""

import asyncio
import tempfile
from datetime import UTC, datetime
from pathlib import Path

import pytest

from memory_store import MemoryStore, get_memory_store


@pytest.fixture
async def temp_memory_store() -> MemoryStore:
    """Cria um memory store com banco de dados temporário."""
    with tempfile.TemporaryDirectory() as tmpdir:
        db_path = Path(tmpdir) / "test_memories.db"
        store = MemoryStore(db_dsn=f"file:///{db_path}")
        await store.initialize()
        yield store


@pytest.mark.asyncio
async def test_memory_store_initialize():
    """Testa inicialização do memory store."""
    with tempfile.TemporaryDirectory() as tmpdir:
        db_path = Path(tmpdir) / "test_init.db"
        store = MemoryStore(db_dsn=f"file:///{db_path}")
        await store.initialize()
        assert db_path.exists()


@pytest.mark.asyncio
async def test_save_and_retrieve_memory(temp_memory_store: MemoryStore):
    """Testa salvar e recuperar uma memória."""
    user_id = "user_123"
    key = "user_preferences"
    content = "Preferência por português, gosta de código limpo"
    metadata = {"category": "preferences", "importance": "high"}

    # Salva memória
    memory_id = await temp_memory_store.save(
        user_id=user_id,
        key=key,
        content=content,
        metadata=metadata,
    )

    assert memory_id == f"{user_id}:{key}"

    # Recupera memória
    retrieved = await temp_memory_store.get(user_id, key)
    assert retrieved is not None
    assert retrieved["content"] == content
    assert retrieved["metadata"] == metadata


@pytest.mark.asyncio
async def test_update_existing_memory(temp_memory_store: MemoryStore):
    """Testa atualizar uma memória existente."""
    user_id = "user_456"
    key = "project_context"

    # Primeira versão
    await temp_memory_store.save(user_id, key, "Contexto v1")
    first = await temp_memory_store.get(user_id, key)
    first_updated = first["updated_at"]

    # Aguarda para garantir diferença de timestamp
    await asyncio.sleep(0.01)

    # Segunda versão
    await temp_memory_store.save(user_id, key, "Contexto v2 atualizado")
    second = await temp_memory_store.get(user_id, key)

    assert second["content"] == "Contexto v2 atualizado"
    assert second["updated_at"] > first_updated


@pytest.mark.asyncio
async def test_memory_with_ttl(temp_memory_store: MemoryStore):
    """Testa memória com TTL (time-to-live)."""
    user_id = "user_ttl"
    key = "temp_data"
    content = "Dados temporários"

    # Salva com TTL de 1 dia
    memory_id = await temp_memory_store.save(
        user_id=user_id,
        key=key,
        content=content,
        ttl_days=1,
    )

    retrieved = await temp_memory_store.get(user_id, key)
    assert retrieved is not None
    assert retrieved["expires_at"] is not None


@pytest.mark.asyncio
async def test_get_all_memories(temp_memory_store: MemoryStore):
    """Testa recuperar todas as memórias de um usuário."""
    user_id = "user_multi"

    # Salva múltiplas memórias
    memories_data = [
        ("pref1", "Preferência 1"),
        ("pref2", "Preferência 2"),
        ("context", "Contexto do projeto"),
    ]

    for key, content in memories_data:
        await temp_memory_store.save(user_id, key, content)

    # Recupera todas
    all_memories = await temp_memory_store.get_all(user_id)

    assert len(all_memories) == 3
    keys = {m["key"] for m in all_memories}
    assert keys == {"pref1", "pref2", "context"}


@pytest.mark.asyncio
async def test_delete_memory(temp_memory_store: MemoryStore):
    """Testa deletar uma memória."""
    user_id = "user_delete"
    key = "to_delete"

    # Salva
    await temp_memory_store.save(user_id, key, "Será deletado")

    # Verifica que existe
    assert await temp_memory_store.get(user_id, key) is not None

    # Deleta
    deleted = await temp_memory_store.delete(user_id, key)
    assert deleted is True

    # Verifica que foi deletado
    assert await temp_memory_store.get(user_id, key) is None

    # Tenta deletar novamente
    deleted_again = await temp_memory_store.delete(user_id, key)
    assert deleted_again is False


@pytest.mark.asyncio
async def test_memory_isolation_between_users(temp_memory_store: MemoryStore):
    """Testa que memórias de um usuário não afetam outro."""
    # Salva memória para user1
    await temp_memory_store.save("user1", "key1", "Conteúdo user1")

    # Salva memória para user2
    await temp_memory_store.save("user2", "key1", "Conteúdo user2")

    # Verifica isolamento
    mem_user1 = await temp_memory_store.get("user1", "key1")
    mem_user2 = await temp_memory_store.get("user2", "key1")

    assert mem_user1["content"] == "Conteúdo user1"
    assert mem_user2["content"] == "Conteúdo user2"


@pytest.mark.asyncio
async def test_cleanup_expired_memories(temp_memory_store: MemoryStore):
    """Testa limpeza de memórias expiradas."""
    user_id = "user_cleanup"

    # Salva memória que nunca expira
    await temp_memory_store.save(user_id, "permanent", "Permanente")

    # Salva memória que expirou (TTL de -1 dia)
    from datetime import timedelta

    import aiosqlite

    expired_date = (datetime.now(UTC) - timedelta(days=1)).isoformat()
    async with aiosqlite.connect(temp_memory_store.db_dsn) as db:
        await db.execute(
            """
            INSERT INTO memories
            (id, user_id, key, content, metadata, created_at, updated_at, expires_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                "user_cleanup:expired",
                user_id,
                "expired",
                "Deve ser removido",
                "{}",
                datetime.now(UTC).isoformat(),
                datetime.now(UTC).isoformat(),
                expired_date,
            ),
        )
        await db.commit()

    # Verifica que expirado está aí antes de limpar
    assert await temp_memory_store.get(user_id, "expired") is None  # Já filtrado

    # Limpa expirados
    removed = await temp_memory_store.cleanup_expired()
    assert removed == 1

    # Verifica que permanente ainda existe
    perm = await temp_memory_store.get(user_id, "permanent")
    assert perm is not None
    assert perm["content"] == "Permanente"


@pytest.mark.asyncio
async def test_memory_store_singleton(temp_memory_store: MemoryStore):
    """Testa que get_memory_store retorna a mesma instância (lazy init)."""
    # Nota: este teste é limitado pois get_memory_store tem instância global
    # Aqui apenas verificamos que a inicialização não falha
    store = await get_memory_store()
    assert isinstance(store, MemoryStore)


@pytest.mark.asyncio
async def test_memory_with_empty_metadata(temp_memory_store: MemoryStore):
    """Testa salvar memória com metadata vazia."""
    user_id = "user_empty_meta"
    key = "simple"
    content = "Conteúdo simples"

    await temp_memory_store.save(user_id, key, content)
    retrieved = await temp_memory_store.get(user_id, key)

    assert retrieved["content"] == content
    assert retrieved["metadata"] == {}
