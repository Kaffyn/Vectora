# Vectora Backend Implementation

## Overview

Complete Go backend implementation for Vectora, including infrastructure, IPC communication, and database persistence.

## Architecture

The system follows a layered architecture:

- Frontend (CLI, App, Installer) communicates via IPC
- IPC Server handles request/response messages
- Handler Registry routes messages to business logic
- Database Store persists all data

## File Structure

```
internal/
├── infra/
│   ├── config.go           Configuration management
│   ├── config_test.go      Config tests
│   └── logger.go           Structured logging
│
├── ipc/
│   ├── types.go            IPC message types
│   ├── server.go           IPC server implementation
│   ├── server_test.go      Server tests
│   └── client.go           IPC client implementation
│
└── db/
    ├── store.go            Database abstraction layer
    └── store_test.go       Store tests

cmd/
└── vectora-daemon/
    └── main.go             Daemon entry point
```

## Components

### 1. Infrastructure (infra/)

Config: Loads configuration from .env and environment variables
Logger: Structured logging using slog package

### 2. IPC Communication (ipc/)

Types: Message protocol definitions
Server: IPC server with handler routing
Client: Client implementation

### 3. Database Layer (db/)

Store: BoltDB abstraction with multiple buckets

## Registered Handlers

- ping: Health check
- server.info: Server information
- db.stats: Database statistics
- workspace.save: Save workspace data
- workspace.get: Get workspace data
- workspace.list: List all workspaces
- workspace.delete: Delete workspace

## Building and Running

Build daemon:
  go build -o vectora-daemon ./cmd/vectora-daemon

Run daemon:
  ./vectora-daemon

Run tests:
  go test ./internal/...

## Configuration

Environment variables or .env file at ~/.Vectora/.env:
- GEMINI_API_KEY
- MAX_RAM_DAEMON
- MAX_RAM_INDEXING
- PREFERRED_LLM_PROVIDER
- LOG_LEVEL

## Database

Location: ~/.Vectora/db/vectora.db
Type: BoltDB embedded key-value store
No separate database server needed

## Next Steps for RAG Pipeline

1. Vector Embeddings: Integrate Chroma-go
2. Document Processing: Chunking and embedding generation
3. Semantic Search: Query vector similarity
4. LLM Integration: Gemini API or local LLM

