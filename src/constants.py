from pathlib import Path

# Versão do Vectora (sincronizar com pyproject.toml)
VERSION = "0.1.0"

_data_dir = Path(__file__).parent.parent / "data"
_data_dir.mkdir(parents=True, exist_ok=True)

_db_file = _data_dir / "vectora.db"

# AsyncSqliteSaver usa aiosqlite que espera um caminho de arquivo (compatível Unix/Windows)
# aiosqlite conecta diretamente: aiosqlite.connect(path)
# Para portabilidade: converte PosixPath/WindowsPath para string
DB_DSN = str(_db_file)
