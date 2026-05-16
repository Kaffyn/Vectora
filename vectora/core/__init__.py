"""Core Layer: Application orchestration and business logic.

The AgentManager is the central coordinator that ties together all services
and the graph, providing a clean interface for CLI and future integrations.

No UI awareness. Pure Python with type hints and async/await.
"""

from core.agent import AgentManager

__all__ = ["AgentManager"]
