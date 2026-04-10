import * as vscode from 'vscode';
import { Client } from './client';
import { ChatViewProvider } from './chat-panel';
import { BinaryManager } from './binary-manager';
import { VectoraInlineProvider } from './inline-completion';
import { InitializeRequest, InitializeResponse } from './types/client';

let coreClient: Client | undefined;
let statusBarItem: vscode.StatusBarItem | undefined;
let chatProvider: ChatViewProvider | undefined;
const binaryManager = new BinaryManager();

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  // Create status bar early
  statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  statusBarItem.command = 'vectora.toggleStatus';
  statusBarItem.text = '$(circle-outline) Vectora: Stopped';
  statusBarItem.tooltip = 'Click to Start Vectora';
  statusBarItem.show();
  context.subscriptions.push(statusBarItem);

  // Register Commands
  context.subscriptions.push(
    vscode.commands.registerCommand('vectora.toggleStatus', async () => {
      if (coreClient && coreClient.isRunning) {
        await stopVectora();
      } else {
        await startVectora(context);
      }
    }),

    vscode.commands.registerCommand('vectora.newSession', async () => {
      await vscode.commands.executeCommand('vectora.chatView.focus');
    }),

    vscode.commands.registerCommand('vectora.explainCode', async () => {
      const editor = vscode.window.activeTextEditor;
      if (editor) {
        const selection = editor.document.getText(editor.selection);
        if (selection) {
          await vscode.commands.executeCommand('vectora.chatView.focus');
          if (chatProvider) await chatProvider.sendMessage(`Explain this code:\n\n\`\`\`\n${selection}\n\`\`\``);
        }
      }
    }),

    vscode.commands.registerCommand('vectora.refactorCode', async () => {
      const editor = vscode.window.activeTextEditor;
      if (editor) {
        const selection = editor.document.getText(editor.selection);
        if (selection) {
          await vscode.commands.executeCommand('vectora.chatView.focus');
          if (chatProvider) await chatProvider.sendMessage(`Refactor this code:\n\n\`\`\`\n${selection}\n\`\`\``);
        }
      }
    })
  );

  // Auto-start
  await startVectora(context);
}

async function startVectora(context: vscode.ExtensionContext) {
  if (statusBarItem) {
    statusBarItem.text = '$(sync~spin) Vectora: Starting...';
    statusBarItem.tooltip = 'Vectora is initializing';
  }

  try {
    const binPath = await binaryManager.ensureBinary();
    
    coreClient = new Client('Vectora Core', binPath, ['acp']);
    await coreClient.start();

    // Initialize with Core
    await coreClient.request<InitializeRequest, InitializeResponse>('initialize', {
      protocolVersion: 1,
      clientCapabilities: {
        fs: { readTextFile: true, writeTextFile: true },
        terminal: true,
      },
      clientInfo: {
        name: 'vectora-vscode',
        title: 'Vectora VS Code',
        version: '0.1.0',
      },
    });

    if (!chatProvider) {
      chatProvider = new ChatViewProvider(coreClient, context);
      context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(ChatViewProvider.viewType, chatProvider)
      );
    } else {
      chatProvider.setClient(coreClient);
    }

    // Ghost Text / Inline Completion
    const inlineProvider = new VectoraInlineProvider(coreClient);
    context.subscriptions.push(
      vscode.languages.registerInlineCompletionItemProvider({ pattern: '**' }, inlineProvider)
    );

    if (statusBarItem) {
      statusBarItem.text = '$(check) Vectora: Ready';
      statusBarItem.tooltip = 'Vectora is running. Click to Stop.';
      statusBarItem.color = new vscode.ThemeColor('statusBarItem.prominentForeground');
    }

    // Monitor for unexpected exit
    coreClient.onExit.event(() => {
        updateStatusStopped();
    });

  } catch (err: any) {
    if (statusBarItem) {
        statusBarItem.text = `$(error) Vectora: Error`;
        statusBarItem.tooltip = `Error: ${err.message}. Click to retry.`;
        statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.errorBackground');
    }
    vscode.window.showErrorMessage(`Vectora failed to start: ${err.message}`);
  }
}

async function stopVectora() {
  if (coreClient) {
    coreClient.stop();
    coreClient = undefined;
  }
  updateStatusStopped();
}

function updateStatusStopped() {
    if (statusBarItem) {
        statusBarItem.text = '$(circle-outline) Vectora: Stopped';
        statusBarItem.tooltip = 'Vectora is offline. Click to Start.';
        statusBarItem.color = undefined;
        statusBarItem.backgroundColor = undefined;
    }
}

export function deactivate(): void {
  if (coreClient) coreClient.stop();
  if (statusBarItem) statusBarItem.dispose();
}
