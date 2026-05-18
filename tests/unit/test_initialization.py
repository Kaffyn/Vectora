"""Testes para Vectora initialization and setup utilities."""

from __future__ import annotations

from pathlib import Path
from unittest.mock import patch

from vectora.services.initialization import (
    ensure_vectora_initialized,
    initialize_vectora_home,
    verify_vectora_setup,
)


class TestInitializeVectoraHome:
    """Testes para initialize_vectora_home()."""

    def test_initialize_creates_home_directory(self, tmp_path: Path) -> None:
        """Verificar que initialize_vectora_home() cria diretório home."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            result = initialize_vectora_home()

            vectora_home = tmp_path / ".vectora"
            assert result == vectora_home
            assert vectora_home.exists()

    def test_initialize_creates_subdirectories(self, tmp_path: Path) -> None:
        """Verificar que initialize_vectora_home() cria subdirs data e logs."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            result = initialize_vectora_home()

            vectora_home = result
            assert (vectora_home / "data").exists()
            assert (vectora_home / "logs").exists()

    def test_initialize_idempotent(self, tmp_path: Path) -> None:
        """Verificar que initialize_vectora_home() é idempotente."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            result1 = initialize_vectora_home()
            result2 = initialize_vectora_home()

            assert result1 == result2
            assert result1.exists()

    def test_initialize_returns_path_object(self, tmp_path: Path) -> None:
        """Verificar que initialize_vectora_home() retorna Path."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            result = initialize_vectora_home()

            assert isinstance(result, Path)

    def test_initialize_all_subdirs_are_directories(self, tmp_path: Path) -> None:
        """Verificar que todos os subdirs criados são diretórios."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            result = initialize_vectora_home()

            assert result.is_dir()
            assert (result / "data").is_dir()
            assert (result / "logs").is_dir()


class TestVerifyVectoraSetup:
    """Testes para verify_vectora_setup()."""

    def test_verify_returns_true_when_all_dirs_exist(self, tmp_path: Path) -> None:
        """Verificar que verify_vectora_setup() retorna True quando tudo existe."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            # Primeiro inicializar
            initialize_vectora_home()
            # Depois verificar
            result = verify_vectora_setup()

            assert result is True

    def test_verify_returns_false_when_home_missing(self, tmp_path: Path) -> None:
        """Verificar que verify_vectora_setup() retorna False se home não existe."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            # Não chamar initialize_vectora_home()
            result = verify_vectora_setup()

            assert result is False

    def test_verify_returns_false_when_data_missing(self, tmp_path: Path) -> None:
        """Verificar que verify_vectora_setup() retorna False se data falta."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            vectora_home = tmp_path / ".vectora"
            vectora_home.mkdir(parents=True, exist_ok=True)
            # Criar apenas data, não logs
            (vectora_home / "data").mkdir(parents=True, exist_ok=True)

            result = verify_vectora_setup()

            assert result is False

    def test_verify_returns_false_when_logs_missing(self, tmp_path: Path) -> None:
        """Verificar que verify_vectora_setup() retorna False se logs falta."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            vectora_home = tmp_path / ".vectora"
            vectora_home.mkdir(parents=True, exist_ok=True)
            # Criar apenas logs, não data
            (vectora_home / "logs").mkdir(parents=True, exist_ok=True)

            result = verify_vectora_setup()

            assert result is False

    def test_verify_checks_all_required_directories(self, tmp_path: Path) -> None:
        """Verificar que verify_vectora_setup() verifica todos os dirs obrigatórios."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            # Criar apenas a home, sem subdirs
            vectora_home = tmp_path / ".vectora"
            vectora_home.mkdir(parents=True, exist_ok=True)

            result = verify_vectora_setup()

            # Deve retornar False porque data e logs estão faltando
            assert result is False


class TestEnsureVectoraInitialized:
    """Testes para ensure_vectora_initialized()."""

    def test_ensure_initializes_if_missing(self, tmp_path: Path) -> None:
        """Verificar que ensure_vectora_initialized() inicializa se necessário."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            # Não deve haver nada inicialmente
            assert not (tmp_path / ".vectora").exists()

            # Chamar ensure
            ensure_vectora_initialized()

            # Agora tudo deve estar criado
            vectora_home = tmp_path / ".vectora"
            assert vectora_home.exists()
            assert (vectora_home / "data").exists()
            assert (vectora_home / "logs").exists()

    def test_ensure_is_idempotent(self, tmp_path: Path) -> None:
        """Verificar que ensure_vectora_initialized() é idempotente."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            ensure_vectora_initialized()
            ensure_vectora_initialized()
            ensure_vectora_initialized()

            # Tudo deve continuar existindo
            vectora_home = tmp_path / ".vectora"
            assert vectora_home.exists()

    def test_ensure_creates_all_subdirectories(self, tmp_path: Path) -> None:
        """Verificar que ensure_vectora_initialized() cria todos os subdirs."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            ensure_vectora_initialized()

            vectora_home = tmp_path / ".vectora"
            assert (vectora_home / "data").is_dir()
            assert (vectora_home / "logs").is_dir()

    def test_ensure_no_return_value(self, tmp_path: Path) -> None:
        """Verificar que ensure_vectora_initialized() retorna None."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            result = ensure_vectora_initialized()

            assert result is None

    def test_ensure_verifies_after_initialization(self, tmp_path: Path) -> None:
        """Verificar que ensure_vectora_initialized() verifica após criar."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            ensure_vectora_initialized()

            # Após ensure, verify deve retornar True
            with patch("pathlib.Path.home", return_value=tmp_path):
                result = verify_vectora_setup()

            assert result is True


class TestInitializationIntegration:
    """Testes de integração das funções de inicialização."""

    def test_initialize_then_verify_succeeds(self, tmp_path: Path) -> None:
        """Verificar que initialize() seguido de verify() funciona."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            initialize_vectora_home()
            result = verify_vectora_setup()

            assert result is True

    def test_ensure_makes_verify_pass(self, tmp_path: Path) -> None:
        """Verificar que ensure() faz verify() passar."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            # Inicialmente, verify deve falhar
            assert verify_vectora_setup() is False

            # Depois de ensure, deve passar
            ensure_vectora_initialized()
            assert verify_vectora_setup() is True

    def test_multiple_ensure_calls_safe(self, tmp_path: Path) -> None:
        """Verificar que múltiplas chamadas a ensure() são seguras."""
        with patch("pathlib.Path.home", return_value=tmp_path):
            for _ in range(5):
                ensure_vectora_initialized()
                assert verify_vectora_setup() is True
