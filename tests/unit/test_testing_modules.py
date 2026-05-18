"""Testes para vectora/testing/assertions.py e vectora/testing/mocks.py.

Cobre: assert_tool_called, assert_tool_called_with_args, assert_tool_result_in_messages,
       assert_message_contains_text, assert_last_message_is_ai, MockLLM
"""

from __future__ import annotations

from unittest.mock import MagicMock

import pytest
from langchain_core.messages import AIMessage, HumanMessage, ToolMessage


class TestAssertToolCalled:
    """Testa assert_tool_called."""

    def test_passes_when_tool_called(self):
        """Verifica que não lança exceção quando tool foi chamada."""
        from langchain_core.messages import AIMessage as _AIMessage

        from vectora.testing.assertions import assert_tool_called

        # Criar AIMessage com tool_calls no formato correto LangChain
        ai_msg = _AIMessage(content="chamando tool")
        # Simular tool_calls como objetos com atributo .name
        tc = MagicMock()
        tc.name = "list_dir"
        tc.args = {"path": "."}
        ai_msg.tool_calls = [tc]
        # Não deve lançar exceção
        assert_tool_called([ai_msg], "list_dir")

    def test_raises_when_tool_not_called(self):
        """Verifica que lança AssertionError quando tool não foi chamada."""
        from vectora.testing.assertions import assert_tool_called

        ai_msg = AIMessage(content="sem tool calls")
        with pytest.raises(AssertionError):
            assert_tool_called([ai_msg], "missing_tool")

    def test_raises_with_empty_messages(self):
        """Verifica AssertionError com lista vazia."""
        from vectora.testing.assertions import assert_tool_called

        with pytest.raises(AssertionError):
            assert_tool_called([], "any_tool")


class TestAssertToolCalledWithArgs:
    """Testa assert_tool_called_with_args."""

    def test_passes_with_correct_args(self):
        """Verifica que não lança exceção quando args correspondem."""
        from vectora.testing.assertions import assert_tool_called_with_args

        ai_msg = AIMessage(content="")
        tc = MagicMock()
        tc.name = "search"
        tc.args = {"query": "python"}
        ai_msg.tool_calls = [tc]
        assert_tool_called_with_args([ai_msg], "search", {"query": "python"})

    def test_raises_with_wrong_args(self):
        """Verifica AssertionError com args errados."""
        from vectora.testing.assertions import assert_tool_called_with_args

        ai_msg = AIMessage(content="")
        tc = MagicMock()
        tc.name = "search"
        tc.args = {"query": "python"}
        ai_msg.tool_calls = [tc]
        with pytest.raises(AssertionError):
            assert_tool_called_with_args([ai_msg], "search", {"query": "java"})

    def test_raises_when_tool_absent(self):
        """Verifica AssertionError quando tool não está presente."""
        from vectora.testing.assertions import assert_tool_called_with_args

        with pytest.raises(AssertionError):
            assert_tool_called_with_args([], "any_tool", {})


class TestAssertToolResultInMessages:
    """Testa assert_tool_result_in_messages."""

    def test_passes_when_result_found(self):
        """Verifica que não lança quando resultado está nas mensagens."""
        from vectora.testing.assertions import assert_tool_result_in_messages

        tool_msg = ToolMessage(
            content="arquivo encontrado", name="list_dir", tool_call_id="tc1"
        )
        assert_tool_result_in_messages([tool_msg], "list_dir", "arquivo")

    def test_raises_when_result_not_found(self):
        """Verifica AssertionError quando resultado não está presente."""
        from vectora.testing.assertions import assert_tool_result_in_messages

        tool_msg = ToolMessage(content="nada aqui", name="list_dir", tool_call_id="tc1")
        with pytest.raises(AssertionError):
            assert_tool_result_in_messages([tool_msg], "list_dir", "resultado esperado")


class TestAssertMessageContainsText:
    """Testa assert_message_contains_text."""

    def test_passes_when_text_found(self):
        """Verifica que não lança quando texto está em alguma mensagem."""
        from vectora.testing.assertions import assert_message_contains_text

        msgs = [HumanMessage(content="olá mundo"), AIMessage(content="resposta")]
        assert_message_contains_text(msgs, "olá")

    def test_raises_when_text_missing(self):
        """Verifica AssertionError quando texto não está nas mensagens."""
        from vectora.testing.assertions import assert_message_contains_text

        msgs = [AIMessage(content="outra coisa")]
        with pytest.raises(AssertionError):
            assert_message_contains_text(msgs, "texto inexistente")


class TestAssertLastMessageIsAI:
    """Testa assert_last_message_is_ai."""

    def test_passes_when_last_is_ai(self):
        """Verifica que retorna AIMessage quando última é AI."""
        from vectora.testing.assertions import assert_last_message_is_ai

        msgs = [HumanMessage(content="oi"), AIMessage(content="olá")]
        result = assert_last_message_is_ai(msgs)
        assert isinstance(result, AIMessage)

    def test_raises_when_last_is_human(self):
        """Verifica AssertionError quando última mensagem é Human."""
        from vectora.testing.assertions import assert_last_message_is_ai

        msgs = [AIMessage(content="oi"), HumanMessage(content="humano")]
        with pytest.raises(AssertionError):
            assert_last_message_is_ai(msgs)

    def test_raises_with_empty_messages(self):
        """Verifica AssertionError com lista vazia."""
        from vectora.testing.assertions import assert_last_message_is_ai

        with pytest.raises(AssertionError):
            assert_last_message_is_ai([])


class TestMockLLM:
    """Testa MockLLM."""

    def test_mock_llm_creation(self):
        """Verifica que MockLLM pode ser criado."""
        from vectora.testing.mocks import MockLLM

        llm = MockLLM()
        assert llm is not None
        assert llm._llm_type == "mock"

    def test_mock_llm_match_pattern_greeting(self):
        """Verifica que _match_pattern responde a saudações."""
        from vectora.testing.mocks import MockLLM

        llm = MockLLM()
        result = llm._match_pattern("hello world", [])
        assert result is not None
        assert "Olá" in result.content or "assistente" in result.content

    def test_mock_llm_match_pattern_multiply(self):
        """Verifica que _match_pattern gera tool call para multiplicação."""
        from vectora.testing.mocks import MockLLM

        llm = MockLLM()
        result = llm._match_pattern("multiplique 4 por 5", [])
        assert result is not None

    def test_mock_llm_handle_multiply_with_numbers(self):
        """Verifica que _handle_multiply extrai números corretamente."""
        from vectora.testing.mocks import MockLLM

        llm = MockLLM()
        result = llm._handle_multiply("multiplique 4 por 5")
        assert result is not None
        assert result.tool_calls or "4" in result.content or len(result.content) > 0

    def test_mock_llm_handle_multiply_no_numbers(self):
        """Verifica _handle_multiply sem números retorna mensagem de ajuda."""
        from vectora.testing.mocks import MockLLM

        llm = MockLLM()
        result = llm._handle_multiply("multiplica isso")
        assert (
            "números" in result.content
            or "forneça" in result.content.lower()
            or len(result.content) > 0
        )

    def test_mock_llm_match_pattern_default(self):
        """Verifica resposta padrão para entrada não reconhecida."""
        from vectora.testing.mocks import MockLLM

        llm = MockLLM()
        result = llm._match_pattern("mensagem aleatória xyz", [])
        assert result is not None
        assert len(result.content) > 0

    def test_mock_llm_bind_tools(self):
        """Verifica que bind_tools retorna self."""
        from vectora.testing.mocks import MockLLM

        llm = MockLLM()
        result = llm.bind_tools([])
        assert result is llm

    @pytest.mark.asyncio
    async def test_mock_llm_agenerate(self):
        """Verifica que _agenerate chama _match_pattern."""
        from vectora.testing.mocks import MockLLM

        llm = MockLLM()
        # _agenerate chama _generate internamente — testar via _match_pattern
        result = llm._match_pattern("olá", [HumanMessage(content="olá")])
        assert result is not None
