# Vectora

O **Vectora** é um projeto open-source (licença Apache 2.0) e self-hosted.
Ele está sendo desenvolvido para ser altamente flexível em sua infraestrutura e armazenamento, integrando tecnologias de ponta em agentes e inteligência artificial.

## O Problema que o Vectora Resolve

O **Vectora** tem o **RAG (Retrieval-Augmented Generation)** em seu coração. Modelos de IA frequentemente sofrem de defasagem de conhecimento e não foram treinados nas versões mais recentes da sua _stack_ (como as últimas atualizações do Next.js, Hono, Bun, Playwright, Axios, Zustand, TypeScript, etc.).

Através do Vectora, a sua IA passa a dominar essas tecnologias de forma instantânea. Basta fornecer a documentação desses projetos para o nosso **Vector Bucket**, e o próprio Vectora realiza a ingestão e indexação automática para você. No futuro, lançaremos uma **Asset Library** para que a comunidade possa publicar e baixar _buckets_ completos com um clique.

Além do conhecimento externo, o Vectora resolve o **problema de contexto da sua própria aplicação**, injetando na IA um entendimento profundo de como cada ponto da sua arquitetura local funciona.

## Arquitetura de Sub-Agent via MCP

O Vectora pode operar de forma autônoma como um Agente Principal, mas **o nosso foco principal é atuar como um Sub-Agent especializado**.

Utilizando o **MCP (Model Context Protocol)**, um Agente Principal (como um assistente integrado ou uma IA orquestradora) executa o Vectora e passa para ele a _session_ atual. Isso permite que o Vectora retome o contexto exato de onde a conversa parou, execute o RAG profundo e devolva a resposta extremamente embasada para o Agente Principal tomar a decisão final.

## Stack Tecnológica

- **Backend / Linguagem:** Python gerenciado pelo [UV](https://github.com/astral-sh/uv)
- **API:** [FastAPI](https://fastapi.tiangolo.com/)
- **UI no Terminal:** [Rich](https://rich.readthedocs.io/) + [Textual](https://textual.textualize.io/)
- **Orquestração de LLMs:** LangChain + LangGraph (Grafos, Fluxos e Memory)
- **Busca em Tempo Real:** [Meilisearch](https://www.meilisearch.com/)
- **Cache:** [Valkey](https://valkey.io/)
- **Memória de Longo Prazo:** PostgreSQL
- **Vector Store (RAG):** Camada de abstração com dois provedores (LanceDB e Qdrant)

---

## Modelos e IA

O Vectora conta com a **VoyageAI** como seu motor principal para:

- **Embeddings:** Geração de embeddings vetoriais para representação semântica robusta de textos e imagens.
- **Reranker:** Reordenação de resultados de busca com alta performance e precisão.

## Deep Agents

Desenvolvido com inspiração na equipe do LangChain / LangGraph, o projeto apresenta capacidades profundas de agentes:

- **ACP (Agent Client Protocol):** Padroniza a integração entre o agente e IDEs (Zed, JetBrains, Neovim) para edições de código contextuais.
- **Harness:** SDK com ferramentas nativas de planejamento (TODOs), sistema de arquivos e delegação para tarefas de longa duração.
- **CLI:** Interface de terminal para execução local de pesquisas, automações e assistência de codificação.
- **Skills:** Injeção dinâmica de instruções apenas quando necessário, otimizando a janela de contexto.
- **Context Management:** Middleware para sumarização de histórico e offloading de dados pesados para o disco.
- **Deep Research:** Ciclos iterativos de busca e análise profunda para síntese de relatórios complexos.
- **Autonomia:** Orquestração via subagentes especializados, isolando contextos técnicos do objetivo principal.

---

## Bancos de Dados e Search

- **PostgreSQL (Memória de Longo Prazo):** Banco de dados relacional para armazenar históricos de conversas, perfis de usuários e dados relacionais que exigem alta consistência e transações ACID.
- **Vector Store (RAG):** O Vectora implementa uma **Camada de Abstração** via LangChain focada em duas opções de bancos de vetores via variável de ambiente (`VECTOR_STORE_TYPE`):
  - **LanceDB:** A escolha padrão, file-based e altamente performático, ideal para a maioria dos casos "zero config". Não possui suporte nativo ao algoritmo HNSW.
  - **Qdrant:** Alternativa que suporta nativamente **HNSW**, perfeita para alta performance de busca e grandes escalas. Pode rodar 100% local no modo standalone (via driver Python/Rust) ou via container Docker se não puder ser instalado localmente.
- **Meilisearch (Busca):** Fornece recursos avançados de busca em tempo real (tolerância a erros de digitação, busca facetada). **Embarcado**, sem necessidade de container.
- **RAG Híbrido:** Uma grande vantagem competitiva da stack do Vectora é a união de tecnologias:
  - Busca por palavra-chave (BM25) via **Meilisearch**.
  - Busca semântica via **Vector DB** (LanceDB ou Qdrant).
  - Reclassificação inteligente com **VoyageAI Reranker**.
- **Valkey (Cache):** Armazena dados frequentemente acessados em memória. Em Docker, um container próprio é usado. No modo standalone local, o sistema oferece um _fallback_ (em memória/arquivo via pickle/json) caso um servidor Valkey/Redis não esteja configurado.

---

## Flexibilidade de Implantação

O Vectora foi arquitetado para ser flexível, rodando conforme as necessidades da sua infraestrutura:

A arquitetura baseia-se em uma estratégia **"Dual-Mode"**:

1. **Docker Compose (Para Servidores/VPS):** A forma mais simples de iniciar e escalar. Orquestra automaticamente containers isolados e interconectados para a aplicação, PostgreSQL, Valkey, etc.
2. **Instalação Standalone (Para Desenvolvimento Local / Bare Metal):** Instale o serviço principal do Vectora localmente via Python/UV. Remove a barreira de entrada enquanto mantém a robustez.
3. **Dependências Opcionais:** Através de _extras_ do Python no gerenciador de pacotes, os usuários instalam apenas o que precisam (ex: `pip install "vectora[qdrant]"` ou `vectora[full]`).
4. **Bring Your Own DB (BYOD):** Total compatibilidade com serviços gerenciados na nuvem (Supabase, Neon, AWS RDS, etc).
5. **Ferramentas de Migração:** O projeto prevê utilitários fáceis para a migração de dados caso o usuário comece local (ex: LanceDB) e decida escalar para uma infraestrutura dedicada (ex: pgvector no VPS).

## Licença

Este projeto está sob a licença **Apache 2.0**.
