"""Base — Sanitização de mensagens e chamada LLM genérica.

Funções compartilhadas por todos os workers:
- _sanitize_for_gemini: Remove mensagens inválidas para o Gemini
- call_llm_with: Invoca qualquer LLM bindado com ferramentas dado o state
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from langchain_core.messages import (
    AIMessage,
    HumanMessage,
    SystemMessage,
    ToolMessage,
    trim_messages,
)

from vectora.config.settings import settings
from vectora.services.memory import get_memory_store
from vectora.services.text import text_service

if TYPE_CHECKING:
    from langchain_core.runnables import Runnable

    from vectora.state import State

logger = logging.getLogger(__name__)


def sanitize_for_gemini(messages: list) -> list:
    """Remove mensagens do início que violam as regras de ordenação do Gemini.

    O Gemini exige:
      - Início com HumanMessage OU AIMessage sem tool_calls
      - AIMessage(tool_calls) deve vir APÓS HumanMessage ou ToolMessage
      - ToolMessage deve vir APÓS AIMessage(tool_calls)

    Remove blocos inválidos do início até encontrar um ponto de partida válido.
    """
    result = list(messages)
    while result:
        first = result[0]
        if isinstance(first, HumanMessage):
            break
        if isinstance(first, AIMessage) and not getattr(first, "tool_calls", None):
            break
        if isinstance(first, AIMessage) and getattr(first, "tool_calls", None):
            result = result[1:]
            while result and isinstance(result[0], ToolMessage):
                result = result[1:]
            continue
        result = result[1:]
    return result


async def build_messages(state: State, system_prompt: str = "") -> list:
    """Monta a lista de mensagens para enviar ao LLM.

    Pipeline:
    1. Trim ao max_context_tokens (sliding window)
    2. Fallback: garante pelo menos 1 HumanMessage
    3. Sanitiza para Gemini
    4. Injeta SystemMessage + memórias + contexto RAG
    """
    all_messages = list(state.get("messages", []))
    session_metadata = state.get("session_metadata", {})
    thread_id = session_metadata.get("thread_id", 1)

    # 1. Sliding window
    trimmed = trim_messages(
        all_messages,
        max_tokens=settings.max_context_tokens,
        strategy="last",
        token_counter=text_service.count_messages_tokens,
    )

    # 2. Fallback: sem HumanMessage → pega a partir da última
    has_human = any(isinstance(m, (HumanMessage, AIMessage)) for m in trimmed)
    if not trimmed or not has_human:
        for i in range(len(all_messages) - 1, -1, -1):
            if isinstance(all_messages[i], HumanMessage):
                trimmed = all_messages[i:]
                break
        else:
            trimmed = all_messages

    # 3. Sanitiza para Gemini
    trimmed = sanitize_for_gemini(trimmed)
    if not trimmed:
        for i in range(len(all_messages) - 1, -1, -1):
            if isinstance(all_messages[i], HumanMessage):
                trimmed = [all_messages[i]]
                break

    # 4. System prompt — fornecido pelo agent ou vazio
    system_content = system_prompt

    # Injeta contexto RAG se disponível (injetado pelo rag_subgraph)
    rag_docs = state.get("rag_docs")
    if rag_docs:
        lines = ["\n\n## Contexto RAG disponível (priorize para responder):"]
        for i, doc in enumerate(rag_docs[:5], 1):
            content = doc.get("page_content", "")
            src = doc.get("metadata", {}).get("source", "")
            lines.append(f"[{i}] {content[:600]}" + (f"\nFonte: {src}" if src else ""))
        system_content += "\n".join(lines)

    # Injeta retrieval_results legado se houver
    retrieval_results = state.get("retrieval_results")
    if retrieval_results:
        parts = []
        for col, docs in retrieval_results.items():
            parts.append(f"\nContexto da coleção '{col}':")
            for i, doc in enumerate(docs, 1):
                parts.append(f"[{i}] {doc['page_content']}")
        system_content += "\n\nUSE O CONTEXTO ABAIXO:\n" + "\n".join(parts)

    system_msg = SystemMessage(content=system_content)

    # 5. Memórias persistentes
    memory_messages: list[SystemMessage] = []
    try:
        store = await get_memory_store()
        user_id = f"user_{thread_id}" if thread_id else "default_user"
        memories = await store.get_all(user_id)
        if memories:
            mem_text = "## MEMÓRIAS:\n" + "\n".join(
                f"- **{m['key']}**: {m['content']}" for m in memories
            )
            memory_messages.append(SystemMessage(content=mem_text))
    except Exception as e:
        logger.warning("Falha ao carregar memórias: %s", e)

    return [system_msg, *memory_messages, *trimmed]


async def invoke_llm(llm: Runnable, state: State, system_prompt: str = "") -> dict:
    """Invoca o LLM com o state atual e retorna {'messages': [AIMessage]}.

    Usado por todos os workers — cada um passa seu próprio LLM bindado.
    """
    from vectora.services.tracer import tracer

    session_id: int | None = None
    try:
        session_id = state.get("session_metadata", {}).get("thread_id")  # type: ignore[assignment]
    except Exception:
        pass

    messages = await build_messages(state, system_prompt=system_prompt)

    response_content = ""
    tool_calls_collected = []

    try:
        async with tracer.span("invoke_llm", "call", session_id=session_id) as span:
            try:
                async for chunk in llm.astream(messages):
                    if hasattr(chunk, "content") and chunk.content:
                        content = chunk.content
                        if isinstance(content, list):
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
                    if hasattr(chunk, "tool_calls") and chunk.tool_calls:
                        tool_calls_collected = chunk.tool_calls

            except Exception as e:
                err = str(e)
                if "RESOURCE_EXHAUSTED" in err or "429" in err:
                    logger.warning("Quota esgotada: %s", err[:200])
                    span.set(status="quota_error")
                    return {
                        "messages": [
                            AIMessage(
                                content=(
                                    "**⚠️ Quota da API atingida.**\n"
                                    "Aguarde alguns minutos ou configure outra chave de API."
                                )
                            )
                        ]
                    }
                raise

            # Extrai token counts do response_metadata (Gemini/OpenAI)
            try:
                model_name = getattr(llm, "model", None) or getattr(
                    llm, "model_name", None
                )
                span.set(model=str(model_name) if model_name else "unknown")
            except Exception:
                pass

    except Exception:
        # Tracer não pode quebrar o fluxo — re-lança só se não for do tracer
        raise

    result = AIMessage(content=response_content, tool_calls=tool_calls_collected)
    logger.debug(
        "LLM response",
        extra={"len": len(response_content), "tool_calls": bool(tool_calls_collected)},
    )
    return {"messages": [result]}
