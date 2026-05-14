# Vectora

O **Vectora** é um agente de IA open-source (licença Apache 2.0), self-hosted e **local-first**.
Projetado para rodar com `uv sync` e nada mais — sem containers, sem serviços externos, sem configuração de infraestrutura.

## O Problema que o Vectora Resolve

O **Vectora** tem o **RAG (Retrieval-Augmented Generation)** em seu coração. Modelos de IA frequentemente sofrem de defasagem de conhecimento e não foram treinados nas versões mais recentes da sua _stack_ (como as últimas atualizações do Next.js, Hono, Bun, Playwright, Axios, Zustand, TypeScript, etc.).

Através do Vectora, a sua IA passa a dominar essas tecnologias de forma instantânea. Basta fornecer a documentação desses projetos para o nosso **Vector Bucket**, e o próprio Vectora realiza a ingestão e indexação automática para você.

Além do conhecimento externo, o Vectora resolve o **problema de contexto da sua própria aplicação**, injetando na IA um entendimento profundo de como cada ponto da sua arquitetura local funciona.

## Arquitetura de Sub-Agent via MCP

O Vectora foi projetado para brilhar primariamente como um **Sub-Agent especializado**.

Utilizando o **MCP (Model Context Protocol)**, um Agente Principal (como um assistente integrado na sua IDE ou uma IA orquestradora) executa o Vectora e passa para ele a _session_ atual. Isso permite que o Vectora retome o contexto exato de onde a conversa parou, execute o RAG profundo na sua base de conhecimento especializada e devolva a resposta embasada para o Agente Principal prosseguir com a tarefa.

## Vectora Code (CLI Principal e Integração ACP)

Apesar da sua forte vocação como Sub-Agent, nós também oferecemos o **Vectora Code**.

O Vectora Code é a nossa interface CLI completa, onde o Vectora atua como o Agente Principal. Ele possui acesso a um conjunto de ferramentas (_Tools_) muito mais amplo do que no modo Sub-Agent, sendo capaz de planejar e executar workflows autônomos inteiros.

Além do uso puro via terminal, toda a potência do Vectora Code também está disponível para ser integrada nativamente em IDEs modernas (como Zed, Neovim, etc.) através do **ACP (Agent Client Protocol)**.

---

## Instalação

O Vectora requer apenas o [uv](https://github.com/astral-sh/uv) instalado.

```bash
# 1. Clone o repositório
git clone https://github.com/seu-usuario/vectora
cd vectora

# 2. Instale as dependências
uv sync

# 3. Configure suas chaves
cp .env.example .env
# Edite .env com sua chave de API de LLM

# 4. Execute o chat TUI
uv run python src/run_chat.py
```

Nenhum Docker, nenhum banco de dados externo, nenhum container. O Vectora cria todos os arquivos necessários no diretório `data/` automaticamente.

---

## Stack Tecnológica

- **Backend / Linguagem:** Python gerenciado pelo [UV](https://github.com/astral-sh/uv)
- **UI no Terminal:** [Rich](https://rich.readthedocs.io/) + [Textual](https://textual.textualize.io/)
- **Orquestração de LLMs:** LangChain + LangGraph (Grafos, Fluxos e Memory)
- **Vector Store (RAG):** [LanceDB](https://lancedb.github.io/lancedb/) — file-based, zero-config
- **Persistência de Sessões:** SQLite via `aiosqlite` — file-based, zero-config

---

## Modelos e IA

O Vectora é compatível com os principais provedores de LLM:

| Provedor       | Variável            | Modelo padrão      |
| -------------- | ------------------- | ------------------ |
| Google Gemini  | `GOOGLE_API_KEY`    | `gemini-2.0-flash` |
| OpenAI         | `OPENAI_API_KEY`    | `gpt-4o`           |
| Anthropic      | `ANTHROPIC_API_KEY` | `claude-opus-4-5`  |
| Ollama (local) | `OLLAMA_BASE_URL`   | `llama3.2`         |

Para RAG com embeddings, o Vectora usa **VoyageAI** (`VOYAGE_API_KEY`) para geração de embeddings e reranking.

---

## Persistência Local (File-Based)

Toda a persistência do Vectora é baseada em arquivos locais. O diretório `data/` é criado automaticamente:

```
data/
├── vectora.db              # Histórico de conversas (SQLite via LangGraph checkpointer)
├── embedding_queue.db      # Fila de embedding com retry (SQLite via SQLAlchemy async)
└── lancedb/                # Vector Store para RAG semântico
    ├── articles/           # Coleção padrão
    ├── wiki/
    ├── api_docs/
    └── knowledge_base/
```

**SQLite** (`aiosqlite` + `langgraph-checkpoint-sqlite`):

- Armazena o histórico completo de todas as conversas com suporte a time-travel
- Gerencia a fila de embedding com retry automático em caso de falha de API
- Totalmente assíncrono, sem bloqueio do event loop

**LanceDB** (file-based, columnar):

- Vector store para busca semântica via RAG
- Cria tabelas por coleção automaticamente na primeira indexação
- Suporte nativo a embeddings via pyarrow schema
- Busca vetorial por similaridade de cosseno

---

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

## Licença

Este projeto está sob a licença **Apache 2.0**.
