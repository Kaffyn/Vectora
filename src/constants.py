"""Global Constants for Vectora Application.

Defines system-wide constants for paths, database configurations, and default values.
Single source of truth for configuration parameters.
"""

from pathlib import Path

# Versão do Vectora (sincronizar com pyproject.toml)
VERSION = "0.1.0rc1"

_data_dir = Path(__file__).parent.parent / "data"
_data_dir.mkdir(parents=True, exist_ok=True)

_db_file = _data_dir / "vectora.db"

# AsyncSqliteSaver usa aiosqlite que espera um caminho de arquivo (compatível Unix/Windows)
# aiosqlite conecta diretamente: aiosqlite.connect(path)
# Para portabilidade: converte PosixPath/WindowsPath para string
DB_DSN = str(_db_file)
