"""Tests for vectora/agents/supervisor.py"""

from __future__ import annotations

from typing import TYPE_CHECKING

import pytest
from langchain_core.messages import AIMessage, HumanMessage
from langgraph.types import Command

from vectora.agents.supervisor import classify_intent, supervisor

if TYPE_CHECKING:
    from vectora.state import State

# ---------------------------------------------------------------------------
# classify_intent
# ---------------------------------------------------------------------------


class TestClassifyIntent:
    def test_direct_greeting(self):
        for text in ("oi", "olá", "hello", "hi", "bom dia", "boa tarde"):
            assert classify_intent(text) == "direct", f"failed for: {text}"

    def test_direct_thanks(self):
        assert classify_intent("obrigado") == "direct"
        assert classify_intent("valeu") == "direct"

    def test_direct_who_are_you(self):
        assert classify_intent("quem é você") == "direct"
        assert classify_intent("o que você faz") == "direct"

    def test_direct_short_fallback(self):
        assert classify_intent("ok") == "direct"

    def test_coder_file_operations(self):
        assert classify_intent("cria um arquivo main.py") == "coder"
        assert classify_intent("edita o script") == "coder"

    def test_coder_git(self):
        assert classify_intent("rode git commit -m 'fix'") == "coder"
        assert classify_intent("git status do projeto") == "coder"

    def test_coder_terminal(self):
        assert classify_intent("executa npm install no terminal") == "coder"
        assert classify_intent("instala o pacote via pip") == "coder"

    def test_search_web_explicit(self):
        assert classify_intent("busca na web sobre Python 3.13") == "search"
        assert classify_intent("pesquisa na internet por notícias") == "search"

    def test_rag_document(self):
        assert classify_intent("o que diz o documento sobre JWT?") == "rag"
        assert classify_intent("busque na base de conhecimento") == "rag"

    def test_rag_de_acordo_com(self):
        assert (
            classify_intent("de acordo com a documentação, como configurar?") == "rag"
        )

    def test_rag_long_question_fallback(self):
        # mensagens longas sem trigger explícito → rag
        result = classify_intent("explique como funciona a arquitetura do sistema?")
        assert result == "rag"


# ---------------------------------------------------------------------------
# supervisor node
# ---------------------------------------------------------------------------


class TestSupervisor:
    @pytest.mark.asyncio
    async def test_greeting_routes_to_direct(self):
        state: State = {
            "messages": [HumanMessage(content="oi")],
            "session_metadata": {},
        }
        cmd = await supervisor(state)
        assert isinstance(cmd, Command)
        assert cmd.goto == "direct"
        assert cmd.update["routing_decision"] == "direct"

    @pytest.mark.asyncio
    async def test_coder_routes_to_coder(self):
        state: State = {
            "messages": [HumanMessage(content="cria um arquivo main.py")],
            "session_metadata": {},
        }
        cmd = await supervisor(state)
        assert cmd.goto == "coder"
        assert cmd.update["routing_decision"] == "coder"

    @pytest.mark.asyncio
    async def test_rag_routes_to_rag_subgraph(self):
        state: State = {
            "messages": [HumanMessage(content="o que diz o documento sobre auth?")],
            "session_metadata": {},
        }
        cmd = await supervisor(state)
        assert cmd.goto == "rag_subgraph"
        assert cmd.update["routing_decision"] == "rag"

    @pytest.mark.asyncio
    async def test_uses_last_human_message(self):
        state: State = {
            "messages": [
                HumanMessage(content="o que diz o documento?"),  # → rag
                AIMessage(content="Resposta"),
                HumanMessage(content="oi"),  # → direct (última)
            ],
            "session_metadata": {},
        }
        cmd = await supervisor(state)
        assert cmd.update["routing_decision"] == "direct"

    @pytest.mark.asyncio
    async def test_empty_messages_defaults_to_direct(self):
        state: State = {"messages": [], "session_metadata": {}}
        cmd = await supervisor(state)
        assert cmd.goto == "direct"

    @pytest.mark.asyncio
    async def test_no_human_message_defaults_to_direct(self):
        state: State = {
            "messages": [AIMessage(content="resposta")],
            "session_metadata": {},
        }
        cmd = await supervisor(state)
        assert cmd.goto == "direct"
