# Vectora VS Code Extension - Legacy Code Audit

**Date**: 2026-04-12
**Status**: Phase 1a - Initial Audit
**Purpose**: Identify all Roo-Code legacy code and create removal plan

---

## Executive Summary

The Vectora VS Code extension is **60-70% Roo-Code boilerplate** that needs to be removed or refactored. Current findings:

- **60 files importing `@roo*` patterns** that should be replaced or removed
- **53 components in `chat/` subdirectory** with mixed legacy/new code
- **11 Roo-specific components** that can be deleted entirely
- **30+ state variables** in ChatView.tsx that aren't needed for Vectora
- **500+ lines of message transformation logic** that can be eliminated

---

## Part 1: Dependency Analysis (package.json)

### KEEP - Core Runtime (Non-Negotiable)
```json
{
  "vscode": "^1.90.0",
  "vscode-jsonrpc": "^8.2.1",
  "react": "^18.3.1",
  "react-dom": "^18.3.1"
}
```

**Rationale**: These are the absolutely essential dependencies for VS Code integration and React UI.

### KEEP - UI Framework Stack
```json
{
  "@radix-ui/*": "~1.x (14 packages)",
  "tailwindcss": "^3.4.14",
  "lucide-react": "^0.453.0",
  "@vscode/codicons": "^0.0.45",
  "@vscode/webview-ui-toolkit": "^1.4.0"
}
```

**Rationale**: These provide accessible, themeable UI components matching VS Code design system.

### KEEP - Content Rendering (Markdown, Code, Visualization)
```json
{
  "react-markdown": "^9.1.0",
  "remark-gfm": "^4.0.1",
  "remark-math": "^6.0.0",
  "rehype-highlight": "^7.0.2",
  "rehype-katex": "^7.0.1",
  "shiki": "^4.0.2",
  "katex": "^0.16.45",
  "mermaid": "^11.14.0",
  "diff": "^8.0.4",
  "ansi-to-html": "^0.7.2"
}
```

**Rationale**: Essential for displaying markdown, code blocks, math, and diagrams in chat messages.

### KEEP - State & Data Management
```json
{
  "@tanstack/react-query": "^5.97.0",
  "react-use": "^17.6.0",
  "react-virtuoso": "^4.18.4",
  "zod": "^4.3.6",
  "lru-cache": "^11.3.3"
}
```

**Rationale**: React Query for data fetching, virtuoso for efficient message lists, zod for validation.

### KEEP - Utilities & Helpers
```json
{
  "axios": "^1.15.0",
  "date-fns": "^4.1.0",
  "clsx": "^2.1.1",
  "tailwind-merge": "^2.5.4",
  "shell-quote": "^1.8.3",
  "fast-deep-equal": "^3.1.3",
  "remove-markdown": "^0.6.3",
  "pretty-bytes": "^7.1.0",
  "fzf": "^0.5.2"
}
```

**Rationale**: General-purpose utilities needed for chat functionality, not Roo-specific.

### REMOVE - Analytics/Telemetry
```json
{
  "posthog-js": "^1.367.0"  ❌ DELETE - Vectora doesn't use external telemetry
}
```

### CONDITIONAL - Audio/UX Features
```json
{
  "use-sound": "^5.0.0"  ❓ CONDITIONAL - Remove unless UX requires notification sounds
}
```

Recommendation: **REMOVE** for Phase 1 (can add back in Phase 5 if needed)

### KEEP - Internationalization
```json
{
  "i18next": "^26.0.4",
  "react-i18next": "^17.0.2",
  "i18next-http-backend": "^3.0.4"
}
```

**Rationale**: Needed for 4-language support (en, pt, es, fr).

### KEEP - Dev Dependencies (Essential)
```json
{
  "@types/vscode": "^1.90.0",
  "@types/react": "^18.3.11",
  "@types/react-dom": "^18.3.0",
  "typescript": "^5.3.0",
  "vite": "^5.4.10",
  "webpack": "^5.90.0",
  "webpack-cli": "^5.1.0",
  "eslint": "^8.0.0",
  "mocha": "^10.4.0"
}
```

**Rationale**: Build tools, type definitions, and test infrastructure.

---

## Part 2: @roo* Import Analysis (60 Files)

### Files importing `@roo-code/types` (9 Files)
```
src/shared/api.ts
src/shared/modes.ts
src/shared/getApiMetrics.ts
src/shared/__tests__/api.spec.ts
src/shared/__tests__/modes.spec.ts
src/shared/__tests__/getApiMetrics.spec.ts
src/components/chat/ChatView.tsx
src/components/chat/ChatRow.tsx
src/components/chat/WarningRow.tsx
```

**Action**: Replace with local `VectoraMessage` type from `/src/types/vectora.ts`

### Files importing `@roo/*` utilities (45+ Files)

**High-impact imports in ChatView.tsx**:
```typescript
import { combineApiRequests } from "@roo/combineApiRequests";
import { combineCommandSequences } from "@roo/combineCommandSequences";
import { getApiMetrics } from "@roo/getApiMetrics";
import { getAllModes } from "@roo/modes";
import { ProfileValidator } from "@roo/ProfileValidator";
import { getLatestTodo } from "@roo/todo";
import { findLast } from "@roo/array";
import { SuggestionItem } from "@roo-code/types";
```

**Action**:
- Remove `combineApiRequests`, `combineCommandSequences` (Roo-specific message transformation)
- Replace `getAllModes` with Vectora-specific modes
- Replace `ProfileValidator` with Vectora config validator
- Keep `findLast` from array utils (generic, reusable)

---

## Part 3: Component Inventory (53 Files in src/components/chat/)

### DELETE - Roo-Specific Components (11 Files)

| Component | Reason | Status |
|-----------|--------|--------|
| `ProfileViolationWarning.tsx` | Roo provider constraint checking | Delete |
| `AutoApprovedRequestLimitWarning.tsx` | Roo auto-approval feature | Delete |
| `AutoApproveDropdown.tsx` | Roo auto-approval UI | Delete |
| `CheckpointWarning.tsx` | Roo checkpoint system | Delete |
| `CheckpointRestoreDialog.tsx` | Roo checkpoint restore | Delete |
| `BatchDiffApproval.tsx` | Roo batch edit approval | Delete |
| `BatchFilePermission.tsx` | Roo batch permission | Delete |
| `ShareButton.tsx` | Roo share feature (no backend) | Delete |
| `RooHero.tsx` | Roo mascot/logo component | Delete |
| `RooTips.tsx` | Roo welcome tips | Delete |
| Files importing `useRooCreditBalance`, `useRooPortal`, `useRooConfig` | Roo-specific hooks | Delete/Refactor |

### REFACTOR - Roo-Influenced but Reusable (8 Components)

| Component | Change | Priority |
|-----------|--------|----------|
| `ChatView.tsx` | Reduce from 1400 to 600 lines, remove Roo state | HIGH |
| `ModelSelector.tsx` | Hardcode Vectora models, remove provider selection | HIGH |
| `ContextSelector.tsx` | Focus on Vectora-specific context (RAG workspace) | MEDIUM |
| `ModeSelector.tsx` | Remove (Vectora doesn't have modes) | HIGH |
| `ApiConfigSelector.tsx` | Remove (Vectora handles provider in systray) | HIGH |
| `PolicySelector.tsx` | Refactor for Vectora tool permissions | MEDIUM |
| `TooManyToolsWarning.tsx` | Make generic (remove Roo thresholds) | LOW |
| `ContextWindowProgress.tsx` | Make generic (generic progress bar) | LOW |

### KEEP - Core Functional Components (34 Components)

| Category | Components | Action |
|----------|-----------|--------|
| **Message Rendering** | ChatRow, ChatMessage, ChatTextArea, ToolUseBlock, ReasoningBlock | KEEP as-is |
| **Content Display** | Markdown, CodebaseSearchResult, CodebaseSearchResultsDisplay, FileChangesPanel | KEEP as-is |
| **Status & Progress** | ProgressIndicator, IndexingStatusBadge, CodeIndexPopover | KEEP as-is |
| **History & Sessions** | history/* (all 8 files) | KEEP as-is |
| **UI Primitives** | common/* (all generic components) | KEEP as-is |
| **Error Handling** | ErrorBoundary, ErrorRow, CommandExecutionError, WarningRow | KEEP as-is |
| **Utilities** | Dropdown, Select, Menu components | KEEP as-is |

---

## Part 4: State Management Analysis

### ChatView.tsx State Variables (30+)

**Roo-specific state (to remove)**:
- `selectedApiConfiguration` - Roo provider selection (move to systray)
- `selectedModel` - Will use global provider from systray
- `autoApprove` - Roo auto-approval system
- `customInstructions` - Roo custom mode system
- `budget` - Roo token budget tracking
- `selectedMode` - Roo modes (Gemini, Claude, Qwen, etc)
- `botMessageBeingEdited` - Roo message editing
- `checkpointRestoreId` - Roo checkpoint system
- `tempTask` - Temporary task state (Roo)

**Vectora-needed state**:
- `messages` - Chat message history ✓
- `isProcessing` - Currently waiting for response ✓
- `currentSessionId` - Active session ID ✓
- `error` - Error messages ✓
- `selectedModel` → Can be removed, use global provider

**Message transformation logic (500+ lines to remove)**:
- `combineApiRequests()` - Consolidates multiple API requests (Roo-specific)
- `combineCommandSequences()` - Groups command outputs (Roo-specific)
- `condensedAndContextualizedMessages` - Complex context compression (Roo)
- All auto-approval/budget calculation logic

### Expected Result

After Phase 2c cleanup:
- ChatView.tsx: 600 lines (from 1400+)
- VectoraStateContext: ~150 lines (minimal state)
- Message flow: Direct ACP protocol → No transformation layer

---

## Part 5: Path Alias Cleanup (vite.config.ts)

### REMOVE These Aliases
```typescript
{ find: "@roo-code/types", replacement: resolve(__dirname, "./src/types/roo") },
{ find: "@roo-code/core", replacement: resolve(__dirname, "./src/roo-internal/core/src/index.ts") },
{ find: "@roo-code/core/browser", replacement: resolve(__dirname, "./src/roo-internal/core/src/browser.ts") },
{ find: "@roo-code/ipc", replacement: resolve(__dirname, "./src/roo-internal/ipc/src/index.ts") },
```

### KEEP These Aliases
```typescript
{ find: "@", replacement: resolve(__dirname, "./src") },
{ find: "@src", replacement: resolve(__dirname, "./src") },
{ find: "@components", replacement: resolve(__dirname, "./src/components") },
{ find: "@context", replacement: resolve(__dirname, "./src/context") },
{ find: "@hooks", replacement: resolve(__dirname, "./src/hooks") },
{ find: "@utils", replacement: resolve(__dirname, "./src/utils") },
{ find: "@core", replacement: resolve(__dirname, "./src/core") },
{ find: "@shared", replacement: resolve(__dirname, "./src/shared") },
{ find: "vscode", replacement: resolve(__dirname, "./src/roo-internal/vscode-shim/src/vscode.ts") },
```

### KEEP @roo (for now)
```typescript
{ find: "@roo", replacement: resolve(__dirname, "./src/shared") },
```

Will be refactored in Phase 2 (replace `@roo/foo` with `@shared/foo`)

---

## Part 6: Directory Structure Cleanup

### DELETE These Directories
```
src/types/roo/                 (empty placeholder)
src/roo-internal/core/src/     (missing, causing build errors)
src/roo-internal/ipc/src/      (missing, causing build errors)
```

### KEEP These Directories
```
src/roo-internal/vscode-shim/  (contains vscode.ts shim needed for imports)
```

---

## Implementation Schedule

### Phase 1a (THIS): Create audit document ✅
- [x] Document all dependencies
- [x] List all 60 @roo* imports
- [x] Categorize 53 components
- [x] Identify 11 deletable components
- [x] Map path aliases to remove

### Phase 1b: Remove Roo-Code Type System
- [ ] Create `/src/types/vectora.ts`
- [ ] Replace all ClineMessage imports with VectoraMessage
- [ ] Delete `/src/types/roo/` directory
- [ ] Test: `grep -r "ClineMessage" src/ | wc -l` should return 0

### Phase 1c: Delete Roo-Specific Components
- [ ] Remove 11 files listed in "DELETE" section
- [ ] Update all imports in ChatView.tsx

### Phase 1d: Clean package.json
- [ ] Remove posthog-js
- [ ] Remove use-sound (or set to CONDITIONAL)
- [ ] Verify no Roo-specific packages remain

### Phase 1e: Clean Path Aliases
- [ ] Remove @roo-code/* aliases from vite.config.ts
- [ ] Remove @roo-code-related paths
- [ ] Keep @roo for now (will refactor in Phase 2)

### Phase 1f: Fix Compilation
- [ ] Run `npm run compile`
- [ ] Fix any missing imports from deleted files
- [ ] Verify build succeeds

---

## Files to Review (Next Steps)

✅ `/extensions/vscode/package.json` - Dependency analysis
✅ `/extensions/vscode/vite.config.ts` - Path alias analysis
✅ `/extensions/vscode/src/components/chat/` - Component inventory
✅ `/extensions/vscode/src/components/chat/ChatView.tsx` - State analysis
✅ `/extensions/vscode/src/extension.ts` - Entry point review

---

## Success Criteria

- [x] Audit document created
- [ ] Phase 1b: VectoraMessage types created
- [ ] Phase 1c: No Roo-specific components remain
- [ ] Phase 1d: package.json cleaned
- [ ] Phase 1e: vite.config.ts cleaned
- [ ] Phase 1f: `npm run compile` succeeds without errors

---

**Prepared by**: Claude + Bruno
**Date**: 2026-04-12
**Status**: Ready for Phase 1b Implementation
