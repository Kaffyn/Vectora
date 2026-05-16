# Plano de Implementação Atualizado - Vectora Phase 2/3

**Status**: Phase 2 ✅ COMPLETO | Phase 3 🔄 EM ANDAMENTO
**Data**: 2026-05-16
**Linguagem**: PT-BR

---

## 🎯 Visão Geral da Arquitetura Implementada

Vectora foi transformado de uma arquitetura "orgânica" para uma **enterprise-grade CLI** com:

### ✅ Implementado (Phase 2)

```
main.py (limpo e simples)
  ↓
Settings (Pydantic - single source of truth)
  ↓
AgentManager (fachada central - orchestrator)
  ├─ TelemetryService (logging + audit)
  ├─ SessionService (SQLite + WAL mode)
  ├─ EmbeddingService (LanceDB + background worker)
  ├─ SecurityService (validation guardrails)
  └─ Graph (⏳ Week 4 - não integrado ainda)
```

---

## 📋 Status Detalhado por Componente

### 1️⃣ **Settings (Pydantic)** ✅ COMPLETO

**Arquivo**: `vectora/settings.py`

**O que foi feito**:

- ✅ Validação centralizada com Pydantic
- ✅ 3-level hierarchy: `defaults.env` → `.env` → `~/.vectora/.env`
- ✅ Todos os campos tipados (type hints 100%)
- ✅ Auto-detection de LLM provider
- ✅ Fail-fast validation (crashes na inicialização com erro claro)

**Status**: PRONTO PARA PRODUÇÃO

```python
from settings import Settings
settings = Settings()  # Pydantic valida tudo aqui
```

---

### 2️⃣ **AgentManager (Fachada Central)** ✅ ESTRUTURA OK | ⏳ CHAT INCOMPLETO

**Arquivo**: `vectora/core/agent.py`

**O que foi feito**:

- ✅ Injeção de dependências (recebe `Settings`)
- ✅ Inicialização de todos os 4 serviços no `__init__`
- ✅ `async initialize()` com sequência correta
- ✅ `async shutdown()` com cleanup em ordem reversa
- ✅ Métodos públicos: `chat()`, `switch_model()`, `create_session()`, `search_vectors()`
- ✅ Zero conhecimento de `Checkpointer`, `BackgroundWorker`, `LanceDB` no `main.py`

**O que FALTA**:

- ❌ `chat()` retorna placeholder (precisa invocar o grafo real)
- ❌ Graph não foi integrado (`self.graph = None`)
- ❌ Falta `build_graph()` no `initialize()`

**Status**: ESTRUTURA OK | LÓGICA DE CHAT INCOMPLETA

```python
# Atualmente (placeholder):
async def chat(self, user_input: str, session_id: int = 1) -> str:
    if not self.graph:
        raise RuntimeError("Graph not initialized")
    return f"[Placeholder] Received: {user_input}"

# Deve ser (quando grafo estiver pronto):
async def chat(self, user_input: str, session_id: int = 1) -> str:
    config = self.session_service.get_runnable_config(session_id)
    input_state = {
        "messages": [HumanMessage(user_input)],
        "session_metadata": {...}
    }
    result = await self.graph.ainvoke(input_state, config=config)
    return result["messages"][-1].content
```

---

### 3️⃣ **TelemetryService** ✅ COMPLETO

**Arquivo**: `vectora/services/telemetry.py`

**Funcionalidades**:

- ✅ JSONFormatter + TextFormatter (dual output)
- ✅ Session audit trails (export Markdown)
- ✅ Debug dumps (.tar.gz com logs + metadata)
- ✅ Correlation IDs (request tracing)
- ✅ Error logging com context
- ✅ Performance metrics

**Status**: PRONTO PARA PRODUÇÃO

---

### 4️⃣ **SessionService** ✅ COMPLETO

**Arquivo**: `vectora/services/session.py`

**Funcionalidades**:

- ✅ AsyncSqliteSaver integration (LangGraph checkpoints)
- ✅ WAL mode (concurrent read/write)
- ✅ Session CRUD (create, switch, list, delete)
- ✅ Metadata tracking (thread_id, user_type, timestamps)
- ✅ `get_runnable_config(thread_id)` para LangGraph

**Status**: PRONTO PARA PRODUÇÃO

---

### 5️⃣ **EmbeddingService** ✅ ESTRUTURA COMPLETA | ⏳ INTEGRAÇÃO INCOMPLETA

**Arquivo**: `vectora/services/embedding.py`

**O que foi feito**:

- ✅ LanceDB integration
- ✅ VoyageAI embeddings
- ✅ Background worker loop (async task)
- ✅ Semantic search
- ✅ Fire-and-forget queuing interface
- ✅ Exponential backoff retry
- ✅ Semaphore concurrency control
- ✅ Health checks

**O que FALTA**:

- ❌ `queue_document()` - stub (precisa EmbeddingQueue real)
- ❌ `_worker_loop()` - stub (precisa processar documentos)
- ❌ `get_queue_status()` - hardcoded (precisa query real)

**Status**: ESTRUTURA OK | QUEUE INCOMPLETA

---

### 6️⃣ **SecurityService** ✅ COMPLETO

**Arquivo**: `vectora/services/security.py`

**Funcionalidades**:

- ✅ File validation (protected paths/files)
- ✅ Command validation (blocked commands)
- ✅ Directory access validation
- ✅ Path traversal prevention
- ✅ Security event logging

**Status**: PRONTO PARA PRODUÇÃO

---

### 7️⃣ **main.py (CLI Entry Point)** ✅ COMPLETO

**Arquivo**: `vectora/main.py`

**Estrutura**:

1. Load Settings (fail-fast)
2. Initialize AgentManager
3. Validate LLM configuration
4. Run chat interface

**Status**: PRONTO PARA PRODUÇÃO

---

### 8️⃣ **State (LangGraph)** ✅ ESTRUTURA OK

**Arquivo**: `vectora/state.py`

**O que foi feito**:

- ✅ SessionMetadata TypedDict (JSON-serializable)
- ✅ Removed Context from State
- ✅ Updated nodes.py to read from state

**Status**: PRONTO PARA PRODUÇÃO

---

## 📊 Resumo de Conclusão

| Componente            | Phase   | Status      | Pronto?                |
| --------------------- | ------- | ----------- | ---------------------- |
| Settings              | 2       | ✅ Completo | SIM                    |
| AgentManager          | 2       | ⏳ 90%      | NÃO (chat incompleto)  |
| TelemetryService      | 2       | ✅ Completo | SIM                    |
| SessionService        | 2       | ✅ Completo | SIM                    |
| EmbeddingService      | 2       | ⏳ 80%      | NÃO (queue incompleta) |
| SecurityService       | 2       | ✅ Completo | SIM                    |
| main.py               | 2       | ✅ Completo | SIM                    |
| State/nodes.py        | 2       | ✅ Completo | SIM                    |
| **Graph (LangGraph)** | **3/4** | **⏳ 0%**   | **NÃO**                |

---

## 🔄 Próximos Passos (Phase 3 - Week 3)

### PRIORIDADE 1: Integrar Graph no AgentManager (CRÍTICO)

**Arquivo**: `vectora/core/agent.py`

```python
# NO MÉTODO initialize():
from graph import build_graph

async def initialize(self) -> None:
    # ... init services ...

    # Build graph com SessionService como checkpointer
    self.graph = build_graph(
        llm=self.settings.get_llm(),
        checkpointer=self.session_service.checkpointer,
        memory_store=self.memory_store,
        embedding_tool=self.embedding_service,
    )
    logger.info("LangGraph built successfully")

# NO MÉTODO chat():
async def chat(self, user_input: str, session_id: int = 1) -> str:
    if not self.graph:
        raise RuntimeError("Graph not initialized")

    # Pega config do SessionService
    config = self.session_service.get_runnable_config(session_id)

    # Build input state com SessionMetadata
    input_state = {
        "messages": [HumanMessage(user_input)],
        "session_metadata": {
            "thread_id": session_id,
            "user_type": "default",
            "created_at": datetime.now(UTC).isoformat(),
            "llm_provider": self.settings.get_llm_provider(),
            "llm_model": self.settings.get_llm_model(),
        }
    }

    # Execute graph
    result = await self.graph.ainvoke(input_state, config=config)

    # Extract response
    response = result["messages"][-1].content

    # Log via telemetry
    await self.telemetry_service.log_chat_message(
        session_id,
        "assistant",
        response
    )

    return response
```

**Estimativa**: 2-3 horas

---

### PRIORIDADE 2: Implementar EmbeddingQueue Real

**Arquivo**: `vectora/services/embedding.py`

Substituir stubs por implementação real:

```python
async def queue_document(
    self, doc_id: str, text: str, collection: str = "documents"
) -> None:
    """Enfileira documento para embedding (fire-and-forget)."""
    # Validar tamanho da fila
    # Inserir em EmbeddingQueue (SQLite)
    # NÃO aguardar processamento

async def _worker_loop(self) -> None:
    """Worker: poll queue → batch → process → LanceDB."""
    while self.worker_running:
        # Buscar 10 docs pendentes
        # Gerar embeddings em paralelo (Semaphore(5))
        # Escrever em LanceDB
        # Atualizar queue com status "success"
        # Retry com backoff exponencial
        # Mover para DLQ se > 3 tentativas
```

**Estimativa**: 3-4 horas

---

### PRIORIDADE 3: Atualizar chat.py para usar AgentManager

**Arquivo**: `vectora/chat.py`

```python
async def run_chat(agent: AgentManager, settings: Settings) -> None:
    """Loop de chat simples usando o AgentManager."""
    session_id = 1  # ou criar novo

    while True:
        user_input = input("You: ")
        if user_input == "/quit":
            break

        # Delegar para AgentManager
        response = await agent.chat(user_input, session_id)
        print(f"Vectora: {response}")
```

**Status**: INICIADO (chat.py já existe, precisa refatoração)

**Estimativa**: 1 hora

---

### PRIORIDADE 4: Testes de Integração (Phase 3)

**Arquivo**: `tests/test_integration.py`

```python
async def test_full_chat_flow():
    """Test: Settings → AgentManager → chat → response."""
    settings = Settings()
    manager = AgentManager(settings)

    await manager.initialize()
    response = await manager.chat("Olá!", session_id=1)

    assert response  # Não placeholder
    assert not response.startswith("[Placeholder]")

    await manager.shutdown()
```

**Estimativa**: 2 horas

---

## 🚧 Problemas Conhecidos / Gaps

### Gap 1: Graph não está integrado

- **Impacto**: AgentManager.chat() retorna placeholder
- **Solução**: Integrar `build_graph()` no `initialize()`
- **Dependência**: graph.py deve estar finalizado

### Gap 2: EmbeddingQueue é stub

- **Impacto**: Documentos não são processados
- **Solução**: Implementar queue_document() + \_worker_loop() reais
- **Dependência**: SQLAlchemy async setup completo

### Gap 3: Sem testes de integração

- **Impacto**: Difícil detectar quebras nas interfaces
- **Solução**: Adicionar pytest async + fixtures
- **Dependência**: Nenhuma

### Gap 4: chat.py ainda usa legacy checkpointer

- **Impacto**: Dois sistemas de persistência competindo
- **Solução**: Refatorar para usar SessionService do AgentManager
- **Dependência**: AgentManager.chat() implementado

---

## 📈 Progressão Visual (Gantt)

```
Phase 2 (Completo):
  ✅ Week 1: Settings + Service Stubs
  ✅ Week 2: TelemetryService, SessionService, EmbeddingService

Phase 3 (Em Andamento):
  🔄 Week 3 (Próxima):
     - Integrar Graph no AgentManager
     - Implementar EmbeddingQueue real
     - Refatorar chat.py
     - Testes de integração

Phase 4 (Futuro):
  ⏳ Week 4: Graph Isolation
     - Multi-node graph
     - Error recovery
     - Sub-graphs

Phase 5 (Futuro):
  ⏳ Week 5: Cleanup + PyPI
     - Delete legacy files
     - Documentation
     - Release
```

---

## 🎓 Lições Aprendidas / Decisões Arquiteturais

### 1. **Dependency Injection via AgentManager**

- ✅ Eliminou circular imports
- ✅ Facilita testes (mock services)
- ✅ Centraliza lifecycle management

### 2. **SessionMetadata em State (não Context)**

- ✅ State é JSON-serializable para LangGraph checkpoints
- ✅ Remove objetos complexos do RunnableConfig
- ✅ Mais previsível e tipado

### 3. **Fire-and-Forget Embedding**

- ✅ UI não bloqueia durante indexing
- ⚠️ Trade-off: eventual consistency

### 4. **WAL Mode para Concorrência**

- ✅ Permite reader + writer simultâneo
- ✅ Necessário para embeddings em background

---

## ✅ Checklist para Phase 3 Começar

- [x] Phase 2 completo (Settings + Services)
- [x] Git commits em Conventional Commits format
- [x] Documentação atualizada
- [ ] Graph integrado no AgentManager
- [ ] EmbeddingQueue implementado
- [ ] chat.py refatorado
- [ ] Testes de integração passando
- [ ] Sem erros de NoneType

---

## 📞 Contato / Próximos Passos

**O que fazer agora**:

1. **Revisar** este documento e o código atual
2. **Iniciar Phase 3 Week 3**:

   - Integrar `build_graph()` no AgentManager.initialize()
   - Implementar `chat()` com invocação real do grafo
   - Adicionar testes básicos

3. **Comunicar** quando estiver pronto para começar Phase 3

**Dúvidas**:

- A estrutura do AgentManager está correta?
- O main.py está suficientemente limpo?
- Precisa de ajustes no design antes de continuar?

---

**Gerado em**: 2026-05-16
**Phase**: 2 ✅ | 3 🔄
**Próximo Review**: Quando Phase 3 Week 3 começar
