from dataclasses import dataclass, field
from datetime import UTC, datetime


@dataclass(kw_only=True, frozen=True, slots=True)
class UserPreferences:
    """User-level preferences that persist across sessions."""

    language: str = "en"
    max_search_results: int = 10
    min_score_threshold: float = 0.5
    preferred_model: str | None = None


@dataclass(kw_only=True, frozen=True, slots=True)
class FeatureFlags:
    """Which tools/features are enabled for this user."""

    enable_web_search: bool = True
    enable_url_fetch: bool = True
    enable_database: bool = True
    enable_rag: bool = True
    enable_mcp: bool = False


@dataclass(kw_only=True, frozen=True, slots=True)
class Context:
    """Immutable context for a LangGraph conversation.

    Set once at conversation start, passed to graph.ainvoke(context=...).
    Contains user info, feature flags, session metadata, and preferences.

    During a conversation, Context is immutable per LangGraph semantics.
    To change context between turns, create a new Context instance.

    Attributes:
        user_id: Unique user identifier
        user_type: User subscription level (plus, enterprise, pro)
        thread_id: Conversation thread ID for persistence
        conversation_id: Optional external conversation ID
        created_at: ISO timestamp when context was created
        preferences: User preferences (language, search settings, etc)
        features: Feature flags controlling tool access
    """

    user_id: str = "default"
    user_type: str
    thread_id: str | int
    conversation_id: str | None = None
    created_at: str | None = None
    preferences: UserPreferences = field(default_factory=UserPreferences)
    features: FeatureFlags = field(default_factory=FeatureFlags)

    def __post_init__(self) -> None:
        """Set created_at if not provided and validate thread_id."""
        if self.thread_id is None:
            msg = "thread_id cannot be None"
            raise ValueError(msg)
        if self.created_at is None:
            object.__setattr__(self, "created_at", datetime.now(UTC).isoformat())
