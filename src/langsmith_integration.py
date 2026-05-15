"""LangSmith Integration for Observability and Tracing.

Configures LangSmith tracing for all LangGraph executions:
- Automatic trace creation for each graph invocation
- Token counting for input/output
- Latency tracking for LLM calls and tool execution
- Error tracking and debugging information
- Cache hit metrics for vector searches
"""

import logging
import os
from typing import Any

from langsmith import Client, tracing_context

logger = logging.getLogger(__name__)


class LangSmithObserver:
    """Manager for LangSmith observability integration.

    Enables/disables tracing based on LANGSMITH_API_KEY environment variable.
    """

    def __init__(self) -> None:
        """Initialize LangSmith observer.

        Raises:
            RuntimeError: If LANGSMITH_API_KEY is set but client initialization fails
                (fail-loud: observability is infrastructure, not optional)
        """
        self.api_key = os.getenv("LANGSMITH_API_KEY")
        self.enabled = bool(self.api_key)
        self.client: Client | None = None

        if self.enabled:
            try:
                self.client = Client(api_key=self.api_key)
                logger.info(
                    "LangSmith tracing enabled",
                    extra={
                        "project": os.getenv("LANGSMITH_PROJECT", "vectora"),
                    },
                )
            except Exception as e:
                msg = (
                    f"LANGSMITH_API_KEY is set but client initialization failed: {e!s}"
                )
                logger.critical(msg)
                # Fail-loud: observability is infrastructure, not optional
                raise RuntimeError(msg) from e
        else:
            logger.info("LangSmith tracing disabled (set LANGSMITH_API_KEY to enable)")

    def get_project_name(self) -> str:
        """Get LangSmith project name from environment."""
        return os.getenv("LANGSMITH_PROJECT", "vectora")

    def get_trace_tags(self, context: dict[str, Any] | None = None) -> list[str]:
        """Get tags for tracing.

        Args:
            context: Optional context with user/thread info

        Returns:
            List of tags for the trace
        """
        tags = ["vectora", "v0.0.1dev1"]

        if context:
            if user_type := context.get("user_type"):
                tags.append(f"user:{user_type}")
            if thread_id := context.get("thread_id"):
                tags.append(f"thread:{thread_id}")

        return tags

    def get_trace_metadata(
        self, context: dict[str, Any] | None = None
    ) -> dict[str, Any]:
        """Get metadata for tracing.

        Args:
            context: Optional context with user/thread info

        Returns:
            Dictionary of metadata for the trace
        """
        metadata = {
            "version": "0.0.1dev1",
            "source": "vectora-chat",
        }

        if context:
            if user_id := context.get("user_id"):
                metadata["user_id"] = user_id
            if user_type := context.get("user_type"):
                metadata["user_type"] = user_type
            if thread_id := context.get("thread_id"):
                metadata["thread_id"] = thread_id
            if correlation_id := context.get("correlation_id"):
                metadata["correlation_id"] = correlation_id

        return metadata


# Global singleton
_observer: LangSmithObserver | None = None


def get_langsmith_observer() -> LangSmithObserver:
    """Get or create the LangSmith observer singleton.

    Returns:
        LangSmithObserver instance
    """
    global _observer
    if _observer is None:
        _observer = LangSmithObserver()
    return _observer


def create_trace_context(
    run_name: str,
    context: dict[str, Any] | None = None,
) -> Any:
    """Create a LangSmith trace context for a graph invocation.

    Args:
        run_name: Name of the run (e.g., "chat_message")
        context: Optional context with user/thread/correlation info

    Returns:
        Context manager for tracing (if enabled) or no-op context
    """
    observer = get_langsmith_observer()

    if not observer.enabled:
        # Return a no-op context manager
        from contextlib import nullcontext

        return nullcontext()

    # Create LangSmith context
    return tracing_context(
        run_type="chain",
        name=run_name,
        tags=observer.get_trace_tags(context),
        metadata=observer.get_trace_metadata(context),
        project_name=observer.get_project_name(),
    )


async def trace_graph_invocation(
    graph_name: str,
    input_data: dict[str, Any],
    context: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """Decorator-compatible function to trace graph invocations.

    Args:
        graph_name: Name of the graph (e.g., "vectora_chat_graph")
        input_data: Input to the graph
        context: Optional context with user/thread/correlation info

    Returns:
        Metadata about the trace
    """
    observer = get_langsmith_observer()

    if not observer.enabled:
        return {"tracing_enabled": False}

    return {
        "tracing_enabled": True,
        "project": observer.get_project_name(),
        "tags": observer.get_trace_tags(context),
        "metadata": observer.get_trace_metadata(context),
    }


def log_tool_execution(
    tool_name: str,
    input_data: dict[str, Any] | None = None,
    output_data: dict[str, Any] | None = None,
    duration_ms: float | None = None,
    error: str | None = None,
) -> None:
    """Log tool execution details for tracing.

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
    """Log LLM API call details for tracing.

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
    """Log vector search execution details.

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
