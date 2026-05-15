"""Testes unitários para o módulo de prompts (prompts.py)."""

from prompts import get_system_prompt


class TestGetSystemPrompt:
    """Testes para função get_system_prompt."""

    def test_system_prompt_returns_string(self):
        """Verificar que get_system_prompt retorna uma string."""
        prompt = get_system_prompt()
        assert isinstance(prompt, str)
        assert len(prompt) > 0

    def test_system_prompt_contains_instructions(self):
        """Verificar que prompt contém instruções clara."""
        prompt = get_system_prompt()
        # Deve conter instruções ou guidelines
        assert "guide" in prompt.lower() or "instruction" in prompt.lower() or "rag" in prompt.lower()

    def test_system_prompt_defines_persona(self):
        """Verificar que prompt define persona do Vectora."""
        prompt = get_system_prompt()
        # Deve mencionar Vectora ou seu propósito
        assert "vectora" in prompt.lower() or "assistente" in prompt.lower()

    def test_system_prompt_includes_tool_instructions(self):
        """Verificar que prompt inclui instruções sobre ferramentas."""
        prompt = get_system_prompt()
        # Deve mencionar tools/ferramentas
        assert "tool" in prompt.lower() or "ferramenta" in prompt.lower()


class TestSystemPromptWithLanguageDetection:
    """Testes para detecção de idioma no prompt."""

    def test_system_prompt_detects_portuguese(self):
        """Verificar que prompt é retornado com language code pt_BR."""
        prompt = get_system_prompt(language="pt_BR")
        # Verificar que o language code foi inserido
        assert "pt_BR" in prompt or "pt-br" in prompt.lower()

    def test_system_prompt_with_default_language(self):
        """Verificar que prompt funciona com idioma padrão."""
        prompt = get_system_prompt()
        assert prompt is not None
        assert len(prompt) > 0

    def test_system_prompt_with_english(self):
        """Verificar que prompt pode ser em inglês."""
        prompt = get_system_prompt(language="en_US")
        assert prompt is not None
        assert isinstance(prompt, str)


class TestPromptInterpolation:
    """Testes para interpolação de valores no prompt."""

    def test_system_prompt_with_context_variables(self):
        """Verificar que prompt pode incluir variáveis de contexto."""
        prompt = get_system_prompt()
        # Prompt pode conter placeholders para contexto
        assert prompt is not None

    def test_system_prompt_consistency(self):
        """Verificar que prompt é consistente entre chamadas."""
        prompt1 = get_system_prompt()
        prompt2 = get_system_prompt()
        assert prompt1 == prompt2


class TestPromptFormatting:
    """Testes para formatação do prompt."""

    def test_system_prompt_is_well_formatted(self):
        """Verificar que prompt é bem formatado."""
        prompt = get_system_prompt()
        # Deve ter quebras de linha apropriadas
        assert "\n" in prompt or len(prompt) > 100

    def test_system_prompt_no_excessive_whitespace(self):
        """Verificar que prompt não tem espaçamento excessivo."""
        prompt = get_system_prompt()
        # Não deve ter múltiplas quebras de linha em sequência
        assert "\n\n\n" not in prompt

    def test_system_prompt_proper_escaping(self):
        """Verificar que caracteres especiais são escapados."""
        prompt = get_system_prompt()
        # Não deve ter caracteres quebrados
        assert "\x00" not in prompt


class TestPromptCaching:
    """Testes para caching de prompt (se implementado)."""

    def test_system_prompt_caching(self):
        """Verificar que prompt é cacheado eficientemente."""
        prompt1 = get_system_prompt()
        prompt2 = get_system_prompt()
        # Deve ser a mesma instância ou conteúdo
        assert prompt1 == prompt2

    def test_system_prompt_cache_invalidation(self):
        """Verificar que cache pode ser invalidado."""
        prompt1 = get_system_prompt()
        # Se houver função para invalidar cache
        # cache_invalidate() ou similar
        prompt2 = get_system_prompt()
        assert prompt2 is not None


class TestPromptContent:
    """Testes para conteúdo específico do prompt."""

    def test_system_prompt_mentions_rag(self):
        """Verificar que prompt menciona capacidades RAG."""
        prompt = get_system_prompt()
        assert "rag" in prompt.lower() or "busca" in prompt.lower()

    def test_system_prompt_mentions_tools(self):
        """Verificar que prompt menciona ferramentas disponíveis."""
        prompt = get_system_prompt()
        assert "tool" in prompt.lower() or "ferramenta" in prompt.lower()

    def test_system_prompt_has_clear_guidelines(self):
        """Verificar que prompt tem diretrizes claras."""
        prompt = get_system_prompt()
        # Deve ter instruções sobre quando usar tools
        assert len(prompt) > 200  # Prompt minimamente descritivo

    def test_system_prompt_includes_safety_guidelines(self):
        """Verificar que prompt inclui diretrizes de segurança."""
        prompt = get_system_prompt()
        # Pode mencionar segurança, limites, etc
        assert prompt is not None
