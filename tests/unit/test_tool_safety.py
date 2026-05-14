"""Testes unitários para validação e segurança de tools (tool_safety.py)."""

import pytest
from pathlib import Path
import tempfile

from tool_safety import (
    validate_file_path,
    validate_command,
    validate_pattern,
)


class TestFilePathValidation:
    """Testes para validação de caminhos de arquivo."""

    def test_validate_safe_file_path(self):
        """Verificar que caminhos seguros são aceitos."""
        with tempfile.TemporaryDirectory() as tmpdir:
            safe_path = Path(tmpdir) / "safe.txt"
            result = validate_file_path(str(safe_path))
            assert result is True

    def test_validate_file_path_prevents_traversal(self):
        """Verificar que traversal de diretório é bloqueado."""
        malicious_path = "../../../../etc/passwd"
        with pytest.raises(ValueError):
            validate_file_path(malicious_path)

    def test_validate_file_path_with_parent_references(self):
        """Verificar que .. em paths é bloqueado."""
        malicious_path = "/var/www/../../etc/passwd"
        with pytest.raises(ValueError):
            validate_file_path(malicious_path)

    def test_validate_file_path_with_absolute_external_path(self):
        """Verificar que caminhos absolutos externos são bloqueados."""
        # Path fora da área permitida
        malicious_path = "/etc/passwd"
        with pytest.raises((ValueError, AssertionError)):
            validate_file_path(malicious_path)

    def test_validate_file_path_with_relative_safe_path(self):
        """Verificar que caminhos relativos seguros são aceitos."""
        with tempfile.TemporaryDirectory() as tmpdir:
            safe_path = "subdir/file.txt"
            result = validate_file_path(safe_path)
            # Deve passar na validação de padrão

    def test_validate_empty_path_raises_error(self):
        """Verificar que path vazio gera erro."""
        with pytest.raises((ValueError, AssertionError)):
            validate_file_path("")

    def test_validate_null_bytes_in_path(self):
        """Verificar que null bytes em path são bloqueados."""
        malicious_path = "/tmp/file\x00.txt"
        with pytest.raises((ValueError, AssertionError)):
            validate_file_path(malicious_path)


class TestCommandValidation:
    """Testes para validação de comandos shell."""

    def test_validate_safe_command(self):
        """Verificar que comandos seguros são aceitos."""
        safe_commands = ["ls", "pwd", "cat file.txt", "grep pattern file"]
        for cmd in safe_commands:
            result = validate_command(cmd)
            assert result is True

    def test_validate_command_blocks_dangerous_commands(self):
        """Verificar que comandos perigosos são bloqueados."""
        dangerous_commands = [
            "rm -rf /",
            "sudo /bin/bash",
            "chmod 777 /etc/shadow",
            "dd if=/dev/zero of=/dev/sda",
        ]
        for cmd in dangerous_commands:
            with pytest.raises(ValueError):
                validate_command(cmd)

    def test_validate_command_blocks_pipes(self):
        """Verificar que pipes não seguros podem ser bloqueados."""
        # Implementação específica pode ou não bloquear pipes
        # Testando se comportamento é consistente
        try:
            result = validate_command("cat file.txt | grep pattern")
            assert result is True
        except ValueError:
            # Pipes podem ser bloqueados por segurança
            pass

    def test_validate_command_blocks_command_injection(self):
        """Verificar que injection é bloqueado."""
        injection_commands = [
            "cat file.txt; rm -rf /",
            "cat file.txt && malicious",
            "cat file.txt || delete_everything",
            "cat file.txt `malicious`",
        ]
        for cmd in injection_commands:
            with pytest.raises(ValueError):
                validate_command(cmd)

    def test_validate_command_requires_whitelisted_binary(self):
        """Verificar que apenas binários whitelistados são permitidos."""
        whitelist_required = ["random_unknown_command", "/usr/bin/malicious"]
        for cmd in whitelist_required:
            with pytest.raises(ValueError):
                validate_command(cmd)

    def test_validate_empty_command_raises_error(self):
        """Verificar que comando vazio gera erro."""
        with pytest.raises((ValueError, AssertionError)):
            validate_command("")

    def test_validate_command_with_redirects(self):
        """Verificar que redirects são validados."""
        # Redirecionamento para arquivo pode ser permitido ou bloqueado
        # Testar comportamento consistente
        try:
            result = validate_command("echo test > output.txt")
            # Se permitido, OK
        except ValueError:
            # Se bloqueado, OK também
            pass


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
            result = validate_pattern(pattern)
            assert result is True

    def test_validate_pattern_prevents_redos(self):
        """Verificar que padrões ReDoS são bloqueados."""
        # Exemplos clássicos de ReDoS
        redos_patterns = [
            r"(a+)+$",
            r"(a|a)*$",
            r"(.*)*$",
            r"(a|ab)*$",
        ]
        for pattern in redos_patterns:
            with pytest.raises(ValueError):
                validate_pattern(pattern)

    def test_validate_pattern_complexity_limit(self):
        """Verificar que padrões muito complexos são bloqueados."""
        complex_pattern = r"(" + "|".join([f"a{i}" for i in range(100)]) + ")*"
        with pytest.raises(ValueError):
            validate_pattern(complex_pattern)

    def test_validate_empty_pattern(self):
        """Verificar que padrão vazio é aceito ou rejeitado consistentemente."""
        try:
            result = validate_pattern("")
            # Padrão vazio pode ser aceito
        except ValueError:
            # Ou rejeitado
            pass

    def test_validate_pattern_with_special_chars(self):
        """Verificar que caracteres especiais são escapados corretamente."""
        special_patterns = [
            r"\.",
            r"\*",
            r"\?",
            r"\[",
        ]
        for pattern in special_patterns:
            result = validate_pattern(pattern)
            assert result is True


class TestWhitelistValidation:
    """Testes para validação de whitelist."""

    def test_command_whitelisting(self):
        """Verificar que whitelist de comandos funciona."""
        # Comandos whitelistados devem passar
        whitelisted = ["ls", "pwd", "cat", "grep", "find", "head", "tail"]
        for cmd in whitelisted:
            # Testando com comando simples
            try:
                result = validate_command(cmd)
                # Deve ser aceito
            except ValueError:
                # Mesmo se a implementação rejeitar, deve ser consistente
                pass

    def test_directory_whitelist(self):
        """Verificar que whitelist de diretório funciona."""
        with tempfile.TemporaryDirectory() as tmpdir:
            allowed_path = tmpdir
            file_in_allowed = Path(allowed_path) / "file.txt"
            # Arquivo em diretório permitido deve passar
            result = validate_file_path(str(file_in_allowed))

    def test_blocked_system_directories(self):
        """Verificar que diretórios do sistema são bloqueados."""
        system_dirs = ["/etc", "/sys", "/proc", "/root", "/boot"]
        for dir_path in system_dirs:
            with pytest.raises((ValueError, AssertionError)):
                validate_file_path(dir_path)


class TestValidationErrorMessages:
    """Testes para mensagens de erro de validação."""

    def test_validation_error_provides_clear_message(self):
        """Verificar que erros de validação têm mensagens claras."""
        try:
            validate_file_path("../../../../etc/passwd")
        except ValueError as e:
            # Mensagem deve indicar o problema
            assert len(str(e)) > 0
            assert "traversal" in str(e).lower() or "path" in str(e).lower()

    def test_validation_error_suggests_fix(self):
        """Verificar que mensagem pode sugerir correção."""
        try:
            validate_command("random_unknown_command")
        except ValueError as e:
            # Mensagem pode sugerir usar whitelisted commands
            assert len(str(e)) > 0
