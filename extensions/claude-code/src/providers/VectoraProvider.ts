import * as vscode from "vscode";
import { MCPClient } from "../utils/mcpClient";
import { MCPTool } from "../utils/types";

interface HistoryEntry {
  tool: string;
  input: string;
  output: string;
  timestamp: Date;
}

export class VectoraProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = "vectora.sidebar";
  private view?: vscode.WebviewView;
  private history: HistoryEntry[] = [];

  constructor(private readonly client: MCPClient) {
    client.on("connected", (tools: MCPTool[]) => {
      this.update(tools, true);
    });
    client.on("disconnected", () => {
      this.update([], false);
    });
  }

  resolveWebviewView(webviewView: vscode.WebviewView): void {
    this.view = webviewView;
    webviewView.webview.options = { enableScripts: true };
    webviewView.webview.html = this.getHtml(
      this.client.getTools(),
      this.client.isConnected()
    );

    webviewView.webview.onDidReceiveMessage(async (msg) => {
      if (msg.type === "callTool") {
        const result = await this.client.callTool(msg.tool, { query: msg.input, path: msg.input });
        const output = result.content.map((c) => c.text).join("\n");
        this.history.unshift({ tool: msg.tool, input: msg.input, output, timestamp: new Date() });
        this.update(this.client.getTools(), true);
      }
    });
  }

  addHistoryEntry(entry: HistoryEntry): void {
    this.history.unshift(entry);
    this.update(this.client.getTools(), this.client.isConnected());
  }

  private update(tools: MCPTool[], connected: boolean): void {
    if (this.view) {
      this.view.webview.html = this.getHtml(tools, connected);
    }
  }

  private getHtml(tools: MCPTool[], connected: boolean): string {
    const statusColor = connected ? "#4caf50" : "#f44336";
    const statusText = connected ? `✓ Conectado (${tools.length} ferramentas)` : "✗ Desconectado";

    const toolsHtml = tools
      .map(
        (t) => `
        <div class="tool">
          <strong>${t.name}</strong>
          <p>${t.description ?? ""}</p>
        </div>`
      )
      .join("");

    const historyHtml = this.history
      .slice(0, 5)
      .map(
        (h) => `
        <div class="history-item">
          <span class="tool-name">${h.tool}</span>
          <span class="time">${h.timestamp.toLocaleTimeString()}</span>
          <p class="input">${h.input}</p>
          <pre class="output">${h.output.slice(0, 200)}${h.output.length > 200 ? "..." : ""}</pre>
        </div>`
      )
      .join("");

    return `<!DOCTYPE html>
<html>
<head>
<style>
  body { font-family: var(--vscode-font-family); font-size: 12px; padding: 8px; }
  .status { padding: 6px 10px; border-radius: 4px; background: ${statusColor}22; color: ${statusColor}; margin-bottom: 12px; font-weight: bold; }
  .section-title { font-size: 11px; text-transform: uppercase; opacity: 0.6; margin: 12px 0 4px; }
  .tool { padding: 6px; border-radius: 4px; margin-bottom: 4px; background: var(--vscode-list-hoverBackground); }
  .tool strong { font-size: 12px; }
  .tool p { margin: 2px 0 0; opacity: 0.7; font-size: 11px; }
  .history-item { border-left: 2px solid var(--vscode-focusBorder); padding: 4px 8px; margin-bottom: 8px; }
  .tool-name { font-weight: bold; font-size: 11px; }
  .time { float: right; opacity: 0.5; font-size: 10px; }
  .input { opacity: 0.7; margin: 2px 0; }
  .output { font-size: 10px; background: var(--vscode-editor-background); padding: 4px; border-radius: 2px; white-space: pre-wrap; word-break: break-word; }
  .shortcuts { opacity: 0.6; font-size: 11px; }
  kbd { background: var(--vscode-keybindingLabel-background); border: 1px solid var(--vscode-keybindingLabel-border); border-radius: 3px; padding: 1px 4px; font-size: 10px; }
</style>
</head>
<body>
  <div class="status">${statusText}</div>

  <div class="section-title">Atalhos</div>
  <div class="shortcuts">
    <kbd>Ctrl+Shift+V S</kbd> Buscar<br/>
    <kbd>Ctrl+Shift+V A</kbd> Analisar<br/>
    <kbd>Ctrl+Shift+V R</kbd> Refatorar<br/>
    <kbd>Ctrl+Shift+V D</kbd> Documentar
  </div>

  ${tools.length > 0 ? `<div class="section-title">Ferramentas (${tools.length})</div>${toolsHtml}` : ""}
  ${this.history.length > 0 ? `<div class="section-title">Histórico</div>${historyHtml}` : ""}
</body>
</html>`;
  }
}
