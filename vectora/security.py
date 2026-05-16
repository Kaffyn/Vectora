"""Security validation module for Vectora.

Provides guardrails for sensitive operations like file editing and terminal execution.
Prevents accidental modifications to system-critical files and directories.
"""

import logging
from pathlib import Path

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


def _normalize_path(path: str) -> Path:
    """Normalize and expand path variables."""
    return Path(path).expanduser().resolve()


def is_path_protected(file_path: str) -> bool:
    """Check if a path is in a protected directory or is a protected file.

    Args:
        file_path: Path to check

    Returns:
        True if the path is protected, False otherwise
    """
    try:
        normalized = _normalize_path(file_path)
        normalized_str = str(normalized).lower()

        # Check protected files
        for protected_file in PROTECTED_FILES:
            protected_normalized = _normalize_path(protected_file)
            if normalized == protected_normalized:
                return True

        # Check protected directories
        for protected_dir in PROTECTED_PATHS:
            try:
                protected_normalized = _normalize_path(protected_dir)
                # Check if file is under protected directory
                if normalized.is_relative_to(protected_normalized):
                    return True
            except (ValueError, AttributeError):
                # is_relative_to not available in older Python, fallback
                if normalized_str.startswith(str(protected_normalized).lower()):
                    return True

        return False
    except Exception as e:
        logger.warning(f"Error checking path protection: {e}")
        return True  # Err on the side of caution


def validate_file_edit(file_path: str) -> tuple[bool, str]:
    """Validate if a file can be safely edited.

    Args:
        file_path: Path to file to edit

    Returns:
        Tuple of (is_safe, message)
    """
    if is_path_protected(file_path):
        return False, f"[BLOCKED] Cannot edit protected file: {file_path}"

    return True, f"[OK] Safe to edit: {file_path}"


def validate_terminal_command(command: str) -> tuple[bool, str]:
    """Validate if a terminal command is safe to execute.

    Args:
        command: Command to execute

    Returns:
        Tuple of (is_safe, message)
    """
    command_lower = command.lower().strip()

    # Check for blocked commands
    for blocked in BLOCKED_COMMANDS:
        if blocked.lower() in command_lower:
            return (
                False,
                f"[BLOCKED] Dangerous command blocked: '{blocked}' detected in command",
            )

    # Check for suspicious patterns
    if command_lower.startswith(("rm", "del")):
        return False, "[BLOCKED] File deletion commands are not allowed for safety"

    return True, f"[OK] Safe to execute: {command}"


def validate_directory_access(directory: str) -> tuple[bool, str]:
    """Validate if a directory can be accessed.

    Args:
        directory: Directory path to access

    Returns:
        Tuple of (is_safe, message)
    """
    try:
        normalized = _normalize_path(directory)
        if normalized.exists() and not normalized.is_dir():
            return False, f"[ERROR] Path is not a directory: {directory}"

        # Check if it's a protected directory
        if is_path_protected(directory):
            return (
                False,
                f"[WARNING] Protected system directory - cannot access: {directory}",
            )

        return True, f"[OK] Safe to access: {directory}"
    except Exception as e:
        return False, f"[ERROR] Error validating directory: {e}"
