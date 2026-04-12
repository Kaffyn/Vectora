/**
 * Test Fixtures - Central Export
 * Todas as fixtures de teste estão centralizadas aqui
 */

// Core response fixtures
export * from "./mockCoreResponses";

// Message fixtures
export * from "./testMessages";

// Workspace fixtures
export * from "./testWorkspaces";

// ============================================================================
// Test Data Builders
// ============================================================================

import type {
  SessionNewResponse,
  PromptResponse,
  VectoraMessage,
  WorkspaceContext,
} from "@/types/core";

/**
 * Builder para criar respostas de sessão customizadas
 */
export class SessionResponseBuilder {
  private response: SessionNewResponse;

  constructor(baseResponse: SessionNewResponse) {
    this.response = { ...baseResponse };
  }

  withSessionId(sessionId: string): this {
    this.response.sessionId = sessionId;
    return this;
  }

  withMetadata(metadata: Record<string, unknown>): this {
    this.response.metadata = { ...this.response.metadata, ...metadata };
    return this;
  }

  build(): SessionNewResponse {
    return this.response;
  }
}

/**
 * Builder para criar respostas de prompt customizadas
 */
export class PromptResponseBuilder {
  private response: PromptResponse;

  constructor(baseResponse: PromptResponse) {
    this.response = JSON.parse(JSON.stringify(baseResponse));
  }

  withSessionId(sessionId: string): this {
    this.response.sessionId = sessionId;
    return this;
  }

  withContent(content: string): this {
    this.response.content = content;
    return this;
  }

  withModel(model: string): this {
    this.response.model = model;
    return this;
  }

  addTokenUsage(usage: { inputTokens: number; outputTokens: number }): this {
    this.response.usage = {
      inputTokens: usage.inputTokens,
      outputTokens: usage.outputTokens,
      totalTokens: usage.inputTokens + usage.outputTokens,
    };
    return this;
  }

  build(): PromptResponse {
    return this.response;
  }
}

/**
 * Builder para criar mensagens customizadas
 */
export class MessageBuilder {
  private message: VectoraMessage;

  constructor(role: "user" | "assistant" | "system" | "tool") {
    this.message = { role, content: "" };
  }

  withContent(content: string): this {
    this.message.content = content;
    return this;
  }

  withName(name: string): this {
    this.message.name = name;
    return this;
  }

  withToolCall(toolCallId: string, toolName: string): this {
    this.message.toolCallId = toolCallId;
    this.message.toolName = toolName;
    return this;
  }

  build(): VectoraMessage {
    return this.message;
  }
}

/**
 * Builder para criar contextos de workspace customizados
 */
export class WorkspaceContextBuilder {
  private context: WorkspaceContext;

  constructor(workspaceId: string, rootPath: string) {
    this.context = {
      workspaceId,
      rootPath,
    };
  }

  withFiles(files: string[]): this {
    this.context.files = files;
    return this;
  }

  withEmbeddings(enabled: boolean, vectorCount: number = 0): this {
    this.context.embeddingEnabled = enabled;
    this.context.vectorCount = vectorCount;
    return this;
  }

  build(): WorkspaceContext {
    return this.context;
  }
}
