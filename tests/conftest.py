import importlib.util
import os
import sys
from pathlib import Path

import pytest

src_path = Path(__file__).parent.parent / "src"
sys.path.insert(0, str(src_path))

# Load setup_logging from our custom logging module
logging_setup_path = src_path / "logging" / "setup.py"
spec = importlib.util.spec_from_file_location("logging_setup", logging_setup_path)
if spec is None or spec.loader is None:
    msg = "Could not load logging.setup module"
    raise ImportError(msg)
logging_module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(logging_module)
setup_logging = logging_module.setup_logging

pytest_plugins = ("pytest_asyncio",)


@pytest.fixture(scope="session", autouse=True)
def _setup_test_environment() -> None:
    """Setup environment and logging for all tests."""
    os.environ.setdefault("LOG_LEVEL", "DEBUG")
    os.environ.setdefault("LOG_JSON", "false")

    setup_logging(json_output=False, log_level="DEBUG")


pytest_asyncio_mode = "auto"
