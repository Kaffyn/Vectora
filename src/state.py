from collections.abc import Sequence
from typing import Annotated, Any, TypedDict

from langgraph.graph.message import BaseMessage, add_messages


class Document(TypedDict, total=False):
    """Estrutura de documento recuperado do RAG"""

    page_content: str
    metadata: dict[str, Any]
    relevance_score: float | None


class State(TypedDict):
    """Estado da conversa com suporte a RAG.

    Campos obrigatórios: messages (histórico)
    Campos opcionais: RAG context, routing metadata
    """

    # Histórico de mensagens (obrigatório, reduzido por add_messages)
    messages: Annotated[Sequence[BaseMessage], add_messages]

    # RAG Context (opcionais)
    retrieval_results: dict[str, list[Document]] | None
    # Exemplo: {"articles": [...], "wiki": [...], "api_docs": [...]}

    selected_rag_source: str | None
    # Qual fonte RAG foi usada: "articles", "wiki", "api_docs", "knowledge_base"

    routing_decision: dict[str, Any] | None
    # Metadata da decisão tomada
    # {"decision": "use_vector_search", "confidence": 0.95, "reason": "..."}
