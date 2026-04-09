# Plano de Implementação: Extensões Vectora

> **Contexto:** O core do Vectora já está funcional — ACP server over stdio, MCP server over stdio, embedding via Gemini, ferramentas agênticas, RAG pipeline, IPC, systray. Agora precisamos construir os **clientes** que conectam ao core: uma extensão para VS Code (ACP) e um MCP server para Gemini CLI.

---

## 1. Visão Geral da Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│                      EDITOR / AGENTE                            │
│                                                                 │
│  ┌──────────────────┐           ┌──────────────────────┐       │
│  │  VS Code         │           │  Gemini CLI          │       │
│  │  (ACP Client)    │           │  (MCP Client)        │       │
│  └─────────────────┘           ──────────┬───────────┘       │
│           │                                │                    │
│           │ ACP (JSON-RPC 2.0)             │ MCP (JSON-RPC 2.0)│
│           │ stdio                          │ stdio              │
│           │                                │                    │
│  ┌────────▼────────────────────────────────▼───────────────┐   │
│  │              Vectora Core (Dual Server)                 │   │
│  │   ┌─────────────────────┐   ┌──────────────────────┐   │   │
│  │   │  ACP Server         │   │  MCP Server          │   │   │
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

**Princípio:** Vectora é **agente** para o VS Code (ACP) e **provedor de ferramentas** para o Gemini CLI (MCP). O core é o mesmo — apenas o protocolo de entrada muda.

---

## 2. VS Code Extension (ACP Client)

### 2.1 Papel: Cliente ACP

- VS Code é o **client** — orquestra a UI e repassa interações ao agente
- Vectora é o **agent** — pensa, pesquisa, executa ferramentas
- Protocolo: ACP over stdio (JSON-RPC 2.0)

### 2.2 Estrutura do Projeto

```
extensions/vscode/
├── package.json              → manifest da extensão
├── tsconfig.json             → TypeScript config
├── src/
│   ├── extension.ts          → entry point, activation
│   ├── acp-client.ts         → cliente ACP over stdio
│   ├── chat-panel.ts         → WebView sidebar (chat UI)
│   ├── permission-handler.ts → UI de aprovação de tool calls
│   ├── diff-provider.ts      → inline diffs no editor
│   ├── session-manager.ts    → gerencia sessões ACP
│   └── types/
│       └── acp.d.ts          → tipos TypeScript do ACP
├── media/
│   ├── icon.svg              → ícone da extensão
│   └── chat.css              → estilos do WebView
└── webpack.config.js         → bundler
```

### 2.3 `package.json` — Manifest

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
          "default": "vectora",
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

### 2.4 `src/acp-client.ts` — Cliente ACP over stdio

```typescript
import * as cp from "child_process";
import * as vscode from "vscode";
import { ACPMessage, ACPSessionUpdate, PermissionRequest } from "./types/acp";

export class ACPClient {
  private process: cp.ChildProcess | null = null;
  private buffer = "";
  private pendingRequests = new Map<number, { resolve: Function; reject: Function }>();
  private nextId = 0;
  public onSessionUpdate?: (update: ACPSessionUpdate) => void;
  public onPermissionRequest?: (req: PermissionRequest) => void;

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

  private request(method: string, params: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const id = this.nextId++;
      this.pendingRequests.set(id, { resolve, reject });
      const msg: ACPMessage = { jsonrpc: "2.0", id, method, params };
      this.write(JSON.stringify(msg) + "\n");
    });
  }

  private notify(method: string, params: any): void {
    this.write(JSON.stringify({ jsonrpc: "2.0", method, params }) + "\n");
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
        this.handleMessage(JSON.parse(line));
      } catch {}
    }
  }

  private handleMessage(msg: any): void {
    if (msg.id !== undefined) {
      const pending = this.pendingRequests.get(msg.id);
      if (pending) {
        this.pendingRequests.delete(msg.id);
        if (msg.error) pending.reject(new Error(msg.error.message));
        else pending.resolve(msg.result);
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

### 2.5 `src/chat-panel.ts` — WebView Sidebar

```typescript
import * as vscode from "vscode";
import { ACPClient } from "./acp-client";

export class ChatPanel {
  private panel: vscode.WebviewPanel | undefined;
  private sessionId: string | undefined;

  constructor(private client: ACPClient) {
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
      if (msg.type === "send") await this.sendMessage(msg.text);
    });
  }

  private async sendMessage(text: string): Promise<void> {
    if (!this.sessionId) {
      this.sessionId = await this.client.newSession(vscode.workspace.workspaceFolders?.[0]?.uri.fsPath || ".");
    }
    this.panel?.webview.postMessage({ type: "user_message", text });
    const result = await this.client.prompt(this.sessionId, text);
    this.panel?.webview.postMessage({ type: "done", stopReason: result.stopReason });
  }

  private handleUpdate(update: any): void {
    if (!this.panel) return;
    switch (update.sessionUpdate) {
      case "agent_message_chunk":
        this.panel.webview.postMessage({
          type: "agent_chunk",
          text: update.content?.[0]?.content?.text || "",
        });
        break;
      case "tool_call":
        this.panel.webview.postMessage({
          type: "tool_call",
          toolCallId: update.toolCallId,
          title: update.title,
          kind: update.kind,
          status: update.status,
        });
        break;
      case "tool_call_update":
        this.panel.webview.postMessage({
          type: "tool_call_update",
          toolCallId: update.toolCallId,
          status: update.status,
        });
        break;
      case "plan":
        this.panel.webview.postMessage({ type: "plan", entries: update.entries });
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
    body { font-family: var(--vscode-font-family); padding: 8px; padding-bottom: 50px; }
    .message { margin: 4px 0; padding: 6px 10px; border-radius: 4px; }
    .user { background: var(--vscode-input-background); text-align: right; }
    .agent { background: var(--vscode-editor-background); }
    .tool { font-size: 0.85em; opacity: 0.7; }
    .plan { border-left: 2px solid var(--vscode-focusBorder); padding-left: 8px; margin: 4px 0; }
    #input-bar { position:fixed; bottom:0; left:0; right:0; padding:8px; background: var(--vscode-editor-background); }
    #input { width: 100%; padding: 6px; border: 1px solid var(--vscode-input-border); background: var(--vscode-input-background); color: var(--vscode-input-foreground); }
  </style>
</head>
<body>
  <div id="chat"></div>
  <div id="input-bar"><input id="input" placeholder="Ask Vectora..." onkeydown="if(event.key==='Enter')send()"></div>
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
      const d = e.data;
      if (d.type === 'user_message') {
        chat.innerHTML += '<div class="message user">' + esc(d.text) + '</div>';
      } else if (d.type === 'agent_chunk') {
        let last = chat.querySelector('.message.agent:last-child');
        if (!last) { last = document.createElement('div'); last.className = 'message agent'; chat.appendChild(last); }
        last.textContent += d.text;
      } else if (d.type === 'done') {
        chat.innerHTML += '<div style="opacity:0.5;font-size:0.8em;">— ' + d.stopReason + ' —</div>';
      } else if (d.type === 'tool_call') {
        chat.innerHTML += '<div class="message tool">🔧 ' + esc(d.title || d.kind) + ' (' + d.status + ')</div>';
      } else if (d.type === 'plan') {
        chat.innerHTML += d.entries.map(e => '<div class="plan">⏳ ' + esc(e.content) + '</div>').join('');
      }
      chat.scrollTop = chat.scrollHeight;
    });
    function esc(t) { const d = document.createElement('div'); d.textContent = t; return d.innerHTML; }
  </script>
</body>
</html>`;
  }
}
```

### 2.6 `src/extension.ts` — Entry Point

```typescript
import * as vscode from "vscode";
import { ACPClient } from "./acp-client";
import { ChatPanel } from "./chat-panel";

let client: ACPClient | undefined;
let chatPanel: ChatPanel | undefined;

export async function activate(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration("vectora");
  const corePath = config.get<string>("corePath") || "vectora";
  const workspacePath = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;

  if (!workspacePath) {
    vscode.window.showErrorMessage("Vectora requires an open workspace folder.");
    return;
  }

  client = new ACPClient(corePath);
  await client.start(workspacePath);

  chatPanel = new ChatPanel(client);

  context.subscriptions.push(
    vscode.commands.registerCommand("vectora.newSession", async () => await chatPanel?.show()),
    vscode.commands.registerCommand("vectora.explainCode", async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor) return;
      const code = editor.document.getText(editor.selection);
      await chatPanel?.show();
      await chatPanel["sendMessage"](`Explain this code:\n\n\`\`\`\n${code}\n\`\`\``);
    }),
  );

  await chatPanel.show();
  context.subscriptions.push({ dispose: () => client?.stop() });
}

export function deactivate() {
  client?.stop();
}
```

---

## 3. Gemini CLI Integration (MCP Server)

### 3.1 Papel: Servidor MCP

- **Gemini CLI** é um **agente** — ele pensa e decide
- **Vectora** é um **MCP Server** — expõe ferramentas e recursos
- Protocolo: MCP over stdio (JSON-RPC 2.0)
- Gemini CLI se conecta ao Vectora como um **MCP client** para acessar:
  - Ferramentas agênticas (read_file, grep_search, run_shell_command, etc.)
  - Recursos RAG (busca semântica na codebase indexada)
  - Contexto do workspace

### 3.2 Estrutura do Projeto

```
extensions/mcp-server/
├── package.json              → manifest do pacote MCP
├── tsconfig.json
├── src/
│   ├── index.ts              → MCP server entry point (stdio)
│   ├── tools.ts              → registro de ferramentas Vectora
│   ├── resources.ts          → recursos RAG (workspace context)
│   └── types/
│       └── mcp.d.ts          → tipos MCP
└── scripts/
    └── install.sh            → script de instalação
```

### 3.3 `src/index.ts` — MCP Server Entry Point

```typescript
#!/usr/bin/env node

import * as readline from "readline";
import { ToolRegistry } from "./tools";
import { ResourceRegistry } from "./resources";

// MCP JSON-RPC over stdio
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false,
});

const tools = new ToolRegistry();
const resources = new ResourceRegistry();

rl.on("line", async (line) => {
  if (!line.trim()) return;

  let msg: any;
  try {
    msg = JSON.parse(line);
  } catch {
    writeError(null, -32700, "Parse error");
    return;
  }

  if (msg.method === "initialize") {
    writeResult(msg.id, {
      protocolVersion: "2024-11-05",
      capabilities: {
        tools: {},
        resources: {},
      },
      serverInfo: {
        name: "vectora-mcp",
        version: "0.1.0",
      },
    });
  } else if (msg.method === "tools/list") {
    writeResult(msg.id, { tools: tools.list() });
  } else if (msg.method === "tools/call") {
    const result = await tools.call(msg.params.name, msg.params.arguments);
    writeResult(msg.id, result);
  } else if (msg.method === "resources/list") {
    writeResult(msg.id, { resources: resources.list() });
  } else if (msg.method === "resources/read") {
    const content = await resources.read(msg.params.uri);
    writeResult(msg.id, { contents: content });
  } else {
    writeError(msg.id, -32601, `Method '${msg.method}' not found`);
  }
});

function writeResult(id: number, result: any): void {
  console.log(JSON.stringify({ jsonrpc: "2.0", id, result }));
}

function writeError(id: number | null, code: number, message: string): void {
  console.log(JSON.stringify({ jsonrpc: "2.0", id, error: { code, message } }));
}
```

### 3.4 `src/tools.ts` — Registro de Ferramentas Vectora

```typescript
import { exec } from "child_process";
import * as fs from "fs";
import * as path from "path";

interface MCPTool {
  name: string;
  description: string;
  inputSchema: {
    type: "object";
    properties: Record<string, { type: string; description: string }>;
    required: string[];
  };
}

export class ToolRegistry {
  private tools: MCPTool[] = [
    {
      name: "read_file",
      description: "Read the content of a file within the Vectora workspace.",
      inputSchema: {
        type: "object",
        properties: {
          path: { type: "string", description: "File path relative to workspace root" },
        },
        required: ["path"],
      },
    },
    {
      name: "write_file",
      description: "Write content to a file within the Vectora workspace.",
      inputSchema: {
        type: "object",
        properties: {
          path: { type: "string", description: "File path relative to workspace root" },
          content: { type: "string", description: "File content" },
        },
        required: ["path", "content"],
      },
    },
    {
      name: "grep_search",
      description: "Search file contents using regex within the Vectora workspace.",
      inputSchema: {
        type: "object",
        properties: {
          pattern: { type: "string", description: "Regex pattern to search for" },
          case_sensitive: { type: "boolean", description: "Whether to match case" },
        },
        required: ["pattern"],
      },
    },
    {
      name: "run_shell_command",
      description: "Execute a shell command within the Vectora workspace directory.",
      inputSchema: {
        type: "object",
        properties: {
          command: { type: "string", description: "Shell command to execute" },
        },
        required: ["command"],
      },
    },
    {
      name: "vector_search",
      description: "Perform semantic search on the Vectora indexed codebase using RAG.",
      inputSchema: {
        type: "object",
        properties: {
          query: { type: "string", description: "Natural language search query" },
          top_k: { type: "integer", description: "Number of results to return" },
        },
        required: ["query"],
      },
    },
  ];

  list(): MCPTool[] {
    return this.tools;
  }

  async call(name: string, args: any): Promise<any> {
    const cwd = process.env.VECTORA_WORKSPACE || process.cwd();

    switch (name) {
      case "read_file":
        return this.readFile(cwd, args.path);

      case "write_file":
        return this.writeFile(cwd, args.path, args.content);

      case "grep_search":
        return this.grepSearch(cwd, args.pattern, args.case_sensitive);

      case "run_shell_command":
        return this.runShellCommand(cwd, args.command);

      case "vector_search":
        return this.vectorSearch(args.query, args.top_k || 5);

      default:
        return {
          content: [{ type: "text", text: `Tool '${name}' not found` }],
          isError: true,
        };
    }
  }

  private readFile(cwd: string, filePath: string): any {
    try {
      const absPath = path.resolve(cwd, filePath);
      if (!absPath.startsWith(cwd)) {
        return { content: [{ type: "text", text: "Access denied: outside workspace" }], isError: true };
      }
      const content = fs.readFileSync(absPath, "utf-8");
      return { content: [{ type: "text", text: content }] };
    } catch (err: any) {
      return { content: [{ type: "text", text: err.message }], isError: true };
    }
  }

  private writeFile(cwd: string, filePath: string, content: string): any {
    try {
      const absPath = path.resolve(cwd, filePath);
      if (!absPath.startsWith(cwd)) {
        return { content: [{ type: "text", text: "Access denied: outside workspace" }], isError: true };
      }
      fs.writeFileSync(absPath, content, "utf-8");
      return { content: [{ type: "text", text: `File written: ${filePath}` }] };
    } catch (err: any) {
      return { content: [{ type: "text", text: err.message }], isError: true };
    }
  }

  private grepSearch(cwd: string, pattern: string, caseSensitive: boolean): any {
    return new Promise((resolve) => {
      const flags = caseSensitive ? "" : "i";
      exec(`grep -r${flags} "${pattern}" --include="*.*" . 2>/dev/null | head -50`, { cwd }, (err, stdout) => {
        if (err && stdout === "") {
          resolve({ content: [{ type: "text", text: "No matches found" }] });
        } else {
          resolve({ content: [{ type: "text", text: stdout || err?.message || "" }] });
        }
      });
    });
  }

  private runShellCommand(cwd: string, command: string): any {
    return new Promise((resolve) => {
      exec(command, { cwd, timeout: 30000 }, (err, stdout, stderr) => {
        const output = stdout + (stderr ? `\nstderr: ${stderr}` : "");
        resolve({ content: [{ type: "text", text: output || "Command executed successfully" }] });
      });
    });
  }

  private async vectorSearch(query: string, topK: number): Promise<any> {
    // Call Vectora IPC or internal API to perform RAG search
    // For now, return a placeholder
    return {
      content: [
        {
          type: "text",
          text: `Vector search for "${query}" — integrate with Vectora chromem-go for real results.`,
        },
      ],
    };
  }
}
```

### 3.5 `src/resources.ts` — Recursos RAG

```typescript
import * as fs from "fs";
import * as path from "path";

interface MCPResource {
  uri: string;
  name: string;
  description: string;
  mimeType?: string;
}

interface MCPResourceContent {
  uri: string;
  mimeType: string;
  text: string;
}

export class ResourceRegistry {
  private cwd: string;

  constructor() {
    this.cwd = process.env.VECTORA_WORKSPACE || process.cwd();
  }

  list(): MCPResource[] {
    return [
      {
        uri: `file://${this.cwd}`,
        name: "Workspace Context",
        description: "Full context of the indexed Vectora workspace including RAG embeddings.",
        mimeType: "text/markdown",
      },
      {
        uri: `vectora://workspace/dependencies`,
        name: "Dependency Graph",
        description: "Codebase dependency graph extracted during indexing.",
        mimeType: "application/json",
      },
      {
        uri: `vectora://workspace/index-status`,
        name: "Index Status",
        description: "Current RAG index status — files indexed, last update time.",
        mimeType: "application/json",
      },
    ];
  }

  async read(uri: string): Promise<MCPResourceContent[]> {
    if (uri.startsWith("file://")) {
      const filePath = uri.replace("file://", "");
      const content = fs.readFileSync(filePath, "utf-8");
      return [{ uri, mimeType: "text/plain", text: content }];
    }

    if (uri === "vectora://workspace/index-status") {
      return [
        {
          uri,
          mimeType: "application/json",
          text: JSON.stringify({
            status: "indexed",
            filesCount: 0,
            lastIndexed: new Date().toISOString(),
          }),
        },
      ];
    }

    return [{ uri, mimeType: "text/plain", text: "Resource not found" }];
  }
}
```

### 3.6 Configuração do Gemini CLI

O Gemini CLI conecta ao Vectora via MCP adicionando o server ao config:

```bash
# ~/.config/gemini/settings.json
{
  "mcpServers": {
    "vectora": {
      "command": "vectora",
      "args": ["mcp"],
      "env": {
        "VECTORA_WORKSPACE": "/path/to/project"
      }
    }
  }
}
```

Ou via variável de ambiente:

```bash
export GEMINI_MCP_SERVERS='{"vectora":{"command":"vectora","args":["mcp"]}}'
gemini
```

---

## 4. Protocolo de Comunicação — Resumo

| Camada            | Protocolo            | Transporte    | Direção          | Papel do Vectora                       |
| ----------------- | -------------------- | ------------- | ---------------- | -------------------------------------- |
| VS Code → Core    | **ACP** JSON-RPC 2.0 | stdio (pipes) | Bidirecional     | **Agent** (pensa, responde, usa tools) |
| Gemini CLI → Core | **MCP** JSON-RPC 2.0 | stdio (pipes) | Request/Response | **Server** (expõe tools + resources)   |
| Core → Embedding  | Gemini Embedding API | HTTPS         | Request/Response | Client                                 |
| Core → LLM        | Gemini/Claude API    | HTTPS         | Stream/Request   | Client                                 |
| Core → Vector DB  | chromem-go           | Local file    | In-process       | Owner                                  |

**Nenhuma comunicação de rede entre clientes e core** — tudo local via stdio. O único tráfego remoto é do core para APIs de IA.

---

## 5. Cronograma de Implementação

### Fase 1: VS Code Extension (Semana 1-2)

- [ ] Scaffold com Yeoman (`yo code`)
- [ ] ACP client over stdio (spawn `vectora acp`)
- [ ] WebView chat panel com streaming
- [ ] Session management (newSession, prompt, cancel)
- [ ] Tool call notifications (tool_call, tool_call_update)
- [ ] Permission request UI (allow/reject tool calls)

### Fase 2: VS Code Tool Integration (Semana 2-3)

- [ ] File system methods (fs/read_text_file, fs/write_text_file)
- [ ] Inline diff provider para edições de arquivo
- [ ] Terminal integration (terminal/create, terminal/output)
- [ ] Slash commands (/embed, /clear, /help)

### Fase 3: MCP Server para Gemini CLI (Semana 3-4)

- [ ] MCP server over stdio (spawn `vectora mcp`)
- [ ] Tool registry (read_file, write_file, grep_search, run_shell_command)
- [ ] Resource registry (workspace context, dependency graph, index status)
- [ ] vector_search tool (integração com chromem-go via IPC)
- [ ] Configuração do Gemini CLI (settings.json)

### Fase 4: Polimento e Testes (Semana 4-5)

- [ ] Testes E2E VS Code (simulated stdio)
- [ ] Testes E2E MCP (JSON-RPC over stdio)
- [ ] Packaging (`vsce package`)
- [ ] README e documentação
- [ ] Publicação no VS Code Marketplace

---

## 6. Notas Técnicas

### ACP vs MCP — Quando usar cada um

| Protocolo | Use quando...                                                                                    | Vectora é...                      |
| --------- | ------------------------------------------------------------------------------------------------ | --------------------------------- |
| **ACP**   | O cliente é uma IDE/editor que quer um **assistente de codificação** com chat, diffs, permissões | **Agent** (ativo, pensa, decide)  |
| **MCP**   | O cliente é um **agente** (Gemini CLI, Claude Code, Cursor) que quer **ferramentas e contexto**  | **Server** (passivo, expõe tools) |

### Compartilhamento de código

O ACP client (TypeScript) e o MCP server (TypeScript/Go) compartilham:

- **Core logic**: Ambos chamam o mesmo Vectora core binário
- **Transport**: Ambos usam stdio com JSON-RPC 2.0
- **Tools**: As mesmas ferramentas (read_file, grep_search, etc.) são expostas em ambos os protocolos

### Vetor de ataque de segurança

- **ACP**: O cliente (VS Code) tem controle total sobre permissões de tool calls
- **MCP**: O servidor (Vectora) aplica Guardian — bloqueia .env, .key, .db, .exe
- Ambos: Workspace-scoped — ferramentas operam apenas dentro do Trust Folder
