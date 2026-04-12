# Vectora

> [!TIP]
> Read this file in another language | Leia esse arquivo em outro idioma.
> [English](README.md) | [Português](README.pt.md)

> [!NOTE]
>
> **RAG híbrido para codebases reais.**
> Entenda seu código como um sistema — não como texto.

---

**RAG tradicional opera em chunks isolados.
O Vectora opera sobre relações reais do sistema.**

- busca semântica (embeddings)
- estrutura de código (arquivos, funções, dependências)
- grafo da codebase
- raciocínio multi-hop

Resultado: respostas baseadas em como o sistema realmente funciona — não apenas em trechos isolados.

---

- **RAG Híbrido (Core):** Recuperação semântica + estrutural + relacional
- **Codebase-Aware:** Entende relações reais entre módulos e arquivos
- **Multi-hop Real:** Navega múltiplos pontos do sistema para responder perguntas complexas
- **Execução Agêntica:** Atua sobre o código com contexto e segurança
- **Privacidade Total:** Dados e embeddings permanecem locais

---

## Por que Vectora?

A maioria das soluções exige combinar múltiplas ferramentas:

- vector database
- framework de RAG
- embeddings
- orchestration

O Vectora integra tudo isso em um único core local, pronto para uso.

---

## Operação Agêntica: Sub-Agente Especialista para sua IDE

O Vectora atua como um **Sub-Agente (Tier 2)** acoplado à sua IDE, projetado para executar tarefas complexas com base em entendimento real da codebase.

Enquanto o agente principal (Tier 1) orquestra a interação com o usuário, o Vectora assume tarefas que exigem **recuperação profunda, análise estrutural e execução controlada**.

- **Delegação de Alto Nível:**
  A IDE pode delegar operações como refatorações multi-arquivo, análise de impacto e navegação entre dependências. O Vectora executa com base no estado real do código.

- **Execução com Consciência de Contexto:**
  Todas as ações são guiadas por recuperação contextual (RAG), garantindo que leituras, escritas e comandos sejam baseados na estrutura e nas relações reais do projeto.

- **Execução Restrita ao Escopo:**
  Ferramentas operam exclusivamente dentro do Trust Folder, garantindo previsibilidade e isolamento.

- **Segurança Transacional (Git Snapshots):**
  Modificações no código são precedidas por snapshots automáticos, permitindo reversão imediata.

---

## Motor de Recuperação: Entendimento Sistêmico da Codebase

RAG tradicional opera em chunks isolados.
O Vectora recupera **contexto conectado**.

- **Camadas de Recuperação:**
  Combina busca semântica (embeddings) com análise estrutural do código (arquivos, funções, módulos e dependências).

- **Grafo da Codebase:**
  O projeto é modelado como um grafo de relações entre entidades, permitindo navegação além de fronteiras de arquivo.

- **Multi-hop Reasoning:**
  Consultas atravessam múltiplos pontos do sistema, conectando dependências e fluxos de execução para responder perguntas que exigem contexto global.

---

## Instalação e Integração

O Vectora foi projetado para se adaptar ao seu workflow:

**1. Binário Portátil (Manual & MCP Server):**

Ideal para controle total ou uso fora de IDEs.

- **Download:** Baixe o `vectora.exe` na página de [Releases](https://github.com/Kaffyn/Vectora/releases).
- **Modo CLI:** Interface interativa para indexação e buscas semânticas rápidas via terminal.
- **Modo MCP Server:** Configure o caminho do binário como um servidor MCP para ferramentas como **Claude Desktop**, **Claude Code (CLI)** e **Gemini CLI**. O core se comunica via **stdio**, expondo seu arsenal agêntico para qualquer cliente compatível.

**2. Extensões de IDE & Clientes CLI:**

- **Extensão VS Code (Modos Agent & Sub-Agent):**
  - **Modo Agent:** Experiência interativa completa com painel de chat dedicado, ferramentas e feedback visual direto na sua IDE.
  - **Modo Sub-Agent:** Integra-se nativamente como um motor especialista em raciocínio de código que pode ser invocado pelo chat principal ou outros agentes (como o **Antigravity**) para lidar com tarefas complexas em toda a codebase.
  - **Bundled:** A extensão já inclui o binário do Vectora Core; não é necessário download separado.

- **Gemini CLI (Modo Sub-Agent):**
  - **Integração:** Opera exclusivamente como um **Sub-Agent** para o Gemini CLI. Ao expor seu RAG e arsenal agêntico via MCP, o Vectora permite que o Gemini CLI entenda e atue sobre bases de código complexas de forma nativa.

---

## SDKs & Protocolos

O Vectora se fundamenta em três pilares de integração para máxima escalabilidade e estabilidade corporativa:

### 1. Protocolos de Orquestração (ACP & MCP)

- **MCP (Model Context Protocol):** Operando focado em Contexto. Permite que agentes "Pai" (como Claude Code ou Antigravity) usem o Vectora como um Sub-Agent focado em busca e RAG profundo.
- **ACP (Agent Client Protocol):** Operando focado em Ação (JSON-RPC via stdio). O Vectora atua conectando de forma estrita extensões do VS Code (Agent) ou CLI ao Core sem overhead.

### 2. Provider Models SDKs

O Core implementa nativamente parsers complexos (streaming, chamadas de tools) em Go através da adoção estrita de SDKs oficiais, garantindo máxima confiabilidade:

- **google/genai** (Gemini 3.1 Pro,3.0 Flash e Embedding 2.0)
- **anthropic-sdk-go** (Claude 4.6 Sonnet/Opus, Claude 4.5 Haiku)
- **openai-go** (GPT-5.4 Pro/Mini, interoperabilidade com Qwen 3.6 API, Text Embeddings 3)
- **voyageai** (Advanced Voyage-3 Large/Code Embeddings)

### 3. Famílias de IA Suportadas (Padrão Abril 2026)

O Vectora foi projetado para operar com as 10 famílias de IA mais potentes do mercado:

| Família        | Modelos de Fronteira (2026)              |
| :------------- | :--------------------------------------- |
| **Google**     | Gemini 3.1 Pro, Gemini 3 Flash, Gemma 4  |
| **Anthropic**  | Claude 4.6 Sonnet/Opus, Claude 4.5 Haiku |
| **OpenAI**     | GPT-5.4 Pro, GPT-5.4 Mini, GPT-5-o1      |
| **Alibaba**    | Qwen 3.6-Plus, Qwen 3.6-Turbo, Qwen-Max  |
| **Voyage AI**  | Voyage-3 Large, Voyage-3 Code            |
| **Meta**       | Muse Spark, Llama 4 (Scout/Maverick)     |
| **Microsoft**  | Phi-4-Reasoning-Vision, Phi-4-Medium     |
| **DeepSeek**   | DeepSeek-V3.2, V3.2-Speciale             |
| **Mistral AI** | Mistral Small 4, Mistral Large 3         |
| **xAI**        | Grok 4.20, Grok 4.1                      |
| **Zhipu AI**   | GLM-5.1, GLM-5-Flash                     |

> [!TIP]
> Para uma validação técnica detalhada, links de documentação oficial e termos de busca para cada modelo, consulte o arquivo **[AGENTS.md](file:///c:/Users/bruno/Desktop/Vectora/AGENTS.md)**.

### 4. Multi-Tenancy Protocol (MTP)

O Core roda como um **Daemon Singleton** no background, consumindo o mínimo de RAM. Abrir mais IDs no editor não cria vários processos do Vectora; em vez disso, a arquitetura utiliza o _Multi-Tenancy Protocol_, estabelecendo Namespaces blindados por conexão IPC, balanceando as filas de requisições de AI de forma isolada por projeto e mitigando vazamento de Trust Folders.

---

## Quick Start (CLI Cobra)

```bash
# 1. Rode o core na pasta do projeto
vectora start

# 2. Configure sua chave (Ex: Gemini)
vectora config --gemini SUA_KEY

# 3. Gere embeddings
vectora embed

# 4. Pergunte sobre o código
vectora ask "Como funciona a autenticação?"
```

---

## Casos de Uso

- **Domínio de Codebase Legado:**
  Entenda sistemas desconhecidos rapidamente com contexto real.

- **Documentação de Nicho:**
  Respostas baseadas em versões específicas de documentação técnica.

- **Correlação de Requisitos:**
  Compare especificações com implementação real.

- **Onboarding Acelerado:**
  Aprenda padrões e decisões arquiteturais direto da codebase.

---

## Stack Tecnológica

- **Linguagem:** Go 1.23+ (Golang)
- **Arquitetura:** Daemon Singleton com Multi-Tenancy Protocol (MTP)
- **Interface:** **Cobra CLI** (Terminal) & **Systray** (Desktop)
- **Vector DB:** **chromem-go** (Motor RAG local-first)
- **Metadata Store:** **BBolt** (Persistência ACID para histórico e logs)
- **Modelos (Default):** Gemini 3.1 Pro (Reasoning) & Gemini Embedding 2 (RAG)
- **Protocolos:** ACP (Agent Client Protocol) & MCP (Model Context Protocol)
- **Inter-Process Communication (IPC):** JSON-RPC 2.0 sobre Named Pipes (Windows) ou Unix Sockets (POSIX)
- **Inferência Local:** Integração com **llama.cpp** (Qwen 3.6)
- **Otimização:** TurboQuant (Compressão de KV-Cache)

---

## Privacidade e Controle (Trust Folder)

- **Diretório Mestre (Config):** Suas configurações globais, logs e chaves de API permanecem isolados em %USERPROFILE%/.vectora.
- **Trust Folder (Escopo):** Por padrão, o local de onde você executa o **Vectora Core** é o seu "Diretório de Confiança". O core só terá permissão para ler ou modificar arquivos dentro deste escopo.
- **Reposicionamento:** Caso deseje conceder acesso a outro local do sistema, utilize a flag --path via CLI (ex: vectora start --path D:\MeuProjeto) para reposicionar o folder de confiança.
- **Hard-Coded Guardian:** Mesmo dentro do Trust Folder, o Vectora ignora automaticamente arquivos sensíveis como .env, .key e .pem.

---

## Toolkit Agêntico (Industrial Grade)

Exposto via MCP/ACP e construído do zero em Go, o Toolkit Agêntico é o arsenal de ferramentas que transforma o Vectora de um simples chatbot em um agente operacional capaz de interagir diretamente com seu sistema — sempre dentro do Trust Folder e com segurança transacional.

### Ferramentas de Arquivo e Sistema

- **`read_file`:** Leitura precisa de arquivos individuais com suporte a paginação por linhas para arquivos grandes.
- **`write_file`:** Escrita completa de novos arquivos ou sobrescrita controlada (com snapshot Git automático).
- **`read_folder`:** Listagem recursiva de diretórios com metadados de estrutura (tamanho, modificação, permissões).
- **`edit`:** Patching inteligente — edição refinada de trechos específicos sem reescrever o arquivo inteiro. Localiza o contexto exato via search-and-replace semântico e aplica alterações cirúrgicas.

### Ferramentas de Busca e Recuperação

- **`find_files`:** Busca rápida por padrões glob (ex: `**/*.ts`, `src/**/*.tsx`) com filtragem por nome e extensão.
- **`grep_search`:** Busca poderosa baseada em ripgrep com suporte a regex completo, filtragem por tipo de arquivo e limite de resultados.
- **`google_search`:** Integração com busca web para trazer contexto externo atualizado — ideal para documentação recente, changelogs e troubleshooting.
- **`web_fetch`:** Fetch direto de URLs específicas, convertendo HTML para markdown e extraindo o conteúdo relevante para o contexto da conversa.

### Ferramentas de Sistema e Terminal

- **`run_shell_command`:** Execução de comandos shell em ambiente controlado com captura de stdout/stderr em tempo real. Suporte a execução em background para processos longos e timeout configurável.

### Ferramentas de Memória e Planejamento

- **`save_memory`:** Persistência de fatos importantes e preferências do usuário em memória de longo prazo (global ou por projeto), permitindo personalização contínua entre sessões.
- **`enter_plan_mode`:** Ativação do modo de planejamento estruturado — antes de executar tarefas complexas, o agente elabora um plano passo-a-passo, valida com o usuário e só então inicia a implementação.

### Segurança e Isolamento

Todas as ferramentas operam exclusivamente dentro do **Trust Folder** definido pelo usuário. O **Hard-Coded Guardian** bloqueia automaticamente leitura/escrita de arquivos sensíveis (`.env`, `.key`, `.pem`, bancos de dados, binários), independentemente da inteligência do modelo ou instruções do prompt.

---

## Funcionalidades Principais

**Cérebro Híbrido (Inteligência na Nuvem + Dados Locais):**

O Vectora utiliza a potência do Gemini 3.1 Pro para raciocínio lógico e o Gemini Embedding 2.0 para criar vetores de alta precisão. Enquanto a inteligência reside na nuvem, os dados de conhecimento e o armazenamento vetorial permanecem locais, garantindo recuperação rápida sob demanda.

**Arsenal Agêntico (Industrial Grade):**

Ao contrário de modelos puramente de chat, o Vectora possui ferramentas reais para interagir com o seu sistema (dentro do Trust Folder):

- **Arquivos:** Ler, escrever, listar e editar (read, write, ls, edit).
- **Suporte a Git via IDE:** O Vectora opera em harmonia com o ambiente Git do usuário, permitindo que o controle de versão integrado da IDE gerencie o histórico e snapshots.
- **Terminal:** Execução de comandos shell com captura de saída em tempo real.
- **Knowledge Search:** Busca semântica profunda em seus workspaces locais usando chromem-go.

## Arquitetura Core

| Componente           | Tecnologia             | Papel                                                                               |
| -------------------- | ---------------------- | ----------------------------------------------------------------------------------- |
| **Vector DB**        | **chromem-go**         | Busca semântica e embeddings                                                        |
| **Key-Value DB**     | **bbolt**              | Histórico de chat, logs, configuração                                               |
| **Motor de IA**      | **Direct Calls**       | Chamadas HTTP/STDIO otimizadas para APIs e `llama.cpp`. Sem overhead de frameworks. |
| **Inferência Local** | **llama.cpp (native)** | Execução de modelos offline (Qwen 3.6) via integração nativa do sistema             |
| **Vectora Core**     | **Cobra + Systray**    | CLI, Systray, IPC (local), JSON-RPC/stdio (remoto)                                  |

---

## Como Usar

1. Execute `vectora.exe`
2. O core inicia no diretório atual
3. Configure sua API key
4. Conecte sua IDE ou agente

---

## Planos para o Futuro

O Vectora Core é o alicerce de um ecossistema maior focado em RAG híbrido e desenvolvimento assistido por IA.

**Modo 100% Local:**

- Integração nativa com **llama.cpp**
- Execução offline completa (incluindo embeddings)
- Suporte a modelos **Qwen 3.5** e **Qwen 3.6**

**TurboQuant (Eficiência Extrema):**

- Compressão extrema de KV Cache (3 a 3.5 bits por valor)
- Processamento de contextos massivos (128k a 1M de tokens) 100% localmente
- Perda de acurácia próxima de zero com tecnologias PolarQuant e corretor QJL

**Evolução do Motor de Recuperação:**

- Aprimoramento do grafo da codebase (relações mais ricas)
- Multi-hop mais eficiente e com menor latência
- Ranking híbrido (semântico + estrutural + relacional)

**Vectora Assets:**

- Marketplace de bases de conhecimento vetoriais
- Documentações oficiais, specs técnicas e datasets curados
- Download e indexação instantânea

**Interfaces Avançadas:**

- **Vectora Desktop (Fyne):** UI nativa para gestão de workspaces e navegação no grafo
- **Vectora TUI (Bubbletea):** Interface de terminal otimizada para produtividade

**Serviços Cloud:**

- **Vectora Web:** acesso remoto ao seu workspace
- **Vectora Auth:** autenticação e controle de acesso (RBAC)
- **Colaboração:** compartilhamento seguro de conhecimento entre equipes

---

_Parte da organização open source Kaffyn._
