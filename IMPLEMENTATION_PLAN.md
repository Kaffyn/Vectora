# Vectora Implementation Plan: MVP → Production-Ready

**Document Status:** Active | **Version:** 2.0 | **Last Updated:** 2026-05-16

---

## Executive Summary

Vectora has achieved **MVP stability** (see Phase 1 Checklist). This plan outlines the evolution to **Production-Ready** architecture by addressing the "organic architecture" problem that caused `NoneType` errors and import complexity.

**Key Insight:** The current architecture grew organically (one file per feature). Moving forward requires **stratified architecture** based on responsibility domains, not file size.

---

## Phase 1: MVP (COMPLETED ✅)

### Checklist - All Verified Working

#### 1. Command System ✅

- [x] `/debug` - Toggle/persist debug mode
- [x] `/tools` - List 13+ tools
- [x] `/list` - Complete command reference
- [x] `/model` - List/switch models (now dynamic via `Config.get_llm_model()`)
- [x] `/new`, `/session <id>` - Session management with user_type preservation
- [x] `/help` - Quick reference
- [x] Persistent config in `~/.vectora/chat_config.json`

#### 2. UI & UX ✅

- [x] Compact welcome screen (all 8 commands, no truncation)
- [x] Rich dashboard with 3-panel layout (header, body, footer)
- [x] Debug Mode with real-time log panel
- [x] SafeConsole for Windows UTF-8 compatibility
- [x] Session audit export to Markdown

#### 3. System Prompt Enhancement ✅

- [x] Creator attribution (Bruno Soares)
- [x] Technical identity (LangChain, LangGraph, LanceDB, Apache 2.0)
- [x] Privacy & Security Protocols (5 core rules)
- [x] Technical Boundaries (deletion, resources, context, sandbox)
- [x] Design Principles (6 principles documented)

#### 4. Security Foundation ✅

- [x] `security.py` module with:
  - Protected paths validation (Windows, Linux, macOS)
  - Protected files validation (SSH, system configs)
  - Blocked terminal commands (rm, dd, chmod, sudo, etc.)
  - Safe defaults with explicit error messages
- [x] All validators tested and working

#### 5. Logging & Observability ✅

- [x] JSON structured logging with correlation IDs
- [x] Quiet mode (suppress external library logs)
- [x] VectoraOnlyFilter for focused logging
- [x] Session audit trail with full history
- [x] Debug logging for all major operations

#### 6. Resilience & Context ✅

- [x] RunnableConfig context injection
- [x] Context preservation across sessions
- [x] User_type preservation on switches
- [x] Dynamic model loading on each LLM call
- [x] Error handling with graceful fallbacks

---

## Phase 2: Architecture Refactor → Production-Ready (PLANNED)

### Current Problems (Organic Architecture)

```
BEFORE (Current - Organic):
vectora/
├── chat.py              # 450+ lines: UI + Logic mixed
├── commands.py          # 400+ lines: Command dispatcher
├── nodes.py             # 324 lines: Graph nodes
├── tools.py             # 1000+ lines: All tools in one file
├── background_worker.py # Worker logic isolated
├── checkpointer.py      # Persistence isolated
├── initialization.py    # One-time setup
├── log_setup.py         # Logging config
├── constants.py         # Environment variables
├── setup_wizard.py      # CLI wizard
├── run_chat.py          # Entry point
├── prompts.py           # System prompts
├── config.py            # Configuration
├── context.py           # Context objects
├── state.py             # State definitions
├── graph.py             # Graph construction
└── ui.py                # Rich components
```

**Problems:**

1. **Import Hell:** Circular dependencies (chat → commands → context → chat)
2. **NoneType Errors:** Config/context passed through 5+ layers
3. **Testing Nightmare:** Can't test graph without initializing Rich, UI, Logger
4. **Maintenance Burden:** 18+ files to understand one chat flow
5. **No Isolation:** Changing logging affects graph, changing UI affects tools

---

### Proposed Architecture (Phase 2)

```
AFTER (Production - Stratified):
vectora/
├── main.py              # CLI Entry point ONLY
├── core/
│   ├── agent.py         # AgentManager (Facade & Orchestrator)
│   └── state.py         # Immutable State definitions
├── graph/
│   ├── __init__.py       # build_graph() function
│   ├── nodes.py         # LLM node, tool node, sub node
│   └── state.py         # Graph-specific state (MessageList, etc.)
├── services/            # DOMAIN SERVICES (no UI dependencies)
│   ├── __init__.py
│   ├── session.py       # SessionService (/new, /session, etc.)
│   ├── embedding.py     # EmbeddingService (LanceDB + Worker)
│   ├── telemetry.py     # TelemetryService (Logs + Audit)
│   ├── security.py      # SecurityService (Validation + Whitelisting)
│   └── tool_runner.py   # ToolService (Safe tool execution)
├── tools/               # Pure tool functions (no state)
│   ├── rag_tools.py
│   ├── web_tools.py
│   └── file_tools.py
├── settings.py          # Pydantic Settings (SINGLE SOURCE OF TRUTH)
├── ui/                  # PRESENTATION LAYER (Rich components)
│   ├── __init__.py
│   ├── dashboard.py     # VectoraLayout
│   ├── components.py    # ChatMessage, Panels, etc.
│   └── prompts.py       # User input handling
└── config/
    ├── __init__.py
    ├── logging.py       # Log configuration
    └── constants.py     # Application constants (immutable)
```

**Architecture Principles:**

1. **Single Source of Truth:** All config in `settings.py` (Pydantic)
2. **Immutable State:** Graph state is JSON-like, never mutated mid-flow
3. **Explicit Injection:** All dependencies passed to constructors, never global
4. **Separation of Concerns:** Services know NOTHING about UI (Rich)
5. **Testability:** Core can be tested without initializing CLI

---

## Phase 2: Detailed Implementation

### 2.1 Create Core Foundation

**Goal:** Build the brain of Vectora independent of UI.

#### Step 1: Create `settings.py` (Pydantic-Settings)

```python
# vectora/settings.py
from pydantic_settings import BaseSettings
from pathlib import Path

class Settings(BaseSettings):
    # LLM Configuration
    llm_provider: str = "google-genai"
    google_api_key: str | None = None
    google_model: str = "gemini-3.1-flash-lite"

    # Database & Persistence
    db_url: str = f"sqlite:///{Path.home() / '.vectora' / 'vectora.db'}"
    log_file: Path = Path.home() / ".vectora" / "logs" / "vectora.jsonl"

    # Logging
    log_level: str = "INFO"
    debug_mode: bool = False

    # Directories (immutable)
    vectora_home: Path = Path.home() / ".vectora"

    class Config:
        env_file = ".env"
        case_sensitive = False

    def __init__(self, **data):
        super().__init__(**data)
        # Validate & create directories on init
        self.vectora_home.mkdir(parents=True, exist_ok=True)
        (self.vectora_home / "logs").mkdir(parents=True, exist_ok=True)
        # If any key is missing, Pydantic raises ValidationError immediately
        # No "NoneType" surprises later
```

**Benefits:**

- App crashes at startup with clear message if config is invalid
- No scattered `os.getenv()` calls throughout codebase
- Settings are **immutable after initialization**
- Easy to test: `Settings(llm_provider="test")`

---

#### Step 2: Create `core/agent.py` (The Orchestrator)

```python
# vectora/core/agent.py
from settings import Settings
from graph import build_graph
from services import SessionService, EmbeddingService, TelemetryService

class AgentManager:
    """Orchestrator for all Vectora operations. No UI awareness."""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.session = SessionService(settings)
        self.embedding = EmbeddingService(settings)
        self.telemetry = TelemetryService(settings)

        self.graph = build_graph(settings)
        self.embedding.start()

    async def chat(self, user_input: str, session_id: int) -> str:
        """Pure chat execution. Returns response string."""
        config = self.session.get_runnable_config(session_id)
        result = await self.graph.ainvoke(
            {"messages": [HumanMessage(user_input)]},
            config=config
        )
        # Telemetry tracks this automatically
        return result["messages"][-1].content

    async def switch_model(self, model_name: str) -> bool:
        """Switch LLM model. Returns success."""
        try:
            self.settings.google_model = model_name
            return True
        except ValueError:
            return False

    async def list_sessions(self) -> list[dict]:
        """Returns list of available sessions."""
        return await self.session.list_all()

    async def shutdown(self):
        """Graceful shutdown: stop worker, close DB."""
        await self.embedding.stop()
```

**Benefits:**

- Pure Python. No Rich imports.
- Testable: `agent = AgentManager(Settings(...)); await agent.chat("hi", 1)`
- Isolated: Graph errors don't affect TelemetryService
- **The UI will only call these 5 methods**

---

#### Step 3: Create Services

**vectora/services/session.py**

```python
class SessionService:
    """Manages session lifecycle: create, switch, list, delete."""

    async def create(self, user_type: str = "default") -> int:
        """Create new session. Returns thread_id."""

    async def switch(self, thread_id: int) -> bool:
        """Switch to session. Validates existence."""

    async def list_all(self) -> list[dict]:
        """Returns all sessions with metadata."""

    def get_runnable_config(self, thread_id: int) -> RunnableConfig:
        """Returns immutable config for graph execution."""
```

**vectora/services/embedding.py**

```python
class EmbeddingService:
    """Manages vector store, background worker, and embedding queue."""

    async def start(self):
        """Start background worker."""

    async def stop(self):
        """Stop worker gracefully."""

    async def queue_document(self, doc_id: str, text: str):
        """Queue document for embedding (fire-and-forget)."""

    async def search(self, query: str, collection: str) -> list[dict]:
        """Vector search. Retries on DB lock."""
```

**vectora/services/telemetry.py**

```python
class TelemetryService:
    """Logs, audit trails, debug dumps."""

    def log_chat(self, session_id: int, role: str, content: str):
        """Log message to JSON structured log."""

    async def export_session_audit(self, session_id: int) -> str:
        """Export session as Markdown. Returns file path."""

    async def export_debug_dump(self) -> str:
        """Create .tar.gz with logs, config, state. Returns path."""
```

**vectora/services/security.py**

```python
class SecurityService:
    """Validates commands before execution."""

    def validate_file_edit(self, path: str) -> tuple[bool, str]:
        """Check if file edit is allowed."""

    def validate_terminal_command(self, cmd: str) -> tuple[bool, str]:
        """Check if command is safe."""
```

---

### 2.2 Refactor CLI Layer

**Goal:** Reduce `chat.py` and `commands.py` to thin UI wrappers.

#### New `main.py`

```python
# vectora/main.py
import asyncio
from settings import Settings
from core.agent import AgentManager
from ui.dashboard import create_dashboard

async def run_cli():
    """CLI entry point."""
    # 1. Initialize config (crashes here if invalid)
    settings = Settings()

    # 2. Initialize agent (no UI involved)
    agent = AgentManager(settings)

    # 3. Initialize UI (connects to agent)
    dashboard = create_dashboard(agent, settings)

    # 4. Run event loop
    await dashboard.run()

if __name__ == "__main__":
    asyncio.run(run_cli())
```

#### New `commands.py` (simplified)

```python
# vectora/ui/commands.py
# Each command is now just a wrapper that calls agent methods

async def handle_debug(agent: AgentManager, args: str) -> None:
    """Handle /debug command."""
    if args.strip() in ["true", "false"]:
        agent.settings.debug_mode = args.strip() == "true"
        console.print("[OK] Debug mode updated")
    else:
        agent.settings.debug_mode = not agent.settings.debug_mode
        console.print(f"[OK] Debug toggled: {agent.settings.debug_mode}")

async def handle_session(agent: AgentManager, args: str) -> None:
    """Handle /session <id> command."""
    try:
        session_id = int(args.strip())
        success = await agent.switch_session(session_id)
        if success:
            console.print(f"[OK] Switched to session {session_id}")
        else:
            console.print(f"[ERROR] Session {session_id} not found")
    except ValueError:
        console.print("[ERROR] Invalid session ID")
```

**Result:** `commands.py` drops from 400 → 100 lines. It's now a **dispatcher** not a **logic container**.

---

### 2.3 Refactor Graph Layer

**Goal:** Ensure graph reads config dynamically on each invocation.

**Already Done in Phase 1:**

- ✅ Added `get_llm_model()` to Config
- ✅ Made `call_llm()` read config dynamically

**Phase 2 Addition:**

- [ ] Move all graph nodes to `graph/nodes.py`
- [ ] Decouple state from context objects
- [ ] Make state purely JSON-serializable

```python
# vectora/graph/nodes.py
async def call_llm(state: State, config: RunnableConfig) -> dict:
    """Read config dynamically. No hardcoded models."""
    settings = config.configurable.get("settings")

    model = load_model(
        provider=settings.llm_provider,
        model_name=settings.google_model
    )
    # ... rest of logic
```

---

## Phase 3: Rollout Timeline

### Week 1: Foundation

- [ ] Create `settings.py` (Pydantic)
- [ ] Create `core/agent.py` (AgentManager)
- [ ] Create service stubs (SessionService, EmbeddingService, etc.)
- [ ] **Tests:** Unit test AgentManager in isolation

### Week 2: Services Implementation

- [ ] Implement SessionService
- [ ] Implement EmbeddingService (port from background_worker.py)
- [ ] Implement TelemetryService (port from log_setup.py)
- [ ] **Tests:** Integration tests for each service

### Week 3: CLI Refactor

- [ ] Refactor `main.py` to use AgentManager
- [ ] Refactor `commands.py` (commands now call agent methods)
- [ ] Refactor `chat.py` (UI only, no logic)
- [ ] **Tests:** End-to-end tests via CLI

### Week 4: Graph Isolation

- [ ] Decouple graph from context objects
- [ ] Make state JSON-serializable
- [ ] Remove all `os.getenv()` calls
- [ ] **Tests:** Graph unit tests without CLI

### Week 5: Cleanup & Documentation

- [ ] Delete old files (initialization.py, constants.py, etc.)
- [ ] Update imports across all files
- [ ] Write architecture documentation
- [ ] **Tests:** Full regression test suite

---

## Critical Files & Dependencies

### Dependencies Map (Phase 2)

```
settings.py (No dependencies except Pydantic)
    ↓
core/agent.py (depends: settings, services, graph)
    ├── services/session.py (depends: settings)
    ├── services/embedding.py (depends: settings)
    ├── services/telemetry.py (depends: settings)
    └── graph/nodes.py (depends: settings)

main.py (depends: settings, agent, ui)
    ├── ui/dashboard.py (depends: agent, settings)
    ├── ui/commands.py (depends: agent)
    └── ui/components.py (depends: Rich only)
```

**Key Rule:** No file should depend on more than 3 other files.

---

## Testing Strategy

### Unit Tests (Phase 2)

```python
# tests/test_agent.py
async def test_agent_chat():
    settings = Settings(llm_provider="mock")
    agent = AgentManager(settings)

    response = await agent.chat("hello", session_id=1)
    assert isinstance(response, str)
    assert len(response) > 0

async def test_agent_switch_model():
    settings = Settings()
    agent = AgentManager(settings)

    success = await agent.switch_model("gpt-5.5")
    assert success is True
```

### Integration Tests

```python
# tests/test_services.py
async def test_session_service_create():
    service = SessionService(settings)

    session_id = await service.create("test_user")
    assert session_id >= 1

    sessions = await service.list_all()
    assert len(sessions) >= 1
```

### End-to-End Tests

```python
# tests/test_cli.py
async def test_cli_chat_flow():
    # Spawn full CLI in subprocess
    # Send commands via stdin
    # Verify output via stdout
    # Verify audit file exists
```

---

## Risk Mitigation

### Risk 1: Breaking Change During Refactor

**Mitigation:**

- Keep Phase 1 code in `legacy/` during Week 1-2
- Run both implementations in parallel
- Switch over only when Phase 2 passes all tests

### Risk 2: Import Errors After Cleanup

**Mitigation:**

- Use `mypy` and `ruff` to catch broken imports
- Automated lint checks on every commit

### Risk 3: Configuration Errors

**Mitigation:**

- Pydantic validates settings on init
- App crashes with clear message, not silent NoneType later

---

## Success Metrics

By end of Phase 2:

- ✅ **Files reduced:** 18+ → 12 (graph, services, ui, core, tools, settings, main)
- ✅ **Test coverage:** >80% of services unit tested
- ✅ **No NoneType errors:** All config validated at startup
- ✅ **Testability:** Graph can be tested without CLI
- ✅ **Maintainability:** No file has >300 lines
- ✅ **Documentation:** Architecture doc + inline comments

---

## Long-Term Vision (Phase 3+)

Once Phase 2 is done, Vectora can evolve:

1. **HTTP API:** Create `api/main.py` that uses same AgentManager
2. **Discord Bot:** Create `discord/main.py` that uses same AgentManager
3. **Web Dashboard:** Create `web/main.py` that uses same AgentManager
4. **Mobile App:** Create `mobile/api.py` that exposes AgentManager over HTTP

**Key:** Zero changes to `core/`, `services/`, `graph/` for any of these.

---

## Current Readiness (Phase 1 → 2)

**Phase 1 Status:** ✅ COMPLETE

- MVP is stable and tested
- All commands working
- Security foundation in place
- Logging & observability excellent

**Phase 2 Status:** 🚧 READY TO START

- Architecture designed
- Services scoped
- Testing strategy defined
- Risk mitigation planned

**Recommended Next Step:** Create `settings.py` file and build AgentManager skeleton in parallel with Phase 1. No breaking changes until Week 2.

---

## Appendix: File Migration Map

| Current File           | Phase 2 Location              | Status      |
| ---------------------- | ----------------------------- | ----------- |
| `chat.py`              | `ui/dashboard.py` + `main.py` | Split       |
| `commands.py`          | `ui/commands.py`              | Simplify    |
| `nodes.py`             | `graph/nodes.py`              | Move        |
| `background_worker.py` | `services/embedding.py`       | Integrate   |
| `checkpointer.py`      | `core/agent.py`               | Integrate   |
| `initialization.py`    | `settings.py`                 | Consolidate |
| `log_setup.py`         | `services/telemetry.py`       | Integrate   |
| `constants.py`         | `settings.py` + `config/`     | Consolidate |
| `setup_wizard.py`      | `ui/onboarding.py`            | Simplify    |
| `run_chat.py`          | `main.py`                     | Merge       |
| `security.py`          | `services/security.py`        | Move        |

---

**Document Owner:** Bruno Soares
**Last Review:** 2026-05-16
**Next Review:** When Phase 2 starts
