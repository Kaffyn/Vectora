# Protocolo de Integração: Paperclip ↔ Vectora

Este documento estabelece o **contrato oficial** de integração entre o ecossistema **Paperclip** (multi-agent framework) e o **Vectora** (RAG + LangGraph hub). É leitura obrigatória para qualquer Agente de IA ou desenvolvedor que vá implementar conectores, plugins ou clientes do Vectora dentro do Paperclip.

A integração tem como objetivo permitir que **N agentes do Paperclip** compartilhem **um único servidor Vectora** (com LanceDB e singleton AgentManager comuns), mantendo **sessões totalmente isoladas** via `thread_id`. Não há duplicação de banco de dados nem cópias de embeddings — apenas um hub central que serve múltiplos clientes de forma concorrente.

---

## 1. Princípios Arquiteturais

A integração Paperclip→Vectora é regida por cinco princípios não-negociáveis que garantem performance, isolamento e estabilidade do sistema.

### 1.1. Hub Centralizado (Single Source of Truth)

Existe **exatamente um** processo Vectora ativo por deployment. Ele atua como **hub cognitivo compartilhado** para todos os agentes do Paperclip, expondo ferramentas, recursos e a capacidade de delegação A2A (Agent-to-Agent).

O Vectora roda como container Docker autônomo (ver `docker-compose.yml` na raiz do projeto). Múltiplos agentes Paperclip se conectam ao mesmo container — não há um Vectora por agente, e isso é fundamental para evitar corrupção do LanceDB e duplicação de embeddings caros.

### 1.2. Sessão por `thread_id`, NÃO por instância

Cada agente Paperclip deve possuir um `thread_id` único e estável durante toda a sua vida útil. Este ID é a **chave de segregação cognitiva** dentro do Vectora — o LangGraph Checkpointer usa o `thread_id` para indexar memórias, histórico e estado de cada conversa dentro de **um único arquivo SQLite**.

Não é necessário (nem permitido) que cada agente possua seu próprio arquivo SQLite. O padrão Checkpointer já garante que sessões com `thread_id` diferentes não enxergam histórico umas das outras. Tentar criar SQLites separados é antipattern e quebra a arquitetura.

### 1.3. Cliente "Burro", Servidor "Inteligente"

O Paperclip atua como **cliente desinformado**: ele apenas envia tarefas em linguagem natural e o `thread_id`. Toda a lógica de raciocínio, decisão de ferramentas, RAG e síntese acontece dentro do Vectora.

Esta separação de responsabilidades evita acoplamento — o Paperclip não precisa saber quais tools existem, qual LLM está rodando ou como o RAG é executado. Ele simplesmente confia que o Vectora resolverá a tarefa.

### 1.4. Comunicação via MCP (Model Context Protocol)

Toda comunicação entre Paperclip e Vectora **deve** usar o protocolo MCP. Não é permitido criar APIs REST customizadas, gRPC paralelos ou qualquer outro protocolo. O MCP fornece tipagem, descoberta de capacidades e tratamento de erros padronizado.

O transport pode ser `stdio` (local, processo filho) ou `sse` (HTTP, multi-container). A escolha depende do deploy, mas o protocolo de aplicação é sempre o mesmo.

### 1.5. Async-first, Sempre

Toda interação com o Vectora é I/O-bound (rede, banco, LLM). Por isso, **todo cliente Paperclip deve ser implementado com `async/await`**. Chamadas síncronas bloqueariam o event loop do agente cliente e degradariam a performance do sistema como um todo.

O `VectoraProxy` (helper oficial em `vectora/mcp/proxy.py`) já implementa esse padrão. Usá-lo é a forma recomendada de integração.

---

## 2. Modos de Conexão

O Vectora suporta dois transports MCP, escolhidos via variável de ambiente `MCP_TRANSPORT`. A decisão entre os dois depende exclusivamente da topologia de deploy do Paperclip.

### 2.1. Modo `stdio` (Local)

Use este modo quando o Paperclip e o Vectora rodam no **mesmo host** (sem containers separados ou comunicação de rede). O Vectora é iniciado como processo filho do agente Paperclip via `uv run vectora-mcp`.

Este é o modo ideal para desenvolvimento local, testes unitários e cenários onde um único agente Paperclip precisa do Vectora exclusivamente. A latência é mínima (IPC via pipes), mas não há concorrência entre agentes — cada processo Paperclip teria seu próprio Vectora-filho.

```python
from vectora.mcp import create_local_proxy

async with create_local_proxy() as vectora:
    result = await vectora.delegate(
        task="Resuma o último PR do repo X",
        thread_id="paperclip_dev_machine_001",
    )
```

### 2.2. Modo `sse` (Remoto Multi-Agent)

Use este modo quando múltiplos agentes Paperclip (em containers, máquinas ou processos diferentes) precisam compartilhar **um único Vectora**. Este é o modo **canônico de produção** para o cenário multi-agent.

O Vectora roda em seu próprio container Docker, escutando em `MCP_HOST:MCP_PORT` (default `0.0.0.0:8000`). Cada agente Paperclip estabelece sua própria conexão HTTP/SSE, mas todos compartilham o mesmo LanceDB, AgentManager e SQLite. O isolamento entre agentes acontece exclusivamente pelo `thread_id`.

```python
from vectora.mcp import create_remote_proxy

VECTORA_URL = "http://vectora.internal:8000/sse"

async with create_remote_proxy(VECTORA_URL) as vectora:
    result = await vectora.delegate(
        task="Analise sentimentos das últimas 50 issues do GitHub",
        thread_id=f"paperclip_agent_{agent_id}",
    )
```

### 2.3. Como Subir o Vectora em Modo Remoto

A configuração de transport é feita via variáveis de ambiente no `docker-compose.yml`. Basta definir `MCP_TRANSPORT=sse` no `.env` do projeto Vectora e subir o container.

```bash
# .env (projeto Vectora)
MCP_TRANSPORT=sse
MCP_HOST=0.0.0.0
MCP_PORT=8000

# Subir o hub
docker compose up -d
```

Após isso, o endpoint `http://vectora:8000/sse` estará disponível para qualquer agente Paperclip que precise conectar.

---

## 3. Contrato do `thread_id`

O `thread_id` é o elemento mais importante deste protocolo, pois é ele que define **quem é cada agente** dentro do Vectora. Erros na geração ou no uso do `thread_id` resultam em vazamento de contexto entre agentes ou perda de memória.

### 3.1. Formato Recomendado

O Paperclip deve gerar `thread_id` no formato `paperclip_<agent_role>_<instance_id>`, onde `<agent_role>` identifica o tipo do agente (ex: `summarizer`, `researcher`, `coder`) e `<instance_id>` é único por instância (UUID, timestamp+random, ou hash de configuração).

Este formato facilita debugging (logs do Vectora mostram qual agente fez cada chamada) e garante unicidade global. Exemplo válido: `paperclip_summarizer_a1b2c3d4`.

### 3.2. Estabilidade Durante a Vida do Agente

O `thread_id` **deve ser persistido** pelo agente Paperclip durante toda a sua vida útil. Se um agente reiniciar e gerar um novo `thread_id`, ele perderá todo o histórico de conversa armazenado no Vectora — efetivamente "amnésia".

Recomenda-se armazenar o `thread_id` no estado do agente Paperclip (ex: arquivo de configuração, banco local, variável de ambiente) e reutilizá-lo após restarts.

### 3.3. Conversão para `int` no Vectora

Internamente, o Vectora atualmente usa `thread_id: int` em algumas APIs por compatibilidade com o LangGraph. O `VectoraProxy.delegate()` faz essa conversão automaticamente quando recebe uma string numérica, mas IDs alfanuméricos puros podem causar erro.

Workaround atual: use `hash(thread_id_str) & 0xFFFFFFFF` se precisar de um `int` estável a partir de uma string. Roadmap futuro: migrar todas as APIs para `thread_id: str` (issue aberta).

---

## 4. APIs Disponíveis via `VectoraProxy`

O `VectoraProxy` expõe quatro famílias de operações que cobrem 100% das necessidades de um agente Paperclip. Cada uma tem semântica e custo diferentes — escolher a operação correta para cada tarefa é crucial para performance.

### 4.1. `delegate(task, thread_id)` — Delegação A2A

Use para tarefas **complexas** que exigem raciocínio, múltiplas ferramentas ou síntese de informações. O Vectora roda seu LangGraph interno completo, decide quais tools chamar, executa RAG/busca/análise, e retorna o resultado final.

Esta é a operação mais poderosa, mas também a mais cara (pode levar até 5 minutos). Use quando o Paperclip não souber decompor a tarefa em chamadas atômicas, ou quando quiser delegar 100% do problem-solving.

```python
result = await vectora.delegate(
    task="Pesquise as 5 melhores práticas de RAG em 2025 e resuma em bullets",
    thread_id="paperclip_researcher_42",
)
```

### 4.2. `call_tool(name, args)` — Ferramenta Atômica

Use quando o Paperclip **já sabe** qual ferramenta invocar. É muito mais rápido que `delegate()` pois não roda LangGraph completo — apenas executa a tool diretamente.

Esta é a operação preferida quando o agente Paperclip tem lógica de decisão própria e precisa apenas executar uma operação específica (ex: ler um arquivo, fazer uma busca vetorial, indexar um documento).

```python
docs = await vectora.call_tool(
    "vector_search_tool",
    {"query": "RAG patterns", "collection": "docs", "limit": 5},
)
```

### 4.3. `get_thread_context(thread_id)` / `get_thread_history(thread_id)` — Resources

Use para **inspecionar o estado** da sessão sem invocar o LLM. Retorna metadados, sumário e mensagens recentes — útil para debugging, dashboards ou para o agente Paperclip decidir se já tem contexto suficiente.

Esta operação é praticamente gratuita (leitura direta do SQLite) e idempotente. Pode ser chamada com frequência sem impacto no servidor.

```python
context = await vectora.get_thread_context("paperclip_agent_42")
print(f"Mensagens nesta sessão: {context['message_count']}")
```

### 4.4. `list_tools()` — Descoberta de Capacidades

Use na inicialização do agente Paperclip para **descobrir dinamicamente** quais tools o Vectora expõe. Isso permite que o cliente se adapte a novas versões do Vectora sem mudanças de código.

Não chame esta API a cada operação — o resultado é estável dentro de uma sessão. Cacheie no início e revalide apenas se houver erro de "tool not found".

```python
tools = await vectora.list_tools()
for tool in tools:
    print(f"- {tool['name']}: {tool['description']}")
```

---

## 5. Padrões de Uso (Receitas)

A seguir estão os padrões recomendados para cenários comuns no Paperclip. Use-os como ponto de partida e adapte conforme a necessidade do seu agente.

### 5.1. Agente Persistente com Memória

Cenário: um agente Paperclip que mantém memória de longo prazo entre execuções (ex: assistente pessoal, code reviewer com histórico).

O agente deve usar um `thread_id` estável armazenado no estado local, abrir a conexão na inicialização e mantê-la durante toda a sessão. Memórias são automaticamente persistidas pelo Checkpointer do Vectora.

```python
class PersistentPaperclipAgent:
    def __init__(self, agent_id: str, vectora_url: str):
        self.thread_id = f"paperclip_persistent_{agent_id}"
        self.vectora_url = vectora_url
        self._proxy = None

    async def __aenter__(self):
        from vectora.mcp import create_remote_proxy
        self._proxy = create_remote_proxy(self.vectora_url)
        await self._proxy.connect()
        return self

    async def __aexit__(self, *args):
        await self._proxy.disconnect()

    async def ask(self, question: str) -> str:
        return await self._proxy.delegate(
            task=question,
            thread_id=self.thread_id,
        )
```

### 5.2. Pool de Agentes Concorrentes

Cenário: vários agentes Paperclip rodando em paralelo, cada um processando uma fila de tarefas independentes.

Cada agente do pool deve ter seu próprio `thread_id` único. As conexões podem ser criadas sob demanda (curtas) ou mantidas em pool (longas) dependendo do padrão de tráfego.

```python
import asyncio
from vectora.mcp import create_remote_proxy

async def process_task(agent_id: int, task: str):
    async with create_remote_proxy("http://vectora:8000/sse") as vectora:
        return await vectora.delegate(
            task=task,
            thread_id=f"paperclip_worker_{agent_id}",
        )

# 10 agentes em paralelo, cada um com sua sessão isolada
results = await asyncio.gather(*[
    process_task(i, f"Tarefa {i}") for i in range(10)
])
```

### 5.3. Agente Híbrido (call_tool + delegate)

Cenário: um agente Paperclip sofisticado que tem sua própria lógica de decisão, mas delega tarefas pesadas ao Vectora.

Use `call_tool()` para operações que o Paperclip já sabe executar e `delegate()` apenas quando a tarefa exigir raciocínio que ele não consegue (ou não quer) fazer localmente.

```python
async with create_remote_proxy(VECTORA_URL) as vectora:
    # Paperclip decide buscar no RAG primeiro (operação atômica)
    docs = await vectora.call_tool(
        "vector_search_tool",
        {"query": user_question, "limit": 5},
    )

    if not enough_context(docs):
        # Sem contexto local, delega análise completa ao Vectora
        answer = await vectora.delegate(
            task=f"Responda: {user_question}. Faça busca web se necessário.",
            thread_id=session_id,
        )
    else:
        # Tem contexto, Paperclip processa localmente
        answer = local_llm.generate(user_question, docs)
```

---

## 6. Tratamento de Erros

Erros na integração Paperclip→Vectora devem ser tratados de forma defensiva. O `VectoraProxy` levanta `VectoraProxyError` para todas as falhas do protocolo, e cabe ao agente Paperclip decidir entre retry, fallback ou propagação.

### 6.1. Timeouts

O Vectora aplica timeouts em camadas: cada ferramenta individual tem seu próprio limite (10-120s), e a delegação A2A tem timeout global de 300s (5 minutos). Quando esses limites são excedidos, a resposta retornada já contém a mensagem de erro formatada — não há exception.

O agente Paperclip deve sempre verificar se o resultado começa com `"Erro:"` ou `"Error:"` para detectar timeouts e decidir se faz retry com tarefa simplificada. Não capturar `TimeoutError` na camada do proxy — ele já é tratado internamente pelo Vectora.

### 6.2. Falhas de Conexão

Erros de rede (servidor offline, DNS, etc.) levantam `VectoraProxyError` durante `proxy.connect()`. O agente Paperclip deve implementar retry exponencial e, se o erro persistir, registrar logs claros e desativar a integração temporariamente.

Não tente reconectar dentro do mesmo `async with` — destrua o proxy e crie um novo na próxima tentativa. O `AsyncExitStack` interno garante limpeza correta de recursos.

### 6.3. Validação de Resposta

Mesmo quando não há erro, valide o conteúdo da resposta antes de usá-la. O Vectora pode retornar mensagens como `"Erro: Ferramenta 'X' excedeu timeout..."` mesmo em status HTTP 200 (são erros de aplicação, não de protocolo).

A regra é: sempre tratar a string retornada como input não-confiável. Se o agente Paperclip espera JSON, parse com try/except. Se espera texto, verifique se não começa com prefixos de erro conhecidos.

---

## 7. Observabilidade

A operação multi-agent só é gerenciável se houver visibilidade do que cada agente está fazendo. O Vectora já emite logs estruturados, e o Paperclip deve complementá-los com correlação por `thread_id`.

### 7.1. Logs do Vectora

Todos os logs do Vectora vão para `~/.vectora/logs/mcp.log` (dentro do container, mapeado para volume). Cada entrada inclui o `thread_id` em `extra`, permitindo filtrar logs por agente Paperclip específico.

Para acompanhar logs em tempo real:

```bash
docker compose logs -f vectora
# ou
docker exec vectora tail -f /root/.vectora/logs/mcp.log
```

### 7.2. Métricas Recomendadas no Paperclip

O agente Paperclip deve emitir métricas próprias para cada interação com o Vectora, no mínimo: latência da chamada, tamanho do payload, taxa de erro e operação (`delegate` vs `call_tool`).

Essas métricas, combinadas com os logs do Vectora, permitem identificar gargalos (ex: um agente específico fazendo muitos `delegate()` desnecessários) e tomar ações corretivas.

### 7.3. LangSmith (opcional)

Se a variável `LANGSMITH_API_KEY` estiver configurada no Vectora, todas as execuções do LangGraph interno são automaticamente rastreadas no LangSmith. Isso fornece **traces completos** de cada delegação — útil para debugar comportamentos inesperados do agente.

O Paperclip não precisa configurar nada para usar isso, basta verificar o dashboard do LangSmith filtrando por `thread_id`.

---

## 8. Roadmap da Integração

Esta seção lista melhorias planejadas para a integração Paperclip↔Vectora, em ordem de prioridade. Contribuições e issues são bem-vindas no repositório do Vectora.

### 8.1. Plugin Oficial Paperclip (futuro)

Próximo passo: criar um plugin nativo do Paperclip dentro desta mesma pasta (`integrations/paperclip/plugin/`). O plugin abstrairá a configuração do `VectoraProxy`, fornecerá decoradores de alto nível e exporá uma API ergonômica seguindo as convenções do Paperclip.

Não há ETA — depende da maturidade do plugin system do Paperclip e do feedback dos primeiros usuários do `VectoraProxy`.

### 8.2. `thread_id: str` Nativo

Atualmente, IDs alfanuméricos precisam de workaround (`hash`). A migração para `str` nativo está planejada e simplificará a geração de `thread_id` no Paperclip — basta concatenar strings sem se preocupar com conversões.

### 8.3. Streaming de Respostas

Hoje, `delegate()` retorna a resposta completa após o LangGraph finalizar (pode levar minutos). Streaming via SSE permitirá que o Paperclip receba a resposta em tempo real, melhorando UX e detectando travamentos mais cedo.

### 8.4. Health Check e Auto-Discovery

Adicionar endpoint `/health` no Vectora SSE permite que orchestrators (Kubernetes, Docker Swarm) façam health checks reais. Combinado com mDNS ou service registry, agentes Paperclip podem descobrir automaticamente o endpoint do Vectora sem hardcode.

---

## 9. Referências

Documentação relacionada para aprofundamento técnico nos componentes citados neste protocolo.

- **Vectora** — Código fonte do servidor MCP: `vectora/mcp/server.py`
- **VectoraProxy** — Cliente oficial: `vectora/mcp/proxy.py`
- **Testes A2A** — Suite de validação: `tests/integration/test_a2a_integration.py`
- **MCP Protocol** — Especificação oficial: https://modelcontextprotocol.io
- **LangGraph Checkpointer** — Padrão de persistência: https://langchain-ai.github.io/langgraph/concepts/persistence/
- **AGENTS.md do projeto Vectora** — Regras gerais de desenvolvimento: `/AGENTS.md`

---

## 10. Changelog Deste Protocolo

Toda mudança neste documento deve ser versionada com commit `docs:` seguindo Conventional Commits, conforme regra do `AGENTS.md` do projeto Vectora.

- **v1.0.0** (2026-05-17) — Versão inicial. Estabelece arquitetura multi-agent, contrato de `thread_id`, modos `stdio`/`sse` e API do `VectoraProxy`.
