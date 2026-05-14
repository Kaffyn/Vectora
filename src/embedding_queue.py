import asyncio
import json
import logging
from datetime import UTC, datetime
from typing import Any, Self
from uuid import uuid4

from sqlalchemy import Column, DateTime, Integer, String, Text
from sqlalchemy.ext.asyncio import AsyncEngine, AsyncSession, create_async_engine
from sqlalchemy.orm import declarative_base, sessionmaker

logger = logging.getLogger(__name__)

Base = declarative_base()


class EmbeddingQueueRecord(Base):  # type: ignore[valid-type,misc]
    """Modelo SQLAlchemy para registros da fila de embedding."""

    __tablename__ = "embedding_queue"

    id = Column(Integer, primary_key=True)
    queue_id = Column(String(36), unique=True, nullable=False)
    text = Column(Text, nullable=False)
    collection = Column(String(255), nullable=False)
    metadata = Column(String(4096), nullable=True)  # String JSON
    status = Column(
        String(20), default="pending"
    )  # pendente, processando, sucesso, falha
    error_message = Column(Text, nullable=True)
    attempt_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=lambda: datetime.now(UTC))
    processed_at = Column(DateTime, nullable=True)


class EmbeddingQueue:
    """Gerenciador de fila para embedding de documentos quando API Voyage falha."""

    def __init__(self: Self, db_url: str) -> None:
        """Inicializa fila de embedding com conexão de banco de dados."""
        self.db_url = db_url
        self.engine: AsyncEngine | None = None
        self.AsyncSessionLocal: sessionmaker[AsyncSession] | None = None

    async def init(self) -> None:
        """Inicializa motor de banco de dados assíncrono e cria tabelas."""
        self.engine = create_async_engine(self.db_url, echo=False)

        async with self.engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)  # type: ignore[attr-defined]

        self.AsyncSessionLocal = sessionmaker(
            self.engine, class_=AsyncSession, expire_on_commit=False
        )

        logger.info("embedding_queue_initialized", extra={"db_url": self.db_url})

    async def enqueue(
        self, text: str, collection: str, metadata: dict[str, Any] | None = None
    ) -> str:
        """Enfileira um documento para embedding.

        Args:
            text: Texto do documento
            collection: Nome da coleção Qdrant
            metadata: Dicionário de metadados opcional

        Returns:
            ID da fila para rastreamento
        """
        queue_id = str(uuid4())
        metadata_json = json.dumps(metadata or {})

        try:
            if self.AsyncSessionLocal is None:
                msg = "AsyncSessionLocal não foi inicializado"
                raise RuntimeError(msg)
            async with self.AsyncSessionLocal() as session:
                record = EmbeddingQueueRecord(
                    queue_id=queue_id,
                    text=text,
                    collection=collection,
                    metadata=metadata_json,
                    status="pending",
                )
                session.add(record)
                await session.commit()

            logger.info(
                "embedding_enqueued",
                extra={
                    "queue_id": queue_id,
                    "collection": collection,
                    "text_length": len(text),
                },
            )

            return queue_id

        except Exception:
            logger.exception(
                "embedding_queue_insert_failed",
                extra={"queue_id": queue_id},
            )
            raise

    async def get_pending(self, limit: int = 10) -> list[EmbeddingQueueRecord]:
        """Obtém documentos pendentes da fila.

        Args:
            limit: Máximo de registros a retornar

        Returns:
            Lista de registros pendentes da fila de embedding
        """
        try:
            if self.AsyncSessionLocal is None:
                msg = "AsyncSessionLocal não foi inicializado"
                raise RuntimeError(msg)
            async with self.AsyncSessionLocal() as session:
                from sqlalchemy import and_, select

                query = select(EmbeddingQueueRecord).where(
                    and_(
                        EmbeddingQueueRecord.status == "pending",
                        EmbeddingQueueRecord.attempt_count < 3,  # Max 3 retries
                    )
                )
                result = await session.execute(query)
                records = result.scalars().all()

                logger.debug(
                    "embedding_queue_get_pending",
                    extra={"count": len(records), "limit": limit},
                )

                return records[:limit]

        except Exception:
            logger.exception("embedding_queue_get_pending_failed")
            return []

    async def mark_processing(self, queue_id: str) -> None:
        """Marca um registro da fila como processando.

        Args:
            queue_id: ID do registro a marcar
        """
        try:
            if self.AsyncSessionLocal is None:
                msg = "AsyncSessionLocal não foi inicializado"
                raise RuntimeError(msg)
            async with self.AsyncSessionLocal() as session:
                from sqlalchemy import update

                query = (
                    update(EmbeddingQueueRecord)
                    .where(EmbeddingQueueRecord.queue_id == queue_id)
                    .values(
                        status="processing",
                        attempt_count=EmbeddingQueueRecord.attempt_count + 1,
                    )
                )
                await session.execute(query)
                await session.commit()

        except Exception:
            logger.exception(
                "embedding_queue_mark_processing_failed",
                extra={"queue_id": queue_id},
            )

    async def mark_success(self, queue_id: str) -> None:
        """Marca um registro da fila como processado com sucesso.

        Args:
            queue_id: ID do registro a marcar
        """
        try:
            if self.AsyncSessionLocal is None:
                msg = "AsyncSessionLocal não foi inicializado"
                raise RuntimeError(msg)
            async with self.AsyncSessionLocal() as session:
                from sqlalchemy import update

                query = (
                    update(EmbeddingQueueRecord)
                    .where(EmbeddingQueueRecord.queue_id == queue_id)
                    .values(status="success", processed_at=datetime.now(UTC))
                )
                await session.execute(query)
                await session.commit()

            logger.info("embedding_queue_marked_success", extra={"queue_id": queue_id})

        except Exception:
            logger.exception(
                "embedding_queue_mark_success_failed",
                extra={"queue_id": queue_id},
            )

    async def mark_failed(self, queue_id: str, error_message: str) -> None:
        """Marca um registro da fila como falha.

        Args:
            queue_id: ID do registro a marcar
            error_message: Mensagem de erro
        """
        try:
            if self.AsyncSessionLocal is None:
                msg = "AsyncSessionLocal não foi inicializado"
                raise RuntimeError(msg)
            async with self.AsyncSessionLocal() as session:
                from sqlalchemy import update

                query = (
                    update(EmbeddingQueueRecord)
                    .where(EmbeddingQueueRecord.queue_id == queue_id)
                    .values(status="failed", error_message=error_message)
                )
                await session.execute(query)
                await session.commit()

            logger.error(
                "embedding_queue_marked_failed",
                extra={"queue_id": queue_id, "error": error_message},
            )

        except Exception:
            logger.exception(
                "embedding_queue_mark_failed_failed",
                extra={"queue_id": queue_id},
            )

    async def close(self) -> None:
        """Fecha a conexão do banco de dados."""
        if self.engine:
            await self.engine.dispose()
            logger.info("embedding_queue_closed")


# Instância singleton global com lock para evitar race condition em async
_queue: EmbeddingQueue | None = None
_queue_lock: asyncio.Lock = asyncio.Lock()


async def get_embedding_queue(db_url: str) -> EmbeddingQueue:
    """Obtém ou cria instância global da fila de embedding de forma thread-safe.

    O `asyncio.Lock` garante que apenas uma coroutine inicializa `_queue`,
    eliminando a race condition quando chamadas simultâneas chegam antes da
    primeira inicialização completar.
    """
    global _queue
    if _queue is not None:
        return _queue
    async with _queue_lock:
        # Double-check após adquirir o lock
        if _queue is None:
            _queue = EmbeddingQueue(db_url)
            await _queue.init()
    return _queue
