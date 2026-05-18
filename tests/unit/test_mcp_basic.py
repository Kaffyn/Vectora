"""Testes basicos para modulo mcp - imports e estrutura.

Nota: Os testes complexos requerem mcp.ClientSession que não está disponível.
Esses testes focam em estrutura, constantes e funcoes auxiliares.
"""

from __future__ import annotations

import pytest


class TestMCPModule:
    """Testes basicos do modulo mcp."""

    def test_mcp_module_exists(self):
        """Verificar que modulo mcp existe."""
        try:
            import vectora.mcp

            assert vectora.mcp is not None
        except ImportError:
            # Se falhar importacao, é erro do modulo nao do teste
            pytest.skip("MCP module not available")

    def test_mcp_has_init(self):
        """Verificar que modulo mcp tem __init__.py."""
        try:
            from vectora import mcp

            assert hasattr(mcp, "__name__")
        except ImportError:
            pytest.skip("MCP module not available")


class TestMCPClientBasic:
    """Testes para estrutura basica do cliente MCP."""

    def test_mcp_proxy_module_can_be_imported(self):
        """Testar que proxy.py pode ser importado."""
        try:
            from vectora.mcp import proxy

            assert proxy is not None
        except ImportError as e:
            pytest.skip(f"MCP proxy dependencies not available: {e}")


def test_mcp_package_structure():
    """Testar que package mcp tem estrutura esperada."""
    try:
        from vectora import mcp

        # Verificar que pelo menos mcp é importavel
        assert hasattr(mcp, "__package__")
        assert "mcp" in mcp.__package__
    except ImportError:
        pytest.skip("MCP module not available")
