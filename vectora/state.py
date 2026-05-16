"""LangGraph State Definition and Message Management.

Defines the state schema for conversation: messages, summary, retrieval results.
Includes reducer for message deduplication and history management.
"""

from collections.abc import Sequence
from typing import Annotated, Any, TypedDict

from langchain_core.messages import BaseMessage
from langgraph.graph.message import add_messages


class Document(TypedDict, total=False):
    """Estrutura de documento recuperado do RAG."""

    page_content: str
    metadata: dict[str, Any]
    relevance_score: float | None


class SessionMetadata(TypedDict, total=False):
    """Session metadata for context tracking (JSON-serializable).

    Replaces complex Context object in RunnableConfig.
    All fields are immutable and part of State (JSON-safe).

    Fields:
    - thread_id: Unique session identifier
    - user_type: User classification (default or custom)
    - created_at: ISO 8601 timestamp
    - llm_provider: Active LLM provider (google-genai, openai, etc.)
    - llm_model: Active model name
    """

    thread_id: int
    user_type: str
    created_at: str  # ISO 8601
    llm_provider: str
    llm_model: str


class State(TypedDict):
    """Estado da conversa com suporte a RAG e histórico gerenciado pelo LangGraph.

    O reducer nativo `add_messages` substitui qualquer lógica manual de sliding window.
    Ele realiza append inteligente, suporta substituição de mensagem por `id` e é
    mantido pela equipe LangChain — tornando-o a fonte da verdade do histórico.

    Session metadata (thread_id, user_type, etc.) agora faz parte do State,
    tornando o estado JSON-serializable. Remova Context do RunnableConfig.

    Campos obrigatórios:
    - messages: Histórico gerenciado por `add_messages` (append automático)
    - session_metadata: Session context (thread_id, user_type, timestamps)

    Campos opcionais:
    - Contexto RAG, routing metadata, histórico resumido
    """

    messages: Annotated[Sequence[BaseMessage], add_messages]
    session_metadata: SessionMetadata

    retrieval_results: dict[str, list[Document]] | None
    selected_rag_source: str | None
    routing_decision: dict[str, Any] | None
    summarized_history: str | None
