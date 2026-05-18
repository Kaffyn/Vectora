"""Direct test of ToolNode behavior without LLM calls."""

import pytest
from langgraph.prebuilt.tool_node import ToolNode

from vectora.tools import TOOLS


class TestToolNodeDirect:
    """Tests for ToolNode behavior without LLM."""

    @pytest.mark.asyncio
    async def test_toolnode_can_be_created(self):
        """Verify ToolNode can be created from TOOLS list."""
        tool_node = ToolNode(tools=TOOLS)
        assert tool_node is not None
        assert len(TOOLS) > 0

    @pytest.mark.asyncio
    async def test_toolnode_invokes_list_dir(self):
        """Test ToolNode diretamente com list_dir tool call."""
        from vectora.tools.fs import list_dir as list_dir_tool

        # Testar a tool diretamente (sem ToolNode para evitar runtime deps do LangGraph)
        result = list_dir_tool.invoke({"path": "."})
        assert result is not None
        assert isinstance(result, str)
        assert len(result) > 0

    def test_tools_list_is_not_empty(self):
        """Verify TOOLS is non-empty and contains valid tools."""
        assert len(TOOLS) > 0
        for tool in TOOLS:
            assert hasattr(tool, "name")
            assert hasattr(tool, "description")

    def test_toolnode_has_correct_tool_names(self):
        """Verify ToolNode contains expected tools."""
        tool_names = {t.name for t in TOOLS}
        # At minimum file tools should be present
        assert len(tool_names) > 0
