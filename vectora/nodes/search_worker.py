"""Search Worker — LLM especializado em busca web e RAG.

Ferramentas disponíveis: web_search, fetch_url, vector_search, embedding
Objetivo: pesquisar informações atuais + consultar base indexada
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from vectora.nodes.base import invoke_llm
from vectora.nodes.tools import SEARCH_TOOLS
from vectora.services.utils import load_llm

if TYPE_CHECKING:
    from vectora.state import State

logger = logging.getLogger(__name__)

# LLM bindado apenas com SEARCH_TOOLS — inicializado uma vez por processo
_search_llm = None


def _get_search_llm() -> object:
    global _search_llm
    if _search_llm is None:
        _search_llm = load_llm().bind_tools(SEARCH_TOOLS)
        logger.debug("search_worker LLM inicializado com %d tools", len(SEARCH_TOOLS))
    return _search_llm


async def search_worker(state: State) -> dict:
    """Worker de busca: responde usando web_search, fetch_url e vector_search.

    O LLM decide autonomamente quais ferramentas usar com base na pergunta.
    Após as ferramentas executarem (via search_tool_node), o resultado é
    processado pelo process_retrieval para cascading automático no LanceDB.
    """
    logger.info("search_worker: processando mensagem")
    return await invoke_llm(_get_search_llm(), state)
