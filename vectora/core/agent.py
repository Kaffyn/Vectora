"""AgentManager: Orchestrator for all Vectora operations.

Pure Python class with NO UI dependencies. Responsible for:
1. Initializing and managing the LangGraph computation
2. Orchestrating all services (session, embedding, telemetry, security)
3. Providing high-level API for CLI and future integrations (HTTP, Discord, etc.)

Design principle: AgentManager is the interface that separates business logic
from UI. CLI, APIs, bots all call the same AgentManager methods.
"""

import logging
from typing import Any

from langchain_core.messages import HumanMessage
from langgraph.graph import MessageGraph
from settings import Settings

logger = logging.getLogger(__name__)


class AgentManager:
    """Orchestrator for Vectora chat agent.

    High-level interface for:
    - Chat execution with context management
    - Model switching and configuration
    - Session lifecycle management
    - Embedding and vector search
    - Logging and audit trails
    - Security validation

    No UI awareness. Can be used by:
    - CLI (main.py)
    - HTTP API (future)
    - Discord Bot (future)
    - Web Dashboard (future)
    """

    def __init__(self, settings: Settings | None = None):
        """Initialize AgentManager with settings and services.

        Args:
            settings: Settings instance (defaults to singleton if None)

        Raises:
            ValidationError: If settings contain invalid configuration
        """
        self.settings = settings or self._load_settings()

        # Initialize services (will be implemented in Week 2)
        self.session_service = None  # SessionService(self.settings)
        self.embedding_service = None  # EmbeddingService(self.settings)
        self.telemetry_service = None  # TelemetryService(self.settings)
        self.security_service = None  # SecurityService(self.settings)

        # Initialize graph (will be implemented in Week 4)
        self.graph = None  # build_graph(self.settings)

        logger.info(
            "AgentManager initialized",
            extra={"llm_provider": self.settings.get_llm_provider()},
        )

    def _load_settings(self) -> Settings:
        """Load Settings singleton.

        Returns:
            Settings instance
        """
        from settings import get_settings

        return get_settings()

    # ========================================================================
    # CHAT OPERATIONS
    # ========================================================================

    async def chat(self, user_input: str, session_id: int = 1) -> str:
        """Execute a chat turn with the agent.

        Args:
            user_input: User message
            session_id: Session/thread ID for context tracking

        Returns:
            Agent response string

        Raises:
            RuntimeError: If graph not initialized or execution fails
        """
        if not self.graph:
            raise RuntimeError("Graph not initialized. Call initialize_graph() first.")

        logger.debug(
            "Chat execution started",
            extra={"session_id": session_id, "input_length": len(user_input)},
        )

        try:
            # TODO: Get runnable config from SessionService
            # config = self.session_service.get_runnable_config(session_id)

            # TODO: Build input state
            # input_state = {"messages": [HumanMessage(user_input)]}

            # TODO: Execute graph
            # result = await self.graph.ainvoke(input_state, config=config)

            # TODO: Extract response from result
            # response = result["messages"][-1].content

            # For now, return a placeholder
            response = f"[Placeholder] Received: {user_input}"

            logger.debug(
                "Chat execution completed",
                extra={
                    "session_id": session_id,
                    "response_length": len(response),
                },
            )

            return response

        except Exception as e:
            logger.exception("Chat execution failed", extra={"session_id": session_id})
            raise

    # ========================================================================
    # MODEL & CONFIGURATION
    # ========================================================================

    async def switch_model(self, provider: str, model: str) -> bool:
        """Switch LLM model.

        Args:
            provider: LLM provider name
            model: Model name

        Returns:
            True if successful, False otherwise
        """
        try:
            self.settings.set_model(provider, model)
            logger.info(
                "Model switched",
                extra={"provider": provider, "model": model},
            )
            return True
        except ValueError as e:
            logger.warning(f"Model switch failed: {e}")
            return False

    def get_available_models(self, provider: str | None = None) -> dict[str, list[str]]:
        """Get available models for providers.

        Args:
            provider: Specific provider or None for all

        Returns:
            Dict mapping provider names to model lists
        """
        all_models = {
            "google-genai": [
                "gemini-3.1-flash-lite",
                "gemini-3.1-flash-lite-preview",
                "gemini-3.1-flash-image-preview",
                "gemini-3.1-pro-preview",
                "gemini-3-flash-preview",
                "gemini-3-pro-image-preview",
            ],
            "openai": [
                "gpt-5.5",
                "gpt-5.4",
                "gpt-5.4-mini",
                "gpt-5.3-codex",
            ],
            "anthropic": [
                "claude-sonnet-4-6",
                "claude-opus-4-6",
                "claude-haiku-4-5-20251001",
            ],
        }

        if provider:
            return {provider: all_models.get(provider, [])}
        return all_models

    # ========================================================================
    # SESSION MANAGEMENT
    # ========================================================================

    async def create_session(self, user_type: str = "default") -> int:
        """Create new chat session.

        Args:
            user_type: User classification for context ("default" or custom)

        Returns:
            New session/thread ID

        Raises:
            RuntimeError: If SessionService not initialized
        """
        if not self.session_service:
            raise RuntimeError("SessionService not initialized")

        session_id = await self.session_service.create(user_type)
        logger.info(
            "Session created", extra={"session_id": session_id, "user_type": user_type}
        )
        return session_id

    async def switch_session(self, session_id: int) -> bool:
        """Switch to a different session.

        Args:
            session_id: Session ID to switch to

        Returns:
            True if successful, False if session not found
        """
        if not self.session_service:
            raise RuntimeError("SessionService not initialized")

        success = await self.session_service.switch(session_id)
        if success:
            logger.info("Session switched", extra={"session_id": session_id})
        else:
            logger.warning("Session switch failed", extra={"session_id": session_id})
        return success

    async def list_sessions(self) -> list[dict]:
        """Get all available sessions.

        Returns:
            List of session metadata dicts with keys:
            - session_id: int
            - created_at: ISO 8601 timestamp
            - message_count: int
            - user_type: str
        """
        if not self.session_service:
            raise RuntimeError("SessionService not initialized")

        return await self.session_service.list_all()

    # ========================================================================
    # EMBEDDING & VECTOR SEARCH
    # ========================================================================

    async def search_vectors(
        self, query: str, collection: str = "web_search"
    ) -> list[dict]:
        """Search vector store using semantic similarity.

        Args:
            query: Search query text
            collection: Collection name ("web_search", "documents", etc.)

        Returns:
            List of matched documents with scores
        """
        if not self.embedding_service:
            logger.warning("EmbeddingService not initialized, returning empty results")
            return []

        try:
            results = await self.embedding_service.search(query, collection)
            logger.debug(
                "Vector search completed",
                extra={"query_length": len(query), "result_count": len(results)},
            )
            return results
        except Exception as e:
            logger.exception("Vector search failed")
            return []

    async def queue_document_for_embedding(
        self, doc_id: str, text: str, collection: str = "documents"
    ) -> bool:
        """Queue document for embedding (fire-and-forget).

        Args:
            doc_id: Unique document identifier
            text: Document text content
            collection: Vector collection name

        Returns:
            True if queued successfully
        """
        if not self.embedding_service:
            logger.warning("EmbeddingService not initialized")
            return False

        try:
            await self.embedding_service.queue_document(doc_id, text, collection)
            logger.debug("Document queued for embedding", extra={"doc_id": doc_id})
            return True
        except Exception as e:
            logger.warning(f"Failed to queue document: {e}")
            return False

    # ========================================================================
    # SECURITY & VALIDATION
    # ========================================================================

    def validate_file_edit(self, file_path: str) -> tuple[bool, str]:
        """Validate if file edit is allowed.

        Args:
            file_path: Path to file

        Returns:
            Tuple of (is_allowed, reason_message)
        """
        if not self.security_service:
            logger.warning("SecurityService not initialized, blocking by default")
            return False, "Security service not available"

        return self.security_service.validate_file_edit(file_path)

    def validate_command(self, command: str) -> tuple[bool, str]:
        """Validate if terminal command is safe to execute.

        Args:
            command: Command string to validate

        Returns:
            Tuple of (is_safe, reason_message)
        """
        if not self.security_service:
            logger.warning("SecurityService not initialized, blocking by default")
            return False, "Security service not available"

        return self.security_service.validate_terminal_command(command)

    # ========================================================================
    # TELEMETRY & LOGGING
    # ========================================================================

    async def export_session_audit(self, session_id: int) -> str:
        """Export session as Markdown audit trail.

        Args:
            session_id: Session to export

        Returns:
            File path to exported Markdown file
        """
        if not self.telemetry_service:
            raise RuntimeError("TelemetryService not initialized")

        return await self.telemetry_service.export_session_audit(session_id)

    async def get_debug_dump(self) -> str:
        """Create debug dump with logs, config, state.

        Returns:
            Path to .tar.gz file containing debug information
        """
        if not self.telemetry_service:
            raise RuntimeError("TelemetryService not initialized")

        return await self.telemetry_service.export_debug_dump()

    # ========================================================================
    # LIFECYCLE
    # ========================================================================

    async def initialize(self) -> None:
        """Initialize all services and graph.

        Called after AgentManager creation to set up dependencies.
        This happens asynchronously to avoid blocking CLI startup.
        """
        logger.info("Initializing AgentManager services...")

        # TODO: Initialize services in Week 2
        # self.session_service = SessionService(self.settings)
        # self.embedding_service = EmbeddingService(self.settings)
        # self.telemetry_service = TelemetryService(self.settings)
        # self.security_service = SecurityService(self.settings)

        # TODO: Start background services
        # await self.embedding_service.start()

        # TODO: Build graph in Week 4
        # self.graph = build_graph(self.settings)

        logger.info("AgentManager initialization complete")

    async def shutdown(self) -> None:
        """Graceful shutdown: stop workers, close connections.

        Should be called on application exit.
        """
        logger.info("Shutting down AgentManager...")

        if self.embedding_service:
            try:
                await self.embedding_service.stop()
            except Exception as e:
                logger.warning(f"Error stopping EmbeddingService: {e}")

        logger.info("AgentManager shutdown complete")
