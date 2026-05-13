from pathlib import Path

_data_dir = Path(__file__).parent.parent / "data"
_data_dir.mkdir(parents=True, exist_ok=True)

_db_file = _data_dir / "vectora.db"

# AsyncSqliteSaver uses aiosqlite which expects a file path (Unix/Windows compatible)
# aiosqlite connects directly: aiosqlite.connect(path)
# For portability: convert PosixPath/WindowsPath to string
DB_DSN = str(_db_file)
