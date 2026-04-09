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

  private getHtml(_webview: vscode.Webview): string {
    const nonce = this.getNonce();
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src vscode-resource: https:; font-src https://cdn.jsdelivr.net; style-src 'unsafe-inline' https://cdn.jsdelivr.net; script-src 'nonce-${nonce}' https://cdn.jsdelivr.net https://unpkg.com;">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <link href="https://cdn.jsdelivr.net/npm/@vscode/codicons/dist/codicon.css" rel="stylesheet" />
  <style>
    :root {
      --bg: var(--vscode-sideBar-background);
      --fg: var(--vscode-sideBar-foreground);
      --input-bg: var(--vscode-input-background);
      --input-fg: var(--vscode-input-foreground);
      --border: var(--vscode-input-border);
      --accent: var(--vscode-button-background);
      --accent-hover: var(--vscode-button-hoverBackground);
      --accent-fg: var(--vscode-button-foreground);
      --code-bg: var(--vscode-textCodeBlock-background);
      --panel-bg: rgba(255, 255, 255, 0.05);
      --glass: rgba(15, 15, 15, 0.6);
      --muted: var(--vscode-descriptionForeground);
    }
    .vscode-light {
      --glass: rgba(240, 240, 240, 0.6);
      --panel-bg: rgba(0, 0, 0, 0.05);
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

    #header {
      padding: 10px 12px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-bottom: 1px solid var(--border);
    }
    .brand { font-size: 13px; font-weight: 700; display: flex; align-items: center; gap: 6px; letter-spacing: 0.8px; text-transform: uppercase; }
    .brand i { color: #4285F4; font-size: 14px; }

    .selectors { display: flex; gap: 4px; }
    select {
      background: var(--input-bg);
      color: var(--fg);
      border: 1px solid var(--border);
      font-size: 11px;
      padding: 2px 4px;
      border-radius: 4px;
      max-width: 80px;
    }

    #mode-bar {
      display: flex;
      padding: 4px 12px;
      gap: 8px;
      border-bottom: 1px solid var(--border);
      background: var(--panel-bg);
    }
    .mode-btn {
      background: none;
      border: none;
      color: var(--muted);
      font-size: 11px;
      padding: 4px 8px;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 4px;
      border-radius: 4px;
    }
    .mode-btn:hover { background: var(--panel-bg); color: var(--fg); }
    .mode-btn.active { color: var(--accent-fg); background: var(--accent); }

    #chat-container {
      flex: 1;
      overflow-y: auto;
      padding: 12px;
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    .message-row { display: flex; flex-direction: column; max-width: 90%; }
    .message-row.user { align-self: flex-end; }
    .message-row.agent { align-self: flex-start; }

    .bubble {
      padding: 8px 12px;
      border-radius: 12px;
      font-size: 13px;
      line-height: 1.5;
    }
    .bubble.user { background: var(--accent); color: var(--accent-fg); border-bottom-right-radius: 2px; }
    .bubble.agent { background: var(--panel-bg); border: 1px solid var(--border); border-bottom-left-radius: 2px; }

    .tool-call {
      font-size: 12px;
      padding: 6px 10px;
      background: var(--glass);
      border: 1px solid var(--border);
      border-left: 3px solid var(--accent);
      border-radius: 4px;
      display: flex;
      align-items: center;
      gap: 8px;
      margin: 4px 0;
    }

    #input-area {
      padding: 12px;
      border-top: 1px solid var(--border);
      background: var(--bg);
    }
    .input-wrapper {
      position: relative;
      background: var(--input-bg);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 8px;
    }
    textarea {
      width: 100%;
      min-height: 40px;
      max-height: 200px;
      background: transparent;
      border: none;
      color: var(--input-fg);
      font-family: inherit;
      resize: none;
      outline: none;
      padding-right: 40px;
    }
    #send-btn, #stop-btn {
      position: absolute;
      right: 8px;
      bottom: 8px;
      background: var(--accent);
      color: var(--accent-fg);
      border: none;
      width: 28px;
      height: 28px;
      border-radius: 6px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    #send-btn:hover { background: var(--accent-hover); }

    #stop-btn {
      right: 40px;
      background: var(--vscode-errorForeground);
      display: none;
    }

    .markdown-body pre {
      background: var(--code-bg);
      padding: 8px;
      border-radius: 6px;
      margin: 8px 0;
      overflow-x: auto;
    }
    .spinner { animation: spin 1s linear infinite; }
    @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
  </style>
</head>
<body>
  <div id="header">
    <div class="header-top">
      <div class="brand"><i class="codicon codicon-circuit-board"></i> VECTORA</div>
      <div class="selectors">
        <select id="policy-select" title="Safety Policy">
          <option value="ask">Ask (Diff)</option>
          <option value="auto">Auto-Edit</option>
          <option value="yolo">YOLO</option>
        </select>
        <select id="provider-select">
          <option value="gemini">Gemini</option>
          <option value="claude">Claude</option>
        </select>
        <select id="model-select">
          <option value="gemini-3-flash">Flash</option>
          <option value="gemini-3.1-pro">Pro</option>
        </select>
      </div>
    </div>
    <div id="mode-bar">
      <button class="mode-btn active" id="mode-fast" title="Quick answers with Gemini Flash">
        <i class="codicon codicon-zap"></i> Fast
      </button>
      <button class="mode-btn" id="mode-planning" title="Multi-step reasoning with Gemini Pro">
        <i class="codicon codicon-layers"></i> Planning
      </button>
    </div>
  </div>
  
  <div id="chat-container"></div>
  
  <div id="input-area">
    <div class="input-wrapper">
      <textarea id="user-input" placeholder="Ask anything..."></textarea>
      <button id="stop-btn" title="Stop generating"><i class="codicon codicon-debug-stop"></i></button>
      <button id="send-btn"><i class="codicon codicon-arrow-up"></i></button>
    </div>
  </div>

  <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
  <script nonce="${nonce}">
    const vscode = acquireVsCodeApi();
    const chat = document.getElementById('chat-container');
    const input = document.getElementById('user-input');
    const sendBtn = document.getElementById('send-btn');
    const stopBtn = document.getElementById('stop-btn');
    const policySelect = document.getElementById('policy-select');
    const providerSelect = document.getElementById('provider-select');
    const modelSelect = document.getElementById('model-select');
    const btnFast = document.getElementById('mode-fast');
    const btnPlanning = document.getElementById('mode-planning');

    let agentBubble = null;
    let mode = 'fast';

    const models = {
      gemini: [
        { id: 'gemini-3-flash', name: 'Flash' },
        { id: 'gemini-3.1-pro', name: 'Pro' }
      ],
      claude: [
        { id: 'claude-4.6-haiku', name: 'Haiku' },
        { id: 'claude-4.6-sonnet', name: 'Sonnet' },
        { id: 'claude-4.6-opus', name: 'Opus' }
      ]
    };

    providerSelect.addEventListener('change', () => {
      const p = providerSelect.value;
      modelSelect.innerHTML = models[p].map(m => '<option value="' + m.id + '">' + m.name + '</option>').join('');
    });

    btnFast.addEventListener('click', () => setMode('fast'));
    btnPlanning.addEventListener('click', () => setMode('planning'));

    function setMode(m) {
      mode = m;
      btnFast.classList.toggle('active', m === 'fast');
      btnPlanning.classList.toggle('active', m === 'planning');
      if (m === 'planning' && providerSelect.value === 'gemini') {
          modelSelect.value = 'gemini-3.1-pro';
      } else if (m === 'fast' && providerSelect.value === 'gemini') {
          modelSelect.value = 'gemini-3-flash';
      }
    }

    sendBtn.addEventListener('click', () => { send(); });
    stopBtn.addEventListener('click', () => { vscode.postMessage({ type: 'cancel' }); });
    input.addEventListener('keydown', (e) => { 
      if (e.key === 'Enter' && !e.shiftKey) { 
        e.preventDefault(); 
        send(); 
      } 
    });

    function send() { 
      const text = input.value.trim();
      if (!text) return; 
      vscode.postMessage({ 
        type: 'send', 
        text,
        provider: providerSelect.value,
        model: modelSelect.value,
        mode: mode,
        policy: policySelect.value
      }); 
      input.value = '';
    }

    function createBubble(type, content) {
      const row = document.createElement('div');
      row.className = 'message-row ' + type;
      const bubble = document.createElement('div');
      bubble.className = 'bubble ' + type;
      if (type === 'agent') {
        const md = document.createElement('div');
        md.className = 'markdown-body';
        md.innerHTML = typeof marked !== 'undefined' ? marked.parse(content || '') : (content || '');
        bubble.appendChild(md);
        agentBubble = md;
      } else {
        bubble.textContent = content || '';
      }
      row.appendChild(bubble);
      chat.appendChild(row);
      chat.scrollTop = chat.scrollHeight;
      return bubble;
    }

    window.addEventListener('message', (e) => {
      const d = e.data;
      switch (d.type) {
        case 'user_message':
          agentBubble = null;
          createBubble('user', d.text);
          break;
        case 'agent_chunk':
          if (!agentBubble) createBubble('agent', '');
          const currentTxt = agentBubble.getAttribute('data-raw') || '';
          const newTxt = currentTxt + d.text;
          agentBubble.setAttribute('data-raw', newTxt);
          if (typeof marked !== 'undefined') {
            agentBubble.innerHTML = marked.parse(newTxt);
          } else {
            agentBubble.textContent = newTxt;
          }
          chat.scrollTop = chat.scrollHeight;
          break;
        case 'set_ui_state':
          sendBtn.disabled = d.isStreaming;
          stopBtn.style.display = d.isStreaming ? 'flex' : 'none';
          input.disabled = d.isStreaming;
          break;
        case 'tool_call':
          const tc = document.createElement('div');
          tc.className = 'tool-call';
          tc.id = 'tc-' + d.toolCallId;
          tc.innerHTML = '<i class="codicon codicon-loading spinner"></i> <span>' + (d.title || 'Using tool...') + '</span>';
          chat.appendChild(tc);
          chat.scrollTop = chat.scrollHeight;
          break;
        case 'tool_call_update':
          const tcu = document.getElementById('tc-' + d.toolCallId);
          if (tcu) {
            if (d.status === 'completed') {
              tcu.querySelector('i').className = 'codicon codicon-check';
              tcu.style.borderColor = '#00FF00';
            } else if (d.status === 'failed') {
              tcu.querySelector('i').className = 'codicon codicon-error';
              tcu.style.borderColor = 'var(--vscode-errorForeground)';
            }
          }
          break;
      }
    });
  </script>
</body>
</html>`;
  }
}
