# Consolidação Core vs Internal — Escopo MVP

> **Contexto:** O `core/` foi construído em outro repositório seguindo exatamente o escopo do MVP. O `internal/` cresceu neste repositório com implementações funcionais que o `core/` ainda não completou.  
> **Estratégia:** Não é mover arquivos cegamente. É **analisar cada módulo de ambos**, identificar o que é funcional no `internal/`, o que é esqueleto no `core/`, e **refatorar o `core/`** para ficar completo, funcional e alinhado ao MVP.
>
> **Destino final:** Tudo consolidado no `core/`. O `internal/` será eliminado. O `cmd/` importará apenas de `core/`.

---

## 1. Visão Geral: Quem Ganha em Cada Módulo

| Módulo           | `internal/` Status                                          | `core/` Status                                  | Decisão                                                                            |
| ---------------- | ----------------------------------------------------------- | ----------------------------------------------- | ---------------------------------------------------------------------------------- |
| **Tray**         | ✅ Funcional (systray completo, provider switch, i18n)      | ⚠️ Mínimo (só "Running" + "Quit")               | **Refatorar `core/tray/`** com lógica do `internal/tray/`                          |
| **IPC**          | ✅ Framework RPC completo (14 endpoints, broadcast, client) | ⚠️ Skeleton (~30 linhas, 2 mensagens)           | **Refatorar `core/api/ipc/`** com lógica do `internal/ipc/`                        |
| **LLM**          | ✅ Dual provider (Gemini + Qwen), tool calling, conversas   | ⚠️ Só Gemini, interface streaming, sem Qwen     | **Refatorar `core/llm/`** com Qwen + tool calling do `internal/llm/`               |
| **Storage**      | ✅ Flat KV + Vector stores funcionais                       | ✅ Engine workspace-scoped com dedup + chunking | **Manter `core/storage/`** (superior), usar interfaces do `internal/db/`           |
| **Tools**        | ✅ 8+ ferramentas funcionais                                | ✅ 3 ferramentas com Guardian (seguro)          | **Refatorar `core/tools/`** — portar tools do `internal/` com segurança do `core/` |
| **RAG Pipeline** | ✅ Pipeline RAG funcional                                   | ⚠️ `core/engine/` com stubs                     | **Refatorar `core/engine/`** — portar RAG do `internal/core/`                      |
| **Auth**         | ❌ Arquivo vazio                                            | ❌ Não existe                                   | **Remover ambos**, implementar depois                                              |
| **i18n**         | ✅ Funcional (4 langs, CSV embedado)                        | ❌ Não existe                                   | **Criar `core/i18n/`** com código do `internal/i18n/`                              |
| **OS Manager**   | ✅ Abstração multi-plataforma funcional                     | ❌ Não existe                                   | **Criar `core/os/`** com código do `internal/os/`                                  |
| **Infra**        | ✅ `.env`, slog logger, notificações                        | ⚠️ `core/config/` sofisticado (YAML + AES-GCM)  | **Manter `core/config/`**, mover logger para `core/telemetry/`                     |
| **gRPC**         | ❌ Não existe                                               | ⚠️ Proto + stub                                 | **Manter como opcional** (desativado por default no MVP)                           |
| **Ingestion**    | ❌ Não existe                                               | ✅ Indexer + parser + dependency graph          | **Manter `core/ingestion/`**                                                       |
| **Policies**     | ❌ Não existe                                               | ✅ Guardian + regras YAML                       | **Manter `core/policies/`**                                                        |
| **Telemetry**    | ⚠️ Logger simples em `internal/infra/`                      | ✅ Logger rotativo com `slog`                   | **Manter `core/telemetry/`**                                                       |

---

## 2. Plano de Consolidação por Módulo

### 2.1 `core/tray/` — Systray

**Problema:** `core/tray/tray.go` é mínimo — só mostra "Running" e "Quit". Falta tudo que o MVP precisa.

**O que `internal/tray/` tem que o `core/` não tem:**

- Menu completo: status, abrir app, abrir CLI, trocar provedor (Gemini/Qwen), trocar idioma (EN/PT/ES/FR), configurações, quit
- Provider switching dinâmico com `switchProvider()`
- Integração com `i18n`, `infra`, `llm`, `os`
- Detecção de status do daemon

**Ação:** Reescrever `core/tray/tray.go` usando `internal/tray/tray.go` como referência, adaptando para as interfaces do `core/` (ex: `core/llm/Router` em vez de `internal/llm/Provider`).

**Manter `core/tray/tray_stub.go`** — já está correto.

---

### 2.2 `core/api/ipc/` — Comunicação Inter-Processos

**Problema:** `core/api/ipc/server.go` tem ~30 linhas com 2 mensagens (`start_index`, `get_status`). É um protótipo.

**O que `internal/ipc/` tem que o `core/` não tem:**

- `protocol.go`: `IPCMessage` tipado (request/response/event), `IPCError` com códigos de erro canônicos
- `server.go`: Named Pipes (Windows) + Unix sockets com protocolo JSON-ND, client tracking, broadcast
- `client.go`: Cliente RPC assíncrono com pending response tracking, event hooks, UUID correlation
- `router.go`: 14 endpoints RPC registrados com dependências injetadas (KVStore, VectorStore, LLM, MessageService, etc.)

**Ação:** Reescrever `core/api/ipc/` usando `internal/ipc/` como referência. Adaptar:

- `protocol.go` → copiar e adaptar para `core/api/ipc/`
- `server.go` → copiar Named Pipes/Unix socket logic
- `client.go` → copiar cliente RPC
- `router.go` → copiar mas **remover endpoints de update** (`app.update.check`, `app.update.execute`)
- **NÃO** copiar `update_handlers.go` (sistema de atualização não existe no MVP)

**Nota:** O `core/api/router.go` (dispatcher multi-protocolo) tem `startIPCServer()` como stub comentário. Após `core/api/ipc/` estar pronto, wire nele.

---

### 2.3 `core/llm/` — Gateway de Modelos

**Problema:** `core/llm/` tem interface streaming boa, mas só Gemini implementado. Sem Qwen local, sem tool calling, sem conversation service.

**O que `internal/llm/` tem que o `core/` não tem:**

- `qwen.go`: Provider local via `llama.cpp` sidecar com STDIO
- `protocol_llama.go`: `LlamaProcess` gerencia ciclo de vida do `llama-cli`
- Tool calling nativo no `CompletionRequest`/`CompletionResponse`
- `service.go`: `MessageService` para conversas (create, rename, delete, list, add)
- `messages.go`: Tipos `Role`, `ChatMessage`, `Conversation`, `LlamaRequest`
- `instruct/`: `prompt.txt` e `tools.json` para system prompts
- `gemini.go` usa `langchaingo` (já funcional com tool calling)

**O que `core/llm/` tem que é melhor:**

- Interface streaming com `io.ReadCloser` (melhor para UI responsiva)
- `context_manager.go`: Gerenciamento de janela de contexto com truncagem
- `prompt_factory.go`: System prompt dinâmico com injeção de policies + RAG context

**Ação:** Refatorar `core/llm/` mesclando o melhor de ambos:

1. **Manter** a interface streaming de `core/llm/provider.go` mas **adicionar** `ToolCall` e `TokenUsage` do `internal/`
2. **Adicionar** `qwen.go` + `protocol_llama.go` do `internal/llm/` adaptados para a interface do `core/`
3. **Adicionar** `service.go` (MessageService) do `internal/llm/`
4. **Adicionar** `messages.go` tipos do `internal/llm/`
5. **Manter** `context_manager.go` e `prompt_factory.go` do `core/llm/`
6. **Manter** `gemini_provider.go` do `core/llm/` mas **adicionar** tool calling support
7. **Manter** `router.go` do `core/llm/`
8. **Copiar** `instruct/prompt.txt` e `instruct/tools.json` do `internal/llm/`
9. **Remover** `internal/llm/gemini.go` (usa langchaingo, redundante com `core/llm/gemini_provider.go` que usa genai SDK — manter genai SDK que é mais direto)

---

### 2.4 `core/storage/` — Persistência

**Problema:** Nenhum problema real. `core/storage/` é superior.

**O que `internal/db/` tem de útil:**

- `interfaces.go`: Interfaces limpas `VectorStore` e `KVStore` como contratos
- `db_test.go`: Testes

**O que `core/storage/` tem que é melhor:**

- `engine.go`: Engine unificada com `IndexFile()` (hash dedup, rollback, chunking)
- `chunker.go`: `SimpleChunker` com overlap e newline-aware boundaries
- Workspace-scoped collections
- Mais sofisticado em todos os aspectos

**Ação:** **Manter `core/storage/` intacto**. Opcionalmente copiar `interfaces.go` do `internal/db/` como contrato se útil. Remover `internal/db/`.

---

### 2.5 `core/tools/` — Ferramentas Agênticas

**Problema:** `core/tools/` tem apenas 3 ferramentas (`read_file`, `grep_search`, `terminal_run`) mas todas com Guardian validation. `internal/tools/` tem 8+ ferramentas mas ZERO segurança.

**O que `internal/tools/` tem que o `core/` não tem:**

- `filesystem.go`: `write_file`, `read_folder`, `edit` (com backup snapshots)
- `search.go`: `find_files` (cross-platform glob)
- `web.go`: `google_search` (DuckDuckGo HTML), `web_fetch` (URL scraping)
- `memory.go`: `save_memory` (persiste em BBolt)
- `engine.go`: `Registry` com `ExecuteStringArgs`

**O que `core/tools/` tem que é melhor:**

- Guardian integration em toda tool
- TrustFolder enforcement
- `sanitizer.go`: UTF-8 validation, output truncation, control char stripping
- Interface com `json.RawMessage` args (mais flexível para JSON-RPC)

**Ação:** Refatorar `core/tools/` portando todas as ferramentas do `internal/tools/` com o padrão de segurança do `core/`:

1. **Manter** `tool.go`, `registry.go`, `sanitizer.go` do `core/tools/`
2. **Manter** `read_file.go`, `grep_search.go`, `terminal_run.go` do `core/tools/`
3. **Portar e adaptar** de `internal/tools/`:
   - `write_file` → `core/tools/write_file.go` (com Guardian + Git snapshot)
   - `read_folder` → `core/tools/read_folder.go` (com Guardian)
   - `edit` → `core/tools/edit.go` (com Guardian + backup)
   - `find_files` → `core/tools/find_files.go` (com Guardian)
   - `google_search` → `core/tools/google_search.go`
   - `web_fetch` → `core/tools/web_fetch.go`
   - `save_memory` → `core/tools/save_memory.go` (com Guardian)
4. **NÃO portar** `enter_plan_mode` / `PlanModeTool` — não faz parte do MVP
5. **NÃO portar** `search_service.go` — redundante com tools individuais
6. Remover `internal/tools/`

---

### 2.6 `core/engine/` — Orquestrador + RAG Pipeline

**Problema:** `core/engine/engine.go` tem a estrutura certa mas os métodos principais são stubs. `StreamQuery()` retorna `"RAG not fully wired yet"`.

**O que `internal/core/rag_pipeline.go` tem que o `core/` não tem:**

- Pipeline RAG funcional completo:
  1. Embed da query via LLM
  2. KNN search no Chromem (top-k=5)
  3. Flatten chunks de texto
  4. Injeção de memória do usuário
  5. System prompt com contexto agêntico + tool registry
  6. LLM completion com tool call handling
- Usa `internal/acp/AgentContext` como condutor das ferramentas

**Ação:** Refatorar `core/engine/`:

1. **Portar** a lógica de RAG do `internal/core/rag_pipeline.go` para `core/engine/rag_pipeline.go`
2. **Adaptar** para usar `core/storage/`, `core/llm/`, `core/tools/`, `core/policies/`
3. **Implementar** `StreamQuery()` de verdade com streaming de tokens
4. **Implementar** `StartIndexation()` usando `core/ingestion/`
5. **Implementar** `GetStatus()` com status real do engine
6. **Manter** `ExecuteTool()` (já funcional com Guardian validation)
7. Remover `internal/core/`

---

### 2.7 `core/i18n/` — Internacionalização (NOVO)

**Problema:** Não existe no `core/`.

**O que `internal/i18n/` tem:**

- `i18n.go`: Sistema completo com CSV embedado, 4 langs (EN/PT/ES/FR), thread-safe, fallback inglês
- `translations.csv`: Traduçoes

**Ação:** Copiar `internal/i18n/` → `core/i18n/` intacto. É self-contained, sem dependências problemáticas. Remover `internal/i18n/`.

---

### 2.8 `core/os/` — OS Manager (NOVO)

**Problema:** Não existe no `core/`.

**O que `internal/os/` tem:**

- `manager.go`: Interface `OSManager` com factory dispatch por `runtime.GOOS`
- `windows/windows.go`: Paths, registry, mutex single-instance, admin detection
- `macos/macos.go`: Paths, single-instance
- `linux/linux.go`: Paths, `.desktop` registration, single-instance

**Ação:** Copiar `internal/os/` → `core/os/`. Simplificar:

- Remover: registro no Windows Registry (shortcuts), `.desktop` files no Linux (não precisa no MVP)
- Manter: paths multi-plataforma, single-instance lock, controle do llama-engine
- Remover `internal/os/`

---

### 2.9 `core/config/` — Configuração

**Problema:** Nenhum. `core/config/` é mais sofisticado que `internal/infra/config.go`.

**O que `core/config/` tem:**

- YAML-based config com versão, default model, fallback cloud
- Secrets com AES-256-GCM encryption
- Workspace isolation via SHA-256 hash do path
- `Manager`, `WorkspaceManager`, `Secrets`, `Types`

**O que `internal/infra/config.go` tem:**

- `.env` loading via `godotenv`
- Config simples com `GeminiAPIKey`

**Ação:** **Manter `core/config/`** completo. Opcionalmente adicionar suporte a `.env` loading como fallback para compatibilidade. Remover `internal/infra/config.go` mas **manter** `internal/infra/logger.go` e notificações (ver abaixo).

---

### 2.10 `core/telemetry/` — Observabilidade

**Problema:** Nenhum. `core/telemetry/` já tem logger rotativo com `slog`.

**O que `internal/infra/logger.go` tem:**

- Logger JSON com `slog` gravando em `~/.Vectora/logs/daemon.log`

**Ação:** **Manter `core/telemetry/`**. É superior (rotating writer). Remover `internal/infra/logger.go`.

**Notificações:** `internal/infra/notify_windows.go` e `notify_unix.go` são úteis. Mover para `core/notify/` ou manter em `internal/infra/` até o systray do `core/` ser funcional.

---

### 2.11 `core/ingestion/` — Pipeline de Ingestão

**Problema:** Nenhum. Já existe e é funcional no `core/`.

**Ação:** **Manter intacto**. Não há equivalente no `internal/`.

---

### 2.12 `core/policies/` — Políticas e Guardian

**Problema:** Nenhum. Já existe e é funcional no `core/`.

**Ação:** **Manter intacto**. Não há equivalente no `internal/`.

---

### 2.13 `core/api/` — API Multi-Protocolo

**Problema:** A estrutura existe mas os servers são stubs.

**O que existe:**

- `core/api/router.go`: Dispatcher com `startJSONRPCServer()`, `startgRPCServer()`, `startIPCServer()` — todos como **stubs**
- `core/api/jsonrpc/server.go`: JSON-RPC via stdio/TCP (parcial)
- `core/api/jsonrpc/methods/tools_call_method.go`: Handler funcional
- `core/api/grpc/`: Proto definition + stub handler
- `core/api/ipc/`: Skeleton (ver seção 2.2)
- `core/api/shared/middleware.go`: Middleware comum

**Ação:**

1. **Completar** `core/api/ipc/` (ver seção 2.2)
2. **Completar** `core/api/jsonrpc/server.go` com loop stdio/TCP real
3. **Manter** `core/api/grpc/` como opcional (desativado por default)
4. **Manter** `core/api/shared/middleware.go`
5. **Wire** tudo no `core/api/router.go` — implementar os `start*Server()` methods

---

## 3. O Que Remover Diretamente (Sem Consolidar)

| Caminho                                     | Motivo                                                      |
| ------------------------------------------- | ----------------------------------------------------------- |
| `internal/auth/`                            | Arquivo vazio, sem uso                                      |
| `internal/core/`                            | RAG pipeline será portado para `core/engine/`               |
| `cmd/tests/`                                | Usar `go test ./...`                                        |
| `cmd/daemon/update.go`                      | Sistema de atualização não existe no MVP                    |
| `cmd/daemon/elevate_windows.go`             | Elevação de privilégios não existe no MVP                   |
| `cmd/daemon/elevate_unix.go`                | Elevação de privilégios não existe no MVP                   |
| `cmd/desktop/`                              | Fyne Desktop não existe no MVP                              |
| `cmd/tui/`                                  | Bubbletea TUI não existe no MVP                             |
| `cmd/setup/` complexo                       | Simplificar — sem UI Fyne, sem LPM/MPM integration complexa |
| `cmd/lpm/`                                  | Simplificar ou integrar ao CLI principal                    |
| `cmd/mpm/`                                  | Simplificar ou integrar ao CLI principal                    |
| `internal/ipc/update_handlers.go`           | Update system não existe no MVP                             |
| `internal/tools/memory.go` — `PlanModeTool` | Modo de planejamento não existe no MVP                      |
| `core/api/ipc/` (atual)                     | Skeleton — será reescrito                                   |
| `core/engine/` — stubs                      | Será implementado com RAG do `internal/core/`               |

---

## 4. `go.mod` — Dependências

**Manter:**

- `github.com/spf13/cobra` — CLI
- `github.com/getlantern/systray` — Systray
- `github.com/philippgille/chromem-go` — Vector DB
- `go.etcd.io/bbolt` — KV DB
- `github.com/tmc/langchaingo` — Gemini/LLM (tool calling)
- `github.com/joho/godotenv` — `.env` loading
- `github.com/google/uuid` — UUIDs para IPC
- `golang.org/x/sys` — Syscalls

**Remover:**

- `fyne.io/fyne/v2` — Sem Desktop GUI
- `github.com/charmbracelet/bubbles` — Sem TUI Bubbletea
- `github.com/charmbracelet/bubbletea` — Sem TUI Bubbletea
- `github.com/charmbracelet/lipgloss` — Sem TUI Bubbletea
- `google.golang.org/grpc` — Se não for usar gRPC no MVP (opcional)
- `github.com/gen2brain/beeep` — Se notificações forem apenas via systray
- `github.com/go-toast/toast` — Se notificações forem apenas via systray
- `github.com/nu7hatch/gouuid` — Se `github.com/google/uuid` for suficiente

---

## 5. Ordem de Execução

A ordem importa para manter o projeto compilando a cada passo:

### Fase 1: Fundações (sem quebrar o build)

1. **Copiar `internal/i18n/` → `core/i18n/`** — self-contained, zero risco
2. **Copiar `internal/os/` → `core/os/`** — self-contained, zero risco (simplificar após copiar)
3. **Simplificar `core/os/`** — remover registry shortcuts, .desktop files

### Fase 2: Storage e Tools (base para tudo)

4. **Portar tools do `internal/tools/` → `core/tools/`** — write_file, read_folder, edit, find_files, google_search, web_fetch, save_memory (todos com Guardian)
5. **Remover `PlanModeTool`** — não existe no MVP
6. **Remover `internal/tools/`** — após portar tudo

### Fase 3: LLM (cérebro)

7. **Adicionar tool calling à interface de `core/llm/provider.go`**
8. **Portar `qwen.go` + `protocol_llama.go` do `internal/llm/` → `core/llm/`**
9. **Portar `service.go` (MessageService) do `internal/llm/` → `core/llm/`**
10. **Portar `messages.go` tipos do `internal/llm/` → `core/llm/`**
11. **Copiar `instruct/` do `internal/llm/` → `core/llm/instruct/`**
12. **Adicionar tool calling ao `gemini_provider.go` do `core/llm/`**
13. **Remover `internal/llm/`**

### Fase 4: RAG Pipeline (coração)

14. **Portar RAG do `internal/core/rag_pipeline.go` → `core/engine/rag_pipeline.go`**
15. **Implementar `StreamQuery()` real no `core/engine/`**
16. **Implementar `StartIndexation()` usando `core/ingestion/`**
17. **Implementar `GetStatus()` real**
18. **Remover `internal/core/`**

### Fase 5: IPC e API (comunicação)

19. **Reescrever `core/api/ipc/`** com lógica do `internal/ipc/` (sem update_handlers)
20. **Completar `core/api/jsonrpc/server.go`**
21. **Implementar `core/api/router.go`** — wire todos os servers
22. **Remover `internal/ipc/`**

### Fase 6: Systray (interface)

23. **Reescrever `core/tray/`** com lógica do `internal/tray/` (adaptar para interfaces do `core/`)
24. **Remover `internal/tray/`**

### Fase 7: Cleanup (remover `internal/`)

25. **Mover `internal/infra/logger.go` → substituir por `core/telemetry/`** (já existe)
26. **Mover notificações** → `core/notify/` ou manter referência no systray
27. **Remover `internal/infra/`**
28. **Remover `internal/auth/`**
29. **Remover `internal/db/`** — `core/storage/` é superior
30. **Remover `internal/acp/`** — integrar ao `core/tools/registry.go`

### Fase 8: Cmd e go.mod

31. **Remover `cmd/desktop/`**, `cmd/tui/`, `cmd/tests/`
32. **Simplificar `cmd/setup/`**, `cmd/lpm/`, `cmd/mpm/`
33. **Remover `cmd/daemon/update.go`**, `elevate_*.go`
34. **Wire `cmd/daemon/`** para importar de `core/` em vez de `internal/`
35. **`go mod tidy`** — limpar dependências não utilizadas
36. **`go build ./...`** — verificar compilação

---

## 6. Estado Final Esperado

```
vectora/
├── cmd/
│   ├── daemon/       → Importa apenas de core/
│   ├── setup/        → Simplificado
│   └── (lpm, mpm)    → Simplificados ou integrados
├── core/
│   ├── api/          → JSON-RPC, gRPC (opcional), IPC
│   ├── config/       → YAML + AES-GCM + workspace isolation
│   ├── engine/       → Orquestrador + RAG Pipeline (FUNCIONAL)
│   ├── i18n/         → 4 idiomas, CSV embedado
│   ├── ingestion/    → Indexer + parser + dependency graph
│   ├── llm/          → Gemini + Qwen, tool calling, conversations
│   ├── notify/       → Notificações OS
│   ├── os/           → Multi-plataforma simplificado
│   ├── policies/     → Guardian + regras YAML
│   ├── storage/      → Dual-store engine com chunking
│   ├── tools/        → 8+ ferramentas com Guardian
│   └── tray/         → Systray completo
├── go.mod
└── go.sum
```

**Zero `internal/`. Zero duplicatas. Zero stubs.**

---

> **Nota:** Cada fase deve terminar com `go build ./...` passando. Se algo quebrar, resolver antes de prosseguir. Esta é uma refatoração incremental, não um rewrite Big Bang.
