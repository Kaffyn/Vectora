# Pós MVP

Esta é uma transição de um **agente de RAG local** para um **orquestrador de sistemas de agentes**. O conceito de _Harness_ (que você encontra em bibliotecas como `langgraph-checkpoint` e `langgraph-pregel`) é o coração dessa evolução.

Aqui está o plano estratégico para a evolução do Vectora (v0.2 a v0.4).

---

### 1. O que é o "Harness" dos Deep Agents?

No contexto de sistemas como os da LangChain, o **Harness** é o _runtime_ que isola o grafo do agente do mundo externo.

- **O que ele faz:** Gerencia a persistência (checkpoints), o streaming de eventos, a injeção de prompts de sistema, e a execução de ferramentas em um ambiente controlado.
- **O que você vai trocar:** Hoje, você tem um `run_chat.py` que faz a orquestração manual. No futuro, você usará o `GraphRunnable` ou `Pregel` para que o grafo "rode sozinho" enquanto sua CLI apenas "escuta" os eventos.

---

### 2. Plano Pós-MVP (Roadmap v0.2 - v0.4)

#### **v0.2: O "Agentic Core" (Pattern: Multi-Agent Architectures)**

- **Design Pattern:** **Supervisor/Worker Pattern**.
- **O que muda:** O `MAIN_NODE` deixa de ser o "faz-tudo". Ele vira um `Supervisor` que delega tarefas para agentes especialistas (`SearchAgent`, `CoderAgent`, `AnalysisAgent`).
- **Integração Tavily:** O Tavily substitui o `DuckDuckGo` + `WebBaseLoader`. O Tavily já retorna o contexto formatado para RAG, eliminando a necessidade de `fetch_url` manual na maioria dos casos.
- **ACP:** Implementação do servidor ACP para que o **Zed/Neovim** possa enviar comandos para o agente (ex: "corrija este erro no arquivo aberto").

#### **v0.3: O "Memory & Human-in-the-loop" (Pattern: Feedback Loops)**

- **Design Pattern:** **Human-in-the-loop (HITL)** e **Memory-as-a-Service**.
- **O que muda:** LangGraph permite pausas (`interrupt_before`). Antes de executar um `file_edit` ou `terminal`, o agente fará um `interrupt`. A CLI perguntará ao usuário: _"O agente quer deletar `data/`. Confirmar?"_.
- **Memória:** Implementação de `LongTermMemory` usando `LanceDB` para criar perfis de usuário persistentes que o agente "lê" no início de cada sessão.

#### **v0.4: "Deep Agents" (Pattern: Recursive Reasoning)**

- **Design Pattern:** **Reflection / Self-Correction**.
- **O que muda:** O agente terá um nó de "crítico". Se a resposta do LLM parecer fraca, o agente volta para o `call_llm` com uma nova instrução: "Você falhou. Critique sua resposta anterior e tente de novo". Isso reduz drasticamente alucinações.

---

### 3. Integração de Tecnologias

#### **Tavily (Substituição de Tools)**

Sim, o **Tavily** substitui suas `web_search` e `fetch_url`. Ele foi treinado para RAG. Ele retorna:

1. Contexto resumido.
2. URLs citadas.
3. Snippet de texto pronto para embedding.

- **Resultado:** Seu `PERSISTENCE_NODE` fica 50% menor.

#### **Sub-Agents e Streaming**

- **Streaming:** Usaremos `astream_events` do LangGraph. A TUI não esperará o LLM terminar. A cada token gerado, a interface atualiza.
- **Sub-Agents:** Você terá um `router.py` que não envia a tarefa para um `node` qualquer, mas sim invoca um `SubGraph` compilado. O sub-agente tem seu próprio State.

---

### 4. Estrutura do Projeto (v0.2+)

```text
src/
├── agents/            # Os "Deep Agents" especialistas
│   ├── supervisor.py  # Orquestrador
│   ├── researcher.py  # RAG/Tavily
│   └── coder.py       # File Ops/Terminal
├── memory/            # Gerenciamento de memória persistente
├── protocol/          # ACP (Agent Client Protocol) e MCP
├── runtime/           # O Harness (Pregel/Pregel-lite)
└── state.py           # Estado compartilhado entre agentes
```

---

### 5. Respostas rápidas às suas dúvidas

- **O que será trocado?** O `MAIN_NODE` monolítico será trocado pelo `Supervisor`. O `WebSearch` e `Fetch` serão trocados pela API do Tavily.
- **Quais Patterns?** Supervisor-Worker (delegação), Actor-based (agentes são atores), e Reflection (Auto-crítica).
- **ACP Integration:** O ACP rodará como um **servidor paralelo ao MCP**. O MCP é para o "cérebro" (Claude Code), o ACP é para o "corpo" (seu editor de texto). O Vectora será um servidor dual: `vcp-mcp` + `vcp-acp`.

### O que você deve fazer agora?

Não tente fazer o `v0.2` sem o MVP `0.1` estar rodando perfeitamente.

**Minha sugestão para você:**
Se você quer o **Deep Agents**, o segredo não é o código agora, é o **"Supervisor"**. Comece a pensar como o supervisor: "Se eu recebo uma tarefa de código, eu não uso o LLM para codar, eu chamo o `CoderAgent`".

Deseja que eu monte um **esboço de um Supervisor Node** simples que já comece a usar essa lógica de delegação?
