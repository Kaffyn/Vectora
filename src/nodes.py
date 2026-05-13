from langchain_core.messages import SystemMessage
from langgraph.prebuilt.tool_node import ToolNode
from langgraph.runtime import Runtime

import logging
from context import Context
from prompts import get_system_prompt
from state import State
from tools import TOOLS
from utils import load_llm

logger = logging.getLogger(__name__)

tool_node = ToolNode(tools=TOOLS)


def call_llm(state: State, runtime: Runtime[Context]) -> State:
    ctx = runtime.context
    user_type = ctx.user_type

    model_provider = "ollama" if user_type == "plus" else "ollama"
    model = "gpt-oss:20b" if user_type == "plus" else "qwen3-coder:30b"

    llm_with_tools = load_llm().bind_tools(TOOLS)
    llm_with_config = llm_with_tools.with_config(
        config={
            "configurable": {
                "model": model,
                "model_provider": model_provider,
            }
        }
    )

    # Prepend Vectora system prompt with auto-detected language
    system_prompt = SystemMessage(content=get_system_prompt())
    messages_with_system = [system_prompt] + list(state["messages"])

    result = llm_with_config.invoke(
        messages_with_system,
    )

    logger.debug(
        "LLM response generated",
        extra={"has_tool_calls": bool(getattr(result, "tool_calls", None))},
    )

    return {"messages": [result]}
