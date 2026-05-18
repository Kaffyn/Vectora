"""Coder Worker — LLM especializado em operações de código e filesystem.

Ferramentas disponíveis: file_read, file_edit, file_write, grep, list_dir, terminal
Objetivo: criar/editar arquivos, executar comandos, navegação de código
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from vectora.nodes.base import invoke_llm
from vectora.nodes.tools import FS_TOOLS, MEMORY_TOOLS
from vectora.services.utils import load_llm

if TYPE_CHECKING:
    from vectora.state import State

logger = logging.getLogger(__name__)

# Coder tem acesso a FS + memória (para lembrar de contexto do projeto)
_CODER_TOOLS = [*FS_TOOLS, *MEMORY_TOOLS]

_coder_llm = None


def _get_coder_llm() -> object:
    global _coder_llm
    if _coder_llm is None:
        if not _CODER_TOOLS:
            # Fallback se file_operations estiver desabilitado
            _coder_llm = load_llm()
            logger.warning("coder_worker: FS tools desabilitados, LLM sem ferramentas")
        else:
            _coder_llm = load_llm().bind_tools(_CODER_TOOLS)
            logger.debug(
                "coder_worker LLM inicializado com %d tools", len(_CODER_TOOLS)
            )
    return _coder_llm


async def coder_worker(state: State) -> dict:
    """Worker de código: cria/edita arquivos, executa terminal e git.

    Especializado em tarefas de desenvolvimento:
    - Ler e editar código-fonte
    - Executar comandos (git, npm, pip, terminal)
    - Criar estrutura de projeto
    - Grep e navegação em arquivos
    """
    logger.info("coder_worker: processando mensagem")
    return await invoke_llm(_get_coder_llm(), state)
