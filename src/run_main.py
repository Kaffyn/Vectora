import asyncio
import os

# Set defaults BEFORE importing
os.environ.setdefault("LOG_LEVEL", "INFO")
os.environ.setdefault("LOG_JSON", "false")
os.environ.setdefault("GOOGLE_API_KEY", "")
os.environ.setdefault("LLM_PROVIDER", "google-genai")

if __name__ == "__main__":
    # Validar que Voyage AI está configurado (obrigatório para Vectora)
    from env import validate_voyage_ai

    validate_voyage_ai()

    from main import main

    asyncio.run(main())
