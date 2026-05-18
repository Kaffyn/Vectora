"""Testes para Logging Configuration (logger.py)."""

from __future__ import annotations

import json
import logging
import os
from queue import Queue
from typing import TYPE_CHECKING
from unittest.mock import MagicMock, patch

import pytest

from vectora.services.logger import (
    JSONFormatter,
    TextFormatter,
    setup_logging,
    setup_queue_handler,
)

if TYPE_CHECKING:
    from pathlib import Path


class TestJSONFormatter:
    """Testes para JSONFormatter."""

    def test_json_formatter_creates_valid_json(self) -> None:
        """Verificar que JSONFormatter cria JSON válido."""
        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="test.logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Test message",
            args=(),
            exc_info=None,
        )

        result = formatter.format(record)

        # Deve ser JSON válido
        parsed = json.loads(result)
        assert isinstance(parsed, dict)

    def test_json_formatter_includes_timestamp(self) -> None:
        """Verificar que JSONFormatter inclui timestamp."""
        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Test",
            args=(),
            exc_info=None,
        )

        result = json.loads(formatter.format(record))

        assert "timestamp" in result
        assert result["timestamp"] is not None

    def test_json_formatter_includes_level(self) -> None:
        """Verificar que JSONFormatter inclui level."""
        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.ERROR,
            pathname="test.py",
            lineno=1,
            msg="Error",
            args=(),
            exc_info=None,
        )

        result = json.loads(formatter.format(record))

        assert result["level"] == "ERROR"

    def test_json_formatter_includes_logger_name(self) -> None:
        """Verificar que JSONFormatter inclui nome do logger."""
        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="my.module",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Test",
            args=(),
            exc_info=None,
        )

        result = json.loads(formatter.format(record))

        assert result["logger"] == "my.module"

    def test_json_formatter_includes_message(self) -> None:
        """Verificar que JSONFormatter inclui mensagem."""
        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Test message",
            args=(),
            exc_info=None,
        )

        result = json.loads(formatter.format(record))

        assert result["message"] == "Test message"

    def test_json_formatter_includes_custom_fields(self) -> None:
        """Verificar que JSONFormatter inclui campos customizados."""
        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Test",
            args=(),
            exc_info=None,
        )
        record.thread_id = "thread-123"
        record.user_type = "premium"

        result = json.loads(formatter.format(record))

        assert result["thread_id"] == "thread-123"
        assert result["user_type"] == "premium"


class TestTextFormatter:
    """Testes para TextFormatter."""

    def test_text_formatter_creates_readable_text(self) -> None:
        """Verificar que TextFormatter cria texto legível."""
        formatter = TextFormatter(fmt="%(message)s")
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Test message",
            args=(),
            exc_info=None,
        )

        result = formatter.format(record)

        assert "Test message" in result

    def test_text_formatter_includes_level(self) -> None:
        """Verificar que TextFormatter inclui level."""
        formatter = TextFormatter(fmt="[%(levelname)s] %(message)s")
        record = logging.LogRecord(
            name="test",
            level=logging.WARNING,
            pathname="test.py",
            lineno=1,
            msg="Warning",
            args=(),
            exc_info=None,
        )

        result = formatter.format(record)

        assert "WARNING" in result

    def test_text_formatter_includes_thread_id_if_present(self) -> None:
        """Verificar que TextFormatter inclui thread_id se presente."""
        formatter = TextFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Test",
            args=(),
            exc_info=None,
        )
        record.thread_id = "42"

        result = formatter.format(record)

        assert "thread=" in result
        assert "42" in result


class TestSetupLogging:
    """Testes para setup_logging()."""

    def test_setup_logging_configures_root_logger(self) -> None:
        """Verificar que setup_logging() configura root logger."""
        # Limpar handlers para teste
        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        setup_logging(json_output=False, log_level="INFO")

        assert len(root.handlers) > 0

        # Limpar após teste
        for handler in root.handlers[:]:
            root.removeHandler(handler)

    def test_setup_logging_respects_log_level(self) -> None:
        """Verificar que setup_logging() respeita log level."""
        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        setup_logging(json_output=False, log_level="WARNING")

        assert root.level == logging.WARNING

        for handler in root.handlers[:]:
            root.removeHandler(handler)

    def test_setup_logging_idempotent(self) -> None:
        """Verificar que setup_logging() é idempotente."""
        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        setup_logging(json_output=False, log_level="INFO")
        initial_count = len(root.handlers)

        # Chamar novamente
        setup_logging(json_output=False, log_level="INFO")

        # Não deve adicionar mais handlers
        assert len(root.handlers) == initial_count

        for handler in root.handlers[:]:
            root.removeHandler(handler)

    def test_setup_logging_reads_env_variables(self) -> None:
        """Verificar que setup_logging() lê variáveis de ambiente."""
        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        with patch.dict(os.environ, {"LOG_LEVEL": "ERROR"}):
            setup_logging(json_output=False)

        assert root.level == logging.ERROR

        for handler in root.handlers[:]:
            root.removeHandler(handler)

    def test_setup_logging_creates_log_directory(self, tmp_path: Path) -> None:
        """Verificar que setup_logging() cria diretório de logs."""
        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        log_file = str(tmp_path / "test.jsonl")
        setup_logging(json_output=True, log_level="INFO", log_file=log_file)

        # Diretório deve ter sido criado
        assert tmp_path.exists()

        for handler in root.handlers[:]:
            root.removeHandler(handler)

    def test_setup_logging_with_json_output(self, tmp_path: Path) -> None:
        """Verificar que setup_logging() cria handler JSON."""
        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        log_file = str(tmp_path / "test.jsonl")
        setup_logging(json_output=True, log_level="INFO", log_file=log_file)

        # Deve ter um handler FileHandler
        has_file_handler = any(
            isinstance(h, logging.FileHandler) for h in root.handlers
        )
        assert has_file_handler

        for handler in root.handlers[:]:
            root.removeHandler(handler)

    def test_setup_logging_without_json_output(self) -> None:
        """Verificar que setup_logging() sem JSON não cria FileHandler."""
        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        setup_logging(json_output=False, log_level="INFO")

        # Não deve ter FileHandler
        has_file_handler = any(
            isinstance(h, logging.FileHandler) for h in root.handlers
        )
        assert not has_file_handler

        for handler in root.handlers[:]:
            root.removeHandler(handler)


class TestSetupQueueHandler:
    """Testes para setup_queue_handler()."""

    def test_setup_queue_handler_adds_handler(self) -> None:
        """Verificar que setup_queue_handler() adiciona handler à queue."""
        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        queue: Queue = Queue()  # type: ignore[type-arg]
        setup_queue_handler(queue)

        # Deve ter adicionado um QueueHandler
        has_queue_handler = any(
            isinstance(h, logging.handlers.QueueHandler) for h in root.handlers
        )
        assert has_queue_handler

        for handler in root.handlers[:]:
            root.removeHandler(handler)

    def test_setup_queue_handler_queue_receives_records(self) -> None:
        """Verificar que records são adicionados à queue."""
        import logging.handlers

        root = logging.getLogger()
        for handler in root.handlers[:]:
            root.removeHandler(handler)

        queue: Queue = Queue()  # type: ignore[type-arg]
        setup_queue_handler(queue)

        # Log uma mensagem
        logger = logging.getLogger("test")
        logger.info("Test message")

        # Queue deve ter um item
        assert not queue.empty()

        for handler in root.handlers[:]:
            root.removeHandler(handler)
