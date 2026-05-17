"""Direct test of ToolNode behavior without LLM calls."""

import asyncio
import sys
from pathlib import Path

# Add vectora to path
sys.path.insert(0, str(Path(__file__).parent / "vectora"))

# Set up debug logging
import os

from state import State  # noqa: TC002

os.environ["LOG_LEVEL"] = "DEBUG"
os.environ["QUIET_MODE"] = "false"

from log_setup import setup_logging

setup_logging(json_output=False, log_level="DEBUG")

import logging

logger = logging.getLogger(__name__)


async def test_toolnode():
    """Test ToolNode directly."""
    logger.info("=" * 80)
    logger.info("TESTING TOOLNODE DIRECTLY")
    logger.info("=" * 80)

    from langchain_core.messages import AIMessage, HumanMessage
    from langgraph.prebuilt.tool_node import ToolNode
    from tools import TOOLS

    # Create a tool node
    tool_node = ToolNode(tools=TOOLS)
    logger.info(f"Created ToolNode with {len(TOOLS)} tools")

    # Create a state with an AIMessage containing tool_calls
    # Simulate what the LLM would return when asking to list files
    ai_message = AIMessage(
        content="I'll list the files in the current directory for you.",
        tool_calls=[
            {
                "name": "list_dir",
                "args": {"path": "."},
                "id": "call_1",
                "type": "tool_call",
            }
        ],
    )

    state: State = {
        "messages": [
            HumanMessage(content="List the files in the current directory"),
            ai_message,
        ],
        "session_metadata": {
            "thread_id": 1,
            "user_type": "test_user",
            "created_at": "2026-05-16T00:00:00",
            "llm_provider": "google-genai",
            "llm_model": "gemini-3.1-flash-lite",
        },
    }

    logger.info("[TEST] Initial state:")
    logger.info(f"  - Messages: {len(state['messages'])}")
    logger.info(f"  - Last message type: {type(state['messages'][-1]).__name__}")
    logger.info(f"  - Tool calls: {len(ai_message.tool_calls)}")

    # Run the tool node with required config
    logger.info("[TEST] Invoking ToolNode...")
    try:
        from langgraph.graph.state import RunnableConfig
        from langgraph.types import StreamMode

        # ToolNode requires a config with a runtime
        config = RunnableConfig(configurable={"thread_id": "test_thread_1"})

        result = await tool_node.ainvoke(state, config=config)

        logger.info("[TEST] ToolNode execution completed")
        logger.info(f"[TEST] Result type: {type(result).__name__}")
        logger.info(
            f"[TEST] Result keys: {list(result.keys()) if isinstance(result, dict) else 'N/A'}"
        )

        if isinstance(result, dict) and "messages" in result:
            logger.info(f"[TEST] Messages in result: {len(result['messages'])}")
            for i, msg in enumerate(result["messages"]):
                logger.info(f"  [{i}] {type(msg).__name__}: {str(msg.content)[:50]}")

        # Display the result
        print("\n" + "=" * 80)
        print("TOOL NODE RESULT:")
        print("=" * 80)
        print(f"Result type: {type(result).__name__}")
        print(f"Result: {result}")

        if isinstance(result, dict) and "messages" in result:
            print(f"\nMessages in result: {len(result['messages'])}")
            for i, msg in enumerate(result["messages"]):
                print(f"\n[{i}] {type(msg).__name__}")
                if hasattr(msg, "content"):
                    print(f"    Content: {str(msg.content)[:100]}")
                if hasattr(msg, "name"):
                    print(f"    Tool: {msg.name}")
                if hasattr(msg, "tool_use_id"):
                    print(f"    Tool ID: {msg.tool_use_id}")

        return result

    except Exception as e:
        logger.exception("[TEST] Error invoking ToolNode")
        print(f"\nERROR: {e}")
        raise


if __name__ == "__main__":
    asyncio.run(test_toolnode())
