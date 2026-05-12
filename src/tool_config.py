"""Configuration for Vectora tools loaded from environment variables."""

import os
from dataclasses import dataclass, field
from typing import Optional


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
    allowed_domains: Optional[list[str]] = field(
        default_factory=lambda: _parse_comma_separated(
            os.getenv("WEB_FETCH_ALLOWED_DOMAINS", "")
        )
    )

    # Database Tool
    enable_database: bool = field(
        default_factory=lambda: os.getenv("ENABLE_DATABASE", "false").lower() == "true"
    )
    database_url: Optional[str] = field(
        default_factory=lambda: os.getenv("DATABASE_URL")
    )
    allowed_tables: Optional[list[str]] = field(
        default_factory=lambda: _parse_comma_separated(
            os.getenv("DATABASE_ALLOWED_TABLES", "")
        )
    )

    # MCP Server
    enable_mcp: bool = field(
        default_factory=lambda: os.getenv("ENABLE_MCP", "false").lower() == "true"
    )
    mcp_server_url: Optional[str] = field(
        default_factory=lambda: os.getenv("MCP_SERVER_URL")
    )


def _parse_comma_separated(value: str) -> Optional[list[str]]:
    """Parse comma-separated string into list, return None if empty."""
    if not value or not value.strip():
        return None
    return [item.strip() for item in value.split(",") if item.strip()]


# Global singleton instance
_config: Optional[ToolConfig] = None


def get_tool_config() -> ToolConfig:
    """Get or create the global tool configuration instance."""
    global _config
    if _config is None:
        _config = ToolConfig()
    return _config
