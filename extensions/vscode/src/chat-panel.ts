import * as vscode from 'vscode';
import { Client } from './client';
import { SessionUpdate, InitializeRequest, InitializeResponse, SessionNewRequest, SessionNewResponse, SessionPromptRequest, PromptResponse } from './types/client';

/**
 * ChatViewProvider manages the Sidebar Webview for Vectora chat.
 */
export class ChatViewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'vectora.chatView';
  private _view?: vscode.WebviewView;
  private isStreaming = false;

  constructor(
    private client: Client,
    private context: vscode.ExtensionContext
  ) {
    this.client.onNotification.event(this.handleNotification, this);
  }

  public setClient(client: Client) {
    this.client = client;
    this.client.onNotification.event(this.handleNotification, this);
  }

  public resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ) {
    this._view = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this.context.extensionUri]
    };

    webviewView.webview.html = this.getHtml(webviewView.webview);

    webviewView.webview.onDidReceiveMessage(async (msg) => {
      switch (msg.type) {
        case 'send':
          await this.sendMessageInternal(msg.text);
          break;
        case 'cancel':
          if (this.client.sessionId) {
            this.client.notify('session/cancel', { sessionId: this.client.sessionId });
            this.updateUIState(false);
            this._view?.webview.postMessage({ type: 'stream_end', stopReason: 'cancelled' });
          }
          break;
      }
    });
  }

  public async sendMessage(text: string): Promise<void> {
    if (this._view) {
      this._view.show?.(true);
      this._view.webview.postMessage({ type: 'inject_message', text });
    } else {
      vscode.commands.executeCommand('vectora.chatView.focus');
    }
  }

  public async clearChat(): Promise<void> {
    this._view?.webview.postMessage({ type: 'clear_chat' });
  }

  private async sendMessageInternal(text: string): Promise<void> {
    if (!text.trim()) return;

    if (!this.client.sessionId) {
      const workspacePath = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      if (!workspacePath) {
        vscode.window.showErrorMessage('Vectora: No workspace folder open.');
        return;
      }
      
      const resp = await this.client.request<SessionNewRequest, SessionNewResponse>('session/new', { 
        cwd: workspacePath 
      });
      this.client.sessionId = resp.sessionId;
    }

    this.updateUIState(true);
    this._view?.webview.postMessage({ type: 'user_message', text });

    try {
      const result = await this.client.request<SessionPromptRequest, PromptResponse>('session/prompt', {
        sessionId: this.client.sessionId!,
        prompt: [{ type: 'text', text }]
      });
      this.updateUIState(false);
      this._view?.webview.postMessage({ type: 'stream_end', stopReason: result.stopReason });
    } catch (err: any) {
      this.updateUIState(false);
      this._view?.webview.postMessage({ type: 'error', message: err.message });
    }
  }

  private updateUIState(streaming: boolean): void {
    this.isStreaming = streaming;
    this._view?.webview.postMessage({ type: 'set_ui_state', isStreaming: streaming });
  }

  private handleNotification(notification: any): void {
    if (notification.method === 'session/update') {
      this.handleSessionUpdate(notification.params as SessionUpdate);
    } else if (notification.method === 'session/request_permission') {
       this.handlePermissionRequest(notification.params);
    }
  }

  private handleSessionUpdate(update: SessionUpdate): void {
    if (!this._view) return;
    const u = update.update;

    switch (u.sessionUpdate) {
      case 'agent_message_chunk': {
        const text = u.content?.[0]?.content?.text || '';
        if (text) this._view.webview.postMessage({ type: 'agent_chunk', text });
        break;
      }
      case 'tool_call':
        this._view.webview.postMessage({ type: 'tool_call', toolCallId: u.toolCallId, title: u.title || u.kind || 'Tool', status: u.status });
        break;
      case 'tool_call_update':
        this._view.webview.postMessage({ type: 'tool_call_update', toolCallId: u.toolCallId, status: u.status });
        break;
      case 'plan':
        this._view.webview.postMessage({ type: 'plan', entries: u.entries || [] });
        break;
    }
  }

  private handlePermissionRequest(req: any): void {
    if (!this._view) return;
    vscode.window.showInformationMessage(
      `Vectora wants to execute: ${req.toolCall?.kind || 'unknown'}`,
      { modal: true },
      'Allow', 'Deny'
    ).then(_choice => {
      // Logic for permission handling via client.request
    });
  }

  private getNonce(): string {
    let text = '';
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    for (let i = 0; i < 32; i++) text += possible.charAt(Math.floor(Math.random() * possible.length));
    return text;
  }

  private getHtml(webview: vscode.Webview): string {
    const distPath = vscode.Uri.joinPath(this.context.extensionUri, 'dist', 'webview');
    const scriptUri = webview.asWebviewUri(vscode.Uri.joinPath(distPath, 'assets', 'index.js'));
    const styleUri = webview.asWebviewUri(vscode.Uri.joinPath(distPath, 'assets', 'index.css'));
    
    const nonce = this.getNonce();

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}'; img-src ${webview.cspSource} https:;">
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
