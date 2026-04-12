# Vectora: Issue Requests & Technical Specification

Este documento detalha o planejamento integral do Vectora, definindo conceitos, exigências técnicas, stack e o roadmap de implementação estruturado por "Issues".

---

## 1. Visão Geral e Conceitos

O **Vectora** é um Agente de IA especialista em RAG (Retrieval-Augmented Generation), projetado para operar como um **Sub-Agente (Tier 2)** de alta performance. Ele possui um banco de dados KV e Vetorial interno, rodando localmente na máquina do usuário.

### 1.1 Modos de Operação

O Vectora opera em dois protocolos distintos para servir diferentes tipos de clientes:

- **Modo ACP (Agent Client Protocol):**
  - **Foco:** Ação e Interface Direta.
  - **Uso:** Extensão VS Code (Interface de Chat própria) e Clientes CLI.
  - **Escopo:** Todas as tools disponíveis (arquivo, terminal, embedding, busca web).
  - **Transporte:** JSON-RPC 2.0 via IPC (Named Pipes no Windows).

- **Modo MCP (Model Context Protocol):**
  - **Foco:** Contexto e RAG para outros Agentes.
  - **Uso:** Servir como sub-agente para **Antigravity**, **Cursor**, **Claude Desktop**, **Gemini CLI**.
  - **Escopo:** **Somente tools relacionadas a embedding e busca profunda**. Tools comuns (leitura de arquivos, terminal) são omitidas para evitar redundância/conflitos com o agente "Pai".
  - **Transporte:** Stdio (padrão MCP).

---

## 2. Tecnologias e Stack

### 2.1 Linguagem e Frameworks

- **Linguagem Principal:** Go 1.23+
- **CLI/Command Handler:** [Cobra](https://github.com/spf13/cobra)
- **GUI (Tray & Status):** [Systray](https://github.com/getlantern/systray)
- **UI (VS Code Panel):** Vite + React (dentro da extensão)

### 2.2 Armazenamento (Local-First)

- **Metadata/Logs DB:** [Bbolt](https://github.com/etcd-io/bbolt) (KV Store ACID).
- **Vector DB:** [Chromem-go](https://github.com/philippgille/chromem-go) (Leve e rápido).
- **Compressão de KV:** **TurboQuant** (Otimização extrema para contextos massivos de 128k a 1M tokens).

### 2.3 Protocolos e API

- **JSON-RPC 2.0:** Protocolo de comunicação para chamadas de ferramentas e mensagens.
- **Inter-process communication (IPC):** [go-ipc](https://github.com/Kilemonn/go-ipc) para comunicação de baixa latência.

---

## 3. SDKs e Provedores de LLM (Padrão 2026)

O Vectora utiliza exclusivamente SDKs oficiais para garantir estabilidade:

| Provedor       | SDK Sugerido                                                       | Modelos Suportados                                           |
| :------------- | :----------------------------------------------------------------- | :----------------------------------------------------------- |
| **Google**     | [go-genai](https://github.com/googleapis/go-genai)                 | Gemini 3.1 Pro, 3.0 Flash, Gemma 4                           |
| **Anthropic**  | [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) | Claude 4.6 Sonnet/Opus, Claude 4.5 Haiku                     |
| **OpenAI**     | [openai-go](https://github.com/openai/openai-go)                   | GPT-5.4 Pro/Mini, GPT-5-o1                                   |
| **VoyageAI**   | [voyageai](https://github.com/AustinfHunter/voyageai)              | Voyage-3 Large/Code (Embeddings)                             |
| **OpenRouter** | [go-sdk](https://github.com/OpenRouterTeam/go-sdk)                 | Qwen 3.6, Muse, Llama 4, Phi-4, DeepSeek, Mistral, Grok, GLM |

---

## 4. Toolkit Agêntico (Capabilities)

### 4.1 Tools Comuns (Apenas Modo ACP)

- `read_file`, `write_file`, `edit_file`
- `list_dir`, `search_files` (Grep/Ripgrep)
- `run_command` (Terminal seguro)
- `web_search` (Google/Brave)

### 4.2 Tools de Embedding (ACP & MCP)

- `index_project`: Escaneia e gera embeddings de toda a codebase.
- `semantic_search`: Busca semântica local no banco vetorial.
- `web_search_to_embed`: Faz busca web, converte para markdown e gera embeddings temporários para consulta.
- `implementation_plan_rag`: Gera planos de implementação enriquecidos com contexto recuperado do RAG.
- `project_graph_analysis`: Analisa relações entre arquivos/funções para RAG estrutural.

---

## 5. Roadmap de Implementação (Issue Requests)

### [Issue #1] Core: Foundation & DB Layer

- [ ] Implementar singleton do daemon Vectora.
- [ ] Configuração do Bbolt para armazenamento de chaves de API e metadados.
- [ ] Integração do Chromem-go para armazenamento de vetores.
- [ ] Implementar TurboQuant para compressão de KV Cache local.

### [Issue #2] API: IPC & Protocols

- [ ] Implementar JSON-RPC 2.0 Server sobre Pipes Nomeados (Windows) / Unix Sockets (Linux).
- [ ] Criar o MCP Server minimalista (Stdio) expondo apenas subset de embedding.
- [ ] Criar o ACP Client/Server para comunicação com a Extensão VS Code.

### [Issue #3] Inference: Providers & SDKs

- [ ] Implementar Provider Factory para alternar entre Gemini, Claude, OpenAI e OpenRouter.
- [ ] Garantir suporte a streaming e tool calling nativo em todos os SDKs.
- [ ] Implementar fallback automático de modelos em caso de rate limit.

### [Issue #4] Tools: The RAG Arsenal

- [ ] Implementar as tools básicas de sistema de arquivos.
- [ ] Implementar a engine de indexação e busca do Chromem-go.
- [ ] Implementar a ferramenta de "Web search to Embed".

#### [Issue #5] Integration: VS Code Ecosystem & CLI Bridge

- [x] **Phase 1-4:** Bootstrap, UI Refactoring, and Basic Chat.
- [x] **Phase 5: Advanced Features & Polish**
    - Implementação de Message Caching & History Storage (`~/.vectora/history/`).
    - Streaming de tokens em tempo real com collapsible "Thinking" blocks.
    - Tool call progress tracking e cancelamento de requisições.
    - Otimizador de Context Window e alertas de token usage.
- [x] **Phase 6-10: Robustness & Provider Integration**
    - Framework de Error Recovery com exponential backoff.
    - Integração multi-provider (Gemini, Claude, OpenAI) no frontend.
    - Analytics e métricas de performance integradas.
    - Refatoração completa para TypeScript e Type-Safety absoluto.
- [x] **Phase 11: Internationalization (i18n) & RTL Support**
    - Suporte nativo para 12 idiomas (incluindo PT-BR, JA, DE, FR).
    - Implementação de suporte RTL (Right-to-Left) para Árabe, Hebraico, Persa e Urdu.
    - Sistema de detecção automática de direção de texto e espelhamento de UI/ícones.
- [/] **Phase 12+: Project Optimization**
    - Melhoria nas Inline Suggestions baseadas em RAG local.
    - Indexação FTS (Full-Text Search) para o histórico local.

### [Issue #6] Gemini CLI & Agentic MCP

- [x] Implementação do Bridge `vectora-geminicli` via MCP.
- [x] Exposição de tools de RAG e File I/O para o Gemini CLI como Main Agent.
- [x] Configuração automática via JSON settings do Google Cloud SDK.
- [ ] Implementação de logs semânticos para auditoria de chamadas de sub-agente.

---

## 6. Exigências de Design e Segurança

- **Aesthetic Excellence:** Webview do VS Code deve seguir padrões premium (Glassmorphism, Dark Mode harmônico).
- **Trust Folder:** Nenhuma tool pode ler/escrever fora do diretório autorizado.
- **Hard-Coded Guardian:** Bloqueio binário de arquivos sensíveis (`.env`, `.key`).
- **Git Snapshots:** Snapshots automáticos antes de edições via `edit_file`.
