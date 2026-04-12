/**
 * Vectora Core Interface Types - ACP Protocol
 * Define types para comunicação entre extensão VS Code e Vectora Core binary
 * baseado no Agent Client Protocol (ACP)
 */

// ============================================================================
// Message Types
// ============================================================================

export type MessageRole = "user" | "assistant" | "system" | "tool";

export interface VectoraMessage {
  role: MessageRole;
  content: string;
  name?: string;
  toolCallId?: string;
  toolName?: string;
}

export interface ToolCall {
  id: string;
  name: string;
  input: Record<string, unknown>;
}

export interface TokenUsage {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  cacheCreationInputTokens?: number;
  cacheReadInputTokens?: number;
}

// ============================================================================
// Session Management
// ============================================================================

export interface SessionNewRequest {
  workspaceId: string;
  provider: string;
  model?: string;
}

export interface SessionNewResponse {
  sessionId: string;
  metadata?: Record<string, unknown>;
}

export interface SessionPromptRequest {
  sessionId: string;
  messages: VectoraMessage[];
  workspaceId?: string;
  temperature?: number;
  topP?: number;
  maxTokens?: number;
  systemPrompt?: string;
}

export interface PromptResponse {
  sessionId: string;
  model: string;
  content: string;
  toolCalls?: ToolCall[];
  usage?: TokenUsage;
  stopReason?: "end_turn" | "tool_use" | "max_tokens";
}

export interface StreamDelta {
  type: "content" | "tool_call" | "usage";
  content?: string;
  toolCall?: ToolCall;
  usage?: TokenUsage;
}

// ============================================================================
// Notification Types
// ============================================================================

export type SessionNotificationType =
  | "session.update"
  | "session.stream_delta"
  | "session.complete"
  | "session.error"
  | "session.tool_call"
  | "session.tool_result";

export interface SessionNotification {
  type: SessionNotificationType;
  sessionId: string;
  timestamp: string;
  data: unknown;
}

// ============================================================================
// Tool Execution
// ============================================================================

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
}

export interface ToolExecutionRequest {
  toolName: string;
  input: Record<string, unknown>;
}

export interface ToolExecutionResult {
  toolName: string;
  content: string;
  isError?: boolean;
}

// ============================================================================
// Workspace Context
// ============================================================================

export interface WorkspaceContext {
  workspaceId: string;
  rootPath: string;
  files?: string[];
  embeddingEnabled?: boolean;
  vectorCount?: number;
}

// ============================================================================
// Error Types
// ============================================================================

export interface CoreError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

// ============================================================================
// RPC Request/Response Envelope
// ============================================================================

export interface RpcRequest<T = any> {
  jsonrpc: "2.0";
  id: string | number;
  method: string;
  params: T;
}

export interface RpcResponse<T = any> {
  jsonrpc: "2.0";
  id: string | number;
  result?: T;
  error?: {
    code: number;
    message: string;
    data?: unknown;
  };
}

export interface RpcNotification<T = any> {
  jsonrpc: "2.0";
  method: string;
  params: T;
}

// ============================================================================
// Provider Configuration
// ============================================================================

export interface ProviderConfig {
  name: string;
  apiKey?: string;
  baseUrl?: string;
  models: string[];
}

export interface ProvidersListResponse {
  providers: Array<{
    name: string;
    available: boolean;
    models: string[];
  }>;
}
