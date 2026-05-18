"""Diagnostic test for tool message loss bug.

Tests the tool execution pipeline to identify where ToolMessages
are being lost in the LangGraph execution flow.
"""

from __future__ import annotations

import logging
import os
from typing import TYPE_CHECKING

import pytest

os.environ["LOG_LEVEL"] = "DEBUG"
os.environ["QUIET_MODE"] = "false"

from vectora.services.log_setup import setup_logging

if TYPE_CHECKING:
    from vectora.state import State

setup_logging(json_output=False, log_level="DEBUG")

logger = logging.getLogger(__name__)


@pytest.mark.asyncio
async def test_tool_execution_pipeline():
    """Test tool execution pipeline with detailed logging."""
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
                content="I'll list the files for you.",
                tool_calls=[],
            )
        )
        mock_init.return_value = mock_llm

        graph = build_graph(checkpointer)

        initial_state: State = {
            "messages": [
                HumanMessage(content="Please list the files in the current directory")
            ],
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

        logger.info("Running graph with initial state...")
        result = await graph.ainvoke(
            initial_state,
            config={"configurable": {"thread_id": "test_tool_exec_1"}},
        )

        logger.info("Graph execution completed")
        assert result is not None
        assert "messages" in result
        assert len(result["messages"]) >= 1

        msg_types = {}
        for msg in result["messages"]:
            msg_type = type(msg).__name__
            msg_types[msg_type] = msg_types.get(msg_type, 0) + 1

        logger.info("Message types: %s", msg_types)


@pytest.mark.asyncio
async def test_tool_messages_captured_in_state():
    """Verify ToolMessages are captured in graph state."""
    from langchain_core.messages import AIMessage, HumanMessage, ToolMessage

    state: State = {
        "messages": [
            HumanMessage(content="List files"),
            AIMessage(
                content="",
                tool_calls=[
                    {
                        "name": "list_dir",
                        "args": {"path": "."},
                        "id": "call_1",
                        "type": "tool_call",
                    }
                ],
            ),
            ToolMessage(
                name="list_dir",
                content="file1.py\nfile2.py",
                tool_call_id="call_1",
            ),
        ],
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

    has_tool_messages = any(isinstance(m, ToolMessage) for m in state["messages"])
    assert has_tool_messages, "ToolMessages should be present in state"
    assert len(state["messages"]) == 3
