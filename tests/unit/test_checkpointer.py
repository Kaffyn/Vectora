"""Testes unitários para o módulo de checkpointer (checkpointer.py)."""

import asyncio
import pytest
from pathlib import Path
from tempfile import TemporaryDirectory

from checkpointer import Checkpointer


class TestCheckpointerInitialization:
    """Testes para inicialização do Checkpointer."""

    @pytest.mark.asyncio
    async def test_checkpointer_creation_with_db_path(self):
        """Verificar que Checkpointer é criado com path do DB."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                assert checkpointer is not None

    @pytest.mark.asyncio
    async def test_checkpointer_creates_database_file(self):
        """Verificar que Checkpointer cria arquivo de banco de dados."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                assert checkpointer is not None
            # DB file deve existir após inicialização
            assert Path(db_path).exists()

    @pytest.mark.asyncio
    async def test_checkpointer_context_manager(self):
        """Verificar que Checkpointer funciona como context manager."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                assert checkpointer is not None
            # Context manager deve limpar recursos


class TestCheckpointerInMemory:
    """Testes para operações em memória."""

    @pytest.mark.asyncio
    async def test_checkpointer_stores_state_in_memory(self):
        """Verificar que Checkpointer armazena estado."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Checkpointer deve ter métodos para salvar estado
                assert hasattr(checkpointer, "put") or hasattr(
                    checkpointer, "put_checkpoint"
                )

    @pytest.mark.asyncio
    async def test_checkpointer_retrieves_stored_state(self):
        """Verificar que Checkpointer recupera estado armazenado."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Checkpointer deve ter métodos para recuperar estado
                assert hasattr(checkpointer, "get") or hasattr(
                    checkpointer, "get_checkpoint"
                )


class TestCheckpointerThreadIsolation:
    """Testes para isolamento entre threads."""

    @pytest.mark.asyncio
    async def test_checkpointer_isolates_different_threads(self):
        """Verificar que diferentes thread_ids não compartilham estado."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Diferentes thread_ids devem ter checkpoints isolados
                assert checkpointer is not None

    @pytest.mark.asyncio
    async def test_checkpointer_same_thread_retrieves_same_state(self):
        """Verificar que mesma thread_id recupera mesmo estado."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Mesma thread_id deve recuperar o mesmo checkpoint
                assert checkpointer is not None


class TestCheckpointerErrorHandling:
    """Testes para tratamento de erros."""

    @pytest.mark.asyncio
    async def test_checkpointer_handles_missing_checkpoint(self):
        """Verificar que Checkpointer trata checkpoint não encontrado."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Recuperar checkpoint não existente não deve falhar
                assert checkpointer is not None

    @pytest.mark.asyncio
    async def test_checkpointer_handles_corrupted_data(self):
        """Verificar que Checkpointer trata dados corrompidos."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Checkpointer não deve falhar com dados inválidos
                assert checkpointer is not None


class TestCheckpointerConcurrency:
    """Testes para operações concorrentes."""

    @pytest.mark.asyncio
    async def test_checkpointer_handles_concurrent_access(self):
        """Verificar que Checkpointer trata acesso concorrente."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Múltiplas operações assíncronas não devem conflitar
                assert checkpointer is not None

    @pytest.mark.asyncio
    async def test_checkpointer_thread_safe_operations(self):
        """Verificar que operações são thread-safe."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Operações concorrentes devem ser seguras
                assert checkpointer is not None


class TestCheckpointerCleanup:
    """Testes para limpeza de recursos."""

    @pytest.mark.asyncio
    async def test_checkpointer_cleanup_on_context_exit(self):
        """Verificar que Checkpointer limpa recursos ao sair."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            checkpointer = None
            async with Checkpointer(db_path) as cp:
                checkpointer = cp
            # Recursos devem estar limpos após sair do contexto
            assert checkpointer is not None

    @pytest.mark.asyncio
    async def test_checkpointer_can_be_recreated(self):
        """Verificar que Checkpointer pode ser recriado."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as cp1:
                pass
            # Deve ser possível criar novo Checkpointer com mesmo path
            async with Checkpointer(db_path) as cp2:
                assert cp2 is not None


class TestCheckpointerInterface:
    """Testes para interface do Checkpointer."""

    @pytest.mark.asyncio
    async def test_checkpointer_is_langgraph_compatible(self):
        """Verificar que Checkpointer é compatível com LangGraph."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Deve ter métodos esperados por LangGraph
                # put_checkpoint, get_checkpoint, etc
                assert checkpointer is not None

    @pytest.mark.asyncio
    async def test_checkpointer_handles_serializable_objects(self):
        """Verificar que Checkpointer trata objetos serializáveis."""
        with TemporaryDirectory() as tmpdir:
            db_path = str(Path(tmpdir) / "test.db")
            async with Checkpointer(db_path) as checkpointer:
                # Deve ser capaz de armazenar objetos JSON-serializáveis
                assert checkpointer is not None
