"""Testes unitários para validação e segurança de tools (tool_safety.py)."""

import tempfile
from pathlib import Path

from vectora.services.security import (
    is_safe_file_path,
    is_safe_regex_pattern,
    is_safe_shell_command,
)


class TestFilePathValidation:
    """Testes para validação de caminhos de arquivo."""

    def test_validate_safe_file_path(self):
        """Verificar que caminhos seguros são aceitos."""
        with tempfile.TemporaryDirectory() as tmpdir:
            safe_path = Path(tmpdir) / "safe.txt"
            result = is_safe_file_path(str(safe_path))
            assert result is True

    def test_validate_file_path_prevents_traversal(self):
        """Verificar que traversal de diretório é bloqueado."""
        malicious_path = "../../../../etc/passwd"
        result = is_safe_file_path(malicious_path)
        assert result is False

    def test_validate_file_path_with_parent_references(self):
        """Verificar que .. em paths é bloqueado."""
        malicious_path = "/var/www/../../etc/passwd"
        result = is_safe_file_path(malicious_path)
        assert result is False

    def test_validate_file_path_with_relative_safe_path(self):
        """Verificar que caminhos relativos seguros são aceitos."""
        safe_path = "subdir/file.txt"
        result = is_safe_file_path(safe_path)
        assert result is True

    def test_validate_dangerous_extensions(self):
        """Verificar que extensões perigosas são bloqueadas."""
        dangerous_paths = ["script.exe", "command.bat", "program.sh"]
        for path in dangerous_paths:
            result = is_safe_file_path(path)
            assert result is False


class TestCommandValidation:
    """Testes para validação de comandos shell."""

    def test_validate_safe_command(self):
        """Verificar que comandos seguros são aceitos."""
        safe_commands = ["ls", "pwd", "cat file.txt", "grep pattern file"]
        for cmd in safe_commands:
            result = is_safe_shell_command(cmd)
            assert result is True

    def test_validate_command_blocks_dangerous_commands(self):
        """Verificar que comandos perigosos são bloqueados."""
        dangerous_commands = [
            "rm -rf /",
            "dd if=/dev/zero of=/dev/sda",
            "mkfs.ext4",
        ]
        for cmd in dangerous_commands:
            result = is_safe_shell_command(cmd)
            assert result is False

    def test_validate_command_allows_unknown_binaries(self):
        """Verificar que binários desconhecidos são permitidos (política blacklist-only).

        A implementação usa blacklist — PERMITE tudo exceto comandos destrutivos.
        Binários desconhecidos são permitidos pois não estão na blacklist.
        """
        unknown_commands = ["random_unknown_command", "custom_tool"]
        for cmd in unknown_commands:
            result = is_safe_shell_command(cmd)
            assert result is True

    def test_validate_safe_npm_command(self):
        """Verificar que npm commands são aceitos."""
        npm_commands = ["npm install", "npm test", "npm run build"]
        for cmd in npm_commands:
            result = is_safe_shell_command(cmd)
            assert result is True

    def test_validate_safe_git_command(self):
        """Verificar que git commands são aceitos."""
        git_commands = ["git status", "git log", "git pull"]
        for cmd in git_commands:
            result = is_safe_shell_command(cmd)
            assert result is True


class TestPatternValidation:
    """Testes para validação de padrões regex."""

    def test_validate_safe_regex_pattern(self):
        """Verificar que padrões seguros são aceitos."""
        safe_patterns = [
            r"^hello$",
            r"\d+",
            r"[a-zA-Z]+",
            r"(foo|bar)",
        ]
        for pattern in safe_patterns:
            result = is_safe_regex_pattern(pattern)
            assert result is True

    def test_validate_pattern_prevents_redos(self):
        """Verificar que padrões ReDoS são bloqueados."""
        redos_patterns = [
            r"(a+)+$",
            r"(a|a)*$",
            r"(.*)*$",
        ]
        for pattern in redos_patterns:
            result = is_safe_regex_pattern(pattern)
            assert result is False

    def test_validate_pattern_with_special_chars(self):
        """Verificar que caracteres especiais são escapados corretamente."""
        special_patterns = [
            r"\.",
            r"\*",
            r"\?",
            r"\[",
        ]
        for pattern in special_patterns:
            result = is_safe_regex_pattern(pattern)
            assert result is True

    def test_validate_invalid_regex(self):
        """Verificar que regex inválido é rejeitado."""
        invalid_patterns = [r"[", r"(unclosed", r"(?P<invalid)"]
        for pattern in invalid_patterns:
            result = is_safe_regex_pattern(pattern)
            assert result is False
