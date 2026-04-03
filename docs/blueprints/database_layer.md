# Blueprint: Database Layer & Persistence

**Status:** Implementation Complete / Industrial Grade
**Module:** `internal/db/`

---

## 1. Overview
The database layer manages both semantic (vector) and relational-style (key-value) persistence. It is designed to be fully self-contained and "local-first".

## 2. Engines
Vectora utilizes two embedded Go-native engines to avoid external dependencies:

### 2.1 Vector Store (Chromem-Go)
- **Purpose**: Storage and retrieval of document embeddings.
- **Implementation**: Uses `chromem.NewPersistentDB` for folder-based persistence.
- **Workflow**: Documents are chunked, vectorized, and stored in collections isolated by Workspace ID (e.g., `ws_abc-123`).

### 2.2 KV Store (BBolt)
- **Purpose**: Storage of application metadata, chat history, and user memories.
- **Implementation**: High-performance, thread-safe key-value store.
- **Hierarchy**:
  - `conversations`: Stores chat sessions as JSON blobs indexed by UUID.
  - `workspaces`: Metadata about local ingestion folders.
  - `memories`: Long-term user preferences for prompt injection.

## 3. Data Safety
- **Single Instance**: BBolt enforces single-process access via file locking.
- **Atollic Transactions**: All writes are transactional to prevent data corruption during system crashes.
- **AppData Directory**: All data is stored in the user's local application data directory (`%LOCALAPPDATA%` on Windows).
