import pytest
from langgraph.graph.state import RunnableConfig

from testing import (
    assert_last_message_is_ai,
    assert_tool_called_with_args,
    assert_tool_result_in_messages,
    human_message,
)


@pytest.mark.asyncio()
class TestGraphFlow:
    """Test suite for the graph execution flow."""

    async def test_graph_single_tool_call(self, test_graph, checkpointer) -> None:
        """Test that graph correctly calls the multiply tool.

        Scenario:
        1. User sends message: "multiply 5 by 3"
        2. Mock LLM creates tool call for multiply(a=5, b=3)
        3. Graph executes the tool
        4. Assert tool was called and result is in final state
        """
        config = RunnableConfig(configurable={"thread_id": "test_1"})

        result = await test_graph.ainvoke(
            {"messages": [human_message("multiply 5 by 3")]},
            config=config,
        )

        messages = result["messages"]

        assert_tool_called_with_args(
            messages,
            tool_name="multiply",
            expected_args={"a": 5.0, "b": 3.0},
        )

        assert_tool_result_in_messages(
            messages,
            tool_name="multiply",
            expected_result=15.0,
        )

    async def test_graph_multiple_tool_calls_in_sequence(self, test_graph) -> None:
        """Test that graph can handle multiple tool calls in sequence."""
        config = RunnableConfig(configurable={"thread_id": "test_2"})

        initial_result = await test_graph.ainvoke(
            {"messages": [human_message("multiply 2 by 3")]},
            config=config,
        )

        assert_tool_called_with_args(
            initial_result["messages"],
            tool_name="multiply",
            expected_args={"a": 2.0, "b": 3.0},
        )

        follow_up_result = await test_graph.ainvoke(
            {"messages": [human_message("Now multiply 4 by 5")]},
            config=config,
        )

        assert_tool_called_with_args(
            follow_up_result["messages"],
            tool_name="multiply",
            expected_args={"a": 4.0, "b": 5.0},
        )

    async def test_graph_preserves_message_history(self, test_graph) -> None:
        """Test that graph preserves message history across invocations."""
        config = RunnableConfig(configurable={"thread_id": "test_3"})

        first_turn = await test_graph.ainvoke(
            {"messages": [human_message("multiply 3 by 4")]},
            config=config,
        )

        first_turn_message_count = len(first_turn["messages"])

        second_turn = await test_graph.ainvoke(
            {"messages": [human_message("What's 2 times 8?")]},
            config=config,
        )

        assert len(second_turn["messages"]) > first_turn_message_count

    async def test_graph_ends_without_tool_call(self, test_graph) -> None:
        """Test that graph can end conversation without tool calls."""
        config = RunnableConfig(configurable={"thread_id": "test_4"})

        result = await test_graph.ainvoke(
            {"messages": [human_message("Hello")]},
            config=config,
        )

        assert_last_message_is_ai(result["messages"])

    async def test_graph_state_persistence(self, test_graph) -> None:
        """Test that graph state persists with same thread_id."""
        config = RunnableConfig(configurable={"thread_id": "test_persistent"})

        await test_graph.ainvoke(
            {"messages": [human_message("multiply 10 by 5")]},
            config=config,
        )

        stored_state = await test_graph.aget_state(config=config)

        assert stored_state.values is not None
        assert len(stored_state.values["messages"]) > 0
