"""Services Layer: Domain-specific business logic without UI dependencies.

Each service encapsulates one responsibility domain and can be tested in isolation.
All services depend only on Settings and standard library/third-party packages.

Services:
- SessionService: Session lifecycle management
- EmbeddingService: Vector store and embeddings
- TelemetryService: Logging and audit trails
- SecurityService: Security validation
"""

from vectora.services.embedding import EmbeddingService
from vectora.services.security import SecurityService
from vectora.services.session import SessionService
from vectora.services.telemetry import TelemetryService

__all__ = [
    "EmbeddingService",
    "SecurityService",
    "SessionService",
    "TelemetryService",
]
