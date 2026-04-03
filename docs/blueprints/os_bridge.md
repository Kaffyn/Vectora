# Blueprint: OS Bridge & Native Integration

**Status:** Implementation Complete / Industrial Grade
**Module:** `internal/os/`

---

## 1. Overview
The OS Bridge handles the lowest-level interactions between the Vectora application and the host operating system. It ensures that the application behaves correctly in a multi-process environment and follows platform-specific standards.

## 2. Infrastructure Patterns
The `vecos.Manager` provides a unified interface for the following operations:

### 2.1 AppData Management
Vectora stores all persistent data in the official system application data directory:
- **Windows**: `%LOCALAPPDATA%\Vectora`
- **Unix**: `~/.config/vectora` or `~/.local/share/vectora`
- **Sub-folders**:
  - `data/`: DB files (bbolt, chromem).
  - `run/`: Active sockets and lock files.
  - `backups/`: File snapshots for GitBridge/Undo.
  - `engines/`: Llama binaries and library assets.

### 2.2 Single Instance Enforcement
To prevent data corruption in the bbolt database, Vectora enforces a single-instance rule:
- **Implementation**: Uses a global system Mutex (Windows) or a File Lock on `.vectlock` (Unix).
- **Behavior**: If a second instance tries to launch as `daemon`, it detects the lock, notifies the user, and immediately exits.

### 2.3 Local Notifications
Vectora uses native system notifications to alert the user of background events (Indexing finished, Agent actions, Errors).
- **Windows**: Desktop Toast notifications.
- **Unix**: `libnotify` / `dbus`.

## 3. Deployment Safety (BR-OS)
- **Path Sanitization**: All incoming paths from tools are sanitized to prevent directory traversal attacks.
- **Permission Management**: Default creation perms for folders are `0755` and files `0644`.
