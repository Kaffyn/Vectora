"""SessionService: Manages chat session lifecycle and context.

Responsibilities:
1. Create, switch, list, delete chat sessions
2. Generate runnable configs for LangGraph execution
3. Maintain session metadata (user_type, created_at, etc.)
4. Persist session state to database

Week 2 implementation task: Port session management from checkpointer.py
"""

import logging
from typing import Any

from langgraph.graph.graph import CompiledGraph
from langgraph.types import RunnableConfig
from settings import Settings

logger = logging.getLogger(__name__)


class SessionService:
    """Manages chat session lifecycle.

    Sessions are the unit of conversation persistence. Each session:
    - Has a unique thread_id
    - Maintains conversation history
    - Tracks user_type for context
    - Stores metadata (created_at, last_activity, etc.)
    """

    def __init__(self, settings: Settings):
        """Initialize SessionService.

        Args:
            settings: Application settings
        """
        self.settings = settings
        # TODO: Initialize database connection in Week 2
        # self.db = AsyncSqliteSaver(settings.db_dsn)

        logger.debug("SessionService initialized")

    async def create(self, user_type: str = "default") -> int:
        """Create new chat session.

        Args:
            user_type: User classification ("default" or custom)

        Returns:
            New session/thread ID
        """
        # TODO: Implement in Week 2
        # 1. Find max existing thread_id
        # 2. Increment and create new session
        # 3. Store metadata to database
        # 4. Return thread_id

        logger.debug("New session created")
        return 1  # Placeholder

    async def switch(self, thread_id: int) -> bool:
        """Switch to existing session.

        Args:
            thread_id: Session ID to activate

        Returns:
            True if session exists, False otherwise
        """
        # TODO: Implement in Week 2
        # 1. Query database for session
        # 2. Validate existence
        # 3. Update current context
        # 4. Return success/failure

        logger.debug(f"Switched to session {thread_id}")
        return True  # Placeholder

    async def list_all(self) -> list[dict]:
        """Get all available sessions.

        Returns:
            List of session metadata dicts:
            [
                {
                    "thread_id": 1,
                    "user_type": "default",
                    "created_at": "2026-05-16T10:30:00",
                    "last_activity": "2026-05-16T10:35:00",
                    "message_count": 15,
                },
                ...
            ]
        """
        # TODO: Implement in Week 2
        # 1. Query database for all sessions
        # 2. Get metadata for each
        # 3. Sort by last_activity
        # 4. Return list

        logger.debug("Listed all sessions")
        return [
            {
                "thread_id": 1,
                "user_type": "default",
                "created_at": "2026-05-16T10:30:00",
                "last_activity": "2026-05-16T10:35:00",
                "message_count": 0,
            }
        ]  # Placeholder

    def get_runnable_config(self, thread_id: int) -> RunnableConfig:
        """Get LangGraph runnable config for session.

        Args:
            thread_id: Session ID

        Returns:
            RunnableConfig with thread_id and context
        """
        # TODO: Implement in Week 2
        # 1. Create RunnableConfig with thread_id
        # 2. Inject settings as configurable
        # 3. Return config for graph.ainvoke()

        from context import Context

        context = Context(thread_id=thread_id, user_type="default")

        return RunnableConfig(
            configurable={
                "thread_id": thread_id,
                "context": context,
            }
        )

    async def delete(self, thread_id: int) -> bool:
        """Delete session and its history.

        Args:
            thread_id: Session to delete

        Returns:
            True if deleted, False if not found
        """
        # TODO: Implement in Week 2
        # 1. Validate session exists
        # 2. Delete from database
        # 3. Return success/failure

        logger.warning(f"Session deleted: {thread_id}")
        return True  # Placeholder

    async def get_history(self, thread_id: int, limit: int = 50) -> list[dict]:
        """Get message history for session.

        Args:
            thread_id: Session ID
            limit: Maximum messages to return

        Returns:
            List of messages with role and content
        """
        # TODO: Implement in Week 2
        # 1. Query database for messages
        # 2. Order by timestamp
        # 3. Limit to N most recent
        # 4. Return formatted messages

        logger.debug(f"Retrieved history for session {thread_id}")
        return []  # Placeholder
