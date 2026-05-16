"""Chat Application Launcher.

Entry point for starting the interactive Textual chat interface.
Initializes graph, checkpointer, and message loop.
"""

import asyncio
import os


async def _run_chat_async() -> None:
    """Lógica assíncrona principal da TUI."""
    import sys
    from pathlib import Path

    # Adicionar src ao PYTHONPATH para imports internos funcionarem
    src_dir = Path(__file__).parent
    if str(src_dir) not in sys.path:
        sys.path.insert(0, str(src_dir))

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
        from chat import run_chat as start_chat_app

        await start_chat_app()


def run_chat() -> None:
    """Entry point para a TUI interativa (CLI: vectora).

    Wrapper síncrono que executa a lógica assíncrona via asyncio.run().
    """
    asyncio.run(_run_chat_async())


if __name__ == "__main__":
    run_chat()
