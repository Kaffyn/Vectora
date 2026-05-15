"""LangGraph Node Implementations for Agent Behavior.

Defines individual nodes: MAIN_NODE (chat), TOOL_NODE (execution), SUB_NODE (isolation).
Each node handles specific responsibilities in the conversation pipeline.
"""

import json
import logging

from langchain_core.language_models import BaseChatModel
from langchain_core.messages import SystemMessage, ToolMessage, trim_messages
from langchain_core.runnables import Runnable
from langgraph.prebuilt.tool_node import ToolNode
from langgraph.runtime import Runtime

from context import Context
from langsmith_integration import create_trace_context
from memory_store import get_memory_store
from prompts import get_system_prompt
from state import State
from tools import TOOLS
from utils import load_llm

logger = logging.getLogger(__name__)

tool_node = ToolNode(tools=TOOLS)

# Inicializa LLM com ferramentas uma única vez no carregamento do módulo (não por invocação)
_llm_base: BaseChatModel | None = None
# bind_tools() retorna Runnable, não BaseChatModel — tipagem correta evita erro de atributo
_llm_with_tools: Runnable | None = None


def _get_llm_with_tools() -> Runnable:
    """Retorna LLM em cache com ferramentas vinculadas (inicializado uma vez por processo)."""
    global _llm_with_tools
    if _llm_with_tools is None:
        _llm_with_tools = load_llm().bind_tools(TOOLS)
        logger.debug("LLM com ferramentas inicializado e em cache")
    return _llm_with_tools


async def call_llm(state: State, runtime: Runtime[Context]) -> dict:
    """Nó principal: invoca o LLM com o histórico completo e ferramentas vinculadas.

    Assíncrono para não bloquear o event loop durante a chamada de rede ao LLM.
    Retorna um dict parcial — o LangGraph atualiza apenas as chaves presentes,
    sem exigir todos os campos do State TypedDict.
    O retorno `{"messages": [result]}` faz append via reducer `add_messages`.
    """
    ctx = runtime.context
    user_type = ctx.user_type

    model_provider = "ollama"
    model = "gpt-oss:20b" if user_type == "plus" else "qwen3-coder:30b"

    # Obtém LLM em cache com ferramentas (vinculado uma vez, reutilizado por invocação)
    llm_with_tools = _get_llm_with_tools()
    llm_with_config = llm_with_tools.with_config(
        config={
            "configurable": {
                "model": model,
                "model_provider": model_provider,
            }
        }
    )

    # Prepara prompt do sistema Vectora com detecção automática de idioma
    system_content = get_system_prompt()

    # Injeta contexto recuperado (RAG), se houver
    if state.get("retrieval_results"):
        context_parts = []
        for collection, docs in state["retrieval_results"].items():
            context_parts.append(f"\nContexto recuperado da coleção '{collection}':")
            for i, doc in enumerate(docs, 1):
                context_parts.append(f"[{i}] {doc['page_content']}")

        system_content += "\n\nUSE O CONTEXTO ABAIXO PARA RESPONDER:\n" + "\n".join(
            context_parts
        )

    system_prompt = SystemMessage(content=system_content)

    # Injeta memórias persistentes do usuário após a system message
    memory_messages: list[SystemMessage] = []
    try:
        memory_store = await get_memory_store()
        user_id = ctx.thread_id or "default_user"
        memories = await memory_store.get_all(user_id)

        if memories:
            memory_content = "## MEMÓRIAS PERSISTENTES:\n"
            for mem in memories:
                memory_content += f"\n- **{mem['key']}**: {mem['content']}\n"

            memory_messages.append(SystemMessage(content=memory_content))
            logger.debug(
                "Memórias injetadas no contexto",
                extra={"count": len(memories), "user_id": user_id},
            )
    except Exception as e:
        logger.warning(
            "Falha ao carregar memórias: %s",
            e,
            extra={"user_id": getattr(ctx, "thread_id", "unknown")},
        )

    # Gerencia o histórico de mensagens (sliding window) para evitar estouro de contexto
    # Mantém as últimas mensagens até ~1000 tokens (estimado pelo modelo)
    trimmed_messages = trim_messages(
        state["messages"],
        max_tokens=1000,
        strategy="last",
        token_counter=llm_with_config,
    )

    messages_with_system = [system_prompt, *memory_messages, *trimmed_messages]

    # Trace com LangSmith se habilitado
    trace_context = await create_trace_context(
        run_name="vectora_llm_call",
        context={
            "user_id": str(ctx.user_id),
            "user_type": ctx.user_type,
            "thread_id": str(ctx.thread_id),
            "correlation_id": str(ctx.correlation_id),
        },
    )

    with trace_context:
        # LangSmith automatically captures duration, tokens, and model info
        # No need for manual timing—let the trace context handle metrics
        result = await llm_with_config.ainvoke(
            messages_with_system,
        )

    logger.debug(
        "Resposta do LLM gerada",
        extra={
            "tem_chamadas_de_ferramentas": bool(getattr(result, "tool_calls", None)),
        },
    )

    # add_messages faz append automaticamente — não sobrescreve o histórico
    return {"messages": [result]}


async def handle_sub_node(state: State, runtime: Runtime[Context]) -> dict:
    """Nó subordinado para workflows complexos em instância separada de LangGraph.

    Por enquanto, passa as mensagens de volta ao nó principal. No futuro,
    pode instanciar um novo grafo para processos mais complexos.
    Retorna um dict vazio pois não altera o estado no momento.
    """
    logger.debug(
        "Sub-node invoked",
        extra={
            "mensagens": len(state["messages"]),
            "retrieval_results": bool(state.get("retrieval_results")),
        },
    )

    return {}


async def process_retrieval(state: State) -> dict:
    """Processa resultados de ferramentas de busca e atualiza o estado RAG.

    Varre as mensagens recentes em busca de ToolMessages de 'vector_search',
    analisa o JSON e preenche 'retrieval_results'.
    """
    messages = state["messages"]
    if not messages:
        return {}

    # Identifica mensagens de ferramentas no turno atual (após o último AIMessage)
    current_retrieval = state.get("retrieval_results") or {}
    new_results_found = False

    # Itera de trás para frente até encontrar uma mensagem que não seja de ferramenta
    for msg in reversed(messages):
        if not isinstance(msg, ToolMessage):
            break

        if msg.name == "vector_search":
            try:
                data = json.loads(msg.content)
                if data.get("status") != "success":
                    continue

                results = data.get("results", [])
                collection = data.get("collection", "default")

                formatted_docs = [
                    {
                        "page_content": r["content"],
                        "metadata": r.get("metadata", {}),
                        "relevance_score": r.get("relevance_score"),
                    }
                    for r in results
                ]

                logger.info(
                    "retrieval_processed",
                    extra={"collection": collection, "count": len(formatted_docs)},
                )

                current_retrieval[collection] = formatted_docs
                new_results_found = True

            except Exception:
                logger.exception("Error processing retrieval for %s", msg.name)

    return {"retrieval_results": current_retrieval} if new_results_found else {}
