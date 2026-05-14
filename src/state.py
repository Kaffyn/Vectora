from collections.abc import Sequence
from typing import Annotated, Any, TypedDict

from langchain_core.messages import BaseMessage
from langgraph.graph.message import add_messages


class Document(TypedDict, total=False):
    """Estrutura de documento recuperado do RAG."""

    page_content: str
    metadata: dict[str, Any]
    relevance_score: float | None


class State(TypedDict):
    """Estado da conversa com suporte a RAG e histórico gerenciado pelo LangGraph.

    O reducer nativo `add_messages` substitui qualquer lógica manual de sliding window.
    Ele realiza append inteligente, suporta substituição de mensagem por `id` e é
    mantido pela equipe LangChain — tornando-o a fonte da verdade do histórico.

    Campos obrigatórios:
    - messages: Histórico gerenciado por `add_messages` (append automático, sem sobrescrita)

    Campos opcionais:
    - Contexto RAG, routing metadata, histórico resumido
    """

    messages: Annotated[Sequence[BaseMessage], add_messages]

    retrieval_results: dict[str, list[Document]] | None
    selected_rag_source: str | None
    routing_decision: dict[str, Any] | None
    summarized_history: str | None
