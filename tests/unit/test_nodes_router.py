"""Testes para vectora/nodes/router.py"""

from __future__ import annotations

from typing import TYPE_CHECKING

import pytest
from langchain_core.messages import AIMessage, HumanMessage
from langgraph.types import Command
from vectora.nodes.router import _classify_intent, route_request

if TYPE_CHECKING:
    from vectora.state import State


class TestClassifyIntent:
    """Testes unitários para _classify_intent()."""

    def test_direct_greeting_oi(self):
        assert _classify_intent("oi") == "direct"

    def test_direct_greeting_hello(self):
        assert _classify_intent("hello") == "direct"

    def test_direct_greeting_bom_dia(self):
        assert _classify_intent("Bom dia!") == "direct"

    def test_direct_obrigado(self):
        assert _classify_intent("obrigado") == "direct"

    def test_direct_who_are_you(self):
        assert _classify_intent("quem é você") == "direct"

    def test_tools_criar_arquivo(self):
        assert _classify_intent("cria um arquivo chamado test.py") == "tools"

    def test_tools_terminal(self):
        assert _classify_intent("execute o comando git status no terminal") == "tools"

    def test_tools_git(self):
        assert _classify_intent("rode git commit -m 'fix'") == "tools"

    def test_tools_search_web_explicit(self):
        assert _classify_intent("busca na web sobre Python 3.13") == "tools"

    def test_tools_npm(self):
        assert _classify_intent("instala o pacote via npm") == "tools"

    def test_rag_documentacao(self):
        assert _classify_intent("o que diz o documento sobre autenticação?") == "rag"

    def test_rag_base_conhecimento(self):
        assert _classify_intent("busque na base de conhecimento") == "rag"

    def test_rag_de_acordo_com(self):
        assert (
            _classify_intent("de acordo com a documentação, como configurar?") == "rag"
        )

    def test_rag_long_question(self):
        # Perguntas longas sem trigger explícito → rag (heurística de comprimento)
        result = _classify_intent(
            "como funciona o sistema de autenticação JWT no projeto?"
        )
        assert result == "rag"

    def test_direct_short_question(self):
        # Perguntas curtas sem trigger → direct
        result = _classify_intent("ok")
        assert result == "direct"


class TestRouteRequest:
    """Testes para route_request()."""

    @pytest.mark.asyncio
    async def test_route_greeting_to_direct(self):
        state: State = {
            "messages": [HumanMessage(content="oi")],
            "session_metadata": {},
        }
        cmd = await route_request(state)

        assert isinstance(cmd, Command)
        assert cmd.goto == "call_llm"
        assert cmd.update["routing_decision"] == "direct"

    @pytest.mark.asyncio
    async def test_route_tools_to_call_llm(self):
        state: State = {
            "messages": [HumanMessage(content="cria um arquivo main.py")],
            "session_metadata": {},
        }
        cmd = await route_request(state)

        assert cmd.goto == "call_llm"
        assert cmd.update["routing_decision"] == "tools"

    @pytest.mark.asyncio
    async def test_route_rag_to_rag_subgraph(self):
        state: State = {
            "messages": [
                HumanMessage(content="o que diz o documento sobre autenticação?")
            ],
            "session_metadata": {},
        }
        cmd = await route_request(state)

        assert cmd.goto == "rag_subgraph"
        assert cmd.update["routing_decision"] == "rag"

    @pytest.mark.asyncio
    async def test_route_long_question_to_rag(self):
        state: State = {
            "messages": [
                HumanMessage(
                    content="como funciona a arquitetura de microserviços do sistema?"
                )
            ],
            "session_metadata": {},
        }
        cmd = await route_request(state)

        assert cmd.goto == "rag_subgraph"
        assert cmd.update["routing_decision"] == "rag"

    @pytest.mark.asyncio
    async def test_route_empty_state_defaults_to_direct(self):
        """Sem mensagens humanas → fallback para direct."""
        state: State = {
            "messages": [AIMessage(content="Olá!")],
            "session_metadata": {},
        }
        cmd = await route_request(state)

        assert cmd.goto == "call_llm"

    @pytest.mark.asyncio
    async def test_route_uses_last_human_message(self):
        """Deve usar a ÚLTIMA HumanMessage, não a primeira."""
        state: State = {
            "messages": [
                HumanMessage(content="o que diz o documento?"),  # primeira → rag
                AIMessage(content="Resposta"),
                HumanMessage(content="oi"),  # última → direct
            ],
            "session_metadata": {},
        }
        cmd = await route_request(state)

        assert cmd.update["routing_decision"] == "direct"
        assert cmd.goto == "call_llm"

    @pytest.mark.asyncio
    async def test_route_empty_messages_defaults_to_direct(self):
        state: State = {"messages": [], "session_metadata": {}}
        cmd = await route_request(state)

        assert cmd.goto == "call_llm"
        assert cmd.update["routing_decision"] == "direct"
