"""Módulo de segurança para execução de ferramentas com whitelisting."""

import re
from pathlib import Path


def is_safe_file_path(path: str, allowed_dirs: list[str] | None = None) -> bool:
    """Verifica se um caminho de arquivo é seguro para leitura/edição.

    Rejeita:
    - Caminhos absolutoscom /../
    - Arquivos fora de allowed_dirs (se especificado)
    - Extensões perigosas

    Args:
        path: Caminho do arquivo
        allowed_dirs: Diretórios permitidos (ex: ["./src", "./data"])

    Returns:
        True se caminho é seguro
    """
    try:
        file_path = Path(path).resolve()

        if ".." in path:
            return False

        dangerous_extensions = {".exe", ".sh", ".bat", ".cmd", ".com", ".pif"}
        if file_path.suffix.lower() in dangerous_extensions:
            return False

        if allowed_dirs:
            for allowed_dir in allowed_dirs:
                allowed_path = Path(allowed_dir).resolve()
                try:
                    file_path.relative_to(allowed_path)
                    return True
                except ValueError:
                    pass
            return False

        return True
    except (ValueError, OSError):
        return False


def is_safe_regex_pattern(pattern: str) -> bool:
    """Valida se um padrão regex é seguro (evita ReDoS).

    Args:
        pattern: Padrão regex

    Returns:
        True se padrão é válido
    """
    dangerous_patterns = [
        r"(.*)*",
        r"(.*)+",
        r"(.+)*",
        r"(.+)+",
        r"(a*)*",
        r"(a+)*",
        r"(a*)+",
        r"(a+)+",
    ]

    if any(dangerous in pattern for dangerous in dangerous_patterns):
        return False

    try:
        re.compile(pattern)
        return True
    except re.error:
        return False


def is_safe_shell_command(command: str) -> bool:
    """Valida se um comando shell é seguro (whitelist de comandos).

    Allowed: git, python, npm, node, etc
    Blocked: rm -rf, dd, etc

    Args:
        command: Comando a validar

    Returns:
        True se comando é permitido
    """
    allowed_commands = {
        "python",
        "node",
        "npm",
        "git",
        "ls",
        "pwd",
        "echo",
        "cat",
        "grep",
        "find",
        "tail",
        "head",
        "wc",
        "sort",
        "uniq",
        "cut",
    }

    forbidden_patterns = {
        "rm ",
        "dd ",
        "mkfs",
        "format",
        "erase",
        "del ",
        "unlink",
    }

    first_word = command.split()[0] if command.split() else ""

    if first_word not in allowed_commands:
        return False

    return all(forbidden not in command for forbidden in forbidden_patterns)
