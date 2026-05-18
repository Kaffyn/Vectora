"""Coder Worker — LLM especializado em operações de código e filesystem.

Ferramentas disponíveis: file_read, file_edit, file_write, grep, list_dir, terminal
Objetivo: criar/editar arquivos, executar comandos, navegação de código
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from vectora.agents._identity import VECTORA_IDENTITY
from vectora.nodes.base import invoke_llm
from vectora.nodes.tools import FS_TOOLS, MEMORY_TOOLS
from vectora.services.utils import load_llm

if TYPE_CHECKING:
    from langchain_core.runnables import Runnable

    from vectora.state import State

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = f"""{VECTORA_IDENTITY}

---

## Seu Papel — Coder Agent

Você é o **Coder Agent** do Vectora. Especializado em desenvolvimento de software e
operações de filesystem.

### Ferramentas disponíveis
- `file_read`, `file_edit`, `file_write` — ler, editar e criar arquivos
- `grep` — busca em código por padrões/regex
- `list_dir` — listar diretórios
- `terminal` — executar comandos shell (git, npm, pip, uv, docker, pytest...)
- `save_memory`, `get_memory` — lembrar contexto do projeto entre sessões

### Git e terminal são livres
Execute qualquer subcomando git (`git status`, `git add`, `git commit`, `git push`,
`git log`, `git diff`...) **sem pedir confirmação ao usuário**. Git é essencial para
desenvolvimento. Apenas `rm -rf`, `mkfs` e equivalentes destrutivos são bloqueados
automaticamente pela tool.

### Proatividade
- Ao criar ou editar código, execute testes automaticamente se existirem
- Use `grep` para navegar no código antes de editar
- Prefira edições cirúrgicas (`file_edit`) a reescritas completas (`file_write`)

### Estilo
- Mostre o código gerado ou editado no resultado
- Explique brevemente o que foi feito e por quê
- Adapte o idioma ao da conversa
"""

# Coder tem acesso a FS + memória (para lembrar de contexto do projeto)
_CODER_TOOLS = [*FS_TOOLS, *MEMORY_TOOLS]

_coder_llm = None


def _get_coder_llm() -> Runnable:
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


async def coder(state: State) -> dict:
    """Agent de código: cria/edita arquivos, executa terminal e git.

    Especializado em tarefas de desenvolvimento:
    - Ler e editar código-fonte
    - Executar comandos (git, npm, pip, terminal)
    - Criar estrutura de projeto
    - Grep e navegação em arquivos
    """
    logger.info("coder: processando mensagem")
    return await invoke_llm(_get_coder_llm(), state, system_prompt=SYSTEM_PROMPT)
