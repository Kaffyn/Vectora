"""LangSmith Observability Configuration.

Configures tracing via environment variables (SDK auto-injection).
All structured logging for tools and LLM calls is handled here.

Environment variables (auto-detected by LangChain SDK):
    LANGSMITH_TRACING=true
    LANGSMITH_ENDPOINT=https://api.smith.langchain.com
    LANGSMITH_API_KEY=<your-api-key>
    LANGSMITH_PROJECT=vectora
"""

import logging
from typing import Any

logger = logging.getLogger(__name__)


def log_tool_execution(
    tool_name: str,
    input_data: dict[str, Any] | None = None,
    output_data: dict[str, Any] | None = None,
    duration_ms: float | None = None,
    error: str | None = None,
) -> None:
    """Log tool execution details for structured observability.

    Args:
        tool_name: Name of the tool executed
        input_data: Tool input (optional)
        output_data: Tool output (optional)
        duration_ms: Execution duration in milliseconds
        error: Error message if tool failed (optional)
    """
    if error:
        logger.warning(
            "tool_execution_failed",
            extra={
                "tool": tool_name,
                "duration_ms": duration_ms,
                "error": error,
            },
        )
    else:
        logger.info(
            "tool_execution_success",
            extra={
                "tool": tool_name,
                "duration_ms": duration_ms,
            },
        )


def log_llm_call(
    model: str,
    input_tokens: int | None = None,
    output_tokens: int | None = None,
    total_tokens: int | None = None,
    duration_ms: float | None = None,
) -> None:
    """Log LLM API call details for structured observability.

    Args:
        model: LLM model name
        input_tokens: Number of input tokens
        output_tokens: Number of output tokens
        total_tokens: Total tokens used
        duration_ms: Call duration in milliseconds
    """
    logger.info(
        "llm_api_call",
        extra={
            "model": model,
            "input_tokens": input_tokens,
            "output_tokens": output_tokens,
            "total_tokens": total_tokens,
            "duration_ms": duration_ms,
        },
    )


def log_vector_search(
    collection: str,
    query: str | None = None,
    num_results: int | None = None,
    *,
    min_score: float | None = None,
    duration_ms: float | None = None,
    cache_hit: bool = False,
) -> None:
    """Log vector search execution details for structured observability.

    Args:
        collection: Vector collection/index name
        query: Search query (optional)
        num_results: Number of results returned
        min_score: Minimum relevance score of results (keyword-only)
        duration_ms: Search duration in milliseconds (keyword-only)
        cache_hit: Whether result came from cache (keyword-only)
    """
    logger.info(
        "vector_search_execution",
        extra={
            "collection": collection,
            "num_results": num_results,
            "min_score": min_score,
            "duration_ms": duration_ms,
            "cache_hit": cache_hit,
        },
    )
