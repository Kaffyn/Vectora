"""Debug wrapper para diagnosticar fluxo de mensagens no LangGraph."""

import json
import logging
from typing import Any

from langchain_core.runnables import RunnableConfig
from langgraph.prebuilt.tool_node import ToolNode
from langgraph.runtime import Runtime

from vectora.state import State

logger = logging.getLogger(__name__)


class DiagnosticToolNode(ToolNode):
    """ToolNode com logging detalhado para diagnosticar perda de ToolMessages."""

    async def ainvoke(
        self, state_input: State, config: RunnableConfig | None = None, **kwargs: Any
    ) -> dict[str, Any]:
        """Executa tools com logging de entrada e saída."""
        logger.info(
            "[TOOL_NODE] ENTRADA",
            extra={
                "messages_count": len(state_input.get("messages", [])),
                "last_message_type": (
                    type(state_input["messages"][-1]).__name__
                    if state_input.get("messages")
                    else None
                ),
            },
        )

        # Log do AIMessage com tool_calls
        if state_input.get("messages"):
            last_msg = state_input["messages"][-1]
            if hasattr(last_msg, "tool_calls") and last_msg.tool_calls:
                logger.info(
                    "[TOOL_NODE] Detectadas tool_calls",
                    extra={
                        "tool_calls_count": len(last_msg.tool_calls),
                        "tool_names": [
                            (tc.get("name") if isinstance(tc, dict) else tc.name)
                            for tc in last_msg.tool_calls
                        ],
                    },
                )
            else:
                logger.info(
                    "[TOOL_NODE] Nenhuma tool_call detectada na última mensagem"
                )

        # Chama ToolNode original
        try:
            result = await super().ainvoke(state_input, config, **kwargs)

            logger.info(
                "[TOOL_NODE] SAÍDA",
                extra={
                    "result_keys": list(result.keys()) if result else [],
                    "result_type": type(result).__name__ if result else "None",
                    "messages_in_result": (
                        len(result.get("messages", []))
                        if result and isinstance(result, dict) and "messages" in result
                        else 0
                    ),
                },
            )

            # Log de cada ToolMessage retornado
            if result and isinstance(result, dict) and "messages" in result:
                for i, msg in enumerate(result["messages"]):
                    logger.info(
                        f"[TOOL_NODE] ToolMessage[{i}]",
                        extra={
                            "type": type(msg).__name__,
                            "tool_use_id": getattr(msg, "tool_use_id", "N/A"),
                            "content_length": len(str(getattr(msg, "content", ""))),
                        },
                    )
            elif result:
                logger.warning(
                    "[TOOL_NODE] Resultado não é dict ou não tem 'messages'",
                    extra={
                        "result_type": type(result).__name__,
                        "has_messages_key": "messages" in result
                        if isinstance(result, dict)
                        else False,
                    },
                )

            return result

        except Exception as e:
            logger.exception(
                "[TOOL_NODE] ERRO ao executar ferramentas",
                extra={"error": str(e)},
            )
            raise

    def invoke(
        self, state_input: State, config: RunnableConfig | None = None, **kwargs: Any
    ) -> dict[str, Any]:
        """Sync wrapper para invoke."""
        logger.info(
            "[TOOL_NODE-SYNC] ENTRADA",
            extra={
                "messages_count": len(state_input.get("messages", [])),
                "last_message_type": (
                    type(state_input["messages"][-1]).__name__
                    if state_input.get("messages")
                    else None
                ),
            },
        )

        try:
            result = super().invoke(state_input, config, **kwargs)

            logger.info(
                "[TOOL_NODE-SYNC] SAÍDA",
                extra={
                    "result_keys": list(result.keys())
                    if isinstance(result, dict)
                    else "N/A",
                    "result_type": type(result).__name__,
                },
            )

            return result

        except Exception as e:
            logger.exception(
                "[TOOL_NODE-SYNC] ERRO ao executar ferramentas",
                extra={"error": str(e)},
            )
            raise


async def call_llm_debug(
    state: State, runtime: Runtime | None = None, **kwargs: Any
) -> dict[str, Any]:
    """Wrapper de call_llm com logging de entrada/saída para debugging."""
    from vectora.nodes.engine import call_llm as original_call_llm

    logger.info(
        "[CALL_LLM] ENTRADA",
        extra={
            "messages_count": len(state.get("messages", [])),
            "messages_summary": [
                {
                    "type": type(m).__name__,
                    "content_length": len(str(m.content)) if m.content else 0,
                }
                for m in state.get("messages", [])[-5:]  # Últimas 5
            ],
        },
    )

    try:
        result = await original_call_llm(state, runtime, **kwargs)

        logger.info(
            "[CALL_LLM] SAÍDA",
            extra={
                "result_keys": list(result.keys())
                if isinstance(result, dict)
                else "N/A",
                "result_type": type(result).__name__,
                "messages_added": (
                    len(result.get("messages", []))
                    if result and isinstance(result, dict) and "messages" in result
                    else 0
                ),
            },
        )

        # Log do AIMessage retornado
        if result and isinstance(result, dict) and "messages" in result:
            for msg in result["messages"]:
                has_tool_calls = bool(getattr(msg, "tool_calls", None))
                logger.info(
                    "[CALL_LLM] AIMessage gerado",
                    extra={
                        "type": type(msg).__name__,
                        "content_length": len(str(msg.content)) if msg.content else 0,
                        "has_tool_calls": has_tool_calls,
                        "tool_calls_count": (
                            len(msg.tool_calls) if has_tool_calls else 0
                        ),
                    },
                )

        return result
    except Exception as e:
        logger.exception(
            "[CALL_LLM] ERRO durante execução",
            extra={"error": str(e)},
        )
        raise
