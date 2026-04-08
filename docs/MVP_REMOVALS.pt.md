# Remoções para Escopo MVP

> **Contexto:** Este documento cataloga tudo que deve ser **removido ou desativado** do código atual para que o Vectora opere estritamente dentro do escopo do MVP definido em `docs/README.pt.md`.
>
> **Princípio:** Remover **apenas** o que é dependente de: Desktop (Fyne), TUI (Bubbletea), Vectora Index (marketplace/asset library), Vectora Web, Vectora Auth, Colaboração/RBAC, Setup Wizard (LPM/MPM complexos).  
> **Mantido:** Systray, Cobra CLI, IPC, i18n/locale, DB (chromem-go + bbolt), ferramentas agênticas, Policies, Storage, LLM Gateway, Ingestion, Telemetry, Config Manager, Trust Folder.

---

## 1. O Que É MVP (NÃO Tocar)

Estes componentes **fazem parte do MVP** e devem ser preservados integralmente:

| Componente                          | Módulo                                       | Motivo                                    |
| ----------------------------------- | -------------------------------------------- | ----------------------------------------- |
| **CLI Cobra + Systray**             | `cmd/daemon/`, `internal/tray/`              | Interface primária do usuário             |
| **IPC (Named Pipes / Unix Socket)** | `internal/ipc/` → mover para `core/api/ipc/` | Comunicação Daemon ↔ UI                  |
| **i18n / Locale**                   | `internal/i18n/`                             | Internacionalização (EN/PT/ES/FR)         |
| **Vector DB (chromem-go)**          | `internal/db/vector.go`, `core/storage/`     | Busca semântica                           |
| **KV DB (bbolt)**                   | `internal/db/store.go`, `core/storage/`      | Metadados, histórico, config              |
| **Ferramentas Agênticas**           | `internal/tools/`, `core/tools/`             | read_file, grep_search, terminal_run, etc |
| **Policies / Guardian**             | `core/policies/`                             | Hard-Coded Guardian + regras YAML         |
| **Storage Engine**                  | `core/storage/`                              | Chunking, indexação, dual-store           |
| **LLM Gateway**                     | `core/llm/`, `internal/llm/`                 | Gemini + Qwen providers                   |
| **Ingestion Pipeline**              | `core/ingestion/`                            | Parser, dependency graph, indexer         |
| **Config Manager**                  | `core/config/`                               | Secrets AES-GCM, workspace isolation      |
| **Telemetry**                       | `core/telemetry/`                            | Logger rotativo com slog                  |
| **API Multi-Protocolo**             | `core/api/`                                  | JSON-RPC (MCP), gRPC, IPC                 |
| **RAG Pipeline**                    | `internal/core/`                             | Fluxo RAG canonico                        |
| **OS Manager**                      | `internal/os/`                               | Abstração multi-plataforma                |

---

## 2. Remoções Completas (Não Fazem Parte do MVP)

### 2.1 `cmd/desktop/` — App Desktop Fyne

- **Arquivos:** `cmd/desktop/main.go`
- **Motivo:** O MVP não inclui Vectora Desktop (Fyne GUI). A interface é via Systray + CLI Cobra.
- **Ação:** Remover binário inteiro. Remover dependência `fyne.io/fyne/v2` do `go.mod`.

### 2.2 `cmd/tui/` — Interface Terminal Bubbletea

- **Arquivos:** `cmd/tui/main.go`
- **Motivo:** O MVP não inclui Vectora TUI (Bubbletea/Bubbles/Lipgloss). O terminal é via CLI Cobra puro (`vectora ask`, `vectora embed`, etc).
- **Ação:** Remover binário inteiro. Remover dependências do `go.mod`:
  - `github.com/charmbracelet/bubbles`
  - `github.com/charmbracelet/bubbletea`
  - `github.com/charmbracelet/lipgloss`

### 2.3 Vectora Index — Marketplace de Datasets

- **Status:** Não há código dedicado ainda, mas qualquer menção a:
  - Catálogo de datasets vetoriais
  - Download de workspaces curados
  - Publicação/compartilhamento de datasets
  - RBAC (Privado/Equipe/Público)
  - Servidor Index (`core/api` dedicado ao Index)
- **Motivo:** O MVP opera exclusivamente com o **Trust Folder local**. Sem marketplace, sem compartilhamento, sem catálogo.
- **Ação:** Remover qualquer endpoint IPC/API relacionado a Index. Remover menções no systray, config, e UI.

### 2.4 Vectora Auth / RBAC / Colaboração

- **Arquivos:** `internal/auth/manager.go` (vazio, placeholder)
- **Motivo:** Sem autenticação, sem controle de acesso, sem colaboração no MVP.
- **Ação:** Remover pacote `internal/auth/` inteiro.

### 2.5 Vectora Web — Acesso Remoto

- **Status:** Mencionado em "Planos para o Futuro" no `docs/README.pt.md`. Nenhum código implementado ainda.
- **Motivo:** Fora do escopo MVP.
- **Ação:** Garantir que nenhum endpoint HTTP remoto seja exposto. O JSON-RPC roda apenas via **stdio** (MCP) e **IPC local**.

---

## 3. Remoções Parciais / Simplificações

### 3.1 `cmd/setup/` — Setup Wizard

- **Arquivos:** `main.go`, `cli.go`, `theme.go`, `ipc_integration.go`, `embed_windows.go`, `embed_other.go`
- **Motivo:** O setup wizard com interface gráfica Fyne e integração LPM/MPM complexa não faz parte do MVP.
- **Ação:** Simplificar para configuração via CLI (`vectora config --gemini KEY`). Remover:
  - `theme.go` (UI Fyne)
  - `embed_windows.go`, `embed_other.go` (embedding de recursos visuais)
  - `ipc_integration.go` (integração complexa com daemon)
  - Manter apenas `cli.go` como configuração simples via CLI, ou fundir em `cmd/daemon/`.

### 3.2 `cmd/lpm/` — Local Package Manager

- **Arquivo:** `main.go`
- **Motivo:** O LPM gerencia download/configuração de `llama.cpp`. No MVP, o binário do llama-cli é gerenciado de forma mais simples.
- **Ação:** Simplificar para um comando `vectora engine download` integrado ao CLI principal, ou remover como binário separado.

### 3.3 `cmd/mpm/` — Model Package Manager

- **Arquivos:** `main.go`, `commands.go`, `mpm.go`
- **Motivo:** O MPM com catálogo de modelos e sistema de download complexo não é essencial no MVP.
- **Ação:** Simplificar para `vectora model download <nome>` integrado ao CLI principal, ou remover como binário separado.

### 3.4 `cmd/daemon/` — Remover Update/Elevation

- **Arquivos:** `update.go`, `elevate_windows.go`, `elevate_unix.go`
- **Motivo:** O sistema de atualização automática de componentes e elevação de privilégios não faz parte do MVP.
- **Ação:** Remover estes arquivos. Manter `main.go` como ponto de entrada do daemon.

### 3.5 `internal/ipc/` → Mover para `core/api/ipc/`

- **Arquivos:** `protocol.go`, `server.go`, `client.go`, `router.go`, `update_handlers.go`
- **Motivo:** O IPC faz parte do MVP, mas está em `internal/`. Deve ser movido para `core/api/ipc/` para consistência com a arquitetura multi-protocolo.
- **Ação:**
  - Mover `protocol.go`, `server.go`, `client.go` para `core/api/ipc/`
  - Remover `update_handlers.go` (sistema de atualização — não existe no MVP)
  - Simplificar `router.go` — manter endpoints relevantes ao MVP:
    - **Manter:** `workspace.query`, `chat.history`, `chat.list`, `chat.create`, `chat.delete`, `chat.rename`, `message.add`, `provider.get`, `provider.set`, `memory.search`, `search.web`, `i18n.get`, `app.health`
    - **Remover:** `app.update.check`, `app.update.execute` (atualização de componentes)

### 3.6 `internal/llm/service.go` — MessageService

- **Motivo:** O serviço de mensagens gerencia conversas com rename/delete. No MVP, manter funcionalidades básicas.
- **Ação:** Manter `create`, `list`, `add`, `history`. **Manter** `rename` e `delete` — são úteis no CLI.

### 3.7 `internal/tools/memory.go` — PlanModeTool

- **Motivo:** O `PlanModeTool` (modo de planejamento estruturado) não faz parte do MVP.
- **Ação:** Remover `PlanModeTool`. Manter `SaveMemoryTool`.

### 3.8 `internal/tools/web.go` — Ferramentas Web

- **Motivo:** `GoogleSearchTool` e `WebFetchTool` são úteis no contexto de codebase para documentação externa.
- **Ação:** **Manter** ambas.

### 3.9 `core/api/` — API Multi-Protocolo

- **Motivo:** A API completa com gRPC e JSON-RPC faz parte do MVP.
- **Ação:**
  - **Manter** `jsonrpc/` (MCP via stdio + TCP dev mode)
  - **Manter** `ipc/` (após mover de `internal/ipc/`)
  - **Manter** `grpc/` — mas marcar como opcional para o MVP (pode ser desativado por config)
  - **Manter** `shared/middleware.go`
  - **Manter** `router.go`

### 3.10 `core/config/` — Config Manager

- **Motivo:** O Config Manager completo com secrets AES-GCM e workspace isolation faz parte do MVP.
- **Ação:** **Manter** todos os arquivos: `manager.go`, `workspace_manager.go`, `secrets.go`, `types.go`.

### 3.11 `core/policies/` — Sistema de Políticas

- **Motivo:** As policies YAML com Guardian, Git passive, RAG priority, system integrity fazem parte do MVP.
- **Ação:** **Manter** todos os arquivos: `guardian.go`, `loader.go`, `schema.go`, `rules/*.yaml`.

### 3.12 `core/ingestion/` — Pipeline de Ingestão

- **Motivo:** O pipeline de indexação on-demand com parser, dependency graph e indexer faz parte do MVP.
- **Ação:** **Manter** todos os arquivos: `indexer.go`, `parser.go`, `dependency_graph.go`.

### 3.13 `core/telemetry/` — Telemetria

- **Motivo:** O logger rotativo com slog faz parte do MVP para observabilidade.
- **Ação:** **Manter** todos os arquivos: `logger.go`, `rotation.go`.

### 3.14 `core/git/` — Git Manager

- **Motivo:** O GitBridge com snapshots atômicos faz parte do MVP (conforme `POLICIES.md` → `git-passive.yaml`).
- **Ação:** **Manter** `manager.go`.

### 3.15 `core/tools/` vs `internal/tools/` — Duplicata de Ferramentas

- **Motivo:** Existem implementações de ferramentas em ambos os diretórios.
- **Ação:** Consolidar em **um único local**. Recomendação: manter `core/tools/` como implementação principal (mais completa, com `sanitizer.go`) e remover `internal/tools/` ou usar `internal/tools/` como referência e deletar `core/tools/`. **Não remover ferramentas** — apenas deduplicar.

### 3.16 `core/storage/` vs `internal/db/` — Duplicata de Storage

- **Motivo:** Existem abstrações de storage em ambos os diretórios.
- **Ação:** Consolidar em **um único local**. Recomendação: manter `core/storage/` (mais completo, com `engine.go`, `chunker.go`, `bbolt_store.go`, `chromem_store.go`, `types.go`) e remover `internal/db/` ou usar como referência. **Não remover storage** — apenas deduplicar.

### 3.17 `core/llm/` vs `internal/llm/` — Duplicata de LLM

- **Motivo:** Existem providers LLM em ambos os diretórios.
- **Ação:** Consolidar em **um único local**. Recomendação: manter `core/llm/` (mais completo, com `provider.go`, `router.go`, `prompt_factory.go`, `context_manager.go`, `gemini_provider.go`) e remover `internal/llm/` ou usar como referência. **Não remover LLM** — apenas deduplicar.

### 3.18 `core/engine/` — Gerenciador de Motor

- **Arquivo:** `engine.go`
- **Motivo:** Verificar se duplica lógica de `core/llm/` ou `core/storage/`.
- **Ação:** Se redundante, remover. Se for o orquestrador central (Engine que une tudo), **manter**.

### 3.19 `internal/os/` — OS Manager

- **Motivo:** O OS Manager é necessário para paths multi-plataforma e single-instance detection.
- **Ação:** **Manter**, mas simplificar:
  - Remover: registro no Windows Registry (shortcuts), `.desktop` files no Linux
  - Manter: detecção de paths (`~/.Vectora` / `%APPDATA%`), single-instance lock, detecção de instância única

### 3.20 `cmd/tests/` — Binário de Testes

- **Arquivo:** `main.go`
- **Motivo:** Testes devem rodar via `go test ./...`, não via binário dedicado.
- **Ação:** Remover.

---

## 4. `go.mod` — Dependências a Remover

| Dependência                          | Motivo                                                 |
| ------------------------------------ | ------------------------------------------------------ |
| `fyne.io/fyne/v2`                    | Sem Desktop GUI (Fyne)                                 |
| `github.com/charmbracelet/bubbles`   | Sem TUI Bubbletea                                      |
| `github.com/charmbracelet/bubbletea` | Sem TUI Bubbletea                                      |
| `github.com/charmbracelet/lipgloss`  | Sem TUI Bubbletea                                      |
| `github.com/gen2brain/beeep`         | Verificar se ainda usado pelo systray; se não, remover |
| `github.com/go-toast/toast`          | Verificar se ainda usado pelo systray; se não, remover |

**Nota:** `getlantern/systray` **NÃO remover** — é essencial para o systray do MVP.

---

## 5. Features Removidas (Resumo Funcional)

| Feature                                    | Status no Plano Original | Status no MVP                         |
| ------------------------------------------ | ------------------------ | ------------------------------------- |
| Vectora Index (marketplace de datasets)    | ✅                       | ❌ Removido                           |
| Vectora Desktop (Fyne)                     | ✅                       | ❌ Removido                           |
| Vectora TUI (Bubbletea)                    | ✅                       | ❌ Removido                           |
| Vectora Web                                | ✅ (futuro)              | ❌ Removido                           |
| Vectora Auth                               | ✅ (futuro)              | ❌ Removido                           |
| RBAC / Colaboração                         | ✅ (futuro)              | ❌ Removido                           |
| Setup Wizard complexo (Fyne + LPM/MPM)     | ✅                       | ❌ Simplificado para CLI              |
| LPM como binário separado                  | ✅                       | ❌ Simplificado ou integrado ao CLI   |
| MPM como binário separado                  | ✅                       | ❌ Simplificado ou integrado ao CLI   |
| Sistema de atualização de componentes      | ✅                       | ❌ Removido                           |
| Elevação de privilégios                    | ✅                       | ❌ Removido                           |
| Binário de testes dedicado                 | ✅                       | ❌ Removido (usar `go test`)          |
| gRPC (opcional)                            | ✅                       | ⚠️ Mantido mas desativado por default |
| Duplicatas de código (tools, storage, LLM) | ✅                       | ❌ Consolidar em um único local       |

---

## 6. Features Mantidas no MVP

| Componente                                                  | Tecnologia                                                                                                                                      |
| ----------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| CLI Cobra (`vectora start`, `vectora embed`, `vectora ask`) | Cobra CLI                                                                                                                                       |
| Systray                                                     | `getlantern/systray`                                                                                                                            |
| IPC (Named Pipes / Unix Socket)                             | `core/api/ipc/`                                                                                                                                 |
| i18n / Locale                                               | `internal/i18n/` (EN/PT/ES/FR)                                                                                                                  |
| Vector DB                                                   | chromem-go                                                                                                                                      |
| KV DB                                                       | bbolt                                                                                                                                           |
| Motor de IA                                                 | Gemini (cloud) + Qwen (local via llama.cpp)                                                                                                     |
| Ferramentas Agênticas                                       | `read_file`, `write_file`, `read_folder`, `edit`, `find_files`, `grep_search`, `run_shell_command`, `save_memory`, `google_search`, `web_fetch` |
| Protocolo                                                   | MCP (stdio) + ACP (JSON-RPC 2.0)                                                                                                                |
| Hard-Coded Guardian                                         | Filtro de arquivos sensíveis + regras YAML                                                                                                      |
| Configuração                                                | `core/config/` com AES-GCM                                                                                                                      |
| Logging                                                     | `core/telemetry/` com logger rotativo                                                                                                           |
| Git Snapshots                                               | `core/git/` (passivo, atômico)                                                                                                                  |
| Policies                                                    | `core/policies/` (scope, git, rag, integrity)                                                                                                   |
| Storage Engine                                              | `core/storage/` (dual-store + chunking)                                                                                                         |
| LLM Gateway                                                 | `core/llm/` (providers + context manager + prompt factory)                                                                                      |
| Ingestion                                                   | `core/ingestion/` (parser, dependency graph, indexer)                                                                                           |
| API                                                         | `core/api/` (JSON-RPC, gRPC opcional, IPC)                                                                                                      |
| OS Manager                                                  | `internal/os/` (paths multi-plataforma simplificados)                                                                                           |
| RAG Pipeline                                                | `internal/core/`                                                                                                                                |
| Trust Folder                                                | Escopo de operação padrão                                                                                                                       |

---

## 7. Ordem Sugerida de Remoção

1. **Remover `cmd/desktop/` e `cmd/tui/`** — elimina dependências Fyne/Bubbletea do `go.mod`
2. **Remover `internal/auth/`** — pacote vazio, sem uso
3. **Remover `cmd/tests/`** — usar `go test ./...`
4. **Remover `update.go`, `elevate_*.go`** do `cmd/daemon/`
5. **Simplificar `cmd/setup/`** — remover UI Fyne, manter CLI básico
6. **Simplificar/remover `cmd/lpm/` e `cmd/mpm/`** — integrar ao CLI ou remover
7. **Mover `internal/ipc/` → `core/api/ipc/`** e remover `update_handlers.go`
8. **Consolidar duplicatas**: `core/tools/` vs `internal/tools/`, `core/storage/` vs `internal/db/`, `core/llm/` vs `internal/llm/`
9. **Remover `PlanModeTool`** de `internal/tools/memory.go`
10. **Simplificar `internal/os/`** — remover registry shortcuts, .desktop files
11. **Limpar `go.mod`** — `go mod tidy` após todas as remoções
12. **Verificar compilação** — `go build ./...` após cada passo

---

> **Nota:** Este documento serve como guia de refatoração. Cada remoção deve ser feita com cuidado para não quebrar dependências cruzadas. Recomenda-se rodar `go build ./...` após cada passo de remoção para garantir que o projeto continua compilando. **Consolidar duplicatas é prioritário** — ter duas implementações de tools, storage e LLM gera confusão e bugs.
