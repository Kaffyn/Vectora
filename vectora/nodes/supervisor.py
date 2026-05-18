"""Supervisor — Decide qual worker deve processar a mensagem atual.

Substitui o router.py com lógica mais rica: além de classificar a intenção,
o supervisor também decide se delegar de volta ao supervisor após um worker
terminar (ex: search_worker fez busca → supervisor decide injetar RAG antes
de responder via direct_worker).

Rota de saída:
  "search_worker"  → busca web + RAG
  "coder_worker"   → filesystem, terminal, git
  "direct_worker"  → resposta direta sem ferramentas
  "rag_subgraph"   → pipeline RAG completo (retrieve→rerank→inject)
"""

from __future__ import annotations

import logging
import re
from typing import TYPE_CHECKING

from langchain_core.messages import HumanMessage
from langgraph.types import Command

if TYPE_CHECKING:
    from vectora.state import State

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Padrões de classificação
# ---------------------------------------------------------------------------

_DIRECT_PATTERNS = re.compile(
    r"^("
    r"oi|olá|ola|hey|hi|hello|tudo bem|tudo bom|como vai|bom dia|boa tarde|boa noite"
    r"|obrigad[oa]|valeu|ok|okay|certo|entendido|perfeito|ótimo|legal|show"
    r"|quem (é você|es você|você é)|o que (você é|você faz|é o vectora)"
    r"|me apresente|apresentação|sobre você"
    r"|thanks|thank you|got it|understood|great|nice|cool"
    r"|who are you|what are you|what can you do"
    r")[!?.,]*$",
    re.IGNORECASE,
)

_CODER_PATTERNS = re.compile(
    r"\b("
    r"cria(r)?|escreve(r)?|edita(r)?|salva(r)?|apaga(r)?|deleta(r)?|move(r)?"
    r"|arquivo|pasta|diretório|directório|código|script|função|classe|módulo"
    r"|terminal|comando|executa(r)?|roda(r)?|instala(r)?|compila(r)?|build"
    r"|git|npm|pip|uv|docker|make|pytest|uvicorn|poetry"
    r"|file|folder|directory|code|function|class|module|run|execute|install|compile"
    r"|create file|write file|edit file|delete file|read file"
    r")\b",
    re.IGNORECASE,
)

_SEARCH_PATTERNS = re.compile(
    r"\b("
    r"busca(r)? na web|pesquisa(r)? na internet|procura(r)? online"
    r"|search|google|notícia(s)?|news|atualidade(s)?"
    r"|o que (é|aconteceu|está acontecendo)|quem (é|foi|inventou)"
    r"|quando (foi|aconteceu)|onde (fica|está)|como (funciona|fazer)"
    r"|acessa(r)? url|abre(r)? link|fetch url|download page"
    r"|search the web|look up|find out|what is|who is|when did|where is"
    r")\b",
    re.IGNORECASE,
)

_RAG_PATTERNS = re.compile(
    r"\b("
    r"documento(s)?|doc(s)?|wiki|base de conhecimento|knowledge base"
    r"|indexad[oa]|embeddings?|lancedb|vectora"
    r"|o que (diz|está escrito|consta)|segundo o(s)? documento(s)?"
    r"|com base no(s)?|de acordo com|conforme o(s)?"
    r"|na documentação|no manual|no guia|no relatório|no projeto"
    r"|document(s)?|indexed|knowledge base|according to|based on"
    r"|in the docs|in the manual|in the guide|in the report"
    r")\b",
    re.IGNORECASE,
)


def classify_intent(text: str) -> str:
    """Classifica intenção da mensagem em: direct | coder | search | rag.

    Prioridade:
      1. direct  — saudações e meta-perguntas curtas
      2. coder   — filesystem, terminal, código, git
      3. search  — busca web explícita
      4. rag     — consulta a documentação / base indexada
      5. rag     — fallback para mensagens longas (> 30 chars)
      6. direct  — fallback final
    """
    stripped = text.strip()

    if _DIRECT_PATTERNS.match(stripped):
        return "direct"
    if _CODER_PATTERNS.search(stripped):
        return "coder"
    if _SEARCH_PATTERNS.search(stripped):
        return "search"
    if _RAG_PATTERNS.search(stripped):
        return "rag"
    if len(stripped) > 30:
        return "rag"
    return "direct"


_WORKER_MAP = {
    "direct": "direct_worker",
    "coder": "coder_worker",
    "search": "search_worker",
    "rag": "rag_subgraph",
}


async def supervisor(state: State) -> Command:
    """Nó supervisor: classifica a intenção e roteia para o worker correto.

    Returns:
        Command com goto = worker alvo e routing_decision atualizado no State.
    """
    messages = state.get("messages", [])

    last_human_text = ""
    for msg in reversed(messages):
        if isinstance(msg, HumanMessage):
            last_human_text = str(msg.content)
            break

    intent = classify_intent(last_human_text)
    worker = _WORKER_MAP[intent]

    logger.info(
        "Supervisor: '%s...' → %s (%s)",
        last_human_text[:60],
        worker,
        intent,
    )

    return Command(
        goto=worker,
        update={"routing_decision": intent},  # type: ignore[arg-type]
    )
