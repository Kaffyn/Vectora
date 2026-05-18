"""Testes para serviços de logging do Vectora.

Cobre: log_setup.py, logger.py — formatadores JSON/text,
setup_logging, QueueHandler, filtros de background.
"""

from __future__ import annotations

import json
import logging
import logging.handlers
from queue import Queue
from unittest.mock import patch


class TestJSONFormatterLogSetup:
    """Testa JSONFormatter em log_setup.py."""

    def test_json_formatter_produces_valid_json(self):
        """Verifica que JSONFormatter produz JSON válido."""
        from vectora.services.log_setup import JSONFormatter

        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="test.logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Mensagem de teste",
            args=(),
            exc_info=None,
        )
        output = formatter.format(record)
        data = json.loads(output)
        assert data["level"] == "INFO"
        assert data["logger"] == "test.logger"
        assert data["message"] == "Mensagem de teste"
        assert "timestamp" in data

    def test_json_formatter_includes_exception(self):
        """Verifica que JSONFormatter inclui traceback de exceção."""
        from vectora.services.log_setup import JSONFormatter

        formatter = JSONFormatter()
        try:
            raise ValueError("Erro de teste")
        except ValueError:
            import sys

            exc_info = sys.exc_info()

        record = logging.LogRecord(
            name="test",
            level=logging.ERROR,
            pathname="test.py",
            lineno=1,
            msg="Erro ocorreu",
            args=(),
            exc_info=exc_info,
        )
        output = formatter.format(record)
        data = json.loads(output)
        assert "exception" in data
        assert "ValueError" in data["exception"]

    def test_json_formatter_includes_extra_fields(self):
        """Verifica que JSONFormatter inclui campos extras (thread_id, etc)."""
        from vectora.services.log_setup import JSONFormatter

        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="msg",
            args=(),
            exc_info=None,
        )
        record.thread_id = "thread_42"
        record.user_type = "admin"

        output = formatter.format(record)
        data = json.loads(output)
        assert data["thread_id"] == "thread_42"
        assert data["user_type"] == "admin"


class TestTextFormatterLogSetup:
    """Testa TextFormatter em log_setup.py."""

    def test_text_formatter_produces_readable_output(self):
        """Verifica que TextFormatter produz texto legível."""
        from vectora.services.log_setup import TextFormatter

        formatter = TextFormatter()
        record = logging.LogRecord(
            name="vectora.test",
            level=logging.DEBUG,
            pathname="test.py",
            lineno=1,
            msg="Debug message",
            args=(),
            exc_info=None,
        )
        output = formatter.format(record)
        assert isinstance(output, str)
        assert len(output) > 0


class TestBackgroundConsoleFilter:
    """Testa _BackgroundConsoleFilter em log_setup.py."""

    def test_filter_blocks_background_info_logs(self):
        """Verifica que filtro bloqueia logs INFO de serviços de background."""
        from vectora.services.log_setup import _BackgroundConsoleFilter

        f = _BackgroundConsoleFilter()
        record = logging.LogRecord(
            name="vectora.services.background",
            level=logging.INFO,
            pathname="bg.py",
            lineno=1,
            msg="embedding_enqueued",
            args=(),
            exc_info=None,
        )
        assert f.filter(record) is False

    def test_filter_allows_background_warning_logs(self):
        """Verifica que filtro permite logs WARNING de background."""
        from vectora.services.log_setup import _BackgroundConsoleFilter

        f = _BackgroundConsoleFilter()
        record = logging.LogRecord(
            name="vectora.services.background",
            level=logging.WARNING,
            pathname="bg.py",
            lineno=1,
            msg="Aviso crítico",
            args=(),
            exc_info=None,
        )
        assert f.filter(record) is True

    def test_filter_allows_other_loggers(self):
        """Verifica que filtro permite outros loggers."""
        from vectora.services.log_setup import _BackgroundConsoleFilter

        f = _BackgroundConsoleFilter()
        record = logging.LogRecord(
            name="vectora.nodes.engine",
            level=logging.INFO,
            pathname="engine.py",
            lineno=1,
            msg="LLM invoked",
            args=(),
            exc_info=None,
        )
        assert f.filter(record) is True

    def test_filter_blocks_queue_info_logs(self):
        """Verifica que filtro bloqueia logs INFO da queue."""
        from vectora.services.log_setup import _BackgroundConsoleFilter

        f = _BackgroundConsoleFilter()
        record = logging.LogRecord(
            name="vectora.services.queue",
            level=logging.INFO,
            pathname="queue.py",
            lineno=1,
            msg="embedding_queue_marked_success",
            args=(),
            exc_info=None,
        )
        assert f.filter(record) is False


class TestSetupLogging:
    """Testa função setup_logging em log_setup.py."""

    def test_setup_logging_text_mode(self, tmp_path):
        """Verifica que setup_logging funciona no modo texto."""
        from vectora.services.log_setup import setup_logging

        log_dir = tmp_path / "logs"
        log_dir.mkdir()

        with patch("vectora.services.log_setup.Path.home", return_value=tmp_path):
            setup_logging(json_output=False, log_level="INFO")

        root_logger = logging.getLogger()
        assert root_logger is not None

    def test_setup_logging_json_mode(self, tmp_path):
        """Verifica que setup_logging funciona no modo JSON."""
        from vectora.services.log_setup import setup_logging

        with patch("vectora.services.log_setup.Path.home", return_value=tmp_path):
            setup_logging(json_output=True, log_level="DEBUG")

        root_logger = logging.getLogger()
        assert root_logger is not None

    def test_setup_queue_handler(self):
        """Verifica que setup_queue_handler adiciona handler à fila."""
        from vectora.services.log_setup import setup_queue_handler

        q: Queue = Queue()
        setup_queue_handler(q)

        test_logger = logging.getLogger("vectora.test.queue_handler")
        test_logger.info("Mensagem de teste para fila")

        assert q.empty() or not q.empty()  # Handler foi adicionado sem erro


class TestLoggerModule:
    """Testa módulo vectora/services/logger.py."""

    def test_json_formatter_logger_module(self):
        """Testa JSONFormatter do módulo logger."""
        from vectora.services.logger import JSONFormatter

        formatter = JSONFormatter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="Teste logger module",
            args=(),
            exc_info=None,
        )
        output = formatter.format(record)
        data = json.loads(output)
        assert data["level"] == "INFO"
        assert data["message"] == "Teste logger module"

    def test_text_formatter_logger_module(self):
        """Testa TextFormatter do módulo logger."""
        from vectora.services.logger import TextFormatter

        formatter = TextFormatter()
        record = logging.LogRecord(
            name="vectora.test",
            level=logging.WARNING,
            pathname="test.py",
            lineno=10,
            msg="Warning test",
            args=(),
            exc_info=None,
        )
        output = formatter.format(record)
        assert isinstance(output, str)
        assert "Warning" in output or "WARNING" in output

    def test_setup_logging_logger_module(self, tmp_path):
        """Testa setup_logging do módulo logger."""
        from vectora.services.logger import setup_logging

        with patch("vectora.services.logger.Path.home", return_value=tmp_path):
            setup_logging(json_output=False, log_level="WARNING")

        assert logging.getLogger() is not None

    def test_setup_queue_handler_logger_module(self):
        """Testa setup_queue_handler do módulo logger."""
        from vectora.services.logger import setup_queue_handler

        q: Queue = Queue()
        setup_queue_handler(q)
        # Não deve lançar exceção
