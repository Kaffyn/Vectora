import * as vscode from "vscode";
import { spawn } from "child_process";
import { AcpClient } from "./core/acpClient";
import { ChatViewProvider } from "./core/providers/chatViewProvider";
import { BinaryManager } from "./core/binaryManager";
import { VectoraInlineProvider } from "./core/providers/inlineCompletionProvider";

let coreClient: AcpClient | undefined;
let statusBarItem: vscode.StatusBarItem | undefined;
let chatProvider: ChatViewProvider | undefined;
let backgroundProcessPid: number | undefined;
const binaryManager = new BinaryManager();

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  // Ensure we only ever have one status bar item
  if (!statusBarItem) {
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.command = "vectora.toggleStatus";
    context.subscriptions.push(statusBarItem);
  }

  updateStatusStopped();
  statusBarItem.show();

  // Register Chat View early to avoid "view not found" errors.
  // Client starts undefined and is set later via setClient() after core connects.
  chatProvider = new ChatViewProvider(undefined, context);
  context.subscriptions.push(vscode.window.registerWebviewViewProvider(ChatViewProvider.viewType, chatProvider));

  // Register Commands
  context.subscriptions.push(
    vscode.commands.registerCommand("vectora.toggleStatus", async () => {
      if (coreClient && coreClient.isConnected) {
        await stopVectora();
      } else {
        await startVectora(context);
      }
    }),

    vscode.commands.registerCommand("vectora.start", async () => {
      await startVectora(context);
    }),

    vscode.commands.registerCommand("vectora.stop", async () => {
      await stopVectora();
    }),

    vscode.commands.registerCommand("vectora.newSession", async () => {
      await vscode.commands.executeCommand("vectora.chatView.focus");
    }),

    vscode.commands.registerCommand("vectora.explainCode", async () => {
      const editor = vscode.window.activeTextEditor;
      if (editor) {
        const selection = editor.document.getText(editor.selection);
        if (selection) {
          await vscode.commands.executeCommand("vectora.chatView.focus");
          if (chatProvider) await chatProvider.sendMessage(`Explain this code:\n\n\`\`\`\n${selection}\n\`\`\``);
        }
      }
    }),

    vscode.commands.registerCommand("vectora.refactorCode", async () => {
      const editor = vscode.window.activeTextEditor;
      if (editor) {
        const selection = editor.document.getText(editor.selection);
        if (selection) {
          await vscode.commands.executeCommand("vectora.chatView.focus");
          if (chatProvider) await chatProvider.sendMessage(`Refactor this code:\n\n\`\`\`\n${selection}\n\`\`\``);
        }
      }
    }),

    vscode.commands.registerCommand("vectora.selectProvider", async () => {
      const providers = [
        { label: "🔵 Gemini", value: "gemini" },
        { label: "🅰️ Claude", value: "claude" },
        { label: "🔴 OpenAI", value: "openai" },
      ];

      const selected = await vscode.window.showQuickPick(providers, {
        title: "Select AI Provider",
        placeHolder: "Choose a provider...",
      });

      if (selected) {
        const config = vscode.workspace.getConfiguration("vectora");
        await config.update("defaultProvider", selected.value, vscode.ConfigurationTarget.Global);
        vscode.window.showInformationMessage(`Provider changed to ${selected.label}`);
      }
    }),
  );

  // Initial attempt
  startVectora(context);
}

async function startVectora(context: vscode.ExtensionContext) {
  if (!statusBarItem) return;

  if (coreClient && coreClient.isConnected) {
    return;
  }

  statusBarItem.text = "$(sync~spin) Vectora: Starting...";
  statusBarItem.tooltip = "Vectora is initializing";
  statusBarItem.backgroundColor = undefined;
  statusBarItem.color = undefined;

  try {
    const binPath = await binaryManager.ensureBinary();

    // To show the Tray, we need the background core started.
    // We'll run 'vectora start' which is detached.
    const bgProcess = spawn(binPath, ["start"], { detached: true, stdio: "ignore" });
    backgroundProcessPid = bgProcess.pid;
    bgProcess.unref();

    // Now connect via 'acp' bridge which talks to the background core
    coreClient = new AcpClient("Vectora Core", binPath, ["acp"]);

    // Wait a bit for the core tray process to open the socket
    let connected = false;
    for (let i = 0; i < 10; i++) {
      try {
        await coreClient.connect();
        connected = true;
        break;
      } catch {
        await new Promise((r) => setTimeout(r, 500));
      }
    }

    if (!connected) throw new Error("Core background service didn't start in time.");

    if (chatProvider) {
      chatProvider.setClient(coreClient);
    }

    // Ghost Text / Inline Completion
    const inlineProvider = new VectoraInlineProvider(coreClient);
    context.subscriptions.push(
      vscode.languages.registerInlineCompletionItemProvider({ pattern: "**" }, inlineProvider),
    );

    updateStatusRunning();

    // Monitor for unexpected disconnection
    coreClient.onConnectionChange.event((isConnected) => {
      if (!isConnected) {
        updateStatusStopped();
      }
    });
  } catch (err: any) {
    statusBarItem.text = `$(error) Vectora: Error`;
    statusBarItem.tooltip = `Error: ${err.message}. Click to retry.`;
    statusBarItem.backgroundColor = new vscode.ThemeColor("statusBarItem.errorBackground");
    statusBarItem.color = undefined;

    // Only show error message if it's a manual start attempt?
    // For now let's keep it visible since user complained about it not starting.
    console.error(`Vectora start error: ${err.message}`);
  }
}

async function stopVectora() {
  if (coreClient) {
    coreClient.disconnect();
    coreClient = undefined;
  }

  // Kill the background process if we stored its PID
  if (backgroundProcessPid) {
    try {
      // On Windows: use taskkill; on Unix: use process.kill()
      const isWin = process.platform === "win32";
      if (isWin) {
        // Kill by PID on Windows
        spawn("taskkill", ["/PID", String(backgroundProcessPid), "/F"], {
          stdio: "ignore"
        }).unref();
      } else {
        // On Unix, kill the process group (negative PID kills entire group)
        process.kill(-backgroundProcessPid, "SIGTERM");
      }
    } catch (err) {
      // Ignore errors if process already terminated
    }
    backgroundProcessPid = undefined;
  }

  // As fallback, kill any remaining vectora processes by name
  // (handles case where PID tracking failed or process renamed)
  try {
    await binaryManager.killAllVectoraProcesses();
  } catch {
    // Ignore errors — process might not exist
  }

  updateStatusStopped();
}

function updateStatusRunning() {
  if (statusBarItem) {
    const config = vscode.workspace.getConfiguration("vectora");
    const provider = config.get<string>("defaultProvider") || "gemini";
    const providerEmoji = {
      gemini: "🔵",
      claude: "🅰️",
      openai: "🔴",
    }[provider] || "🤖";

    statusBarItem.text = `${providerEmoji} Vectora: Running`;
    statusBarItem.tooltip = `Connected to ${provider}. Click to select provider.`;
    statusBarItem.command = "vectora.selectProvider";
    statusBarItem.color = new vscode.ThemeColor("statusBar.foreground");
    statusBarItem.backgroundColor = undefined;
  }
}

function updateStatusStopped() {
  if (statusBarItem) {
    statusBarItem.text = "$(circle-outline) Vectora: Stopped";
    statusBarItem.tooltip = "Vectora is offline. Click to Start.";
    statusBarItem.command = "vectora.toggleStatus";
    statusBarItem.color = undefined;
    statusBarItem.backgroundColor = undefined;
  }
}

export function deactivate(): void {
  if (coreClient) coreClient.disconnect();
  if (backgroundProcessPid) {
    try {
      const isWin = process.platform === "win32";
      if (isWin) {
        spawn("taskkill", ["/PID", String(backgroundProcessPid), "/F"], {
          stdio: "ignore"
        }).unref();
      } else {
        process.kill(-backgroundProcessPid, "SIGTERM");
      }
    } catch {
      // Ignore errors
    }
  }
  if (statusBarItem) statusBarItem.dispose();
}
