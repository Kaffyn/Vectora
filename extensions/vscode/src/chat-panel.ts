import * as vscode from 'vscode';
import { ACPClient } from './acp-client';
import { SessionUpdate } from './types/acp';

/**
 * ChatPanel manages the WebView sidebar for Vectora chat.
 * Renders Markdown via marked.js with CSP protection.
 */
export class ChatPanel {
  private panel: vscode.WebviewPanel | undefined;
  private isStreaming = false;
  private currentAgentMessageElementId: string | null = null;

  constructor(
    private client: ACPClient,
    private context: vscode.ExtensionContext
  ) {
    this.client.onSessionUpdate.event(this.handleSessionUpdate, this);
    this.client.onPermissionRequest.event(this.handlePermissionRequest, this);
  }

  public async show(): Promise<void> {
    if (this.panel) {
      this.panel.reveal(vscode.ViewColumn.One);
      return;
    }

    this.panel = vscode.window.createWebviewPanel(
      'vectoraChat',
      'Vectora AI',
      vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [this.context.extensionUri],
      }
    );

    this.panel.webview.html = this.getHtml();
    this.panel.onDidDispose(() => { this.panel = undefined; });

    this.panel.webview.onDidReceiveMessage(async (msg) => {
      switch (msg.type) {
        case 'send':
          await this.sendMessageInternal(msg.text);
          break;
        case 'cancel':
          if (this.client.sessionId) {
            this.client.cancel(this.client.sessionId);
            this.updateUIState(false);
            this.panel?.webview.postMessage({ type: 'stream_end', stopReason: 'cancelled' });
          }
          break;
      }
    });
  }

  /**
   * Public method to send a message programmatically (e.g., from commands).
   */
  public async sendMessage(text: string): Promise<void> {
    await this.show();
    setTimeout(() => {
      this.panel?.webview.postMessage({ type: 'inject_message', text });
    }, 100);
  }

  public async clearChat(): Promise<void> {
    this.panel?.webview.postMessage({ type: 'clear_chat' });
  }

  private async sendMessageInternal(text: string): Promise<void> {
    if (!text.trim()) return;

    if (!this.client.sessionId) {
      const workspacePath = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      if (!workspacePath) {
        vscode.window.showErrorMessage('Vectora: No workspace folder open.');
        return;
      }
      await this.client.newSession(workspacePath);
    }

    this.updateUIState(true);
    this.panel?.webview.postMessage({ type: 'user_message', text });

    try {
      const result = await this.client.prompt(this.client.sessionId!, text);
      this.updateUIState(false);
      this.panel?.webview.postMessage({ type: 'stream_end', stopReason: result.stopReason });
    } catch (err: any) {
      this.updateUIState(false);
      this.panel?.webview.postMessage({ type: 'error', message: err.message });
    }
  }

  private updateUIState(streaming: boolean): void {
    this.isStreaming = streaming;
    this.panel?.webview.postMessage({ type: 'set_ui_state', isStreaming: streaming });
  }

  private handleSessionUpdate(update: SessionUpdate): void {
    if (!this.panel) return;
    const u = update.update;

    switch (u.sessionUpdate) {
      case 'agent_message_chunk': {
        const text = u.content?.[0]?.content?.text || '';
        if (text) this.panel.webview.postMessage({ type: 'agent_chunk', text });
        break;
      }
      case 'tool_call':
        this.panel.webview.postMessage({ type: 'tool_call', toolCallId: u.toolCallId, title: u.title || u.kind || 'Tool', status: u.status });
        break;
      case 'tool_call_update':
        this.panel.webview.postMessage({ type: 'tool_call_update', toolCallId: u.toolCallId, status: u.status });
        break;
      case 'plan':
        this.panel.webview.postMessage({ type: 'plan', entries: u.entries || [] });
        break;
    }
  }

  private handlePermissionRequest(req: any): void {
    if (!this.panel) return;
    vscode.window.showInformationMessage(
      `Vectora wants to execute: ${req.toolCall?.kind || 'unknown'}`,
      { modal: true },
      'Allow', 'Deny'
    );
  }

  private getNonce(): string {
    let text = '';
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    for (let i = 0; i < 32; i++) text += possible.charAt(Math.floor(Math.random() * possible.length));
    return text;
  }

  private getHtml(): string {
    const nonce = this.getNonce();
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; script-src 'nonce-${nonce}' https://cdn.jsdelivr.net;">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    :root {
      --bg: var(--vscode-editor-background);
      --fg: var(--vscode-editor-foreground);
      --input-bg: var(--vscode-input-background);
      --input-fg: var(--vscode-input-foreground);
      --border: var(--vscode-input-border);
      --accent: var(--vscode-button-background);
      --accent-fg: var(--vscode-button-foreground);
      --code-bg: var(--vscode-textCodeBlock-background);
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: var(--vscode-font-family);
      color: var(--fg);
      background: var(--bg);
      display: flex;
      flex-direction: column;
      height: 100vh;
      overflow: hidden;
    }
    #chat-container { flex: 1; overflow-y: auto; padding: 10px; scroll-behavior: smooth; }
    .message-row { margin-bottom: 12px; display: flex; flex-direction: column; }
    .message-row.user { align-items: flex-end; }
    .message-bubble {
      max-width: 90%;
      padding: 8px 12px;
      border-radius: 8px;
      font-size: 13px;
      line-height: 1.5;
      word-wrap: break-word;
    }
    .user .message-bubble {
      background: var(--input-bg);
      border: 1px solid var(--border);
      color: var(--input-fg);
    }
    .agent .message-bubble { background: transparent; padding-left: 0; }
    .markdown-body { font-size: 13px; }
    .markdown-body pre {
      background: var(--code-bg);
      padding: 8px;
      border-radius: 4px;
      overflow-x: auto;
      margin: 4px 0;
    }
    .markdown-body code {
      font-family: var(--vscode-editor-font-family);
      font-size: var(--vscode-editor-font-size);
    }
    .tool-call {
      font-size: 12px;
      background: var(--code-bg);
      border-left: 3px solid var(--accent);
      padding: 6px;
      margin: 4px 0;
      opacity: 0.8;
    }
    #input-area {
      padding: 10px;
      border-top: 1px solid var(--border);
      background: var(--bg);
      display: flex;
      gap: 8px;
    }
    textarea {
      flex: 1;
      background: var(--input-bg);
      color: var(--input-fg);
      border: 1px solid var(--border);
      border-radius: 4px;
      padding: 8px;
      resize: none;
      font-family: inherit;
      min-height: 40px;
      max-height: 150px;
    }
    textarea:focus { outline: 1px solid var(--accent); }
    textarea:disabled { opacity: 0.5; }
    button {
      background: var(--accent);
      color: var(--accent-fg);
      border: none;
      border-radius: 4px;
      padding: 0 16px;
      cursor: pointer;
      font-weight: bold;
    }
    button:disabled { opacity: 0.5; cursor: not-allowed; }
    #cancel-btn {
      background: transparent;
      border: 1px solid var(--vscode-errorForeground);
      color: var(--vscode-errorForeground);
      display: none;
    }
  </style>
</head>
<body>
  <div id="chat-container"></div>
  <div id="input-area">
    <textarea id="user-input" placeholder="Ask Vectora..." rows="1"></textarea>
    <button id="send-btn">Send</button>
    <button id="cancel-btn">Stop</button>
  </div>
  <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
  <script nonce="${nonce}">
    const vscode = acquireVsCodeApi();
    const chat = document.getElementById('chat-container');
    const input = document.getElementById('user-input');
    const sendBtn = document.getElementById('send-btn');
    const cancelBtn = document.getElementById('cancel-btn');
    let agentBubble = null;

    sendBtn.addEventListener('click', () => { send(input.value.trim()); input.value = ''; input.style.height = 'auto'; });
    cancelBtn.addEventListener('click', () => { vscode.postMessage({ type: 'cancel' }); });
    input.addEventListener('keydown', (e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); send(input.value.trim()); input.value = ''; input.style.height = 'auto'; } });
    input.addEventListener('input', function() { this.style.height = 'auto'; this.style.height = Math.min(this.scrollHeight, 150) + 'px'; });

    function send(text) { if (!text) return; vscode.postMessage({ type: 'send', text }); }
    function scrollBottom() { chat.scrollTop = chat.scrollHeight; }

    function createBubble(type, content) {
      const row = document.createElement('div');
      row.className = 'message-row ' + type;
      const bubble = document.createElement('div');
      bubble.className = 'message-bubble ' + type;
      if (type === 'agent') {
        const md = document.createElement('div');
        md.className = 'markdown-body';
        md.innerHTML = content || '';
        bubble.appendChild(md);
        agentBubble = md;
      } else {
        bubble.textContent = content || '';
      }
      row.appendChild(bubble);
      chat.appendChild(row);
      scrollBottom();
      return bubble;
    }

    window.addEventListener('message', (e) => {
      const d = e.data;
      switch (d.type) {
        case 'inject_message':
          input.value = d.text;
          send(d.text);
          break;
        case 'clear_chat':
          chat.innerHTML = '';
          agentBubble = null;
          break;
        case 'user_message':
          agentBubble = null;
          createBubble('user', d.text);
          break;
        case 'stream_start':
          createBubble('agent', '');
          break;
        case 'agent_chunk':
          if (!agentBubble) createBubble('agent', '');
          const txt = (agentBubble.innerText || '') + d.text;
          agentBubble.innerHTML = marked.parse(txt);
          scrollBottom();
          break;
        case 'stream_end':
          const st = document.createElement('div');
          st.style.cssText = 'font-size:10px;opacity:0.5;margin-top:4px;';
          st.textContent = '— ' + d.stopReason + ' —';
          if (agentBubble?.parentElement) agentBubble.parentElement.appendChild(st);
          scrollBottom();
          break;
        case 'set_ui_state':
          sendBtn.style.display = d.isStreaming ? 'none' : 'block';
          cancelBtn.style.display = d.isStreaming ? 'block' : 'none';
          input.disabled = d.isStreaming;
          if (!d.isStreaming) input.focus();
          break;
        case 'error':
          const err = document.createElement('div');
          err.style.cssText = 'color:var(--vscode-errorForeground);padding:10px;';
          err.textContent = 'Error: ' + d.message;
          chat.appendChild(err);
          scrollBottom();
          break;
        case 'tool_call':
          const tc = document.createElement('div');
          tc.className = 'tool-call';
          tc.textContent = '🛠️ ' + (d.title || '') + ' (' + d.status + ')';
          chat.appendChild(tc);
          scrollBottom();
          break;
        case 'plan':
          const pl = document.createElement('div');
          pl.style.cssText = 'border-left:2px solid var(--accent);padding:4px 8px;margin:4px 0;font-size:0.9em;';
          (d.entries || []).forEach(en => {
            const pe = document.createElement('div');
            pe.textContent = (en.status === 'completed' ? '✅' : '⏳') + ' ' + en.content;
            pl.appendChild(pe);
          });
          chat.appendChild(pl);
          scrollBottom();
          break;
      }
    });
  </script>
</body>
</html>`;
  }
}
