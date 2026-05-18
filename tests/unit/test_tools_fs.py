"""Testes para vectora/tools/fs.py"""

from __future__ import annotations


def test_fs_module_imports():
    """Verificar que fs pode ser importado."""
    from vectora.tools import fs

    assert fs is not None


def test_fs_has_logger():
    """Verificar que fs tem logger."""
    from vectora.tools import fs

    assert hasattr(fs, "logger")


def test_fs_has_path():
    """Verificar que fs usa Path."""
    from vectora.tools import fs

    assert hasattr(fs, "Path")


def test_fs_has_functions_to_call():
    """Verificar que fs tem funcoes."""
    import inspect

    from vectora.tools import fs

    # Contar funcoes publicas
    funcs = [
        name
        for name, obj in inspect.getmembers(fs, inspect.isfunction)
        if not name.startswith("_")
    ]
    assert len(funcs) > 0
