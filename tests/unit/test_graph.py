"""Testes para graph.py (LangGraph Construction)."""

from __future__ import annotations

from unittest.mock import MagicMock, patch

from vectora.graph import build_graph


class TestBuildGraph:
    """Testes para funcao build_graph()."""

    def test_build_graph_returns_compiled_state_graph(self) -> None:
        """Verificar que build_graph retorna CompiledStateGraph."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_compiled = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = mock_compiled

            result = build_graph(checkpointer=checkpointer)

            assert result is mock_compiled

    def test_build_graph_creates_state_graph_with_state(self) -> None:
        """Verificar que StateGraph e criado com State schema."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            build_graph(checkpointer=checkpointer)

            assert mock_state_graph.called

    def test_build_graph_adds_all_required_nodes(self) -> None:
        """Verificar que todos os nos sao adicionados ao grafo."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            build_graph(checkpointer=checkpointer)

            assert mock_builder.add_node.call_count == 4

    def test_build_graph_adds_call_llm_node(self) -> None:
        """Verificar que no call_llm e adicionado."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            with patch("vectora.graph.call_llm_debug"):
                build_graph(checkpointer=checkpointer)

            node_names = [call[0][0] for call in mock_builder.add_node.call_args_list]
            assert "call_llm" in node_names

    def test_build_graph_adds_tools_node(self) -> None:
        """Verificar que no tools e adicionado."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            build_graph(checkpointer=checkpointer)

            node_names = [call[0][0] for call in mock_builder.add_node.call_args_list]
            assert "tools" in node_names

    def test_build_graph_adds_process_retrieval_node(self) -> None:
        """Verificar que no process_retrieval e adicionado."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            with patch("vectora.graph.process_retrieval"):
                build_graph(checkpointer=checkpointer)

            node_names = [call[0][0] for call in mock_builder.add_node.call_args_list]
            assert "process_retrieval" in node_names

    def test_build_graph_adds_sub_node(self) -> None:
        """Verificar que no sub_node e adicionado."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            with patch("vectora.graph.handle_sub_node"):
                build_graph(checkpointer=checkpointer)

            node_names = [call[0][0] for call in mock_builder.add_node.call_args_list]
            assert "sub_node" in node_names

    def test_build_graph_adds_edges(self) -> None:
        """Verificar que arestas sao adicionadas ao grafo."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            build_graph(checkpointer=checkpointer)

            assert mock_builder.add_edge.called

    def test_build_graph_adds_conditional_edges(self) -> None:
        """Verificar que arestas condicionais sao adicionadas."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            build_graph(checkpointer=checkpointer)

            assert mock_builder.add_conditional_edges.called

    def test_build_graph_compiles_graph_with_checkpointer(self) -> None:
        """Verificar que grafo e compilado com checkpointer."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            build_graph(checkpointer=checkpointer)

            mock_builder.compile.assert_called_once_with(checkpointer=checkpointer)

    def test_build_graph_logs_info(self) -> None:
        """Verificar que logs informativos sao gerados."""
        checkpointer = MagicMock()

        with patch("vectora.graph.StateGraph") as mock_state_graph:
            mock_builder = MagicMock()
            mock_state_graph.return_value = mock_builder
            mock_builder.compile.return_value = MagicMock()

            with patch("vectora.graph.logger") as mock_logger:
                build_graph(checkpointer=checkpointer)

                assert mock_logger.info.called
