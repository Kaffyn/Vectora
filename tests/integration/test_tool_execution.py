#!/usr/bin/env python3
"""Diagnostic test for tool message loss bug.

This script tests the tool execution pipeline to identify where ToolMessages
are being lost in the LangGraph execution flow.
"""

import asyncio
import logging
import sys
from pathlib import Path

# Add vectora to path
sys.path.insert(0, str(Path(__file__).parent / "vectora"))

# Set up debug logging BEFORE importing anything else
import os

from state import State  # noqa: TC002

os.environ["LOG_LEVEL"] = "DEBUG"
os.environ["QUIET_MODE"] = "false"  # Enable all logs

from log_setup import setup_logging

setup_logging(json_output=False, log_level="DEBUG")

logger = logging.getLogger(__name__)


async def test_tool_execution():
    """Test tool execution pipeline with detailed logging."""
    logger.info("=" * 80)
    logger.info("STARTING TOOL EXECUTION DIAGNOSTIC TEST")
    logger.info("=" * 80)

    # Import after logging is set up
    from initialization import ensure_vectora_initialized

    ensure_vectora_initialized()

    # Load environment
    from dotenv import load_dotenv

    load_dotenv()
    load_dotenv(Path.home() / ".vectora" / ".env")

    # Import graph components
    from graph import build_graph
    from langchain_core.messages import HumanMessage
    from langgraph.checkpoint.memory import MemorySaver
    from settings import settings

    # Ensure LLM is configured
    if not settings.get_llm_provider():
        logger.error(
            "LLM provider not configured. Please set VECTORA_LLM_PROVIDER and required API keys."
        )
        return None

    logger.info(f"Using LLM provider: {settings.get_llm_provider()}")
    logger.info(f"File operations enabled: {settings.enable_file_operations}")

    try:
        # Use in-memory checkpointer for testing
        checkpointer = MemorySaver()

        # Build the graph
        graph = build_graph(checkpointer)
        logger.info("Graph built successfully")

        # Create initial state with a message that should trigger tool use
        initial_state: State = {
            "messages": [
                HumanMessage(
                    content="Please list the files in the current directory (use the list_dir tool)"
                )
            ],
            "session_metadata": {
                "thread_id": 1,
                "user_type": "test_user",
                "created_at": "2026-05-16T00:00:00",
                "llm_provider": settings.get_llm_provider(),
                "llm_model": settings.get_llm_model(),
            },
        }

        logger.info(
            "[TEST] Initial state messages count: %d", len(initial_state["messages"])
        )
        logger.info("[TEST] Initial message: %s", initial_state["messages"][0].content)

        # Run the graph
        logger.info("[TEST] Running graph with initial state...")
        result = await graph.ainvoke(
            initial_state,
            config={"configurable": {"thread_id": 1}},
        )

        # Analyze results
        logger.info("[TEST] Graph execution completed")
        logger.info("[TEST] Final messages count: %d", len(result["messages"]))

        print("\n" + "=" * 80)
        print("MESSAGE TRACE:")
        print("=" * 80)

        for i, msg in enumerate(result["messages"]):
            msg_type = type(msg).__name__
            msg_content = str(msg.content)[:100] if msg.content else "(empty)"
            print(f"\n[{i}] {msg_type}")
            print(f"    Content: {msg_content}")
            if hasattr(msg, "tool_calls"):
                print(f"    Tool calls: {len(msg.tool_calls) if msg.tool_calls else 0}")
            if hasattr(msg, "name"):
                print(f"    Tool name: {msg.name}")
            if hasattr(msg, "tool_use_id"):
                print(f"    Tool use ID: {msg.tool_use_id}")

        print("\n" + "=" * 80)
        print("SUMMARY:")
        print("=" * 80)

        # Count message types
        msg_types = {}
        for msg in result["messages"]:
            msg_type = type(msg).__name__
            msg_types[msg_type] = msg_types.get(msg_type, 0) + 1

        for msg_type, count in msg_types.items():
            print(f"{msg_type}: {count}")

        # Check for tool messages
        has_tool_messages = any(
            type(msg).__name__ == "ToolMessage" for msg in result["messages"]
        )
        print(f"\nToolMessages present: {has_tool_messages}")

        if not has_tool_messages:
            print("[ERROR] No ToolMessages found! Tool results are not being captured.")
        else:
            print("[OK] ToolMessages found in state.")

        return result

    finally:
        logger.info("Test completed")


if __name__ == "__main__":
    asyncio.run(test_tool_execution())
