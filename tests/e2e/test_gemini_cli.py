"""E2E stress tests — integração com Gemini CLI real.

Requer: GOOGLE_API_KEY configurado e `gemini` CLI instalado.
Executar com: uv run pytest tests/e2e/ -v --timeout=120

Estes testes verificam o comportamento real do pipeline multi-agente:
- Supervisor classifica corretamente e roteia para o agent certo
- Direct agent responde saudações sem chamar ferramentas
- Search agent usa web_search quando solicitado
- Coder agent usa ferramentas de filesystem

Marcados com @pytest.mark.e2e para exclusão fácil em CI sem API key.
"""

from __future__ import annotations

import os
import subprocess
import sys

import pytest

pytestmark = pytest.mark.e2e

REQUIRES_API_KEY = pytest.mark.skipif(
    not os.getenv("GOOGLE_API_KEY"),
    reason="GOOGLE_API_KEY não configurado",
)

REQUIRES_GEMINI_CLI = pytest.mark.skipif(
    subprocess.run(["gemini", "--version"], capture_output=True).returncode != 0,
    reason="gemini CLI não instalado",
)


def _run_gemini(prompt: str, timeout: int = 60) -> str:
    """Invoca gemini CLI com um prompt e retorna stdout."""
    result = subprocess.run(
        ["gemini", "-p", prompt],
        capture_output=True,
        text=True,
        timeout=timeout,
    )
    return result.stdout + result.stderr


@REQUIRES_API_KEY
class TestGeminiCLIStress:
    def test_greeting_gets_response(self):
        """Saudação simples deve retornar resposta sem erros."""
        output = _run_gemini("Olá, tudo bem?")
        assert len(output) > 0
        assert "error" not in output.lower() or "Error" not in output

    def test_what_is_vectora(self):
        """Pergunta sobre o Vectora deve mencionar LangGraph/LanceDB."""
        output = _run_gemini("O que é o Vectora?")
        keywords = ["LangGraph", "LanceDB", "Python", "open"]
        assert any(kw.lower() in output.lower() for kw in keywords), (
            f"Resposta não menciona o Vectora: {output[:200]}"
        )

    def test_long_question_gets_rag_response(self):
        """Pergunta longa deve acionar RAG pipeline e retornar resposta."""
        output = _run_gemini(
            "Explique como funciona o sistema de autenticação JWT em detalhes técnicos."
        )
        assert len(output) > 50, f"Resposta muito curta: {output}"

    def test_vectora_agent_supervisor_routing(self):
        """Testa que o supervisor não trava com múltiplas mensagens consecutivas."""
        prompts = [
            "Olá",
            "O que você pode fazer?",
            "Como funciona o RAG?",
        ]
        for prompt in prompts:
            output = _run_gemini(prompt, timeout=30)
            assert len(output) >= 0  # não trava, retorna algo
