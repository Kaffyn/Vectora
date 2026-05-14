# 🤖 Vectora Self-Awareness: Como Vectora Responde Sobre Si Próprio

## 📋 Resumo

Vectora é um **assistente RAG-first** que pode, teoricamente, responder perguntas sobre si próprio usando **embeddings de sua própria documentação**. Mas isso requer que a documentação seja **indexada previamente** no vector store.

---

## 🔄 Como Funciona (Fluxo Completo)

### 1️⃣ **Indexação** (Pré-requisito)

Antes de Vectora responder perguntas sobre si, sua documentação deve ser indexada:

```python
# Terminal
vectora chat

# Dentro do Vectora:
user> ingest_docs ./README.md ./MVP_SCOPE.md ./DEPLOYMENT.md
```

Isso:

- Lê os documentos
- Gera embeddings via Voyage AI
- Armazena em LanceDB (vector store local)

### 2️⃣ **Pergunta do Usuário**

```
user> Tell me about Vectora and its capabilities
```

### 3️⃣ **Processamento no LangGraph**

O Vectora passa por um grafo de 4 nós:

```
[MAIN_NODE] - Recebe pergunta
    ↓
[SUMMARIZER_NODE] - Mantém histórico comprimido
    ↓
[TOOL_NODE] - Executa ferramentas em paralelo
    ├─ vector_search("Vectora capabilities")  ← RAG!
    ├─ web_search() (opcional)
    └─ [outras ferramentas]
    ↓
[SUB_NODE] - Processa resultado complexo
    ↓
[LLM] - Gera resposta
    ↓
Response
```

### 4️⃣ **Vector Search Execution**

```python
# No TOOL_NODE:
results = await vector_search(
    query="Vectora capabilities",
    collection="articles",  # onde está indexado
    limit=5
)

# Retorna JSON com:
{
  "status": "success",
  "results": [
    {
      "id": "doc_001",
      "score": 0.92,  # Relevância (0-1)
      "content": "Vectora é um assistente de IA open-source...",
      "metadata": {"source": "README.md", "section": "Core Capabilities"}
    },
    {
      "id": "doc_002",
      "score": 0.88,
      "content": "O Vectora foi projetado para brilhar primariamente como um Sub-Agent...",
      "metadata": {"source": "README.md", "section": "Arquitetura"}
    },
    ...
  ]
}
```

### 5️⃣ **Reranking** (Opcional)

Se configurado, Voyage AI reordena por relevância:

```python
# Antes:
[0.92, 0.88, 0.85, 0.72, 0.65]

# Depois do reranking:
[0.98, 0.91, 0.83, 0.70, 0.68]
```

### 6️⃣ **Resposta Final**

O LLM vê:

```
System Prompt:
> You are Vectora, an advanced AI assistant...
> Use vector_search() for indexed knowledge
> Cite sources when using retrieved documents

Context (injected by RAG):
> Retrieved 5 documents about your capabilities:
> - README.md (score 0.92): "Vectora é um agente de IA open-source..."
> - README.md (score 0.88): "O Vectora foi projetado para brilhar como Sub-Agent..."
> - MVP_SCOPE.md (score 0.85): "O Vectora é um assistente RAG completo..."
> ...

User Query:
> Tell me about Vectora and its capabilities
```

**Resposta gerada:**

```
I am **Vectora**, an advanced AI assistant with RAG (Retrieval-Augmented Generation)
at my core. Based on my documentation:

**Core Capabilities:**
- 🔍 **RAG**: I use vector search to find relevant documents from indexed collections
- 🌐 **Web Search**: Real-time information retrieval via DuckDuckGo
- 📄 **URL Fetching**: Extract and analyze content from URLs
- 🛠️ **Tool Integration**: 11 tools for code manipulation, file operations, and more
- 🔌 **MCP**: As a Model Context Protocol Sub-Agent for Claude Code/Paperclip

**Architecture:**
I'm designed as a Sub-Agent with my own LangGraph (4-node pattern with auto-summarization,
auto-retry, and sliding window history). I maintain local-first persistence using SQLite
and LanceDB.

[Sources: README.md (confidence 0.92), MVP_SCOPE.md (confidence 0.85)]
```

---

## 🎯 Cenários de Resposta

### Cenário 1: Pergunta Bem Indexada

```
user> What is your vector store?
```

**Resultado:** ✅ Resposta com embedding matching

```
Vectora uses LanceDB as its vector store - a file-based,
zero-config database perfect for local-first applications...
[Source: README.md, score: 0.95]
```

---

### Cenário 2: Pergunta Parcialmente Indexada

```
user> How do I deploy Vectora to production?
```

**Resultado:** ✅ Encontra documentação + complementa com knowledge

```
According to my deployment documentation:

1. Build Docker image: `docker build -t vectora:0.1.0 .`
2. Push to GHCR: `docker push ghcr.io/kaffyn/vectora:0.1.0`
3. Deploy to VPS with SSH: `ssh user@vps docker pull...`

[Source: DEPLOYMENT.md, confidence: 0.91]

Additionally, I can help you set up the required secrets and certificates
for your specific infrastructure.
```

---

### Cenário 3: Pergunta Não Indexada

```
user> What is Vectora's code coverage?
```

**Resultado:** ⚠️ Sem RAG match, mas Vectora responde com knowledge

```
I don't have that information in my indexed documentation. However, based on my
training and recent telemetry: Vectora v0.1.0 achieves 80.85% test coverage across
282 tests (228 passing), meeting the MVP requirement of >80%.

For the most current coverage metrics, you can run: pytest --cov=src
```

---

## 🔧 Configuração Necessária

### Para Funcionar Completo:

```bash
# 1. Environment variables obrigatórios
export GOOGLE_API_KEY="your-key"
export VOYAGE_API_KEY="your-key"
export LLM_PROVIDER="google-genai"
export ENABLE_RAG=true

# 2. Iniciar Vectora
vectora chat

# 3. Indexar documentação
ingest_docs ./README.md ./MVP_SCOPE.md ./DEPLOYMENT.md ./VECTORA_SELF_AWARENESS.md

# 4. Fazer pergunta
Tell me about Vectora
```

---

## 📊 Estado Atual (v0.1.0)

| Aspecto             | Status          | Nota                        |
| ------------------- | --------------- | --------------------------- |
| System Prompt       | ✅ Pronto       | Define persona e instruções |
| Vector Search       | ✅ Pronto       | Funciona com LanceDB        |
| Voyage AI Embedding | ✅ Pronto       | Integrado e testado         |
| Reranking           | ✅ Pronto       | BM25 ou Voyage AI           |
| Documentação        | ⚠️ Não Indexada | Precisa de `ingest_docs`    |
| Self-Referencing    | ⚠️ Manual       | Requer pré-indexação        |

---

## 🚀 Próximas Melhorias

### v0.2.0: Smart Self-Awareness

1. **Auto-Indexing** - Vectora indexa sua própria documentação na inicialização
2. **Live Updates** - Quando documentação muda, reindexar automaticamente
3. **Metadata Enrichment** - Adicionar tags, versões, etc aos documentos
4. **Cross-Reference** - Vectora compreende relacionamentos entre docs

### v0.3.0: Meta-Cognition

1. **Versioning** - "Tell me about Vectora v0.1.0 vs v0.2.0"
2. **Architecture Diagrams** - Embeddings de imagens + vector search
3. **Capability Discovery** - "What can you do with web_search tool?"
4. **Source Attribution** - "Show me where you learned that"

---

## 🧪 Teste Você Mesmo

### Quick Test (Sem Indexação Prévia)

```bash
# Vectora responde com system prompt knowledge
user> What are your core tools?

# Resposta (sem RAG, apenas system prompt):
Based on my system prompt, my core tools include:
- vector_search() for semantic search
- web_search() for real-time data
- fetch_url() for content extraction
- file operations (read, edit, grep)
- terminal execution
- MCP tool integration
```

### Full Test (Com Indexação)

```bash
# 1. Setup com .env completo
cp .env.example .env
# [configure GOOGLE_API_KEY e VOYAGE_API_KEY]

# 2. Iniciar chat
python src/run_chat.py

# 3. Indexar
> ingest_docs README.md MVP_SCOPE.md DEPLOYMENT.md

# 4. Testar self-reference
> What is Vectora?
> How am I architected?
> What tools do I have?
> How do I deploy?
> Tell me about my MCP integration
```

---

## 🎓 Conceitos Importantes

### RAG (Retrieval-Augmented Generation)

```
Query → Embedding → Vector Search → Retrieved Docs → LLM Context → Response

Benefits:
✅ Current information (não dependente de knowledge cutoff)
✅ Source attribution (citar documentação)
✅ Reduced hallucination (baseado em fatos)
✅ Local first (nenhum chamada externa para buscar)
```

### Why Vectora is "Self-Aware"

Vectora não é self-aware no sentido filosófico, mas **instrumentalmente**:

- Conhece seu próprio sistema prompt
- Pode buscar (retrieve) documentação sobre si
- Pode reconhecer suas próprias limitações
- Pode aprender sobre si mesmo via RAG

---

## ❓ FAQ

**P: Vectora realmente "entende" a si próprio?**
A: Não de verdade. Ele usa pattern matching (embeddings) + LLM generation. É como um espelho que lê textos sobre si mesmo.

**P: Se eu perguntar "O que você pensa sobre você?", qual é a resposta?**
A: Ele responderia com base no que está indexado:

```
I'm an advanced AI assistant with RAG capabilities, designed as a Sub-Agent
for Claude Code using the Model Context Protocol. My purpose is to provide
semantic search over indexed knowledge combined with real-time web search...
[cita documentation]
```

**P: Isso é circulação de informação?**
A: Sim, parcialmente. Mas é útil porque:

1. Mantém respostas sincronizadas com documentação
2. Permite versioning ("Eu como v0.1" vs "Eu como v0.2")
3. Reduz alucinação

**P: Como isso é diferente de GPT falando sobre si?**
A: GPT foi treinado em dados estatáticos; Vectora **busca dados atuais** de documentação
viva. Se você atualizar o README, Vectora responderá com informação atualizada.

---

## 📚 Referências

- [RAG Fundamentals](https://arxiv.org/abs/2005.11401)
- [LanceDB Docs](https://lancedb.github.io/)
- [Voyage AI Embeddings](https://docs.voyageai.com/)
- [Model Context Protocol](https://modelcontextprotocol.io/)

---

**Status:** ✅ Ready for v0.1.0 (requer pré-indexação manual)  
**Próxima Milestone:** v0.2.0 auto-indexing  
**Última Atualização:** 2026-05-14
