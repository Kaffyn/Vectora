/**
 * Vectora Types - Central Export Point
 * Consolidates all type definitions for easy importing throughout the extension
 */

// Core ACP/IPC Types
export * from "./core";
export * from "./ipc";
export * from "./vectora";

// Client.d.ts types (re-export as named exports)
export type {
  JsonRpcRequest,
  JsonRpcResponse,
  JsonRpcNotification,
  JsonRpcError,
  InitializeRequest,
  ClientCapabilities,
  FsCapabilities,
  ClientInfo,
  InitializeResponse,
  AgentCapabilities,
  PromptCapabilities,
  McpCapabilities,
  SessionNewRequest as ClientSessionNewRequest,
  McpServer,
  SessionNewResponse as ClientSessionNewResponse,
  SessionLoadRequest,
  SessionCancelRequest,
  SessionPromptRequest as ClientSessionPromptRequest,
  ContentBlock,
  Resource,
  PromptResponse as ClientPromptResponse,
  StopReason,
  SessionUpdate as ClientSessionUpdate,
  UpdateData,
  UpdateType,
  PlanEntry,
  ToolContent,
  ToolLocation,
  ToolKind,
  ToolCallStatus,
  RequestPermissionRequest,
  PermissionOption,
  PermissionKind,
  RequestPermissionResponse,
  FSReadRequest,
  FSReadResponse,
  FSWriteRequest,
  FSCompletionRequest,
  FSCompletionResponse,
} from "./client";

// Common utility types
export type Nullable<T> = T | null;
export type Optional<T> = T | undefined;

export type AsyncResult<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };
