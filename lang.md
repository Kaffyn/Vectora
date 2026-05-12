# LangChain, LangGraph e Deep Agents

A arquitetura de agentes do Vectora é construída sob uma base robusta e de última geração, aproveitando ecossistemas consolidados para orquestração de LLMs e workflows complexos. Esta stack nos permite ir muito além de simples chamadas de API (chatbots básicos), habilitando agentes com memória, autonomia, capacidade de uso de ferramentas e reflexão profunda.

## LangChain

O **LangChain** atua como o framework base para integrações com os Modelos de Linguagem de Grande Escala (LLMs). No Vectora, ele é responsável pelas abstrações de baixo nível, como:

- Padronização das chamadas aos diferentes provedores de IA.
- Gerenciamento de prompts estruturados e conversão de saída (Output Parsers).
- Definição estruturada de ferramentas (Tools) que o agente pode invocar (como busca no banco vetorial e acesso local).

O uso do LangChain simplifica o ecossistema, permitindo que a aplicação seja _model-agnostic_ (agnóstica a modelo), facilitando a troca entre modelos locais e serviços de nuvem de acordo com a preferência do usuário.

## LangGraph

Enquanto o LangChain cuida das integrações básicas, o **LangGraph** atua como o orquestrador do "cérebro" do Vectora. Diferente de fluxos lineares, o LangGraph permite construir aplicações baseadas em grafos, permitindo ciclos, recursividade e _loops_.

Isso é um pilar essencial para os Agentes do Vectora, pois permite:

- **Fluxos de Raciocínio Complexos:** O agente pode observar, planejar, tentar uma ação, verificar o resultado e iterar caso tenha falhado (ciclos ReAct e reflexão).
- **Gerenciamento de Estado (State):** O estado da conversa e as variáveis temporárias fluem de maneira previsível através das arestas do grafo.
- **Memória de Longo Prazo Resiliente:** Através do mecanismo nativo de _checkpoint_, o estado de cada etapa do grafo é persistido (ex: no PostgreSQL via `langgraph-checkpoint-postgres`). Isso oferece tolerância a falhas e permite _time-travel_, habilitando que tarefas longas sejam pausadas e retomadas a qualquer momento.

## Deep Agents

Desenvolvido com inspiração na equipe do LangChain / LangGraph, o projeto apresenta capacidades profundas de agentes:

- **ACP (Agent Client Protocol):** Padroniza a integração entre o agente e IDEs (Zed, JetBrains, Neovim) para edições de código contextuais.
- **Harness:** SDK com ferramentas nativas de planejamento (TODOs), sistema de arquivos e delegação para tarefas de longa duração.
- **CLI:** Interface de terminal para execução local de pesquisas, automações e assistência de codificação.
- **Skills:** Injeção dinâmica de instruções apenas quando necessário, otimizando a janela de contexto.
- **Context Management:** Middleware para sumarização de histórico e offloading de dados pesados para o disco.
- **Deep Research:** Ciclos iterativos de busca e análise profunda para síntese de relatórios complexos.
- **Autonomia:** Orquestração via subagentes especializados, isolando contextos técnicos do objetivo principal.
