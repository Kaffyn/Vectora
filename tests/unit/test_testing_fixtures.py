"""Testes para vectora/testing/fixtures.py"""

from __future__ import annotations


def test_fixtures_module_imports():
    """Verificar que fixtures pode ser importado."""
    from vectora.testing import fixtures

    assert fixtures is not None


def test_fixtures_module_has_content():
    """Verificar que fixtures tem conteudo."""
    import inspect

    from vectora.testing import fixtures

    members = dict(inspect.getmembers(fixtures))
    assert len(members) > 0
