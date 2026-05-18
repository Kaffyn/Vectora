"""Configurações e fixtures globais para testes unitários."""

from __future__ import annotations

import asyncio
from contextlib import AsyncExitStack


def pytest_configure(config) -> None:  # type: ignore[no-untyped-def]
    """Configurar pytest com mocks globais."""
    # Mockar asyncio.ExitStack com contextlib.AsyncExitStack no início
    asyncio.ExitStack = AsyncExitStack  # type: ignore[assignment]
