import * as vscode from 'vscode';
import { ACPClient } from './acp-client';
import { ChatPanel } from './chat-panel';

let client: ACPClient | undefined;
let chatPanel: ChatPanel | undefined;
let statusBarItem: vscode.StatusBarItem | undefined;

/**
 * Activates the Vectora VS Code extension.
 * This is called when VS Code activates the extension (on startup).
 */
export async function activate(context: vscode.ExtensionContext): Promise<void> {
  const config = vscode.workspace.getConfiguration('vectora');
  const corePath = config.get<string>('corePath') || 'vectora';
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];

  if (!workspaceFolder) {
    vscode.window.showErrorMessage('Vectora requires an open workspace folder.');
    return;
  }

  // Create status bar item
  statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  statusBarItem.text = '$(loading~spin) Vectora connecting...';
  statusBarItem.show();
  context.subscriptions.push(statusBarItem);

  try {
    // Initialize ACP client
    client = new ACPClient(corePath);
    const initResult = await client.start(workspaceFolder.uri.fsPath);

    // Update status bar
    statusBarItem.text = `$(check) Vectora (${initResult.agentInfo.name} v${initResult.agentInfo.version})`;
    statusBarItem.tooltip = `Vectora ACP connected\nProvider: ${initResult.agentInfo.name}\nVersion: ${initResult.agentInfo.version}`;

    // Initialize chat panel
    chatPanel = new ChatPanel(client, context);

    // Register commands
    context.subscriptions.push(
      vscode.commands.registerCommand('vectora.newSession', async () => {
        if (client?.sessionId) {
          client.cancel(client.sessionId);
        }
        await chatPanel?.show();
        const newSessionId = await client?.newSession(workspaceFolder.uri.fsPath);
        vscode.window.showInformationMessage(`New Vectora session: ${newSessionId}`);
      }),

      vscode.commands.registerCommand('vectora.explainCode', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
          vscode.window.showWarningMessage('No active editor to explain.');
          return;
        }
        const selection = editor.selection;
        const code = selection.isEmpty
          ? editor.document.getText()
          : editor.document.getText(selection);

        await chatPanel?.show();
        await (chatPanel as any).sendMessage(
          `Explain the following code in detail:\n\n\`\`\`${editor.document.languageId}\n${code}\n\`\`\``
        );
      }),

      vscode.commands.registerCommand('vectora.refactorCode', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
          vscode.window.showWarningMessage('No active editor to refactor.');
          return;
        }
        const selection = editor.selection;
        const code = selection.isEmpty
          ? editor.document.getText()
          : editor.document.getText(selection);

        await chatPanel?.show();
        await (chatPanel as any).sendMessage(
          `Refactor the following code to be cleaner, more efficient, and follow best practices.\nExplain your changes:\n\n\`\`\`${editor.document.languageId}\n${code}\n\`\`\``
        );
      })
    );

    // Auto-open chat panel
    await chatPanel.show();

    // Listen for configuration changes
    context.subscriptions.push(
      vscode.workspace.onDidChangeConfiguration((e) => {
        if (e.affectsConfiguration('vectora')) {
          vscode.window.showInformationMessage('Vectora configuration changed. Reload to apply.');
        }
      })
    );

    // Handle client errors
    client.onError((msg) => {
      statusBarItem!.text = '$(error) Vectora error';
      statusBarItem!.tooltip = msg;
      vscode.window.showErrorMessage(msg);
    });

  } catch (err: any) {
    statusBarItem.text = '$(error) Vectora failed';
    statusBarItem.tooltip = err.message;
    vscode.window.showErrorMessage(`Vectora failed to start: ${err.message}`);
  }

  // Cleanup on deactivate
  context.subscriptions.push({
    dispose: () => client?.stop(),
  });
}

/**
 * Deactivates the Vectora VS Code extension.
 * Called when VS Code deactivates the extension.
 */
export function deactivate(): void {
  client?.stop();
  statusBarItem?.dispose();
}
