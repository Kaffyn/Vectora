# Plano de Implementação: Extensões Vectora

> **Contexto:** O core do Vectora já está funcional — ACP server over stdio, embedding via Gemini, ferramentas agênticas, RAG pipeline, IPC, systray. Agora precisamos construir os **clientes** que conectam ao core: uma extensão para VS Code e um plugin para Gemini CLI.

---

## 1. Visão Geral da Arquitetura

```
┌─────────────────────────────────────────────────────────┐
│                     EDITOR / CLI                        │
│                                                         │
│  ┌──────────────┐        ┌──────────────────────────┐  │
│  │ VS Code Ext. │        │   Gemini CLI Extension   │  │
│  │ (ACP Client) │        │    (ACP Client)          │  │
│  └──────┬───────┘        └──────────┬───────────────┘  │
│         │                           │                   │
│         │ stdio (JSON-RPC 2.0)      │ stdio             │
│         │                           │                   │
│  ┌──────▼───────────────────────────▼───────────────┐  │
│  │           Vectora Core (ACP Server)              │  │
│  │   ├── initialize / session / prompt / tools      │  │
│  │   ├── Engine (RAG + Tools + LLM Router)          │  │
│  │   └── Chromem-go + BBolt (local storage)         │  │
│  └──────────────────────────────────────────────────┘  │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │           Gemini Embedding API (remote)           │  │
│  │              Claude API (remote)                  │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**Princípio:** O Vectora Core é o **único** processo inteligente. As extensões são **clientes ACP** burros — apenas renderizam UI e repassam JSON-RPC.

---

## 2. VS Code Extension

### 2.1 Estrutura do Projeto

```
extensions/vscode/
├── package.json              → manifest da extensão
├── tsconfig.json             → TypeScript config
├── src/
│   ├── extension.ts          → entry point, activation
│   ├── acp-client.ts         → cliente ACP over stdio
│   ├── chat-panel.ts         → WebView sidebar (chat UI)
│   ├── tool-handler.ts       → aprovações de tool calls
│   ├── diff-provider.ts      → inline diffs no editor
│   ├── session-manager.ts    → gerencia sessões ACP
│   └── types/
│       └── acp.d.ts          → tipos TypeScript do ACP
├── media/
│   ├── icon.svg              → ícone da extensão
│   └── chat.css              → estilos do WebView
└── webpack.config.js         → bundler
```

### 2.2 `package.json` — Manifest

```json
{
  "name": "vectora-vscode",
  "displayName": "Vectora AI",
  "description": "Local AI coding assistant — RAG + agentic tools via ACP",
  "version": "0.1.0",
  "publisher": "kaffyn",
  "engines": { "vscode": "^1.90.0" },
  "categories": ["AI", "Chat", "Programming Languages"],
  "activationEvents": ["onStartupFinished"],
  "contributes": {
    "viewsContainers": {
      "activitybar": [
        {
          "id": "vectora",
          "title": "Vectora",
          "icon": "media/icon.svg"
        }
      ]
    },
    "views": {
      "vectora": [
        {
          "id": "vectora.chat",
          "type": "webview",
          "name": "Chat"
        }
      ]
    },
    "commands": [
      {
        "command": "vectora.newSession",
        "title": "New Vectora Session"
      },
      {
        "command": "vectora.explainCode",
        "title": "Explain with Vectora"
      }
    ],
    "configuration": {
      "title": "Vectora",
      "properties": {
        "vectora.corePath": {
          "type": "string",
          "default": "vectora.exe",
          "description": "Path to Vectora core binary"
        },
        "vectora.workspace": {
          "type": "string",
          "description": "Workspace folder for RAG indexing"
        }
      }
    }
  },
  "main": "./dist/extension.js",
  "scripts": {
    "compile": "webpack --mode production",
    "watch": "webpack --mode development --watch",
    "vsce:package": "vsce package"
  }
}
```

### 2.3 `src/acp-client.ts` — Cliente ACP over stdio

```typescript
import * as cp from "child_process";
import * as vscode from "vscode";
import { ACPMessage, ACPSessionUpdate, ToolCall, PermissionRequest } from "./types/acp";

export class ACPClient {
  private process: cp.ChildProcess | null = null;
  private buffer = "";
  private pendingRequests = new Map<number, { resolve: Function; reject: Function }>();
  private nextId = 0;
  private onSessionUpdate?: (update: ACPSessionUpdate) => void;
  private onPermissionRequest?: (req: PermissionRequest) => void;

  constructor(private corePath: string) {}

  async start(workspacePath: string): Promise<void> {
    this.process = cp.spawn(this.corePath, ["acp"], {
      stdio: ["pipe", "pipe", "pipe"],
      cwd: workspacePath,
    });

    this.process.stdout!.on("data", (data) => this.onData(data));
    this.process.stderr!.on("data", (data) => {
      vscode.window.showErrorMessage(`Vectora: ${data.toString()}`);
    });

    // Handshake: initialize
    const result = await this.request("initialize", {
      protocolVersion: 1,
      clientCapabilities: {
        fs: { readTextFile: true, writeTextFile: true },
        terminal: true,
      },
      clientInfo: {
        name: "vectora-vscode",
        title: "Vectora VS Code",
        version: "0.1.0",
      },
    });

    if (result.protocolVersion !== 1) {
      throw new Error(`Protocol version mismatch: expected 1, got ${result.protocolVersion}`);
    }
  }

  async newSession(cwd: string): Promise<string> {
    const result = await this.request("session/new", { cwd });
    return result.sessionId;
  }

  async prompt(sessionId: string, prompt: string): Promise<{ stopReason: string }> {
    return this.request("session/prompt", {
      sessionId,
      prompt: [{ type: "text", text: prompt }],
    });
  }

  async readFile(sessionId: string, path: string): Promise<string> {
    const result = await this.request("fs/read_text_file", { sessionId, path });
    return result.content;
  }

  async writeFile(sessionId: string, path: string, content: string): Promise<void> {
    await this.request("fs/write_text_file", { sessionId, path, content });
  }

  cancelSession(sessionId: string): void {
    this.notify("session/cancel", { sessionId });
  }

  // JSON-RPC request/response
  private request(method: string, params: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const id = this.nextId++;
      this.pendingRequests.set(id, { resolve, reject });

      const msg: ACPMessage = { jsonrpc: "2.0", id, method, params };
      this.write(JSON.stringify(msg) + "\n");
    });
  }

  private notify(method: string, params: any): void {
    const msg = { jsonrpc: "2.0", method, params };
    this.write(JSON.stringify(msg) + "\n");
  }

  private write(data: string): void {
    this.process?.stdin?.write(data);
  }

  private onData(chunk: Buffer): void {
    this.buffer += chunk.toString();

    while (true) {
      const idx = this.buffer.indexOf("\n");
      if (idx === -1) break;

      const line = this.buffer.substring(0, idx);
      this.buffer = this.buffer.substring(idx + 1);

      if (!line.trim()) continue;

      try {
        const msg = JSON.parse(line);
        this.handleMessage(msg);
      } catch {
        // Skip malformed lines (stderr leakage)
      }
    }
  }

  private handleMessage(msg: any): void {
    if (msg.id !== undefined) {
      // Response to our request
      const pending = this.pendingRequests.get(msg.id);
      if (pending) {
        this.pendingRequests.delete(msg.id);
        if (msg.error) {
          pending.reject(new Error(msg.error.message));
        } else {
          pending.resolve(msg.result);
        }
      }
    } else if (msg.method === "session/update") {
      this.onSessionUpdate?.(msg.params.update);
    } else if (msg.method === "session/request_permission") {
      this.onPermissionRequest?.(msg.params);
    }
  }

  stop(): void {
    this.process?.kill();
    this.process = null;
  }
}
```

### 2.4 `src/chat-panel.ts` — WebView Sidebar

```typescript
import * as vscode from "vscode";
import { ACPClient } from "./acp-client";

export class ChatPanel {
  private panel: vscode.WebviewPanel | undefined;
  private sessionId: string | undefined;

  constructor(
    private client: ACPClient,
    private context: vscode.ExtensionContext,
  ) {
    this.client.onSessionUpdate = (update) => this.handleUpdate(update);
    this.client.onPermissionRequest = (req) => this.handlePermission(req);
  }

  async show(): Promise<void> {
    if (this.panel) {
      this.panel.reveal(vscode.ViewColumn.One);
      return;
    }

    this.panel = vscode.window.createWebviewPanel("vectoraChat", "Vectora Chat", vscode.ViewColumn.One, {
      enableScripts: true,
      retainContextWhenHidden: true,
    });

    this.panel.webview.html = this.getHtml();
    this.panel.webview.onDidReceiveMessage(async (msg) => {
      if (msg.type === "send") {
        await this.sendMessage(msg.text);
      } else if (msg.type === "permission") {
        // Forward permission decision
        this.panel?.webview.postMessage({
          type: "permission_result",
          optionId: msg.optionId,
        });
      }
    });
  }

  private async sendMessage(text: string): Promise<void> {
    if (!this.sessionId) {
      this.sessionId = await this.client.newSession(vscode.workspace.workspaceFolders?.[0]?.uri.fsPath || ".");
    }

    // Show user message in UI
    this.panel?.webview.postMessage({ type: "user_message", text });

    // Send to ACP
    const result = await this.client.prompt(this.sessionId, text);

    // Show stop reason
    this.panel?.webview.postMessage({ type: "done", stopReason: result.stopReason });
  }

  private handleUpdate(update: any): void {
    if (!this.panel) return;

    switch (update.sessionUpdate) {
      case "agent_message_chunk":
        // Stream text to UI
        this.panel.webview.postMessage({
          type: "agent_chunk",
          text: update.content?.[0]?.content?.text || "",
        });
        break;

      case "tool_call":
        // Show tool call pending
        this.panel.webview.postMessage({
          type: "tool_call",
          toolCallId: update.toolCallId,
          title: update.title,
          kind: update.kind,
          status: update.status,
        });
        break;

      case "tool_call_update":
        // Update tool call status
        this.panel.webview.postMessage({
          type: "tool_call_update",
          toolCallId: update.toolCallId,
          status: update.status,
          content: update.content,
        });
        break;

      case "plan":
        // Show plan entries
        this.panel.webview.postMessage({
          type: "plan",
          entries: update.entries,
        });
        break;
    }
  }

  private handlePermission(req: any): void {
    if (!this.panel) return;
    this.panel.webview.postMessage({
      type: "permission_request",
      toolCallId: req.toolCall.toolCallId,
      options: req.options,
    });
  }

  private getHtml(): string {
    return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: var(--vscode-font-family); padding: 8px; }
    .message { margin: 4px 0; padding: 6px 10px; border-radius: 4px; }
    .user { background: var(--vscode-input-background); text-align: right; }
    .agent { background: var(--vscode-editor-background); }
    .tool { font-size: 0.85em; opacity: 0.7; }
    .plan { border-left: 2px solid var(--vscode-focusBorder); padding-left: 8px; margin: 4px 0; }
    input { width: 100%; padding: 6px; border: 1px solid var(--vscode-input-border); background: var(--vscode-input-background); color: var(--vscode-input-foreground); }
  </style>
</head>
<body>
  <div id="chat"></div>
  <div style="position:fixed;bottom:0;left:0;right:0;padding:8px;">
    <input id="input" placeholder="Ask Vectora..." onkeydown="if(event.key==='Enter')send()">
  </div>
  <script>
    const chat = document.getElementById('chat');
    const vscode = acquireVsCodeApi();

    function send() {
      const input = document.getElementById('input');
      if (!input.value.trim()) return;
      vscode.postMessage({ type: 'send', text: input.value });
      input.value = '';
    }

    window.addEventListener('message', (e) => {
      const { type, text, stopReason, entries, toolCallId, status, title, kind, options } = e.data;

      if (type === 'user_message') {
        chat.innerHTML += '<div class="message user">' + escapeHtml(text) + '</div>';
      } else if (type === 'agent_chunk') {
        let last = chat.querySelector('.message.agent:last-child');
        if (!last) {
          last = document.createElement('div');
          last.className = 'message agent';
          chat.appendChild(last);
        }
        last.textContent += text;
      } else if (type === 'done') {
        chat.innerHTML += '<div style="opacity:0.5;font-size:0.8em;">— ' + stopReason + ' —</div>';
      } else if (type === 'tool_call') {
        chat.innerHTML += '<div class="message tool">🔧 ' + escapeHtml(title || kind) + ' (' + status + ')</div>';
      } else if (type === 'tool_call_update') {
        // Update existing tool call status
      } else if (type === 'plan') {
        const planHtml = entries.map(e => '<div class="plan">⏳ ' + escapeHtml(e.content) + '</div>').join('');
        chat.innerHTML += planHtml;
      } else if (type === 'permission_request') {
        const optsHtml = options.map(o =>
          '<button onclick="vscode.postMessage({type:\\"permission\\",optionId:\\"'+o.optionId+'\\"})">' + escapeHtml(o.name) + '</button>'
        ).join(' ');
        chat.innerHTML += '<div style="padding:4px;">Approve? ' + optsHtml + '</div>';
      }

      chat.scrollTop = chat.scrollHeight;
    });

    function escapeHtml(t) {
      const d = document.createElement('div');
      d.textContent = t;
      return d.innerHTML;
    }
  </script>
</body>
</html>`;
  }
}
```

### 2.5 `src/diff-provider.ts` — Inline Diffs

```typescript
import * as vscode from "vscode";

export class DiffProvider {
  private decorations: vscode.TextEditorDecorationType[] = [];

  // Apply a diff received from ACP tool_call_update
  applyDiff(filePath: string, oldText: string, newText: string): vscode.Disposable {
    const uri = vscode.Uri.file(filePath);

    return vscode.workspace.openTextDocument(uri).then((doc) => {
      return vscode.window.showTextDocument(doc).then((editor) => {
        // Create decoration for added/removed lines
        const addedDecoration = vscode.window.createTextEditorDecorationType({
          backgroundColor: new vscode.ThemeColor("diffEditor.insertedLineBackground"),
          isWholeLine: true,
        });

        const removedDecoration = vscode.window.createTextEditorDecorationType({
          backgroundColor: new vscode.ThemeColor("diffEditor.removedLineBackground"),
          isWholeLine: true,
        });

        this.decorations.push(addedDecoration, removedDecoration);

        // Highlight the changed region
        const range = editor.document.getText().includes(oldText)
          ? this.findRange(editor.document, oldText)
          : undefined;

        if (range) {
          editor.setDecorations(removedDecoration, [range]);
        }

        return {
          dispose: () => {
            addedDecoration.dispose();
            removedDecoration.dispose();
          },
        };
      });
    });
  }

  private findRange(doc: vscode.TextDocument, text: string): vscode.Range | undefined {
    const content = doc.getText();
    const start = content.indexOf(text);
    if (start === -1) return undefined;

    const startPos = doc.positionAt(start);
    const endPos = doc.positionAt(start + text.length);
    return new vscode.Range(startPos, endPos);
  }
}
```

### 2.6 `src/extension.ts` — Entry Point

```typescript
import * as vscode from "vscode";
import { ACPClient } from "./acp-client";
import { ChatPanel } from "./chat-panel";
import { DiffProvider } from "./diff-provider";

let client: ACPClient | undefined;
let chatPanel: ChatPanel | undefined;
let diffProvider: DiffProvider | undefined;

export async function activate(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration("vectora");
  const corePath = config.get<string>("corePath") || "vectora";
  const workspacePath = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;

  if (!workspacePath) {
    vscode.window.showErrorMessage("Vectora requires an open workspace folder.");
    return;
  }

  // Initialize ACP client
  client = new ACPClient(corePath);
  await client.start(workspacePath);

  // Initialize components
  chatPanel = new ChatPanel(client, context);
  diffProvider = new DiffProvider();

  // Register commands
  context.subscriptions.push(
    vscode.commands.registerCommand("vectora.newSession", async () => {
      await chatPanel?.show();
    }),
    vscode.commands.registerCommand("vectora.explainCode", async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor) return;

      const selection = editor.selection;
      const selectedCode = editor.document.getText(selection);

      await chatPanel?.show();
      await chatPanel["sendMessage"](`Explain this code:\n\n\`\`\`\n${selectedCode}\n\`\`\``);
    }),
  );

  // Auto-open chat panel
  await chatPanel.show();

  context.subscriptions.push({
    dispose: () => client?.stop(),
  });
}

export function deactivate() {
  client?.stop();
}
```

---

## 3. Gemini CLI Extension

### 3.1 Visão Geral

O **Gemini CLI** (antigo `gemini`) é uma interface de linha de comando da Google para desenvolvimento assistido por IA. A extensão Vectora permite que o Gemini CLI use o Vectora como **backend de agente** via ACP.

### 3.2 Estrutura do Projeto

```
extensions/gemini-cli/
├── package.json
├── tsconfig.json
├── src/
│   ├── index.ts              → CLI entry point
│   ├── acp-client.ts         → mesmo cliente ACP (compartilhado)
│   ├── session.ts            → sessão interativa REPL
│   ├── tool-executor.ts      → executa tools localmente
│   └── types/
│       └── acp.d.ts          → tipos ACP
└── scripts/
    └── install.sh            → script de instalação
```

### 3.3 `src/index.ts` — CLI Entry Point

```typescript
#!/usr/bin/env node

import * as readline from "readline";
import * as path from "path";
import { ACPClient } from "./acp-client";
import { Session } from "./session";

async function main() {
  const corePath = process.env.VECTORA_CORE_PATH || "vectora";
  const cwd = process.cwd();

  console.log("🚀 Vectora for Gemini CLI");
  console.log("Connecting to Vectora core...");

  // Start ACP client
  const client = new ACPClient(corePath);
  await client.start(cwd);

  // Create new session
  const sessionId = await client.newSession(cwd);

  console.log("✅ Connected. Type your questions or commands.\n");

  // REPL loop
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  const session = new Session(client, sessionId, rl);
  await session.run();
}

main().catch((err) => {
  console.error("Fatal error:", err.message);
  process.exit(1);
});
```

### 3.4 `src/session.ts` — REPL Interativo

```typescript
import * as readline from "readline";
import { ACPClient } from "./acp-client";

export class Session {
  private toolCalls = new Map<string, { title: string; kind: string }>();

  constructor(
    private client: ACPClient,
    private sessionId: string,
    private rl: readline.Interface,
  ) {
    this.client.onSessionUpdate = (update) => this.handleUpdate(update);
  }

  async run(): Promise<void> {
    while (true) {
      const input = await this.prompt();
      if (!input || input.trim() === "/exit" || input.trim() === "/quit") {
        break;
      }

      // Handle slash commands
      if (input.startsWith("/")) {
        await this.handleCommand(input);
        continue;
      }

      // Send prompt to Vectora
      process.stdout.write("\n");
      const result = await this.client.prompt(this.sessionId, input);

      if (result.stopReason === "refusal") {
        console.log("⚠️  Vectora refused to answer.");
      }
      console.log("");
    }

    this.client.stop();
    this.rl.close();
  }

  private prompt(): Promise<string> {
    return new Promise((resolve) => {
      this.rl.question(" > ", resolve);
    });
  }

  private async handleCommand(cmd: string): Promise<void> {
    const parts = cmd.split(" ");
    switch (parts[0]) {
      case "/embed":
        // Trigger embedding of current directory
        console.log("📦 Embedding workspace...");
        // This would call vectora embed via subprocess
        break;
      case "/clear":
        this.toolCalls.clear();
        console.log("🧹 Session cleared.");
        break;
      case "/help":
        console.log(`
Commands:
  /embed          - Embed current workspace
  /clear          - Clear session context
  /help           - Show this help
  /exit, /quit    - Exit
        `);
        break;
      default:
        console.log("Unknown command. Type /help for available commands.");
    }
  }

  private handleUpdate(update: any): void {
    switch (update.sessionUpdate) {
      case "agent_message_chunk":
        const text = update.content?.[0]?.content?.text || "";
        process.stdout.write(text);
        break;

      case "tool_call":
        console.log(`\n🔧 ${update.title || update.kind}...`);
        this.toolCalls.set(update.toolCallId, {
          title: update.title,
          kind: update.kind,
        });
        break;

      case "tool_call_update":
        const tc = this.toolCalls.get(update.toolCallId);
        if (tc && update.status === "completed") {
          console.log(`✅ ${tc.title || tc.kind} done.\n`);
        }
        break;

      case "plan":
        console.log("\n📋 Plan:");
        update.entries.forEach((e: any) => {
          console.log(`  ${e.status === "completed" ? "✅" : "⏳"} ${e.content}`);
        });
        console.log("");
        break;
    }
  }
}
```

### 3.5 Integração com Gemini CLI Existente

O Gemini CLI original (do pacote `@google/gemini-cli`) pode ser **patcheado** para usar o Vectora como agente backend:

```typescript
// Patch: node_modules/@google/gemini-cli/dist/agent.js
const { ACPClient } = require("vectora-gemini-cli");

// Override the default agent
const originalAgent = Agent.create;
Agent.create = async (options) => {
  const acpClient = new ACPClient(process.env.VECTORA_CORE_PATH || "vectora");
  await acpClient.start(process.cwd());

  return {
    async chat(messages) {
      const prompt = messages.map((m) => m.content).join("\n");
      const result = await acpClient.prompt("default", prompt);
      return { content: result };
    },
    async tools() {
      // Vectora tools are exposed via ACP
      return acpClient.getAvailableTools();
    },
  };
};
```

---

## 4. Cronograma de Implementação

### Fase 1: VS Code Extension (Semana 1-2)

- [ ] Scaffold com Yeoman (`yo code`)
- [ ] ACP client over stdio
- [ ] WebView chat panel
- [ ] Session management
- [ ] Streaming de respostas

### Fase 2: Tool Calls & Permissões (Semana 2-3)

- [ ] UI de aprovação de tool calls
- [ ] Diff provider para inline edits
- [ ] File system integration (read/write via ACP)

### Fase 3: Gemini CLI Extension (Semana 3-4)

- [ ] CLI scaffolding
- [ ] REPL interativo
- [ ] Slash commands (/embed, /clear, /help)
- [ ] Patch do Gemini CLI original

### Fase 4: Polimento (Semana 4-5)

- [ ] Testes E2E
- [ ] Packaging (`vsce package`)
- [ ] README e documentação
- [ ] Publicação no VS Code Marketplace

---

## 5. Protocolo de Comunicação — Resumo

| Camada           | Protocolo            | Transporte    | Direção          |
| ---------------- | -------------------- | ------------- | ---------------- |
| Extensão → Core  | ACP JSON-RPC 2.0     | stdio (pipes) | Bidirecional     |
| Core → Embedding | Gemini Embedding API | HTTPS         | Request/Response |
| Core → LLM       | Gemini/Claude API    | HTTPS         | Stream/Request   |
| Core → Vector DB | chromem-go           | Local file    | In-process       |

**Nenhuma comunicação de rede entre extensão e core** — tudo local via stdio. O único tráfego remoto é do core para APIs de IA.
