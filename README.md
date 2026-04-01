# Vectora

> [!TIP]
> Read this file in another language.
> English [README.md] | Portuguese [README.pt.md]

**A private NotebookLM that runs entirely on your machine.**

Vectora is a local AI assistant that learns from whatever you give it — documents, code, papers, images — and answers questions based strictly on that content. Think Google NotebookLM, but running on your hardware, with your data never leaving your machine.

No cloud dependency. No recurring cost. No data leaving your machine.

---

## The Problem

You know when you ask an AI about something very specific — a particular version of a framework, an internal document, a niche research paper — and it either makes something up or gives you a generic answer that misses the point entirely?

That happens because the AI has no access to _your_ context. Vectora fixes that. Feed it your files, point it at a knowledge base, and it answers from exactly that — nothing more, nothing less.

---

## How It Works

Vectora embeds your files and downloaded knowledge bases into isolated local vector databases. When you ask a question, it retrieves the most semantically relevant context from whichever workspaces you have active and sends everything — along with your question — to the language model.

```markdown
AI Base Model (Qwen / Gemini)     ← always present, fully general
    + Workspace: Godot 4.2        ← downloaded from Vectora Index
    + Workspace: Physics Papers   ← downloaded from Vectora Index
    + Workspace: Your Files       ← added by you, private
              ↓
         Your Question
              ↓
        Retrieval across active workspaces
              ↓
           Answer
```

Each workspace is a completely isolated namespace. Contexts never bleed into each other. You control which workspaces are active per session.

---

## Vectora Index

The Index is a curated marketplace of knowledge bases — pre-built vector datasets published by the community and reviewed by Kaffyn before becoming available for download.

From inside the Vectora app, you can browse the full catalog with search and filters, read a lightweight README for each dataset describing its contents, download any dataset directly into your local Vectora as a new workspace, and publish your own knowledge bases for others to use.

**Examples of what you'll find in the Index:**

- Godot 4.x documentation (per version)
- Frontend and backend framework references
- Engineering, physics, and computer science papers
- Game design resources, language specs, and more

Every dataset downloaded from the Index is embedded and stored locally. After download, no network request is made at query time.

---

## What Can You Do With It?

**Study & Research**
Drop PDFs, papers, or notes into a workspace. Ask Vectora to explain, summarize, cross-reference, or quiz you. Everything stays local and private.

**Development**
Combine an engine documentation workspace with your own codebase workspace. Get answers that are aware of both the API contract and your actual implementation.

**Deep Work**
Use Gemini mode to index images, PDFs, and audio alongside text — all processed and stored locally after indexing.

**IDE Integration**
Expose any workspace as an MCP server, feeding precise context directly into tools like Cursor, VS Code, or Claude Code.

---

## AI Providers

Vectora supports two providers out of the box, with the engine built to accommodate more in the future:

**Qwen (Local / Offline)**
Runs entirely on your hardware via `llama.cpp`. No internet required. Supports text and code using the Qwen3 models (see section below for details). Ideal for fully private workflows.

**Gemini (Cloud / Multimodal)**
Uses your own Gemini API key, stored only in your local config. Unlocks multimodal indexing — PDFs, images, and audio are all supported. The key never leaves your machine.

Both providers include dedicated embedding models. Vectora does not rely on a separate embedding service.

## Qwen Official Models

Vectora supports the new **Qwen3** lineage, optimized for different development fronts:

**Coding & Reasoning**

- **Qwen3-Coder-Next (80B):** The state-of-the-art for massive refactoring and system architecture.
- **Qwen3-4B-Thinking (2507):** Logical reasoning model (Chain-of-Thought) for complex bug resolution.

**Vision & Multimodal (Thinking VL)**

- **Qwen3-VL-Thinking (2B/8B):** Vision models that "think" about the image, ideal for analyzing UI bug screenshots or architecture diagrams.
- **Qwen3-VL-Embedding (2B):** Vectorization of visual assets and diagrams for semantic search in GDDs.

**Audio & Speech (ASR/TTS)**

- **Qwen3-ASR (0.6B):** Ultra-fast transcription of sprint meetings and feedback audio.
- **Qwen3-TTS-VoiceDesign (1.7B):** High-fidelity voice synthesis (12Hz) for real-time dialogue prototyping.

**RAG & Embeddings**

- **Qwen3-Embedding (0.6B/4B/8B):** The vector search engines that power chromem-go. **We recommend the 0.6B version** for the strict 2GB RAM limit, ensuring your code context is precisely retrieved without compromising system performance.

---

## Interfaces

Vectora is not a single app — it is an ecosystem of interfaces sharing a common core via IPC, all orchestrated by a lightweight systray daemon:

| Interface           | Description                                                                                               |
| ------------------- | --------------------------------------------------------------------------------------------------------- |
| **Systray**         | The core daemon. Lives near your clock, orchestrates everything, ~5MB RAM.                                |
| **Web UI (Wails)**  | Local desktop app powered by Next.js. Chat interface, workspace management, settings, and Index browsing. |
| **CLI (Bubbletea)** | Terminal interface. Minimal footprint, instant response.                                                  |
| **MCP Server**      | Exposes Vectora's knowledge to external AI tools and IDEs.                                                |
| **ACP Agent**       | Autonomous agent mode with filesystem and terminal access.                                                |

---

## Agentic Toolkit

When operating in MCP or ACP mode, Vectora exposes a shared set of tools built from scratch in Go:

- **Filesystem:** `read_file`, `write_file`, `read_folder`, `edit`
- **Search:** `find_files`, `grep_search`, `google_search`, `web_fetch`
- **System:** `run_shell_command`
- **Memory:** `save_memory`, `enter_plan_mode`

> [!IMPORTANT]
> Every write or shell action triggers an automatic Git snapshot via `GitBridge` before execution. Any agentic action can be fully rolled back with a single `undo` command.

---

## Architecture

Vectora is written entirely in Go. The core runs as a lightweight systray daemon and spawns other interfaces on demand via IPC.

| Component       | Technology               | Role                                                        |
| --------------- | ------------------------ | ----------------------------------------------------------- |
| Vector DB       | chromem-go               | Semantic search and embeddings                              |
| Key-Value DB    | bbolt                    | Chat history, logs, config                                  |
| AI Engine       | langchaingo              | LLM and embedding provider abstraction (Gemini, extensible) |
| Local Inference | llama.cpp (sidecar)      | Offline model execution (Qwen)                              |
| Installer       | Fyne                     | Cross-platform setup wizard                                 |
| Tray            | systray                  | Core daemon and orchestrator                                |
| Web UI          | Wails + Next.js (static) | Local desktop chat interface                                |
| CLI             | Bubbletea                | Terminal interface                                          |
| Index Server    | Go (net/http)            | Vector dataset catalog and distribution                     |

The Web UI is built with Next.js in static export mode, embedded into the Wails binary via `go:embed`. The frontend communicates with the Go backend through Wails bindings — no HTTP server, no Node.js runtime, direct JS→Go function calls.

Designed to operate under **4GB of RAM** on modest hardware.

---

## Roadmap

- [ ] Full end-to-end integration (in progress)
- [ ] Public first release
- [ ] Vectora Index public launch
- [ ] Multimodal indexing (images, PDFs) via Gemini
- [ ] Audio transcription and indexing
- [ ] Vectora site and documentation

---

_Part of the [Kaffyn](https://github.com/Kaffyn) open source organization._
