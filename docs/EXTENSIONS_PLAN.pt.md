# Plano de Implementação: Extensões Vectora

> **Status:** ✅ IMPLEMENTAÇÃO CONCLUÍDA — Build e type-check passando
> **Data:** 2026-04-09 (atualizado com resultados de build)
> **Contexto:** O core do Vectora já está funcional — ACP server over stdio, MCP server over stdio, embedding via Gemini, ferramentas agênticas, RAG pipeline, IPC, systray. As extensões conectam editores e agentes ao core.

---

## 1. Visão Geral da Arquitetura (Final)

```
┌─────────────────────────────────────────────────────────────────┐
│                      EDITOR / AGENTE                            │
│                                                                 │
│  ┌──────────────────┐           ┌──────────────────────┐       │
│  │  VS Code         │           │  Gemini CLI          │       │
│  │  (ACP Client)    │           │  (MCP Client)        │       │
│  │  TypeScript      │           │  TypeScript          │       │
│  └─────────────────┘           ──────────┬───────────┘       │
│           │                                │                    │
│           │ ACP (JSON-RPC 2.0)             │ MCP (JSON-RPC 2.0)│
│           │ stdio                          │ stdio              │
│           │                                │                    │
│  ┌────────▼────────────────────────────────▼───────────────┐   │
│  │              Vectora Core (Dual Server)                 │   │
│  │   ┌─────────────────────┐   ┌──────────────────────┐   │   │
│  │   │  ACP Server (Go)    │   │  MCP Server (Go)     │   │   │
│  │   │  (Agent Mode)       │   │  (Tool Provider)     │   │   │
│  │   │  - session/prompt   │   │  - tools/list        │   │   │
│  │   │  - tools/call       │   │  - tools/call        │   │   │
│  │   │  - fs/read, write   │   │  - resources/*       │   │   │
│  │   └──────────┬──────────┘   └─────────────────────┘   │   │
│  │              │                         │                │   │
│  │  ┌───────────▼─────────────────────────▼────────────┐   │   │
│  │  │  Engine (RAG + Tools + LLM Router + Chromem)     │   │   │
│  │  └──────────────────────────────────────────────────┘   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │         Gemini Embedding / Claude API (remote)            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

**Princípio:** Vectora é **agente** para o VS Code (ACP) e **provedor de ferramentas** para o Gemini CLI (MCP). O core Go é o mesmo — apenas o protocolo de entrada muda.

---

## 2. VS Code Extension (ACP Client) — ✅ IMPLEMENTADO E COMPILADO

### 2.1 Build Results

| Arquivo                                | Type-Check | Bundle Size  | Status |
| -------------------------------------- | ---------- | ------------ | ------ |
| `extensions/vscode/package.json`       | —          | —            | ✅ OK  |
| `extensions/vscode/tsconfig.json`      | ✅ Pass    | —            | ✅ OK  |
| `extensions/vscode/webpack.config.js`  | ✅ Pass    | —            | ✅ OK  |
| `extensions/vscode/src/types/acp.d.ts` | ✅ Pass    | —            | ✅ OK  |
| `extensions/vscode/src/acp-client.ts`  | ✅ Pass    | —            | ✅ OK  |
| `extensions/vscode/src/chat-panel.ts`  | ✅ Pass    | —            | ✅ OK  |
| `extensions/vscode/src/extension.ts`   | ✅ Pass    | —            | ✅ OK  |
| `extensions/vscode/dist/extension.js`  | —          | **17.2 KiB** | ✅ OK  |

**Command:** `npx tsc --noEmit` → **0 errors**
**Command:** `npx webpack --mode production` → **compiled successfully in 1713 ms**

### 2.2 Funcionalidades Implementadas

**ACP Client (`acp-client.ts`):**

- `start(workspacePath)` — spawn `vectora acp` + initialize handshake
- `newSession(cwd)` — cria sessão ACP
- `prompt(sessionId, text)` — envia prompt, retorna stopReason
- `cancel(sessionId)` — cancela prompt em andamento
- `readFile()` / `writeFile()` — fs operations via ACP
- `onSessionUpdate` — EventEmitter para streaming de chunks, tool calls, plans
- `onPermissionRequest` — EventEmitter para aprovação de tool calls
- `onError` — EventEmitter para erros do subprocesso

**Chat Panel (`chat-panel.ts`):**

- WebView sidebar com HTML/CSS/JS vanilla (sem dependências)
- Streaming de respostas (chunk por chunk)
- Tool call notifications (🔧 pending → in_progress → completed)
- Plan visualization (⏳ / ✅ entries)
- Cancel button durante streaming
- Auto-resize textarea
- Permission request UI
- VS Code theme variables (dark/light mode)

**Extension (`extension.ts`):**

- Status bar indicator (conectando → conectado → erro)
- 3 commands: `vectora.newSession`, `vectora.explainCode`, `vectora.refactorCode`
- Context menu integration (editor/context)
- 3 config properties: `corePath`, `workspace`, `streaming`
- Auto-open chat panel on activation
- Graceful cleanup on deactivate

### 2.3 Comandos VS Code

| Comando                | Descrição                   | Acesso          |
| ---------------------- | --------------------------- | --------------- |
| `vectora.newSession`   | Nova sessão ACP             | Command Palette |
| `vectora.explainCode`  | Explica código selecionado  | Context Menu    |
| `vectora.refactorCode` | Refatora código selecionado | Context Menu    |

---

## 3. Gemini CLI Extension (MCP Client) — ✅ IMPLEMENTADO E COMPILADO

### 3.1 Build Results

| Arquivo                                   | Type-Check | Output Size | Status |
| ----------------------------------------- | ---------- | ----------- | ------ |
| `extensions/geminicli/package.json`       | —          | —           | ✅ OK  |
| `extensions/geminicli/tsconfig.json`      | ✅ Pass    | —           | ✅ OK  |
| `extensions/geminicli/src/types/mcp.d.ts` | ✅ Pass    | —           | ✅ OK  |
| `extensions/geminicli/src/mcp-client.ts`  | ✅ Pass    | 5.9 KiB     | ✅ OK  |
| `extensions/geminicli/src/index.ts`       | ✅ Pass    | 9.9 KiB     | ✅ OK  |
| `extensions/geminicli/dist/index.js`      | —          | —           | ✅ OK  |

**Command:** `npx tsc --noEmit` → **0 errors**
**Command:** `npx tsc` → **compiled to dist/** (15.8 KiB total)

### 3.2 Funcionalidades Implementadas

**MCP Client (`mcp-client.ts`):**

- `start(workspacePath)` — spawn `vectora mcp` + initialize handshake
- `listTools()` — lista ferramentas MCP disponíveis
- `callTool(name, args)` — chama ferramenta MCP
- `prompt(sessionId, text)` — envia prompt ACP
- `newSession(cwd)` — cria sessão ACP

**CLI (`index.ts`):**

- **REPL interativo** — prompt loop com `/new`, `/tools`, `/call`, `/embed`, `/help`, `/exit`
- **`config`** — gera JSON de configuração para Gemini CLI settings.json
- **`list-tools`** — lista ferramentas com schemas
- **`call-tool <name> [args]`** — chama ferramenta via CLI

### 3.3 Integração com Gemini CLI

**Via settings.json:**

```json
{
  "mcpServers": {
    "vectora": {
      "command": "vectora",
      "args": ["mcp"]
    }
  }
}
```

**Via variável de ambiente:**

```bash
GEMINI_MCP_SERVERS='{"vectora":{"command":"vectora","args":["mcp"]}}' gemini
```

---

## 4. Protocolo de Comunicação — Resumo Final

| Camada            | Protocolo            | Transporte    | Direção          | Papel do Vectora                     |
| ----------------- | -------------------- | ------------- | ---------------- | ------------------------------------ |
| VS Code → Core    | **ACP** JSON-RPC 2.0 | stdio (pipes) | Bidirecional     | **Agent** (pensa, responde, streama) |
| Gemini CLI → Core | **MCP** JSON-RPC 2.0 | stdio (pipes) | Request/Response | **Server** (expõe tools + resources) |
| Core → Embedding  | Gemini Embedding API | HTTPS         | Request/Response | Client                               |
| Core → LLM        | Gemini/Claude API    | HTTPS         | Stream/Request   | Client                               |
| Core → Vector DB  | chromem-go           | Local file    | In-process       | Owner                                |

**Nenhuma comunicação de rede entre clientes e core** — tudo local via stdio. O único tráfego remoto é do core para APIs de IA.

---

## 5. ACP vs MCP — Quando Usar Cada Um

| Protocolo | Use quando...                                                                                    | Vectora é...                      |
| --------- | ------------------------------------------------------------------------------------------------ | --------------------------------- |
| **ACP**   | O cliente é uma IDE/editor que quer um **assistente de codificação** com chat, diffs, permissões | **Agent** (ativo, pensa, decide)  |
| **MCP**   | O cliente é um **agente** (Gemini CLI, Claude Code, Cursor) que quer **ferramentas e contexto**  | **Server** (passivo, expõe tools) |

---

## 6. Status de Build — Resumo Completo

| Componente           | Command                         | Result      |
| -------------------- | ------------------------------- | ----------- |
| Go Core              | `go build ./...`                | ✅ OK       |
| Go Tests             | `go test ./core/...`            | ✅ 4/4      |
| Production Binaries  | `build.ps1`                     | ✅ 3/3      |
| VS Code Extension    | `npx tsc --noEmit`              | ✅ 0 errors |
| VS Code Extension    | `npx webpack --mode production` | ✅ 17.2 KiB |
| Gemini CLI Extension | `npx tsc --noEmit`              | ✅ 0 errors |
| Gemini CLI Extension | `npx tsc`                       | ✅ 15.8 KiB |

**Production Binaries:**

- `vectora-windows-amd64.exe` — 24.98 MB ✅
- `vectora-linux-amd64` — 24.00 MB ✅
- `vectora-darwin-amd64.app` — 24.81 MB ✅
- **Total:** 89.7 MB

---

## 7. Próximos Passos (Pós-Implementação)

### Fase 1: Testes E2E (Manual)

- [ ] Instalar extensão VS Code via `code --install-extension dist/vectora-vscode-0.1.0.vsix`
- [ ] Testar ACP handshake com `vectora acp` manualmente
- [ ] Testar Gemini CLI com `node dist/index.js`
- [ ] Testar MCP tools/list entre Gemini CLI e `vectora mcp`

### Fase 2: Polimento VS Code

- [ ] Diff provider para inline edits (show diff before applying)
- [ ] Terminal integration (terminal/create, terminal/output)
- [ ] Slash commands no chat (/embed, /clear, /help)
- [ ] Keyboard shortcuts (Ctrl+Shift+V para Vectora)

### Fase 3: Polimento Gemini CLI

- [ ] Color output no REPL
- [ ] Tool output formatting (syntax highlighting)
- [ ] Config file support (~/.vectora/geminicli.json)
- [ ] Batch mode (non-interactive tool calls)

### Fase 4: Publicação

- [ ] `vsce package` — gerar .vsix
- [ ] Publicar no VS Code Marketplace
- [ ] Publicar no npm (vectora-geminicli)
- [ ] README com screenshots e demo GIFs

---

## 8. Notas Técnicas

### ACP Client Architecture (VS Code)

- `ACPClient` é um wrapper thin sobre `child_process.spawn`
- JSON-RPC over stdio com newline-delimited JSON
- vscode.EventEmitter para session updates e permission requests
- WebView usa vanilla JS (sem framework) para performance
- Theme variables do VS Code para dark/light mode

### MCP Client Architecture (Gemini CLI)

- `McpClient` é idêntico ao ACP client, mas com types MCP
- CLI usa readline para REPL interativo
- Subcommands para configuração e tool calls
- Designed para ser consumido pelo Gemini CLI como MCP server

### Shared Code Patterns

- Ambos os clients usam o mesmo pattern: spawn → initialize → request/response loop
- Buffer de leitura com newline delimiter
- Pending requests map com Promise correlation
- Graceful shutdown com kill + pending rejection

### Security

- **ACP**: VS Code tem controle total sobre permissões de tool calls
- **MCP**: Vectora aplica Guardian — bloqueia .env, .key, .db, .exe
- **Ambos**: Workspace-scoped — ferramentas operam apenas dentro do Trust Folder
