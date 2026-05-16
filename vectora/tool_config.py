"""Tool Configuration with Environment Variable Parsing.

Dataclass managing all tool-specific settings: RAG, embedding, web search, MCP.
Provides property-based normalization for database URLs and path handling.

Todos os paths usam a estrutura roaming: ~/.vectora/
"""

import os
from dataclasses import dataclass, field
from pathlib import Path

from constants import EMBEDDING_QUEUE_DSN, LANCEDB_DIR


@dataclass
class ToolConfig:
    """Configuração de ferramentas do Vectora — MVP local-first.

    Toda a infraestrutura é file-based com roaming (~/.vectora/):
    - Vector Store: LanceDB (diretório local em ~/.vectora/data/lancedb)
    - Embedding Queue: SQLite (arquivo local em ~/.vectora/data/embedding_queue.db)
    - Checkpointer: SQLite (arquivo local em ~/.vectora/data/vectora.db)

    Não há dependências de containers externos (PostgreSQL, Qdrant, Valkey).
    """

    # Ferramenta de Busca Web (web_search e fetch_url via Tavily)
    enable_web_search: bool = field(
        default_factory=lambda: os.getenv("ENABLE_WEB_SEARCH", "true").lower() == "true"
    )
    tavily_api_key: str = field(default_factory=lambda: os.getenv("TAVILY_API_KEY", ""))

    # Operações de Arquivo (file_read, file_edit, grep, list_dir, terminal)
    enable_file_operations: bool = field(
        default_factory=lambda: (
            os.getenv("ENABLE_FILE_OPERATIONS", "true").lower() == "true"
        )
    )

    # Servidor MCP (Vector Context Protocol)
    enable_mcp: bool = field(
        default_factory=lambda: os.getenv("ENABLE_MCP", "false").lower() == "true"
    )
    mcp_server_url: str | None = field(
        default_factory=lambda: os.getenv("MCP_SERVER_URL")
    )
    mcp_transport_type: str = field(
        default_factory=lambda: os.getenv("MCP_TRANSPORT_TYPE", "stdio")
    )
    mcp_command: str | None = field(default_factory=lambda: os.getenv("MCP_COMMAND"))
    mcp_command_args: list[str] | None = field(
        default_factory=lambda: _parse_comma_separated(
            os.getenv("MCP_COMMAND_ARGS", "")
        )
    )
    mcp_timeout: int = field(
        default_factory=lambda: int(os.getenv("MCP_TIMEOUT", "30"))
    )

    # RAG — Vector Store: LanceDB (file-based, roaming em ~/.vectora/data/lancedb)
    enable_rag: bool = field(
        default_factory=lambda: os.getenv("ENABLE_RAG", "true").lower() == "true"
    )
    lancedb_dir: str = field(
        default_factory=lambda: os.getenv("LANCEDB_DIR", LANCEDB_DIR)
    )

    # RAG — Embedding (Voyage AI)
    voyage_api_key: str = field(default_factory=lambda: os.getenv("VOYAGE_API_KEY", ""))
    embedding_model: str = field(
        default_factory=lambda: os.getenv("EMBEDDING_MODEL", "voyage-3-lite")
    )
    embedding_dims: int = field(
        default_factory=lambda: int(os.getenv("EMBEDDING_DIMS", "512"))
    )

    # RAG — Embedding Queue (SQLite com enfileiramento assíncrono em ~/.vectora/data/)
    embedding_queue_enabled: bool = field(
        default_factory=lambda: (
            os.getenv("EMBEDDING_QUEUE_ENABLED", "true").lower() == "true"
        )
    )
    embedding_queue_db: str = field(
        default_factory=lambda: os.getenv("EMBEDDING_QUEUE_DB", EMBEDDING_QUEUE_DSN)
    )

    # RAG — Busca Vetorial
    default_search_top_k: int = field(
        default_factory=lambda: int(os.getenv("RAG_SEARCH_TOP_K", "10"))
    )
    search_min_score: float = field(
        default_factory=lambda: float(os.getenv("RAG_SEARCH_MIN_SCORE", "0.5"))
    )

    # RAG — Reranking (Voyage AI)
    reranker_type: str = field(
        default_factory=lambda: os.getenv("RERANKER_TYPE", "voyage")
    )
    reranker_model: str = field(
        default_factory=lambda: os.getenv("RERANKER_MODEL", "reranker-2-lite")
    )
    reranker_top_k: int = field(
        default_factory=lambda: int(os.getenv("RAG_RERANKER_TOP_K", "5"))
    )

    @property
    def lancedb_path(self) -> Path:
        """Retorna o caminho absoluto do diretório LanceDB, criando se necessário."""
        path = Path(self.lancedb_dir)
        path.mkdir(parents=True, exist_ok=True)
        return path

    @property
    def embedding_queue_url(self) -> str:
        """Normalize embedding queue DB URL format.

        Accepts both shorthand `:memory:` and full SQLAlchemy URLs.
        Returns proper SQLAlchemy URL format.
        """
        if self.embedding_queue_db == ":memory:":
            return "sqlite+aiosqlite:///:memory:"
        if self.embedding_queue_db.startswith(
            "sqlite://"
        ) or self.embedding_queue_db.startswith("sqlite+aiosqlite://"):
            return self.embedding_queue_db
        # Assume file path, convert to SQLAlchemy URL
        if not self.embedding_queue_db.startswith("sqlite"):
            # Treat as file path
            return f"sqlite+aiosqlite:///{self.embedding_queue_db}"
        return self.embedding_queue_db


def _parse_comma_separated(value: str) -> list[str] | None:
    """Analisa string separada por vírgulas em lista, retorna None se vazio."""
    if not value or not value.strip():
        return None
    return [item.strip() for item in value.split(",") if item.strip()]


# Instância singleton global
_config: ToolConfig | None = None


def get_tool_config() -> ToolConfig:
    """Get or create the global tool configuration instance."""
    global _config
    if _config is None:
        _config = ToolConfig()
    return _config
