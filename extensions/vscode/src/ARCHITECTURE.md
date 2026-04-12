# Vectora VS Code Extension - Architecture

## Overview

O Vectora VS Code Extension é uma camada de apresentação para o Vectora Core, um servidor AI agêntico que executa localmente ou remotamente. A extensão comunica com o Core usando JSON-RPC 2.0 sobre stdio e integrando com o VS Code APIs para edição e exploração de arquivos.

## Message Flow

### 1. User Input → WebView
```
User types in ChatTextArea
    ↓
ChatTextArea onChange event
    ↓
vscode.postMessage({ type: 'prompt', text: 'user input' })
```

### 2. WebView → Extension Process
```
WebviewPanel.webview.onDidReceiveMessage()
    ↓
chat-panel.ts receives message
    ↓
Call client.prompt({ sessionId, messages[] })
```

### 3. Extension → Core (JSON-RPC Request)
```
AcpClient.prompt(SessionPromptRequest)
    ↓
JSON-RPC Request: { jsonrpc: "2.0", id: 1, method: "session/prompt", params: {...} }
    ↓
Core stdin/stdout
```

### 4. Core Processing
```
Vectora Core receives request
    ↓
Parse SessionPromptRequest
    ↓
Select provider (Gemini, Claude, OpenAI, etc)
    ↓
Send to selected LLM
```

### 5. Core → Extension (Streaming Notifications)
```
Core sends notifications during processing:
    ↓
session/stream_delta: { type: "content", delta: "token" }
session/stream_delta: { type: "content", delta: "another" }
    ↓
session/complete: { type: "complete", usage: {...} }
```

### 6. Extension → WebView (Updates)
```
AcpClient receives notifications
    ↓
onSessionUpdate event fired
    ↓
chat-panel.ts receives update
    ↓
vscode.webview.postMessage({ type: 'session_update', data: {...} })
```

### 7. WebView → UI Re-render
```
ChatView component receives message
    ↓
useVectoraState() updates state
    ↓
Component re-renders with new content
    ↓
User sees streamed response in real-time
```

## File Organization

### Core Infrastructure
- **extension.ts**: Entry point, registers commands, manages lifecycle
- **client.ts** (AcpClient): JSON-RPC client, handles all communication with Core
- **chat-panel.ts**: WebviewPanel provider, bridges extension and webview

### Context & State
- **context/VectoraStateContext.tsx**: Global state management
  - useVectoraState(): read state
  - useCoreConnection(): manage connection
  - useSession(): manage active session
  - useProviderSelection(): provider/model selection
  - useEmbedding(): RAG/embedding status

### Types
- **types/core.ts**: ACP protocol types
  - SessionNewRequest/Response
  - SessionPromptRequest/PromptResponse
  - SessionUpdate (notifications)
  - ToolCall, TokenUsage
- **types/ipc.ts**: IPC/MCP protocol types
  - WorkspaceQuery/Response (RAG search)
  - EmbeddingStatus, IndexingProgress
  - FileContent, FileChange operations
- **types/vectora.ts**: Vectora-specific types
  - VectoraMessage, ExtensionState
  - ChatMessage, ChatSession
  - Provider, ProviderSelection
- **types/index.ts**: Central export point

### UI Components
- **components/chat/ChatView.tsx**: Main chat interface
- **components/chat/ChatRow.tsx**: Individual message rows
- **components/chat/ChatTextArea.tsx**: User input textarea
- **components/chat/CommandExecution.tsx**: Tool execution display
- **components/common/CodeBlock.tsx**: Code syntax highlighting
- **components/common/DiffView.tsx**: File diff display

### Utils
- **utils/format.ts**: Number and date formatting
- **utils/clipboard.ts**: Clipboard operations
- **utils/command-parser.ts**: Command parsing and execution
- **utils/highlight.ts**: Text highlighting for search
- **utils/highlighter.ts**: Code syntax highlighting
- **utils/markdown.ts**: Markdown utilities
- **utils/model-utils.ts**: Model/token calculations

## Request/Response Patterns

### Session Creation
```typescript
// Request
{
  jsonrpc: "2.0",
  id: 1,
  method: "session/new",
  params: {
    workspaceId: "workspace-1",
    provider: "gemini",
    model: "gemini-pro"
  }
}

// Response
{
  jsonrpc: "2.0",
  id: 1,
  result: {
    sessionId: "session-abc123",
    metadata: {}
  }
}
```

### Prompt Request with Streaming
```typescript
// Request
{
  jsonrpc: "2.0",
  id: 2,
  method: "session/prompt",
  params: {
    sessionId: "session-abc123",
    messages: [
      { role: "user", content: "What is 2+2?" }
    ]
  }
}

// Streaming Notifications
{ method: "session/stream_delta", params: { type: "content", delta: "4" } }
{ method: "session/stream_delta", params: { type: "usage", usage: {...} } }
{ method: "session/update", params: { type: "complete" } }
```

## Error Handling

### Connection Errors
```
Core crashes/exits
    ↓
Process.on('exit') triggered
    ↓
onConnectionChange.fire(false)
    ↓
UI shows "Connection Lost" status
    ↓
User can reconnect
```

### Request Timeout
```
AcpClient.request() with 60s timeout
    ↓
No response after 60s
    ↓
Promise.race() times out
    ↓
Throw: "Request 'session/prompt' timed out"
    ↓
Error handler shows to user
```

### JSON-RPC Errors
```
Core returns error response
    ↓
{ error: { code: -32600, message: "Invalid Request" } }
    ↓
AcpClient converts to Error
    ↓
Chat panel handles gracefully
    ↓
User sees error message in chat
```

## State Management Flow

```
VectoraStateProvider (Context)
    ├── coreStatus: "starting" | "running" | "stopped"
    ├── isConnected: boolean
    ├── selectedProvider: string
    ├── selectedModel: string
    ├── currentSessionId: string
    ├── isProcessing: boolean
    ├── availableProviders: Provider[]
    ├── availableModels: string[]
    └── error: string | null

Hooks consume this state:
    ├── useCoreConnection() - connect/disconnect
    ├── useSession() - create, send, cancel
    ├── useProviderSelection() - select provider
    └── useEmbedding() - indexing status
```

## Performance Considerations

### 1. Streaming
- Avoid waiting for complete response
- Stream tokens to UI in real-time
- Update UI incrementally

### 2. Caching
- Cache provider list
- Cache available models
- Cache last N chat sessions

### 3. Request Handling
- Use appropriate timeouts per operation
- Cancel long-running requests on demand
- Implement retry logic for transient failures

## Security Considerations

### 1. IPC Communication
- JSON-RPC over stdio (local process only)
- No network exposure by default
- Core binary must be verified/signed

### 2. File Operations
- Only operate on workspace root
- Validate file paths
- Restrict to configured workspace

### 3. Credentials
- Never store API keys in extension state
- Use VS Code secret storage if needed
- Never log sensitive data

## Future Enhancements

### Phase 3+
1. Real-time collaboration
2. Advanced RAG features
3. Streaming optimization
4. Performance improvements
5. Plugin system for tools

### Known Limitations
- Single session per workspace
- No message persistence (yet)
- No advanced caching (yet)
- No tool execution (future)
