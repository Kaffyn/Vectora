"""LangGraph Construction for Multi-Node Agentic Workflow.

Builds compiled state graph with MAIN_NODE, TOOL_NODE, SUB_NODE pattern.
Coordinates conversation flow, tool execution, and state management.
"""

import logging

from langgraph.constants import END, START
from langgraph.graph.state import CompiledStateGraph, StateGraph
from langgraph.prebuilt.tool_node import tools_condition
from langgraph.pregel.main import BaseCheckpointSaver

from vectora.context import Context
from vectora.nodes.debug import DiagnosticToolNode, call_llm_debug
from vectora.nodes.engine import handle_sub_node, process_retrieval
from vectora.state import State
from vectora.tools import TOOLS

logger = logging.getLogger(__name__)


def build_graph(
    checkpointer: BaseCheckpointSaver,
) -> CompiledStateGraph[State, Context, State, State]:
    """Constrói LangGraph com padrão 3-node: MAIN_NODE, TOOL_NODE, SUB_NODE.

    Fluxo:
    - MAIN_NODE (call_llm): Invoca LLM com histórico deslizante
    - TOOL_NODE (tool_node): Executa ferramentas em paralelo
    - SUB_NODE (handle_sub_node): Workflows complexos em instância separada
    """
    logger.info("Building LangGraph with 3-node pattern: call_llm, tools, sub_node")
    builder = StateGraph(  # type: ignore[type-arg,arg-type]
        state_schema=State,
        context_schema=Context,
        input_schema=State,
        output_schema=State,
    )

    # Create diagnostic nodes for debugging message loss
    diagnostic_tool_node = DiagnosticToolNode(tools=TOOLS)

    builder.add_node("call_llm", call_llm_debug)
    builder.add_node("tools", diagnostic_tool_node)
    builder.add_node("process_retrieval", process_retrieval)
    builder.add_node("sub_node", handle_sub_node)

    builder.add_edge(START, "call_llm")
    builder.add_conditional_edges("call_llm", tools_condition, ["tools", END])
    builder.add_edge("tools", "process_retrieval")
    builder.add_edge("process_retrieval", "call_llm")

    compiled = builder.compile(checkpointer=checkpointer)
    logger.info("Graph compiled successfully with 3-node pattern")
    return compiled
