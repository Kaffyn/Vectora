import logging

from langchain.tools import BaseTool, tool
from langgraph.prebuilt.tool_node import ToolRuntime

from context import Context
from state import State

logger = logging.getLogger(__name__)


@tool
def multiply(a: float, b: float, runtime: ToolRuntime[Context, State]) -> float:  # noqa: ARG001
    """Multiply a * b and returns the result

    Args:
        a: float multiplicand
        b: float multiplier

    Returns:
        the resulting float of the equation a * b
    """
    result = a * b
    logger.info(
        "multiply tool executed",
        extra={"a": a, "b": b, "result": result},
    )
    return result


TOOLS: list[BaseTool] = [multiply]
TOOLS_BY_NAME: dict[str, BaseTool] = {tool.name: tool for tool in TOOLS}
