# Vectora

O **Vectora** é um agente de IA open-source (licença Apache 2.0), self-hosted e **local-first**.
Projetado para rodar com `uv sync` e nada mais — sem containers, sem serviços externos, sem configuração de infraestrutura pesada.

## O Problema que o Vectora Resolve

O **Vectora** tem o **RAG (Retrieval-Augmented Generation)** em seu coração. Modelos de IA frequentemente sofrem de defasagem de conhecimento e não foram treinados nas versões mais recentes da sua _stack_ (como as últimas atualizações do Next.js, Hono, Bun, Playwright, Axios, Zustand, TypeScript, etc.).

Através do Vectora, a sua IA passa a dominar essas tecnologias de forma instantânea. Basta fornecer a documentação desses projetos para o nosso **Vector Bucket**, e o próprio Vectora realiza a ingestão e indexação automática para você.

Além do conhecimento externo, o Vectora resolve o **problema de contexto da sua própria aplicação**, injetando na IA um entendimento profundo de como cada ponto da sua arquitetura local funciona.

## Arquitetura de Sub-Agent via MCP

O Vectora foi projetado para brilhar primariamente como um **Sub-Agent especializado**.

Utilizando o **MCP (Model Context Protocol)**, um Agente Principal (como o Paperclip, um assistente integrado na sua IDE ou uma IA orquestradora) executa o Vectora e utiliza suas ferramentas de busca semântica. Isso permite que o Vectora forneça o contexto exato necessário para que o Agente Principal prossiga com a tarefa com precisão cirúrgica.

---

## Instalação

O Vectora requer apenas o [uv](https://github.com/astral-sh/uv) instalado.

```bash
# 1. Clone o repositório
git clone https://github.com/seu-usuario/vectora
cd vectora

# 2. Instale as dependências e configure o ambiente
uv sync

# 3. Configure suas chaves
cp .env.example .env
# Edite .env com sua VOYAGE_API_KEY e chaves de LLM

# 4. Execute o chat TUI
uv run vectora
```

Nenhum Docker, nenhum banco de dados externo, nenhum container. O Vectora cria todos os arquivos necessários no diretório `data/` automaticamente.

---

## Servidor MCP (Model Context Protocol)

O Vectora pode rodar como um servidor de contexto para outros agentes (como o **Paperclip** ou **Claude Desktop**).

### Como rodar:
```bash
uv run vectora-mcp
```
O servidor iniciará via **Stdio**. Para configurar em outros agentes, use as seguintes configurações:

**Exemplo de configuração (JSON):**
```json
{
  "mcpServers": {
    "vectora": {
      "command": "uv",
      "args": ["--directory", "/caminho/para/vectora", "run", "vectora-mcp"]
    }
  }
}
```

---

## Stack Tecnológica

- **Backend / Linguagem:** Python 3.14 gerenciado pelo [UV](https://github.com/astral-sh/uv)
- **UI no Terminal:** [Rich](https://rich.readthedocs.io/) + [Textual](https://textual.textualize.io/)
- **Orquestração de LLMs:** LangChain 0.3 + LangGraph 0.2+
- **Vector Store (RAG):** [LanceDB](https://lancedb.github.io/lancedb/) — file-based, zero-config
- **Persistência de Sessões:** SQLite via `aiosqlite` — file-based, zero-config
- **Protocolo de Contexto:** [MCP](https://modelcontextprotocol.io/) (Model Context Protocol)

---

## Persistência Local (File-Based)

Toda a persistência do Vectora é baseada em arquivos locais no diretório `data/`:

```
data/
├── vectora.db              # Histórico de conversas (SQLite via LangGraph checkpointer)
├── embedding_queue.db      # Fila de embedding com retry (SQLite via SQLAlchemy async)
└── lancedb/                # Vector Store para RAG semântico
```

**SQLite** (`aiosqlite` + `langgraph-checkpoint-sqlite`):
- Armazena o histórico completo de todas as conversas com suporte a time-travel.
- Gerencia a fila de embedding com retry automático em caso de falha de API.

**LanceDB** (file-based, columnar):
- Vector store para busca semântica via RAG de alta performance.
- Ingestão automática via ferramenta `ingest_docs`.
- Reranking integrado via VoyageAI.

---

## Licença

Este projeto está sob a licença **Apache 2.0**.
