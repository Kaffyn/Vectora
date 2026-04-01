# Plano de Implementação do Sistema Vectora

Este documento detalha o plano de implementação do sistema Vectora, consolidando a filosofia, arquitetura, estrutura de repositório e regras de negócio essenciais extraídas dos documentos `README.pt.md`, `CONTRIBUTING.pt.md` e `BussinesRules.pt.md`.

## 1. Visão Geral e Filosofia

O Vectora é um assistente de IA privado que opera localmente, permitindo aos usuários interagir com seus próprios dados sem dependência de serviços em nuvem. A filosofia central é construída sobre:

- **Local Primeiro:** Todas as funcionalidades críticas devem operar offline. Provedores de nuvem são opcionais.
- **Baixo Consumo:** O daemon do systray deve consumir menos de 5MB de RAM em idle, com o sistema completo abaixo de 4GB sob carga.
- **Go Puro:** Preferência por Go, evitando CGO, C++ e interpretadores em runtime, a menos que indispensável.

## 2. Arquitetura do Sistema

A arquitetura do Vectora é modular, com um daemon central que orquestra todas as operações e interfaces.

### 2.1 Daemon Central (Systray)

- O `cmd/vectora` atua como a única fonte de verdade para o estado em runtime.
- Gerencia workspaces, conexões LLM e expõe um servidor IPC.
- Inicializa e encerra processos de interface sob demanda.
- Executa o GitBridge para garantir snapshots antes de qualquer operação de escrita.

### 2.2 Interfaces

Todas as interfaces são clientes leves e sem estado que se comunicam com o daemon via IPC.

- **Web UI (`cmd/vectora-web`):** Aplicação desktop com Wails e um frontend Next.js estático embarcado. Interface principal de chat, gerenciamento de workspaces e configurações.
- **CLI (`cmd/vectora-cli`):** Interface de terminal via Bubbletea para interação mínima e imediata.
- **Servidor MCP (`internal/mcp`):** Expõe o conhecimento do Vectora para ferramentas e IDEs externas.
- **Agente ACP (`internal/acp`):** Modo agente autônomo com acesso ao sistema de arquivos e terminal.

### 2.3 Motor de IA (`internal/llm`)

- Utiliza `langchaingo` como abstração para provedores LLM e de embedding (Gemini, Qwen/llama.cpp).
- **Gemini:** Integrado via API Key do usuário, suporta indexação multimodal.
- **Qwen (Local):** Roda offline via `llama.cpp` (sidecar), otimizado para texto e código.
- A API Key da Gemini é armazenada criptografada localmente e nunca exposta em logs ou payloads.

### 2.4 Camada de Armazenamento (`internal/db`)

- **`chromem-go`:** Para armazenamento de vetores.
- **`bbolt`:** Para armazenamento chave-valor (histórico, configuração).
- Cada workspace tem uma coleção `chromem-go` e um bucket `bbolt` isolados (`ws:<workspace_id>`).

### 2.5 Pipeline RAG (`internal/core`)

- Responsável por recuperação, mesclagem e re-ranqueamento de contexto de múltiplos workspaces ativos antes da chamada ao LLM.
- `IndexWorkspace` é assíncrono e reporta progresso.

### 2.6 Toolkit Agêntico (`internal/tools`)

Conjunto de ferramentas para operações de sistema de arquivos, busca, web, shell e memória.

- **`read_file`, `write_file`, `read_folder`, `edit`** (Filesystem)
- **`find_files`, `grep_search`, `google_search`, `web_fetch`** (Busca/Web)
- **`run_shell_command`** (Sistema)
- **`save_memory`, `enter_plan_mode`** (Memória)
- **Regra crucial:** Tools que alteram o estado (i.e., escrita ou execução shell) devem chamar `git.Bridge.Snapshot()` antes de sua execução para permitir reversão.

### 2.7 GitBridge (`internal/git`)

- Fornece funcionalidades de `Snapshot()` e `Restore()` para garantir a segurança das operações de escrita e execução de shell realizadas pelas ferramentas agênticas.

### 2.8 Contrato IPC (`internal/ipc`)

- Comunicação entre daemon e interfaces via Unix Domain Sockets (Linux/macOS) ou Named Pipes (Windows).
- Formato de mensagem: JSON delimitado por newline.
- Define métodos para `Workspace`, `Provider`, `Tools`, `Index` e `Session`, além de eventos proativos.

## 3. Estrutura do Repositório

O projeto segue as convenções de layout de projetos Go.

```
vectora/
├── cmd/            # Pontos de entrada dos binários (sem lógica de negócio)
├── internal/       # Toda a lógica de negócio (privada, não importável externamente)
│   ├── core/       # Pipeline RAG, gerenciamento de workspaces
│   ├── db/         # Wrappers de banco de dados
│   ├── llm/        # Abstração de provedor LLM
│   ├── ipc/        # Servidor e cliente IPC
│   ├── tray/       # UI do systray e ciclo de vida
│   ├── tools/      # Toolkit agêntico
│   ├── git/        # GitBridge
│   ├── mcp/        # Servidor MCP
│   ├── acp/        # Agente ACP
│   ├── index/      # Cliente do Vectora Index
│   └── infra/      # Transversal (logging, config, erros)
├── pkg/            # Packages públicos (reservado para SDK futuro)
├── web/            # Frontend Next.js (embarcado via Wails)
├── index-server/   # Servidor HTTP do Vectora Index (independente)
├── assets/         # Assets estáticos
├── scripts/        # Scripts de build, release, setup
├── tests/          # Suites de testes
└── docs/           # Documentação
```

## 4. Regras de Negócio e Arquitetura Críticas

- **RN-01 — Orçamento de RAM:** Rigoroso controle de consumo de memória.
- **RN-02 — Sem Rede no Core:** `internal/core`, `internal/db`, `internal/tools`, `internal/git` não devem fazer chamadas de rede.
- **RN-03 — Isolamento de Workspaces:** Workspaces isolados; agregação apenas na camada RAG.
- **RN-04 — Proteção de Escrita:** Todas as operações de escrita/shell requerem snapshot GitBridge prévio.
- **RN-05 — Abstração de Provedor:** Nenhum arquivo fora de `internal/llm` pode importar SDKs de provedor diretamente.
- **RN-06 — Interfaces Sem Estado:** Binários de interface são descartáveis e não mantêm estado.
- **RN-07 — Frontend Embarcado:** Frontend Next.js 100% estático, sem CDN externo, embarcável via `go:embed`.
- **RN-08 — Curadoria do Index:** Downloads do Vectora Index requerem verificação de assinatura Kaffyn.
- **TDD (Red-Green-Refactor):** Todas as lógicas de negócio devem ser precedidas por testes que falham.
- **Padrão 300%:** Cada funcionalidade testada em Happy Path, Negative e Edge Case.
- **SSOT (Single Source of Truth):** Este plano e as regras de negócio são a fonte primária para a implementação.

## 5. Tecnologias Chave

- **Linguagem:** Go
- **Banco Vetorial:** `chromem-go`
- **Banco Chave-Valor:** `bbolt`
- **Motor de IA:** `langchaingo` (com integrações Gemini e `llama.cpp` para Qwen)
- **Instalador:** `Fyne`
- **Daemon/Systray:** `systray`
- **Web UI:** `Wails` + `Next.js` (estático)
- **CLI:** `Bubbletea`
- **Servidor Index:** Go (`net/http`)

## 6. Requisitos e Construção

- **Requisitos:** Go 1.22+, Node.js 20+, Wails CLI.
- **Build:** Uso de `Makefile` para compilação de todos os binários (`make build`).
- **Testes:** `make test`, `make test-integration`, `make test-race`. Todos os PRs devem passar nos testes.

## 7. Convenções

- **Conventional Commits:** `<tipo>(escopo): <descrição>`.
- **Pull Requests:** Referenciar issue aberta, requer aprovações (especialmente para `internal/core` e `internal/db`), sem novas dependências CGO sem discussão, Go doc para funções públicas.
- **O que não fazer:** Adicionar estado a interfaces, chamar SDKs de provedor fora de `internal/llm`, contornar o GitBridge, adicionar packages a `pkg/` sem discussão, embutir URLs de CDN externo, introduzir dependências de servidor.
