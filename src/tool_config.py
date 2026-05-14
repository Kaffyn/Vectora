import os
from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class ToolConfig:
    """Configuração de ferramentas do Vectora — MVP local-first.

    Toda a infraestrutura é file-based:
    - Vector Store: LanceDB (diretório local)
    - Embedding Queue: SQLite (arquivo local)
    - Checkpointer: SQLite (arquivo local)

    Não há dependências de containers externos (PostgreSQL, Qdrant, Valkey).
    """

    # Ferramenta de Busca Web
    enable_web_search: bool = field(
        default_factory=lambda: os.getenv("ENABLE_WEB_SEARCH", "true").lower() == "true"
    )

    # Ferramenta de Busca de URL
    enable_web_fetch: bool = field(
        default_factory=lambda: os.getenv("ENABLE_WEB_FETCH", "true").lower() == "true"
    )
    max_fetch_size: int = field(
        default_factory=lambda: int(os.getenv("WEB_FETCH_MAX_SIZE", "5000"))
    )
    allowed_domains: list[str] | None = field(
        default_factory=lambda: _parse_comma_separated(
            os.getenv("WEB_FETCH_ALLOWED_DOMAINS", "")
        )
    )

    # Operações de Arquivo (file_read, file_edit, grep, list_dir, terminal)
    enable_file_operations: bool = field(
        default_factory=lambda: (
            os.getenv("ENABLE_FILE_OPERATIONS", "true").lower() == "true"
        )
    )

    # Servidor MCP
    enable_mcp: bool = field(
        default_factory=lambda: os.getenv("ENABLE_MCP", "false").lower() == "true"
    )
    mcp_server_url: str | None = field(
        default_factory=lambda: os.getenv("MCP_SERVER_URL")
    )
    mcp_transport_type: str = field(
        default_factory=lambda: os.getenv("MCP_TRANSPORT_TYPE", "http")
    )
    mcp_command: str | None = field(default_factory=lambda: os.getenv("MCP_COMMAND"))
    mcp_command_args: list[str] | None = field(
        default_factory=lambda: _parse_comma_separated(
            os.getenv("MCP_COMMAND_ARGS", "")
        )
    )

    # RAG — Vector Store: LanceDB (file-based, zero-config)
    enable_rag: bool = field(
        default_factory=lambda: os.getenv("ENABLE_RAG", "true").lower() == "true"
    )
    lancedb_dir: str = field(
        default_factory=lambda: os.getenv("LANCEDB_DIR", "./data/lancedb")
    )

    # RAG — Embedding (Voyage AI)
    voyage_api_key: str = field(default_factory=lambda: os.getenv("VOYAGE_API_KEY", ""))
    embedding_model: str = field(
        default_factory=lambda: os.getenv("EMBEDDING_MODEL", "voyage-4")
    )
    embedding_dims: int = field(
        default_factory=lambda: int(os.getenv("EMBEDDING_DIMS", "1024"))
    )

    # RAG — Embedding Queue (SQLite fallback para falhas de API)
    embedding_queue_enabled: bool = field(
        default_factory=lambda: (
            os.getenv("EMBEDDING_QUEUE_ENABLED", "true").lower() == "true"
        )
    )
    embedding_queue_db: str = field(
        default_factory=lambda: os.getenv(
            "EMBEDDING_QUEUE_DB", "sqlite+aiosqlite:///./data/embedding_queue.db"
        )
    )

    # RAG — Busca Vetorial
    default_search_top_k: int = field(
        default_factory=lambda: int(os.getenv("RAG_SEARCH_TOP_K", "10"))
    )
    search_min_score: float = field(
        default_factory=lambda: float(os.getenv("RAG_SEARCH_MIN_SCORE", "0.5"))
    )

    # RAG — Reranking (BM25 local ou Voyage API)
    reranker_type: str = field(
        default_factory=lambda: os.getenv("RERANKER_TYPE", "bm25")
    )
    reranker_model: str = field(
        default_factory=lambda: os.getenv("RERANKER_MODEL", "reranker-2.5-large")
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
