"""ToolNodes especializados por categoria de ferramenta.

Cada ToolNode agrupa apenas as ferramentas relevantes para o worker correspondente,
evitando que um worker de código veja ferramentas de busca e vice-versa.
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from langgraph.prebuilt import ToolNode

from vectora.config.settings import settings
from vectora.tools.fs import file_edit, file_read, file_write, grep, list_dir, terminal
from vectora.tools.memory import delete_memory, get_memory, save_memory
from vectora.tools.rag import embedding, ingest_docs, vector_search
from vectora.tools.web import fetch_url, web_search

if TYPE_CHECKING:
    from langchain_core.tools import BaseTool

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Grupos de ferramentas por domínio
# ---------------------------------------------------------------------------

#: Ferramentas de busca e pesquisa (search_worker)
SEARCH_TOOLS: list[BaseTool] = [web_search, fetch_url, vector_search]
if settings.enable_rag:
    SEARCH_TOOLS.append(embedding)

#: Ferramentas de filesystem e terminal (coder_worker)
FS_TOOLS: list[BaseTool] = []
if settings.enable_file_operations:
    FS_TOOLS.extend([file_read, file_edit, file_write, grep, list_dir, terminal])

#: Ferramentas de memória (disponível para qualquer worker)
MEMORY_TOOLS: list[BaseTool] = [save_memory, get_memory, delete_memory]

#: Ferramentas RAG de ingestão (disponível para coder/search)
RAG_TOOLS: list[BaseTool] = [vector_search]
if settings.enable_rag:
    RAG_TOOLS.extend([embedding, ingest_docs])

#: Conjunto completo (fallback / supervisor)
ALL_TOOLS: list[BaseTool] = list(
    {t.name: t for t in [*SEARCH_TOOLS, *FS_TOOLS, *MEMORY_TOOLS, *RAG_TOOLS]}.values()
)

# ---------------------------------------------------------------------------
# ToolNodes prontos para uso nos workers
# ---------------------------------------------------------------------------

search_tool_node = ToolNode(tools=SEARCH_TOOLS)
coder_tool_node = ToolNode(tools=FS_TOOLS) if FS_TOOLS else ToolNode(tools=[])
memory_tool_node = ToolNode(tools=MEMORY_TOOLS)
all_tool_node = ToolNode(tools=ALL_TOOLS)

logger.debug(
    "ToolNodes inicializados",
    extra={
        "search": len(SEARCH_TOOLS),
        "coder": len(FS_TOOLS),
        "memory": len(MEMORY_TOOLS),
        "all": len(ALL_TOOLS),
    },
)

__all__ = [
    "ALL_TOOLS",
    "FS_TOOLS",
    "MEMORY_TOOLS",
    "RAG_TOOLS",
    "SEARCH_TOOLS",
    "all_tool_node",
    "coder_tool_node",
    "memory_tool_node",
    "search_tool_node",
]
