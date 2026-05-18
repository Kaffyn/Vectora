"""Testes para services/debug_dump.py.

Cobre: generate_debug_dump, create_qa_report
"""

from __future__ import annotations

import tarfile
from pathlib import Path
from unittest.mock import patch

import pytest


class TestGenerateDebugDump:
    """Testa generate_debug_dump."""

    @pytest.mark.asyncio
    async def test_generates_tar_gz_file(self, tmp_path):
        """Verifica que generate_debug_dump cria arquivo .tar.gz."""
        from vectora.services.debug_dump import generate_debug_dump

        output = str(tmp_path / "test_dump.tar.gz")

        with patch("vectora.services.debug_dump.settings") as mock_settings:
            mock_settings.get_llm_provider.return_value = "test_provider"
            mock_settings.enable_rag = False
            result = await generate_debug_dump(
                output_file=output, include_databases=False, include_logs=False
            )

        assert result == output
        assert Path(output).exists()

    @pytest.mark.asyncio
    async def test_tar_contains_info_json(self, tmp_path):
        """Verifica que o tar.gz contém INFO.json com metadados."""
        from vectora.services.debug_dump import generate_debug_dump

        output = str(tmp_path / "info_test.tar.gz")

        with patch("vectora.services.debug_dump.settings") as mock_settings:
            mock_settings.get_llm_provider.return_value = "openai"
            mock_settings.enable_rag = True
            await generate_debug_dump(
                output_file=output, include_databases=False, include_logs=False
            )

        with tarfile.open(output, "r:gz") as tar:
            names = tar.getnames()
        assert "INFO.json" in names

    @pytest.mark.asyncio
    async def test_auto_generates_filename(self, tmp_path):
        """Verifica que gera nome de arquivo automaticamente quando None."""
        import os
        from pathlib import Path

        from vectora.services.debug_dump import generate_debug_dump

        original_cwd = Path.cwd()
        os.chdir(tmp_path)
        try:
            with patch("vectora.services.debug_dump.settings") as mock_settings:
                mock_settings.get_llm_provider.return_value = "test"
                mock_settings.enable_rag = False
                result = await generate_debug_dump(
                    output_file=None, include_databases=False, include_logs=False
                )

            assert result.startswith("vectora_debug_")
            assert result.endswith(".tar.gz")
            assert Path(result).exists()
        finally:
            os.chdir(original_cwd)

    @pytest.mark.asyncio
    async def test_includes_env_safe_when_env_exists(self, tmp_path):
        """Verifica que .env.safe é incluído quando .env existe (sem secrets)."""
        import os
        from pathlib import Path

        from vectora.services.debug_dump import generate_debug_dump

        env_file = tmp_path / ".env"
        env_file.write_text("LOG_LEVEL=DEBUG\nCOHERE_API_KEY=secret123\nDEBUG=true\n")

        original_cwd = Path.cwd()
        os.chdir(tmp_path)
        try:
            output = str(tmp_path / "env_test.tar.gz")
            with patch("vectora.services.debug_dump.settings") as mock_settings:
                mock_settings.get_llm_provider.return_value = "test"
                mock_settings.enable_rag = False
                await generate_debug_dump(
                    output_file=output, include_databases=False, include_logs=False
                )

            with tarfile.open(output, "r:gz") as tar:
                names = tar.getnames()
            assert ".env.safe" in names

            # .env.safe não deve ter a linha com KEY
            with tarfile.open(output, "r:gz") as tar:
                f = tar.extractfile(".env.safe")
                content = f.read().decode()
            assert "secret123" not in content
            assert "LOG_LEVEL=DEBUG" in content
        finally:
            os.chdir(original_cwd)

    @pytest.mark.asyncio
    async def test_raises_on_write_failure(self, tmp_path):
        """Verifica que exceção é re-levantada em caso de falha."""
        from vectora.services.debug_dump import generate_debug_dump

        with patch("vectora.services.debug_dump.settings") as mock_settings:
            mock_settings.get_llm_provider.return_value = "test"
            mock_settings.enable_rag = False
            with patch(
                "vectora.services.debug_dump.tarfile.open",
                side_effect=OSError("disk full"),
            ):
                with pytest.raises(OSError, match="disk full"):
                    await generate_debug_dump(output_file="/invalid/path/test.tar.gz")


class TestCreateQAReport:
    """Testa create_qa_report."""

    @pytest.mark.asyncio
    async def test_creates_markdown_report(self, tmp_path):
        """Verifica que create_qa_report retorna markdown válido."""
        from vectora.services.debug_dump import create_qa_report, generate_debug_dump

        output = str(tmp_path / "dump.tar.gz")
        with patch("vectora.services.debug_dump.settings") as mock_settings:
            mock_settings.get_llm_provider.return_value = "test"
            mock_settings.enable_rag = False
            dump_path = await generate_debug_dump(
                output_file=output, include_databases=False, include_logs=False
            )

        report = await create_qa_report(
            debug_dump_path=dump_path,
            tester_name="Tester Silva",
            test_scenario="Teste de embedding",
            bug_description="O embedding falhou com erro 500",
            severity="high",
        )

        assert "Tester Silva" in report
        assert "Teste de embedding" in report
        assert "O embedding falhou com erro 500" in report
        assert "HIGH" in report
        assert dump_path in report
        assert "## Informações do Teste" in report

    @pytest.mark.asyncio
    async def test_default_severity_is_medium(self, tmp_path):
        """Verifica que severidade padrão é medium."""
        from vectora.services.debug_dump import create_qa_report, generate_debug_dump

        output = str(tmp_path / "dump2.tar.gz")
        with patch("vectora.services.debug_dump.settings") as mock_settings:
            mock_settings.get_llm_provider.return_value = "test"
            mock_settings.enable_rag = False
            dump_path = await generate_debug_dump(
                output_file=output, include_databases=False, include_logs=False
            )

        report = await create_qa_report(
            debug_dump_path=dump_path,
            tester_name="Tester",
            test_scenario="Cenário",
            bug_description="Bug",
        )

        assert "MEDIUM" in report
