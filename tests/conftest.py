import os
import sys
from pathlib import Path

import pytest

src_path = Path(__file__).parent.parent / "src"
sys.path.insert(0, str(src_path))

# Import setup_logging from log_setup module
from log_setup import setup_logging

pytest_plugins = ("pytest_asyncio",)


@pytest.fixture(scope="session", autouse=True)
def _setup_test_environment() -> None:
    """Setup environment and logging for all tests."""
    os.environ.setdefault("LOG_LEVEL", "DEBUG")
    os.environ.setdefault("LOG_JSON", "false")

    setup_logging(json_output=False, log_level="DEBUG")


pytest_asyncio_mode = "auto"
