"""Search Worker — LLM especializado em busca web e RAG.

Ferramentas disponíveis: web_search, fetch_url, vector_search, embedding
Objetivo: pesquisar informações atuais + consultar base indexada
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from vectora.agents._identity import VECTORA_IDENTITY
from vectora.nodes.base import invoke_llm
from vectora.nodes.tools import SEARCH_TOOLS
from vectora.services.utils import load_llm

if TYPE_CHECKING:
    from vectora.state import State

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = f"""{VECTORA_IDENTITY}

---

## Seu Papel — Search Agent

Você é o **Search Agent** do Vectora. Especializado em pesquisa e recuperação de informação.

### Ferramentas disponíveis
- `web_search` — busca web em tempo real via Tavily
- `fetch_url` — extrai conteúdo de uma URL específica
- `vector_search` — busca semântica na base indexada (LanceDB)
- `embedding` — enfileira documento para indexação assíncrona

### Estratégia RAG-first
1. **Prefira `vector_search`** se o tema já foi pesquisado antes — é instantâneo (local)
2. Use `web_search` para informações atuais ou não indexadas
3. Após `web_search` ou `fetch_url`, o `process_retrieval` faz cascading automático
   para LanceDB — **não chame `embedding` manualmente** depois de uma busca web

### Fire-and-forget embedding
Quando `embedding` retornar `"status": "fire_and_forget"`, os docs foram **enfileirados**,
não indexados ainda. Informe o usuário: use `/rag` para acompanhar o progresso.

### Estilo
- Cite fontes com URL ou título
- Indique qual ferramenta usou e por quê
- Adapte o idioma ao da conversa
"""

_search_llm = None


def _get_search_llm() -> object:
    global _search_llm
    if _search_llm is None:
        _search_llm = load_llm().bind_tools(SEARCH_TOOLS)
        logger.debug("search_worker LLM inicializado com %d tools", len(SEARCH_TOOLS))
    return _search_llm


async def search(state: State) -> dict:
    """Agent de busca: responde usando web_search, fetch_url e vector_search.

    O LLM decide autonomamente quais ferramentas usar com base na pergunta.
    Após as ferramentas executarem (via search_tools node), o resultado é
    processado pelo process_retrieval para cascading automático no LanceDB.
    """
    logger.info("search: processando mensagem")
    return await invoke_llm(_get_search_llm(), state, system_prompt=SYSTEM_PROMPT)
