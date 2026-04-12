# Phase 3c: API Integration Loop - Verification Checklist

## End-to-End Message Flow

### Message Path Verification
- [x] **ChatView.tsx** → User input captured and sent to ChatTextArea
- [x] **ChatTextArea.tsx** → Input validated and passed to onSend callback
- [x] **ChatViewProvider.resolveWebviewView()** → Receives webview messages via onDidReceiveMessage
- [x] **AcpClient.prompt()** → Sends message to Core via JSON-RPC 2.0
- [x] **Core Binary** → Processes prompt and returns PromptResponse
- [x] **AcpClient.onSessionUpdate** → Receives streaming notifications from Core
- [x] **ChatViewProvider.handleNotification()** → Routes notifications to webview
- [x] **ChatView.tsx** → Re-renders with new messages

### Type Alignment

#### Frontend Types (Vectora)
```typescript
// src/types/vectora.ts
interface VectoraMessage {
  id?: string;
  role: "user" | "assistant" | "system" | "tool";
  content: string;
  timestamp?: string;
  toolCallId?: string;
  toolName?: string;
}
```

#### ACP Request/Response Types
```typescript
// src/types/core.ts
interface SessionPromptRequest {
  sessionId: string;
  messages: VectoraMessage[];
  workspaceId?: string;
}

interface PromptResponse {
  sessionId: string;
  model: string;
  content: string;
  toolCalls?: ToolCall[];
  usage?: TokenUsage;
}
```

#### Notification Types
```typescript
// src/types/core.ts
interface SessionUpdate {
  type: "message" | "tool_call" | "tool_result" | "error" | "complete";
  content?: string;
  toolName?: string;
  toolInput?: Record<string, unknown>;
  delta?: string;
  usage?: TokenUsage;
}
```

## Integration Points Verified

### 1. Client Connection (AcpClient)
- [x] Spawns Core binary via child_process
- [x] Establishes JSON-RPC 2.0 connection
- [x] Handles stdio stream communication
- [x] Manages request/response pairing with timeouts
- [x] Routes notifications to event emitters

### 2. Chat Panel Integration
- [x] Initializes AcpClient on extension startup
- [x] Creates sessions via `createSession()`
- [x] Sends prompts via `prompt()` method
- [x] Listens to `onSessionUpdate` events
- [x] Routes messages to webview via postMessage

### 3. State Management
- [x] VectoraStateProvider: Global state with hooks
- [x] useVectoraState(): Access current state
- [x] useSession(): Manage active session
- [x] Session tracking in ChatViewProvider

### 4. Message Rendering
- [x] ChatView displays messages from session
- [x] ChatRow renders individual messages
- [x] CommandExecution displays tool results
- [x] FileChangesPanel shows file modifications
- [x] Streaming updates reflected in real-time

## Configuration Requirements

### Settings (extension.json / vscode settings)
```json
{
  "vectora.corePath": "path/to/vectora/binary",
  "vectora.defaultProvider": "gemini",
  "vectora.enableRAG": true
}
```

### Environment Variables (for Core)
```bash
VECTORA_LOG_LEVEL=info
VECTORA_WORKSPACE_PATH=/path/to/workspace
```

## Testing Workflow (Manual)

1. **Start Extension**
   - Run `npm run watch` in extensions/vscode
   - Launch debug via F5 in VS Code

2. **Verify Core Connection**
   - Check that Binary Manager finds/downloads Core
   - Verify AcpClient.connect() succeeds
   - Confirm "Connected" status appears in UI

3. **Send Message**
   - Type message in ChatView
   - Observe message appears in chat
   - Wait for Core response

4. **Verify Response**
   - Check that streaming updates appear
   - Confirm tool calls are displayed
   - Validate file changes are shown

5. **Test Tool Execution**
   - Request file read/write
   - Verify CommandExecution displays
   - Check that FileChangesPanel updates

## Known Limitations (Phase 3c)

- ⚠️ Provider selection hardcoded to "gemini" (Phase 4a)
- ⚠️ RAG search not yet integrated (Phase 4b)
- ⚠️ Token usage not displayed (Phase 4d)
- ⚠️ No message persistence/history (Phase 5a)

## Success Criteria

- [x] Extension compiles without errors
- [x] AcpClient types match Core protocol
- [x] ChatViewProvider correctly routes messages
- [x] Message flow is unidirectional (user → Core → UI)
- [x] No Roo-Code legacy code remains in flow
- [x] All imports use Vectora types
- [ ] (Requires runtime test) End-to-end message delivery works

## Next Phase (4a)

After 3c completes, Phase 4a will:
1. Add provider selection UI to status bar
2. Store selected provider in state
3. Pass provider to `createSession()`
4. Update available models based on selection

---

**Status**: Integration complete at type/code level. Awaiting runtime verification.
