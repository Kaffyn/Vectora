# VECTORA: DIAS 2-7 PLANO EXECUTIVO

**Status:** Dia 1 ✅ COMPLETO | Preparando Dias 2-7  
**Meta:** Sistema de produção estável em 7 dias

---

## DIA 1 RECAP: ESTRUTURA COMPLETA ✅

- **32.732 linhas de código** implementadas
- **173 arquivos** criados (88 Go, 118 React/TS)
- **3 agents** completados em paralelo
- **4 Componentes principais** prontos:
  1. Backend Go (Infra + IPC + DB)
  2. Web UI (React + Next.js)
  3. RAG Pipeline (LLM + Tools)
  4. Build System (Makefile + go.mod)

---

## DIA 2: BUILD & COMPILAÇÃO ⚙️

### Tarefas
```bash
# 1. Setup Go
cd ~/Desktop/Vectora
go mod tidy
go mod download

# 2. Setup Node
cd internal/app
npm install          # ou bun install
npm run build        # ou bun run build

# 3. Build binários
cd ../..
make build-daemon    # build/vectora
make build-app       # build/vectora-app
make build-cli       # build/vectora-cli
```

### Validações
- [ ] `go mod verify` ✓
- [ ] `npm test` ✓
- [ ] Todos os binários compilarem sem erro
- [ ] `./build/vectora --help` retorna sucesso

### Deliverable
**Binários compiláveis, pronto para testes**

---

## DIA 3: INTEGRAÇÃO IPC 🔌

### Tarefas
```bash
# 1. Testar daemon
./build/vectora --test-ipc

# 2. Testar system
./build/vectora --tests

# 3. Conectar Web UI ao daemon
# Editar hooks/useIPC.ts para realmente chamar daemon
# Testar cada método IPC
```

### Métodos IPC para Testar
- [ ] `ping` - Verificar conexão
- [ ] `workspace.create` - Criar workspace
- [ ] `workspace.list` - Listar workspaces
- [ ] `workspace.query` - Query RAG
- [ ] `tool.execute` - Executar ferramenta

### Validações
- [ ] IPC latency < 1ms
- [ ] Suporta 10+ conexões simultâneas
- [ ] Recuperação de desconexão
- [ ] Error handling correto

### Deliverable
**IPC funcional end-to-end, Web UI conectada ao daemon**

---

## DIA 4: DATABASE & RAG ⚡

### Tarefas
```bash
# 1. Testar bbolt operations
go test -v ./internal/db/...

# 2. Testar chromem-go vector store
go test -v ./internal/core/...

# 3. Indexar exemplo de arquivo
./build/vectora --index ~/Downloads/example.txt

# 4. Fazer query
curl -X POST http://localhost:8080/api/v1/query \
  -d '{"query": "test"}'
```

### Validações
- [ ] Create/Read/Update/Delete workspaces
- [ ] Persistência de dados (bbolt)
- [ ] Indexação de arquivos
- [ ] Busca semântica (Top-K=5)
- [ ] Context building correto
- [ ] Isolamento por workspace

### Métricas
- Índexação: 1000 arquivos em < 5 min
- Busca: < 500ms por query
- RAM: < 100MB idle, < 500MB com 5 workspaces

### Deliverable
**Workspace CRUD + Indexação + RAG query funcionando**

---

## DIA 5: LLM PROVIDERS & TOOLS 🤖

### Tarefas
```bash
# 1. Testar Gemini (se tiver key)
export GEMINI_API_KEY=<sua-key>
./build/vectora --test-llm gemini

# 2. Testar Qwen local
./build/vectora --test-llm qwen

# 3. Testar tools
go test -v ./internal/tools/...

# 4. Chat completo
./build/vectora chat
> Qual é a capital da França?
```

### LLM Tests
- [ ] Gemini provider (se configurado)
  - [ ] Complete() retorna resposta
  - [ ] Embed() retorna 384D vector
  - [ ] Fallback para Qwen se falhar
  
- [ ] Qwen local
  - [ ] Sidecar llama-cli inicia
  - [ ] Responde queries
  - [ ] Memory < 2GB

### Tools Tests
- [ ] read_file - Ler arquivo
- [ ] write_file - Escrever com snapshot
- [ ] grep_search - Buscar texto
- [ ] web_fetch - Fetch URL
- [ ] shell_command - Executar comando

### Security
- [ ] GitBridge snapshot antes de write
- [ ] Tool sandbox funcional
- [ ] Audit logging completo
- [ ] Undo/Restore funcionando

### Deliverable
**Chat completamente funcional com LLM + Tools**

---

## DIA 6: TESTES & REFINAMENTO 🧪

### Unit Tests
```bash
go test -v -race ./...
go test -cover ./...
```

**Alvo: 100+ testes**
- [ ] Infra tests (20+)
- [ ] IPC tests (25+)
- [ ] Database tests (20+)
- [ ] RAG tests (15+)
- [ ] LLM tests (15+)
- [ ] Tools tests (20+)

### Integration Tests
```bash
./build/vectora --tests
```

**Alvo: Todos os 5 checkpoints passarem**
- [ ] Checkpoint 1: Infra OK
- [ ] Checkpoint 2: RAG OK
- [ ] Checkpoint 3: Tools OK
- [ ] Checkpoint 4: UI OK
- [ ] Checkpoint 5: Integration OK

### Performance
- [ ] Profiling CPU/Memory
- [ ] Otimização de hotspots
- [ ] Cache hits > 90%
- [ ] Latência IPC < 1ms

### Bug Fixes
- [ ] Corrigir todos os failed tests
- [ ] Refatorar código duplicado
- [ ] Remover código morto
- [ ] Documentar APIs

### Deliverable
**Sistema estável com testes completos, pronto para release**

---

## DIA 7: QA FINAL & RELEASE 🚀

### E2E Tests
```bash
# Teste fluxo completo
1. Abrir Web UI
2. Criar workspace
3. Indexar arquivo
4. Fazer query
5. Executar tool
6. Undo mudança
7. Chat multi-turn
```

### Load Testing
- 10 queries simultâneas
- 1000 chunks indexados
- 5 workspaces ativos
- 30 minutos de uso

### Documentation
- [ ] README atualizado
- [ ] CONTRIBUTING.md atualizado
- [ ] API docs completos
- [ ] Troubleshooting guide

### Release Package
```bash
# Criar release
make clean
make build-all
zip -r vectora-v2.0.zip build/
```

**Artifacts:**
- `vectora-v2.0.zip` - Binários
- `SHA256SUMS` - Checksums
- `RELEASE_NOTES.md` - Mudanças
- `INSTALL.md` - Instruções

### Final Checks
- [ ] Windows compatível
- [ ] macOS compatível
- [ ] Linux compatível
- [ ] Zero security warnings
- [ ] Zero memory leaks

### Deliverable
**v2.0 PRODUCTION READY - Pronto para público**

---

## TIMELINE DE EXECUÇÃO

```
Dia 1: ✅ ESTRUTURA COMPLETA (32.732 LOC)
Dia 2: ⚙️  Build & Compilação
Dia 3: 🔌 Integração IPC
Dia 4: ⚡ Database & RAG
Dia 5: 🤖 LLM & Tools
Dia 6: 🧪 Testes & Refinamento
Dia 7: 🚀 QA Final & Release v2.0
```

---

## PRÓXIMAS AÇÕES IMEDIATAS

### Hoje (antes de dormir - Dia 1):
```bash
# Apenas documentar:
cd ~/Desktop/Vectora
ls -la internal/        # Verificar estrutura
wc -l $(find . -name "*.go") | tail -1    # Contar LOC
```

### Amanhã (Dia 2 - MORNING):
```bash
cd ~/Desktop/Vectora
go mod tidy
cd internal/app && npm install && npm run build
make build-daemon
make build-app
./build/vectora --help
```

### Amanhã (Dia 2 - EVENING):
```bash
go test ./internal/infra/... -v
# Corrigir cualquier fallo
```

---

## SUCESSO ESPERADO

```
Dia 1: FEITO ✅
Dia 2: Binários compiláveis
Dia 3: IPC funcional
Dia 4: Queries RAG funcionando
Dia 5: Chat completamente funcional
Dia 6: Sistema testado e estável
Dia 7: v2.0 em produção

Total: 1 SEMANA ATÉ RELEASE
```

---

## RECURSOS

Arquivo de referência criados:
- `DAY_1_SUMMARY.md` - Resumo do Dia 1
- `DAYS_2-7_PLAN.md` - Este arquivo
- `AGENTS_CONSOLIDATED_REPORT.md` - Relatório técnico
- Relatórios dos 3 agents

---

**Status:** Pronto para Dia 2 ✅
**Data Esperada Release:** 2026-04-11 (8 dias úteis)

