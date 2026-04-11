# Plano Mestre: Estabilização, Segurança e Modernização Multi-SDK

Este plano detalha a execução técnica de todas as decisões arquiteturais consolidadas no `issue_report.md`. A meta é entregar um ecossistema Vectora resiliente, com comunicação unificada e integração nativa com os SDKs de última geração.

## User Review Required

> [!IMPORTANT] > **Modelo de Comunicação Unificado**: Unificaremos a camada de transporte em **IPC (Named Pipes/Sockets) + JSON-RPC 2.0**.
>
> - **Go**: `github.com/sourcegraph/jsonrpc2` para lidar com a complexidade de streams bidirecionais.
> - **TS/VS Code**: `vscode-jsonrpc` para integração nativa.
> - As extensões e UIs nunca acessarão os SDKs de IA diretamente, consumindo apenas o Core.

> [!WARNING] > **Sistema de Update (Windows)**: Implementaremos um binário auxiliar `updater.exe` para permitir a substituição de `vectora.exe` em execução, com mecanismo de **auto-rollback** se a nova versão falhar no health check.

> [!NOTE] > **Gestão de Dados**: Mudanças no schema do banco ou nos embeddings dispararão uma **re-indexação automática em background** com notificação visual ao usuário, garantindo a integridade dos resultados de busca.

## Proposed Changes

---

### 1. Extensão VS Code (Interface e Estabilidade)

#### [MODIFY] [extension.ts](file:///c:/Users/bruno/Desktop/Vectora/extensions/vscode/src/extension.ts) [binary-manager.ts](file:///c:/Users/bruno/Desktop/Vectora/extensions/vscode/src/binary-manager.ts) [chat-panel.ts](file:///c:/Users/bruno/Desktop/Vectora/extensions/vscode/src/chat-panel.ts)

- **Ação**: Registro imediato do provedor de Webview para evitar erro de "View Not Found".
- **Ação**: Adicionar suporte a caminhos de instalação `%LOCALAPPDATA%\Vectora`.
- **Ação**: Atualizar CSP para permitir assets gerados pelo Vite.
- **Ação**: Migrar para `vscode-jsonrpc` no cliente de comunicação.

---

### 2. Core Foundation (Segurança e Singleton)

#### [MODIFY] [main.go](file:///c:/Users/bruno/Desktop/Vectora/cmd/core/main.go)

- **Singleton Híbrido**: Lógica de File Lock (`.vectora.lock`) combinada com validação de PID ativo.
- **Handshake Auth**: Geração e validação de tokens temporários para o canal IPC.

#### [MODIFY] [server.go](file:///c:/Users/bruno/Desktop/Vectora/core/api/ipc/server.go)

- Migrar para `sourcegraph/jsonrpc2`.
- Middleware de sanitização de logs (mascaramento de conteúdo sensível).

---

### 3. Core Engine (Modernização Multi-SDK)

#### [MODIFY] [gemini_provider.go](file:///c:/Users/bruno/Desktop/Vectora/core/llm/gemini_provider.go)

- **Migração**: Adotar `google.golang.org/genai`.
- Ativar suporte a `ThinkingConfig` e Embeddings nativos.

#### [MODIFY] [claude_provider.go](file:///c:/Users/bruno/Desktop/Vectora/core/llm/claude_provider.go)

- **Migração**: Adotar `anthropic-sdk-go`.
- Ativar suporte a **Prompt Caching**.

#### [NEW] [voyage_provider.go](file:///c:/Users/bruno/Desktop/Vectora/core/llm/voyage_provider.go)

- Implementação nativa via SDK oficial para embeddings profissionais.

#### [MODIFY] [engine/query.go](file:///c:/Users/bruno/Desktop/Vectora/core/engine/query.go)

- **Memória**: Implementar limpeza agressiva de buffers e monitoramento via `pprof` local.

---

### 4. Workspaces & Dados (Resiliência)

#### [MODIFY] [workspace.go](file:///c:/Users/bruno/Desktop/Vectora/cmd/core/workspace.go)

- Implementar detecção de versão de schema e rotina de re-indexação em background.
- Melhorar comandos CLI (aliases plural e exibição de paths).

---

## Open Questions

- Nenhuma questão pendente. O design foi fechado com as definições de memória, updater e schema.

## Verification Plan

### Automated Tests

- `go mod tidy` para integrar as novas libs JSON-RPC e SDKs de IA.
- Teste unitário da lógica de Singleton (PID check).
- Mock de stream JSON-RPC para validar o handshake e sanitização de logs.

### Manual Verification

1. Build via `.\build.ps1`.
2. Validar carregamento da Webview no VS Code.
3. Testar o "auto-update" (simular novo binário) e observar o updater em ação.
4. Forçar mismatch de schema no banco para validar a re-indexação automática.
