"""Main entry point for Vectora CLI.

Refactored to use Settings + AgentManager (Dependency Injection).
No scattered os.getenv() calls. No complex context objects in configurable.

This is the "brain" that initializes everything with proper dependency injection:
1. Load Settings (fail-fast on validation errors)
2. Initialize AgentManager (which injects all services)
3. Run the CLI (thin UI layer)

Simple, testable, maintainable.
"""

import asyncio
import logging
import os
import sys
from pathlib import Path

# Configure UTF-8 on Windows before any output
os.environ.setdefault("PYTHONIOENCODING", "utf-8")

# Add vectora to path for imports
vectora_dir = Path(__file__).parent
if str(vectora_dir) not in sys.path:
    sys.path.insert(0, str(vectora_dir))

# Setup logging early (before importing rest of app)
from vectora.log_setup import setup_logging

setup_logging()

logger = logging.getLogger(__name__)


async def main() -> None:
    """Main entry point for Vectora CLI.

    Initialization sequence:
    1. Load Settings (with validation)
    2. Initialize AgentManager
    3. Run CLI dashboard
    """
    try:
        logger.info("Starting Vectora CLI...")

        # ====================================================================
        # STEP 1: Load Settings (Single Source of Truth)
        # ====================================================================
        from vectora.settings import Settings

        try:
            settings = Settings()
            logger.info(
                "Settings loaded successfully",
                extra={
                    "provider": settings.get_llm_provider(),
                    "version": settings.version,
                },
            )
        except Exception as e:
            logger.critical(
                "Failed to initialize Settings. Configuration error:",
                extra={"error": str(e)},
            )
            print(
                f"\n❌ Configuration Error:\n{e}\n\n"
                "Please check your environment variables or ~/.vectora/.env file."
            )
            sys.exit(1)

        # ====================================================================
        # STEP 2: Initialize AgentManager (Orchestrator)
        # ====================================================================
        from vectora.agent import AgentManager

        try:
            agent = AgentManager(settings)
            await agent.initialize()
            logger.info("AgentManager initialized successfully")
        except Exception as e:
            logger.critical(
                "Failed to initialize AgentManager",
                extra={"error": str(e)},
            )
            print(f"\n❌ Initialization Error:\n{e}")
            sys.exit(1)

        # ====================================================================
        # STEP 3: Validate LLM Configuration
        # ====================================================================
        available_providers = settings.get_available_providers()
        if not available_providers:
            logger.warning("No LLM providers configured. Running setup wizard...")
            from vectora.setup_wizard import run_setup_sync

            run_setup_sync()
            # After setup, reload settings
            settings = Settings()
            agent = AgentManager(settings)
            await agent.initialize()

        # ====================================================================
        # STEP 4: Run CLI Dashboard
        # ====================================================================
        from vectora.chat import run_chat

        await run_chat(agent=agent, settings=settings)

    except KeyboardInterrupt:
        logger.info("Chat interrupted by user")
        print("\n\nGoodbye! 👋")
        sys.exit(0)
    except Exception as e:
        logger.exception("Unexpected error in main loop")
        print(f"\n❌ Unexpected Error:\n{e}")
        sys.exit(1)
    finally:
        logger.info("Vectora CLI shutdown")


def run() -> None:
    """Synchronous entry point (called by `vectora` CLI command).

    This wrapper converts the async main() to sync for the CLI.
    """
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n\nGoodbye! 👋")
        sys.exit(0)


if __name__ == "__main__":
    run()
