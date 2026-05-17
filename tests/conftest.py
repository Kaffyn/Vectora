"""Pytest Configuration and Global Test Fixtures.

Configures pytest environment, sets up logging, and provides reusable fixtures
for all tests: mock_llm, temp_db, test_context, test_graph, checkpointer, etc.
"""

import os

import pytest

from vectora.services.log_setup import setup_logging
from vectora.testing.fixtures import (
    checkpointer,
    mock_llm,
    temp_db,
    test_context,
    test_graph,
    vector_store_dir,
)

pytest_plugins = ("pytest_asyncio",)


@pytest.fixture(scope="session", autouse=True)
def _setup_test_environment() -> None:
    """Setup environment and logging for all tests.

    Configures SQLite + LanceDB for lightweight testing without
    external dependencies (PostgreSQL, Qdrant).
    """
    # Logging
    os.environ.setdefault("LOG_LEVEL", "DEBUG")
    os.environ.setdefault("LOG_JSON", "false")

    # Vector Store: LanceDB (lightweight, local file-based)
    os.environ.setdefault("VECTOR_STORE_TYPE", "lancedb")
    os.environ.setdefault("LANCEDB_DIR", "./data/lancedb")

    # LLM (use mock in tests)
    os.environ.setdefault("LLM_PROVIDER", "mock")

    # Disable external services in tests
    os.environ.setdefault("ENABLE_SEMANTIC_CACHING", "false")
    os.environ.setdefault("ENABLE_WEB_SEARCH", "false")

    setup_logging(json_output=False, log_level="DEBUG")


pytest_asyncio_mode = "auto"
