"""Configuration for Vectora tools loaded from environment variables."""

import os
from dataclasses import dataclass, field


@dataclass
class ToolConfig:
    """Configuration for external tools with environment variable support."""

    # Web Search Tool
    enable_web_search: bool = field(
        default_factory=lambda: os.getenv("ENABLE_WEB_SEARCH", "true").lower() == "true"
    )

    # Web Fetch Tool
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

    # Database Tool
    enable_database: bool = field(
        default_factory=lambda: os.getenv("ENABLE_DATABASE", "false").lower() == "true"
    )
    database_url: str | None = field(default_factory=lambda: os.getenv("DATABASE_URL"))
    allowed_tables: list[str] | None = field(
        default_factory=lambda: _parse_comma_separated(
            os.getenv("DATABASE_ALLOWED_TABLES", "")
        )
    )

    # MCP Server
    enable_mcp: bool = field(
        default_factory=lambda: os.getenv("ENABLE_MCP", "false").lower() == "true"
    )
    mcp_server_url: str | None = field(
        default_factory=lambda: os.getenv("MCP_SERVER_URL")
    )

    # RAG - Embedding (Voyage AI)
    voyage_api_key: str = field(default_factory=lambda: os.getenv("VOYAGE_API_KEY", ""))
    embedding_model: str = field(
        default_factory=lambda: os.getenv("EMBEDDING_MODEL", "voyage-4")
    )
    # Options: voyage-4 (general), voyage-4-lite (cost), voyage-code-3 (code), etc
    embedding_dims: int = field(
        default_factory=lambda: int(os.getenv("EMBEDDING_DIMS", "1024"))
    )

    # RAG - Embedding Queue (Fallback)
    embedding_queue_enabled: bool = field(
        default_factory=lambda: os.getenv("EMBEDDING_QUEUE_ENABLED", "true").lower()
        == "true"
    )
    embedding_queue_db: str = field(
        default_factory=lambda: os.getenv(
            "EMBEDDING_QUEUE_DB", "sqlite:///./data/embedding_queue.db"
        )
    )

    # RAG - Vector Store (Qdrant)
    qdrant_url: str = field(
        default_factory=lambda: os.getenv("QDRANT_URL", "http://localhost:6333")
    )
    qdrant_api_key: str | None = field(
        default_factory=lambda: os.getenv("QDRANT_API_KEY")
    )
    qdrant_collections: list[str] = field(
        default_factory=lambda: _parse_comma_separated(
            os.getenv("QDRANT_COLLECTIONS", "articles,wiki,api_docs,knowledge_base")
        )
        or ["articles", "wiki", "api_docs", "knowledge_base"]
    )

    # RAG - Vector Search
    default_search_top_k: int = field(
        default_factory=lambda: int(os.getenv("RAG_SEARCH_TOP_K", "10"))
    )
    search_min_score: float = field(
        default_factory=lambda: float(os.getenv("RAG_SEARCH_MIN_SCORE", "0.5"))
    )

    # RAG - Reranking (BM25 or Voyage)
    reranker_type: str = field(
        default_factory=lambda: os.getenv("RERANKER_TYPE", "bm25")
    )
    # "bm25" = local (MVP), "voyage" = API
    reranker_model: str = field(
        default_factory=lambda: os.getenv("RERANKER_MODEL", "reranker-2.5-large")
    )
    # Options: reranker-2.5-large, reranker-2.5-mbxl, etc
    reranker_top_k: int = field(
        default_factory=lambda: int(os.getenv("RAG_RERANKER_TOP_K", "5"))
    )


def _parse_comma_separated(value: str) -> list[str] | None:
    """Parse comma-separated string into list, return None if empty."""
    if not value or not value.strip():
        return None
    return [item.strip() for item in value.split(",") if item.strip()]


# Global singleton instance
_config: ToolConfig | None = None


def get_tool_config() -> ToolConfig:
    """Get or create the global tool configuration instance."""
    global _config
    if _config is None:
        _config = ToolConfig()
    return _config
