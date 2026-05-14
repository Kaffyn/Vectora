import logging
import os
import sys
from pathlib import Path

# Configuração de logging rigorosa: redireciona tudo para arquivo para não poluir o stdout (JSON-RPC)
log_dir = Path("logs")
log_dir.mkdir(exist_ok=True)
log_file = log_dir / "mcp.log"

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.FileHandler(log_file, encoding="utf-8")],
)

logger = logging.getLogger("vectora-mcp")

try:
    from mcp.server.fastmcp import FastMCP
    from tools import fetch_url, grep, list_dir, vector_search, web_search
except ImportError as e:
    logger.error(f"Falha ao importar dependências do MCP: {e}")
    sys.exit(1)

# Inicializa o servidor FastMCP (Vector Context Protocol)
mcp = FastMCP(
    "Vectora-Agent",
    description="Infraestrutura de contexto e busca semântica (RAG) local-first",
)

# Adiciona ferramentas do Vectora ao servidor MCP
# O FastMCP gerencia automaticamente a conversão de docstrings em descrições de ferramentas
mcp.add_tool(vector_search)
mcp.add_tool(web_search)
mcp.add_tool(fetch_url)
mcp.add_tool(list_dir)
mcp.add_tool(grep)

# DICA: Você também pode adicionar recursos (Resources) aqui se desejar 
# expor arquivos de log ou configurações via MCP futuramente.

if __name__ == "__main__":
    logger.info("Iniciando servidor MCP Vectora via Stdio...")
    
    # FastMCP.run() detecta automaticamente se deve usar stdio (padrão) ou sse
    # Ele lida com o handshake JSON-RPC de forma transparente.
    try:
        mcp.run()
    except Exception as e:
        logger.exception("Erro crítico no servidor MCP")
        sys.exit(1)
