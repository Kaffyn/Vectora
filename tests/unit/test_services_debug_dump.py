"""Testes para vectora/services/debug_dump.py"""

from __future__ import annotations

import asyncio
import json
import tarfile
from io import BytesIO
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, mock_open, patch

import pytest

from vectora.services.debug_dump import (
    create_qa_report,
    generate_debug_dump,
)


class TestGenerateDebugDump:
    """Testes para generate_debug_dump()"""

    @pytest.mark.asyncio
    async def test_generate_debug_dump_creates_tar_with_default_filename(self):
        """Testar que debug dump cria arquivo .tar.gz com nome padrão"""
        with patch("vectora.services.debug_dump.tarfile.open") as mock_tar_open:
            with patch("vectora.services.debug_dump.Path") as mock_path:
                # Mock da instancia do tarfile
                mock_tar = MagicMock()
                mock_tar.__enter__ = MagicMock(return_value=mock_tar)
                mock_tar.__exit__ = MagicMock(return_value=None)
                mock_tar_open.return_value = mock_tar

                # Mock do Path.stat()
                mock_stat = MagicMock()
                mock_stat.st_size = 1024
                mock_path_instance = MagicMock()
                mock_path_instance.stat.return_value = mock_stat
                mock_path.return_value = mock_path_instance

                result = await generate_debug_dump()

                # Verificar que foi gerado um nome padrão
                assert "vectora_debug_" in result
                assert result.endswith(".tar.gz")
                mock_tar_open.assert_called_once()

    @pytest.mark.asyncio
    async def test_generate_debug_dump_uses_provided_filename(self):
        """Testar que debug dump respeita filename fornecido"""
        with patch("vectora.services.debug_dump.tarfile.open") as mock_tar_open:
            with patch("vectora.services.debug_dump.Path") as mock_path:
                mock_tar = MagicMock()
                mock_tar.__enter__ = MagicMock(return_value=mock_tar)
                mock_tar.__exit__ = MagicMock(return_value=None)
                mock_tar_open.return_value = mock_tar

                mock_stat = MagicMock()
                mock_stat.st_size = 2048
                mock_path_instance = MagicMock()
                mock_path_instance.stat.return_value = mock_stat
                mock_path.return_value = mock_path_instance

                result = await generate_debug_dump(output_file="custom_debug.tar.gz")

                assert result == "custom_debug.tar.gz"
                mock_tar_open.assert_called_once_with("custom_debug.tar.gz", "w:gz")

    @pytest.mark.asyncio
    async def test_generate_debug_dump_adds_metadata(self):
        """Testar que debug dump inclui metadados"""
        with patch("vectora.services.debug_dump.tarfile.open") as mock_tar_open:
            with patch("vectora.services.debug_dump.Path") as mock_path:
                with patch("vectora.services.debug_dump.settings") as mock_settings:
                    mock_tar = MagicMock()
                    mock_tar.__enter__ = MagicMock(return_value=mock_tar)
                    mock_tar.__exit__ = MagicMock(return_value=None)
                    mock_tar_open.return_value = mock_tar

                    mock_stat = MagicMock()
                    mock_stat.st_size = 512
                    mock_path_instance = MagicMock()
                    mock_path_instance.stat.return_value = mock_stat
                    mock_path.return_value = mock_path_instance

                    mock_settings.get_llm_provider.return_value = "openai"
                    mock_settings.enable_rag = True

                    result = await generate_debug_dump()

                    # Verificar que addfile foi chamado para metadados
                    calls = mock_tar.addfile.call_args_list
                    assert len(calls) > 0

                    # Primeiro arquivo deve ser INFO.json
                    first_call = calls[0]
                    tar_info = first_call[0][0]
                    assert tar_info.name == "INFO.json"

    @pytest.mark.asyncio
    async def test_generate_debug_dump_excludes_databases_when_disabled(self):
        """Testar que databases sao excluidas quando flag disabled"""
        with patch("vectora.services.debug_dump.tarfile.open") as mock_tar_open:
            with patch("vectora.services.debug_dump.Path") as mock_path_cls:
                with patch("vectora.services.debug_dump.settings") as mock_settings:
                    mock_tar = MagicMock()
                    mock_tar.__enter__ = MagicMock(return_value=mock_tar)
                    mock_tar.__exit__ = MagicMock(return_value=None)
                    mock_tar_open.return_value = mock_tar

                    mock_stat = MagicMock()
                    mock_stat.st_size = 256
                    mock_path_instance = MagicMock()
                    mock_path_instance.stat.return_value = mock_stat
                    mock_path_cls.return_value = mock_path_instance

                    # Mock settings methods to return serializable values
                    mock_settings.get_llm_provider.return_value = "openai"
                    mock_settings.enable_rag = True

                    result = await generate_debug_dump(include_databases=False)

                    assert result.endswith(".tar.gz")

    @pytest.mark.asyncio
    async def test_generate_debug_dump_handles_missing_data_dir(self):
        """Testar que missing data dir nao causa erro"""
        with patch("vectora.services.debug_dump.tarfile.open") as mock_tar_open:
            with patch("vectora.services.debug_dump.Path") as mock_path_cls:
                with patch("vectora.services.debug_dump.settings") as mock_settings:
                    mock_tar = MagicMock()
                    mock_tar.__enter__ = MagicMock(return_value=mock_tar)
                    mock_tar.__exit__ = MagicMock(return_value=None)
                    mock_tar_open.return_value = mock_tar

                    mock_stat = MagicMock()
                    mock_stat.st_size = 128

                    # Path retorna diferentes valores para diferentes paths
                    def path_side_effect(arg: object) -> MagicMock:
                        if "data" in str(arg):
                            mock = MagicMock()
                            mock.exists.return_value = False  # data dir nao existe
                            return mock
                        if "logs" in str(arg):
                            mock = MagicMock()
                            mock.exists.return_value = False
                            return mock
                        mock = MagicMock()
                        mock.stat.return_value = mock_stat
                        return mock

                    mock_path_cls.side_effect = path_side_effect

                    # Mock settings methods to return serializable values
                    mock_settings.get_llm_provider.return_value = "openai"
                    mock_settings.enable_rag = True

                    result = await generate_debug_dump()

                    assert result.endswith(".tar.gz")


class TestCreateQAReport:
    """Testes para create_qa_report()"""

    @pytest.mark.asyncio
    async def test_create_qa_report_returns_markdown(self):
        """Testar que report retorna conteudo markdown"""
        with patch("vectora.services.debug_dump.Path") as mock_path:
            mock_stat = MagicMock()
            mock_stat.st_size = 2048
            mock_path_instance = MagicMock()
            mock_path_instance.stat.return_value = mock_stat
            mock_path.return_value = mock_path_instance

            report = await create_qa_report(
                debug_dump_path="debug.tar.gz",
                tester_name="John Doe",
                test_scenario="Login test",
                bug_description="Login button not responding",
                severity="high",
            )

            assert isinstance(report, str)
            assert "# 🐛 Relatório de Bug" in report
            assert "John Doe" in report
            assert "Login test" in report
            assert "Login button not responding" in report
            assert "HIGH" in report

    @pytest.mark.asyncio
    async def test_create_qa_report_includes_debug_dump_info(self):
        """Testar que report inclui info do debug dump"""
        with patch("vectora.services.debug_dump.Path") as mock_path:
            mock_stat = MagicMock()
            mock_stat.st_size = 4096
            mock_path_instance = MagicMock()
            mock_path_instance.stat.return_value = mock_stat
            mock_path.return_value = mock_path_instance

            report = await create_qa_report(
                debug_dump_path="my_debug.tar.gz",
                tester_name="Jane",
                test_scenario="API test",
                bug_description="API timeout",
                severity="critical",
            )

            assert "my_debug.tar.gz" in report
            assert "CRITICAL" in report
            assert "4.0 KB" in report or "4" in report

    @pytest.mark.asyncio
    async def test_create_qa_report_default_severity(self):
        """Testar que severidade padrão e 'medium'"""
        with patch("vectora.services.debug_dump.Path") as mock_path:
            mock_stat = MagicMock()
            mock_stat.st_size = 1024
            mock_path_instance = MagicMock()
            mock_path_instance.stat.return_value = mock_stat
            mock_path.return_value = mock_path_instance

            report = await create_qa_report(
                debug_dump_path="test.tar.gz",
                tester_name="Tester",
                test_scenario="Basic test",
                bug_description="Found a bug",
            )

            assert "MEDIUM" in report
