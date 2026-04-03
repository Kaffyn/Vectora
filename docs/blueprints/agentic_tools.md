# Blueprint: Agentic Tools & ACP

**Status:** Implementation Complete / Industrial Grade
**Module:** `internal/tools/` & `internal/acp/`

---

## 1. Overview
The agentic toolkit allows the LLM to interact with the host system (filesystem, web, shell). Every tool is exposed via a JSON Schema contract.

## 2. The ACP (Agent Context Protocol)
The ACP registry manages the "arsenal" of tools available to the Agent.
- **Tool Mapping**: Transparently maps LLM JSON calls (tool_calls) to binary Go execution.
- **Registry Pool**: All tools are instantiated in a singular repository before being injected into the RAG Pipeline.

## 3. Toolkit Categories
### 3.1 Filesystem (Safe Write)
- `read_file`, `write_file`, `edit`, `find_files`.
- **Rule**: Every mutation (`write`, `edit`) triggers an automatic **GitBridge** snapshot.

### 3.2 Information Retrieval
- `grep_search`: recursive pure-Go search.
- `google_search`: no-key DuckDuckGo fallback for open web scraping.
- `web_fetch`: full HTML reader with token truncation.

## 4. Safety & Governance
- **Snapshot ID**: Every mutation tool returns a `SnapshotID`, allowing the user UI to offer a one-click "Undo".
- **Context Limit**: Web fetch results are truncated at 8000 tokens to protect inference window.
