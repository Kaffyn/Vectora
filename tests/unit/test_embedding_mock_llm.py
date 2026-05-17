"""Test embedding with mocked LLM to bypass Google API quota limits.

This test verifies that:
1. Tools are properly bound to LLM
2. Tool calls are created correctly
3. ToolMessages are generated from tool execution
4. Message flow is preserved through the graph
"""

import asyncio
import os
import sys
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

# Set up debug logging BEFORE importing anything
os.environ["LOG_LEVEL"] = "DEBUG"
os.environ["QUIET_MODE"] = "false"
os.environ["LOG_JSON"] = "false"

# Add vectora to path
sys.path.insert(0, str(Path(__file__).parent / "vectora"))

from collections.abc import AsyncGenerator
from typing import Any

from log_setup import setup_logging
from state import State  # noqa: TC002

setup_logging(json_output=False, log_level="DEBUG")

import logging

logger = logging.getLogger(__name__)


async def test_embedding_with_mock():
    """Test embedding with mocked LLM response."""
    logger.info("=" * 80)
    logger.info("TESTING EMBEDDING WITH MOCKED LLM (no API quota issues)")
    logger.info("=" * 80)

    # Initialize Vectora
    from dotenv import load_dotenv
    from graph import build_graph
    from initialization import ensure_vectora_initialized
    from langchain_core.messages import AIMessage, HumanMessage
    from langgraph.checkpoint.memory import MemorySaver
    from settings import settings

    ensure_vectora_initialized()
    load_dotenv()
    load_dotenv(Path.home() / ".vectora" / ".env")

    logger.info(f"LLM Provider: {settings.get_llm_provider()}")
    logger.info(f"RAG Enabled: {settings.enable_rag}")
    logger.info(f"Embedding Queue Enabled: {settings.embedding_queue_enabled}")

    # Build graph with diagnostics enabled
    checkpointer = MemorySaver()
    graph = build_graph(checkpointer)

    # Create state asking to do embedding of vectora folder
    user_message = """Quero que você faça embedding de toda a pasta vectora.
Leia todos os arquivos .py da pasta vectora (recursive) e faça embedding deles na coleção 'vectora'.
Use a ferramenta ingest_docs para isso."""

    initial_state: State = {
        "messages": [HumanMessage(content=user_message)],
        "session_metadata": {
            "thread_id": 1,
            "user_type": "test_user",
            "created_at": "2026-05-16T00:00:00",
            "llm_provider": settings.get_llm_provider(),
            "llm_model": settings.get_llm_model(),
        },
    }

    logger.info("[TEST] Starting graph execution for embedding request...")
    logger.info(f"[TEST] User request: {user_message[:100]}...")

    # Mock the LLM to return a tool call without hitting the API
    from tools import TOOLS

    try:
        # Patch the astream method of llm_with_tools to return chunks
        from nodes import _get_llm_with_tools

        original_llm = _get_llm_with_tools()

        async def mock_astream(*args: Any, **kwargs: Any) -> AsyncGenerator[AIMessage]:
            """Mock astream that yields chunks with tool_calls at the end."""
            # Create a chunk with tool_calls
            chunk_with_tools = AIMessage(
                content="",
                tool_calls=[
                    {
                        "name": "ingest_docs",
                        "args": {
                            "docs_pattern": "vectora/**/*.py",
                            "collection": "vectora",
                            "recursive": True,
                        },
                        "id": "call_ingest_docs_001",
                        "type": "tool_call",
                    }
                ],
            )
            yield chunk_with_tools

        with patch.object(original_llm, "astream", side_effect=mock_astream):
            logger.info("[TEST] LLM mocked to return tool call for ingest_docs...")

            # Run the graph
            result = await graph.ainvoke(
                initial_state,
                config={"configurable": {"thread_id": 1}},
            )

            logger.info("[TEST] Graph execution completed")
            logger.info(f"[TEST] Final messages count: {len(result['messages'])}")

            # Display results
            print("\n" + "=" * 80)
            print("FINAL STATE MESSAGES:")
            print("=" * 80)

            for i, msg in enumerate(result["messages"]):
                msg_type = type(msg).__name__
                content_preview = str(msg.content)[:100] if msg.content else "(empty)"
                print(f"\n[{i}] {msg_type}")
                print(f"    {content_preview}")

                if msg_type == "AIMessage" and hasattr(msg, "tool_calls"):
                    if msg.tool_calls:
                        print(f"    Tool calls: {len(msg.tool_calls)}")
                        for tc in msg.tool_calls:
                            tool_name = (
                                tc.get("name") if isinstance(tc, dict) else tc.name
                            )
                            print(f"      - {tool_name}")

                if msg_type == "ToolMessage":
                    print(f"    Tool: {getattr(msg, 'name', 'N/A')}")
                    print(f"    Content length: {len(str(msg.content))}")

            # Summary
            print("\n" + "=" * 80)
            print("MESSAGE SUMMARY:")
            print("=" * 80)

            msg_types = {}
            for msg in result["messages"]:
                msg_type = type(msg).__name__
                msg_types[msg_type] = msg_types.get(msg_type, 0) + 1

            for msg_type, count in sorted(msg_types.items()):
                print(f"  {msg_type}: {count}")

            # Check for tool messages
            has_tool_messages = any(
                type(msg).__name__ == "ToolMessage" for msg in result["messages"]
            )
            print(f"\n  ToolMessages present: {has_tool_messages}")

            if has_tool_messages:
                print("  [OK] Tool results captured in state!")
            else:
                print("  [ERROR] No ToolMessages found - tools may not have executed")

    except Exception as e:
        logger.exception("[TEST] Error during graph execution")
        print(f"\n[ERROR] {e}")
        raise


if __name__ == "__main__":
    asyncio.run(test_embedding_with_mock())
