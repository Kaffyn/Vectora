/**
 * Vectora Extension Types - Core Message Definitions
 * Define tipos básicos de mensagens e estruturas compartilhadas
 */

// ============================================================================
// Message Types - Unificados para comunicação ACP/IPC
// ============================================================================

export type MessageRole = "user" | "assistant" | "system" | "tool";

export interface VectoraMessage {
  role: MessageRole;
  content: string;
  name?: string;
  toolCallId?: string;
  toolName?: string;
  timestamp?: string;
}

// ============================================================================
// Session Update - Streaming from Core
// ============================================================================

export type SessionUpdateType =
  | "message"
  | "tool_call"
  | "tool_result"
  | "error"
  | "complete"
  | "stream_delta"
  | "usage";

export interface SessionUpdate {
  type: SessionUpdateType;
  sessionId?: string;
  content?: string;
  toolName?: string;
  toolInput?: Record<string, unknown>;
  delta?: string; // Para streaming de tokens
  usage?: {
    inputTokens: number;
    outputTokens: number;
    totalTokens: number;
  };
  error?: {
    code: string;
    message: string;
  };
  timestamp?: string;
}

// ============================================================================
// User Request/Command
// ============================================================================

export interface UserRequest {
  type: "prompt" | "tool_execution" | "context_update";
  content: string;
  metadata?: Record<string, unknown>;
}

// ============================================================================
// Provider Information
// ============================================================================

export interface Provider {
  id: string;
  name: string;
  icon?: string;
  available: boolean;
  models?: string[];
}

export interface ProviderSelection {
  providerId: string;
  modelId?: string;
}

// ============================================================================
// Extension State
// ============================================================================

export interface ExtensionState {
  coreStatus: "starting" | "running" | "stopped" | "error";
  isConnected: boolean;
  selectedProvider: string | null;
  selectedModel: string | null;
  workspace: string | null;
  currentSessionId: string | null;
  isProcessing: boolean;
  availableProviders: Provider[];
  availableModels: string[];
  error: string | null;
  workspaceEmbeddingStatus?: {
    isIndexing: boolean;
    progress: number;
    vectorCount: number;
  };
}

// ============================================================================
// Chat History
// ============================================================================

export interface ChatMessage extends VectoraMessage {
  id: string;
  timestamp: string;
  sessionId?: string;
  usage?: {
    inputTokens: number;
    outputTokens: number;
    totalTokens: number;
  };
}

export interface ChatSession {
  id: string;
  title: string;
  provider: string;
  model: string;
  createdAt: string;
  updatedAt: string;
  messages: ChatMessage[];
  metadata?: Record<string, unknown>;
}

// ============================================================================
// Tool Information
// ============================================================================

export interface Tool {
  name: string;
  description: string;
  inputSchema: {
    type: string;
    properties: Record<string, unknown>;
    required?: string[];
  };
}

export interface ToolResult {
  toolName: string;
  content: string;
  isError: boolean;
  metadata?: Record<string, unknown>;
}

// ============================================================================
// Configuration & Settings
// ============================================================================

export interface VectoraSettings {
  coreExecutablePath?: string;
  corePort?: number;
  defaultProvider?: string;
  enableRAG?: boolean;
  enableEmbedding?: boolean;
  ragChunkSize?: number;
  ragOverlapSize?: number;
  theme?: "light" | "dark" | "auto";
}

// ============================================================================
// Error Types
// ============================================================================

export interface VectoraError {
  code: string;
  message: string;
  context?: string;
  details?: Record<string, unknown>;
}

// ============================================================================
// Feature Flags
// ============================================================================

export interface FeatureFlags {
  ragEnabled: boolean;
  embeddingEnabled: boolean;
  streamingEnabled: boolean;
  multiProviderEnabled: boolean;
  advancedUIEnabled: boolean;
}
