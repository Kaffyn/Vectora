"""Tests for vectora/agents/direct.py"""

from __future__ import annotations

from typing import TYPE_CHECKING
from unittest.mock import AsyncMock, patch

import pytest
from langchain_core.messages import AIMessage, HumanMessage

from vectora.agents.direct import SYSTEM_PROMPT, direct

if TYPE_CHECKING:
    from vectora.state import State


def test_system_prompt_exists():
    assert isinstance(SYSTEM_PROMPT, str)
    assert len(SYSTEM_PROMPT) > 200


def test_system_prompt_contains_identity():
    assert "Vectora" in SYSTEM_PROMPT
    assert "LangGraph" in SYSTEM_PROMPT
    assert "LanceDB" in SYSTEM_PROMPT


def test_system_prompt_describes_role():
    assert "Direct Agent" in SYSTEM_PROMPT
    assert "RAG" in SYSTEM_PROMPT


@pytest.mark.asyncio
async def test_direct_calls_invoke_llm():
    state: State = {
        "messages": [HumanMessage(content="oi")],
        "session_metadata": {},
    }
    mock_response = {"messages": [AIMessage(content="Olá!")]}

    with patch("vectora.agents.direct.invoke_llm", new_callable=AsyncMock) as mock_llm:
        mock_llm.return_value = mock_response
        result = await direct(state)

    mock_llm.assert_called_once()
    call_kwargs = mock_llm.call_args
    assert call_kwargs.kwargs.get("system_prompt") == SYSTEM_PROMPT
    assert result == mock_response


@pytest.mark.asyncio
async def test_direct_passes_system_prompt():
    state: State = {"messages": [HumanMessage(content="test")], "session_metadata": {}}

    with patch("vectora.agents.direct.invoke_llm", new_callable=AsyncMock) as mock_llm:
        mock_llm.return_value = {"messages": [AIMessage(content="ok")]}
        await direct(state)

    _, kwargs = mock_llm.call_args
    assert kwargs["system_prompt"] is SYSTEM_PROMPT
