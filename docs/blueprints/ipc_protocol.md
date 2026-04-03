# Blueprint: IPC Protocol (JSON-ND)

**Status:** Implementation Complete / Industrial Grade
**Module:** `internal/ipc/`

---

## 1. Overview
The IPC (Inter-Process Communication) layer is the nervous system of Vectora. It allows the Daemon (orchestrator) to talk to lightweight interfaces (Web, CLI, MCP).

## 2. Transport Layer
Vectora prioritizes local, zero-port communication:

- **Windows**: Uses **Named Pipes** (`\\.\pipe\vectora`) with an `AF_UNIX` (Unix Sockets) wrapper for Windows 10/11. Fallback to `127.0.0.1:42780` (TCP loopback) if pipes fail.
- **Linux / macOS**: Uses native **Unix Domain Sockets** located in the user's AppData directory (`run/vectora.sock`).

## 3. Protocol: JSON-ND
To simplify streaming and avoid heavy headers, we use **Newline-Delimited JSON**.

### 3.1 Message Envelope
Every frame sent through the socket must follow this JSON structure:
```json
{
  "id": "uuid-v4",
  "type": "request | response | event",
  "method": "workspace.query",
  "payload": { ... },
  "error": null
}
```

## 4. Operational Rules (RN-IPC)
1. **Size Limit**: Frames are limited to **4MB** to protect daemon memory.
2. **Master-Slave**: The Daemon acts as the Master, consuming only `request` types from clients and emitting `response` or `event`.
3. **Graceful Failure**: Disconnected clients are automatically reaped from the `clients` map.
4. **Broadcast**: The server supports a broadcast mode to push real-time updates (like indexing progress) to all connected UIs.
