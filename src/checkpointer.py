from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager

from langgraph.checkpoint.sqlite.aio import AsyncSqliteSaver

from constants import DB_DSN


@asynccontextmanager
async def Checkpointer(
    db_dsn: str | None = None,
) -> AsyncGenerator[AsyncSqliteSaver]:
    """Constrói checkpointer SQLite assíncrono para persistência local de conversas.

    Usa aiosqlite via AsyncSqliteSaver do LangGraph. O arquivo SQLite é criado
    automaticamente no diretório `data/` na raiz do projeto.

    Args:
        db_dsn: Caminho para o arquivo SQLite. Se None, usa o padrão de `constants.DB_DSN`.
    """
    conn_string = db_dsn or DB_DSN
    async with AsyncSqliteSaver.from_conn_string(conn_string) as checkpointer:
        yield checkpointer
