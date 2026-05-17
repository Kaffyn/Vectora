"""Chat Application Launcher.

Entry point for starting the interactive chat interface.
Initializes directories, loads configuration, and starts the chat loop.
"""

import asyncio
import os


async def _run_chat_async() -> None:
    """Main async logic for chat initialization and execution."""
    import sys
    from pathlib import Path

    from dotenv import load_dotenv

    # Ensure Vectora home directory structure is initialized
    from vectora.initialization import ensure_vectora_initialized

    ensure_vectora_initialized()

    # Load .env automatically before everything
    # Search in: .env (local), ~/.vectora/.env (global)
    load_dotenv()  # Load .env from CWD
    load_dotenv(Path.home() / ".vectora" / ".env")  # Load global config

    # Adicionar vectora ao PYTHONPATH para imports internos funcionarem
    vectora_dir = Path(__file__).parent
    if str(vectora_dir) not in sys.path:
        sys.path.insert(0, str(vectora_dir))

    os.environ.setdefault("LOG_LEVEL", "INFO")
    os.environ.setdefault("LOG_JSON", "false")

    # Validar que Voyage AI está configurado (obrigatório para Vectora)
    from vectora.services.env import validate_voyage_ai

    validate_voyage_ai()

    from vectora.settings import settings

    if not settings.get_llm_provider():
        from vectora.setup_wizard import run_setup_sync

        run_setup_sync()
    else:
        from vectora.chat import run_chat as start_chat_app

        await start_chat_app()


def run_chat() -> None:
    """Entry point para a TUI interativa (CLI: vectora).

    Wrapper síncrono que executa a lógica assíncrona via asyncio.run().
    """
    # Force UTF-8 encoding on Windows for proper console output
    os.environ.setdefault("PYTHONIOENCODING", "utf-8")
    asyncio.run(_run_chat_async())


if __name__ == "__main__":
    run_chat()
