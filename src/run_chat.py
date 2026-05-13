"""Entry point for running the Vectora chat TUI."""

import os

if __name__ == "__main__":
    os.environ.setdefault("DB_DSN", "sqlite:///./vectora.db")
    os.environ.setdefault("LOG_LEVEL", "INFO")
    os.environ.setdefault("LOG_JSON", "false")

    from chat import run_chat

    run_chat()
