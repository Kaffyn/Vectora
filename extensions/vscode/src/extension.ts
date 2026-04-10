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
  // Ensure we only ever have one status bar item
  if (!statusBarItem) {
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.command = 'vectora.toggleStatus';
    context.subscriptions.push(statusBarItem);
  }
  
  updateStatusStopped();
  statusBarItem.show();

  // Register Commands
  context.subscriptions.push(
    vscode.commands.registerCommand('vectora.toggleStatus', async () => {
      if (coreClient && coreClient.isRunning) {
        await stopVectora();
      } else {
        await startVectora(context);
      }
    }),

    vscode.commands.registerCommand('vectora.start', async () => {
        await startVectora(context);
    }),

    vscode.commands.registerCommand('vectora.stop', async () => {
        await stopVectora();
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

  // Initial attempt
  startVectora(context);
}

async function startVectora(context: vscode.ExtensionContext) {
  if (!statusBarItem) return;
  
  if (coreClient && coreClient.isRunning) {
      return; 
  }

  statusBarItem.text = '$(sync~spin) Vectora: Starting...';
  statusBarItem.tooltip = 'Vectora is initializing';
  statusBarItem.backgroundColor = undefined;
  statusBarItem.color = undefined;

  try {
    const binPath = await binaryManager.ensureBinary();
    
    // To show the Tray, we need the background core started.
    // We'll run 'vectora start' which is detached.
    const cp = require('child_process');
    cp.spawn(binPath, ['start'], { detached: true, stdio: 'ignore' }).unref();

    // Now connect via 'acp' bridge which talks to the background core
    coreClient = new Client('Vectora Core', binPath, ['acp']);
    
    // Wait a bit for the core tray process to open the socket
    let connected = false;
    for (let i = 0; i < 10; i++) {
        try {
            await coreClient.start();
            connected = true;
            break;
        } catch {
            await new Promise(r => setTimeout(r, 500));
        }
    }

    if (!connected) throw new Error("Core background service didn't start in time.");

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

    statusBarItem.text = '$(check) Vectora: Ready';
    statusBarItem.tooltip = 'Vectora is running. Click to Stop.';
    statusBarItem.color = new vscode.ThemeColor('statusBarItem.prominentForeground');
    statusBarItem.backgroundColor = undefined;

    // Monitor for unexpected exit
    coreClient.onExit.event(() => {
        updateStatusStopped();
    });

  } catch (err: any) {
    statusBarItem.text = `$(error) Vectora: Error`;
    statusBarItem.tooltip = `Error: ${err.message}. Click to retry.`;
    statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.errorBackground');
    statusBarItem.color = undefined;
    
    // Only show error message if it's a manual start attempt? 
    // For now let's keep it visible since user complained about it not starting.
    console.error(`Vectora start error: ${err.message}`);
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
