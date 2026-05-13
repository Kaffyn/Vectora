"""Entry point for running the Vectora CLI."""

import asyncio
import os

# Set defaults BEFORE importing (before constants.py is loaded)
os.environ.setdefault("DB_DSN", "sqlite:///./vectora.db")
os.environ.setdefault("LOG_LEVEL", "INFO")
os.environ.setdefault("LOG_JSON", "false")
os.environ.setdefault("GOOGLE_API_KEY", "")
os.environ.setdefault("LLM_PROVIDER", "google-genai")

if __name__ == "__main__":
    from main import main

    asyncio.run(main())
