"""LangGraph Node Implementations for Agent Behavior.

Defines individual nodes: MAIN_NODE (chat), TOOL_NODE (execution), SUB_NODE (isolation).
Each node handles specific responsibilities in the conversation pipeline.
"""

import json
import logging
from contextlib import nullcontext
from datetime import UTC, datetime

from langchain_core.language_models import BaseChatModel
from langchain_core.messages import (
    AIMessage,
    HumanMessage,
    SystemMessage,
    ToolMessage,
    trim_messages,
)
from langchain_core.runnables import Runnable, RunnableConfig
from langgraph.prebuilt.tool_node import ToolNode
from langgraph.runtime import Runtime

from vectora.config.settings import settings
from vectora.context import Context
from vectora.prompts import get_system_prompt
from vectora.services.memory import get_memory_store
from vectora.services.utils import load_llm
from vectora.state import State
from vectora.tools import TOOLS, embedding

logger = logging.getLogger(__name__)


def _simple_token_counter(messages: list) -> int:
    """Simple token counter without requiring LLM initialization.

    Estimates tokens as ~4 characters per token (rough approximation).
    Avoids triggering lazy LLM initialization which causes config leakage.
    """
    total_length = sum(len(str(msg.content or "")) for msg in messages)
    return total_length // 4


def _sanitize_for_gemini(messages: list) -> list:
    """Remove mensagens do início que violam as regras de ordenação do Gemini.

    O Gemini exige que a sequência de mensagens respeite este padrão:
      - Deve começar com HumanMessage OU AIMessage sem tool_calls
      - AIMessage(tool_calls) deve vir IMEDIATAMENTE após HumanMessage ou ToolMessage
      - ToolMessage deve vir IMEDIATAMENTE após AIMessage(tool_calls)

    O trim_messages(strategy="last") pode cortar o histórico no meio de um
    exchange de ferramentas, deixando ToolMessages ou AIMessage(tool_calls) no
    início do slice sem o contexto necessário. Isso gera:
      400 Bad Request: "function call turn comes immediately after a user turn"

    Este sanitizador remove do início qualquer mensagem que crie uma sequência
    inválida, até encontrar um ponto de partida válido.

    Args:
        messages: Lista de mensagens após trim_messages

    Returns:
        Lista com início válido para o Gemini
    """
    result = list(messages)

    while result:
        first = result[0]

        # HumanMessage: início válido sempre
        if isinstance(first, HumanMessage):
            break

        # AIMessage sem tool_calls: início válido (resposta de modelo standalone)
        if isinstance(first, AIMessage) and not getattr(first, "tool_calls", None):
            break

        # AIMessage COM tool_calls no início: inválido — Gemini exige que function
        # call venha após user turn. Pular este AI + todos os ToolMessages seguintes
        # (eles são um bloco atômico: separar quebraria a sequência igualmente).
        if isinstance(first, AIMessage) and getattr(first, "tool_calls", None):
            result = result[1:]
            while result and isinstance(result[0], ToolMessage):
                result = result[1:]
            continue

        # ToolMessage no início: orphan — não tem AIMessage(tool_calls) anterior
        if isinstance(first, ToolMessage):
            result = result[1:]
            continue

        # Qualquer outro tipo desconhecido: remover por segurança
        result = result[1:]

    return result


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
            extra={"user_id": f"user_{thread_id}"},
        )

    # Gerencia o histórico de mensagens (sliding window) para evitar estouro de contexto
    # NÃO usar llm_with_tools como token_counter pois força inicialização lazy do LLM
    # com argumentos inválidos do LangGraph (__pregel_*, thread_id, etc)
    trimmed_messages = trim_messages(
        state["messages"],
        max_tokens=settings.max_context_tokens,
        strategy="last",
        token_counter=_simple_token_counter,
    )

    # Garante pelo menos 1 HumanMessage no histórico.
    # Sem isso, o Google GenAI lança "ValueError: contents are required".
    #
    # Caso crítico: N ToolMessages consecutivos (ex: 10 ferramentas paralelas)
    # consomem todos os max_context_tokens no trim_messages. Os últimos [-3:]
    # ainda seriam só ToolMessages — insuficiente.
    #
    # Solução: caminhar para trás no histórico completo até encontrar a última
    # HumanMessage e incluir tudo a partir dela. Isso preserva o par
    # HumanMessage → AIMessage(tool_calls) → ToolMessage[0..N] que o LLM precisa
    # para fazer sentido dos resultados das ferramentas.
    if not trimmed_messages or not any(
        isinstance(m, (HumanMessage, AIMessage)) for m in trimmed_messages
    ):
        last_human_idx: int | None = None
        for i in range(len(state["messages"]) - 1, -1, -1):
            if isinstance(state["messages"][i], HumanMessage):
                last_human_idx = i
                break

        if last_human_idx is not None:
            trimmed_messages = state["messages"][last_human_idx:]
            logger.debug(
                "trim_messages fallback: usando histórico a partir da última HumanMessage",
                extra={
                    "last_human_idx": last_human_idx,
                    "total_messages": len(state["messages"]),
                    "trimmed_count": len(trimmed_messages),
                },
            )
        else:
            # Sem nenhuma HumanMessage no histórico (raro) — envia tudo
            trimmed_messages = state["messages"]
            logger.warning(
                "trim_messages fallback: nenhuma HumanMessage encontrada, enviando histórico completo",
                extra={"total_messages": len(state["messages"])},
            )

    # Sanitiza sequência para Gemini: remove ToolMessages/AIMessage(tool_calls)
    # órfãos no início que causam "function call turn" 400 Bad Request.
    original_len = len(trimmed_messages)
    trimmed_messages = _sanitize_for_gemini(trimmed_messages)
    if len(trimmed_messages) != original_len:
        logger.debug(
            "sanitize_for_gemini: removidas %d mensagens inválidas do início",
            original_len - len(trimmed_messages),
            extra={"remaining": len(trimmed_messages)},
        )

    # Última linha de defesa: se sanitização zerou tudo, usar a última HumanMessage
    if not trimmed_messages:
        for i in range(len(state["messages"]) - 1, -1, -1):
            if isinstance(state["messages"][i], HumanMessage):
                trimmed_messages = [state["messages"][i]]
                logger.warning(
                    "sanitize_for_gemini: histórico inválido completo — usando apenas última HumanMessage",
                )
                break

    messages_with_system = [system_prompt, *memory_messages, *trimmed_messages]

    # LangSmith tracing é auto-injetado via env vars (LANGSMITH_API_KEY, etc)
    response_content = ""
    tool_calls_collected = []
    try:
        with nullcontext():
            # LangSmith automatically captures duration, tokens, and model info
            # Use astream() to get complete messages without the event overhead
            # which may bypass some of the config leakage issues from astream_events()
            async for chunk in llm_with_tools.astream(
                messages_with_system,
            ):
                # astream() returns AIMessage chunks with partial content and tool_calls
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

                # Capture tool_calls from chunks (may accumulate across multiple chunks)
                if hasattr(chunk, "tool_calls") and chunk.tool_calls:
                    tool_calls_collected = chunk.tool_calls
    except Exception as e:
        error_str = str(e)
        if "RESOURCE_EXHAUSTED" in error_str or "429" in error_str:
            logger.warning("Quota da API esgotada: %s", error_str[:200])
            return {
                "messages": [
                    AIMessage(
                        content=(
                            "**⚠️ Limite de quota da API atingido.**\n\n"
                            "Seu plano atual esgotou as requisições disponíveis. "
                            "Aguarde alguns minutos ou configure uma chave de API diferente."
                        )
                    )
                ]
            }
        raise

    # Create AIMessage from collected response, preserving tool_calls
    result = AIMessage(
        content=response_content,
        tool_calls=tool_calls_collected,
    )

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
