from collections.abc import Sequence
from typing import Annotated, Any, TypedDict

from langgraph.graph.message import BaseMessage


class Document(TypedDict, total=False):
    """Estrutura de documento recuperado do RAG."""

    page_content: str
    metadata: dict[str, Any]
    relevance_score: float | None


def _reduce_messages(
    messages: Sequence[BaseMessage], new_messages: Sequence[BaseMessage]
) -> Sequence[BaseMessage]:
    """Reduz histórico mantendo últimas N mensagens e sumarizando anteriores.

    Mantém as últimas 10 mensagens intactas. Se houver mais de 15 mensagens,
    as anteriores são marcadas para resumo futuro (quando LangSmith estiver ativo).

    Este é um placeholder para auto-summarização em 80% do token limit.
    No futuro, integrará com LangSmith para contar tokens reais.
    """
    max_recent_messages = 10
    summarization_threshold = 15

    all_messages = list(messages) + list(new_messages)

    if len(all_messages) <= summarization_threshold:
        return tuple(all_messages)

    recent = all_messages[-max_recent_messages:]
    return tuple(recent)


class State(TypedDict):
    """Estado da conversa com suporte a RAG e histórico deslizante.

    Campos obrigatórios:
    - messages: Histórico deslizante com auto-summarização

    Campos opcionais:
    - Contexto RAG, routing metadata, histórico resumido
    """

    messages: Annotated[Sequence[BaseMessage], _reduce_messages]

    retrieval_results: dict[str, list[Document]] | None
    selected_rag_source: str | None
    routing_decision: dict[str, Any] | None
    summarized_history: str | None
