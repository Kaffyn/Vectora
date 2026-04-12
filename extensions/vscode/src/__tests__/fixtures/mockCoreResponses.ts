/**
 * Mock Core Responses - Fixture for Testing
 * Simula respostas do Vectora Core para testes unitários e de integração
 */

import type {
  SessionNewResponse,
  PromptResponse,
  StreamDelta,
  RpcResponse,
  ProvidersListResponse,
  TokenUsage,
  ToolCall,
} from "@/types/core";

// ============================================================================
// Session Management Fixtures
// ============================================================================

export const mockSessionNewResponse: SessionNewResponse = {
  sessionId: "test-session-001",
  metadata: {
    provider: "anthropic",
    model: "claude-3-5-sonnet-20241022",
    createdAt: new Date().toISOString(),
  },
};

export const mockSessionNewResponseOllama: SessionNewResponse = {
  sessionId: "test-session-ollama-001",
  metadata: {
    provider: "ollama",
    model: "llama2",
    baseUrl: "http://localhost:11434",
  },
};

// ============================================================================
// Prompt Response Fixtures
// ============================================================================

export const mockPromptResponse: PromptResponse = {
  sessionId: "test-session-001",
  model: "claude-3-5-sonnet-20241022",
  content: "This is a test response from the Claude model.",
  usage: {
    inputTokens: 150,
    outputTokens: 50,
    totalTokens: 200,
  },
  stopReason: "end_turn",
};

export const mockPromptResponseWithToolCalls: PromptResponse = {
  sessionId: "test-session-001",
  model: "claude-3-5-sonnet-20241022",
  content: "I'll analyze the code and provide refactoring suggestions.",
  toolCalls: [
    {
      id: "tool-call-001",
      name: "read_file",
      input: {
        path: "src/main.ts",
      },
    },
  ],
  usage: {
    inputTokens: 200,
    outputTokens: 100,
    totalTokens: 300,
  },
  stopReason: "tool_use",
};

export const mockStreamingPromptResponse: PromptResponse = {
  sessionId: "test-session-001",
  model: "claude-3-5-sonnet-20241022",
  content: "Streaming response content...",
  usage: {
    inputTokens: 150,
    outputTokens: 250,
    totalTokens: 400,
  },
  stopReason: "end_turn",
};

// ============================================================================
// Stream Delta Fixtures
// ============================================================================

export const mockStreamDeltas: StreamDelta[] = [
  {
    type: "content",
    content: "The function ",
  },
  {
    type: "content",
    content: "calculateTotal ",
  },
  {
    type: "content",
    content: "can be optimized by using reduce().",
  },
  {
    type: "usage",
    usage: {
      inputTokens: 150,
      outputTokens: 50,
      totalTokens: 200,
    },
  },
];

export const mockToolCallStreamDeltas: StreamDelta[] = [
  {
    type: "content",
    content: "Let me ",
  },
  {
    type: "content",
    content: "read ",
  },
  {
    type: "content",
    content: "the file.",
  },
  {
    type: "tool_call",
    toolCall: {
      id: "tool-call-stream-001",
      name: "read_file",
      input: {
        path: "src/index.ts",
      },
    },
  },
];

// ============================================================================
// RPC Response Fixtures
// ============================================================================

export const mockRpcResponseSuccess: RpcResponse<PromptResponse> = {
  jsonrpc: "2.0",
  id: 1,
  result: mockPromptResponse,
};

export const mockRpcResponseError: RpcResponse = {
  jsonrpc: "2.0",
  id: 1,
  error: {
    code: -32601,
    message: "Method not found",
    data: {
      method: "session.invalid_method",
    },
  },
};

export const mockRpcResponseTimeout: RpcResponse = {
  jsonrpc: "2.0",
  id: 2,
  error: {
    code: -32603,
    message: "Internal error",
    data: {
      reason: "Request timeout",
    },
  },
};

// ============================================================================
// Provider Configuration Fixtures
// ============================================================================

export const mockProvidersListResponse: ProvidersListResponse = {
  providers: [
    {
      name: "anthropic",
      available: true,
      models: [
        "claude-3-5-sonnet-20241022",
        "claude-3-5-haiku-20241022",
        "claude-3-opus-20250219",
      ],
    },
    {
      name: "ollama",
      available: false,
      models: ["llama2", "mistral", "neural-chat"],
    },
    {
      name: "openai",
      available: true,
      models: ["gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"],
    },
  ],
};

// ============================================================================
// Tool Call Fixtures
// ============================================================================

export const mockToolCall: ToolCall = {
  id: "tool-001",
  name: "read_file",
  input: {
    path: "src/utils.ts",
  },
};

export const mockToolCalls: ToolCall[] = [
  {
    id: "tool-001",
    name: "read_file",
    input: { path: "src/main.ts" },
  },
  {
    id: "tool-002",
    name: "write_file",
    input: {
      path: "src/main.ts",
      content: "// refactored code",
    },
  },
];

// ============================================================================
// Token Usage Fixtures
// ============================================================================

export const mockTokenUsageSmall: TokenUsage = {
  inputTokens: 100,
  outputTokens: 50,
  totalTokens: 150,
};

export const mockTokenUsageLarge: TokenUsage = {
  inputTokens: 5000,
  outputTokens: 2000,
  totalTokens: 7000,
};

export const mockTokenUsageWithCache: TokenUsage = {
  inputTokens: 1000,
  outputTokens: 500,
  totalTokens: 1500,
  cacheCreationInputTokens: 500,
  cacheReadInputTokens: 0,
};

// ============================================================================
// Error Response Fixtures
// ============================================================================

export const mockErrorInvalidParams = {
  code: -32602,
  message: "Invalid params",
  data: {
    param: "sessionId",
    reason: "Expected UUID format",
  },
};

export const mockErrorSessionNotFound = {
  code: -32001,
  message: "Session not found",
  data: {
    sessionId: "invalid-session-id",
  },
};

export const mockErrorCoreUnreachable = {
  code: -32009,
  message: "Core process unreachable",
  data: {
    reason: "Process terminated unexpectedly",
  },
};

// ============================================================================
// Complex Response Fixtures
// ============================================================================

export const mockMultiToolResponse: PromptResponse = {
  sessionId: "test-session-001",
  model: "claude-3-5-sonnet-20241022",
  content:
    "I'll analyze the codebase and create a refactoring plan. Let me start by exploring the structure.",
  toolCalls: [
    {
      id: "tool-call-001",
      name: "read_file",
      input: { path: "src/components/index.ts" },
    },
    {
      id: "tool-call-002",
      name: "read_file",
      input: { path: "src/utils/helpers.ts" },
    },
    {
      id: "tool-call-003",
      name: "list_files",
      input: { path: "src", recursive: true },
    },
  ],
  usage: {
    inputTokens: 300,
    outputTokens: 150,
    totalTokens: 450,
  },
  stopReason: "tool_use",
};

export const mockLongFormResponse: PromptResponse = {
  sessionId: "test-session-001",
  model: "claude-3-5-sonnet-20241022",
  content: `# Code Analysis Report

## Overview
The codebase shows good structure but has several optimization opportunities.

## Issues Found
1. Unused imports in component files
2. Missing error handling in async operations
3. Inefficient database queries

## Recommendations
- Use ESLint to catch unused imports
- Add try-catch blocks for async functions
- Implement query optimization with proper indexing

## Priority
High - These changes will improve both performance and maintainability.`,
  usage: {
    inputTokens: 200,
    outputTokens: 500,
    totalTokens: 700,
  },
  stopReason: "end_turn",
};
