"""Testes para o módulo vectora.mcp.__init__.py (lazy imports)."""

from __future__ import annotations


class TestMcpModuleExports:
    """Testes para exports e lazy imports do módulo mcp."""

    def test_mcp_module_has_all_attribute(self) -> None:
        """Verificar que módulo define __all__."""
        import vectora.mcp as mcp_module

        assert hasattr(mcp_module, "__all__")
        assert isinstance(mcp_module.__all__, list)

    def test_mcp_module_all_contains_expected_exports(self) -> None:
        """Verificar que __all__ contém os exports esperados."""
        import vectora.mcp as mcp_module

        expected_exports = [
            "MCPClient",
            "VectoraProxy",
            "create_local_proxy",
            "create_remote_proxy",
            "run",
        ]
        for export in expected_exports:
            assert export in mcp_module.__all__

    def test_mcp_client_lazy_import(self) -> None:
        """Verificar que MCPClient é importado lazily."""
        import vectora.mcp as mcp_module

        # Acessar MCPClient dispara __getattr__
        mcp_client = mcp_module.MCPClient
        assert mcp_client is not None

    def test_vectora_proxy_lazy_import(self) -> None:
        """Verificar que VectoraProxy é importado lazily."""
        import vectora.mcp as mcp_module

        # Acessar VectoraProxy dispara __getattr__
        vectora_proxy = mcp_module.VectoraProxy
        assert vectora_proxy is not None

    def test_create_local_proxy_lazy_import(self) -> None:
        """Verificar que create_local_proxy é importado lazily."""
        import vectora.mcp as mcp_module

        create_local = mcp_module.create_local_proxy
        assert callable(create_local)

    def test_create_remote_proxy_lazy_import(self) -> None:
        """Verificar que create_remote_proxy é importado lazily."""
        import vectora.mcp as mcp_module

        create_remote = mcp_module.create_remote_proxy
        assert callable(create_remote)

    def test_run_lazy_import(self) -> None:
        """Verificar que run é importado lazily."""
        import vectora.mcp as mcp_module

        run = mcp_module.run
        assert callable(run)

    def test_getattr_raises_for_nonexistent_attribute(self) -> None:
        """Verificar que __getattr__ lança AttributeError para atributos inexistentes."""
        import pytest

        import vectora.mcp as mcp_module

        with pytest.raises(AttributeError, match="has no attribute"):
            _ = mcp_module.NONEXISTENT_ATTRIBUTE

    def test_mcp_client_import_actually_returns_class(self) -> None:
        """Verificar que MCPClient é realmente a classe MCPClient."""
        import vectora.mcp as mcp_module
        from vectora.mcp.client import MCPClient as DirectMCPClient

        lazy_client = mcp_module.MCPClient
        assert lazy_client is DirectMCPClient

    def test_vectora_proxy_import_actually_returns_class(self) -> None:
        """Verificar que VectoraProxy é realmente a classe VectoraProxy."""
        import vectora.mcp as mcp_module
        from vectora.mcp.proxy import VectoraProxy as DirectVectoraProxy

        lazy_proxy = mcp_module.VectoraProxy
        assert lazy_proxy is DirectVectoraProxy

    def test_run_import_actually_returns_function(self) -> None:
        """Verificar que run é realmente a função run."""
        import vectora.mcp as mcp_module
        from vectora.mcp.server import run as direct_run

        lazy_run = mcp_module.run
        assert lazy_run is direct_run
