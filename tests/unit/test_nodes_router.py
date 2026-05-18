"""Testes para vectora/agents/supervisor.py"""

from __future__ import annotations

from typing import TYPE_CHECKING

import pytest
from langchain_core.messages import AIMessage, HumanMessage
from langgraph.types import Command

from vectora.agents.supervisor import classify_intent, supervisor

if TYPE_CHECKING:
    from vectora.state import State


class TestClassifyIntent:
    """Testes unitários para classify_intent()."""

    def test_direct_greeting_oi(self):
        assert classify_intent("oi") == "direct"

    def test_direct_greeting_hello(self):
        assert classify_intent("hello") == "direct"

    def test_direct_greeting_bom_dia(self):
        assert classify_intent("Bom dia!") == "direct"

    def test_direct_obrigado(self):
        assert classify_intent("obrigado") == "direct"

    def test_direct_who_are_you(self):
        assert classify_intent("quem é você") == "direct"

    def test_coder_criar_arquivo(self):
        assert classify_intent("cria um arquivo chamado test.py") == "coder"

    def test_coder_terminal(self):
        assert classify_intent("execute o comando git status no terminal") == "coder"

    def test_coder_git(self):
        assert classify_intent("rode git commit -m 'fix'") == "coder"

    def test_search_web_explicit(self):
        assert classify_intent("busca na web sobre Python 3.13") == "search"

    def test_coder_npm(self):
        assert classify_intent("instala o pacote via npm") == "coder"

    def test_rag_documentacao(self):
        assert classify_intent("o que diz o documento sobre autenticação?") == "rag"

    def test_rag_base_conhecimento(self):
        assert classify_intent("busque na base de conhecimento") == "rag"

    def test_rag_de_acordo_com(self):
        assert (
            classify_intent("de acordo com a documentação, como configurar?") == "rag"
        )

    def test_rag_long_question(self):
        # Perguntas longas sem trigger explícito → rag (heurística de comprimento)
        result = classify_intent(
            "como funciona o sistema de autenticação JWT no projeto?"
        )
        assert result == "rag"

    def test_direct_short_question(self):
        # Perguntas curtas sem trigger → direct
        result = classify_intent("ok")
        assert result == "direct"


class TestSupervisor:
    """Testes para supervisor()."""

    @pytest.mark.asyncio
    async def test_route_greeting_to_direct(self):
        state: State = {
            "messages": [HumanMessage(content="oi")],
            "session_metadata": {},
        }
        cmd = await supervisor(state)

        assert isinstance(cmd, Command)
        assert cmd.goto == "direct"
        assert cmd.update["routing_decision"] == "direct"

    @pytest.mark.asyncio
    async def test_route_coder_to_coder(self):
        state: State = {
            "messages": [HumanMessage(content="cria um arquivo main.py")],
            "session_metadata": {},
        }
        cmd = await supervisor(state)

        assert cmd.goto == "coder"
        assert cmd.update["routing_decision"] == "coder"

    @pytest.mark.asyncio
    async def test_route_rag_to_rag_subgraph(self):
        state: State = {
            "messages": [
                HumanMessage(content="o que diz o documento sobre autenticação?")
            ],
            "session_metadata": {},
        }
        cmd = await supervisor(state)

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
        cmd = await supervisor(state)

        assert cmd.goto == "rag_subgraph"
        assert cmd.update["routing_decision"] == "rag"

    @pytest.mark.asyncio
    async def test_route_empty_state_defaults_to_direct(self):
        """Sem mensagens humanas → fallback para direct."""
        state: State = {
            "messages": [AIMessage(content="Olá!")],
            "session_metadata": {},
        }
        cmd = await supervisor(state)

        assert cmd.goto == "direct"

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
        cmd = await supervisor(state)

        assert cmd.update["routing_decision"] == "direct"
        assert cmd.goto == "direct"

    @pytest.mark.asyncio
    async def test_route_empty_messages_defaults_to_direct(self):
        state: State = {"messages": [], "session_metadata": {}}
        cmd = await supervisor(state)

        assert cmd.goto == "direct"
        assert cmd.update["routing_decision"] == "direct"
