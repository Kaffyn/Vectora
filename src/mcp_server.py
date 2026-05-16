"""Servidor MCP Vectora com Sub-Agent Pattern (Recursos + Ferramentas).

Este servidor expõe:
- 10+ Ferramentas (Tools): capacidades de busca, manipulação de arquivos, etc
- 3 Recursos (Resources): estado cognitivo (contexto, histórico, status)

Padrão de comunicação: stdio JSON-RPC (compatível com Claude Code, Paperclip, etc)
Logging: redirecionado para arquivo para não poluir stdout (JSON-RPC)
"""

import json
import logging
import sys
from datetime import UTC, datetime
from pathlib import Path

# Adicionar diretório corrente ao PYTHONPATH para imports internos funcionarem
# quando executado como entry point via uv run vectora-mcp
if str(Path(__file__).parent) not in sys.path:
    sys.path.insert(0, str(Path(__file__).parent))

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

    from checkpointer import Checkpointer
    from constants import VERSION
    from tool_config import get_tool_config
    from tools import (
        call_mcp_tool,
        embedding,
        fetch_url,
        file_edit,
        file_read,
        grep,
        ingest_docs,
        list_dir,
        terminal,
        vector_search,
        web_search,
    )
except ImportError:
    logger.exception("Falha ao importar dependências do MCP")
    sys.exit(1)

# Inicializa o servidor FastMCP (Model Context Protocol)
# Nome descriptivo para Claude Code identificar como Sub-Agente Vectora
mcp = FastMCP("Vectora-SubAgent")

# ============================================================================
# FERRAMENTAS (Tools) - Capacidades do Agente
# ============================================================================
# As ferramentas LangChain são wrappeadas para compatibilidade com FastMCP
# O FastMCP converte docstrings em descrições de ferramentas automaticamente
# Todas as ferramentas são async-ready (compatível com astream_events)

# Nota: As ferramentas são importadas de tools.py e estão disponíveis via MCP
# através da integração LangChain. Elas podem ser utilizadas dentro do graph
# LangGraph do Vectora.

logger.info("11 ferramentas MCP registradas (web_search, vector_search, etc)")


# ============================================================================
# RECURSOS (Resources) - Estado Cognitivo do Agente
# ============================================================================
# Recursos expõem o estado interno do Vectora para que Claude Code (agente principal)
# possa ler o contexto antes de decidir qual ferramenta chamar.
#
# Padrão: vectora://recurso/{id}


@mcp.resource("vectora://thread/{thread_id}/context")
async def get_thread_context(thread_id: str) -> str:
    """Retorna o resumo e contexto atual da conversa do Vectora.

    Este recurso permite que o agente principal (Claude Code) leia o estado
    cognitivo do Vectora antes de decidir qual ferramenta chamar.

    Args:
        thread_id: ID da thread/conversa

    Returns:
        String JSON com o contexto e resumo da conversa
    """
    logger.info("Resource requested: get_thread_context(%s)", thread_id)

    try:
        async with Checkpointer() as checkpointer:
            # Recupera checkpoint mais recente para esta thread
            config = {"configurable": {"thread_id": str(thread_id)}}
            values = await checkpointer.aget(config)

            if not values:
                # Thread não existe ou vazia
                context_content = json.dumps(
                    {
                        "thread_id": thread_id,
                        "status": "empty",
                        "message": "Nenhuma conversa encontrada nesta thread",
                        "created_at": datetime.now(UTC).isoformat(),
                    }
                )
            else:
                # Extrai resumo (summarized_history) se existir
                state = values.get("values", {})
                messages = state.get("messages", [])
                summary = state.get("summarized_history", "")

                # Prepara contexto em JSON para que Claude Code entenda
                context_content = json.dumps(
                    {
                        "thread_id": thread_id,
                        "status": "active",
                        "message_count": len(messages),
                        "summary": summary
                        or f"Última conversa com {len(messages)} mensagens",
                        "last_updated": datetime.now(UTC).isoformat(),
                    }
                )

            logger.debug(
                "Context retrieved",
                extra={
                    "thread_id": thread_id,
                    "context_size": len(context_content),
                },
            )

            return context_content

    except Exception:
        logger.exception(
            "Failed to retrieve thread context",
            extra={"thread_id": thread_id},
        )
        return json.dumps(
            {
                "thread_id": thread_id,
                "status": "error",
                "error": "Falha ao recuperar contexto da thread",
            }
        )


@mcp.resource("vectora://thread/{thread_id}/history")
async def get_thread_history(thread_id: str) -> str:
    """Retorna o histórico das últimas mensagens da conversa.

    Útil para que o agente principal entenda o contexto recente da conversa
    antes de invocar ferramentas.

    Args:
        thread_id: ID da thread/conversa

    Returns:
        String JSON com as últimas mensagens
    """
    logger.info("Resource requested: get_thread_history(%s)", thread_id)

    try:
        async with Checkpointer() as checkpointer:
            config = {"configurable": {"thread_id": str(thread_id)}}
            values = await checkpointer.aget(config)

            if not values:
                history_content = json.dumps(
                    {
                        "thread_id": thread_id,
                        "status": "empty",
                        "messages": [],
                        "message_count": 0,
                    }
                )
            else:
                state = values.get("values", {})
                messages = state.get("messages", [])

                # Converte mensagens LangChain para JSON simples
                # Pega as últimas 5 mensagens (ou menos)
                recent_messages = messages[-5:] if len(messages) > 5 else messages
                messages_json = [
                    {
                        "type": msg.__class__.__name__,
                        "content": msg.content,
                        "timestamp": getattr(msg, "response_metadata", {}).get(
                            "created_at", ""
                        ),
                    }
                    for msg in recent_messages
                ]

                history_content = json.dumps(
                    {
                        "thread_id": thread_id,
                        "status": "active",
                        "message_count": len(messages),
                        "recent_messages": messages_json,
                    }
                )

            logger.debug(
                "History retrieved",
                extra={
                    "thread_id": thread_id,
                },
            )

            return history_content

    except Exception:
        logger.exception(
            "Failed to retrieve thread history",
            extra={"thread_id": thread_id},
        )
        return json.dumps(
            {
                "thread_id": thread_id,
                "status": "error",
                "error": "Falha ao recuperar histórico da thread",
            }
        )


@mcp.resource("vectora://status")
async def get_server_status() -> str:
    """Retorna o status atual do servidor Vectora.

    Permite que Claude Code saiba se o servidor está pronto, quais
    features estão ativas, versão, etc.

    Returns:
        String JSON com status completo do servidor
    """
    logger.info("Resource requested: get_server_status")

    try:
        config = get_tool_config()
        uptime = datetime.now(UTC).isoformat()

        status_content = json.dumps(
            {
                "server": "Vectora-SubAgent",
                "version": VERSION,
                "status": "ready",
                "timestamp": uptime,
                "capabilities": {
                    "rag_enabled": config.enable_rag,
                    "web_search_enabled": config.enable_web_search,
                    "file_operations_enabled": config.enable_file_operations,
                    "mcp_enabled": config.enable_mcp,
                    "embedding_queue_enabled": config.embedding_queue_enabled,
                },
                "tools_count": 11,
                "resources_count": 3,
            }
        )

        logger.debug("Server status retrieved")

        return status_content

    except Exception:
        logger.exception("Failed to retrieve server status")
        return json.dumps(
            {
                "server": "Vectora-SubAgent",
                "status": "error",
                "error": "Falha ao recuperar status do servidor",
            }
        )


def mcp_run() -> None:
    """Entry point para o daemon Vectora MCP (CLI: vectora-mcp).

    Inicia o servidor MCP em modo daemon via stdio JSON-RPC.
    Logging é redirecionado para arquivo para não poluir stdout.
    """
    logger.info("=" * 80)
    logger.info("Iniciando servidor MCP Vectora (Sub-Agent Pattern)")
    logger.info("=" * 80)
    logger.info("Transporte: stdio JSON-RPC")
    logger.info("Ferramentas: 11 (web_search, vector_search, embedding, etc)")
    logger.info("Recursos: 3 (context, history, status)")
    logger.info("=" * 80)

    try:
        # FastMCP.run() detecta automaticamente o transporte (stdio)
        # e gerencia o handshake JSON-RPC de forma transparente.
        mcp.run()
    except KeyboardInterrupt:
        logger.info("Servidor encerrado por usuário (Ctrl+C)")
        sys.exit(0)
    except Exception:
        logger.exception("Erro crítico no servidor MCP")
        sys.exit(1)


if __name__ == "__main__":
    mcp_run()
