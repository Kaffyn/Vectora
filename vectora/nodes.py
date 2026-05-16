"""LangGraph Node Implementations for Agent Behavior.

Defines individual nodes: MAIN_NODE (chat), TOOL_NODE (execution), SUB_NODE (isolation).
Each node handles specific responsibilities in the conversation pipeline.
"""

import json
import logging
from contextlib import nullcontext
from datetime import UTC, datetime

from context import Context
from langchain_core.language_models import BaseChatModel
from langchain_core.messages import AIMessage, SystemMessage, ToolMessage, trim_messages
from langchain_core.runnables import Runnable, RunnableConfig
from langgraph.prebuilt.tool_node import ToolNode
from langgraph.runtime import Runtime
from memory_store import get_memory_store
from prompts import get_system_prompt
from state import State
from tools import TOOLS, embedding
from utils import load_llm

logger = logging.getLogger(__name__)


def _simple_token_counter(messages: list) -> int:
    """Simple token counter without requiring LLM initialization.

    Estimates tokens as ~4 characters per token (rough approximation).
    Avoids triggering lazy LLM initialization which causes config leakage.
    """
    total_length = sum(len(str(msg.content or "")) for msg in messages)
    return total_length // 4


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


def _filter_llm_config(config: dict) -> dict:
    """Remove metadados internos do LangGraph/LangChain que causam erros nos modelos.

    O ChatGoogleGenerativeAI e outros modelos são estritamente tipados e
    rejeitam argumentos desconhecidos. O LangGraph injeta variáveis de controle
    internas (thread_id, __pregel_*, checkpoint_*) que não são parâmetros válidos
    do modelo. Este filtro remove essas chaves, deixando apenas o que o LLM entende.

    Args:
        config: Dicionário configurable do RunnableConfig

    Returns:
        Dicionário filtrado sem metadados do LangGraph
    """
    # Lista de chaves que o LangGraph injeta e que causam "Unexpected argument" errors
    blacklist = {
        "thread_id",
        "context",
        "__pregel_runtime",
        "__pregel_replay_state",
        "__pregel_task_id",
        "__pregel_send",
        "__pregel_read",
        "__pregel_checkpointer",
        "checkpoint_map",
        "checkpoint_id",
        "checkpoint_ns",
        "__pregel_scratchpad",
        "__pregel_call",
    }
    return {k: v for k, v in config.items() if k not in blacklist}


async def call_llm(state: State, runtime: Runtime[Context]) -> dict:
    """Nó principal: invoca o LLM com o histórico completo e ferramentas vinculadas.

    Phase 2 Refactor: Lê session_metadata de state ao invés de Context no runtime.
    Isso torna o state JSON-serializable e remove dependência de objetos complexos
    no RunnableConfig.

    Assíncrono para não bloquear o event loop durante a chamada de rede ao LLM.
    Retorna um dict parcial — o LangGraph atualiza apenas as chaves presentes,
    sem exigir todos os campos do State TypedDict.
    O retorno `{"messages": [result]}` faz append via reducer `add_messages`.
    """
    # Lê session metadata de state (JSON-serializable, sem Context)
    session_metadata = state.get("session_metadata", {})
    user_type = session_metadata.get("user_type", "default")
    thread_id = session_metadata.get("thread_id", 1)
    llm_provider = session_metadata.get("llm_provider", "google-genai")
    llm_model = session_metadata.get("llm_model", "gemini-3.1-flash-lite")

    logger.debug(
        "LLM configuration loaded from state",
        extra={
            "provider": llm_provider,
            "model": llm_model,
            "thread_id": thread_id,
            "user_type": user_type,
        },
    )

    # Obtém LLM em cache com ferramentas (vinculado uma vez, reutilizado por invocação)
    # Don't use with_config() as it leaks configurable dict to model kwargs
    llm_with_tools = _get_llm_with_tools()

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
        user_id = f"user_{thread_id}" if thread_id else "default_user"
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
    # Mantém as últimas mensagens até ~1000 tokens (estimado por contador simples)
    # NÃO usar llm_with_tools como token_counter pois força inicialização lazy do LLM
    # com argumentos inválidos do LangGraph (__pregel_*, thread_id, etc)
    trimmed_messages = trim_messages(
        state["messages"],
        max_tokens=1000,
        strategy="last",
        token_counter=_simple_token_counter,
    )

    messages_with_system = [system_prompt, *memory_messages, *trimmed_messages]

    # LangSmith tracing é auto-injetado via env vars (LANGSMITH_API_KEY, etc)
    response_content = ""
    with nullcontext():
        # LangSmith automatically captures duration, tokens, and model info
        # Use astream() to get complete messages without the event overhead
        # which may bypass some of the config leakage issues from astream_events()
        async for chunk in llm_with_tools.astream(
            messages_with_system,
        ):
            # astream() returns AIMessage chunks with partial content
            if hasattr(chunk, "content") and chunk.content:
                # Handle content as string or list of content blocks
                content = chunk.content
                if isinstance(content, list):
                    # Extract text from content blocks
                    for item in content:
                        if isinstance(item, dict) and "text" in item:
                            response_content += item["text"]
                        elif isinstance(item, str):
                            response_content += item
                        else:
                            response_content += str(item)
                elif isinstance(content, dict) and "text" in content:
                    response_content += content["text"]
                else:
                    response_content += str(content)

    # Create AIMessage from collected response
    result = AIMessage(content=response_content)

    logger.debug(
        "Resposta do LLM gerada",
        extra={
            "response_length": len(response_content),
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


async def process_retrieval(state: State, runtime: Runtime[Context]) -> dict:
    """Nó 2: Persistência inteligente (Fire-and-Forget) de resultados Tavily.

    Monitora ToolMessages de 'web_search' e 'fetch_url', enfileira conteúdo
    para embedding assíncrono (fire-and-forget), e retorna status ao agente.
    """
    messages = state["messages"]
    if not messages:
        return {}

    current_retrieval = state.get("retrieval_results") or {}
    new_results_found = False

    # Processa últimas ToolMessages de web_search ou fetch_url
    for msg in reversed(messages):
        if not isinstance(msg, ToolMessage):
            break
        if msg.name not in ["web_search", "fetch_url"]:
            continue

        # Parse e valida JSON da Tavily
        try:
            data = json.loads(msg.content)
        except json.JSONDecodeError:
            logger.warning(
                "Failed to parse tool result as JSON",
                extra={"tool": msg.name, "content_preview": msg.content[:100]},
            )
            continue

        # Extrai lista de resultados (dict com "results" ou lista direta)
        results = _extract_tavily_results(data, msg.name)
        if not results:
            continue

        # Formata documentos e enfileira para embedding
        formatted_docs = await _process_tavily_results(results, msg.name, embedding)
        if not formatted_docs:
            continue

        # Armazena e marca sucesso
        current_retrieval[msg.name] = formatted_docs
        new_results_found = True
        logger.info(
            "retrieval_results_updated",
            extra={"source": msg.name, "doc_count": len(formatted_docs)},
        )

    return {"retrieval_results": current_retrieval} if new_results_found else {}


def _extract_tavily_results(data: dict | list, tool_name: str) -> list[dict] | None:
    """Extrai lista de resultados de estrutura Tavily flexível."""
    if isinstance(data, dict):
        return data.get("results", [])
    if isinstance(data, list):
        return data
    logger.warning(
        "Unexpected Tavily result format",
        extra={"tool": tool_name, "type": type(data).__name__},
    )
    return None


async def _process_tavily_results(
    results: list[dict], source: str, embedding_tool: Runnable
) -> list[dict]:
    """Processa resultados Tavily e enfileira para embedding (fire-and-forget)."""
    formatted_docs = []

    for r in results:
        content = r.get("content", "").strip()
        if not content:
            continue

        # Cria documento estruturado
        doc = {
            "page_content": content,
            "metadata": {
                "title": r.get("title", ""),
                "url": r.get("url", ""),
                "source": source,
            },
        }
        formatted_docs.append(doc)

        # Enfileira documento para embedding assíncrono (fire-and-forget)
        try:
            # Não aguarda — fire-and-forget pattern
            _ = embedding_tool.astream(
                input={"text": content},
                config=RunnableConfig(),
            )
            logger.debug(
                "Document enqueued for embedding",
                extra={
                    "source": source,
                    "title": doc["metadata"]["title"][:50],
                    "content_length": len(content),
                },
            )
        except Exception as e:
            logger.warning(
                "Failed to enqueue document for embedding",
                extra={
                    "source": source,
                    "title": doc["metadata"]["title"][:50],
                    "error": str(e),
                },
            )

    return formatted_docs
