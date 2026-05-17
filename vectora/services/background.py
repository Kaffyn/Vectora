"""Background Worker para Processamento de Embeddings Fire-and-Forget.

Loop assíncrono que:
1. Busca documentos pendentes da fila de embedding a cada 5 segundos
2. Processa até 10 documentos em paralelo (limitado por Semaphore(5))
3. Gera embeddings via Voyage AI
4. Escreve em LanceDB com idempotência (via queue_id como document ID)
5. Retry com exponential backoff (1s → 2s → 4s) até 3 tentativas
6. Move para DLQ após 3 falhas para auditoria manual
"""

import asyncio
import contextlib
import logging
import traceback
from pathlib import Path
from typing import Any

from pydantic import SecretStr

try:
    import lancedb
except ImportError:
    lancedb = None

try:
    import pyarrow as pa
except ImportError:
    pa = None

try:
    from langchain_voyageai import VoyageAIEmbeddings
except ImportError:
    VoyageAIEmbeddings = None

from vectora.config.settings import settings
from vectora.services.queue import EmbeddingQueueRecord, get_embedding_queue

logger = logging.getLogger(__name__)

# Tempos de backoff exponencial (segundos)
RETRY_BACKOFF = [1, 2, 4]  # 1s → 2s → 4s
MAX_RETRIES = 3
POLLING_INTERVAL = 5  # Busca a cada 5 segundos
MAX_PARALLEL = 5  # Max 5 embeddings simultâneos (Semaphore)
BATCH_SIZE = 10  # Processa até 10 registros por batch


class BackgroundEmbeddingWorker:
    """Worker assíncrono para processamento de embeddings em larga escala."""

    def __init__(self) -> None:
        """Inicializa o worker com configuração global."""
        self.config = settings
        self.running = False
        self.task: asyncio.Task[None] | None = None
        self.semaphore = asyncio.Semaphore(MAX_PARALLEL)
        # Semaphore(1) para proteger escritas em LanceDB contra race conditions
        # LanceDB não suporta múltiplas escritas simultâneas no mesmo diretório
        self.lancedb_semaphore = asyncio.Semaphore(1)
        # Contadores em memória para o painel /rag
        self.processed_count: int = 0
        self.failed_count: int = 0

    async def _get_queue(self) -> Any:
        """Obtém a queue (singleton lazy-loaded)."""
        return await get_embedding_queue(self.config.embedding_queue_dsn)

    async def start(self) -> None:
        """Inicia o worker como asyncio.Task."""
        if self.running:
            logger.warning("Worker já está rodando")
            return

        if not self.config.enable_rag:
            logger.warning("RAG desabilitado, worker não iniciando")
            return

        self.running = True

        # Executar reconciliação na startup (recuperar records travados)
        await self._reconcile_startup()

        self.task = asyncio.create_task(self._run_loop())
        logger.info("BackgroundEmbeddingWorker iniciado")

    async def stop(self, timeout_seconds: int = 30) -> None:
        """Para o worker gracefully.

        Args:
            timeout_seconds: Segundos para aguardar a terminação
        """
        if not self.running:
            return

        logger.info("Parando BackgroundEmbeddingWorker...")
        self.running = False

        if self.task:
            try:
                async with asyncio.timeout(timeout_seconds):
                    await self.task
            except TimeoutError:
                logger.warning(
                    "Worker não terminou a tempo, cancelando",
                    extra={"timeout_seconds": timeout_seconds},
                )
                self.task.cancel()
                with contextlib.suppress(asyncio.CancelledError):
                    await self.task

        logger.info("BackgroundEmbeddingWorker parou")

    async def _reconcile_startup(self) -> None:
        """Recupera records travados em 'processing' na startup."""
        try:
            queue = await self._get_queue()
            await queue.reconcile()
            logger.info("Reconciliação de startup concluída")
        except Exception:
            logger.exception("Erro ao reconciliar startup")

    async def _run_loop(self) -> None:
        """Loop principal: fetch pending → process → retry/success/dlq."""
        while self.running:
            try:
                # Obter queue (lazy-loaded do singleton)
                queue = await self._get_queue()

                # Buscar até BATCH_SIZE documentos pendentes
                pending = await queue.get_pending(limit=BATCH_SIZE)

                if not pending:
                    # Nenhum documento para processar, aguardar POLLING_INTERVAL
                    await asyncio.sleep(POLLING_INTERVAL)
                    continue

                logger.debug("Batch encontrado", extra={"count": len(pending)})

                # Processar batch em paralelo (limitado por Semaphore)
                await asyncio.gather(
                    *[self._process_record(record, queue) for record in pending],
                    return_exceptions=True,
                )

                # Pequena pausa antes do próximo batch
                await asyncio.sleep(1)

            except asyncio.CancelledError:
                logger.info("Worker foi cancelado")
                break
            except Exception:
                logger.exception("Erro no loop principal do worker")
                await asyncio.sleep(POLLING_INTERVAL)

    async def _process_record(
        self, record: EmbeddingQueueRecord, queue: Any | None = None
    ) -> None:
        """Processa um registro individual com retry exponencial.

        Args:
            record: Registro da fila de embedding
            queue: Fila de embedding (obtém do singleton se None)
        """
        if queue is None:
            queue = await self._get_queue()

        queue_id = record.queue_id
        attempt = 0

        while attempt < MAX_RETRIES:
            try:
                async with self.semaphore:
                    # Marcar como processing
                    await queue.mark_processing(queue_id)

                    # Gerar embedding via Voyage AI
                    embedding_vector = await self._generate_embedding(record.text)

                    # Escrever em LanceDB (idempotente via queue_id)
                    await self._write_to_lancedb(record, embedding_vector)

                    # Marcar como success
                    await queue.mark_success(queue_id)
                    self.processed_count += 1

                    logger.info(
                        "embedding_processed_success",
                        extra={
                            "queue_id": queue_id,
                            "collection": record.collection,
                        },
                    )
                    return  # Sucesso, sair do loop de retry

            except Exception as e:
                attempt += 1
                error_trace = traceback.format_exc()
                logger.warning(
                    "embedding_processing_failed",
                    extra={
                        "queue_id": queue_id,
                        "attempt": attempt,
                        "max_retries": MAX_RETRIES,
                        "error": str(e),
                        "traceback": error_trace,
                    },
                )

                if attempt < MAX_RETRIES:
                    # Exponential backoff antes de retry
                    backoff_time = RETRY_BACKOFF[attempt - 1]
                    logger.info(
                        "embedding_retry_backoff",
                        extra={
                            "queue_id": queue_id,
                            "backoff_seconds": backoff_time,
                            "attempt": attempt,
                        },
                    )
                    await asyncio.sleep(backoff_time)
                else:
                    # 3 falhas, mover para DLQ com stack trace completo
                    self.failed_count += 1
                    dlq_reason = f"{e!s}\n\nStack trace:\n{error_trace}"
                    try:
                        await queue.mark_dlq(queue_id, dlq_reason)
                    except Exception:
                        logger.exception(
                            "Erro ao mover para DLQ",
                            extra={"queue_id": queue_id},
                        )
                    else:
                        logger.info(
                            "embedding_moved_to_dlq",
                            extra={
                                "queue_id": queue_id,
                                "reason": str(e),
                            },
                        )

    async def _generate_embedding(self, text: str) -> list[float]:
        """Gera embedding via Voyage AI.

        Args:
            text: Texto para embeddar

        Returns:
            Lista de floats representando o embedding

        Raises:
            ValueError: Se VOYAGE_API_KEY não estiver configurado
            ImportError: Se langchain_voyageai não estiver instalado
        """
        if not self.config.voyage_api_key:
            msg = "VOYAGE_API_KEY não configurado"
            raise ValueError(msg)

        if VoyageAIEmbeddings is None:
            msg = "langchain_voyageai não está instalado"
            raise ImportError(msg)

        embeddings_model = VoyageAIEmbeddings(
            api_key=SecretStr(self.config.voyage_api_key),
            model=self.config.embedding_model,
        )

        # embed_query é bloqueante (HTTP síncrono ~2s por chunk).
        # asyncio.to_thread() move para a thread pool do SO, liberando o event loop
        # para o spinner da UI e demais tarefas async enquanto a API responde.
        return await asyncio.to_thread(embeddings_model.embed_query, text)

    async def _write_to_lancedb(
        self, record: EmbeddingQueueRecord, vector: list[float]
    ) -> None:
        """Escreve documento em LanceDB (idempotente via queue_id).

        Args:
            record: Registro com metadata
            vector: Embedding vector

        Raises:
            ImportError: Se lancedb não estiver instalado
        """
        if lancedb is None:
            msg = "lancedb não está instalado"
            raise ImportError(msg)

        db = await lancedb.connect_async(str(Path(self.config.lancedb_dir)))

        # Schema para documento
        schema = pa.schema(
            [
                pa.field("id", pa.string()),
                pa.field("vector", pa.list_(pa.float32(), len(vector))),
                pa.field("text", pa.string()),
                pa.field("metadata", pa.string()),
            ]
        )

        # Abrir ou criar tabela
        try:
            table = await db.open_table(record.collection)
        except Exception:
            table = await db.create_table(record.collection, schema=schema)
            logger.info(
                "lancedb_table_created",
                extra={"collection": record.collection},
            )

        # Adicionar documento (queue_id como document ID para idempotência)
        doc = {
            "id": record.queue_id,  # Chave primária = queue_id
            "vector": vector,
            "text": record.text,
            "metadata": record.doc_metadata or "{}",
        }

        # Protege escrita em LanceDB com semaphore(1) contra race conditions
        async with self.lancedb_semaphore:
            await table.add([doc])

        logger.debug(
            "lancedb_document_written",
            extra={
                "queue_id": record.queue_id,
                "collection": record.collection,
            },
        )


# Singleton global com lock
_worker: BackgroundEmbeddingWorker | None = None
_worker_lock: asyncio.Lock = asyncio.Lock()


async def get_background_worker() -> BackgroundEmbeddingWorker:
    """Obtém ou cria instância singleton do worker (thread-safe).

    Returns:
        Instância do BackgroundEmbeddingWorker

    Note:
        Usa asyncio.Lock para evitar race condition em múltiplas
        inicializações simultâneas.
    """
    global _worker

    if _worker is not None:
        return _worker

    async with _worker_lock:
        # Double-check após adquirir lock
        if _worker is None:
            _worker = BackgroundEmbeddingWorker()

    return _worker
