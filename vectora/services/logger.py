"""Logging Configuration for Vectora.

Structured logging with JSON and text formatters, file rotation,
and QueueHandler support for real-time UI streaming in debug mode.

This module replaces the top-level log_setup.py.
"""

import json
import logging
import logging.handlers
import os
from datetime import UTC, datetime
from pathlib import Path
from queue import Queue
from typing import Any


class JSONFormatter(logging.Formatter):
    """Format logs as structured JSON for production observability."""

    def format(self, record: logging.LogRecord) -> str:
        log_obj: dict[str, Any] = {
            "timestamp": datetime.now(UTC).isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
        }

        if record.exc_info:
            log_obj["exception"] = self.formatException(record.exc_info)

        # Optional structured fields
        for field in ("thread_id", "user_type", "model"):
            if hasattr(record, field):
                log_obj[field] = getattr(record, field)

        for field in ("retrieval_source", "reranking_applied", "routing_decision"):
            if hasattr(record, field):
                log_obj[field] = getattr(record, field)

        return json.dumps(log_obj, ensure_ascii=False)


class TextFormatter(logging.Formatter):
    """Format logs as readable text for development."""

    def format(self, record: logging.LogRecord) -> str:
        prefix = f"[{record.levelname:8}] {record.name:20} | "

        if hasattr(record, "thread_id"):
            prefix += f"thread={record.thread_id:>3} | "

        msg = record.getMessage()
        if record.exc_info:
            msg += "\n" + self.formatException(record.exc_info)

        return prefix + msg


def setup_logging(
    json_output: bool | None = None,
    log_level: str | None = None,
    log_file: str | None = None,
) -> None:
    """Configure application logging with console and optional JSON file output.

    Sets up dual handlers: human-readable console output and structured JSON
    file output. External library noise is suppressed in quiet mode.

    Args:
        json_output: Enable JSON file output. Defaults to LOG_JSON env var.
        log_level: Log level (DEBUG/INFO/WARNING/ERROR). Defaults to LOG_LEVEL env var.
        log_file: Path to JSON log file. Defaults to LOG_FILE env var.

    Environment variables:
        LOG_LEVEL: Logging level (default: INFO)
        LOG_JSON: Enable JSON output (default: false)
        LOG_FILE: JSON log file path (default: ~/.vectora/logs/vectora.jsonl)
        QUIET_MODE: Suppress external library logs (default: true)
    """
    if json_output is None:
        json_output = os.getenv("LOG_JSON", "false").lower() == "true"

    if log_level is None:
        log_level = os.getenv("LOG_LEVEL", "INFO")

    if log_file is None:
        default_log = Path.home() / ".vectora" / "logs" / "vectora.jsonl"
        log_file = os.getenv("LOG_FILE", str(default_log))

    root_logger = logging.getLogger()
    root_logger.setLevel(getattr(logging, log_level))

    # Idempotent — skip if handlers already configured
    if root_logger.handlers:
        return

    text_formatter = TextFormatter(
        fmt="%(asctime)s | %(name)s | %(levelname)s | %(message)s",
        datefmt="%H:%M:%S",
    )

    console_handler = logging.StreamHandler()
    console_handler.setFormatter(text_formatter)
    console_handler.setLevel(getattr(logging, log_level))
    root_logger.addHandler(console_handler)

    # Suppress external library verbosity in quiet mode
    if os.getenv("QUIET_MODE", "true").lower() == "true":
        _noisy_loggers = [
            "langchain",
            "langchain_core",
            "langchain_google_genai",
            "langgraph",
            "google",
            "google.genai",
            "google.genai._api_client",
            "google_genai",
            "google_genai._api_client",
            "httpx",
            "urllib3",
            "requests",
            "asyncio",
            "aiosqlite",
            "httpcore",
            # Voyage AI SDK — evita output espúrio durante embed_query()
            "voyageai",
            "voyageai.client",
        ]
        for name in _noisy_loggers:
            logging.getLogger(name).setLevel(logging.CRITICAL)

    if json_output:
        log_path = Path(log_file)
        log_path.parent.mkdir(parents=True, exist_ok=True)

        json_handler = logging.FileHandler(log_path, mode="a", encoding="utf-8")
        json_handler.setFormatter(JSONFormatter())
        json_handler.setLevel(getattr(logging, log_level))
        root_logger.addHandler(json_handler)


def setup_queue_handler(log_queue: Queue) -> None:  # type: ignore[type-arg]
    """Add a QueueHandler for real-time log streaming to the UI in debug mode.

    Args:
        log_queue: Queue instance to receive formatted log records
    """
    queue_handler = logging.handlers.QueueHandler(log_queue)
    queue_handler.setFormatter(
        TextFormatter(
            fmt="%(asctime)s | %(name)s | %(levelname)s | %(message)s",
            datefmt="%H:%M:%S",
        )
    )
    logging.getLogger().addHandler(queue_handler)
