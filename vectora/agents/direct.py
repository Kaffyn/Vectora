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

from vectora.agents._identity import VECTORA_IDENTITY
from vectora.nodes.base import invoke_llm
from vectora.nodes.tools import MEMORY_TOOLS
from vectora.services.utils import load_llm

if TYPE_CHECKING:
    from vectora.state import State

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = f"""{VECTORA_IDENTITY}

---

## Seu Papel — Direct Agent

Você é o **Direct Agent** do Vectora. Entrega respostas diretas sem ferramentas de busca
ou filesystem.

### Quando você é acionado
- Saudações, agradecimentos e conversas simples
- Perguntas sobre o que o Vectora é ou faz
- Síntese final após o pipeline RAG injetar contexto no histórico
- Respostas de conhecimento geral

### Contexto RAG
Quando houver `## Contexto Recuperado (RAG)` no histórico, **priorize-o** e cite fontes
usando `[N]`.

### Memória
Use `save_memory` / `get_memory` quando o usuário pedir para lembrar algo relevante
entre sessões.

### Estilo
- Conciso e direto, sem introduções desnecessárias
- Markdown para respostas estruturadas
- Adapte o idioma ao da conversa
"""

_direct_llm = None


def _get_direct_llm() -> object:
    global _direct_llm
    if _direct_llm is None:
        _direct_llm = load_llm().bind_tools(MEMORY_TOOLS)
        logger.debug("direct_worker LLM inicializado (memory tools only)")
    return _direct_llm


async def direct(state: State) -> dict:
    """Agent direto: responde sem ferramentas de busca ou filesystem.

    Casos de uso principais:
    - Saudações e conversas simples
    - Síntese de resultados RAG já presentes em state['rag_docs']
    - Perguntas de conhecimento geral
    - Output final consolidado após pipeline RAG
    """
    logger.info("direct: processando mensagem")
    return await invoke_llm(_get_direct_llm(), state, system_prompt=SYSTEM_PROMPT)
