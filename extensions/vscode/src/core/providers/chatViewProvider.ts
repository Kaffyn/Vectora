import * as vscode from "vscode";
import { AcpClient } from "../acpClient";
import type { SessionUpdate } from "../types";

/**
 * ChatViewProvider manages the Sidebar Webview for Vectora chat.
 */
export class ChatViewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = "vectora.chatView";
  private _view?: vscode.WebviewView;
  private isStreaming = false;
  private sessionId?: string;

  constructor(
    private client: AcpClient | undefined,
    private context: vscode.ExtensionContext,
  ) {
    if (this.client) {
      this.client.onSessionUpdate.event(this.handleNotification, this);
    }
  }

  public setClient(client: AcpClient) {
    this.client = client;
    this.client.onSessionUpdate.event(this.handleNotification, this);
  }

  public resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken,
  ) {
    this._view = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this.context.extensionUri],
    };

    webviewView.webview.html = this.getHtml(webviewView.webview);

    // Show connecting state if client is not ready
    if (!this.client || !this.client.isConnected) {
      webviewView.webview.postMessage({ type: "connecting", message: "Vectora is connecting..." });
    }

    webviewView.webview.onDidReceiveMessage(async (msg) => {
      switch (msg.type) {
        case "prompt":
          await this.sendMessageInternal(msg.text);
          break;

        case "cancel":
          if (this.client && this.sessionId) {
            this.client.cancelSession(this.sessionId);
            this._view?.webview.postMessage({ type: "stream_end", stopReason: "cancelled" });
          }
          break;

        case "clear":
          this.sessionId = undefined;
          await this.clearChat();
          break;
      }
    });
  }

  public async sendMessage(text: string): Promise<void> {
    if (this._view) {
      this._view.show?.(true);
      this._view.webview.postMessage({ type: "inject_message", text });
    } else {
      vscode.commands.executeCommand("vectora.chatView.focus");
    }
  }

  public async clearChat(): Promise<void> {
    this._view?.webview.postMessage({ type: "clear_chat" });
    this.sessionId = undefined;
  }

  private async sendMessageInternal(text: string): Promise<void> {
    if (!text.trim()) return;

    if (!this.client || !this.client.isConnected) {
      await vscode.commands.executeCommand("vectora.start");
      // Wait for connection to establish
      for (let i = 0; i < 20; i++) {
        if (this.client?.isConnected) break;
        await new Promise((r) => setTimeout(r, 500));
      }
    }

    if (!this.client || !this.client.isConnected) {
      this._view?.webview.postMessage({ type: "error", message: "Vectora Core is not running. Failed to auto-start." });
      return;
    }

    if (!this.sessionId) {
      const workspacePath = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      if (!workspacePath) {
        vscode.window.showErrorMessage("Vectora: No workspace folder open.");
        return;
      }

      try {
        const config = vscode.workspace.getConfiguration("vectora");
        const provider = config.get<string>("defaultProvider") || "gemini";

        const resp = await this.client.createSession({
          workspaceId: workspacePath,
          provider,
        });
        this.sessionId = resp.sessionId;
      } catch (err: any) {
        this._view?.webview.postMessage({ type: "error", message: `Failed to create session: ${err.message}` });
        return;
      }
    }

    this._view?.webview.postMessage({ type: "user_message", id: Date.now().toString(), text });

    try {
      const result = await this.client.prompt({
        sessionId: this.sessionId,
        messages: [{ role: "user", content: text }],
      });
      this._view?.webview.postMessage({ type: "stream_end", stopReason: "completed" });
    } catch (err: any) {
      this._view?.webview.postMessage({ type: "error", message: err.message });
    }
  }

  private handleNotification(update: SessionUpdate): void {
    if (!this._view) return;

    switch (update.type) {
      case "message":
        // Streaming message content
        if (update.content) {
          this._view.webview.postMessage({ type: "agent_chunk", text: update.content });
        }
        break;

      case "tool_call":
        // Tool execution started
        this._view.webview.postMessage({
          type: "tool_call",
          toolName: update.toolName,
          toolInput: update.toolInput,
        });
        break;

      case "tool_result":
        // Tool execution completed
        this._view.webview.postMessage({
          type: "tool_result",
          toolName: update.toolName,
          result: update.content,
        });
        break;

      case "error":
        // Error occurred
        this._view.webview.postMessage({
          type: "error",
          message: update.content || "Unknown error",
        });
        break;

      case "complete":
        // Session completed
        this._view.webview.postMessage({ type: "stream_end", stopReason: "completed" });
        break;
    }
  }


  private getNonce(): string {
    let text = "";
    const possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    for (let i = 0; i < 32; i++) text += possible.charAt(Math.floor(Math.random() * possible.length));
    return text;
  }

  private getHtml(webview: vscode.Webview): string {
    const distPath = vscode.Uri.joinPath(this.context.extensionUri, "dist", "webview");
    const scriptUri = webview.asWebviewUri(vscode.Uri.joinPath(distPath, "assets", "index.js"));
    const styleUri = webview.asWebviewUri(vscode.Uri.joinPath(distPath, "assets", "index.css"));

    const nonce = this.getNonce();

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}' ${webview.cspSource}; img-src ${webview.cspSource} https:;">
    <link rel="stylesheet" type="text/css" href="${styleUri}">
    <title>Vectora Chat</title>
</head>
<body>
    <div id="root"></div>
    <script type="module" nonce="${nonce}" src="${scriptUri}"></script>
</body>
</html>`;
  }
}
