# 🔍 Observability Guide - Vectora v0.1.0

**Objetivo:** Rastreabilidade completa de uma request através da pipeline Fire-and-Forget.

---

## 📊 Conceito: Correlation ID

Cada request (user input) recebe um **UUID único** (`correlation_id`) que a segue por toda a pipeline:

```
User Input
    ↓ [correlation_id: abc-123]
MAIN_NODE (chat processing)
    ↓ [correlation_id: abc-123]
web_search() → enfileira embeddings
    ↓ [correlation_id: abc-123]
BackgroundWorker processa
    ↓ [correlation_id: abc-123]
LanceDB indexa documentos
    ↓ [correlation_id: abc-123]
Response para o usuário
```

---

## 🔗 Fluxo de Observabilidade Completo

### Exemplo: User pesquisa "Next.js 16"

**Timestamp:** 2026-05-14T14:30:00Z  
**Correlation ID:** `550e8400-e29b-41d4-a716-446655440000`

### Logs Esperados

```json
// LOG 1: Chat inicia processamento
{
  "timestamp": "2026-05-14T14:30:00Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "main_node_started",
  "user_input": "Pesquise sobre Next.js 16",
  "user_id": "user123",
  "thread_id": 1
}

// LOG 2: Web search iniciado
{
  "timestamp": "2026-05-14T14:30:01Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "web_search_started",
  "query": "Next.js 16 new features 2025",
  "results_count": 10
}

// LOG 3: URLs fetched
{
  "timestamp": "2026-05-14T14:30:03Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "fetch_url_completed",
  "url": "https://nextjs.org/blog/next-16",
  "content_length": 5432,
  "latency_ms": 1200
}

// LOG 4: Embeddings enfileirados (Fire-and-Forget)
{
  "timestamp": "2026-05-14T14:30:04Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "embedding_enqueued",
  "queue_id": "queue-001",
  "document_count": 5,
  "collection": "articles"
}

// LOG 5: Response enviada ao user (TUI responsiva)
{
  "timestamp": "2026-05-14T14:30:04.5Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "response_sent",
  "response_time_ms": 4500,
  "status": "fire_and_forget",
  "message": "Salvos na fila de index. Background worker está processando..."
}

// LOG 6: Background Worker inicia processamento (alguns segundos depois)
{
  "timestamp": "2026-05-14T14:30:09Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "background_worker_started",
  "queue_id": "queue-001",
  "batch_size": 5
}

// LOG 7: Embedding gerado
{
  "timestamp": "2026-05-14T14:30:11Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "embedding_processed",
  "queue_id": "queue-001",
  "document_id": "doc-001",
  "vector_dimension": 1024,
  "latency_ms": 2500
}

// LOG 8: Documento escrito em LanceDB
{
  "timestamp": "2026-05-14T14:30:11.5Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "event": "lancedb_document_written",
  "queue_id": "queue-001",
  "collection": "articles",
  "document_count": 5
}

// LOG 9: Follow-up question (5 minutos depois)
{
  "timestamp": "2026-05-14T14:35:00Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655441111",  // Novo correlation_id
  "event": "vector_search_executed",
  "query": "O que você aprendeu sobre Next.js 16?",
  "collection": "articles",
  "results_count": 3,
  "latency_ms": 150
}

// LOG 10: Response baseada em dados indexados
{
  "timestamp": "2026-05-14T14:35:00.5Z",
  "correlation_id": "550e8400-e29b-41d4-a716-446655441111",
  "event": "response_sent",
  "response_time_ms": 500,
  "source": "vector_search",
  "referenced_documents": ["doc-001", "doc-002", "doc-003"]
}
```

---

## 📈 Métricas Rastreadas

### Por Node

| Node                | Métrica          | Esperado |
| ------------------- | ---------------- | -------- |
| MAIN_NODE           | Response time    | <1s      |
| web_search          | Latência         | <3s      |
| fetch_url           | Latência por URL | <2s      |
| embedding (enqueue) | Latência         | <200ms   |
| BackgroundWorker    | Throughput       | 5 docs/s |
| vector_search       | Latência         | <100ms   |

### Performance Budget

```
┌─ Total Request Latency: <5 segundos
│
├─ MAIN_NODE: 0.1s
├─ web_search: 2.5s
├─ fetch_url: 1.5s
├─ embedding (enqueue): 0.2s
└─ Response: 0.7s
  ──────────
   Total: 5.0s ✓ (Fire-and-Forget)
```

---

## 🔗 Rastreabilidade: Como Debugar com Correlation ID

### Caso de Uso 1: User reporta "Resposta incorreta"

**Passos:**

1. User fornece o **correlation_id** (copiar de log)
2. Buscar todos os logs com esse correlation_id:
   ```bash
   grep "550e8400-e29b-41d4-a716-446655440000" logs/mcp.log
   ```
3. Analisar a cadeia:
   - Qual web_search foi feito?
   - Quais URLs foram fetched?
   - Qual embedding foi gerado?
   - O vector_search encontrou o documento?

### Caso de Uso 2: "Documento não aparece em vector_search"

**Rastreamento:**

1. Encontrar o `queue_id` quando foi enfileirado
2. Rastrear:
   - ✓ embedding_enqueued (foi à fila)
   - ✓ embedding_processed (foi processado)
   - ✓ lancedb_document_written (foi escrito)
   - ✓ vector_search_executed (foi encontrado)

Se faltou um passo, a culpa está naquele nó específico.

### Caso de Uso 3: "Sistema travou"

**Investigação:**

1. Pegar o último correlation_id antes do travamento
2. Rastrear se foi para DLQ:
   ```bash
   grep "embedding_moved_to_dlq" logs/mcp.log
   ```
3. Ver o `dlq_reason` (stack trace completo)

---

## 📋 Checklist de Observabilidade para v0.1.0

- [x] Context tem correlation_id
- [x] correlation_id é gerado automaticamente se não fornecido
- [x] MAIN_NODE loga com correlation_id
- [x] BackgroundWorker loga com correlation_id
- [x] embedding_queue logs têm correlation_id
- [ ] web_search loga com correlation_id (future improvement)
- [ ] vector_search loga com correlation_id (future improvement)
- [ ] Métricas são coletadas por nó
- [x] Stack traces completos em DLQ
- [x] Debug dump contém logs estruturados

---

## 🛠️ Implementação para Developers

### Adicionar correlation_id a um log:

```python
logger.info(
    "my_event",
    extra={
        "correlation_id": context.correlation_id,
        "custom_field": "value"
    }
)
```

### Acessar context em um nó:

```python
async def my_node(state: State, context: Context) -> dict:
    logger.info(
        "processing",
        extra={"correlation_id": context.correlation_id}
    )
    # ... implementação
    return {}
```

---

## 📊 Dashboard de Observabilidade (Futuro)

Para v0.2.0, considerar:

1. **Dashboard Web** mostrando:

   - Requisições por correlation_id
   - Latência de cada nó
   - Taxa de sucesso/erro
   - Top queries

2. **Alertas**:

   - Latência >5s
   - Taxa de erro >5%
   - Documentos em DLQ

3. **Exportação**:
   - Prometheus metrics
   - ELK Stack integration
   - Jaeger tracing

---

## 🎯 Para QAs: Como Usar Correlation ID

Quando reportar um bug:

1. **Copie o correlation_id:**

   ```bash
   tail -100 logs/mcp.log | grep "correlation_id" | tail -1
   ```

2. **Inclua no relatório:**

   ```markdown
   ## Correlation ID

   550e8400-e29b-41d4-a716-446655440000

   ## Logs Relevantes

   [Cole aqui as linhas com esse correlation_id]
   ```

3. **Gere debug dump:**
   ```bash
   python -m src.debug_dump
   ```

---

**Versão:** 1.0  
**Última atualização:** 2026-05-14  
**Status:** Production Ready ✅
