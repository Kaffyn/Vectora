"""Telemetry Consent Management — LGPD/GDPR Compliance.

Manages user opt-in/opt-out decisions for sending anonymized operational
telemetry (LangSmith traces) to Vectora's developer dashboard.

Legal basis: LGPD Art. 7, inciso I — consentimento do titular.

Consent lifecycle:
  1. First run → ConsentManager.has_answered() == False
  2. Setup wizard shows explicit prompt (separate from ToS, specific purpose)
  3. User answers → set_consent(True|False)
  4. Decision persisted in ~/.vectora/consent.json  (audit trail)
  5. User can revoke at any time → /privacidade disable
     or revoke() → re-serialized immediately, LangSmith disabled at next
     startup (or immediately, since main.py re-reads at startup).

LGPD references:
  Art. 7 — bases legais para tratamento
  Art. 8 — requisitos do consentimento (gratuito, informado, inequívoco)
  Art. 18 — direitos do titular (revogação a qualquer tempo)
"""

import json
import logging
from datetime import UTC, datetime
from pathlib import Path

logger = logging.getLogger(__name__)

# Default location — can be overridden in tests
CONSENT_FILE: Path = Path.home() / ".vectora" / "consent.json"


class ConsentManager:
    """Read/write user telemetry consent with LGPD-compliant audit trail.

    All decisions are persisted in a JSON file so the developer can
    demonstrate compliance if ever audited by ANPD.

    Attributes:
        consent_file: Path to the JSON file storing consent decisions.
    """

    def __init__(self, consent_file: Path | None = None) -> None:
        """Initialize manager.

        Args:
            consent_file: Optional override for consent file location
                (useful in tests).
        """
        self.consent_file: Path = consent_file or CONSENT_FILE

    # ------------------------------------------------------------------
    # Read helpers
    # ------------------------------------------------------------------

    def _load(self) -> dict:
        """Load raw consent JSON, returning empty dict on any error."""
        try:
            if self.consent_file.exists():
                return json.loads(self.consent_file.read_text(encoding="utf-8"))
        except Exception:
            logger.debug("consent_file_unreadable", exc_info=True)
        return {}

    def has_answered(self) -> bool:
        """Return True if the user has already made a consent decision.

        Returns:
            False when "telemetry" key is missing or "not_asked".
        """
        data = self._load()
        return data.get("telemetry", "not_asked") not in ("not_asked", "")

    def is_consented(self) -> bool:
        """Return True only if the user explicitly opted in.

        Returns:
            True when telemetry == "true", False for anything else
            (including "not_asked").
        """
        return self._load().get("telemetry") == "true"

    def get_status(self) -> str:
        """Return human-readable consent status string.

        Returns:
            One of: "not_asked" | "consented" | "declined"
        """
        value = self._load().get("telemetry", "not_asked")
        if value == "true":
            return "consented"
        if value == "false":
            return "declined"
        return "not_asked"

    # ------------------------------------------------------------------
    # Write helpers
    # ------------------------------------------------------------------

    def set_consent(self, value: bool, vectora_version: str = "unknown") -> None:
        """Persist a consent decision with a timestamped audit entry.

        Creates or updates ~/.vectora/consent.json.  If the file already
        contains a decision it is updated (to support re-consent after
        policy changes).

        Args:
            value: True = opt-in, False = opt-out.
            vectora_version: Version string stored for audit purposes.
        """
        data = self._load()

        # First-time: record when consent was first requested
        if "asked_at" not in data:
            data["asked_at"] = datetime.now(UTC).isoformat()

        data["telemetry"] = "true" if value else "false"
        data["answered_at"] = datetime.now(UTC).isoformat()
        data["vectora_version"] = vectora_version
        # LGPD Art. 8 §5 — identify the legal basis explicitly
        data["legal_basis"] = "consentimento_art7_inciso_I_lgpd"
        data["purpose"] = (
            "telemetria_operacional_anonima: tempo_execucao, "
            "tipo_erro, modelo_llm, versao_vectora"
        )

        self.consent_file.parent.mkdir(parents=True, exist_ok=True)
        self.consent_file.write_text(
            json.dumps(data, indent=2, ensure_ascii=False),
            encoding="utf-8",
        )

        logger.info(
            "telemetry_consent_saved",
            extra={"consent": value, "version": vectora_version},
        )

    def revoke(self, vectora_version: str = "unknown") -> None:
        """Revoke consent (opt-out).  Alias for set_consent(False).

        Args:
            vectora_version: Current Vectora version.
        """
        self.set_consent(False, vectora_version=vectora_version)
        logger.info("telemetry_consent_revoked")


# Module-level singleton for convenient import
_consent_manager: ConsentManager | None = None


def get_consent_manager() -> ConsentManager:
    """Return the global ConsentManager singleton."""
    global _consent_manager
    if _consent_manager is None:
        _consent_manager = ConsentManager()
    return _consent_manager
