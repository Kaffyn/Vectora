"""Vectora initialization and setup utilities.

Handles automatic directory creation, environment setup, and verification.
Ensures all required directories and configurations exist before runtime.
"""

import logging
from pathlib import Path

logger = logging.getLogger(__name__)


def initialize_vectora_home() -> Path:
    """Initialize Vectora home directory structure.

    Creates the following directories if they don't exist:
    - ~/.vectora/
    - ~/.vectora/data/ (for SQLite database)
    - ~/.vectora/logs/ (for application logs)

    Returns:
        Path to the Vectora home directory (~/.vectora)

    Raises:
        OSError: If directories cannot be created due to permissions
    """
    vectora_home = Path.home() / ".vectora"

    # Create main directory
    vectora_home.mkdir(parents=True, exist_ok=True)
    logger.info(f"Vectora home initialized: {vectora_home}")

    # Create subdirectories
    subdirs = ["data", "logs"]
    for subdir in subdirs:
        subdir_path = vectora_home / subdir
        subdir_path.mkdir(parents=True, exist_ok=True)
        logger.info(f"Created directory: {subdir_path}")

    return vectora_home


def verify_vectora_setup() -> bool:
    """Verify that Vectora home directory is properly initialized.

    Checks for the existence of:
    - ~/.vectora/
    - ~/.vectora/data/
    - ~/.vectora/logs/

    Returns:
        True if all required directories exist, False otherwise
    """
    vectora_home = Path.home() / ".vectora"

    required_dirs = [
        vectora_home,
        vectora_home / "data",
        vectora_home / "logs",
    ]

    missing_dirs = [d for d in required_dirs if not d.exists()]

    if missing_dirs:
        logger.warning(f"Missing Vectora directories: {[str(d) for d in missing_dirs]}")
        return False

    logger.info("Vectora setup verified successfully")
    return True


def ensure_vectora_initialized() -> None:
    """Ensure Vectora is properly initialized.

    Automatically initializes Vectora home if not already done.
    This should be called at application startup.
    """
    if not verify_vectora_setup():
        logger.info("Initializing Vectora home directory...")
        initialize_vectora_home()
        verify_vectora_setup()
