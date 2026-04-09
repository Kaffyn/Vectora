import * as vscode from 'vscode';
import { ACPClient } from './acp-client';
import { ChatPanel } from './chat-panel';
import { InitializeResponse } from './types/acp';

let client: ACPClient | undefined;
let chatPanel: ChatPanel | undefined;
let statusBarItem: vscode.StatusBarItem | undefined;
let currentSessionId: string | undefined;

/**
 * Activates the Vectora VS Code extension.
 * This is called when VS Code activates the extension (on startup).
 */
export async function activate(context: vscode.ExtensionContext): Promise<void> {
  const config = vscode.workspace.getConfiguration('vectora');
  const corePath = config.get<string>('corePath') || 'vectora';
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];

  if (!workspaceFolder) {
    vscode.window.showErrorMessage('Vectora requires an open workspace folder to start.');
    return;
  }

  statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  statusBarItem.text = '$(loading~spin) Vectora starting...';
  statusBarItem.tooltip = 'Initializing Vectora Core...';
  statusBarItem.show();
  context.subscriptions.push(statusBarItem);

  try {
    client = new ACPClient(corePath);
    const initResult = await client.start(workspaceFolder.uri.fsPath);
    currentSessionId = await client.newSession(workspaceFolder.uri.fsPath);

    updateStatusBar(true, initResult.agentInfo);

    chatPanel = new ChatPanel(client, context);

    context.subscriptions.push(
      vscode.commands.registerCommand('vectora.newSession', async () => {
        if (client && currentSessionId) {
          client.cancel(currentSessionId);
        }
        currentSessionId = await client?.newSession(workspaceFolder.uri.fsPath);
        vscode.window.showInformationMessage('New Vectora session started.');
        await chatPanel?.clearChat();
        await chatPanel?.show();
      }),

      vscode.commands.registerCommand('vectora.explainCode', async () => {
        const codeContext = getCodeContext();
        if (!codeContext) return;
        await chatPanel?.show();
        await chatPanel?.sendMessage(
          `Explain the following code in detail:\n\n\`\`\`${codeContext.language}\n${codeContext.code}\n\`\`\``
        );
      }),

      vscode.commands.registerCommand('vectora.refactorCode', async () => {
        const codeContext = getCodeContext();
        if (!codeContext) return;
        await chatPanel?.show();
        await chatPanel?.sendMessage(
          `Refactor the following code to be cleaner, more efficient, and follow best practices.\nExplain your changes:\n\n\`\`\`${codeContext.language}\n${codeContext.code}\n\`\`\``
        );
      })
    );

    context.subscriptions.push(
      vscode.workspace.onDidChangeConfiguration((e) => {
        if (e.affectsConfiguration('vectora.corePath')) {
          vscode.window.showWarningMessage('Changes to Core Path require reloading VS Code.');
        }
      })
    );

    client.onError.event((msg) => {
      updateStatusBar(false, undefined, msg);
      vscode.window.showErrorMessage(`Vectora: ${msg}`);
    });

  } catch (err: any) {
    console.error('Vectora activation error:', err);
    updateStatusBar(false, undefined, err.message);
    vscode.window.showErrorMessage(`Vectora failed to start: ${err.message}`);
  }
}

export function deactivate(): void {
  if (client) client.stop();
  if (statusBarItem) statusBarItem.dispose();
}

function updateStatusBar(isConnected: boolean, agentInfo?: { name: string; version: string }, errorMsg?: string): void {
  if (!statusBarItem) return;
  if (errorMsg) {
    statusBarItem.text = '$(error) Vectora Error';
    statusBarItem.tooltip = errorMsg;
  } else if (isConnected && agentInfo) {
    statusBarItem.text = `$(check) Vectora: ${agentInfo.name}`;
    statusBarItem.tooltip = `Vectora Active\nProvider: ${agentInfo.name}\nVersion: ${agentInfo.version}`;
  } else {
    statusBarItem.text = '$(circle-outline) Vectora Disconnected';
    statusBarItem.tooltip = 'Vectora is not connected';
  }
}

function getCodeContext(): { code: string; language: string } | null {
  const editor = vscode.window.activeTextEditor;
  if (!editor) {
    vscode.window.showWarningMessage('No active editor found.');
    return null;
  }
  const selection = editor.selection;
  const code = selection.isEmpty ? editor.document.getText() : editor.document.getText(selection);
  if (!code.trim()) {
    vscode.window.showWarningMessage('No code selected.');
    return null;
  }
  return { code, language: editor.document.languageId };
}
