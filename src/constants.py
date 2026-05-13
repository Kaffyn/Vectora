from pathlib import Path

_data_dir = Path(__file__).parent.parent / "data"
_data_dir.mkdir(parents=True, exist_ok=True)

# Use absolute path for SQLite database (works on both Unix and Windows)
# Format: <absolute_path>/vectora.db (aiosqlite expects direct path, not sqlite:// URI on Windows)
DB_DSN = str(_data_dir / "vectora.db")
