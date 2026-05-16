"""Testes para IgnorePatternValidator."""

import tempfile
from pathlib import Path

import pytest

from vectora.ignore_validator import IgnorePatternValidator


@pytest.fixture
def temp_dir():
    """Create temporary directory for testing."""
    with tempfile.TemporaryDirectory() as tmpdir:
        yield Path(tmpdir)


def test_builtin_patterns(temp_dir):
    """Testa que padrões embutidos ignoram paths sensíveis."""
    validator = IgnorePatternValidator(temp_dir)

    # Deve ignorar node_modules
    assert validator.should_ignore(Path("node_modules/package/index.js"))
    assert validator.should_ignore("node_modules/package/index.js")

    # Deve ignorar venv
    assert validator.should_ignore(Path(".venv/lib/python3.13"))
    assert validator.should_ignore("venv/bin/activate")

    # Deve ignorar __pycache__
    assert validator.should_ignore("__pycache__/module.pyc")

    # Deve ignorar .env e secrets
    assert validator.should_ignore(".env")
    assert validator.should_ignore(".env.local")
    assert validator.should_ignore(".env.production")

    # Deve ignorar chaves privadas
    assert validator.should_ignore("id_rsa")
    assert validator.should_ignore(".ssh/id_ed25519")

    # Deve ignorar credentials
    assert validator.should_ignore("aws_credentials")
    assert validator.should_ignore("credentials.json")

    print("[OK] Builtin patterns working correctly")


def test_gitignore_file(temp_dir):
    """Testa carregamento de .gitignore."""
    # Criar .gitignore
    gitignore = temp_dir / ".gitignore"
    gitignore.write_text(
        """\
# Python
*.pyc
__pycache__/

# Custom patterns
build/
dist/
.mypy_cache/

# Comments should be ignored
"""
    )

    validator = IgnorePatternValidator(temp_dir)

    # Deve ignorar padrões do .gitignore
    assert validator.should_ignore("build/output.txt")
    assert validator.should_ignore("dist/package.tar.gz")
    assert validator.should_ignore(".mypy_cache/module.json")

    print("[OK] .gitignore patterns loaded correctly")


def test_npmignore_file(temp_dir):
    """Testa carregamento de .npmignore."""
    npmignore = temp_dir / ".npmignore"
    npmignore.write_text(
        """\
node_modules/
.git/
coverage/
src/
tests/
"""
    )

    validator = IgnorePatternValidator(temp_dir)

    # Deve ignorar padrões do .npmignore
    assert validator.should_ignore("node_modules/package/index.js")
    assert validator.should_ignore(".git/objects/ab/1234567890")
    assert validator.should_ignore("coverage/report.html")

    print("[OK] .npmignore patterns loaded correctly")


def test_filter_files(temp_dir):
    """Testa filtragem de lista de arquivos."""
    validator = IgnorePatternValidator(temp_dir)

    files = [
        "src/main.py",
        "node_modules/package/index.js",
        "tests/test_main.py",
        ".env",
        "build/app.exe",
        "README.md",
    ]

    filtered = validator.filter_files(files)
    filtered_names = [f.name for f in filtered]

    # Deve manter apenas arquivos válidos
    assert "main.py" in filtered_names
    assert "test_main.py" in filtered_names
    assert "README.md" in filtered_names

    # Deve remover ignorados
    assert "index.js" not in filtered_names
    assert ".env" not in str(filtered)
    assert "app.exe" not in filtered_names

    print("[OK] File filtering working correctly")


def test_no_false_positives(temp_dir):
    """Testa que não ignora arquivos legítimos."""
    validator = IgnorePatternValidator(temp_dir)

    # Não deveria ignorar esses
    assert not validator.should_ignore("src/app.py")
    assert not validator.should_ignore("src/config/settings.py")
    assert not validator.should_ignore("tests/unit/test_api.py")
    assert not validator.should_ignore("docs/README.md")
    assert not validator.should_ignore("LICENSE")

    print("[OK] No false positives")


def test_ignore_summary(temp_dir):
    """Testa summary de padrões carregados."""
    validator = IgnorePatternValidator(temp_dir)
    summary = validator.get_ignored_patterns_summary()

    assert "builtin_patterns" in summary
    assert "total_patterns" in summary
    assert summary["builtin_patterns"] > 0
    assert summary["total_patterns"] >= summary["builtin_patterns"]

    print(f"[OK] Loaded {summary['total_patterns']} patterns total")


if __name__ == "__main__":
    # Run tests manually
    with tempfile.TemporaryDirectory() as tmpdir:
        tmpdir = Path(tmpdir)
        test_builtin_patterns(tmpdir)
        test_gitignore_file(tmpdir)
        test_npmignore_file(tmpdir)
        test_filter_files(tmpdir)
        test_no_false_positives(tmpdir)
        test_ignore_summary(tmpdir)

    print("\n✅ All ignore validator tests passed!")
