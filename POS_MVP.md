# Roadmap Pós-MVP

Evolução estratégica do Vectora de um **agente RAG local** para um **orquestrador de Deep Agents** escalável, multi-especialista e pronto para produção.

O pré-requisito absoluto para qualquer item desta lista é que o **v0.1.0 esteja estável em produção**. Não construir sobre areia.

---

## v0.2 — Agentic Core

Esta versão transforma o MAIN_NODE monolítico num **Supervisor** que delega para agentes especialistas. A arquitetura Supervisor/Worker é o padrão correto para sistemas que precisam de escalabilidade e especialização.

### Multi-Agent Architecture

```
Supervisor (router inteligente)
  ├── SearchAgent   → RAG, Tavily, web research
  ├── CoderAgent    → File ops, terminal, git
  └── AnalysisAgent → Reasoning profundo, síntese, critique
```

Cada agente tem seu próprio grafo LangGraph, ferramentas especializadas e contexto isolado. O Supervisor decide qual agente é mais adequado para cada tarefa recebida.

Itens concretos:

- [ ] Implementar `agents/supervisor.py` com routing inteligente
- [ ] Implementar `agents/search_agent.py` (RAG + Tavily)
- [ ] Implementar `agents/coder_agent.py` (files + terminal)
- [ ] Migrar `web_search` e `fetch_url` para Tavily API nativa (melhor contexto RAG que DuckDuckGo)
- [ ] Suporte a execução paralela de sub-agentes (`asyncio.gather`)

### Infra Escalável

Para produção real com múltiplos usuários simultâneos:

- [ ] PostgreSQL como alternativa ao SQLite (connection pool, concurrent writes)
- [ ] Qdrant Cloud como alternativa ao LanceDB (vector store gerenciado, escalável)
- [ ] Valkey/Redis como cache distribuído (embeddings, respostas frequentes)
- [ ] Migrações automáticas SQLite → PostgreSQL

### Observabilidade Production-Grade

- [ ] Prometheus metrics (latência, tokens/call, taxa de erro por tool)
- [ ] Jaeger distributed tracing (trace completo de cada delegação A2A)
- [ ] LangSmith com `client_thread_id` nos metadados (identificar qual agente Paperclip causou o quê)
- [ ] Alert system (Slack/Discord/PagerDuty para erros críticos)

---

## v0.3 — Memory & Human-in-the-Loop

Esta versão adiciona **consciência temporal** (o agente lembra de versões anteriores de si mesmo) e **controle humano** (o usuário pode revisar antes de ações destrutivas).

### Human-in-the-Loop (HITL)

LangGraph suporta `interrupt_before` — o grafo pausa e aguarda confirmação humana antes de executar nós específicos. Casos de uso:

- Antes de `file_edit` ou `file_write` em produção
- Antes de `terminal` com comandos de mutação (git push, rm, deploy)
- Antes de `ingest_docs` em coleções críticas

A CLI exibirá um Rich Panel pedindo confirmação antes de prosseguir.

### Long-Term Memory aprimorada

A base já existe (`save_memory`, `get_memory`) mas precisa evoluir:

- [ ] User profile persistente — preferências, estilo de código, projetos ativos
- [ ] Consolidação automática — memórias antigas resumidas para evitar overflow
- [ ] Memórias contextuais — recuperadas automaticamente com base no assunto da conversa
- [ ] TTL inteligente — memórias expiram por uso, não apenas por tempo

---

## v0.4 — Deep Agents e Self-Correction

O ponto alto do roadmap: **agentes que se auto-criticam** e **corrigem respostas fracas** antes de entregar ao usuário.

### Reflection Pattern

Um nó adicional no grafo avalia a qualidade da resposta gerada:

```
call_llm → [resposta fraca?] → critique_node → call_llm (nova tentativa)
         → [resposta ok?]   → END
```

Critérios de qualidade: coerência, completude, fundamentação em fontes, tom adequado.

### Self-Correction em Loop

Se a resposta não passar no critique, o agente recebe a crítica como nova instrução: _"Sua resposta anterior foi incompleta porque X. Tente novamente com foco em Y."_ Isso reduz alucinações drasticamente em tarefas complexas.

### Streaming de Respostas

Hoje `delegate()` bloqueia até o LangGraph finalizar. Com `astream_events` do LangGraph:

- [ ] CLI exibe tokens em tempo real (streaming token-by-token)
- [ ] MCP SSE envia eventos incrementais para o Paperclip
- [ ] Cancelamento mid-stream suportado

---

## Débitos Técnicos Prioritários

Itens que devem ser resolvidos antes ou durante v0.2, pois criam riscos reais:

### `thread_id: str` nativo

Atualmente, IDs alfanuméricos são convertidos via `hash(s) & 0xFFFFFFFF`. Em larga escala, colisões de hash são matematicamente inevitáveis (birthday paradox). A migração para `str` nativo no LangGraph Checkpointer é simples e deve ser feita no início de v0.2.

### SSE Heartbeat

Em modo SSE, conexões ociosas são silenciosamente fechadas por firewalls e load balancers (típico: 30–60s). Tarefas de delegação longas falham sem feedback. Solução: `mcp.settings.heartbeat_interval = 25` (abaixo do threshold padrão de 30s).

### LangSmith Multi-Agent Correlation

Sem `client_thread_id` nos traces do LangSmith, é impossível auditá-los em produção multi-agent. Cada trace aparece sem identificação do agente cliente que o causou.

---

## Plugins & Ecossistema (v0.3+)

Extensões planejadas para o ecossistema Vectora:

- [ ] **VSCode Extension** — Usar Vectora inline no editor sem sair do contexto de código
- [ ] **ACP Protocol Server** — Agent Client Protocol para Zed/Neovim enviarem comandos ao agente
- [ ] **Plugin oficial Paperclip** — `integrations/paperclip/plugin/` com API ergonômica de alto nível
- [ ] **Vectora Asset Library** — Buckets de embeddings pré-treinados (Next.js docs, FastAPI, etc.)
- [ ] **Gemini CLI Plugin** — Usar Vectora como ferramenta dentro do Gemini CLI

---

## Versão Resumida do Roadmap

```
v0.1.0 ── MVP ─────────────────────────────────┐
                                                │
          └──► v0.2: Multi-Agent + PostgreSQL   │
                    + Tavily + Observabilidade  │
                                                │
          └──► v0.3: HITL + Long-Term Memory    │
                    + Streaming                 │
                                                │
          └──► v0.4: Deep Agents + Reflection   │
                    + Self-Correction           │
                                                │
          └──► v1.0: Plugins + Asset Library    │
                    + ACP + VSCode              ┘
```
