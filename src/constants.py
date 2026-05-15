"""Global Constants for Vectora Application.

Defines system-wide constants for paths, database configurations, and default values.
Single source of truth for configuration parameters.

Diretório de roaming: ~/.vectora/ (Unix/macOS) ou %USERPROFILE%/.vectora/ (Windows).
Estrutura:
- ~/.vectora/data/          → SQLite databases, LanceDB vector store
- ~/.vectora/logs/          → Application logs
- ~/.vectora/keys/          → Sensitive credentials (auth tokens, API keys)
- ~/.vectora/.env           → Environment-specific configuration
- ~/.vectora/mcp.config.json → MCP client configuration
"""

import os
from pathlib import Path

# Versão do Vectora (sincronizar com pyproject.toml)
VERSION = "0.1.0rc1"

# Diretório de roaming: ~/.vectora/
# Em Windows: %USERPROFILE%/.vectora/
# Em Unix: ~/.vectora/
VECTORA_HOME = Path.home() / ".vectora"
VECTORA_HOME.mkdir(parents=True, exist_ok=True)

# Sub-diretórios
DATA_DIR = VECTORA_HOME / "data"
LOGS_DIR = VECTORA_HOME / "logs"
KEYS_DIR = VECTORA_HOME / "keys"

# Garantir criação de sub-diretórios
DATA_DIR.mkdir(parents=True, exist_ok=True)
LOGS_DIR.mkdir(parents=True, exist_ok=True)
KEYS_DIR.mkdir(parents=True, exist_ok=True)

# Arquivos de banco de dados
DB_FILE = DATA_DIR / "vectora.db"
EMBEDDING_QUEUE_FILE = DATA_DIR / "embedding_queue.db"

# AsyncSqliteSaver usa aiosqlite que espera um caminho de arquivo (compatível Unix/Windows)
# Para portabilidade: converte PosixPath/WindowsPath para string
DB_DSN = str(DB_FILE)
EMBEDDING_QUEUE_DSN = str(EMBEDDING_QUEUE_FILE)

# LanceDB (vector store local)
LANCEDB_DIR = str(DATA_DIR / "lancedb")

# Arquivos de configuração
ENV_FILE = VECTORA_HOME / ".env"
MCP_CONFIG_FILE = VECTORA_HOME / "mcp.config.json"

# Log file
LOG_FILE = LOGS_DIR / "vectora.log"

# Rastreabilidade: env var para sobrescrever paths (testabilidade)
_override_data_dir = os.getenv("VECTORA_DATA_DIR")
if _override_data_dir:
    DB_DSN = str(Path(_override_data_dir) / "vectora.db")
    EMBEDDING_QUEUE_DSN = str(Path(_override_data_dir) / "embedding_queue.db")
    LANCEDB_DIR = str(Path(_override_data_dir) / "lancedb")
