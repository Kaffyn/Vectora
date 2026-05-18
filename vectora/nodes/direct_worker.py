"""Direct Worker — LLM sem ferramentas para respostas diretas.

Usado para:
- Saudações e meta-perguntas simples
- Síntese de contexto RAG já injetado no State
- Respostas que não requerem ferramentas externas
- Output final após outros workers completarem seu trabalho
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from vectora.nodes.base import invoke_llm
from vectora.nodes.tools import MEMORY_TOOLS
from vectora.services.utils import load_llm

if TYPE_CHECKING:
    from vectora.state import State

logger = logging.getLogger(__name__)

# Direct worker tem acesso apenas a memória (para personalização)
_direct_llm = None


def _get_direct_llm() -> object:
    global _direct_llm
    if _direct_llm is None:
        _direct_llm = load_llm().bind_tools(MEMORY_TOOLS)
        logger.debug("direct_worker LLM inicializado (memory tools only)")
    return _direct_llm


async def direct_worker(state: State) -> dict:
    """Worker direto: responde sem ferramentas de busca ou filesystem.

    Casos de uso principais:
    - Saudações e conversas simples
    - Síntese de resultados RAG já presentes em state['rag_docs']
    - Perguntas de conhecimento geral
    - Output final consolidado após pipeline RAG
    """
    logger.info("direct_worker: processando mensagem")
    return await invoke_llm(_get_direct_llm(), state)
