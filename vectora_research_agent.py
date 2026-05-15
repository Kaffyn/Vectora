#!/usr/bin/env python3
"""Vectora Research Agent - Cria 150 arquivos com pesquisa profunda via web_search + fetch_url.

Este script invoca o Vectora como um agente autônomo que:
1. Recebe uma lista de 15 tópicos (cada um gera 10 arquivos)
2. Para cada arquivo: faz 10 web_searches, acessa sites e cria sumário
3. Organiza arquivos por tópico em diretório output/research/

Exemplos de tópicos: Next.js 16, Godot 4.7, Hono, FastAPI 0.120, etc.

USAGE:
    python vectora_research_agent.py

OUTPUT:
    output/research/
    ├── next-js-16/
    │   ├── 01_intro.md
    │   ├── 02_features.md
    │   └── ...
    ├── godot-47/
    │   ├── 01_intro.md
    │   └── ...
    └── ...
"""

import asyncio
import logging
from datetime import UTC, datetime
from pathlib import Path

# Configurar logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

# Importar componentes Vectora
import sys

sys.path.insert(0, "src")

from context import Context
from graph import build_graph

# ============================================================================
# TÓPICOS PARA PESQUISA (15 tópicos × 10 arquivos = 150 arquivos)
# ============================================================================

RESEARCH_TOPICS = [
    {
        "slug": "next-js-16",
        "name": "Next.js 16",
        "description": "Framework React com SSR, ISR e API Routes",
        "search_queries": [
            "Next.js 16 new features 2025",
            "Next.js 16 performance improvements",
            "Next.js 16 app router guide",
            "Next.js 16 database integration",
            "Next.js 16 deployment best practices",
            "Next.js 16 vs alternatives comparison",
            "Next.js 16 API development tutorial",
            "Next.js 16 middleware and authentication",
            "Next.js 16 image optimization",
            "Next.js 16 caching strategies",
        ],
    },
    {
        "slug": "godot-47",
        "name": "Godot 4.7",
        "description": "Engine de desenvolvimento de jogos open-source",
        "search_queries": [
            "Godot 4.7 new features release",
            "Godot 4.7 3D development guide",
            "Godot 4.7 GDScript improvements",
            "Godot 4.7 physics engine updates",
            "Godot 4.7 performance optimization",
            "Godot 4.7 graphics and rendering",
            "Godot 4.7 multiplayer networking",
            "Godot 4.7 asset pipeline workflow",
            "Godot 4.7 game examples and tutorials",
            "Godot 4.7 vs Unreal Engine comparison",
        ],
    },
    {
        "slug": "hono-js",
        "name": "Hono",
        "description": "Web framework ultrarrápido para Edge Computing",
        "search_queries": [
            "Hono framework JavaScript guide",
            "Hono vs Express comparison 2025",
            "Hono Cloudflare Workers deployment",
            "Hono routing and middleware",
            "Hono database integration examples",
            "Hono testing and validation",
            "Hono performance benchmarks",
            "Hono TypeScript support",
            "Hono authentication patterns",
            "Hono RESTful API development",
        ],
    },
    {
        "slug": "fastapi-0120",
        "name": "FastAPI 0.120",
        "description": "Framework web Python moderno com async/await",
        "search_queries": [
            "FastAPI 0.120 changelog features",
            "FastAPI async database querying",
            "FastAPI dependency injection system",
            "FastAPI OpenAPI/Swagger documentation",
            "FastAPI authentication and security",
            "FastAPI testing and debugging",
            "FastAPI background tasks",
            "FastAPI deployment on production",
            "FastAPI WebSocket support",
            "FastAPI vs Django REST Framework",
        ],
    },
    {
        "slug": "rust-web-2025",
        "name": "Rust Web Development 2025",
        "description": "Desenvolvimento web com Rust e frameworks modernos",
        "search_queries": [
            "Rust Axum web framework 2025",
            "Rust WebAssembly and frontend development",
            "Rust async runtime tokio comparison",
            "Rust web development security best practices",
            "Rust database ORMs sqlx vs diesel",
            "Rust testing frameworks and practices",
            "Rust web API development tutorial",
            "Rust performance optimization web apps",
            "Rust error handling in web services",
            "Rust microservices architecture",
        ],
    },
    {
        "slug": "openai-api-2025",
        "name": "OpenAI API 2025",
        "description": "LLMs, embeddings e AI assistants com OpenAI",
        "search_queries": [
            "OpenAI API 2025 models and pricing",
            "OpenAI ChatGPT API fine-tuning guide",
            "OpenAI embeddings for semantic search",
            "OpenAI function calling integration",
            "OpenAI prompt engineering best practices",
            "OpenAI multi-turn conversations handling",
            "OpenAI safety and content moderation",
            "OpenAI API rate limiting and quotas",
            "OpenAI batch processing endpoints",
            "OpenAI vs other LLM providers comparison",
        ],
    },
    {
        "slug": "react-19",
        "name": "React 19",
        "description": "Library JavaScript para interfaces de usuário",
        "search_queries": [
            "React 19 new features and APIs",
            "React 19 server components guide",
            "React 19 hooks best practices",
            "React 19 performance optimization",
            "React 19 state management patterns",
            "React 19 forms and validation",
            "React 19 testing strategies",
            "React 19 styling solutions",
            "React 19 accessibility improvements",
            "React 19 migration guide",
        ],
    },
    {
        "slug": "kubernetes-2025",
        "name": "Kubernetes 2025",
        "description": "Orquestração de containers e cloud native",
        "search_queries": [
            "Kubernetes 1.32 features and updates",
            "Kubernetes cluster architecture design",
            "Kubernetes networking and service mesh",
            "Kubernetes persistent storage solutions",
            "Kubernetes security best practices",
            "Kubernetes monitoring and observability",
            "Kubernetes cost optimization strategies",
            "Kubernetes GitOps and ArgoCD",
            "Kubernetes multi-cluster management",
            "Kubernetes vs Docker Swarm comparison",
        ],
    },
    {
        "slug": "postgres-17",
        "name": "PostgreSQL 17",
        "description": "Banco de dados relacional avançado",
        "search_queries": [
            "PostgreSQL 17 new features release",
            "PostgreSQL 17 performance improvements",
            "PostgreSQL JSON and document capabilities",
            "PostgreSQL full-text search optimization",
            "PostgreSQL replication and high availability",
            "PostgreSQL backup and recovery strategies",
            "PostgreSQL query optimization tips",
            "PostgreSQL monitoring and diagnostics",
            "PostgreSQL security hardening",
            "PostgreSQL scaling and partitioning",
        ],
    },
    {
        "slug": "llm-fine-tuning",
        "name": "LLM Fine-tuning 2025",
        "description": "Treinamento customizado de modelos de linguagem",
        "search_queries": [
            "LLM fine-tuning techniques and best practices",
            "LoRA and parameter efficient tuning",
            "QLoRA for consumer GPU fine-tuning",
            "Fine-tuning dataset preparation and curation",
            "Fine-tuning evaluation metrics and benchmarks",
            "Open-source models fine-tuning guide",
            "Fine-tuning for domain specific tasks",
            "Fine-tuning safety and RLHF alignment",
            "Fine-tuning inference optimization",
            "Fine-tuning cost analysis and ROI",
        ],
    },
    {
        "slug": "webassembly-2025",
        "name": "WebAssembly 2025",
        "description": "Executar código compilado no browser e edge",
        "search_queries": [
            "WebAssembly 2025 features and standards",
            "WebAssembly vs JavaScript performance",
            "WebAssembly Rust compilation toolchain",
            "WebAssembly in browser optimization",
            "WebAssembly for game development",
            "WebAssembly Node.js server-side usage",
            "WebAssembly debugging and profiling",
            "WebAssembly memory management",
            "WebAssembly thread support and workers",
            "WebAssembly production deployment patterns",
        ],
    },
    {
        "slug": "tailwind-css-4",
        "name": "Tailwind CSS 4",
        "description": "Framework CSS utilitário moderno",
        "search_queries": [
            "Tailwind CSS 4 new features 2025",
            "Tailwind CSS utility-first approach guide",
            "Tailwind CSS responsive design patterns",
            "Tailwind CSS component customization",
            "Tailwind CSS dark mode implementation",
            "Tailwind CSS animation and transitions",
            "Tailwind CSS with React best practices",
            "Tailwind CSS accessibility guidelines",
            "Tailwind CSS performance optimization",
            "Tailwind CSS vs other CSS frameworks",
        ],
    },
    {
        "slug": "docker-container",
        "name": "Docker Container 2025",
        "description": "Containerização e deployment de aplicações",
        "search_queries": [
            "Docker 27 new features and improvements",
            "Docker image optimization best practices",
            "Docker networking for multi-container apps",
            "Docker volumes and persistent storage",
            "Docker security hardening guide",
            "Docker debugging and troubleshooting",
            "Docker registry and image distribution",
            "Docker compose for development environments",
            "Docker for CI/CD pipelines",
            "Docker resource limits and optimization",
        ],
    },
    {
        "slug": "rag-systems-2025",
        "name": "RAG Systems 2025",
        "description": "Retrieval-Augmented Generation com LLMs",
        "search_queries": [
            "RAG retrieval-augmented generation architecture",
            "RAG vector embeddings and semantic search",
            "RAG document chunking strategies",
            "RAG vector databases comparison (Qdrant, Milvus)",
            "RAG reranking and ranking models",
            "RAG with LLMs prompt engineering",
            "RAG evaluation metrics and benchmarks",
            "RAG for chatbots and QA systems",
            "RAG hallucination reduction techniques",
            "RAG production deployment patterns",
        ],
    },
    {
        "slug": "prompt-engineering",
        "name": "Prompt Engineering 2025",
        "description": "Otimização de prompts para LLMs",
        "search_queries": [
            "Prompt engineering best practices 2025",
            "Few-shot learning and examples",
            "Chain-of-thought prompting techniques",
            "Prompt injection and security risks",
            "Prompt optimization and testing frameworks",
            "Prompt templating for consistency",
            "Prompt engineering for specialized domains",
            "Vision models and multimodal prompts",
            "Prompt evaluation and metrics",
            "Advanced prompting techniques like ReAct",
        ],
    },
]


# ============================================================================
# PROMPT PARA VECTORA EXECUTAR A TAREFA
# ============================================================================


def create_research_prompt(topic: dict) -> str:
    """Cria um prompt estruturado para Vectora pesquisar um tópico específico."""
    queries = "\n".join(f"  - {q}" for q in topic["search_queries"])

    return f"""TAREFA: Pesquisa Profunda e Criação de Documentação

TÓPICO: {topic["name"]}
DESCRIÇÃO: {topic["description"]}

INSTRUÇÕES:
1. Execute as 10 buscas no Google em sequência:
{queries}

2. Para CADA busca:
   - Use web_search() para encontrar resultados
   - Use fetch_url() para acessar os 3 primeiros links
   - Extraia informações relevantes

3. APÓS todas as buscas, crie 10 arquivos Markdown bem estruturados:
   - 01_introducao.md — Overview geral do {topic["name"]}
   - 02_historia_evolucao.md — Evolução e histórico
   - 03_recursos_principais.md — Principais features
   - 04_arquitetura.md — Arquitetura técnica
   - 05_instalacao_setup.md — Como começar
   - 06_exemplos_praticos.md — Exemplos de código
   - 07_performance.md — Performance e otimização
   - 08_seguranca.md — Considerações de segurança
   - 09_comparacao_alternativas.md — vs alternativas
   - 10_roadmap_futuro.md — Roadmap e futuro

4. Cada arquivo deve ter:
   - Entre 800-1500 palavras
   - Markdown com headers (h2, h3)
   - Listas e code blocks quando apropriado
   - Referências aos sites que consultou

5. Organize em: output/research/{topic["slug"]}/

FORMATO DE RESPOSTA:
Quando terminar cada arquivo, confirme:
✅ Arquivo: output/research/{topic["slug"]}/NM_nome.md (X palavras)

Quando terminar o tópico completo:
🎯 Tópico '{topic["name"]}' completado! 10 arquivos criados em output/research/{topic["slug"]}/
"""


async def orchestrate_research():
    """Orquestra a pesquisa com Vectora - cria 150 arquivos com 15 tópicos."""
    logger.info("🚀 Iniciando Vectora Research Agent")
    logger.info(f"   Tópicos: {len(RESEARCH_TOPICS)}")
    logger.info(f"   Arquivos esperados: {len(RESEARCH_TOPICS) * 10}")

    # Criar diretório de saída
    output_dir = Path("output/research")
    output_dir.mkdir(parents=True, exist_ok=True)

    # Configurar Vectora
    context = Context(
        thread_id="research-agent",
        user_id="research",
        user_type="researcher",
    )

    graph = build_graph()
    logger.info("✅ Grafo LangGraph do Vectora carregado")

    # Processar cada tópico
    total_files = 0
    completed_topics = 0

    for i, topic in enumerate(RESEARCH_TOPICS, 1):
        logger.info(f"\n{'=' * 80}")
        logger.info(f"[{i}/{len(RESEARCH_TOPICS)}] Pesquisando: {topic['name']}")
        logger.info(f"{'=' * 80}")

        # Preparar tópico
        topic_dir = output_dir / topic["slug"]
        topic_dir.mkdir(exist_ok=True)

        prompt = create_research_prompt(topic)

        try:
            # Invocar Vectora com o prompt de pesquisa
            logger.info("📝 Prompt enviado para Vectora... (3 searches + summarize)")

            config = {"configurable": {"thread_id": context.thread_id}}

            result = await graph.ainvoke(
                {
                    "messages": [
                        {"role": "user", "content": prompt},
                    ]
                },
                config=config,
            )

            # Processar resposta
            if result and "messages" in result:
                last_message = result["messages"][-1]
                response = (
                    last_message.content
                    if hasattr(last_message, "content")
                    else str(last_message)
                )

                logger.info(
                    f"📖 Resposta Vectora (primeiros 500 chars):\n{response[:500]}..."
                )

                # Contar arquivos criados (simplificado - conta linhas com "Arquivo:")
                files_created = response.count("Arquivo:") + response.count("arquivo:")
                total_files += files_created

                completed_topics += 1
                logger.info(f"✅ {topic['name']}: {files_created} arquivos criados")
            else:
                logger.warning(f"⚠️  Resposta vazia para {topic['name']}")

        except Exception as e:
            logger.error(f"❌ Erro ao processar {topic['name']}: {e}", exc_info=True)

    # Resumo final
    logger.info(f"\n{'=' * 80}")
    logger.info("🏁 PESQUISA CONCLUÍDA")
    logger.info(f"{'=' * 80}")
    logger.info(f"✅ Tópicos completados: {completed_topics}/{len(RESEARCH_TOPICS)}")
    logger.info(
        f"📄 Arquivos criados: ~{total_files} (esperado: ~{len(RESEARCH_TOPICS) * 10})"
    )
    logger.info(f"📁 Diretório: {output_dir.absolute()}")
    logger.info(f"⏱️  Timestamp: {datetime.now(UTC).isoformat()}")


if __name__ == "__main__":
    asyncio.run(orchestrate_research())
