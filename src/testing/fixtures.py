import tempfile
from collections.abc import AsyncGenerator
from pathlib import Path

import pytest
from langchain_core.messages import SystemMessage
from langgraph.checkpoint.sqlite.aio import AsyncSqliteSaver
from langgraph.graph.state import CompiledStateGraph

from context import Context
from prompts import get_system_prompt
from state import State
from testing.mocks import MockLLM


@pytest.fixture()
def test_context() -> Context:
    """Provide a test context with user_type='plus'."""
    return Context(user_type="plus")


@pytest.fixture()
def mock_llm() -> MockLLM:
    """Provide a mock LLM for deterministic testing."""
    return MockLLM()


@pytest.fixture()
async def temp_db() -> AsyncGenerator[str, None]:
    """Create a temporary SQLite database for testing.

    Yields the DSN path, cleans up after test.
    """
    temp_dir = tempfile.mkdtemp()
    db_path = Path(temp_dir) / "test.db"

    yield f"sqlite:///{db_path}"

    if db_path.exists():
        db_path.unlink()


@pytest.fixture()
async def checkpointer(temp_db: str) -> AsyncGenerator[AsyncSqliteSaver, None]:
    """Provide an AsyncSqliteSaver with temporary database.

    The checkpointer is used for persisting graph state.
    """
    async with AsyncSqliteSaver(conn_string=temp_db) as saver:
        yield saver


@pytest.fixture()
async def test_graph(
    mock_llm: MockLLM, checkpointer: AsyncSqliteSaver
) -> CompiledStateGraph[State, Context, State, State]:
    """Provide a compiled graph with mock LLM.

    This is the main fixture for testing the graph execution.
    Uses a mock LLM instead of real Ollama for deterministic responses.
    """

    async def call_llm_with_mock(state: State, runtime):
        """Call LLM using the mock instead of real Ollama."""
        from tools import TOOLS

        llm_with_tools = mock_llm.bind_tools(TOOLS)

        # Prepend Vectora system prompt with auto-detected language
        system_prompt = SystemMessage(content=get_system_prompt())
        messages_with_system = [system_prompt] + list(state["messages"])

        result = llm_with_tools.invoke(messages_with_system)
        return {"messages": [result]}

    from langgraph.constants import END, START
    from langgraph.graph.state import StateGraph
    from langgraph.prebuilt.tool_node import ToolNode, tools_condition

    from tools import TOOLS

    builder = StateGraph(
        state_schema=State,
        context_schema=Context,
        input_schema=State,
        output_schema=State,
    )

    tool_node = ToolNode(tools=TOOLS)

    builder.add_node("call_llm", call_llm_with_mock)
    builder.add_node("tools", tool_node)

    builder.add_edge(START, "call_llm")
    builder.add_conditional_edges("call_llm", tools_condition, ["tools", END])
    builder.add_edge("tools", "call_llm")

    return builder.compile(checkpointer=checkpointer)
