# VS Code Extension Refactoring: Complete Stabilization & Vectora Integration

## Context & Objectives

The Vectora VS Code extension was adapted from Roo-Code (legacy multi-provider AI agent framework) but integration is incomplete. Current state:

- **60-70% boilerplate from Roo-Code** that Vectora doesn't use
- **Incompatible message types** (expects `ClineMessage`, receives `ACP SessionUpdate`)
- **Over-engineered state** (10+ modes, multi-provider selection, auto-approval—Vectora doesn't need)
- **Structural build failures** (missing files due to incomplete adaptation)
- **No actual Vectora integration** (UI exists but doesn't connect to core)

### Goals

1. **Phase 1**: Remove all Roo-Code legacy code and cruft
2. **Phase 2**: Rebuild with stable, minimal foundations that actually connect to Vectora Core
3. **Phase 3**: Refactor UI components to use Vectora's actual ACP/IPC protocols
4. **Phase 4**: Add Vectora-specific features (embedding visualization, RAG result presentation, provider selection)
5. **Phase 5+**: Advanced features (performance, advanced RAG UI, streaming enhancements)

---

## Phase 1: Cleanup & Legacy Removal (1a-1f)

### 1a: Audit & Document Dependencies

**Goal**: Create definitive list of what must stay vs. what must go

**Actions**:

1. Analyze all `package.json` dependencies (KEEP, REMOVE, CONDITIONAL)
2. Scan all `/src` imports for `@roo*` patterns (60+ files)
3. Document component inventory (53 chat/ components)
   - Active (used by ChatView)
   - Unused/Roo-specific
   - Reusable as-is

**Deliverable**: `/extensions/vscode/LEGACY_AUDIT.md` (reference)

**Files to review**:

- `/extensions/vscode/package.json`
- `/extensions/vscode/vite.config.ts` (path aliases)
- `/extensions/vscode/src/components/chat/*.tsx`
- `/extensions/vscode/src/extension.ts`

---

### 1b: Remove Roo-Code Type System

**Goal**: Eliminate `@roo-code/types` and replace with Vectora ACP types

**Delete**:

- `/src/types/roo/` (directory)
- All `ClineMessage`, `ClineAsk`, `ClineSay*`, `isRetiredProvider` imports

**Create `/src/types/vectora.ts`**:

```typescript
export interface VectoraMessage {
  role: "user" | "assistant" | "system" | "tool";
  content: string;
  toolCallId?: string;
  toolName?: string;
}

export interface SessionUpdate {
  type: "message" | "tool_call" | "tool_result" | "error" | "complete";
  content?: string;
  toolName?: string;
  toolInput?: Record<string, unknown>;
}
```

**Refactor**:

- `/src/components/chat/ChatView.tsx` (largest Cline importer)
- `/src/components/chat/ChatRow.tsx`
- `/src/shared/*.ts` (check for Cline types)

---

### 1c: Remove Roo-Specific Components

**Goal**: Delete UI components only in Roo-Code

**Delete these files**:

```
src/components/chat/
  ├── ProfileViolationWarning.tsx
  ├── AutoApprovedRequestLimitWarning.tsx
  ├── AutoApproveDropdown.tsx
  ├── CheckpointWarning.tsx
  ├── CheckpointRestoreDialog.tsx
  ├── BatchDiffApproval.tsx
  ├── BatchFilePermission.tsx
  ├── ShareButton.tsx
```

**Update imports** in ChatView.tsx and related files

---

### 1d: Remove Unnecessary Dependencies

**Goal**: Trim `package.json` of Roo-Code cruft

**Remove**:

- `posthog-js` (Vectora doesn't use external telemetry)
- Other REMOVE items from audit

**Keep essential stack**:

- React, React Query, React Virtuoso
- Radix UI, Tailwind, Lucide
- Markdown rendering (remark, rehype, shiki, katex, mermaid)
- VS Code integration (vscode-jsonrpc CRITICAL)
- Utilities (zod, axios, date-fns, shell-quote)

**Update `vite.config.ts`**:

- Remove `@roo-code/types`, `@roo-code/core`, `@roo-code/ipc` aliases (if unused)
- Keep `@roo` → `/src/shared` (refactor in Phase 2)

---

### 1e: Remove @roo-code Internal Stubs

**Goal**: Clean up `/src/roo-internal/` placeholder directories

**Action**:

1. Remove unused path aliases in `vite.config.ts`:
   - `@roo-code/core`, `@roo-code/core/browser`, `@roo-code/ipc`
2. Keep `vscode` → `/src/roo-internal/vscode-shim/src/vscode.ts`
3. Delete empty/incomplete stub directories

---

### 1f: Fix Compilation Errors from Cleanup

**Goal**: Make extension compile after removing legacy code

**Actions**:

1. Run `npm run compile` and capture errors
2. For each missing import/component:
   - Trace to deleted files
   - Remove import OR replace with Vectora equivalent
3. Create stub implementations if needed

**Tests**:

- `npm run compile` succeeds
- `npm run build:webview` produces `dist/webview/`
- `webpack --mode production` produces `dist/extension.js`

---

## Phase 2: Reconstruction & Stable Normalization (2a-2d)

### 2a: Create Vectora Core Interface Types

**Goal**: Define clean types matching Vectora's actual ACP/IPC protocols

**Create `/src/types/core.ts`**:

```typescript
export interface SessionNewRequest {
  workspace_id: string;
  provider: string; // "gemini", "claude", "openai", etc
}

export interface SessionNewResponse {
  session_id: string;
  metadata: Record<string, unknown>;
}

export interface SessionPromptRequest {
  session_id: string;
  messages: VectoraMessage[];
  workspace_id?: string;
}

export interface PromptResponse {
  session_id: string;
  model: string;
  content: string;
  toolCalls?: ToolCall[];
  usage?: TokenUsage;
}

export interface VectoraMessage {
  role: "user" | "assistant" | "system" | "tool";
  content: string;
  name?: string;
  toolCallId?: string;
  toolName?: string;
}

export interface TokenUsage {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
}
```

**Create `/src/types/ipc.ts`** (IPC-specific):

```typescript
export interface WorkspaceQueryRequest {
  query: string;
  conversation_id: string;
}

export interface WorkspaceQueryResponse {
  answer: string;
  sources: SearchResult[];
}

export interface SearchResult {
  filename: string;
  content: string;
  relevance: number;
}
```

**Reference**:

- `/core/api/acp/types.go`
- `/core/api/ipc/protocol.go`
- `API_ARCHITECTURE.md`

---

### 2b: Refactor Client & IPC Communication

**Goal**: Update `/src/client.ts` to use new Vectora types

**Current client** (good):

- Uses `vscode-jsonrpc` correctly
- Spawns Core binary correctly
- Implements notification listening

**Updates**:

1. Type all RPC methods with Vectora types
2. Replace `Cline*` type references with `Vectora*`
3. Document ACP message flow

**Files**:

- `/src/client.ts` - Update types
- `/src/chat-panel.ts` - Update message handling
- `/src/extension.ts` - Update RPC call sites

---

### 2c: Create Minimal Vectora-Specific State Context

**Goal**: Replace Roo's 500+ lines of state with Vectora's minimal needs

**Create `/src/context/VectoraStateContext.tsx`**:

```typescript
export interface VectoraState {
  coreStatus: "starting" | "running" | "stopped" | "error";
  isConnected: boolean;
  selectedProvider: string | null;
  workspace: string | null;
  currentSessionId: string | null;
  isProcessing: boolean;
  availableModels: string[];
  selectedModel: string;
  error: string | null;
}
```

**Remove**:

- `/src/hooks/useRooConfig.ts`
- `/src/hooks/useCloudUpsell.ts`

**Create new hooks** in `/src/hooks/`:

- `useVectoraState.ts` - Access context
- `useCoreConnection.ts` - Manage lifecycle
- `useSession.ts` - Session management

---

### 2d: Establish Stable Message Flow Architecture

**Goal**: Document & implement single message flow

**Create `/src/ARCHITECTURE.md`** in extension:

```markdown
# Extension Architecture

## Message Flow

1. User types in ChatTextArea
2. Webview: vscode.postMessage({ type: 'prompt', text })
3. chat-panel.ts receives message
4. Calls: client.request('session/prompt', { messages })
5. Core processes, returns PromptResponse
6. Extension listens to notifications:
   - session/update (streaming chunks)
   - session/stream_delta (token-by-token)
   - session/complete (done)
7. Webview receives postMessage updates and re-renders

## Files & Responsibilities

- **extension.ts**: Entry point
- **client.ts**: JSON-RPC ACP client
- **chat-panel.ts**: WebviewProvider bridge
- **ChatView.tsx**: Main UI
- **VectoraStateContext.tsx**: Minimal state
```

---

## Phase 3: UI Component Refactoring & API Closure (3a-3c)

### 3a: Simplify ChatView.tsx

**Goal**: Reduce from 1400+ lines of Roo logic to clean Vectora logic

**Current problems**:

- 30+ state variables (modes, auto-approval, budget)
- 500+ lines message transformation (combineApiRequests)
- Heavy Roo type dependencies
- Complex message interleaving

**Refactored ChatView**:

1. Accept `VectoraMessage[]` from context
2. Render via ChatRow for each message
3. Handle streaming via notifications
4. Simple task state: `{ sessionId, isProcessing, currentTask }`
5. No mode selection, no budget tracking, no auto-approval

**Changes**:

- Remove 400+ lines Roo state logic
- Remove `combineApiRequests`, `combineCommandSequences`
- Replace with direct ACP message handling
- Keep message rendering (ChatRow, ToolUseBlock, FileChangesPanel)

---

### 3b: Refactor Core Integration Components

**Goal**: Update components to use clean ACP

**Refactor**:

1. **ModelSelector.tsx** → hardcode Vectora models
2. **ContextSelector.tsx** → Workspace/RAG context
   - Add "RAG Enabled" toggle
   - Show indexing status
3. **CommandExecution.tsx** → Show tool results with Vectora format
4. **FileChangesPanel.tsx** → Keep as-is

**Remove components**:

- API budget display
- Mode selection UI
- Provider selection (systray handles this)
- Auto-approval UI

---

### 3c: Close the API Integration Loop

**Goal**: Verify end-to-end message flow works

**Test flow**:

1. Launch extension → Core connects
2. Select provider from systray
3. Type message → sent to Core
4. Core responds with streaming updates
5. UI updates in real-time
6. Tool calls displayed and executed

**Checklist**:

- [ ] `/src/client.ts` uses Vectora ACP types
- [ ] `/src/chat-panel.ts` receives/sends proper messages
- [ ] `/src/types/core.ts` matches Core's actual interfaces
- [ ] ChatView renders without Roo logic
- [ ] Tool execution displays work
- [ ] File changes display works
- [ ] No Roo type imports in chat components

---

## Phase 4: UX Enhancement & Vectora-Specific Features (4a-4d)

### 4a: Implement Provider Selection UI

**Goal**: Add UI to switch providers (Gemini, Claude, OpenAI, etc)

**Note**: Systray already expanded to 8 providers (commit 4ddf38a)

**Extension integration**:

1. Listen for provider changes via IPC
2. Update `VectoraStateContext.selectedProvider`
3. Pass provider to `session/prompt` requests

**Components**:

- Add badge in status bar: "🤖 Gemini"
- Click to show provider menu

---

### 4b: Add RAG/Embedding Search Display Component

**Goal**: Show search results from workspace.query (RAG feature)

**New component `/src/components/chat/RagSearchResults.tsx`**:

- Display matching files for user query
- Show file names with relevance scores
- Show snippet previews

**Show in ChatView**:

- After user message
- Call `workspace.query` IPC method
- Display in special card

---

### 4c: Add Embedding Status & Indexing Progress

**Goal**: Show RAG index status

**Listen for IPC notifications**:

- `workspace.embed.progress` → Show progress bar
- `workspace.embed.complete` → Show ✅ badge

**New component `/src/components/chat/EmbeddingStatus.tsx`**:

- Show: "Indexing 150 files... 45% complete"
- Show: "✅ 2,500 vectors indexed"

---

### 4d: Improve Chat UX for Vectora Features

**Goal**: Add Vectora-specific UX improvements

**Changes**:

1. **Welcome screen**: Replace Roo hero with Vectora logo
   - Show: "Connected to Vectora Core"
   - Show: Available providers
   - Show: Current workspace

2. **Command palette**: Add Vectora commands
   - `Vectora: Select Provider`
   - `Vectora: Index Workspace`
   - `Vectora: View Embeddings`

3. **Message rendering**: Add Vectora formatting
   - Show source files in RAG results
   - Show tool outputs
   - Show embedding tokens

4. **Context window indicator**: Show token usage
   - "1,250 / 200,000 tokens used"

---

## Phase 5: Advanced Features & Polish (Optional)

### 5a: Message Caching & History Optimization

- Cache chat sessions locally
- Persist to `~/.Vectora/extension-cache.json`
- Load on startup

### 5b: Streaming & Real-Time Updates

- Token-by-token streaming display
- Show thinking tokens separately
- Better tool call display during execution

### 5c: Advanced RAG Features

- Highlight search results in code
- Show relevance scores
- Allow selecting specific files for context
- Show embedding vector distance metrics

### 5d: Performance & Bundle Size

- Tree-shake unused dependencies
- Lazy load components
- Code-split webview assets
- Optimize icon bundles

---

## Summary Timeline

| Phase | Duration | Goal                       | Deliverable                     |
| ----- | -------- | -------------------------- | ------------------------------- |
| 1     | 2-3 days | Remove 60% legacy code     | Compiling, Roo-free codebase    |
| 2     | 2-3 days | Rebuild with Vectora types | Clean message flow, working IPC |
| 3     | 2-3 days | Refactor UI components     | End-to-end functional chat      |
| 4     | 3-4 days | UX enhancements            | Polished, Vectora-specific UI   |
| 5     | 2+ weeks | Optional improvements      | Performance, advanced features  |

---

## File Manifest

**To Delete**:

- `/src/types/roo/` (directory)
- 11+ Roo-specific component files (ProfileViolationWarning, AutoApproved*, Checkpoint*, BatchDiff*, Share*)

**To Create**:

- `/src/types/core.ts` (ACP types)
- `/src/types/ipc.ts` (IPC types)
- `/src/context/VectoraStateContext.tsx` (new state)
- `/src/components/chat/RagSearchResults.tsx` (RAG display)
- `/src/components/chat/EmbeddingStatus.tsx` (indexing)
- `/src/ARCHITECTURE.md` (documentation)

**To Refactor** (major):

- `/src/components/chat/ChatView.tsx` (1400 → 600 lines)
- `/src/client.ts` (update types)
- `/src/chat-panel.ts` (update message handling)
- `/src/extension.ts` (update RPC calls)
- `/vite.config.ts` (remove @roo-code aliases)
- `/package.json` (remove posthog-js, etc)

**To Keep As-Is**:

- All message rendering components
- All content components (Markdown, code, etc)
- All history components
- All UI primitives
- Core IPC implementation

---

## Verification Steps

### Phase 1 Verification

```bash
npm run compile           # No errors
npm run build:webview    # dist/webview/ created
webpack --mode prod      # dist/extension.js created
```

### Phase 2 Verification

```bash
npm run compile           # Still works
grep -r "@roo-code/types" src/  # No results
grep -r "ClineMessage" src/     # No results
```

### Phase 3 Verification

```bash
# Launch extension in VS Code
# - Core connects
# - Send message to Core
# - Verify response appears in ChatView
# - Verify tool calls display correctly
```

### Phase 4 Verification

```bash
# - Provider selector shows in status bar
# - RAG search results display
# - Embedding progress visible
```

---

**Status**: Plan Ready for Review and User Approval
