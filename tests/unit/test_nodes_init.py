"""Testes para vectora/nodes/__init__.py (package exports)."""

from __future__ import annotations


class TestNodesModuleExports:
    """Testes para exports do módulo nodes."""

    def test_nodes_module_has_all_attribute(self) -> None:
        """Verificar que módulo define __all__."""
        import vectora.nodes as nodes_module

        assert hasattr(nodes_module, "__all__")
        assert isinstance(nodes_module.__all__, list)

    def test_nodes_module_all_contains_expected_exports(self) -> None:
        """Verificar que __all__ contém os exports esperados."""
        import vectora.nodes as nodes_module

        expected_exports = [
            "_extract_tavily_results",
            "_get_llm_with_tools",
            "_process_tavily_results",
            "call_llm",
            "handle_sub_node",
            "process_retrieval",
        ]
        for export in expected_exports:
            assert export in nodes_module.__all__

    def test_extract_tavily_results_import(self) -> None:
        """Verificar que _extract_tavily_results é importado."""
        import vectora.nodes as nodes_module

        assert hasattr(nodes_module, "_extract_tavily_results")
        assert callable(nodes_module._extract_tavily_results)

    def test_get_llm_with_tools_import(self) -> None:
        """Verificar que _get_llm_with_tools é importado."""
        import vectora.nodes as nodes_module

        assert hasattr(nodes_module, "_get_llm_with_tools")
        assert callable(nodes_module._get_llm_with_tools)

    def test_process_tavily_results_import(self) -> None:
        """Verificar que _process_tavily_results é importado."""
        import vectora.nodes as nodes_module

        assert hasattr(nodes_module, "_process_tavily_results")
        assert callable(nodes_module._process_tavily_results)

    def test_call_llm_import(self) -> None:
        """Verificar que call_llm é importado."""
        import vectora.nodes as nodes_module

        assert hasattr(nodes_module, "call_llm")
        assert callable(nodes_module.call_llm)

    def test_handle_sub_node_import(self) -> None:
        """Verificar que handle_sub_node é importado."""
        import vectora.nodes as nodes_module

        assert hasattr(nodes_module, "handle_sub_node")
        assert callable(nodes_module.handle_sub_node)

    def test_process_retrieval_import(self) -> None:
        """Verificar que process_retrieval é importado."""
        import vectora.nodes as nodes_module

        assert hasattr(nodes_module, "process_retrieval")
        assert callable(nodes_module.process_retrieval)
