import os
import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from logging.setup import setup_logging

pytest_plugins = ("pytest_asyncio",)


@pytest.fixture(scope="session", autouse=True)
def _setup_test_environment() -> None:  # noqa: ARG001
    """Setup environment and logging for all tests."""
    os.environ.setdefault("LOG_LEVEL", "DEBUG")
    os.environ.setdefault("LOG_JSON", "false")

    setup_logging(json_output=False, log_level="DEBUG")


pytest_asyncio_mode = "auto"
