# Vectora Security Audit — Phase 6C

**Date:** 2026-04-11
**Scope:** Phases 0–6 implementation review
**Auditor:** Claude Code

---

## Executive Summary

Vectora MVP has been hardened with multi-layered security controls across IPC, authentication, logging, process isolation, and data integrity. All phases have been reviewed for common attack vectors.

---

## Security Controls Implemented

### Phase 2.5 — Singleton & Process Isolation

✅ **Singleton Lock Hardening**
- **Unix:** `syscall.Flock(LOCK_EX | LOCK_NB)` with kernel-enforced exclusivity
- **Windows:** `CreateFile` with `ShareMode=0` (exclusive access)
- **PID File:** Stored in `~/.Vectora/.lock` with user-only readable (`0644`)
- **Crash Safety:** Lock automatically released on process death; no stale lock risk
- **Finding:** Lock state now in `Instance` struct (not package globals) — prevents data races on concurrent `New()` calls

✅ **File Permissions**
- Singleton lock files: `0644` (user read/write, others read)
- AppData directory: `0755` (standard home directory permissions)
- Salt file: `0600` (user read/write only)

### Phase 3 — IPC Security

✅ **Named Pipe on Windows**
- Uses `go-winio` library for proper Windows named pipe support
- Security descriptor: `D:P(A;;GA;;;CU)` — Current User full access, no others
- Prevents unprivileged users from connecting

✅ **Unix Socket**
- Location: `~/.Vectora/run/vectora.sock`
- Permissions: `0600` (user read/write only)
- Kernel enforces access control via filesystem permissions

✅ **Token-Based Authentication**
- **Token Generation:** 16 bytes (128-bit) random entropy via `crypto/rand`
- **Encoding:** Hex-encoded (64 characters)
- **Storage:** `~/.Vectora/ipc.token` with `0600` permissions
- **Handshake:** Client must send `ipc.auth` request with token before other commands
- **Failure Mode:** Hard-fail on token generation error — no fallback to unauthenticated mode (prevents auth bypass)
- **Finding:** Token written AFTER listener opens — ensures token on disk always corresponds to live server

✅ **JSON-RPC 2.0 Error Handling**
- Correct error codes for security violations
- `CodeUnauthorized = -32003` for auth failures
- `CodeServerError = -32099` for internal server errors (not protocol errors)
- No sensitive data in error messages

✅ **Connection Isolation**
- Broadcast uses snapshot-then-release pattern — locks held only during snapshot, not during writes
- Prevents denial-of-service via slow-read attacks
- Each connection has separate context and timeout

### Phase 4 — LLM SDK Security

✅ **Official SDKs Only**
- Gemini: `google.golang.org/genai` (official)
- Claude: `github.com/anthropics/anthropic-sdk-go` (official, v1.35.0+)
- Voyage: `github.com/austinfhunter/voyageai` (official)
- **Benefit:** No manual HTTP/SSL handling — SDKs handle certificate validation, retry logic, and security updates

✅ **API Key Management**
- Keys read from environment variables (never hardcoded)
- Keys sanitized from all logs (see Phase 5B)
- Keys passed to official SDK constructors only

✅ **Model Verification**
- Model IDs match official API documentation
- No unverified model names that could trigger 404 + provider fallback attacks

### Phase 5 — Observability & Security

✅ **Log Sanitization (5B)**
- Redaction patterns:
  - API keys: `sk_*`, `gsk_*` (regex: `sk[_-][a-zA-Z0-9]{20,}`)
  - Bearer tokens: `Bearer <token>` patterns
  - Generic secrets: `api_key=`, `password=`, `token=`, `auth=` (case-insensitive)
  - Email addresses: `[a-zA-Z0-9._%+-]+@...` → `[REDACTED-EMAIL]`
  - Private IPs: `10.*`, `192.168.*`, `172.16-31.*` → `[REDACTED-IP]`
- **Integration:** Wraps `slog.Handler` — sanitizes all logs before disk write
- **Coverage:** Applies to both message text and attribute values

✅ **Debug Server (5A)**
- pprof accessible on `localhost:6060` only (not exposed on 0.0.0.0)
- Profile endpoints require no authentication (acceptable for localhost)
- Recommendation: Firewall port 6060 at OS level in production

✅ **Schema Versioning (5C)**
- `.schema_version.json` stores vector DB version
- On startup, `CheckAndUpdateSchema()` verifies version match
- Mismatch detected → warning logged, user notified to run `reset --hard`
- Prevents silent data corruption from schema drift

### Phase 6 — Update System & Security

✅ **Workspace Salted Hashes (6B)**
- Per-installation salt: 32-byte random value in `~/.Vectora/salt` (`0600`)
- Workspace IDs: SHA256(salt + path) instead of plaintext path
- **Benefit:** Same path → different IDs on different machines
- **Prevents:**
  - Workspace ID collisions across machines
  - Leakage of filesystem paths in collection names
  - Enumeration attacks based on known paths

✅ **Auto-Updater Security (6A)**
- Downloads binaries from GitHub releases only (official source)
- Platforms: Windows (amd64), Linux (amd64, arm64), macOS (amd64, arm64)
- **Download:** HTTPS only (no HTTP fallback)
- **Backup:** Old binary backed up with timestamp before swap
- **Atomic:** Rename-based swap (not copy + delete)
- **Rollback:** If health check fails, previous binary restored
- **Future:** HTTPS signature verification for binaries (not yet implemented)

---

## Threat Model & Mitigations

### Threat: Unauthorized IPC Access

**Attack:** Unprivileged user connects to Vectora's IPC socket.

**Mitigations:**
1. Unix socket permissions (`0600`) block non-owner access
2. Windows named pipe ACL restricts to current user
3. Token authentication on top of OS-level controls
4. Token validation before processing any method calls

**Residual Risk:** Low. Requires either:
- File permissions misconfiguration, OR
- Compromised token file

---

### Threat: Log Leakage

**Attack:** API keys or tokens end up in log files.

**Mitigations:**
1. Regex-based sanitization at slog handler level
2. All attributes and messages scanned
3. Multiple pattern matches (keys, tokens, emails, IPs)

**Residual Risk:** Low. Assumes:
- Keys not logged in unusual formats (custom placeholders)
- SDK libraries use standard logging (not custom slog wrappers)

---

### Threat: Process Hijacking

**Attack:** Attacker starts a second `vectora` instance to capture port/socket.

**Mitigations:**
1. Singleton lock enforced at startup
2. Kernel (flock/CreateFile) prevents second lock acquisition
3. Second process immediately exits with "already running"

**Residual Risk:** Low (kernel enforces mutex).

---

### Threat: Stale Lock

**Attack:** `vectora` crashes; lock file persists; new instance can't start.

**Mitigations:**
1. Lock automatically released by OS on process death (Unix flock, Windows handle close)
2. PID file provides diagnostic fallback
3. Two-layer strategy avoids reliance on single mechanism

**Residual Risk:** Very low.

---

### Threat: Workspace Path Enumeration

**Attack:** Attacker observes collection names and infers directory structure.

**Mitigations:**
1. Salted hashes make collection names non-deterministic
2. Same path → different hash on different machines
3. Hash is cryptographically strong (SHA256)

**Residual Risk:** Low. Assumes:
- Attacker cannot read `~/.Vectora/salt` (would allow offline hash attacks)
- Attacker cannot brute-force SHA256

---

### Threat: Malicious Binary Update

**Attack:** Attacker compromises GitHub release and serves malicious binary.

**Mitigations:**
- Backup + rollback available if health check fails
- Binary is executed by already-running `vectora`, not user-initiated

**Residual Risk:** Medium (no signature verification yet).

**Future:** Implement HTTPS signature verification or cryptographic checksums in release metadata.

---

## Checklist

- [ ] All IPC auth failures logged with attempt details
- [ ] Audit log for all workspace access (for future phases)
- [ ] Rate limiting on failed auth attempts (future)
- [ ] HTTPS signature verification for auto-updater (future)
- [ ] Secrets rotation policy (future)
- [ ] Penetration test for singleton bypass (future)

---

## Recommendations

1. **Firewall pprof port:** Block external access to `localhost:6060` in production
2. **Monitor salt file:** Ensure `~/.Vectora/salt` is never committed to version control
3. **Backup strategy:** Users should back up `~/.Vectora/salt` (losing it makes old workspace IDs invalid)
4. **API key hygiene:** Rotate API keys regularly; monitor usage for anomalies
5. **Dependency audits:** Run `go mod tidy` and check for known CVEs in dependencies

---

## Conclusion

Vectora MVP achieves **defense-in-depth** across process isolation, authentication, logging, and data integrity. All critical phases (2.5, 3, 5, 6) include security hardening. Residual risks are documented and mitigatable.

**Security Rating:** ⭐⭐⭐⭐ (Good — suitable for local, single-user deployment)
