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
from vectora.services.log_setup import setup_logging

setup_logging()

logger = logging.getLogger(__name__)


def _configure_langsmith(settings: "Settings") -> list:  # noqa: F821
    """Configura LangSmith para dois canais independentes.

    Retorna lista de callbacks a adicionar ao RunnableConfig do LangGraph.

    ─────────────────────────────────────────────────────────────────
    Canal A — Usuário (LANGSMITH_API_KEY / LANGSMITH_TRACING=true)
    ─────────────────────────────────────────────────────────────────
    Configurado via variáveis de ambiente no .env local do usuário.
    LangChain/LangGraph detecta automaticamente — zero código extra.
    Quando ativo, o usuário vê seus traces completos (inputs, outputs).
    Independente do consentimento de telemetria do Vectora.

    ─────────────────────────────────────────────────────────────────
    Canal B — Desenvolvedor (VECTORA_LANGSMITH_API_KEY + consentimento)
    ─────────────────────────────────────────────────────────────────
    Só ativo quando:
      1. vectora_langsmith_api_key presente (injetada em build via CI/CD)
      2. Usuário consentiu via /privacidade enable (LGPD Art. 7)

    Usa build_sanitized_tracer() que cria um langsmith.Client com
    hide_inputs=True e hide_outputs=True — hard gate por código, não
    por variável de ambiente. O tracer é adicionado como callback
    explícito no RunnableConfig e NÃO interfere com o Canal A.

    Args:
        settings: Settings inicializadas com campos vectora_langsmith_* e
                  langsmith_*.

    Returns:
        Lista de BaseCallbackHandler. Vazia se telemetria do dev inativa.
    """
    from vectora.services.consent import get_consent_manager

    callbacks: list = []

    # ── Canal A: usuário — configurado via env vars, LangChain cuida sozinho ──
    # Se LANGSMITH_TRACING=true + LANGSMITH_API_KEY estão presentes (via .env
    # local do usuário ou override de dev), o LangChain já ativou o tracer
    # global automaticamente ao importar. Não precisamos fazer nada.
    if settings.langsmith_tracing and settings.langsmith_api_key:
        logger.info(
            "langsmith_user_tracing_active",
            extra={"project": settings.langsmith_project},
        )
    else:
        # Garante que env vars não deixem rastro se não configurado
        os.environ.setdefault("LANGSMITH_TRACING", "false")

    # ── Canal B: desenvolvedor — tracer sanitizado como callback explícito ───
    consent = get_consent_manager()
    if consent.is_consented() and settings.vectora_langsmith_api_key:
        try:
            from vectora.services.privacy import build_sanitized_tracer

            dev_tracer = build_sanitized_tracer(
                api_key=settings.vectora_langsmith_api_key,
                project_name=settings.vectora_langsmith_project,
                endpoint=settings.vectora_langsmith_endpoint,
            )
            callbacks.append(dev_tracer)
            logger.info(
                "langsmith_dev_telemetry_active",
                extra={"project": settings.vectora_langsmith_project},
            )
        except Exception:
            # Telemetria nunca deve bloquear o app
            logger.warning("langsmith_dev_tracer_build_failed", exc_info=True)

    return callbacks


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
        from vectora.config.settings import Settings

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
        # STEP 1b: Configure LangSmith (two independent channels)
        # ====================================================================
        # Returns list of callbacks for the developer's sanitized channel.
        # The user's own channel (LANGSMITH_API_KEY) is auto-detected by
        # LangChain from env vars — no explicit handling needed.
        telemetry_callbacks = _configure_langsmith(settings)

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
            from vectora.services.setup_wizard import run_setup_sync

            run_setup_sync()
            # After setup, reload settings
            settings = Settings()
            agent = AgentManager(settings)
            await agent.initialize()

        # ====================================================================
        # STEP 4: Run CLI Dashboard
        # ====================================================================
        from vectora.ui.chat import run_chat

        await run_chat(
            agent=agent,
            settings=settings,
            telemetry_callbacks=telemetry_callbacks,
        )

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
