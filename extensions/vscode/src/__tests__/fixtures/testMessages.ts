/**
 * Test Messages - Fixture for Testing
 * Mensagens de teste para simular interações do usuário
 */

import type { VectoraMessage, RpcRequest } from "@/types/core";

// ============================================================================
// User Messages
// ============================================================================

export const mockUserMessages: VectoraMessage[] = [
  {
    role: "user",
    content: "Explain what this function does",
  },
  {
    role: "user",
    content: "How can I improve this code?",
  },
  {
    role: "user",
    content: "Write a unit test for the calculateTotal function",
  },
  {
    role: "user",
    content: "What are the dependencies in this project?",
  },
];

// ============================================================================
// Conversation Message Sequences
// ============================================================================

export const mockConversationSequence: VectoraMessage[] = [
  {
    role: "user",
    content: "Analyze the main.ts file",
  },
  {
    role: "assistant",
    content:
      "I'll analyze the main.ts file for you. Let me read it first.",
    toolCallId: "tool-call-001",
    toolName: "read_file",
  },
  {
    role: "tool",
    content: "File contents: export function main() { ... }",
    toolCallId: "tool-call-001",
  },
  {
    role: "assistant",
    content:
      "The main.ts file contains the entry point of your application. It exports a main function that initializes the app.",
  },
];

export const mockRAGConversation: VectoraMessage[] = [
  {
    role: "user",
    content: "How does the authentication system work in our codebase?",
  },
  {
    role: "assistant",
    content:
      "Based on the codebase, I found relevant files about authentication. Let me retrieve the detailed information.",
  },
  {
    role: "user",
    content: "Can you show me the implementation?",
  },
  {
    role: "assistant",
    content:
      "The authentication system uses JWT tokens. Here's the implementation...",
  },
];

// ============================================================================
// System Prompts
// ============================================================================

export const mockSystemPromptBasic =
  "You are a helpful AI coding assistant. Help the user with code analysis, refactoring, and explanations.";

export const mockSystemPromptWithRAG =
  "You are an expert code reviewer with access to the entire codebase via RAG. Provide detailed analysis based on the actual code structure.";

export const mockSystemPromptWithTools =
  "You are an AI assistant with access to tools like reading files, executing code, and modifying documents. Use these tools to help the user effectively.";

// ============================================================================
// RPC Request Fixtures
// ============================================================================

export const mockSessionNewRequest: RpcRequest = {
  jsonrpc: "2.0",
  id: 1,
  method: "session.new",
  params: {
    workspaceId: "test-workspace",
    provider: "anthropic",
    model: "claude-3-5-sonnet-20241022",
  },
};

export const mockPromptRequest: RpcRequest = {
  jsonrpc: "2.0",
  id: 2,
  method: "session.prompt",
  params: {
    sessionId: "test-session-001",
    messages: [
      {
        role: "user",
        content: "Explain this code",
      },
    ],
  },
};

export const mockPromptWithSystemPromptRequest: RpcRequest = {
  jsonrpc: "2.0",
  id: 3,
  method: "session.prompt",
  params: {
    sessionId: "test-session-001",
    messages: [
      {
        role: "user",
        content: "Review my code",
      },
    ],
    systemPrompt: mockSystemPromptWithRAG,
    temperature: 0.7,
    maxTokens: 2000,
  },
};

export const mockStreamPromptRequest: RpcRequest = {
  jsonrpc: "2.0",
  id: 4,
  method: "session.prompt.stream",
  params: {
    sessionId: "test-session-001",
    messages: [
      {
        role: "user",
        content: "Generate a comprehensive analysis",
      },
    ],
  },
};

// ============================================================================
// Message with Different Content Types
// ============================================================================

export const mockShortMessage: VectoraMessage = {
  role: "user",
  content: "Hello",
};

export const mockLongMessage: VectoraMessage = {
  role: "user",
  content: `I need help with a complex refactoring task. My application has:
1. Multiple microservices written in Node.js
2. A React frontend
3. A PostgreSQL database
4. Docker containers for deployment

The main issues I'm facing:
- Slow API response times
- High memory usage in the backend
- Complex state management in the frontend

Can you help me analyze and suggest improvements?`,
};

export const mockCodeMessage: VectoraMessage = {
  role: "user",
  content: `Here's a function that needs optimization:

\`\`\`typescript
function processArray(arr: number[]) {
  const result = [];
  for (let i = 0; i < arr.length; i++) {
    for (let j = 0; j < arr.length; j++) {
      if (arr[i] + arr[j] === 10) {
        result.push([arr[i], arr[j]]);
      }
    }
  }
  return result;
}
\`\`\`

How can I make this more efficient?`,
};

export const mockJSONMessage: VectoraMessage = {
  role: "user",
  content: JSON.stringify(
    {
      action: "analyze",
      target: "performance",
      files: ["src/main.ts", "src/utils.ts"],
    },
    null,
    2,
  ),
};

// ============================================================================
// Conversation Context
// ============================================================================

export const mockContextWithHistory: VectoraMessage[] = [
  {
    role: "system",
    content:
      "You are a TypeScript expert. You have been helping the user refactor their codebase.",
  },
  {
    role: "user",
    content: "Start analyzing my project",
  },
  {
    role: "assistant",
    content: "I'll analyze your project structure. Let me start by examining the files.",
  },
  {
    role: "user",
    content: "Great! What do you think about the architecture?",
  },
  {
    role: "assistant",
    content:
      "The architecture is well-organized with clear separation of concerns. However, I noticed some optimization opportunities...",
  },
];

// ============================================================================
// Error Messages
// ============================================================================

export const mockErrorMessage: VectoraMessage = {
  role: "assistant",
  content: "I encountered an error while processing your request. Please try again.",
};

export const mockWarningMessage: VectoraMessage = {
  role: "assistant",
  content:
    "Warning: This operation might affect performance. Proceed with caution.",
};

// ============================================================================
// Message Batch Fixtures
// ============================================================================

export const mockMessageBatch: VectoraMessage[] = [
  { role: "user", content: "What is the purpose of index.ts?" },
  { role: "assistant", content: "The index.ts file serves as the entry point..." },
  { role: "user", content: "Can you refactor it?" },
  { role: "assistant", content: "Here's an optimized version..." },
  { role: "user", content: "What about error handling?" },
];

// ============================================================================
// Unicode and Special Characters
// ============================================================================

export const mockUnicodeMessage: VectoraMessage = {
  role: "user",
  content:
    "Please help me with my code! 🚀 Análise em português: função está lenta 😞",
};

export const mockSpecialCharsMessage: VectoraMessage = {
  role: "user",
  content: "Check this regex: /^[a-zA-Z0-9_.-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$/",
};

export const mockMultilineMessage: VectoraMessage = {
  role: "user",
  content: `Line 1
Line 2
Line 3

Code block:
\`\`\`
const x = 42;
\`\`\`

Final line`,
};
