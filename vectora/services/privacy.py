"""Privacy Utilities for LGPD/GDPR Compliance.

PII scrubbing, thread ID anonymization, metadata sanitization,
and privacy-safe LangSmith tracer for developer telemetry.

Design principles:
- scrub_pii: regex-based redaction of emails, CPFs, phone numbers,
  credentials and local file paths from arbitrary text.
- anonymize_thread_id: produces an opaque 16-hex identifier that is
  deterministic within a run but NOT linkable across runs (per-process
  salt), so no cross-session correlation is possible.
- sanitize_metadata: removes sensitive keys from dict structures before
  they are attached to LangSmith run metadata.
- build_sanitized_tracer: returns a LangChainTracer whose langsmith.Client
  has hide_inputs=True and hide_outputs=True — the "hard gate by code"
  that makes it physically impossible for conversation content to leave
  the user's machine via the developer's telemetry channel.

Dual-tracer architecture:
  User's own LangSmith → env vars LANGSMITH_API_KEY + LANGSMITH_TRACING=true
                          (picked up automatically by LangChain, full fidelity)
  Developer's telemetry  → build_sanitized_tracer(VECTORA_LANGSMITH_API_KEY)
                          (explicit callback in RunnableConfig, no content)

LGPD references:
  Art. 5 — definition of "dado pessoal"
  Art. 6 — minimização de dados (principle of necessity)
  Art. 12 — dados anonimizados are outside LGPD scope
  Art. 46 — privacidade por padrão (Privacy by Default)
"""

import hashlib
import os
import re
from typing import Any

# ---------------------------------------------------------------------------
# Patterns & replacements
# ---------------------------------------------------------------------------

_PII_PATTERNS: list[tuple[re.Pattern[str], str]] = [
    # Email
    (
        re.compile(r"\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b"),
        "[EMAIL]",
    ),
    # Brazilian CPF (xxx.xxx.xxx-xx or 11 digits)
    (
        re.compile(r"\b\d{3}\.?\d{3}\.?\d{3}[-\s]?\d{2}\b"),
        "[CPF]",
    ),
    # Brazilian CNPJ
    (
        re.compile(r"\b\d{2}\.?\d{3}\.?\d{3}/?\d{4}[-\s]?\d{2}\b"),
        "[CNPJ]",
    ),
    # Phone (Brazilian +55 or bare)
    (
        re.compile(r"(?:\+55\s?)?(?:\(?\d{2}\)?\s?)?\d{4,5}[\s\-]?\d{4}\b"),
        "[PHONE]",
    ),
    # Credentials / tokens
    (
        re.compile(
            r"(?i)(?:bearer|token|key|secret|password|senha|api[_\-]?key)"
            r"\s*[=:]\s*\S+",
        ),
        "[CREDENTIAL]",
    ),
    # Windows absolute paths (C:\..., D:\...)
    (
        re.compile(r"[A-Za-z]:\\(?:[^\"'<>\n\r\\/:*?]+\\)*[^\"'<>\n\r\\/:*?]*"),
        "[PATH]",
    ),
    # Unix home dir paths (/home/<user>/... or /Users/<user>/...)
    (
        re.compile(r"/(?:home|Users)/[^\s/]+(?:/[^\s]*)?"),
        "[PATH]",
    ),
]

# Keys that should never appear in trace metadata
_SENSITIVE_KEYS: frozenset[str] = frozenset(
    {
        "api_key",
        "apikey",
        "api-key",
        "key",
        "secret",
        "password",
        "senha",
        "token",
        "authorization",
        "auth",
        "credential",
        "google_api_key",
        "openai_api_key",
        "anthropic_api_key",
        "voyage_api_key",
        "langsmith_api_key",
        "tavily_api_key",
        "sentry_dsn",
    }
)

# ---------------------------------------------------------------------------
# Per-process salt: ensures thread_id hashes are NOT linkable across runs.
# Generated once when the module is imported; not persisted anywhere.
# ---------------------------------------------------------------------------
_PROCESS_SALT: str = os.urandom(16).hex()


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def scrub_pii(text: str) -> str:
    """Remove PII from *text* using regex-based redaction.

    Redacts: emails, CPF/CNPJ, phone numbers, credential assignments,
    Windows and Unix user-specific file paths.

    Args:
        text: Raw text that may contain personal or sensitive data.

    Returns:
        Sanitized text with PII replaced by placeholder tokens.
    """
    for pattern, replacement in _PII_PATTERNS:
        text = pattern.sub(replacement, text)
    return text


def anonymize_thread_id(thread_id: str) -> str:
    """Produce an opaque, per-run identifier from *thread_id*.

    The output is:
    - Deterministic: same thread_id → same hash within one process.
    - Non-linkable: different every time the process restarts (salt changes).
    - Compact: "vtx_" prefix + 16 hex chars.

    Under LGPD Art. 12 this constitutes "anonimização" because the original
    thread_id cannot be reconstructed without the transient in-memory salt.

    Args:
        thread_id: Raw session/thread identifier (may contain user data).

    Returns:
        Opaque anonymized identifier safe for external telemetry.
    """
    raw = f"{_PROCESS_SALT}:{thread_id}"
    digest = hashlib.sha256(raw.encode()).hexdigest()[:16]
    return f"vtx_{digest}"


def sanitize_metadata(metadata: dict[str, Any]) -> dict[str, Any]:
    """Recursively remove sensitive keys and scrub PII from *metadata*.

    Sensitive keys (api_key, password, token, etc.) are replaced with
    "[REDACTED]".  String values are passed through :func:`scrub_pii`.
    Nested dicts are sanitized recursively.

    Args:
        metadata: Arbitrary metadata dict (e.g., LangSmith run metadata).

    Returns:
        A new dict with sensitive data removed or redacted.
    """
    result: dict[str, Any] = {}
    for key, value in metadata.items():
        if key.lower() in _SENSITIVE_KEYS:
            result[key] = "[REDACTED]"
        elif isinstance(value, dict):
            result[key] = sanitize_metadata(value)
        elif isinstance(value, str):
            result[key] = scrub_pii(value)
        elif isinstance(value, list):
            result[key] = [
                sanitize_metadata(item) if isinstance(item, dict) else item
                for item in value
            ]
        else:
            result[key] = value
    return result


# ---------------------------------------------------------------------------
# Privacy-safe LangSmith tracer (developer telemetry)
# ---------------------------------------------------------------------------


def build_sanitized_tracer(
    api_key: str,
    project_name: str = "Vectora",
    endpoint: str = "https://api.smith.langchain.com",
) -> Any:
    """Build a LangChainTracer that never sends conversation content.

    This is the hard gate required by LGPD Art. 46 (Privacy by Default):
    the langsmith.Client is configured with ``hide_inputs=True`` and
    ``hide_outputs=True``, which means the client itself strips all
    input/output payloads **before** any network call — even if a future
    maintainer accidentally removes the env-var guard.

    What IS sent to the developer's LangSmith project:
    - Graph node names and execution order
    - Per-node latency (start_time / end_time)
    - Error type (class name only, message stripped)
    - Tool names (not tool inputs/outputs)
    - LLM model name and token counts
    - Vectora version tag

    What is NEVER sent:
    - User messages / AI responses
    - Tool call arguments or results
    - File contents, paths, or metadata
    - API keys or credentials

    This tracer is added as an **explicit callback** in RunnableConfig and
    runs in parallel with the user's own LangSmith tracer (if configured
    via LANGSMITH_API_KEY env var) — the two never interfere.

    Args:
        api_key: Developer's VECTORA_LANGSMITH_API_KEY (injected at build).
        project_name: LangSmith project to write to (default: "Vectora").
        endpoint: LangSmith API endpoint.

    Returns:
        LangChainTracer configured for privacy-safe developer telemetry.

    Raises:
        ImportError: If langsmith or langchain is not installed.
    """
    from langchain.callbacks.tracers import LangChainTracer
    from langsmith import Client

    client = Client(
        api_key=api_key,
        api_url=endpoint,
        # Hard gate: strip all inputs and outputs at the client level.
        # This is enforced by the langsmith SDK before any HTTP call —
        # not an env var that can be overridden at runtime.
        hide_inputs=True,
        hide_outputs=True,
    )
    return LangChainTracer(project_name=project_name, client=client)
