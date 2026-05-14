# Vectora

O **Vectora** é um agente de IA open-source (licença Apache 2.0), self-hosted e **local-first**.
Projetado para ser instalado com um comando (`uv tool install vectora`) ou via Docker — sem serviços externos, sem configuração de infraestrutura pesada.

## O Problema que o Vectora Resolve

O **Vectora** tem o **RAG (Retrieval-Augmented Generation)** em seu coração. Modelos de IA frequentemente sofrem de defasagem de conhecimento e não foram treinados nas versões mais recentes da sua _stack_ (como as últimas atualizações do Next.js, Hono, Bun, Playwright, Axios, Zustand, TypeScript, etc.).

Através do Vectora, a sua IA passa a dominar essas tecnologias de forma instantânea. Basta fornecer a documentação desses projetos para o nosso **Vector Bucket**, e o próprio Vectora realiza a ingestão e indexação automática para você.

Além do conhecimento externo, o Vectora resolve o **problema de contexto da sua própria aplicação**, injetando na IA um entendimento profundo de como cada ponto da sua arquitetura local funciona.

## Arquitetura de Sub-Agent via MCP

O Vectora foi projetado para brilhar primariamente como um **Sub-Agent especializado**.

Utilizando o **MCP (Model Context Protocol)**, um Agente Principal (como o Paperclip, um assistente integrado na sua IDE ou uma IA orquestradora) executa o Vectora e utiliza suas ferramentas de busca semântica. Isso permite que o Vectora forneça o contexto exato necessário para que o Agente Principal prossiga com a tarefa com precisão cirúrgica.

---

## Pré-requisitos

O Vectora suporta múltiplos provedores de LLM e recomenda usar **Gemini + Voyage AI** por oferecerem **APIs gratuitas com excelentes limites**:

### Provedores de LLM Suportados

| Provider          | API Key                                                       | Modelo Padrão    | Recomendação               |
| ----------------- | ------------------------------------------------------------- | ---------------- | -------------------------- |
| **Google Gemini** | [Obter gratuitamente](https://aistudio.google.com/app/apikey) | gemini-3.0-flash | ✅ Recomendado (free tier) |
| **Ollama**        | Nenhuma (local)                                               | gpt-oss:20b      | Local sem custos           |
| **OpenAI**        | [Obter aqui](https://platform.openai.com/api-keys)            | gpt-4o           | Pago                       |
| **Anthropic**     | [Obter aqui](https://console.anthropic.com/)                  | claude-opus-4-1  | Pago                       |

### Embedding (RAG) - Obrigatório

| Provider      | API Key                                             | Recomendação                                |
| ------------- | --------------------------------------------------- | ------------------------------------------- |
| **Voyage AI** | [Obter gratuitamente](https://www.voyageai.com/api) | ✅ Recomendado (free tier com bons limites) |

---

## Instalação

Escolha uma das duas opções abaixo:

### Opção 1: Instalação via UV (Recomendado)

Requer apenas o [uv](https://github.com/astral-sh/uv) instalado.

```bash
# Instale o Vectora globalmente
uv tool install vectora

# Configure suas chaves (será criado em ~/.vectora/)
vectora setup

# Execute o chat TUI
vectora chat
```

### Opção 2: Instalação via Docker

```bash
# Crie um arquivo .env com suas configurações
cp .env.example .env
# Edite .env com suas API keys

# Build da imagem
docker build -t vectora:0.1.0 .

# Execute o container
docker run -it \
  --env-file .env \
  -v ~/.vectora:/root/.vectora \
  vectora:0.1.0

# Ou use docker-compose (lê .env automaticamente)
docker-compose up
```

O Vectora cria todos os arquivos necessários no diretório `~/.vectora/` automaticamente:

- Banco de dados de sessões (SQLite)
- Vector store para RAG (LanceDB)
- Logs estruturados
- Configurações

---

## Servidor MCP (Model Context Protocol)

O Vectora pode rodar como um servidor de contexto para outros agentes (como o **Paperclip** ou **Claude Desktop**).

### Configuração de Ambiente

Antes de executar, certifique-se de que o arquivo `.env` está configurado com as chaves necessárias:

```bash
cp .env.example .env
# Edite .env com suas API keys (LLM_PROVIDER, GOOGLE_API_KEY, VOYAGE_API_KEY, etc)
```

### Via UV:

```bash
vectora mcp-server
```

### Via Docker:

```bash
docker run -it \
  --env-file .env \
  -v ~/.vectora:/root/.vectora \
  vectora:0.1.0 \
  vectora mcp-server
```

O servidor iniciará via **Stdio JSON-RPC**. Para configurar em outros agentes, use as seguintes configurações:

**Exemplo de configuração em Claude Desktop (JSON):**

```json
{
  "mcpServers": {
    "vectora": {
      "command": "uv",
      "args": ["tool", "run", "vectora", "mcp-server"]
    }
  }
}
```

**Ou via Docker:**

```json
{
  "mcpServers": {
    "vectora": {
      "command": "docker",
      "args": [
        "run",
        "-it",
        "-e",
        "GOOGLE_API_KEY=seu-api-key",
        "-v",
        "~/.vectora:/root/.vectora",
        "vectora:0.1.0",
        "vectora",
        "mcp-server"
      ]
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

Toda a persistência do Vectora é baseada em arquivos locais no diretório `~/.vectora/`:

```
~/.vectora/
├── config.toml             # Configurações (não-sensível)
├── .env                    # Variáveis de ambiente (sensível)
├── data/
│   ├── vectora.db          # Histórico de conversas (SQLite via LangGraph checkpointer)
│   ├── embedding_queue.db  # Fila de embedding com retry (SQLite via SQLAlchemy async)
│   └── lancedb/            # Vector Store para RAG semântico
├── logs/                   # Logs estruturados
│   ├── vectora.log
│   ├── mcp_client.log
│   └── mcp_server.log
└── keys/                   # Chaves de API encriptadas (opcional)
```

**SQLite** (`aiosqlite` + `langgraph-checkpoint-sqlite`):

- Armazena o histórico completo de todas as conversas com suporte a time-travel.
- Gerencia a fila de embedding com retry automático em caso de falha de API.

**LanceDB** (file-based, columnar):

- Vector store para busca semântica via RAG de alta performance.
- Ingestão automática via ferramenta `ingest_docs`.
- Reranking integrado via VoyageAI.

**Logs e Configuração:**

- Logs estruturados em JSON para análise e debugging
- Configurações salvas em TOML para fácil edição
- Suporte a LangSmith para rastreamento de execução

---

## Licença

Este projeto está sob a licença **Apache 2.0**.
