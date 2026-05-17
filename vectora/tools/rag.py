"""RAG tools: embedding assíncrono, busca vetorial e ingestão de documentos."""

import asyncio
import json
import logging
from datetime import UTC, datetime
from typing import Any

from langchain.tools import tool
from langchain_community.document_loaders import DirectoryLoader, TextLoader
from langchain_core.documents import Document as LCDoc
from langchain_text_splitters import RecursiveCharacterTextSplitter

from vectora.services.queue import get_embedding_queue
from vectora.settings import settings

try:
    import lancedb
    from langchain_voyageai import VoyageAIEmbeddings, VoyageAIRerank
    from pydantic import SecretStr
except ImportError:
    lancedb = None
    VoyageAIEmbeddings = None
    VoyageAIRerank = None
    SecretStr = None

logger = logging.getLogger(__name__)


@tool
async def embedding(
    text: str, collection: str = "articles", metadata: dict[str, Any] | None = None
) -> str:
    """Enfileira documento para embedding assíncrono (fire-and-forget).

    Em vez de bloquear esperando Voyage AI (5+ segundos), enfileira o documento
    imediatamente. Um background worker processa a fila e indexa em LanceDB.

    Args:
        text: Texto do documento a indexar
        collection: Nome da coleção LanceDB (articles, wiki, api_docs, knowledge_base)
        metadata: Metadados opcionais (source, author, timestamp, etc)

    Returns:
        JSON com status fire_and_forget + queue_id, ou error
    """
    if not settings.enable_rag:
        logger.warning("embedding tool called but RAG disabled")
        return json.dumps(
            {
                "status": "error",
                "error": "RAG is disabled. Enable ENABLE_RAG=true to use this tool.",
            }
        )

    if not settings.embedding_queue_enabled:
        logger.error("embedding called but queue not enabled")
        return json.dumps(
            {"status": "error", "error": "Embedding queue not configured."}
        )

    try:
        queue = await get_embedding_queue(settings.embedding_queue_dsn)
        queue_id = await queue.enqueue(text, collection, metadata)

        logger.info(
            "embedding_enqueued",
            extra={
                "queue_id": queue_id,
                "collection": collection,
                "text_length": len(text),
            },
        )

        return json.dumps(
            {
                "status": "fire_and_forget",
                "queue_id": queue_id,
                "collection": collection,
                "message": "Document enqueued for async embedding and indexing.",
            }
        )

    except Exception as e:
        logger.exception(
            "embedding_enqueue_failed",
            extra={"collection": collection, "text_length": len(text)},
        )
        return json.dumps(
            {
                "status": "error",
                "error": str(e) or "Failed to enqueue embedding",
                "collection": collection,
            }
        )


@tool
async def vector_search(
    query: str, collection: str = "articles", limit: int = 5
) -> str:  # noqa: PLR0911
    """Busca o banco de dados vetorial LanceDB por documentos similares.

    LanceDB é file-based — nenhum container ou servidor é necessário.

    Args:
        query: String da consulta de busca
        collection: Nome da tabela LanceDB
        limit: Número máximo de resultados a retornar

    Returns:
        JSON com documentos e scores de similaridade
    """
    if not settings.enable_rag:
        logger.warning("vector_search tool called but RAG disabled")
        return "RAG is disabled. Enable ENABLE_RAG=true to use this tool."

    try:
        if lancedb is None or VoyageAIEmbeddings is None:
            return "LanceDB or Voyage AI dependencies missing."

        if not settings.voyage_api_key:
            logger.error("vector_search called but VOYAGE_API_KEY not configured")
            return json.dumps(
                {"status": "failed", "error": "VOYAGE_API_KEY not configured"}
            )

        embeddings_model = VoyageAIEmbeddings(
            api_key=SecretStr(settings.voyage_api_key),
            model=settings.embedding_model,
        )

        query_vector = embeddings_model.embed_query(query)

        db = await lancedb.connect_async(str(settings.lancedb_dir))

        try:
            async with asyncio.timeout(10):
                table = await db.open_table(collection)
        except TimeoutError:
            logger.exception(f"LanceDB open_table timed out for {collection}")
            return json.dumps(
                {"status": "error", "error": "Vector store access timed out"}
            )
        except Exception:
            logger.warning("LanceDB table not found", extra={"collection": collection})
            return json.dumps(
                {
                    "status": "no_results",
                    "message": f"Collection '{collection}' not found or empty",
                }
            )

        try:
            async with asyncio.timeout(10):
                search_results = await (
                    table.vector_search(query_vector).limit(limit).to_pandas()
                )
        except TimeoutError:
            logger.exception(f"vector_search timed out for collection {collection}")
            return json.dumps(
                {"status": "error", "error": "Search timed out after 10s"}
            )

        results = [
            {
                "id": str(row["id"]),
                "score": float(row.get("_distance", 0.0)),
                "content": row["text"],
                "metadata": json.loads(row.get("metadata", "{}")),
            }
            for _, row in search_results.iterrows()
        ]

        # Reranking opcional — melhora precisão
        if results and settings.reranker_type == "voyage" and VoyageAIRerank:
            try:
                reranker = VoyageAIRerank(
                    api_key=SecretStr(settings.voyage_api_key),
                    model=settings.reranker_model,
                    top_k=settings.reranker_top_k,
                )
                docs_to_rerank = [
                    LCDoc(page_content=str(r["content"]), metadata=r["metadata"])
                    for r in results
                ]
                reranked_docs = reranker.compress_documents(docs_to_rerank, query)
                results = [
                    {
                        "content": doc.page_content,
                        "metadata": doc.metadata,
                        "relevance_score": getattr(doc, "relevance_score", None),
                    }
                    for doc in reranked_docs
                ]
                logger.info("vector_search_reranked", extra={"top_k": len(results)})
            except Exception as rerank_err:
                logger.warning(f"Reranking failed, using raw results: {rerank_err}")
                for r in results:
                    r["reranking_status"] = "unavailable"

        logger.info(
            "vector_search completed",
            extra={
                "query": query,
                "collection": collection,
                "result_count": len(results),
            },
        )

        return json.dumps(
            {
                "status": "success",
                "results": results,
                "query": query,
                "collection": collection,
            }
        )

    except Exception as e:
        logger.exception(
            "vector_search_failed", extra={"query": query, "collection": collection}
        )
        return json.dumps(
            {"status": "failed", "error": str(e) or "Vector search failed"}
        )


@tool
async def ingest_docs(
    directory_path: str,
    collection: str = "articles",
    glob_pattern: str = "**/*.md",
) -> str:
    """Lê todos os arquivos de um diretório e os indexa no banco vetorial.

    Útil para popular o conhecimento do agente com documentação local.

    Args:
        directory_path: Caminho da pasta contendo os documentos
        collection: Nome da coleção LanceDB de destino
        glob_pattern: Padrão de busca de arquivos (ex: **/*.md, **/*.txt)

    Returns:
        Status da ingestão com contagem de sucessos/falhas
    """
    from pathlib import Path

    from tool_safety import is_safe_file_path

    if not settings.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_file_path(directory_path):
        return f"Access denied: {directory_path} is outside allowed directory"

    path = Path(directory_path)
    if not path.is_dir():
        return f"Not a directory: {directory_path}"

    loader = DirectoryLoader(
        str(path),
        glob=glob_pattern,
        loader_cls=TextLoader,
        loader_kwargs={"encoding": "utf-8"},
    )

    try:
        docs = loader.load()
    except Exception as e:
        logger.exception(f"Error loading documents from {directory_path}: {e}")
        return f"Error loading documents: {e!s}"

    if not docs:
        return f"No documents found in {directory_path} matching {glob_pattern}"

    splitter = RecursiveCharacterTextSplitter(chunk_size=1000, chunk_overlap=100)
    chunks = splitter.split_documents(docs)

    success_count = 0
    fail_count = 0

    for chunk in chunks:
        chunk_metadata = chunk.metadata or {}
        chunk_metadata.update(
            {"ingested_at": datetime.now(UTC).isoformat(), "source_dir": directory_path}
        )

        res = ""
        async for stream_chunk in embedding.astream(
            {
                "text": chunk.page_content,
                "collection": collection,
                "metadata": chunk_metadata,
            }
        ):
            if isinstance(stream_chunk, str):
                res += stream_chunk

        if '"status": "fire_and_forget"' in res:
            success_count += 1
        else:
            fail_count += 1

    logger.info(
        "ingest_docs_completed",
        extra={
            "collection": collection,
            "success": success_count,
            "fail": fail_count,
            "total_chunks": len(chunks),
        },
    )

    return json.dumps(
        {
            "status": "completed",
            "total_files": len(docs),
            "total_chunks": len(chunks),
            "indexed": success_count,
            "failed": fail_count,
            "collection": collection,
        }
    )
