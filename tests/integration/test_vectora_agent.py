"""Teste reduzido do Vectora Research Agent - 2 tópicos para demonstração."""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING
from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from langchain_core.messages import HumanMessage

from vectora.context import Context
from vectora.graph import build_graph
from vectora.services.checkpoint import Checkpointer

if TYPE_CHECKING:
    from vectora.state import State

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)


@pytest.mark.asyncio
async def test_vectora_agent_builds_and_runs() -> None:
    """Teste que o grafo Vectora pode ser construído e executado."""
    from langgraph.checkpoint.memory import MemorySaver

    checkpointer = MemorySaver()

    with patch("vectora.services.utils.init_chat_model") as mock_init:
        mock_llm = MagicMock()
        mock_llm.bind_tools = MagicMock(return_value=mock_llm)
        mock_llm.with_config = MagicMock(return_value=mock_llm)
        mock_llm.ainvoke = AsyncMock(
            return_value=MagicMock(
                content="I'm Vectora, ready to help.",
                tool_calls=[],
            )
        )
        mock_init.return_value = mock_llm

        graph = build_graph(checkpointer)
        assert graph is not None

        state: State = {
            "messages": [HumanMessage(content="Who are you?")],
            "session_metadata": {
                "thread_id": 1,
                "user_type": "user",
                "created_at": "2026-05-17T00:00:00Z",
                "llm_provider": "google-genai",
                "llm_model": "gemini-2.0-flash",
            },
            "retrieval_results": None,
            "selected_rag_source": None,
            "routing_decision": None,
            "summarized_history": None,
        }

        config = {"configurable": {"thread_id": "test-agent-1"}}
        result = await graph.ainvoke(state, config=config)

        assert result is not None
        assert "messages" in result
        assert len(result["messages"]) >= 1
        logger.info("Agent test passed with %d messages", len(result["messages"]))


@pytest.mark.asyncio
async def test_context_creation() -> None:
    """Teste que Context pode ser criado corretamente."""
    context = Context(
        thread_id="test-research-agent",
        user_id="test",
        user_type="researcher",
    )

    assert context.thread_id == "test-research-agent"
    assert context.user_id == "test"
    assert context.user_type == "researcher"


@pytest.mark.asyncio
async def test_checkpointer_as_context_manager() -> None:
    """Teste que Checkpointer funciona como async context manager."""
    import tempfile
    from pathlib import Path

    with tempfile.TemporaryDirectory() as tmpdir:
        db_path = str(Path(tmpdir) / "test_agent.db")
        async with Checkpointer(db_path) as checkpointer:
            assert checkpointer is not None
