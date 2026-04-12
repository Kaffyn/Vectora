/**
 * Vectora IPC/MCP Interface Types
 * Define types para comunicação IPC entre extensão e Vectora Core
 * e MCP (Medley Client Protocol) para acesso a workspace
 */

// ============================================================================
// Workspace Query - RAG/Embedding Search
// ============================================================================

export interface SearchResult {
  filename: string;
  content: string;
  startLine?: number;
  endLine?: number;
  relevance: number;
  score?: number;
}

export interface WorkspaceQueryRequest {
  query: string;
  conversationId?: string;
  limit?: number;
  threshold?: number;
}

export interface WorkspaceQueryResponse {
  answer: string;
  sources: SearchResult[];
  confidence?: number;
}

// ============================================================================
// Embedding/Indexing
// ============================================================================

export interface EmbeddingMetadata {
  filename: string;
  totalLines: number;
  totalTokens: number;
  language?: string;
}

export interface EmbeddingStatus {
  isIndexing: boolean;
  progress: number; // 0-100
  filesProcessed: number;
  filesTotal: number;
  vectorCount: number;
  lastUpdate: string; // ISO timestamp
}

export interface IndexingRequest {
  workspaceId: string;
  includePatterns?: string[];
  excludePatterns?: string[];
}

export interface IndexingProgress {
  filesCurrent: number;
  filesTotal: number;
  percentComplete: number;
  currentFile?: string;
  estimatedTimeRemaining?: number; // seconds
}

// ============================================================================
// File Operations
// ============================================================================

export interface FileContent {
  path: string;
  content: string;
  language?: string;
  encoding?: string;
}

export interface FileChangeRequest {
  path: string;
  content: string;
  operation: "create" | "update" | "delete";
}

export interface FileChangeResult {
  path: string;
  success: boolean;
  message?: string;
}

// ============================================================================
// Workspace Structure
// ============================================================================

export interface FileTreeNode {
  path: string;
  name: string;
  type: "file" | "directory";
  children?: FileTreeNode[];
  language?: string;
}

export interface WorkspaceStructure {
  workspaceId: string;
  rootPath: string;
  fileTree: FileTreeNode[];
  totalFiles: number;
  totalDirectories: number;
}

// ============================================================================
// Conversation/Context
// ============================================================================

export interface ConversationContext {
  conversationId: string;
  messages: Array<{
    role: "user" | "assistant";
    content: string;
    timestamp: string;
  }>;
  relevantFiles?: string[];
  embeddingContext?: SearchResult[];
}

export interface SaveConversationRequest {
  conversationId: string;
  title: string;
  summary?: string;
  metadata?: Record<string, unknown>;
}

// ============================================================================
// Configuration
// ============================================================================

export interface WorkspaceConfig {
  workspaceId: string;
  enableRAG: boolean;
  enableEmbedding: boolean;
  embeddingModel?: string;
  chunkSize?: number;
  overlapSize?: number;
  excludePatterns?: string[];
}

// ============================================================================
// Notifications
// ============================================================================

export type IpcNotificationType =
  | "workspace.query.result"
  | "workspace.embedding.progress"
  | "workspace.embedding.complete"
  | "workspace.file.changed"
  | "workspace.index.updated";

export interface IpcNotification<T = any> {
  type: IpcNotificationType;
  data: T;
  timestamp: string;
}

// ============================================================================
// Error Types
// ============================================================================

export interface IpcError {
  code: string;
  message: string;
  workspaceId?: string;
  details?: Record<string, unknown>;
}

// ============================================================================
// Request/Response Wrappers
// ============================================================================

export interface IpcRequest<T = any> {
  id: string;
  method: string;
  params: T;
}

export interface IpcResponse<T = any> {
  id: string;
  result?: T;
  error?: IpcError;
}
