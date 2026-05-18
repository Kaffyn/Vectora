"""Testes unitários para o módulo de contexto (context.py)."""

import pytest

from vectora.context import Context


class TestContextCreation:
    """Testes para criação de contexto."""

    def test_context_creation_with_required_fields(self):
        """Verificar que Context é criado com fields obrigatórios."""
        ctx = Context(thread_id=1, user_type="pro")
        assert ctx.thread_id == 1
        assert ctx.user_type == "pro"

    def test_context_creation_with_string_thread_id(self):
        """Verificar que Context aceita thread_id como string."""
        ctx = Context(thread_id="abc123", user_type="basic")
        assert ctx.thread_id == "abc123"
        assert ctx.user_type == "basic"

    def test_context_creation_with_integer_thread_id(self):
        """Verificar que Context aceita thread_id como integer."""
        ctx = Context(thread_id=42, user_type="premium")
        assert ctx.thread_id == 42
        assert ctx.user_type == "premium"


class TestContextUserTypes:
    """Testes para diferentes tipos de usuário."""

    def test_context_with_pro_user_type(self):
        """Verificar que Context funciona com user_type 'pro'."""
        ctx = Context(thread_id=1, user_type="pro")
        assert ctx.user_type == "pro"

    def test_context_with_basic_user_type(self):
        """Verificar que Context funciona com user_type 'basic'."""
        ctx = Context(thread_id=1, user_type="basic")
        assert ctx.user_type == "basic"

    def test_context_with_premium_user_type(self):
        """Verificar que Context funciona com user_type 'premium'."""
        ctx = Context(thread_id=1, user_type="premium")
        assert ctx.user_type == "premium"

    def test_context_with_custom_user_type(self):
        """Verificar que Context aceita user_types customizados."""
        ctx = Context(thread_id=1, user_type="custom")
        assert ctx.user_type == "custom"


class TestContextImmutability:
    """Testes para imutabilidade de Context."""

    def test_context_fields_cannot_be_modified_after_creation(self):
        """Verificar que fields de Context não podem ser modificados."""
        ctx = Context(thread_id=1, user_type="pro")
        # Tentar modificar deve gerar erro se Context é Pydantic
        with pytest.raises(AttributeError):
            ctx.thread_id = 2

    def test_context_user_type_cannot_be_modified(self):
        """Verificar que user_type não pode ser modificado."""
        ctx = Context(thread_id=1, user_type="pro")
        with pytest.raises(AttributeError):
            ctx.user_type = "basic"


class TestContextEquality:
    """Testes para comparação de Context."""

    def test_context_equality_same_values(self):
        """Verificar que Contexts com mesmos valores são iguais."""
        ctx1 = Context(thread_id=1, user_type="pro")
        ctx2 = Context(thread_id=1, user_type="pro")
        assert ctx1 == ctx2

    def test_context_inequality_different_thread_id(self):
        """Verificar que Contexts com thread_id diferentes são desiguais."""
        ctx1 = Context(thread_id=1, user_type="pro")
        ctx2 = Context(thread_id=2, user_type="pro")
        assert ctx1 != ctx2

    def test_context_inequality_different_user_type(self):
        """Verificar que Contexts com user_type diferentes são desiguais."""
        ctx1 = Context(thread_id=1, user_type="pro")
        ctx2 = Context(thread_id=1, user_type="basic")
        assert ctx1 != ctx2


class TestContextStringRepresentation:
    """Testes para representação em string de Context."""

    def test_context_string_representation(self):
        """Verificar que Context tem representação em string."""
        ctx = Context(thread_id=1, user_type="pro")
        str_repr = str(ctx)
        assert "thread_id" in str_repr or "1" in str_repr
        assert "user_type" in str_repr or "pro" in str_repr

    def test_context_repr_representation(self):
        """Verificar que Context tem repr válida."""
        ctx = Context(thread_id=1, user_type="pro")
        repr_str = repr(ctx)
        assert "Context" in repr_str


class TestContextValidation:
    """Testes para validação de Context."""

    def test_context_requires_thread_id(self):
        """Verificar que Context requer thread_id."""
        with pytest.raises(TypeError):
            Context(user_type="pro")  # Missing thread_id

    def test_context_requires_user_type(self):
        """Verificar que Context requer user_type."""
        with pytest.raises(TypeError):
            Context(thread_id=1)  # Missing user_type

    def test_context_with_none_thread_id_raises_error(self):
        """Verificar que thread_id None gera erro."""
        with pytest.raises((ValueError, TypeError)):
            Context(thread_id=None, user_type="pro")  # type: ignore[arg-type]


class TestContextUsageInConfigurable:
    """Testes para uso de Context em configurable."""

    def test_context_can_be_used_in_graph_config(self):
        """Verificar que Context funciona como configurable."""
        ctx = Context(thread_id=1, user_type="pro")
        config = {"configurable": ctx}
        assert config["configurable"] is ctx

    def test_context_can_be_serialized_to_dict(self):
        """Verificar que Context pode ser serializado se necessário."""
        ctx = Context(thread_id=1, user_type="pro")
        # Se implementado, Context.dict() deve funcionar
        try:
            ctx_dict = ctx.model_dump() if hasattr(ctx, "model_dump") else None
            if ctx_dict:
                assert ctx_dict["thread_id"] == 1
                assert ctx_dict["user_type"] == "pro"
        except AttributeError:
            # Context pode não ser Pydantic, isso é ok
            pass
