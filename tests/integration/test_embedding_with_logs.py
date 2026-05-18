"""Test embedding with full DEBUG logging to diagnose tool message loss."""

from __future__ import annotations

import logging
import os
from typing import TYPE_CHECKING

import pytest

os.environ["LOG_LEVEL"] = "DEBUG"
os.environ["QUIET_MODE"] = "false"
os.environ["LOG_JSON"] = "false"

from vectora.services.log_setup import setup_logging

if TYPE_CHECKING:
    from vectora.state import State

setup_logging(json_output=False, log_level="DEBUG")

logger = logging.getLogger(__name__)


@pytest.mark.asyncio
async def test_embedding_with_debug_logging():
    """Test embedding of vectora folder with debugging."""
    from pathlib import Path
    from unittest.mock import AsyncMock, MagicMock, patch

    from langchain_core.messages import HumanMessage
    from langgraph.checkpoint.memory import MemorySaver

    from vectora.graph import build_graph

    checkpointer = MemorySaver()

    with patch("vectora.services.utils.init_chat_model") as mock_init:
        mock_llm = MagicMock()
        mock_llm.bind_tools = MagicMock(return_value=mock_llm)
        mock_llm.with_config = MagicMock(return_value=mock_llm)
        mock_llm.ainvoke = AsyncMock(
            return_value=MagicMock(
                content="I'll embed those files for you.",
                tool_calls=[],
            )
        )
        mock_init.return_value = mock_llm

        graph = build_graph(checkpointer)

        user_message = "Quero que você faça embedding de toda a pasta vectora."

        initial_state: State = {
            "messages": [HumanMessage(content=user_message)],
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

        result = await graph.ainvoke(
            initial_state,
            config={"configurable": {"thread_id": "test_embed_1"}},
        )

        assert result is not None
        assert "messages" in result
        assert len(result["messages"]) >= 1

        logger.info(
            "Graph execution completed with %d messages", len(result["messages"])
        )

        msg_types = {}
        for msg in result["messages"]:
            msg_type = type(msg).__name__
            msg_types[msg_type] = msg_types.get(msg_type, 0) + 1

        logger.info("Message types: %s", msg_types)
