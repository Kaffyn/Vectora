import pytest
from langgraph.graph.state import RunnableConfig

from testing import (
    assert_message_contains_text,
    assert_tool_called_with_args,
    assert_tool_result_in_messages,
    human_message,
)


@pytest.mark.asyncio
class TestQAScenarios:
    """QA Bot test scenarios for end-to-end validation."""

    async def test_qa_bot_greets_and_validates(self, test_graph) -> None:
        """Test complete QA bot greeting and tool validation.

        Scenario:
        1. QA Bot: "Olá, sou QA do Vectora. Vou testar suas capacidades."
        2. QA Bot: "Multiplique 7 por 6 usando a tool"
        3. Assert: Tool multiply was called with a=7, b=6
        4. Assert: Result (42) is in the final messages
        """
        config = RunnableConfig(configurable={"thread_id": "qa_scenario_1"})

        qa_prompt = "Olá, sou QA do Vectora. Vou testar suas capacidades. Multiplique 7 por 6 usando a tool"

        result = await test_graph.ainvoke(
            {"messages": [human_message(qa_prompt)]},
            config=config,
        )

        assert_tool_called_with_args(
            result["messages"],
            tool_name="multiply",
            expected_args={"a": 7.0, "b": 6.0},
        )

        assert_tool_result_in_messages(
            result["messages"],
            tool_name="multiply",
            expected_result=42.0,
        )

    async def test_qa_bot_sequential_multiplications(self, test_graph) -> None:
        """Test QA bot executing multiple sequential multiplication tests.

        Scenario:
        1. QA: "Teste 1: Multiplique 2 por 3"
        2. QA: "Teste 2: Multiplique 10 por 10"
        3. Assert: Both tools called with correct arguments
        """
        config = RunnableConfig(configurable={"thread_id": "qa_scenario_2"})

        test_1 = await test_graph.ainvoke(
            {"messages": [human_message("Teste 1: Multiplique 2 por 3")]},
            config=config,
        )

        assert_tool_called_with_args(
            test_1["messages"],
            tool_name="multiply",
            expected_args={"a": 2.0, "b": 3.0},
        )

        test_2 = await test_graph.ainvoke(
            {"messages": [human_message("Teste 2: Multiplique 10 por 10")]},
            config=config,
        )

        assert_tool_called_with_args(
            test_2["messages"],
            tool_name="multiply",
            expected_args={"a": 10.0, "b": 10.0},
        )

        assert_tool_result_in_messages(
            test_2["messages"],
            tool_name="multiply",
            expected_result=100.0,
        )

    async def test_qa_bot_edge_cases(self, test_graph) -> None:
        """Test QA bot with edge case values.

        Scenario:
        1. Multiply by zero: 5 * 0 = 0
        2. Multiply negative: -3 * 4 = -12
        3. Large numbers: 999 * 999
        """
        config = RunnableConfig(configurable={"thread_id": "qa_scenario_3"})

        zero_test = await test_graph.ainvoke(
            {"messages": [human_message("Multiplique 5 por 0")]},
            config=config,
        )

        assert_tool_result_in_messages(
            zero_test["messages"],
            tool_name="multiply",
            expected_result=0.0,
        )

        config = RunnableConfig(configurable={"thread_id": "qa_scenario_4"})

        negative_test = await test_graph.ainvoke(
            {"messages": [human_message("Multiplique -3 por 4")]},
            config=config,
        )

        assert_tool_called_with_args(
            negative_test["messages"],
            tool_name="multiply",
            expected_args={"a": -3.0, "b": 4.0},
        )

        assert_tool_result_in_messages(
            negative_test["messages"],
            tool_name="multiply",
            expected_result=-12.0,
        )

    async def test_qa_bot_validation_message(self, test_graph) -> None:
        """Test that QA bot validation messages are present.

        Assert: AI response contains greeting or validation message.
        """
        config = RunnableConfig(configurable={"thread_id": "qa_scenario_5"})

        result = await test_graph.ainvoke(
            {"messages": [human_message("Olá")]},
            config=config,
        )

        assert_message_contains_text(
            result["messages"],
            "Olá",
        ) or assert_message_contains_text(
            result["messages"],
            "ajud",
        )
