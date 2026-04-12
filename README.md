# Vectora

> [!TIP]
> Read this file in another language | Leia esse arquivo em outro idioma.
> [English](README.md) | [Português](README.pt.md)

> [!NOTE]
>
> **Hybrid RAG for real-world codebases.**
> Understand your code as a system — not as text.

---

**Traditional RAG operates on isolated chunks.
Vectora operates on real system relationships.**

- semantic search (embeddings)
- code structure (files, functions, dependencies)
- codebase graph
- multi-hop reasoning

Result: answers based on how the system actually works — not just on isolated snippets.

---

- **Hybrid RAG (Core):** Semantic + structural + relational retrieval
- **Codebase-Aware:** Understands real relationships between modules and files
- **Real Multi-hop:** Navigates multiple points in the system to answer complex questions
- **Agentic Execution:** Acts on code with context and safety
- **Total Privacy:** Data and embeddings stay local

---

## Why Vectora?

Most solutions require combining multiple tools:

- vector database
- RAG framework
- embeddings
- orchestration

Vectora integrates all of this into a single local core, ready for use.

---

## Agentic Operation: Specialized Sub-Agent for your IDE

Vectora acts as a **Sub-Agent (Tier 2)** coupled to your IDE, designed to execute complex tasks based on a real understanding of the codebase.

While the main agent (Tier 1) orchestrates user interaction, Vectora takes on tasks that require **deep retrieval, structural analysis, and controlled execution**.

- **High-Level Delegation:**
  The IDE can delegate operations such as multi-file refactorings, impact analysis, and dependency navigation. Vectora executes based on the actual state of the code.

- **Context-Aware Execution:**
  All actions are guided by contextual retrieval (RAG), ensuring that reads, writes, and commands are based on the actual structure and relationships of the project.

- **Scope-Restricted Execution:**
  Tools operate exclusively within the Trust Folder, ensuring predictability and isolation.

- **Transactional Security (Git Snapshots):**
  Code modifications are preceded by automatic snapshots, allowing for immediate reversion.

---

## Retrieval Engine: Systemic Understanding of the Codebase

Traditional RAG operates on isolated chunks.
Vectora retrieves **connected context**.

- **Retrieval Layers:**
  Combines semantic search (embeddings) with structural code analysis (files, functions, modules, and dependencies).

- **Codebase Graph:**
  The project is modeled as a graph of relationships between entities, allowing for navigation beyond file boundaries.

- **Multi-hop Reasoning:**
  Queries traverse multiple points in the system, connecting dependencies and execution flows to answer questions that require global context.

---

## Installation and Integration

Vectora is designed to adapt to your workflow:

**1. Portable Binary (Manual & MCP Server):**

Ideal for full control or usage outside of IDEs.

- **Download:** Download `vectora.exe` from the [Releases](https://github.com/Kaffyn/Vectora/releases) page.
- **CLI Mode:** Interactive interface for indexing and quick semantic searches via terminal.
- **MCP Server Mode:** Configure the binary path as an MCP server for tools like **Claude Desktop**, **Claude Code (CLI)**, and **Gemini CLI**. The core communicates via **stdio**, exposing its agentic arsenal to any compatible client.

**2. IDE Extensions & CLI Clients:**

- **VS Code Extension (Agent & Sub-Agent Mode):**
  - **Agent Mode:** Full interactive experience with a dedicated chat panel, tools, and visual feedback directly in your IDE.
  - **Sub-Agent Mode:** Integrates natively as a specialized code-reasoning engine that can be invoked by the main chat or other agents (like **Antigravity**) to handle complex, repo-wide tasks.
  - **Bundled:** The extension includes the Vectora Core binary; no separate download required.

- **Gemini CLI (Sub-Agent Mode):**
  - **Integration:** Operates exclusively as a **Sub-Agent** for the Gemini CLI. By exposing its RAG and agentic arsenal via MCP, Vectora allows the Gemini CLI to understand and act on complex codebases natively.

- **Claude Code (Sub-Agent Mode via MCP):**
  - **Integration:** Connect Vectora as a specialized code analysis and RAG engine via the MCP (Model Context Protocol).
  - **Setup:** Configure `~/.claude/settings.json` to add Vectora MCP server. No extension required.
  - **Capabilities:** Vectora exposes 11 specialized tools including semantic search, pattern analysis, test generation, and code refactoring.
  - **Documentation:** See [CLAUDE_CODE_INTEGRATION.md](CLAUDE_CODE_INTEGRATION.md) for complete setup instructions and examples.

---

## SDKs & Protocols

Vectora relies on three foundational integration pillars for maximum scalability and enterprise stability:

### 1. Orchestration Protocols (ACP & MCP)

- **MCP (Model Context Protocol):** Context-focused operation. Enables Parent Agents (like Claude Code or Antigravity) to use Vectora as a deep-search and RAG Sub-Agent.
- **ACP (Agent Client Protocol):** Action-focused operation (JSON-RPC over stdio). Vectora directly connects VS Code extensions (Agent) or CLI to the Core with zero overhead.

### 2. Provider Models SDKs

The Core natively implements complex parsers (streaming, tool calls) in Go through the strict adoption of official SDKs, ensuring maximum reliability:

- **google/genai** (Gemini 3.1 Pro, Flash, and Embedding 2.0)
- **anthropic-sdk-go** (Claude 4.6 Sonnet/Opus, Claude 4.5 Haiku)
- **openai-go** (GPT-5.4 Pro/Mini, Qwen 3.6 API interoperability, Text Embeddings 3)
- **voyageai** (Advanced Voyage-3 Large/Code Embeddings)

### 3. Supported AI Families (Standard April 2026)

Vectora is designed to work with the 10 most powerful AI families on the market:

| Family         | Frontier Models (2026)                   |
| :------------- | :--------------------------------------- |
| **Google**     | Gemini 3.1 Pro, Gemini 3 Flash, Gemma 4  |
| **Anthropic**  | Claude 4.6 Sonnet/Opus, Claude 4.5 Haiku |
| **OpenAI**     | GPT-5.4 Pro, GPT-5.4 Mini, GPT-5-o1      |
| **Alibaba**    | Qwen 3.6-Plus, Qwen 3.6-Turbo, Qwen-Max  |
| **Voyage AI**  | Voyage-3 Large, Voyage-3 Code            |
| **Meta**       | Muse Spark, Llama 4 (Scout/Maverick)     |
| **Microsoft**  | Phi-4-Reasoning-Vision, Phi-4-Medium     |
| **DeepSeek**   | DeepSeek-V3.2, V3.2-Speciale             |
| **Mistral AI** | Mistral Small 4, Mistral Large 3         |
| **xAI**        | Grok 4.20, Grok 4.1                      |
| **Zhipu AI**   | GLM-5.1, GLM-5-Flash                     |

> [!TIP]
> For deep technical verification, official documentation links, and search terms for each model, see **[AGENTS.md](file:///c:/Users/bruno/Desktop/Vectora/AGENTS.md)**.

### 4. Multi-Tenancy Protocol (MTP)

The Core runs as a **Singleton Daemon** in the background, consuming minimal RAM. Opening multiple IDEs does not spawn multiple Vectora processes; instead, the architecture handles concurrent environments through the _Multi-Tenancy Protocol_, establishing shielded Namespaces bound per IPC connection, balancing AI request queues in isolated scopes, and protecting Trust Folder boundaries.

---

## Quick Start (CLI Cobra)

```bash
# 1. Start the core in the project folder
vectora start

# 2. Configure your key (Ex: Gemini)
vectora config --gemini YOUR_KEY

# 3. Generate embeddings
vectora embed

# 4. Ask about the code
vectora ask "How does authentication work?"
```

## 📚 Documentation & Integration Guides

Complete reference and setup guides for integrating Vectora with different environments:

- **[CLAUDE_CODE_INTEGRATION.md](CLAUDE_CODE_INTEGRATION.md)** — Configure Vectora as an MCP sub-agent for Claude Code. Quick start, configuration options, examples, and troubleshooting.

- **[MCP_PROTOCOL_REFERENCE.md](MCP_PROTOCOL_REFERENCE.md)** — Technical reference for all 11 MCP tools exposed by Vectora. Complete API documentation with examples, error handling, and tool chains.

- **[TOOLS_DOCUMENTATION.md](TOOLS_DOCUMENTATION.md)** — Comprehensive user documentation for all Vectora tools (read_file, write_file, grep_search, etc). When to use, when not to use, performance characteristics.

- **[TOOLS_API_REFERENCE.md](TOOLS_API_REFERENCE.md)** — Technical API reference for all tools with function signatures, parameter details, return types, and error handling patterns.

- **[LANGUAGE_SUPPORT.md](LANGUAGE_SUPPORT.md)** — Multi-language support documentation. Language preference flow from systray to LLM with conversation examples in Portuguese and English.

- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)** — Summary of tools documentation improvements and query optimization architecture.

---

## 🧪 Testing

O Vectora utiliza uma estratégia de testes em pirâmide para garantir estabilidade tanto no Core quanto na Extensão.

### Go Core

Testes unitários e de integração com motor de mock persistente.

```bash
go test ./... -race
```

### VS Code Extension

Testes unitários para o cliente IPC e integração de UI via `vscode-test`.

```bash
cd extensions/vscode
npm install
npm test
```

### E2E Smoke Tests

Validação do fluxo completo usando o Core mockado. Rodado automaticamente como parte da suíte `npm test`.

---

## 📜 License

MIT

---

## Use Cases

- **Legacy Codebase Mastery:**
  Understand unknown systems quickly with real context.

- **Niche Documentation:**
  Answers based on specific versions of technical documentation.

- **Requirement Correlation:**
  Compare specifications with actual implementation.

- **Accelerated Onboarding:**
  Learn patterns and architectural decisions directly from the codebase.

---

## Tech Stack

- **Language:** Go 1.23+ (Golang)
- **Architecture:** Singleton Daemon with Multi-Tenancy Protocol (MTP)
- **Interface:** **Cobra CLI** (Terminal) & **Systray** (Desktop)
- **Vector DB:** **chromem-go** (Local-first RAG engine)
- **Metadata Store:** **BBolt** (ACID persistence for history and logs)
- **Models (Default):** Gemini 3.1 Pro (Reasoning) & Gemini Embedding 2 (RAG)
- **Protocols:** ACP (Agent Client Protocol) & MCP (Model Context Protocol)
- **Inter-Process Communication (IPC):** JSON-RPC 2.0 over Named Pipes (Windows) or Unix Sockets (POSIX)
- **Local Inference:** **llama.cpp** integration (Qwen 3.6)
- **Optimization:** TurboQuant (KV-Cache Compression)

---

## Privacy and Control (Trust Folder)

- **Master Directory (Config):** Your global settings, logs, and API keys stay isolated in %USERPROFILE%/.vectora.
- **Trust Folder (Scope):** By default, the location where you run **Vectora Core** is your "Trust Folder". The core will only have permission to read or modify files within this scope.
- **Repositioning:** If you want to grant access to another location on the system, use the --path flag via CLI (e.g., vectora start --path D:\MyProject) to reposition the trust folder.
- **Hard-Coded Guardian:** Even within the Trust Folder, Vectora automatically ignores sensitive files like .env, .key, and .pem.

---

## Agentic Toolkit (Industrial Grade)

Exposed via MCP/ACP and built from scratch in Go, the Agentic Toolkit is the arsenal of tools that transforms Vectora from a simple chatbot into an operational agent capable of directly interacting with your system — always within the Trust Folder and with transactional security.

### File and System Tools

- **`read_file`:** Precise reading of individual files with line pagination support for large files.
- **`write_file`:** Complete writing of new files or controlled overwrite (with automatic Git snapshot).
- **`read_folder`:** Recursive directory listing with structure metadata (size, modification, permissions).
- **`edit`:** Intelligent patching — refined editing of specific snippets without rewriting the entire file. Locates the exact context via semantic search-and-replace and applies surgical changes.

### Search and Retrieval Tools

- **`find_files`:** Fast search using glob patterns (e.g., `**/*.ts`, `src/**/*.tsx`) with name and extension filtering.
- **`grep_search`:** Powerful ripgrep-based search with full regex support, file type filtering, and result limiting.
- **`google_search`:** Web search integration to bring in up-to-date external context — ideal for recent documentation, changelogs, and troubleshooting.
- **`web_fetch`:** Direct URL fetching, converting HTML to markdown and extracting the relevant content for the conversation context.

### System and Terminal Tools

- **`run_shell_command`:** Shell command execution in a controlled environment with real-time stdout/stderr capture. Supports background execution for long-running processes and configurable timeout.

### Memory and Planning Tools

- **`save_memory`:** Persistence of important facts and user preferences in long-term memory (global or per-project), enabling continuous personalization across sessions.
- **`enter_plan_mode`:** Activation of structured planning mode — before executing complex tasks, the agent elaborates a step-by-step plan, validates it with the user, and only then begins implementation.

### Security and Isolation

All tools operate exclusively within the user-defined **Trust Folder**. The **Hard-Coded Guardian** automatically blocks reading/writing of sensitive files (`.env`, `.key`, `.pem`, databases, binaries), regardless of the model's intelligence or prompt instructions.

---

## Key Features

**Hybrid Brain (Cloud Intelligence + Local Data):**

Vectora uses the power of Gemini 3.1 Pro for logical reasoning and Gemini Embedding 2.0 to create high-precision vectors. While intelligence resides in the cloud, knowledge data and vector storage remain local, ensuring fast on-demand retrieval.

**Agentic Arsenal (Industrial Grade):**

Unlike purely chat models, Vectora has real tools to interact with your system (within the Trust Folder):

- **Files:** Read, write, list, and edit (read, write, ls, edit).
- **IDE Git Support:** Vectora operates in harmony with the user's Git environment, allowing the IDE's integrated version control to manage history and snapshots.
- **Terminal:** Shell command execution with real-time output capture.
- **Knowledge Search:** Deep semantic search in your local workspaces using chromem-go.

## Core Architecture

| Component           | Technology             | Role                                                                       |
| ------------------- | ---------------------- | -------------------------------------------------------------------------- |
| **Vector DB**       | **chromem-go**         | Semantic search and embeddings                                             |
| **Key-Value DB**    | **bbolt**              | Chat history, logs, configuration                                          |
| **IA Motor**        | **Direct Calls**       | Optimized HTTP/STDIO calls to APIs and `llama.cpp`. No framework overhead. |
| **Local Inference** | **llama.cpp (native)** | Offline model execution (Qwen3) via native system integration              |
| **Vectora Core**    | **Cobra + Systray**    | CLI, Systray, IPC (local), JSON-RPC/stdio (remote)                         |

---

## How to Use

1. Run `vectora.exe`
2. The core starts in the current directory
3. Configure your API key
4. Connect your IDE or agent

---

## Future Plans

Vectora Core is the foundation of a larger ecosystem focused on hybrid RAG and AI-assisted development.

**100% Local Mode:**

- Native integration with **llama.cpp**
- Complete offline execution (including embeddings)
- Support for **Qwen 3.5** and **Qwen 3.6** models

**TurboQuant (Extreme Efficiency):**

- Extreme KV Cache compression (3 to 3.5 bits per value)
- Processing of massive contexts (128k to 1M tokens) 100% locally
- Near-zero accuracy loss with PolarQuant and QJL Corrector technology

**Recovery Engine Evolution:**

- Codebase graph enhancement (richer relationships)
- More efficient multi-hop with lower latency
- Hybrid ranking (semantic + structural + relational)

**Vectora Assets:**

- Vector knowledge base marketplace
- Official documentations, tech specs, and curated datasets
- Instant download and indexing

**Advanced Interfaces:**

- **Vectora Desktop (Fyne):** Native UI for workspace management and graph navigation
- **Vectora TUI (Bubbletea):** Optimized terminal interface for productivity

**Cloud Services:**

- **Vectora Web:** remote access to your workspace
- **Vectora Auth:** authentication and access control (RBAC)
- **Collaboration:** secure knowledge sharing between teams

---

_Part of the Kaffyn open source organization._
