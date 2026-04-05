# VECTORA: TIMELINE DE DESENVOLVIMENTO

**Status:** Roadmap Executivo — Fases e Milestones
**Versão:** 1.0
**Data:** 2026-04-05
**Idioma:** Português (PT-BR)
**Escopo:** Cronograma consolidado para implementação do Vectora v2.0

---

## VISÃO GERAL DO CRONOGRAMA

```
    Total Estimado: 24-28 semanas (6-7 meses)
    Data Alvo: Setembro/Outubro 2026
    Status Atual: Planning Complete — Pronto para Implementação

    Semanas  1-4   : Fase 1 - Fundação
    Semanas  5-8   : Fase 2 - RAG & IA
    Semanas  9-11  : Fase 3 - Ferramentas & Agência
    Semanas 12-15  : Fase 4 - Interfaces
    Semanas 16-18  : Fase 5 - Integração
    Semanas 19-28  : Fase 6 - Polimento & Release
```

---

## FASE 1: FUNDAÇÃO (Semanas 1-4)

### Objetivo
Estabelecer infraestrutura base: logging, IPC, banco de dados.

### Tarefas

#### Semana 1: Setup e Infraestrutura
- [ ] **Dia 1-2:** Setup repositório, estrutura Go, Makefile
  - CI/CD pipeline inicial (GitHub Actions)
  - Pre-commit hooks para code quality
  - Documentação CONTRIBUTING.md aprovada

- [ ] **Dia 3-5:** ConfigLoader e Logger
  - Implementar `internal/infra/config.go`
  - Logger estruturado com slog + JSON
  - Rotação de logs (10MB, 5 arquivos)
  - Testes unitários

**Deliverables:**
- ✅ Repositório pronto para desenvolvimento
- ✅ Logger centralizado funcionando
- ✅ Tests passando (config + logger)

#### Semana 2: IPC Server
- [ ] **Dia 6-10:** Implementar servidor IPC
  - Socket/Named Pipe (Windows + Unix)
  - Protocolo JSON-ND
  - Message routing
  - Error handling
  - Rate limiting (1000 req/min)

**Deliverables:**
- ✅ IPC server respondendo a requisições
- ✅ Múltiplos clientes simultâneos
- ✅ Testes de stress (100 conexões)

#### Semana 3: Banco de Dados
- [ ] **Dia 11-15:** Implementar camada de persistência
  - bbolt para metadata (workspaces, sessions)
  - Chromem-go para vetores
  - Schemas e migrations
  - Isolamento por workspace

**Deliverables:**
- ✅ Store interface implementada
- ✅ Vector store com isolamento
- ✅ Testes ACID

#### Semana 4: Testes de Infraestrutura
- [ ] **Dia 16-20:** Suite completa de testes
  - Unit tests (config, logger, IPC, db)
  - Integration tests
  - Memory leak detection
  - Performance profiling

**Deliverables:**
- ✅ 300+ testes passando
- ✅ >80% code coverage (infra)
- ✅ Zero memory leaks

**Checkpoint 1:** Infraestrutura base aprovada ✅

---

## FASE 2: RAG & IA (Semanas 5-8)

### Objetivo
Implementar motor RAG completo com suporte a Gemini e Qwen local.

### Tarefas

#### Semana 5: LLM Providers
- [ ] **Dia 21-25:** Implementar providers
  - Provider interface base
  - Gemini provider (langchaingo)
  - Qwen provider (sidecar llama-cli)
  - Health checks e fallback

**Deliverables:**
- ✅ Gemini API funcionando
- ✅ Qwen local respondendo
- ✅ Fallback entre providers

#### Semana 6: Engines Manager
- [ ] **Dia 26-30:** Gerenciador de llama.cpp
  - Hardware detector (CPU/GPU)
  - Build catalog (embarcado)
  - Download com resume
  - SHA256 verification
  - Process lifecycle

**Deliverables:**
- ✅ Detecção de hardware automática
- ✅ Builds baixando e verificando
- ✅ Sidecar llama-cli spawning

#### Semana 7: RAG Pipeline
- [ ] **Dia 31-35:** Orquestração RAG
  - Embedding (query vectorization)
  - Semantic search (Top-K=5)
  - Context construction
  - Prompt engineering
  - Tool calling integration

**Deliverables:**
- ✅ Query completa < 2s latência
- ✅ Sources retornando corretamente
- ✅ Testes de relevância

#### Semana 8: Workspace Manager
- [ ] **Dia 36-40:** Gerenciamento de workspaces
  - Create/Read/Update/Delete
  - Indexing pipeline
  - Progress monitoring
  - Chunk management

**Deliverables:**
- ✅ CRUD completo
- ✅ Indexação de 1000 arquivos em < 5 min
- ✅ Event streaming (index.progress)

**Checkpoint 2:** RAG end-to-end funcionando ✅

---

## FASE 3: FERRAMENTAS & AGÊNCIA (Semanas 9-11)

### Objetivo
Implementar toolkit de ferramentas e controle autônomo (ACP).

### Tarefas

#### Semana 9: ACP Registry & Tools
- [ ] **Dia 41-45:** Implementar ACP
  - Registry interface
  - Tool registration
  - Execution sandbox
  - JSON schema validation

- [ ] Filesystem tools:
  - read_file, write_file, edit_file
  - find_files, read_folder

- [ ] Information tools:
  - grep_search, web_search, web_fetch

**Deliverables:**
- ✅ Registry com 6+ tools
- ✅ JSON schema validation
- ✅ Isolamento de execução

#### Semana 10: GitBridge & Undo
- [ ] **Dia 46-50:** Snapshot e rollback
  - Snapshot antes de mutação
  - Hash-based deduplication
  - Restore functionality
  - Auditoria completa

- [ ] Shell tool:
  - shell_command com timeout
  - Streaming de output
  - Error handling

**Deliverables:**
- ✅ Undo/Redo funcional
- ✅ Auditoria de todas operações
- ✅ 90 dias de retenção

#### Semana 11: Testes Agênticos
- [ ] **Dia 51-55:** Suite de testes
  - Tool execution tests
  - Concurrency tests
  - Security tests (RN-TOOL-01 a 04)
  - Performance baseline

**Deliverables:**
- ✅ 100+ testes de tools
- ✅ Race condition free
- ✅ Segurança validada

**Checkpoint 3:** ACP completo e auditado ✅

---

## FASE 4: INTERFACES (Semanas 12-15)

### Objetivo
Implementar Web UI, CLI, e integrações.

### Tarefas

#### Semana 12: Web UI Base
- [ ] **Dia 56-60:** Setup Next.js + Wails
  - Estrutura de projeto
  - Layout principal
  - Sidebar + header
  - Zustand stores
  - IPC hooks

**Deliverables:**
- ✅ App scaffolding pronto
- ✅ Build < 3MB
- ✅ Boot < 1.5s

#### Semana 13: Aba Chat
- [ ] **Dia 61-65:** Implementar chat
  - ChatFeed component
  - InputArea (auto-grow)
  - MessageBubble (Markdown)
  - Streaming respostas
  - Source badges

**Deliverables:**
- ✅ Chat funcional
- ✅ Streaming LLM
- ✅ Tool calls visualizados

#### Semana 14: Abas Código + Index
- [ ] **Dia 66-70:** Explorador e workspaces
  - FileTree com Monaco editor
  - Terminal integrado
  - WorkspaceList + upload
  - DatasetBrowser (Index)

**Deliverables:**
- ✅ File browser completo
- ✅ Code editor syntax highlighting
- ✅ Workspace CRUD

#### Semana 15: Aba Manager + CLI
- [ ] **Dia 71-75:** Package manager UI + CLI TUI
  - LPMPanel e MPMPanel
  - ConfigurationPanel
  - Bubbletea CLI setup
  - Integration com setup

**Deliverables:**
- ✅ UI para controlar pacotes
- ✅ CLI TUI responsivo
- ✅ Beide interfaces síncronos

**Checkpoint 4:** Todas interfaces básicas ✅

---

## FASE 5: INTEGRAÇÃO (Semanas 16-18)

### Objetivo
Conectar Setup Installer, Package Managers, e refinar fluxos.

### Tarefas

#### Semana 16: Setup Installer
- [ ] **Dia 76-80:** Implementar Fyne wizard
  - 8 screens state machine
  - Hardware detection
  - Build selection + download
  - Model selection + download
  - Configuration wizard

**Deliverables:**
- ✅ Setup wizard end-to-end
- ✅ Hardware-aware defaults
- ✅ Progress monitoring

#### Semana 17: Package Managers
- [ ] **Dia 81-85:** LPM + MPM CLI finais
  - JSON output schemas
  - Resume download support
  - Error messages claros
  - Retry logic (exponential backoff)

**Deliverables:**
- ✅ LPM binary < 20MB
- ✅ MPM binary < 15MB
- ✅ JSON parsing robusto

#### Semana 18: Build Integration
- [ ] **Dia 86-90:** Integração de builds
  - Build.ps1 completo (Windows)
  - Makefile completo (Unix)
  - Checksum generation
  - Distribution packaging

**Deliverables:**
- ✅ Windows installer (.exe)
- ✅ macOS package (.dmg)
- ✅ Linux installers

**Checkpoint 5:** Setup e builds funcionais ✅

---

## FASE 6: POLIMENTO & RELEASE (Semanas 19-28)

### Objetivo
Testes E2E, otimizações, documentação, e release.

### Tarefas

#### Semanas 19-20: E2E Testing
- [ ] **Dia 91-100:** Suite completa E2E
  - Playwright tests
  - Setup flow testing
  - Chat integration
  - Code editor workflow
  - Package manager ops

**Deliverables:**
- ✅ 50+ E2E tests
- ✅ Zero flakiness
- ✅ Coverage de user journeys

#### Semanas 21-22: Performance & Optimization
- [ ] **Dia 101-110:** Profiling e otimizações
  - RAM profiling (target: < 500MB com 5 workspaces)
  - Latency optimization (IPC, UI render)
  - Bundle size reduction (Next.js)
  - Cache optimization (Chromem-go)

**Deliverables:**
- ✅ Daemon < 100MB idle
- ✅ IPC < 1ms latência
- ✅ Chat response < 100ms UI feedback

#### Semanas 23-24: Security Hardening
- [ ] **Dia 111-120:** Security audit
  - Input validation review
  - IPC security (encryption optional)
  - Tool execution restrictions
  - API key storage review

**Deliverables:**
- ✅ Security audit passed
- ✅ Documentação de threats
- ✅ Mitigation strategies

#### Semanas 25-26: Documentation
- [ ] **Dia 121-130:** Documentação final
  - User manual (PT/EN)
  - API documentation
  - Development guide
  - Troubleshooting guide
  - Video tutorials (optional)

**Deliverables:**
- ✅ Documentação em PT/EN
- ✅ API reference completa
- ✅ Sample workflows

#### Semanas 27-28: Release & Support
- [ ] **Dia 131-140:** Release
  - Beta testing (50+ users)
  - Bug fix iteration
  - Release notes
  - GitHub release
  - Community support setup

**Deliverables:**
- ✅ v1.0 released
- ✅ Installers para Windows/macOS/Linux
- ✅ Community channels (GitHub Discussions)

---

## MILESTONES & CHECKPOINTS

### Checkpoint 1 (Semana 4)
**Critério:** Infraestrutura base funcionando
- [ ] IPC server testado
- [ ] Logger produção-ready
- [ ] Banco de dados operacional
- [ ] Tests > 80% coverage

### Checkpoint 2 (Semana 8)
**Critério:** RAG end-to-end
- [ ] Query < 2s latência
- [ ] Gemini + Qwen funcionando
- [ ] Workspaces gerenciáveis
- [ ] Index em progresso

### Checkpoint 3 (Semana 11)
**Critério:** ACP completo
- [ ] Tools executando
- [ ] GitBridge auditando
- [ ] RN-TOOL validadas
- [ ] Security pass

### Checkpoint 4 (Semana 15)
**Critério:** Todas interfaces básicas
- [ ] Chat, Código, Index, Manager
- [ ] CLI TUI
- [ ] Build < 50MB
- [ ] Boot < 2s

### Checkpoint 5 (Semana 18)
**Critério:** Setup e release-ready
- [ ] Installer executável
- [ ] Auto-download models
- [ ] Configuration flow
- [ ] First-run experience

### Release (Semana 28)
**Critério:** v1.0 Public
- [ ] Todos os testes passando
- [ ] Documentação completa
- [ ] Instaladores para 3 OSes
- [ ] Community support ativo

---

## RECURSOS E TEAM

### Estimativa de Dedicação

| Role | Horas/Semana | Duração | Total |
|------|-------------|---------|-------|
| Backend Engineer (Go) | 40h | 28 sem | 1120h |
| Frontend Engineer (React) | 40h | 15 sem | 600h |
| DevOps/Build | 10h | 28 sem | 280h |
| QA/Testing | 20h | 28 sem | 560h |
| Documentation | 10h | 28 sem | 280h |

**Total Estimado:** ~2840 horas (~1.3 FTE ano inteiro)

### Dependências Externas

- Gemini API (Google) — Já integrado
- Hugging Face Hub — Downloads modelo
- GitHub — Releases, CI/CD
- Go 1.22+ — Toolchain
- Node.js 20+ — Frontend build
- Bun — Package manager

---

## RISCOS E MITIGAÇÕES

### Risk 1: Latência RAG > 2s
**Probabilidade:** Média
**Impacto:** Alto (UX degradada)
**Mitigação:**
- Caching de embeddings
- Top-K reduzido (5)
- Parallel processing
- Profile early (Semana 7)

### Risk 2: Memory leaks em daemon
**Probabilidade:** Média
**Impacto:** Alto (crash em uso prolongado)
**Mitigação:**
- pprof profiling (Semana 4)
- Stress tests (100+ workspaces)
- GC tuning
- Regular audits

### Risk 3: Inconsistência entre Setup/App
**Probabilidade:** Alta
**Impacto:** Médio (confusão do usuário)
**Mitigação:**
- Shared IPC protocol (rígido)
- E2E tests (Semana 19)
- Feature flags para rollout
- Versioning API

### Risk 4: Qwen local performance inadequada
**Probabilidade:** Média
**Impacto:** Alto (necessário Gemini sempre)
**Mitigação:**
- Benchmark early (Semana 6)
- Fallback para Gemini
- Modelo selection recomendado
- A/B testing com users

---

## PRÓXIMOS PASSOS IMEDIATOS

### Semana 1 (Próxima)
1. Finalizar aprovação deste plano com team
2. Setup repositório (branching strategy)
3. Criar issues no GitHub com user stories
4. Iniciar Semana 1 tarefas (config + logger)

### Comunicação
- Daily standup (15 min) — Slack
- Weekly sync (1h) — Vídeo
- Checkpoint reviews (2h) — Apresentação
- Retro bi-semanais

---

**Status:** Pronto para kick-off
**Aprovação Necessária:** ✅
**Data de Início Recomendada:** Próxima segunda-feira

---

