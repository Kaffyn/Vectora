import type { Message } from "./core/types";

export const MOCK_MESSAGES: Message[] = [
  {
    id: "1",
    role: "agent",
    content:
      "# Welcome to Vectora\n\nI am your professional AI coding assistant. I can help you with:\n\n- **RAG**: Answering questions based on your codebase.\n- **Agentic Tools**: Running shell commands, editing files, and more.\n- **Modern UI**: Rendered with React and Tailwind CSS.\n\nTry asking me something!",
    timestamp: Date.now() - 1000 * 60 * 5,
  },
  {
    id: "2",
    role: "user",
    content: "Show me how you handle code snippets and tool calls.",
    timestamp: Date.now() - 1000 * 60 * 2,
  },
  {
    id: "3",
    role: "agent",
    content:
      'Certainly! I can render code with syntax highlighting:\n\n```go\nfunc Hello() {\n    fmt.Println("Hello from Vectora Core")\n}\n```\n\nAnd I can show you the tools I am using in real-time:',
    timestamp: Date.now() - 1000 * 60 * 1,
    tools: [
      { id: "t1", title: "Searching codebase...", status: "completed" },
      { id: "t2", title: "Reading engine.go", status: "completed" },
      { id: "t3", title: "Synthesizing answer", status: "in_progress" },
    ],
  },
];
