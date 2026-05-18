"""Testes para vectora/services/telemetry.py"""

from __future__ import annotations


def test_telemetry_module_imports():
    """Verificar que telemetry pode ser importado."""
    import vectora.services.telemetry

    assert vectora.services.telemetry is not None


def test_telemetry_has_logger():
    """Verificar que telemetry tem logger."""
    import vectora.services.telemetry as tel

    assert hasattr(tel, "logger")


def test_telemetry_imports_expected_modules():
    """Verificar imports esperados."""
    import vectora.services.telemetry as tel

    # Deve ter imports basicos
    assert hasattr(tel, "logging")
