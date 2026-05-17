"""Memory tools: persistência de memórias entre sessões."""

import json
import logging
from typing import Any

from langchain.tools import tool

logger = logging.getLogger(__name__)


@tool
async def save_memory(
    key: str,
    content: str,
    metadata: dict[str, Any] | None = None,
    ttl_days: int | None = None,
) -> str:
    """Salva uma memória persistente global para uso em futuras conversas.

    As memórias são armazenadas no SQLite e recuperadas automaticamente
    ao iniciar novas sessões com o mesmo usuário.

    Args:
        key: Chave única da memória (ex: 'user_preferences', 'project_context')
        content: Conteúdo da memória (string com informações a recordar)
        metadata: Metadados adicionais
        ttl_days: Dias até expiração automática (None = nunca expira)

    Returns:
        JSON com status saved/failed
    """
    try:
        from vectora.memory_store import get_memory_store

        user_id = "default_user"
        memory_store = await get_memory_store()
        memory_id = await memory_store.save(
            user_id=user_id,
            key=key,
            content=content,
            metadata=metadata,
            ttl_days=ttl_days,
        )

        logger.info(
            "memory_saved",
            extra={"key": key, "memory_id": memory_id, "ttl_days": ttl_days},
        )

        return json.dumps(
            {
                "status": "saved",
                "memory_id": memory_id,
                "key": key,
                "expires_in_days": ttl_days,
            }
        )

    except Exception as e:
        logger.exception("save_memory_failed", extra={"key": key})
        return json.dumps({"status": "failed", "error": str(e), "key": key})


@tool
async def get_memory(key: str | None = None) -> str:
    """Recupera memórias persistentes salvas anteriormente.

    Args:
        key: Chave da memória específica. Se None, retorna todas as memórias.

    Returns:
        JSON com conteúdo da memória ou lista de memórias
    """
    try:
        from vectora.memory_store import get_memory_store

        user_id = "default_user"
        memory_store = await get_memory_store()

        if key is not None:
            memory = await memory_store.get(user_id, key)
            if memory is None:
                logger.warning("memory_not_found", extra={"key": key})
                return json.dumps({"status": "not_found", "key": key})

            logger.debug("memory_retrieved", extra={"key": key})
            return json.dumps(
                {
                    "status": "found",
                    "key": key,
                    "content": memory["content"],
                    "metadata": memory["metadata"],
                    "updated_at": memory["updated_at"],
                }
            )

        all_memories = await memory_store.get_all(user_id)
        logger.debug("all_memories_retrieved", extra={"count": len(all_memories)})
        return json.dumps(
            {
                "status": "success",
                "count": len(all_memories),
                "memories": [
                    {
                        "key": m["key"],
                        "content": m["content"],
                        "metadata": m["metadata"],
                        "updated_at": m["updated_at"],
                    }
                    for m in all_memories
                ],
            }
        )

    except Exception as e:
        logger.exception("get_memory_failed", extra={"key": key})
        return json.dumps({"status": "failed", "error": str(e)})


@tool
async def delete_memory(key: str) -> str:
    """Deleta uma memória persistente.

    Args:
        key: Chave da memória a deletar

    Returns:
        JSON com status deleted/not_found/failed
    """
    try:
        from vectora.memory_store import get_memory_store

        user_id = "default_user"
        memory_store = await get_memory_store()
        deleted = await memory_store.delete(user_id, key)

        if deleted:
            logger.info("memory_deleted", extra={"key": key})
            return json.dumps({"status": "deleted", "key": key})

        logger.warning("memory_not_found_for_deletion", extra={"key": key})
        return json.dumps({"status": "not_found", "key": key})

    except Exception as e:
        logger.exception("delete_memory_failed", extra={"key": key})
        return json.dumps({"status": "failed", "error": str(e), "key": key})
