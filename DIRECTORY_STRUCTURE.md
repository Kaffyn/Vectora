# VECTORA - Complete Directory Structure

**Status:** Structure initialized and verified
**Date:** 2026-04-05
**Total Directories:** 29 core packages
**Total Scaffold Files:** 60+

---

## Directory Hierarchy

```
vectora/
├── build/                            # Build output and artifacts
├── cmd/                              # Command-line applications
│   ├── vectora/                      # Main daemon orchestrator
│   ├── vectora-app/                  # Desktop app wrapper (Wails)
│   └── vectora-installer/            # Windows setup wizard (Fyne)
│
├── internal/                         # Private packages (not exported)
│   ├── infra/                        # Infrastructure & configuration
│   ├── ipc/                          # Inter-process communication
│   ├── db/                           # Data persistence layer
│   ├── core/                         # Core business logic
│   ├── llm/                          # LLM provider abstraction
│   ├── engines/                      # AI processing engines (RAG, cache)
│   ├── acp/                          # Anthropic Compute Protocol
│   ├── tools/                        # Built-in tools (filesystem, web, shell)
│   ├── git/                          # Git integration & snapshots
│   └── app/                          # Frontend application (Next.js + React)
│
├── pkg/                              # Public packages
│   ├── types/                        # Shared type definitions
│   └── utils/                        # Utility functions
│
├── tests/                            # Test suite
│   ├── integration/                  # Integration tests
│   ├── e2e/                          # End-to-end tests
│   ├── fixtures/                     # Test fixtures
│   └── mocks/                        # Mock objects
│
├── docs/                             # Documentation
├── scripts/                          # Build and utility scripts
│
└── Root Configuration Files
    ├── .env                          # Environment variables
    ├── .gitignore                    # Git ignore rules
    ├── go.mod                        # Go module definition
    ├── go.sum                        # Go dependencies
    ├── Makefile                      # Unix build targets
    ├── build.ps1                     # Windows build script
    ├── package.json                  # Node.js dependencies
    ├── tsconfig.json                 # TypeScript configuration
    ├── next.config.js                # Next.js configuration
    └── tailwind.config.js            # Tailwind CSS configuration
```

---

## Package Organization

### Command Line Applications (cmd/)

#### cmd/vectora - Main Daemon
```
cmd/vectora/
├── main.go          # Entry point, setup
├── app.go           # Application initialization
├── cli.go           # CLI command handling
└── flags.go         # Command-line flags parsing
```
**Responsibility:** Orchestrate the entire Vectora system
- Initialize logging and configuration
- Start IPC server for client communication
- Manage component lifecycle
- Handle CLI commands

#### cmd/vectora-app - Desktop Wrapper
```
cmd/vectora-app/
├── main.go          # Wails app entry
└── frontend.go      # Frontend integration
```
**Responsibility:** Provide desktop application wrapper
- Wails framework integration
- Desktop window management
- Native system integration

#### cmd/vectora-installer - Setup Wizard
```
cmd/vectora-installer/
├── main.go          # Installer entry
└── wizard.go        # Setup flow
```
**Responsibility:** Windows installation and setup
- Guided installation wizard (8 screens)
- System configuration
- Dependency installation

---

### Internal Packages (internal/)

#### internal/infra - Infrastructure
```
internal/infra/
├── config.go        # Configuration management
├── logger.go        # Structured logging
└── env.go           # Environment variables
```
**Responsibility:** System infrastructure
- Configuration loading and management
- Structured logging (JSON format)
- Environment variable handling
- System notifications

#### internal/ipc - Inter-Process Communication
```
internal/ipc/
├── server.go        # IPC server (socket/pipe)
├── message.go       # Message protocol
├── handlers.go      # Request handlers
├── router.go        # Message routing
├── client.go        # Client library
├── protocol.go      # Protocol definitions
└── ipc_test.go      # Tests
```
**Responsibility:** Client-daemon communication
- JSON-ND protocol over Unix sockets/Named pipes
- Message routing and dispatching
- Rate limiting and logging
- Request/response handling

#### internal/db - Data Persistence
```
internal/db/
├── db.go            # Database interface
├── migrations.go    # Schema migrations
├── query.go         # Query builder
├── store.go         # Storage interface
├── vector.go        # Vector operations
├── memory_service.go # In-memory implementation
├── interfaces.go    # Abstract interfaces
└── db_test.go       # Tests
```
**Responsibility:** Data persistence layer
- bbolt for metadata and chat history
- chromem-go for vector embeddings
- Database migrations and schema
- Query interface

#### internal/core - Core Engine
```
internal/core/
├── engine.go        # Main execution engine
├── workspace.go     # Workspace management
├── registry.go      # Tool/component registry
└── rag_pipeline.go  # RAG implementation
```
**Responsibility:** Core business logic
- RAG pipeline orchestration
- Workspace management
- Tool registry maintenance
- Request processing and execution

#### internal/llm - LLM Provider Abstraction
```
internal/llm/
├── provider.go      # Provider interface
├── gemini.go        # Google Gemini provider
├── qwen.go          # Qwen LLM provider
├── service.go       # LLM service orchestration
├── messages.go      # Message types
└── protocol_llama.go # Llama protocol support
```
**Responsibility:** LLM integration
- Provider abstraction interface
- Gemini API integration
- Qwen LLM support
- Health checks and fallback logic
- Token counting and cost tracking

#### internal/engines - AI Processing Engines
```
internal/engines/
├── rag.go           # RAG pipeline
├── cache.go         # Response caching
└── fallback.go      # Fallback strategy
```
**Responsibility:** AI processing engines
- RAG (Retrieval-Augmented Generation) pipeline
- Response caching for performance
- Fallback logic for LLM failures
- Vector similarity search

#### internal/acp - Anthropic Compute Protocol
```
internal/acp/
├── tool.go          # Tool registry
├── security.go      # Security rules enforcement
├── executor.go      # Tool execution
└── agent.go         # Agent logic
```
**Responsibility:** Tool execution framework
- Tool registry and discovery
- Security rules enforcement
- Safe tool execution
- Agent/agentic reasoning support

#### internal/tools - Built-in Tools
```
internal/tools/
├── filesystem.go    # File operations
├── web.go           # Web search
└── shell.go         # Shell command execution
```
**Responsibility:** Built-in capabilities
- Filesystem access (read, write, search)
- Web search integration
- Shell command execution
- Rate limiting per tool

#### internal/git - Git Integration
```
internal/git/
├── bridge.go        # Git command wrapper
├── operations.go    # Git operations
└── snapshot.go      # Workspace snapshots
```
**Responsibility:** Git integration
- Workspace version control
- Git operations wrapper
- Snapshot creation/restore (undo)
- Commit history management

#### internal/app - Frontend Application
```
internal/app/
├── app/
│   ├── page.tsx     # Root page
│   ├── layout.tsx   # Root layout
│   └── globals.css  # Global styles
├── components/      # React components
│   ├── Header.tsx
│   ├── Sidebar.tsx
│   └── ChatPanel.tsx
├── hooks/           # Custom React hooks
│   ├── useChat.ts
│   └── useWorkspace.ts
├── store/           # State management (Zustand/Redux)
│   ├── store.ts
│   ├── chat.ts
│   └── workspace.ts
├── services/        # API services
│   ├── api.ts       # REST/gRPC client
│   └── ipc.ts       # IPC communication
├── utils/           # Utility functions
│   ├── formatters.ts
│   └── validators.ts
├── styles/          # CSS modules
│   ├── theme.css
│   └── components.css
└── public/          # Static assets
    └── favicon.ico
```
**Responsibility:** Web-based user interface
- Next.js 15+ with React 19
- TypeScript for type safety
- Tailwind CSS for styling
- IPC communication with daemon
- 4 main tabs (Chat, Files, Settings, Extensions)

---

### Public Packages (pkg/)

#### pkg/types - Type Definitions
```
pkg/types/
├── types.go         # Core domain types
├── messages.go      # Message types
└── config.go        # Configuration structures
```
**Responsibility:** Shared type definitions
- Domain models
- Message protocol types
- Configuration structures
- Exported interfaces

#### pkg/utils - Utilities
```
pkg/utils/
├── string.go        # String utilities
├── file.go          # File utilities
└── validation.go    # Validation helpers
```
**Responsibility:** Utility functions
- String manipulation
- File operations
- Input validation
- Common helpers

---

### Test Suite (tests/)

```
tests/
├── integration/     # Integration tests
│   └── integration_test.go
├── e2e/             # End-to-end tests
│   └── e2e_test.go
├── fixtures/        # Test fixtures
│   └── fixtures.go
└── mocks/           # Mock implementations
    └── mocks.go
```

**Test Categories:**
- **Unit Tests:** Alongside source code files (e.g., `db_test.go`)
- **Integration Tests:** IPC communication, database operations
- **E2E Tests:** Full system workflows
- **Fixtures:** Sample data and configurations
- **Mocks:** Interface implementations for testing

---

## Development Workflow

### Phase 1: Backend Implementation (Go)
1. **Initialize Module:**
   ```bash
   go mod init github.com/vectora/vectora
   go get github.com/hashicorp/go-hclog
   go get github.com/etcd-io/bbolt
   go get github.com/go-echarts/go-echarts
   ```

2. **Implement Infrastructure:**
   - `internal/infra/config.go` - Configuration loading
   - `internal/infra/logger.go` - Structured logging
   - `internal/infra/env.go` - Environment variables

3. **Implement IPC Layer:**
   - `internal/ipc/server.go` - Socket/pipe server
   - `internal/ipc/message.go` - Protocol definition
   - `internal/ipc/handlers.go` - Request handling

4. **Implement Core Engine:**
   - `internal/core/engine.go` - Main orchestrator
   - `internal/core/workspace.go` - Workspace management
   - `internal/core/registry.go` - Tool registry

5. **Add Persistence:**
   - `internal/db/db.go` - Database interface
   - `internal/db/migrations.go` - Schema setup
   - `internal/db/vector.go` - Vector operations

### Phase 2: Frontend Setup (Next.js)
1. **Initialize Next.js:**
   ```bash
   npm install next react react-dom typescript
   npm install -D tailwindcss postcss autoprefixer
   npm install zustand axios
   ```

2. **Configure TypeScript:**
   - Edit `tsconfig.json`
   - Set strict mode enabled

3. **Setup Tailwind CSS:**
   - Configure `tailwind.config.js`
   - Setup `internal/app/styles/globals.css`

4. **Implement Pages:**
   - `internal/app/app/page.tsx` - Chat interface
   - Navigation and layout components

5. **Implement State Management:**
   - `internal/app/store/store.ts` - Global store
   - `internal/app/hooks/useChat.ts` - Chat state

### Phase 3: Integration & Testing
1. **IPC Integration:**
   - Daemon-to-frontend communication
   - Message type definitions in `pkg/types/`

2. **Database Integration:**
   - Chat history persistence
   - Workspace configuration

3. **Testing:**
   - Unit tests for each package
   - Integration tests in `tests/integration/`
   - E2E tests in `tests/e2e/`

### Phase 4: Deployment
1. **Build System:**
   - Update `Makefile` for Unix builds
   - Update `build.ps1` for Windows builds

2. **Installers:**
   - Windows installer via `cmd/vectora-installer/`
   - Cross-platform binary distribution

3. **Auto-Update:**
   - Update checking mechanism
   - Versioning strategy

---

## File Count Summary

| Component | Files | Purpose |
|-----------|-------|---------|
| cmd/vectora | 4 | Main daemon |
| cmd/vectora-app | 3 | Desktop wrapper |
| cmd/vectora-installer | 5 | Setup wizard |
| internal/infra | 3 | Infrastructure |
| internal/ipc | 7 | IPC protocol |
| internal/db | 8 | Persistence |
| internal/core | 4 | Business logic |
| internal/llm | 6 | LLM integration |
| internal/engines | 3 | AI processing |
| internal/acp | 3 | Tool execution |
| internal/tools | 3 | Built-in tools |
| internal/git | 3 | Git integration |
| internal/app | 15+ | Frontend app |
| pkg/types | 3 | Type definitions |
| pkg/utils | 3 | Utilities |
| tests/* | 4 | Test suite |
| Root | 10 | Configuration |
| **TOTAL** | **90+** | **Complete scaffold** |

---

## Technology Stack

### Backend
- **Language:** Go 1.22+
- **Database:** bbolt (metadata), chromem-go (vectors)
- **IPC:** Unix sockets / Named pipes
- **Logging:** Structured JSON logging
- **HTTP:** Standard library / gRPC

### Frontend
- **Framework:** Next.js 15+
- **UI Library:** React 19+
- **Language:** TypeScript 5+
- **Styling:** Tailwind CSS 3+
- **State:** Zustand or Redux Toolkit
- **HTTP Client:** Axios or TanStack Query

### Desktop
- **Framework:** Wails (Go + Web)
- **Installer:** Fyne GUI
- **Package:** Windows EXE, Linux AppImage, macOS DMG

### Communication
- **Protocol:** JSON-ND (JSON Newline Delimited)
- **Transport:** Unix sockets / Windows Named pipes
- **Fallback:** TCP localhost

---

## Next Steps

1. **Verify Structure:** ✅ Complete
2. **Initialize Go Module:** `go mod init github.com/vectora/vectora`
3. **Setup npm/TypeScript:** `npm install && npm run build`
4. **Implement Core Packages:** Start with `internal/infra/`
5. **Write Tests:** Add unit tests as you develop
6. **Build Frontend:** Implement Next.js components
7. **Integration Testing:** Test IPC communication
8. **Deployment:** Create installers and distribution

---

**This structure follows Go best practices and Next.js conventions while supporting Vectora's distributed architecture with daemon + frontend model.**
