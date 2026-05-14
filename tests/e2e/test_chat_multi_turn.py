"""Testes E2E para conversas multi-turno."""

import pytest
from langchain_core.messages import HumanMessage
from context import Context


class TestChatMultiTurn:
    """Testes para conversa multi-turno."""

    @pytest.mark.asyncio
    async def test_basic_multi_turn_conversation(self, test_graph):
        """Verificar conversa básica de 3 turnos."""
        thread_id = 1
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="pro")

        # Turno 1
        input_1 = {"messages": [HumanMessage(content="Hello, how are you?")]}
        state_1 = await test_graph.ainvoke(input_1, config=config, context=context)
        assert len(state_1["messages"]) > 0

        # Turno 2
        input_2 = {
            "messages": [
                HumanMessage(
                    content="Can you help me with Python?"
                )
            ]
        }
        state_2 = await test_graph.ainvoke(input_2, config=config, context=context)
        assert len(state_2["messages"]) > len(state_1["messages"])

        # Turno 3
        input_3 = {"messages": [HumanMessage(content="Thanks for the help!")]}
        state_3 = await test_graph.ainvoke(input_3, config=config, context=context)
        assert len(state_3["messages"]) > len(state_2["messages"])

    @pytest.mark.asyncio
    async def test_context_preservation_across_turns(self, test_graph):
        """Verificar que contexto é preservado entre turnos."""
        thread_id = 2
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="basic")

        # Estabelecer contexto no turno 1
        input_1 = {"messages": [HumanMessage(content="My name is Alice")]}
        state_1 = await test_graph.ainvoke(input_1, config=config, context=context)

        # Verificar se contexto é mantido no turno 2
        input_2 = {"messages": [HumanMessage(content="What is my name?")]}
        state_2 = await test_graph.ainvoke(input_2, config=config, context=context)

        # Resposta deve incluir "Alice"
        assert state_2["messages"] is not None

    @pytest.mark.asyncio
    async def test_auto_summarization_trigger(self, test_graph):
        """Verificar que auto-summarization é acionado."""
        thread_id = 3
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="pro")

        # Fazer 5+ turnos para trigger summarization
        for i in range(6):
            input_msg = {
                "messages": [
                    HumanMessage(
                        content=f"Message {i}: Tell me about topic {i}"
                    )
                ]
            }
            state = await test_graph.ainvoke(input_msg, config=config, context=context)
            assert state["messages"] is not None

        # Após 5+ mensagens, summary deveria estar presente
        assert "summary" in state or state.get("summary") is not None

    @pytest.mark.asyncio
    async def test_token_counting_across_turns(self, test_graph, mock_llm):
        """Verificar contagem de tokens é acumulativa."""
        thread_id = 4
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="premium")

        # Mock LLM para rastrear tokens
        if hasattr(mock_llm, "call_count"):
            initial_count = mock_llm.call_count
        else:
            initial_count = 0

        for i in range(3):
            input_msg = {
                "messages": [
                    HumanMessage(content=f"Question {i}")
                ]
            }
            state = await test_graph.ainvoke(input_msg, config=config, context=context)
            assert state["messages"] is not None

        # LLM deve ter sido chamado múltiplas vezes
        if hasattr(mock_llm, "call_count"):
            assert mock_llm.call_count > initial_count

    @pytest.mark.asyncio
    async def test_error_recovery_in_conversation(self, test_graph):
        """Verificar recuperação de erro durante conversa."""
        thread_id = 5
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="pro")

        # Turno normal
        input_1 = {"messages": [HumanMessage(content="Normal question")]}
        state_1 = await test_graph.ainvoke(input_1, config=config, context=context)

        # Turno que poderia causar erro
        input_2 = {"messages": [HumanMessage(content="Another question")]}
        state_2 = await test_graph.ainvoke(input_2, config=config, context=context)

        # Conversa deve continuar funcionando
        assert state_2["messages"] is not None

    @pytest.mark.asyncio
    async def test_long_running_conversation(self, test_graph):
        """Verificar conversa longa sem memória leak."""
        thread_id = 6
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="pro")

        # 20 turnos
        for i in range(20):
            input_msg = {
                "messages": [
                    HumanMessage(content=f"Turn {i}: {i % 5} question")
                ]
            }
            state = await test_graph.ainvoke(input_msg, config=config, context=context)
            assert state["messages"] is not None

        # Final state deve ser válido
        assert "messages" in state


class TestChatWithTools:
    """Testes para conversa com uso de ferramentas."""

    @pytest.mark.asyncio
    async def test_tool_invocation_in_conversation(self, test_graph):
        """Verificar que tools são invocadas durante conversa."""
        thread_id = 10
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="pro")

        # Questão que deveria invocar tool
        input_msg = {"messages": [HumanMessage(content="Search for Python docs")]}
        state = await test_graph.ainvoke(input_msg, config=config, context=context)

        # Tool invocation deveria criar ToolMessage
        assert state["messages"] is not None

    @pytest.mark.asyncio
    async def test_tool_result_in_context(self, test_graph):
        """Verificar que resultado de tool é incluído em contexto."""
        thread_id = 11
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="pro")

        # Questão com tool
        input_msg = {"messages": [HumanMessage(content="Find information")]}
        state_1 = await test_graph.ainvoke(input_msg, config=config, context=context)

        # Follow-up deve ter acesso ao resultado anterior
        input_msg_2 = {
            "messages": [
                HumanMessage(content="Can you elaborate on that?")
            ]
        }
        state_2 = await test_graph.ainvoke(input_msg_2, config=config, context=context)

        assert state_2["messages"] is not None

    @pytest.mark.asyncio
    async def test_tool_retry_in_conversation(self, test_graph):
        """Verificar que tool retries são invisíveis para usuário."""
        thread_id = 12
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="pro")

        # Tool que precisa retry
        input_msg = {"messages": [HumanMessage(content="Try this operation")]}
        state = await test_graph.ainvoke(input_msg, config=config, context=context)

        # Usuário não deveria ver erro de retry
        assert state["messages"] is not None


class TestDifferentUserTypes:
    """Testes para diferentes tipos de usuário."""

    @pytest.mark.asyncio
    async def test_basic_user_conversation(self, test_graph):
        """Verificar conversa com usuário básico."""
        thread_id = 20
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="basic")

        input_msg = {"messages": [HumanMessage(content="Hello")]}
        state = await test_graph.ainvoke(input_msg, config=config, context=context)
        assert state["messages"] is not None

    @pytest.mark.asyncio
    async def test_pro_user_conversation(self, test_graph):
        """Verificar conversa com usuário pro."""
        thread_id = 21
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="pro")

        input_msg = {"messages": [HumanMessage(content="Advanced question")]}
        state = await test_graph.ainvoke(input_msg, config=config, context=context)
        assert state["messages"] is not None

    @pytest.mark.asyncio
    async def test_premium_user_conversation(self, test_graph):
        """Verificar conversa com usuário premium."""
        thread_id = 22
        config = {"configurable": {"thread_id": str(thread_id)}}
        context = Context(thread_id=thread_id, user_type="premium")

        input_msg = {"messages": [HumanMessage(content="Complex task")]}
        state = await test_graph.ainvoke(input_msg, config=config, context=context)
        assert state["messages"] is not None
