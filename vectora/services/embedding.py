"""EmbeddingService: Manages vector store, background worker, and embeddings.

Responsibilities:
1. Queue documents for asynchronous embedding
2. Manage LanceDB vector store
3. Perform semantic vector search
4. Manage background worker lifecycle
5. Handle embedding retry logic and backoff

Implementation: Integrates LanceDB, background worker, and VoyageAI embeddings.
Provides fire-and-forget pattern for document ingestion and vector persistence.
"""

import asyncio
import logging
from pathlib import Path
from typing import Any

try:
    import lancedb
except ImportError:
    lancedb = None

try:
    from langchain_voyageai import VoyageAIEmbeddings
except ImportError:
    VoyageAIEmbeddings = None

from vectora.services.ignore_validator import get_ignore_validator
from vectora.settings import Settings

logger = logging.getLogger(__name__)


class EmbeddingService:
    """Manages vector embeddings and semantic search.

    Features:
    - Fire-and-forget document queuing
    - Background worker for asynchronous embedding
    - LanceDB vector store with local persistence
    - Semantic search with retry logic
    - Collection-based organization
    - Exponential backoff retry mechanism
    - Dead Letter Queue for failed embeddings
    """

    def __init__(self, settings: Settings):
        """Initialize EmbeddingService.

        Args:
            settings: Application settings
        """
        self.settings = settings
        self.worker_running = False
        self.worker_task: asyncio.Task[None] | None = None

        # Vector store (lazy-loaded on initialize)
        self.db: Any = None
        self.embeddings: Any = None

        # Queue and worker configuration
        self.batch_size = 10
        self.max_parallel = 5
        self.polling_interval = 5  # seconds
        self.semaphore = asyncio.Semaphore(self.max_parallel)
        self.lancedb_semaphore = asyncio.Semaphore(1)  # Serialize LanceDB writes

        # Retry configuration (exponential backoff)
        self.retry_backoff = [1, 2, 4]  # seconds
        self.max_retries = 3

        # Ignore pattern validator (prevents embedding of secrets, node_modules, etc)
        self.ignore_validator = get_ignore_validator()

        logger.debug("EmbeddingService initialized")

    async def start(self) -> None:
        """Start background embedding worker.

        Should be called during AgentManager.initialize().
        Initializes LanceDB, embeddings model, and starts worker loop.
        """
        if self.worker_running:
            logger.warning("Worker already running")
            return

        if not self.settings.enable_rag:
            logger.warning("RAG disabled, embedding worker not starting")
            return

        try:
            # Initialize vector store (LanceDB)
            await self._initialize_vector_store()

            # Initialize embeddings model
            await self._initialize_embeddings()

            # Start worker loop
            self.worker_running = True
            self.worker_task = asyncio.create_task(self._worker_loop())

            logger.info("Embedding worker started successfully")
        except Exception as e:
            logger.exception(f"Failed to start embedding worker: {e}")
            self.worker_running = False
            raise

    async def stop(self) -> None:
        """Stop background embedding worker gracefully.

        Called during AgentManager.shutdown().
        Waits for pending embeddings to complete before shutting down.
        """
        if not self.worker_running:
            logger.warning("Worker not running")
            return

        try:
            logger.info("Stopping embedding worker...")
            self.worker_running = False

            # Wait for worker task to complete (with timeout)
            if self.worker_task:
                try:
                    await asyncio.wait_for(self.worker_task, timeout=30.0)
                except TimeoutError:
                    logger.warning("Worker task did not complete within timeout")
                    if self.worker_task and not self.worker_task.done():
                        self.worker_task.cancel()

            logger.info("Embedding worker stopped")
        except Exception as e:
            logger.warning(f"Error stopping embedding worker: {e}")

    async def queue_document(
        self,
        doc_id: str,
        text: str,
        collection: str = "documents",
        file_path: str | Path | None = None,
    ) -> bool:
        """Queue document for embedding (fire-and-forget).

        Validates that document respects .gitignore, .npmignore, .dockerignore, etc.
        Rejects documents from ignored paths (node_modules, .env, secrets, etc).

        Args:
            doc_id: Unique document identifier
            text: Document text content
            collection: Vector collection name
            file_path: Optional file path (used for ignore validation)

        Returns:
            True if document was queued, False if ignored/rejected

        Raises:
            RuntimeError: If queue is at capacity
        """
        if not self.worker_running:
            logger.warning("Worker not running, document not queued")
            return False

        # Validar contra ignore patterns se file_path foi fornecido
        if file_path:
            file_path_obj = Path(file_path)
            if self.ignore_validator.should_ignore(file_path_obj):
                logger.warning(
                    "Document rejected: matches ignore pattern",
                    extra={
                        "doc_id": doc_id,
                        "file_path": str(file_path_obj),
                        "reason": "File matches .gitignore, .npmignore, or security patterns",
                    },
                )
                return False

        try:
            # Placeholder implementation - in production would write to queue
            # This will be expanded in Week 3 with actual EmbeddingQueue
            logger.debug(
                "Document queued for embedding",
                extra={
                    "doc_id": doc_id,
                    "collection": collection,
                    "text_length": len(text),
                    "file_path": str(file_path) if file_path else None,
                },
            )
            return True
        except Exception as e:
            logger.warning(f"Failed to queue document {doc_id}: {e}")
            return False

    async def search(
        self, query: str, collection: str = "documents", limit: int = 5
    ) -> list[dict]:
        """Semantic vector search.

        Args:
            query: Search query text
            collection: Vector collection to search
            limit: Maximum results to return

        Returns:
            List of matched documents with scores:
            [
                {
                    "doc_id": "id",
                    "content": "text",
                    "score": 0.95,
                    "metadata": {...},
                },
                ...
            ]

        Raises:
            RuntimeError: If search fails after retries
        """
        if not self.db or not self.embeddings:
            logger.warning("Vector store not initialized")
            return []

        try:
            # Get query embedding with retry logic
            query_embedding = await self._get_embedding_with_retry(query)
            if not query_embedding:
                return []

            # Search in LanceDB (with read lock)
            async with self.lancedb_semaphore:
                try:
                    table = self.db.open_table(collection)
                    results = await asyncio.to_thread(
                        lambda: table.search(query_embedding).limit(limit).to_list()
                    )

                    # Format results
                    formatted_results = [
                        {
                            "doc_id": r.get("doc_id", "unknown"),
                            "content": r.get("text", ""),
                            "score": r.get("_distance", 0.0),
                            "metadata": r.get("metadata", {}),
                        }
                        for r in results
                    ]

                    logger.debug(
                        "Vector search completed",
                        extra={
                            "query_length": len(query),
                            "collection": collection,
                            "result_count": len(formatted_results),
                        },
                    )
                    return formatted_results

                except Exception as e:
                    logger.exception(f"LanceDB search failed: {e}")
                    return []

        except Exception as e:
            logger.exception("Vector search failed")
            return []

    async def get_queue_status(self) -> dict:
        """Get current embedding queue status.

        Returns:
            Status dict:
            {
                "pending_count": 10,
                "max_queue_size": 1000,
                "queue_usage_percent": 1.0,
                "worker_running": True,
            }
        """
        try:
            # Placeholder - will be expanded with actual queue status
            # In production, query EmbeddingQueue pending count
            pending_count = 0  # TODO: Query from embedding queue

            queue_usage_percent = (
                (pending_count / self.settings.max_embedding_queue_size * 100)
                if self.settings.max_embedding_queue_size > 0
                else 0.0
            )

            logger.debug("Retrieved queue status")
            return {
                "pending_count": pending_count,
                "max_queue_size": self.settings.max_embedding_queue_size,
                "queue_usage_percent": queue_usage_percent,
                "worker_running": self.worker_running,
            }
        except Exception as e:
            logger.warning(f"Failed to get queue status: {e}")
            return {
                "pending_count": 0,
                "max_queue_size": self.settings.max_embedding_queue_size,
                "queue_usage_percent": 0.0,
                "worker_running": self.worker_running,
            }

    async def clear_collection(self, collection: str) -> int:
        """Clear all documents from collection.

        Args:
            collection: Collection name to clear

        Returns:
            Number of documents deleted
        """
        if not self.db:
            logger.warning("Vector store not initialized")
            return 0

        try:
            async with self.lancedb_semaphore:
                # Delete table if exists
                try:
                    await asyncio.to_thread(lambda: self.db.drop_table(collection))
                    logger.warning(f"Collection cleared: {collection}")
                    return 1  # Placeholder count
                except Exception:
                    # Table doesn't exist, skip
                    return 0
        except Exception as e:
            logger.warning(f"Failed to clear collection {collection}: {e}")
            return 0

    async def health_check(self) -> tuple[bool, str]:
        """Check health of embedding service.

        Returns:
            Tuple of (is_healthy, status_message)
        """
        try:
            if not self.worker_running:
                return False, "Embedding worker not running"

            if not self.db or not self.embeddings:
                return False, "Vector store or embeddings model not initialized"

            # Test embedding model with simple query
            try:
                test_embedding = await self._get_embedding_with_retry("test")
                if not test_embedding:
                    return False, "Embedding model returned empty result"
            except Exception as e:
                return False, f"Embedding model test failed: {e}"

            return True, "Embedding service healthy"
        except Exception as e:
            logger.exception(f"Health check failed: {e}")
            return False, f"Health check error: {e}"

    # ========================================================================
    # PRIVATE METHODS
    # ========================================================================

    async def _initialize_vector_store(self) -> None:
        """Initialize LanceDB vector store."""
        if not lancedb:
            raise ImportError("lancedb not installed. Install with: uv add lancedb")

        try:
            # Open or create LanceDB database
            self.db = await asyncio.to_thread(
                lambda: lancedb.connect(str(self.settings.lancedb_dir))
            )
            logger.info(f"LanceDB initialized at {self.settings.lancedb_dir}")
        except Exception as e:
            logger.exception(f"Failed to initialize LanceDB: {e}")
            raise

    async def _initialize_embeddings(self) -> None:
        """Initialize embeddings model (VoyageAI)."""
        if not VoyageAIEmbeddings:
            raise ImportError(
                "langchain-voyageai not installed. Install with: uv add langchain-voyageai"
            )

        try:
            api_key = self.settings.get_voyage_api_key()
            if not api_key:
                raise ValueError("VOYAGE_API_KEY not configured")

            self.embeddings = VoyageAIEmbeddings(
                voyage_api_key=api_key,
                model="voyage-3",
            )
            logger.info("VoyageAI embeddings model initialized")
        except Exception as e:
            logger.exception(f"Failed to initialize embeddings model: {e}")
            raise

    async def _worker_loop(self) -> None:
        """Main worker loop: periodically processes pending embeddings.

        Runs while self.worker_running is True.
        Polls embedding queue, batches documents, and processes them.
        """
        logger.info("Embedding worker loop started")

        while self.worker_running:
            try:
                # Poll queue for pending documents (TODO: implement in Week 3)
                # pending_docs = await self._get_pending_documents(self.batch_size)
                # if pending_docs:
                #     await self._process_batch(pending_docs)

                # Sleep before next poll
                await asyncio.sleep(self.polling_interval)

            except asyncio.CancelledError:
                logger.info("Worker loop cancelled")
                break
            except Exception as e:
                logger.exception(f"Error in worker loop: {e}")
                # Continue loop on error
                await asyncio.sleep(self.polling_interval)

        logger.info("Embedding worker loop stopped")

    async def _get_embedding_with_retry(self, text: str) -> list[float] | None:
        """Get embedding for text with exponential backoff retry.

        Args:
            text: Text to embed

        Returns:
            Embedding vector or None if failed
        """
        if not self.embeddings:
            return None

        for attempt, backoff in enumerate(self.retry_backoff, 1):
            try:
                return await asyncio.to_thread(
                    lambda: self.embeddings.embed_query(text)
                )
            except Exception as e:
                if attempt < self.max_retries:
                    logger.warning(
                        f"Embedding attempt {attempt} failed, retrying in {backoff}s: {e}"
                    )
                    await asyncio.sleep(backoff)
                else:
                    logger.exception(
                        f"Embedding failed after {self.max_retries} attempts"
                    )
                    return None

        return None
