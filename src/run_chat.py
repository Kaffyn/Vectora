"""Chat Application Launcher.

Entry point for starting the interactive Textual chat interface.
Initializes graph, checkpointer, and message loop.
"""

import asyncio
import os

if __name__ == "__main__":
    os.environ.setdefault("LOG_LEVEL", "INFO")
    os.environ.setdefault("LOG_JSON", "false")

    # Validar que Voyage AI está configurado (obrigatório para Vectora)
    from env import validate_voyage_ai

    validate_voyage_ai()

    from config import Config

    config = Config.instance()
    if not config.get_llm_provider():
        from setup_wizard import run_setup

        run_setup()
    else:
        from chat import run_chat

        asyncio.run(run_chat())
