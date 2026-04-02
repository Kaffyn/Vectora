# Plano de Arquitetura de Comunicação e API (Vectora)

Este plano traça o mapa definitivo de engenharia para o backend logístico do Vectora. Aqui integramos a espinha dorsal de Sockets com a exposição estendida e a infraestrutura local em Go.

## User Review Required

> [!IMPORTANT]
>
> - O Toolkit Agêntico (Read, Write, Memory, Shell) sofrerá interações diretas via IPC ou via ACP.
> - O Transporte de MCP servirá apenas requisições Standard I/O (StdIn/StdOut), significando que IDEs conectadas abrirão o daemon Vectora num container/stream isolado. Validar abordagem.

---

## Proposed Changes

### 1. Comunicação Interna Host (IPC: Inter-Process Communication)

_Responsável por ligar o Daemon ao Web UI (Next.js/Wails) e CLI sem trafegar na stack HTTP Rest, economizando overhead brutal de TCP networking._

#### [NEW] `internal/ipc/protocol.go`

- Protocolo `JSON-ND` assíncrono (Request/Response pareado por UUID).
- Formato rigoroso sem bloqueios.

#### [NEW] `internal/ipc/server.go` e `client.go`

- Escuta no SO Host: `~/.Vectora/run/vectora.sock` ou Pipe Global Win32.
- **Handlers Distribuídos (`internal/ipc/handlers`)**:
  - `workspace.query` (RAG puro)
  - `session.history`
  - `provider.set`

---

### 2. Integração Externa de Rede (HTTP APIs)

_Conexões out-of-socket desenhadas para internet ou sub-redes._

#### [NEW] `internal/index/client.go`

- Módulo HTTP cliente que consome o Vectora Index JSON.
- Proxy-Aware (Herda config corporativo local).
- Streamador de arquivos pesados (`.gguf` / bancos) que informa porcentagem via IPC Event Callbacks invés de segurar blocos de RAM enormes.

---

### 3. Integração Cross-Software (MCP & ACP)

_Padrões abertos arquitetados no repositório para expor o "Cérebro" de busca e edição do Vectora pra softwares de terceiros (Como Cursor e Plugins)._

#### [NEW] `internal/mcp/server.go` (Model Context Protocol)

- Padrão oficial OpenSource para IA (Empregado pelo Claude Desktop / Cursor).
- Escuta em **STDIO** (O processo Host invoca `vectora.exe --mcp`, e o daemon conversa exclusivamente via Pipe Padrão Terminal StdIn/Out usando JSON-RPC).
- Transforma os `Workspaces` do Vectora em Ferramentas Universais acessíveis pela IDE local.

#### [NEW] `internal/acp/agent.go` (Agent Context Protocol)

- Definição do Agente e empacotador isolado de contexto, abstraindo decisões para o "Agente interno" tomar caso seja chamado por fora.

---

### 4. Toolkit Engine (Cinto de Ferramentas da IA)

_A biblioteca utilitária `internal/tools`. Os "Braços" que a LLM executa. Passarão pelo sistema de Undo em conformidade._

#### [NEW] `internal/tools/engine.go`

- O Registrador (`registry` de chamadas suportadas). Converte Payload do LLM Toolcall (Langchaingo) em structs físicas de Go.

#### Mapeamento de Tools OBRIGATÓRIAS O.S

1. **[NEW] `internal/tools/filesystem.go`**
   - `read_file`, `write_file`, `read_folder`, `edit`.
2. **[NEW] `internal/tools/search.go`**
   - `find_files`, `grep_search`.
3. **[NEW] `internal/tools/system.go`**
   - `run_shell_command` (Execução isolada de Terminal streaming output).
4. **[NEW] `internal/tools/memory.go`**
   - `save_memory` (Chave-Valor no BBolt).
   - `enter_plan_mode` (Ação Meta onde IA decide se fragmentar p/ resolver tarefa).
5. **[NEW] `internal/tools/web.go`**
   - `google_search` e `web_fetch`.

## Open Questions

> [!WARNING]
> Foi removida a menção de `GitBridge` como dependência externa via prompt nas configs do README. Pretende manter Backup local raw (cópia crua na pasta backups) para os reverts das modificações do SDK (`write_file`, `run_shell_command`) ao invés do Git Bridge antigo?

## Verification Plan

1. Teste de `shell_stream`: Simularemos um `ping 8.8.8.8` gerando fluxo pro LLM.
2. Testes STDIO do lado MCP consumindo pacote falso local.
3. Teste exaustivo IPC Sockets para Zero-Delay.
