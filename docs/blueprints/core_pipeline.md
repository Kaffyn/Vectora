# Blueprint: RAG Pipeline & Orchestration

**Status:** Implementation Complete / Industrial Grade
**Module:** `internal/core/rag_pipeline.go`

---

## 1. Overview
The RAG (Retrieval-Augmented Generation) Pipeline is the central orchestrator of Vectora. It manages the flow of information between the user's interface (via IPC) and the storage/inference backends.

## 2. Architecture Logic
The pipeline follows a synchronous standard flow for queries:

1. **Embedding Injection**: The user query is vectorized using the active `llm.Provider`.
2. **Semantic Search**: The vector is sent to `internal/db` (Chromem-Go), which performs a KNN (K-Nearest Neighbors) search across active workspaces.
3. **Context Construction**: The top 5 relevant chunks are extracted and flattened into a single context string.
4. **Agentic Tool Selection**: The `AgentContext` registry is polled for available tools, providing the LLM with the ability to act on the system.
5. **LLM Synthesis**: The final prompt (System Prompt + Context + User Query + Tool Specs) is sent to the LLM.

## 3. Key Constraints
- **TOP K = 5**: Fixed to protect RAM on modest hardware.
- **Low Latency**: The workflow is designed for instant feedback.
- **Privacy**: No external network calls are made during search or context preparation.

## 4. ReAct Loop (Future)
The system currently supports single-shot tool calling. Future iterations will implement a maximum 3-turn ReAct loop to allow the LLM to verify search results before answering.
