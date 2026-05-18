"""Tests for vectora/agents/search.py"""

from __future__ import annotations

from typing import TYPE_CHECKING
from unittest.mock import AsyncMock, patch

import pytest
from langchain_core.messages import AIMessage, HumanMessage

from vectora.agents.search import SYSTEM_PROMPT, search

if TYPE_CHECKING:
    from vectora.state import State


def test_system_prompt_exists():
    assert isinstance(SYSTEM_PROMPT, str)
    assert len(SYSTEM_PROMPT) > 200


def test_system_prompt_contains_identity():
    assert "Vectora" in SYSTEM_PROMPT
    assert "LangGraph" in SYSTEM_PROMPT


def test_system_prompt_describes_tools():
    assert "web_search" in SYSTEM_PROMPT
    assert "vector_search" in SYSTEM_PROMPT
    assert "fetch_url" in SYSTEM_PROMPT


def test_system_prompt_describes_rag_first():
    assert "vector_search" in SYSTEM_PROMPT
    # estratégia RAG-first deve estar documentada
    assert "LanceDB" in SYSTEM_PROMPT or "local" in SYSTEM_PROMPT.lower()


@pytest.mark.asyncio
async def test_search_calls_invoke_llm():
    state: State = {
        "messages": [HumanMessage(content="busca na web sobre Python")],
        "session_metadata": {},
    }
    mock_response = {"messages": [AIMessage(content="Resultado")]}

    with patch("vectora.agents.search.invoke_llm", new_callable=AsyncMock) as mock_llm:
        mock_llm.return_value = mock_response
        result = await search(state)

    mock_llm.assert_called_once()
    assert result == mock_response


@pytest.mark.asyncio
async def test_search_passes_system_prompt():
    state: State = {"messages": [HumanMessage(content="test")], "session_metadata": {}}

    with patch("vectora.agents.search.invoke_llm", new_callable=AsyncMock) as mock_llm:
        mock_llm.return_value = {"messages": [AIMessage(content="ok")]}
        await search(state)

    _, kwargs = mock_llm.call_args
    assert kwargs["system_prompt"] is SYSTEM_PROMPT
