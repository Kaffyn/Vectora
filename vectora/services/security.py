"""SecurityService: Security validation and guardrails.

Responsibilities:
1. Validate file operations (prevent dangerous edits)
2. Validate terminal commands (prevent destructive operations)
3. Validate directory access
4. Maintain protected paths and blocked commands lists
5. Log all security violations

Week 2 implementation task: Move from security.py and enhance
"""

import logging
from pathlib import Path

from settings import Settings

logger = logging.getLogger(__name__)

# System-critical directories that should never be modified
PROTECTED_PATHS = {
    # Windows system paths
    r"C:\Windows",
    r"C:\Program Files",
    r"C:\Program Files (x86)",
    r"C:\ProgramData",
    r"C:\System Volume Information",
    r"C:\$Recycle.Bin",
    # Linux/Unix system paths
    "/bin",
    "/sbin",
    "/usr/bin",
    "/usr/sbin",
    "/usr/local/bin",
    "/etc",
    "/boot",
    "/sys",
    "/proc",
    "/dev",
    "/root",
    # macOS system paths
    "/Library/System",
    "/System/Library",
    "/usr/libexec",
}

# System-critical files that should never be edited
PROTECTED_FILES = {
    # Windows critical files
    r"C:\boot.ini",
    r"C:\bootmgr",
    r"C:\ntdetect.com",
    # Linux/Unix critical files
    "/etc/passwd",
    "/etc/shadow",
    "/etc/sudoers",
    "/etc/fstab",
    "/etc/hostname",
    # Shell profiles
    "~/.bashrc",
    "~/.bash_profile",
    "~/.zshrc",
    "~/.profile",
    # SSH keys and config
    "~/.ssh/config",
    "~/.ssh/authorized_keys",
}

# Dangerous terminal commands that should be blocked
BLOCKED_COMMANDS = {
    "rm -rf",
    "rmdir",
    "rm ",
    "mkfs",
    "dd",
    "format",
    "del ",
    "erase",
    "chown",
    "chmod",
    "sudo",
    "su ",
    ":(){:|:&};:",  # Fork bomb
}


class SecurityService:
    """Validates operations before execution.

    Implements guardrails to prevent dangerous operations:
    - File deletion
    - System file modification
    - Privileged command execution
    - Directory traversal attacks

    Design principle: Fail-secure. Errors default to blocking.
    """

    def __init__(self, settings: Settings):
        """Initialize SecurityService.

        Args:
            settings: Application settings
        """
        self.settings = settings
        self.protected_paths = PROTECTED_PATHS
        self.protected_files = PROTECTED_FILES
        self.blocked_commands = BLOCKED_COMMANDS

        logger.debug("SecurityService initialized")

    def validate_file_edit(self, file_path: str) -> tuple[bool, str]:
        """Validate if file edit is allowed.

        Args:
            file_path: Path to file to edit

        Returns:
            Tuple of (is_allowed, reason_message)
        """
        try:
            normalized = self._normalize_path(file_path)
            normalized_str = str(normalized).lower()

            # Check protected files
            for protected_file in self.protected_files:
                protected_normalized = self._normalize_path(protected_file)
                if normalized == protected_normalized:
                    reason = f"[BLOCKED] Cannot edit protected file: {file_path}"
                    logger.warning(reason)
                    return False, reason

            # Check protected directories
            for protected_dir in self.protected_paths:
                try:
                    protected_normalized = self._normalize_path(protected_dir)
                    if normalized.is_relative_to(protected_normalized):
                        reason = f"[BLOCKED] File in protected directory: {file_path}"
                        logger.warning(reason)
                        return False, reason
                except (ValueError, AttributeError):
                    if normalized_str.startswith(str(protected_normalized).lower()):
                        reason = f"[BLOCKED] File in protected directory: {file_path}"
                        logger.warning(reason)
                        return False, reason

            logger.debug(f"File edit allowed: {file_path}")
            return True, f"[OK] Safe to edit: {file_path}"

        except Exception as e:
            logger.exception(f"Error validating file: {e}")
            return False, f"[ERROR] Validation error: {e}"

    def validate_terminal_command(self, command: str) -> tuple[bool, str]:
        """Validate if terminal command is safe to execute.

        Args:
            command: Command string to validate

        Returns:
            Tuple of (is_safe, reason_message)
        """
        command_lower = command.lower().strip()

        # Check for blocked commands
        for blocked in self.blocked_commands:
            if blocked.lower() in command_lower:
                reason = f"[BLOCKED] Dangerous command blocked: '{blocked}' detected"
                logger.warning(f"Command blocked: {command}")
                return False, reason

        # Check for deletion patterns
        if command_lower.startswith(("rm", "del")):
            reason = "[BLOCKED] File deletion commands not allowed for safety"
            logger.warning(f"Deletion command blocked: {command}")
            return False, reason

        logger.debug(f"Command validation passed: {command}")
        return True, f"[OK] Safe to execute: {command}"

    def validate_directory_access(self, directory: str) -> tuple[bool, str]:
        """Validate if directory can be accessed.

        Args:
            directory: Directory path to access

        Returns:
            Tuple of (is_allowed, reason_message)
        """
        try:
            normalized = self._normalize_path(directory)

            if normalized.exists() and not normalized.is_dir():
                reason = f"[ERROR] Path is not a directory: {directory}"
                logger.warning(reason)
                return False, reason

            # Check if it's a protected directory
            for protected_dir in self.protected_paths:
                try:
                    protected_normalized = self._normalize_path(protected_dir)
                    if normalized.is_relative_to(protected_normalized):
                        reason = f"[WARNING] Protected system directory: {directory}"
                        logger.warning(reason)
                        return False, reason
                except (ValueError, AttributeError):
                    if (
                        str(normalized)
                        .lower()
                        .startswith(str(protected_normalized).lower())
                    ):
                        reason = f"[WARNING] Protected system directory: {directory}"
                        logger.warning(reason)
                        return False, reason

            logger.debug(f"Directory access allowed: {directory}")
            return True, f"[OK] Safe to access: {directory}"

        except Exception as e:
            logger.exception(f"Error validating directory: {e}")
            return False, f"[ERROR] Validation failed: {e}"

    @staticmethod
    def _normalize_path(path: str) -> Path:
        """Normalize and expand path variables.

        Args:
            path: Path string

        Returns:
            Normalized, expanded Path object
        """
        return Path(path).expanduser().resolve()

    def log_security_event(
        self,
        event_type: str,
        status: str,
        resource: str,
        reason: str | None = None,
    ) -> None:
        """Log security-relevant events.

        Args:
            event_type: Type of event ("file_edit", "command_exec", "access")
            status: "allowed" or "blocked"
            resource: Path or command being validated
            reason: Reason for block (if blocked)
        """
        logger.warning(
            f"Security event: {event_type}={status}",
            extra={
                "event_type": event_type,
                "status": status,
                "resource": resource,
                "reason": reason,
            },
        )
