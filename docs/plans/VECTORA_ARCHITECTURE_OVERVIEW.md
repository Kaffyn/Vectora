# VECTORA: ARQUITETURA DE SISTEMA

**Status:** Visão Geral Consolidada — Todos os Componentes
**Versão:** 1.0
**Data:** 2026-04-05
**Idioma:** Português (PT-BR)
**Escopo:** Mapa completo de arquitetura do Vectora

---

## DIAGRAMA GERAL DO SISTEMA

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          VECTORA v2.0 - Arquitetura                         │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         CAMADA DE APRESENTAÇÃO                               │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Web UI    │  │  CLI (TUI)   │  │ MCP Server   │  │  Setup Installer │  │
│  │  (Wails +   │  │ (Bubbletea)  │  │ (Cursor/VS)  │  │   (Fyne GUI)     │  │
│  │  Next.js)   │  │              │  │              │  │                  │  │
│  │  4 Abas     │  │  Chat mode   │  │  Knowledge   │  │  8 Screens       │  │
│  │             │  │  Terminal    │  │  exposure    │  │  Wizard Flow     │  │
│  └──────┬──────┘  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘  │
│         │                 │                  │                  │             │
│         └─────────────────┼──────────────────┼──────────────────┘             │
│                           │ IPC (JSON-ND)    │ Process Spawn                 │
│                    ┌──────▼──────────────────▼──────┐                        │
│                    │  DAEMON CORE (Go)              │                        │
│                    │  ~100MB RAM idle               │                        │
│                    └──────┬───────────────┬──────────┘                        │
│                           │               │                                  │
└───────────────────────────┼───────────────┼──────────────────────────────────┘
                            │               │
         ┌──────────────────┴─┬─────────────┴──────────────┬─────────────────┐
         │                    │                            │                  │
┌────────▼────────┐  ┌────────▼────────┐     ┌────────────▼─────────┐  ┌──────▼──────┐
│  IPC Server     │  │  Core Engine    │     │   Package Manager    │  │  External   │
│                 │  │                 │     │                      │  │  Integration│
│ • Socket/Pipe   │  │ • RAG Pipeline  │     │ • LPM Controller     │  │             │
│ • Message Route │  │ • Workspace Mgr │     │ • MPM Controller     │  │ • Gemini API│
│ • Rate Limit    │  │ • Tool Registry │     │ • Setup Integration  │  │ • HF Hub    │
│ • Logging       │  │ • IPC Handlers  │     │ • Config Management  │  │ • llama.cpp │
│                 │  │                 │     │                      │  │   (sidecar) │
└─────────────────┘  └────────┬────────┘     └──────────┬───────────┘  └─────────────┘
                               │                         │
         ┌─────────────────────┴─────────────┬───────────┴──────────┐
         │                                   │                      │
┌────────▼──────────────┐        ┌────────────▼────────┐   ┌────────▼──────────┐
│  Storage Layer        │        │  LLM Providers      │   │  Tools & ACP      │
│                       │        │                     │   │                   │
│ • bbolt (metadata)    │        │ • Gemini Provider   │   │ • Tool Registry   │
│ • chromem-go (vectors)│        │ • Qwen Provider     │   │ • Filesystem      │
│ • Git Bridge          │        │ • Provider Abstract │   │ • Web Search      │
│ • Snapshots (undo)    │        │ • Health Checks     │   │ • Shell Command   │
│ • Audit Logs          │        │ • Fallback Logic    │   │ • Error Handling  │
│                       │        │                     │   │ • Security Rules  │
└───────────────────────┘        └─────────────────────┘   └───────────────────┘
```

---

## 1. TOPOLOGIA DE COMPONENTES

### 1.1 Daemon Central (cmd/vectora)

**Responsabilidade:** Orquestrador de todo sistema

```
┌─ Daemon (Go) ─────────────────────────┐
│                                        │
│  ┌─ Entry Point ────────────────────┐ │
│  │ • main() setup                   │ │
│  │ • Logger initialization          │ │
│  │ • Config loading                 │ │
│  │ • IPC server start               │ │
│  │ • Component initialization       │ │
│  └──────────────────────────────────┘ │
│                                        │
│  ┌─ IPC Server ──────────────────────┐ │
│  │ • Listen(socket/pipe)            │ │
│  │ • Accept connections             │ │
│  │ • Route messages                 │ │
│  │ • Handle events                  │ │
│  └──────────────────────────────────┘ │
│                                        │
│  ┌─ Core Components ─────────────────┐ │
│  │ • RAG Engine                     │ │
│  │ • Workspace Manager              │ │
│  │ • Tool Registry                  │ │
│  │ • LLM Selector                   │ │
│  │ • Package Manager Controller     │ │
│  └──────────────────────────────────┘ │
│                                        │
└────────────────────────────────────────┘
```

**Performance Target:**
- Boot time: < 2 segundos
- Memory idle: < 100MB
- Memory com 5 workspaces: < 500MB
- IPC latency: < 1ms

---

### 1.2 Web UI (cmd/vectora-app + internal/app)

**Stack:** Wails v3 + Next.js 14 + TailwindCSS

```
┌─ Vectora App ──────────────────────────┐
│  (Wails Desktop Window)                │
│                                         │
│  ┌─ Navigation ─────────────────────┐  │
│  │ • Sidebar (4 abas)               │  │
│  │ • Header (contexto)              │  │
│  │ • Modals/Dialogs                 │  │
│  └──────────────────────────────────┘  │
│                                         │
│  ┌─ Aba: Chat ───────────────────────┐ │
│  │ • ChatFeed (histórico)           │ │
│  │ • InputArea (send query)         │ │
│  │ • Workspace selector             │ │
│  │ • Streaming respostas            │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Aba: Código ──────────────────────┐ │
│  │ • FileTree (file browser)        │ │
│  │ • CodeEditor (Monaco)            │ │
│  │ • Terminal (integrado)           │ │
│  │ • DiffViewer                     │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Aba: Index ───────────────────────┐ │
│  │ • WorkspaceList (CRUD)           │ │
│  │ • WorkspaceDetail (chunks)       │ │
│  │ • DatasetBrowser (HF Index)      │ │
│  │ • Upload interface               │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Aba: Manager ─────────────────────┐ │
│  │ • LPM Panel (builds)             │ │
│  │ • MPM Panel (models)             │ │
│  │ • Configuration                  │ │
│  │ • Progress monitoring            │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ State Management ────────────────┐ │
│  │ • Zustand stores (4+)            │ │
│  │ • IPC hooks                      │ │
│  │ • Error boundaries               │ │
│  └──────────────────────────────────┘ │
│                                         │
└─────────────────────────────────────────┘
```

**Performance Target:**
- Build size: < 50MB
- Boot time: < 1.5 segundos
- Frame rate: 60 FPS (smooth)
- Memory: < 200MB

---

### 1.3 CLI/TUI (cmd/vectora/bubbletea.go)

**Stack:** Bubbletea + Lipgloss + Bubbles

```
┌─ Vectora CLI ──────────────────────────┐
│  (Terminal User Interface)             │
│                                         │
│  ┌─ Commands ────────────────────────┐ │
│  │ • vectora chat --workspace godot │ │
│  │ • vectora query "..."            │ │
│  │ • vectora index /path/to/files   │ │
│  │ • vectora undo                   │ │
│  │ • vectora --tests                │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Components ──────────────────────┐ │
│  │ • Viewport (scrollable history)  │ │
│  │ • Text Input (com histórico)     │ │
│  │ • Spinner (animado)              │ │
│  │ • Progress Bar (indexação)       │ │
│  │ • Markdown renderer              │ │
│  └──────────────────────────────────┘ │
│                                         │
└─────────────────────────────────────────┘
```

**Performance Target:**
- Binary size: < 10MB
- Boot: < 100ms
- Memory: < 10MB
- No-Color support para legacy terminals

---

### 1.4 Setup Installer (cmd/vectora-installer)

**Stack:** Fyne GUI (golang)

```
┌─ Setup Installer ──────────────────────┐
│  (First-run experience)                │
│                                         │
│  ┌─ Screen 1: Welcome ───────────────┐ │
│  │ • Logo                           │ │
│  │ • Explicação do Vectora          │ │
│  │ • Buttons: Next / Cancel         │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Screen 2: Hardware Detection ────┐ │
│  │ • CPU/RAM/GPU info               │ │
│  │ • Recomendações                  │ │
│  │ • Status: scanning...            │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Screens 3-4: LPM Setup ─────────┐ │
│  │ • Build selection (list)         │ │
│  │ • Download progress              │ │
│  │ • Verification (SHA256)          │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Screens 5-6: MPM Setup ─────────┐ │
│  │ • Model selection (Qwen3 list)   │ │
│  │ • Download progress              │ │
│  │ • Quantization choice            │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Screen 7: Configuration ─────────┐ │
│  │ • Gemini API key (optional)      │ │
│  │ • Desktop shortcut               │ │
│  │ • Auto-launch option             │ │
│  └──────────────────────────────────┘ │
│                                         │
│  ┌─ Screen 8: Success ───────────────┐ │
│  │ • Summary installed items        │ │
│  │ • Launch Vectora button          │ │
│  │ • Open docs button               │ │
│  └──────────────────────────────────┘ │
│                                         │
└─────────────────────────────────────────┘
```

**Performance Target:**
- Binary size: < 15MB
- HW detection: < 2s
- Download resumable

---

## 2. FLUXOS DE DADOS PRINCIPAIS

### 2.1 Fluxo de Chat (Query RAG)

```
User typing in Chat UI
         ↓
┌─ Web UI ────────────────────────────────────┐
│ Input validation                            │
│ Show placeholder                            │
└─────────────────────────────────────────────┘
         ↓
┌─ IPC Request ───────────────────────────────┐
│ Method: workspace.query                     │
│ Payload: { ws_id, query }                   │
│ Timeout: 60s                                │
└─────────────────────────────────────────────┘
         ↓
┌─ Daemon: RAG Pipeline ──────────────────────┐
│ 1. Embed query (384D vector)               │
│ 2. Search similar (Top-K=5)                │
│ 3. Construct context                       │
│ 4. Load tool specs (ACP)                   │
│ 5. Build final prompt                      │
│ 6. Call LLM provider                       │
│    ├─ Try Gemini (if configured)          │
│    └─ Fallback Qwen local (if available)  │
│ 7. Parse tool calls (if any)               │
│ 8. Stream response                         │
└─────────────────────────────────────────────┘
         ↓
┌─ IPC Response (Streaming) ──────────────────┐
│ Event: index.stream_chunk { chunk: "..." }  │
│ Multiple events until finish                │
│ Final: workspace.query_complete {sources} │
└─────────────────────────────────────────────┘
         ↓
┌─ Web UI: Update Display ────────────────────┐
│ Append chunks to message                    │
│ Render Markdown in real-time               │
│ Show sources badges                         │
│ Show tool calls visualization               │
└─────────────────────────────────────────────┘
```

**Latência Total Target:** < 2 segundos (até primeiro chunk)

---

### 2.2 Fluxo de Indexação

```
User clicks "Index" in Index tab
         ↓
┌─ Web UI: Modal ────────────────────────────┐
│ Select directory                           │
│ Input workspace name                       │
│ Click "Indexar"                           │
└────────────────────────────────────────────┘
         ↓
┌─ IPC Request ──────────────────────────────┐
│ Method: workspace.index                    │
│ Payload: { ws_id, path }                   │
└────────────────────────────────────────────┘
         ↓
┌─ Daemon: Index Pipeline ────────────────────┐
│ 1. Read files from directory               │
│ 2. Semantic chunking (max 512 tokens)      │
│ 3. For each chunk:                         │
│    ├─ Generate embedding (384D)            │
│    ├─ Store in chromem-go                  │
│    └─ Emit progress event                  │
│ 4. Update workspace metadata               │
│ 5. Set status = "done"                     │
└────────────────────────────────────────────┘
         ↓
┌─ IPC Events (Streaming) ────────────────────┐
│ index.progress {                            │
│   ws_id: "...",                            │
│   files_done: 10,                          │
│   files_total: 45,                         │
│   percent: 22                              │
│ }                                          │
└────────────────────────────────────────────┘
         ↓
┌─ Web UI: Progress Bar ──────────────────────┐
│ Update progress %                          │
│ Show file count                            │
│ Show ETA                                   │
│ Enable chat once done                      │
└────────────────────────────────────────────┘
```

**Velocidade Target:** 1000 arquivos em 2-5 minutos (conforme hardware)

---

### 2.3 Fluxo de Tool Execution

```
LLM decides to execute tool (e.g., write_file)
         ↓
┌─ Daemon: ACP Registry ──────────────────────┐
│ Check tool exists                          │
│ Validate arguments (JSON schema)           │
│ Check security rules:                      │
│  ├─ write_file → Always needs snapshot     │
│  ├─ shell_command → May need approval      │
│  └─ read_file → Usually OK                 │
└─────────────────────────────────────────────┘
         ↓
┌─ Daemon: GitBridge ─────────────────────────┐
│ snapshot(file_path) → hash antes           │
│ Store metadata em bbolt                    │
│ Return snapshot_id                         │
└─────────────────────────────────────────────┘
         ↓
┌─ Daemon: Tool Execution ────────────────────┐
│ Execute tool with arguments                │
│ Capture output/errors                      │
│ Measure execution time                     │
│ Log to audit trail                         │
└─────────────────────────────────────────────┘
         ↓
┌─ IPC Response ──────────────────────────────┐
│ Method: tool.execute response              │
│ Payload: {                                 │
│   result: "output text",                   │
│   snapshot_id: "uuid",                     │
│   execution_time_ms: 123                   │
│ }                                          │
└─────────────────────────────────────────────┘
         ↓
┌─ Web UI: Show Result ───────────────────────┐
│ Display tool output                        │
│ Show "Undo" button                         │
│ Add to message metadata                    │
└─────────────────────────────────────────────┘

User clicks "Undo"
         ↓
┌─ IPC Request ──────────────────────────────┐
│ Method: tool.undo                          │
│ Payload: { snapshot_id }                   │
└─────────────────────────────────────────────┘
         ↓
┌─ Daemon: GitBridge.Restore ─────────────────┐
│ Load snapshot metadata                     │
│ Restore file from hash                     │
│ Remove snapshot record                     │
└─────────────────────────────────────────────┘
```

---

### 2.4 Fluxo de Setup Installer

```
User runs vectora-setup.exe
         ↓
┌─ Setup Wizard: Welcome ─────────────────────┐
│ Screen 1: Show Vectora explanation         │
│ User clicks: Next                          │
└─────────────────────────────────────────────┘
         ↓
┌─ Setup Wizard: Hardware Detection ──────────┐
│ Screen 2: Detect CPU/RAM/GPU               │
│ Call LPMController.Detect()                │
│  └─ Run: lpm detect --json                 │
│ Show results                               │
│ Auto-recommend best build + model          │
└─────────────────────────────────────────────┘
         ↓
┌─ Setup Wizard: Build Selection ─────────────┐
│ Screen 3: Show list from lpm list --json   │
│ Highlight recommended build                │
│ User selects build (or accept rec.)        │
│ Click: Next                                │
└─────────────────────────────────────────────┘
         ↓
┌─ Setup Wizard: Build Download ──────────────┐
│ Screen 4: Call LPMController.Install()     │
│  └─ Run: lpm install --id <id> --json      │
│ Monitor progress from JSON events          │
│ Show progress bar                          │
│ Verify SHA256                              │
│ Click: Next                                │
└─────────────────────────────────────────────┘
         ↓
┌─ Setup Wizard: Model Selection ─────────────┐
│ Screen 5: Show list from mpm list --json   │
│ Highlight recommended model                │
│ User selects model + quantization          │
│ Click: Next                                │
└─────────────────────────────────────────────┘
         ↓
┌─ Setup Wizard: Model Download ──────────────┐
│ Screen 6: Call MPMController.Install()     │
│  └─ Run: mpm install --id <id> --json      │
│ Monitor progress                           │
│ Show progress bar                          │
│ Verify SHA256                              │
│ Click: Next                                │
└─────────────────────────────────────────────┘
         ↓
┌─ Setup Wizard: Configuration ───────────────┐
│ Screen 7: Show optional Gemini API key     │
│ Show desktop shortcut option               │
│ Show auto-launch option                    │
│ User enters/skips config                   │
│ Click: Next                                │
└─────────────────────────────────────────────┘
         ↓
┌─ Setup Wizard: Success ─────────────────────┐
│ Screen 8: Show summary:                    │
│  ✓ Llama.cpp [build_name]                 │
│  ✓ Qwen3 [model_name]                     │
│  ✓ Configuration saved                    │
│ Buttons: Launch Vectora / Close            │
│ User clicks: Launch Vectora                │
└─────────────────────────────────────────────┘
         ↓
┌─ Daemon Starts ─────────────────────────────┐
│ Load configuration                         │
│ Start IPC server                           │
│ Spawn llama-cli sidecar                    │
│ Ready for Web UI                           │
└─────────────────────────────────────────────┘
         ↓
┌─ Web UI Launches ───────────────────────────┐
│ Show Chat tab                              │
│ Ready for queries                          │
└─────────────────────────────────────────────┘
```

---

## 3. ESTRUTURA DE DIRETÓRIOS

### 3.1 Estrutura Completa do Repositório

```
vectora/
├── cmd/
│   ├── vectora/                    # Daemon principal + CLI
│   │   ├── main.go                 # Entry point
│   │   ├── app.go                  # App struct (Wails)
│   │   ├── cli.go                  # Bubbletea setup
│   │   └── flags.go                # Flag definitions
│   ├── vectora-app/                # Web UI (Wails wrapper)
│   │   ├── main.go                 # Wails entry
│   │   └── frontend.go             # Next.js embedding
│   └── vectora-installer/          # Setup Installer (Fyne)
│       ├── main.go                 # Fyne app entry
│       ├── wizard.go               # State machine
│       └── screens/                # 8 screen implementations
│
├── internal/
│   ├── infra/                      # Infrastructure
│   │   ├── config.go               # Config loader
│   │   ├── logger.go               # Structured logging
│   │   └── env.go                  # Env variable handling
│   │
│   ├── ipc/                        # Inter-Process Communication
│   │   ├── server.go               # IPC server
│   │   ├── message.go              # Message structures
│   │   └── handlers.go             # Request handlers
│   │
│   ├── db/                         # Data persistence
│   │   ├── store.go                # bbolt wrapper
│   │   ├── vector.go               # chromem-go wrapper
│   │   └── migrations.go           # Schema migrations
│   │
│   ├── core/                       # RAG & Engine
│   │   ├── rag_pipeline.go         # RAG orchestration
│   │   ├── workspace.go            # Workspace manager
│   │   └── chunking.go             # Semantic chunking
│   │
│   ├── llm/                        # LLM Providers
│   │   ├── provider.go             # Provider interface
│   │   ├── gemini.go               # Gemini implementation
│   │   ├── qwen.go                 # Qwen sidecar wrapper
│   │   └── manager.go              # Provider selection
│   │
│   ├── engines/                    # Llama.cpp Management
│   │   ├── catalog.go              # Build catalog (embedded)
│   │   ├── detector.go             # Hardware detection
│   │   ├── downloader.go           # Download manager
│   │   ├── manager.go              # Lifecycle management
│   │   └── process.go              # Subprocess control
│   │
│   ├── acp/                        # Autonomous Control Protocol
│   │   ├── registry.go             # Tool registry
│   │   ├── executor.go             # Tool execution
│   │   └── security.go             # Security rules
│   │
│   ├── tools/                      # Tool Implementations
│   │   ├── filesystem.go           # read/write/find
│   │   ├── web.go                  # search/fetch
│   │   ├── shell.go                # shell_command
│   │   └── git.go                  # git operations
│   │
│   ├── git/                        # GitBridge (Snapshots)
│   │   ├── bridge.go               # Snapshot manager
│   │   ├── store.go                # Snapshot storage
│   │   └── restore.go              # Rollback logic
│   │
│   └── app/                        # Web UI (Next.js)
│       ├── app/                    # Next.js app router
│       │   ├── (main)/
│       │   │   ├── layout.tsx      # Root layout
│       │   │   └── page.tsx        # Index
│       │   ├── chat/
│       │   ├── codigo/
│       │   ├── index/
│       │   └── manager/
│       ├── components/             # React components
│       │   ├── Chat/
│       │   ├── Codigo/
│       │   ├── Index/
│       │   ├── Manager/
│       │   ├── Common/
│       │   └── UI/
│       ├── hooks/                  # Custom hooks
│       ├── store/                  # Zustand stores
│       ├── services/               # API services
│       ├── utils/                  # Utilities
│       ├── styles/                 # CSS + Tailwind
│       └── public/                 # Static assets
│
├── pkg/                            # Shared packages
│   ├── types/                      # Shared types
│   └── utils/                      # Shared utilities
│
├── tests/                          # Test suites
│   ├── integration/
│   ├── e2e/
│   ├── fixtures/
│   └── mocks/
│
├── build/                          # Build outputs (ignored)
│   ├── vectora
│   ├── vectora-app
│   └── vectora-installer
│
├── docs/                           # Documentation
│   ├── README.md
│   ├── README.pt.md
│   ├── CONTRIBUTING.md
│   ├── CONTRIBUTING.pt.md
│   └── architecture/
│
├── scripts/                        # Build scripts
│   └── embed_assets.sh
│
├── Makefile                        # Unix build
├── build.ps1                       # Windows build
├── go.mod
├── go.sum
├── package.json (Web UI)
├── next.config.js
├── tailwind.config.js
└── tsconfig.json
```

### 3.2 Home Directory User Data

```
%USERPROFILE%/.Vectora/
├── .env                            # Configuration
├── data/
│   ├── vectora.db                  # bbolt (metadata)
│   └── chroma_ws_{id}/             # chromem-go collections
├── engines/
│   ├── catalog.json                # Available builds
│   ├── llama-cpp-{version}/        # Installed builds
│   └── qwen-{model}.gguf           # Model files
├── backups/
│   └── {uuid}-*.bak                # Snapshots
├── logs/
│   ├── daemon.log*
│   ├── wails.log*
│   └── ipc.log*
├── run/
│   └── vectora.sock                # IPC socket (Unix)
└── temp/
    └── downloads/                  # Temporary files
```

---

## 4. PROTOCOLOS E INTERFACES

### 4.1 IPC Message Protocol

```
┌─ Request ────────────────────────────────┐
│ {                                        │
│   "id": "req-1-1234567890",             │
│   "type": "request",                    │
│   "method": "workspace.query",          │
│   "payload": {                          │
│     "ws_id": "godot-4.3",               │
│     "query": "What is viewport?"        │
│   }                                     │
│ }                                       │
└──────────────────────────────────────────┘
           ↓ (30s timeout)
┌─ Response ───────────────────────────────┐
│ {                                        │
│   "id": "req-1-1234567890",             │
│   "type": "response",                   │
│   "payload": {                          │
│     "answer": "Viewport is...",         │
│     "sources": [...],                   │
│     "thinking": "..."                   │
│   },                                    │
│   "error": null                         │
│ }                                       │
└──────────────────────────────────────────┘

┌─ Event (Push) ───────────────────────────┐
│ {                                        │
│   "id": "evt-uuid",                     │
│   "type": "event",                      │
│   "method": "index.progress",           │
│   "payload": {                          │
│     "ws_id": "godot",                   │
│     "files_done": 12,                   │
│     "files_total": 45,                  │
│     "percent": 26                       │
│   }                                     │
│ }                                       │
└──────────────────────────────────────────┘
```

### 4.2 Main IPC Methods

| Method | Direction | Payload In | Response | Purpose |
|--------|-----------|-----------|----------|---------|
| `workspace.query` | Request | `{ws_id, query}` | `{answer, sources, thinking}` | RAG query |
| `workspace.index` | Request | `{ws_id, path}` | `{job_id, status}` | Index files |
| `workspace.create` | Request | `{name, description}` | `{ws_id}` | Create WS |
| `workspace.list` | Request | `{}` | `{workspaces}` | List all |
| `tool.execute` | Request | `{tool, args}` | `{result, snapshot_id}` | Run tool |
| `tool.undo` | Request | `{snapshot_id}` | `{success}` | Undo change |
| `package.lpm_list` | Request | `{}` | `{builds}` | LPM list |
| `package.lpm_install` | Request | `{build_id}` | `{job_id}` | LPM install |
| `package.mpm_list` | Request | `{}` | `{models}` | MPM list |
| `package.mpm_install` | Request | `{model_id}` | `{job_id}` | MPM install |
| `config.get` | Request | `{}` | `{config}` | Get settings |
| `config.set` | Request | `{config}` | `{success}` | Update settings |
| `index.progress` | Event | N/A | N/A | Index progress (push) |
| `index.completed` | Event | N/A | N/A | Index done (push) |
| `tool_completed` | Event | N/A | N/A | Tool result (push) |

---

## 5. MATRIZ DE DEPENDÊNCIAS

```
┌─ Web UI ────────────────────────┐
│ Depends: IPC Client             │
│ Calls: workspace.*, tool.*, etc │
└────────────────┬────────────────┘
                 │
         ┌───────▼────────┐
         │  IPC Server    │
         │  (Daemon)      │
         └───────┬────────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
    │            │            │
┌───▼────┐  ┌───▼───────┐  ┌─▼──────┐
│  Core  │  │ LLM/Tools │  │ Package│
│  RAG   │  │ Provider  │  │ Manager│
└───┬────┘  └───┬───────┘  └─┬──────┘
    │           │            │
┌───▼─────────────────────────▼──┐
│  Storage Layer                 │
│  (bbolt + chromem-go)          │
└────────────────────────────────┘
    │           │
┌───▼────┐  ┌───▼──────┐
│Qwen    │  │Gemini    │
│sidecar │  │API       │
└────────┘  └──────────┘
```

---

## 6. CICLO DE VIDA DE EXECUÇÃO

### Boot do Daemon

```
1. Parse flags + env vars
2. Load config from ~/.Vectora/.env
3. Initialize logger
4. Connect to databases (bbolt, chromem-go)
5. Initialize IPC server (socket/pipe)
6. Initialize component managers:
   ├─ LLM Manager
   ├─ Engine Manager
   ├─ Workspace Manager
   ├─ Tool Registry
   └─ ACP Registry
7. Start listening on IPC
8. Ready for connections
   └─ Total time: ~2 segundos
```

### Query Completa (Início ao Fim)

```
t=0ms     User types in Chat
t=100ms   User presses Enter
t=150ms   IPC request sent
t=200ms   Daemon receives request
t=250ms   Embed query (384D)
t=350ms   Search Top-K=5 chunks
t=400ms   Construct context
t=450ms   Build prompt
t=500ms   Call LLM provider
t=800ms   LLM responds (streaming)
t=1000ms  First chunk in UI
t=1500ms  Full response streamed
t=2000ms  Final response complete
        + Sources available
        + Ready for next query
```

---

## 7. DECISÕES ARQUITETURAIS

### 7.1 Por que Go para Backend?

- ✅ Performance (< 1ms IPC latency)
- ✅ Lightweight (100MB daemon idle)
- ✅ Concurrency (goroutines)
- ✅ Binary size (5MB core)
- ✅ Static typing (safety)

### 7.2 Por que Wails + Next.js para UI?

- ✅ Desktop native (WebView2)
- ✅ Modern web stack (React, TypeScript)
- ✅ Static export (no Node.js runtime)
- ✅ Performance (fast builds)
- ✅ Developer experience

### 7.3 Por que IPC em vez de HTTP?

- ✅ No network overhead
- ✅ Local-only security
- ✅ Socket/pipe flexibility
- ✅ Zero external dependencies
- ✅ Deterministic (testable)

### 7.4 Por que chromem-go para Vetores?

- ✅ In-process (no separate service)
- ✅ Lightweight (embarcado)
- ✅ Isolamento por workspace
- ✅ Semanticamente correto
- ✅ Sem cloud dependency

### 7.5 Por que CLI-only para MPM/LPM?

- ✅ Reutilizável (Setup + App)
- ✅ Testável (JSON output)
- ✅ Pequeno (< 20MB)
- ✅ Sem overhead de GUI
- ✅ Composable (unix philosophy)

---

## 8. SEGURANÇA E ISOLAMENTO

### 8.1 Isolamento de Dados

```
Workspace A (Godot)
└─ Vector Collection (chromem-go)
   └─ 2540 chunks
   └─ ZERO acesso de Workspace B

Workspace B (Physics Papers)
└─ Vector Collection (chromem-go)
   └─ 1200 chunks
   └─ ZERO acesso de Workspace A
```

### 8.2 Tool Execution Security

```
User Input → Validation → ACP Registry → Security Check
                                              ↓
                                    ┌─────────┴─────────┐
                                    │                   │
                            write_file (risco alto)  read_file (OK)
                                    │                   │
                            GitBridge.Snapshot()    Execute
                                    │                   │
                            Execute Tool          Return Result
                                    │
                            Store Snapshot (Undo)
```

### 8.3 API Key Protection

```
.env file (encrypted on disk via OS)
     ↓
ConfigLoader (in-memory only)
     ↓
Gemini API (never logged)
     ↓
Responses only (no key in logs)
```

---

## CONCLUSÃO

Esta arquitetura é otimizada para:
- **Performance:** Latência mínima, RAM eficiente
- **Privacidade:** Zero cloud para dados locais
- **Usabilidade:** Interfaces intuitivas
- **Extensibilidade:** Plugins de tools via ACP
- **Manutenibilidade:** Componentização clara

**Status:** ✅ Pronto para Implementação
**Próximo Passo:** Kick-off da Fase 1 (Infraestrutura)

---

