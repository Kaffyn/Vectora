import * as vscode from 'vscode';
import { ACPClient } from './acp-client';
import { SessionUpdate, ToolCallStatus } from './types/acp';

/**
 * ChatPanel manages the WebView sidebar for Vectora chat.
 * It renders the chat UI, handles user input, and displays
 * streaming responses and tool call updates from the ACP agent.
 */
export class ChatPanel {
  private panel: vscode.WebviewPanel | undefined;
  private messageBuffer = '';
  private isStreaming = false;

  constructor(
    private client: ACPClient,
    private context: vscode.ExtensionContext
  ) {
    // Wire up ACP client events
    this.client.onSessionUpdate(this.handleSessionUpdate, this);
    this.client.onPermissionRequest(this.handlePermissionRequest, this);
  }

  /**
   * Shows or reveals the chat panel.
   */
  async show(): Promise<void> {
    if (this.panel) {
      this.panel.reveal(vscode.ViewColumn.One);
      return;
    }

    this.panel = vscode.window.createWebviewPanel(
      'vectoraChat',
      'Vectora Chat',
      vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [this.context.extensionUri],
      }
    );

    this.panel.webview.html = this.getHtml();
    this.panel.onDidDispose(() => {
      this.panel = undefined;
    });

    // Handle messages from the webview
    this.panel.webview.onDidReceiveMessage(async (msg) => {
      switch (msg.type) {
        case 'send':
          await this.sendMessage(msg.text);
          break;
        case 'cancel':
          if (this.client.sessionId) {
            this.client.cancel(this.client.sessionId);
            this.panel?.webview.postMessage({ type: 'done', stopReason: 'cancelled' });
            this.isStreaming = false;
            this.messageBuffer = '';
          }
          break;
        case 'permission':
          // Forward permission decision back to ACP client
          // This would need a response mechanism in the ACP protocol
          break;
      }
    });
  }

  /**
   * Sends a user message to the agent.
   */
  private async sendMessage(text: string): Promise<void> {
    if (!this.client.sessionId) {
      const workspacePath = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      if (!workspacePath) {
        vscode.window.showErrorMessage('No workspace folder open. Vectora requires a workspace.');
        return;
      }
      await this.client.newSession(workspacePath);
    }

    // Show user message in UI
    this.panel?.webview.postMessage({ type: 'user_message', text });

    // Clear buffer and start streaming
    this.messageBuffer = '';
    this.isStreaming = true;
    this.panel?.webview.postMessage({ type: 'stream_start' });

    try {
      const result = await this.client.prompt(this.client.sessionId!, text);
      this.isStreaming = false;
      this.panel?.webview.postMessage({ type: 'done', stopReason: result.stopReason });
    } catch (err: any) {
      this.isStreaming = false;
      this.panel?.webview.postMessage({ type: 'error', message: err.message });
    }
  }

  /**
   * Handles session update events from the ACP agent.
   */
  private handleSessionUpdate(update: SessionUpdate): void {
    if (!this.panel || !this.isStreaming) return;

    const u = update.update;
    switch (u.sessionUpdate) {
      case 'agent_message_chunk':
        // Stream text chunks to the UI
        const text = u.content?.[0]?.content?.text || '';
        if (text) {
          this.messageBuffer += text;
          this.panel.webview.postMessage({ type: 'agent_chunk', text });
        }
        break;

      case 'tool_call':
        // Show tool call pending notification
        this.panel.webview.postMessage({
          type: 'tool_call',
          toolCallId: u.toolCallId,
          title: u.title || u.kind || 'Tool',
          status: u.status,
        });
        break;

      case 'tool_call_update':
        // Update tool call status
        this.panel.webview.postMessage({
          type: 'tool_call_update',
          toolCallId: u.toolCallId,
          status: u.status,
          content: u.content,
        });
        break;

      case 'plan':
        // Show plan entries
        this.panel.webview.postMessage({
          type: 'plan',
          entries: u.entries || [],
        });
        break;
    }
  }

  /**
   * Handles permission request events from the ACP agent.
   */
  private handlePermissionRequest(req: { toolCall: { toolCallId: string }; options: any[] }): void {
    if (!this.panel) return;
    this.panel.webview.postMessage({
      type: 'permission_request',
      toolCallId: req.toolCall.toolCallId,
      options: req.options,
    });
  }

  /**
   * Generates the HTML for the chat WebView.
   */
  private getHtml(): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    :root {
      --bg: var(--vscode-editor-background);
      --fg: var(--vscode-editor-foreground);
      --input-bg: var(--vscode-input-background);
      --input-fg: var(--vscode-input-foreground);
      --input-border: var(--vscode-input-border);
      --focus: var(--vscode-focusBorder);
      --user-bg: var(--vscode-input-background);
      --agent-bg: var(--vscode-editor-inactiveSelectionBackground);
      --tool-bg: var(--vscode-textBlockQuote-background);
      --plan-bg: var(--vscode-textLink-foreground);
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: var(--vscode-font-family);
      font-size: var(--vscode-font-size);
      color: var(--fg);
      background: var(--bg);
      display: flex;
      flex-direction: column;
      height: 100vh;
      overflow: hidden;
    }
    #chat {
      flex: 1;
      overflow-y: auto;
      padding: 8px;
      padding-bottom: 60px;
    }
    .message {
      margin: 4px 0;
      padding: 6px 10px;
      border-radius: 6px;
      white-space: pre-wrap;
      word-wrap: break-word;
      line-height: 1.5;
    }
    .user {
      background: var(--user-bg);
      text-align: right;
      margin-left: 20%;
    }
    .agent {
      background: var(--agent-bg);
      margin-right: 10%;
    }
    .tool {
      font-size: 0.85em;
      opacity: 0.75;
      background: var(--tool-bg);
      border-left: 3px solid var(--focus);
    }
    .plan {
      border-left: 2px solid var(--focus);
      padding: 4px 8px;
      margin: 4px 0;
      font-size: 0.9em;
    }
    .plan-entry { padding: 2px 0; }
    .status-done { opacity: 0.5; font-size: 0.8em; text-align: center; padding: 4px; }
    .status-error { color: var(--vscode-errorForeground); font-size: 0.9em; padding: 4px 8px; background: var(--vscode-inputValidation-errorBackground); }
    #input-bar {
      position: fixed;
      bottom: 0;
      left: 0;
      right: 0;
      padding: 8px;
      background: var(--bg);
      border-top: 1px solid var(--input-border);
      display: flex;
      gap: 8px;
    }
    #input {
      flex: 1;
      padding: 6px 10px;
      border: 1px solid var(--input-border);
      border-radius: 4px;
      background: var(--input-bg);
      color: var(--input-fg);
      font-family: inherit;
      font-size: inherit;
      resize: none;
    }
    #input:focus { outline: 1px solid var(--focus); }
    #send-btn {
      padding: 6px 16px;
      border: none;
      border-radius: 4px;
      background: var(--focus);
      color: var(--vscode-button-foreground);
      cursor: pointer;
      font-weight: bold;
    }
    #send-btn:hover { opacity: 0.9; }
    #send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
    #cancel-btn {
      padding: 6px 12px;
      border: 1px solid var(--vscode-errorForeground);
      border-radius: 4px;
      background: transparent;
      color: var(--vscode-errorForeground);
      cursor: pointer;
    }
    .thinking { display: inline-block; animation: pulse 1.5s infinite; }
    @keyframes pulse { 0%, 100% { opacity: 0.4; } 50% { opacity: 1; } }
  </style>
</head>
<body>
  <div id="chat"></div>
  <div id="input-bar">
    <textarea id="input" rows="1" placeholder="Ask Vectora..." onkeydown="handleKey(event)"></textarea>
    <button id="send-btn" onclick="send()">Send</button>
    <button id="cancel-btn" onclick="cancel()" style="display:none;">Cancel</button>
  </div>
  <script>
    const chat = document.getElementById('chat');
    const input = document.getElementById('input');
    const sendBtn = document.getElementById('send-btn');
    const cancelBtn = document.getElementById('cancel-btn');
    const vscode = acquireVsCodeApi();
    let agentMessageEl = null;

    function handleKey(e) {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        send();
      }
    }

    function send() {
      const text = input.value.trim();
      if (!text) return;
      vscode.postMessage({ type: 'send', text });
      input.value = '';
      input.style.height = 'auto';
    }

    function cancel() {
      vscode.postMessage({ type: 'cancel' });
      cancelBtn.style.display = 'none';
      sendBtn.disabled = false;
    }

    window.addEventListener('message', (e) => {
      const d = e.data;
      switch (d.type) {
        case 'user_message':
          if (agentMessageEl) agentMessageEl = null;
          appendMessage(d.text, 'user');
          break;
        case 'stream_start':
          agentMessageEl = appendMessage('', 'agent');
          cancelBtn.style.display = 'inline-block';
          sendBtn.disabled = true;
          break;
        case 'agent_chunk':
          if (!agentMessageEl) agentMessageEl = appendMessage('', 'agent');
          agentMessageEl.textContent += d.text;
          scrollToBottom();
          break;
        case 'done':
          cancelBtn.style.display = 'none';
          sendBtn.disabled = false;
          const statusEl = document.createElement('div');
          statusEl.className = 'status-done';
          statusEl.textContent = '— ' + d.stopReason + ' —';
          chat.appendChild(statusEl);
          scrollToBottom();
          break;
        case 'error':
          cancelBtn.style.display = 'none';
          sendBtn.disabled = false;
          const errEl = document.createElement('div');
          errEl.className = 'status-error';
          errEl.textContent = 'Error: ' + d.message;
          chat.appendChild(errEl);
          scrollToBottom();
          break;
        case 'tool_call':
          const toolEl = document.createElement('div');
          toolEl.className = 'message tool';
          toolEl.textContent = '🔧 ' + esc(d.title) + ' (' + d.status + ')';
          chat.appendChild(toolEl);
          scrollToBottom();
          break;
        case 'tool_call_update':
          // Could update existing tool call element
          break;
        case 'plan':
          const planEl = document.createElement('div');
          planEl.className = 'plan';
          d.entries.forEach(e => {
            const entry = document.createElement('div');
            entry.className = 'plan-entry';
            entry.textContent = (e.status === 'completed' ? '✅' : '⏳') + ' ' + e.content;
            planEl.appendChild(entry);
          });
          chat.appendChild(planEl);
          scrollToBottom();
          break;
      }
    });

    function appendMessage(text, cls) {
      const el = document.createElement('div');
      el.className = 'message ' + cls;
      el.textContent = text;
      chat.appendChild(el);
      scrollToBottom();
      return el;
    }

    function scrollToBottom() {
      chat.scrollTop = chat.scrollHeight;
    }

    function esc(t) {
      const d = document.createElement('div');
      d.textContent = t || '';
      return d.innerHTML;
    }

    // Auto-resize textarea
    input.addEventListener('input', () => {
      input.style.height = 'auto';
      input.style.height = Math.min(input.scrollHeight, 120) + 'px';
    });
  </script>
</body>
</html>`;
  }
}
