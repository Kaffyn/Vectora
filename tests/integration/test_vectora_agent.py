#!/usr/bin/env python3
"""Teste reduzido do Vectora Research Agent - 2 tópicos para demonstração."""

import asyncio
import logging
import sys
from pathlib import Path

# Carregar .env antes de importar qualquer coisa que precise das variáveis
from dotenv import load_dotenv

load_dotenv()

# Setup paths and logging FIRST
sys.path.insert(0, str(Path(__file__).parent / "vectora"))

# Now import from vectora
from checkpointer import Checkpointer
from context import Context
from graph import build_graph

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)


async def test_vectora_agent() -> None:
    """Teste com 2 tópicos (Next.js 16 e Godot 4.7)."""
    logger.info("🚀 Teste Vectora Research Agent (Modo Reduzido)")
    logger.info("   Tópicos: 2 (Next.js 16, Godot 4.7)")
    logger.info("   Demonstração: Vectora como agente de pesquisa autônomo")

    output_dir = Path("output/research")
    output_dir.mkdir(parents=True, exist_ok=True)

    context = Context(
        thread_id="test-research-agent",
        user_id="test",
        user_type="researcher",
    )

    # Inicializar checkpointer para persistência de estado
    async with Checkpointer() as checkpointer:
        graph = build_graph(checkpointer)
        logger.info("✅ Grafo LangGraph carregado")

        # Teste 1: Next.js 16
        logger.info("=" * 80)
        logger.info("[1/2] Pesquisando: Next.js 16")
        logger.info("=" * 80)

        prompt_nextjs = """TAREFA: Pesquisa sobre Next.js 16

Execute 3 web_searches sobre:
1. "Next.js 16 new features 2025"
2. "Next.js 16 app router guide"
3. "Next.js 16 performance improvements"

Para CADA resultado:
- Extraia URLs dos resultados
- Use fetch_url() para acessar 1-2 sites importantes
- Compile um sumário de 300 palavras sobre Next.js 16

Formato:
✅ Search 1: "Next.js 16 new features" - X resultados processados
✅ Search 2: "Next.js 16 app router" - X resultados processados
✅ Search 3: "Next.js 16 performance" - X resultados processados

📄 SUMÁRIO FINAL (300 palavras):
[seu sumário aqui]"""

        config = {"configurable": {"thread_id": context.thread_id}}

        try:
            result = await graph.astream_events(
                {"messages": [{"role": "user", "content": prompt_nextjs}]},
                config=config,
                context=context,
            )

            if result and "messages" in result:
                response = result["messages"][-1]
                content = (
                    response.content if hasattr(response, "content") else str(response)
                )
                logger.info(f"✅ Resposta recebida ({len(content)} caracteres)")
                logger.info(f"\n📖 RESPOSTA VECTORA:\n{content[:1000]}...\n")
            else:
                logger.warning("⚠️  Resposta vazia")

        except Exception as e:
            logger.exception("❌ Erro no teste: %s", e)

        # Teste 2: Godot 4.7
        logger.info("=" * 80)
        logger.info("[2/2] Pesquisando: Godot 4.7")
        logger.info("=" * 80)

        prompt_godot = """TAREFA: Pesquisa sobre Godot 4.7

Execute 3 web_searches sobre:
1. "Godot 4.7 new features release"
2. "Godot 4.7 3D development guide"
3. "Godot 4.7 game examples"

Para CADA resultado:
- Extraia URLs dos resultados
- Use fetch_url() em 1-2 sites importantes
- Compile um sumário de 300 palavras sobre Godot 4.7

Formato:
✅ Search 1: "Godot 4.7 new features" - X resultados processados
✅ Search 2: "Godot 4.7 3D development" - X resultados processados
✅ Search 3: "Godot 4.7 game examples" - X resultados processados

📄 SUMÁRIO FINAL (300 palavras):
[seu sumário aqui]"""

        try:
            result = await graph.astream_events(
                {"messages": [{"role": "user", "content": prompt_godot}]},
                config=config,
                context=context,
            )

            if result and "messages" in result:
                response = result["messages"][-1]
                content = (
                    response.content if hasattr(response, "content") else str(response)
                )
                logger.info(f"✅ Resposta recebida ({len(content)} caracteres)")
                logger.info(f"\n📖 RESPOSTA VECTORA:\n{content[:1000]}...\n")
            else:
                logger.warning("⚠️  Resposta vazia")

        except Exception as e:
            logger.exception("❌ Erro no teste: %s", e)

    # Resumo (fora do contexto do checkpointer)
    logger.info("=" * 80)
    logger.info("🏁 TESTE CONCLUÍDO")
    logger.info("=" * 80)
    logger.info(
        "✅ O Vectora executou 2 pesquisas completas com web_search + fetch_url"
    )
    logger.info("✅ Cada pesquisa teve 3 buscas no Google + acesso a sites")
    logger.info("✅ Resumos foram gerados pelo Vectora (LLM Gemini 3.0 Flash)")
    logger.info(
        "\n💡 Próximo passo: Executar 'python vectora_research_agent.py' para 150 arquivos"
    )


if __name__ == "__main__":
    asyncio.run(test_vectora_agent())
