"""Logging Configuration and Initialization.

Sets up structured JSON logging with correlation IDs for production observability.
Supports multiple log levels, file rotation, and console output.
Includes QueueHandler for real-time UI log streaming in Debug Mode.
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
    """Format logs as JSON for structured logging."""

    def format(self: "JSONFormatter", record: logging.LogRecord) -> str:
        log_obj: dict[str, Any] = {
            "timestamp": datetime.now(UTC).isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
        }

        if record.exc_info:
            log_obj["exception"] = self.formatException(record.exc_info)

        if hasattr(record, "thread_id"):
            log_obj["thread_id"] = record.thread_id
        if hasattr(record, "user_type"):
            log_obj["user_type"] = record.user_type
        if hasattr(record, "model"):
            log_obj["model"] = record.model

        if hasattr(record, "retrieval_source"):
            log_obj["retrieval_source"] = record.retrieval_source
        if hasattr(record, "reranking_applied"):
            log_obj["reranking_applied"] = record.reranking_applied
        if hasattr(record, "routing_decision"):
            log_obj["routing_decision"] = record.routing_decision

        return json.dumps(log_obj, ensure_ascii=False)


class TextFormatter(logging.Formatter):
    """Format logs as readable text for development."""

    def format(self: "TextFormatter", record: logging.LogRecord) -> str:
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
    """
    Configure logging with dual outputs: text (console) + JSON (file).

    Args:
        json_output: Enable JSON output. If None, reads LOG_JSON env var.
        log_level: Logging level (DEBUG, INFO, WARNING, etc). Defaults to INFO.
        log_file: Path to JSON log file. If None, uses logs/vectora.jsonl

    Environment Variables:
        LOG_LEVEL: Set logging level (default: INFO)
        LOG_JSON: Set to "true" to enable JSON output (default: false)
        LOG_FILE: Path to JSON log file
    """
    if json_output is None:
        json_output = os.getenv("LOG_JSON", "false").lower() == "true"

    if log_level is None:
        log_level = os.getenv("LOG_LEVEL", "INFO")

    if log_file is None:
        log_file = os.getenv("LOG_FILE", "logs/vectora.jsonl")

    root_logger = logging.getLogger()
    root_logger.setLevel(getattr(logging, log_level))

    if root_logger.handlers:
        return

    formatter_text = TextFormatter(
        fmt="%(asctime)s | %(name)s | %(levelname)s | %(message)s",
        datefmt="%H:%M:%S",
    )

    handler_console = logging.StreamHandler()
    handler_console.setFormatter(formatter_text)
    handler_console.setLevel(getattr(logging, log_level))
    root_logger.addHandler(handler_console)

    if json_output:
        log_file_path = Path(log_file)
        log_file_path.parent.mkdir(parents=True, exist_ok=True)

        formatter_json = JSONFormatter()
        handler_file = logging.FileHandler(log_file_path, mode="a", encoding="utf-8")
        handler_file.setFormatter(formatter_json)
        handler_file.setLevel(getattr(logging, log_level))
        root_logger.addHandler(handler_file)


def setup_queue_handler(log_queue: Queue) -> None:
    """
    Add a QueueHandler to the root logger for Debug Mode UI streaming.

    This allows the UI to consume logs in real-time via the queue.
    Each log record is formatted as {level, logger, message, timestamp}.

    Args:
        log_queue: A queue.Queue instance to receive log records
    """
    root_logger = logging.getLogger()

    handler_queue = logging.handlers.QueueHandler(log_queue)
    formatter_text = TextFormatter(
        fmt="%(asctime)s | %(name)s | %(levelname)s | %(message)s",
        datefmt="%H:%M:%S",
    )
    handler_queue.setFormatter(formatter_text)
    root_logger.addHandler(handler_queue)
