import * as vscode from "vscode";
import { MCPClient } from "./utils/mcpClient";
import { VectoraProvider } from "./providers/VectoraProvider";
import { showToolSelector } from "./providers/ToolSelector";
import { vectoraSearch } from "./commands/vectora-search";
import { vectoraDocs } from "./commands/vectora-docs";
import { vectoraAnalyze } from "./commands/vectora-analyze";
import { vectoraRefactor } from "./commands/vectora-refactor";

let client: MCPClient | undefined;
let statusBar: vscode.StatusBarItem;

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  const workspacePath =
    vscode.workspace.workspaceFolders?.[0]?.uri.fsPath ??
    process.env["VECTORA_WORKSPACE"] ??
    "";

  // Status bar
  statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  statusBar.text = "$(loading~spin) Vectora";
  statusBar.tooltip = "Vectora AI - conectando...";
  statusBar.command = "vectora.selectTool";
  statusBar.show();
  context.subscriptions.push(statusBar);

  // MCP Client
  client = new MCPClient(workspacePath);

  client.on("connected", (tools) => {
    statusBar.text = `$(sparkle) Vectora (${tools.length})`;
    statusBar.tooltip = `Vectora AI - ${tools.length} ferramentas disponíveis`;
    vscode.window.showInformationMessage(`Vectora conectado com ${tools.length} ferramentas.`);
  });

  client.on("disconnected", () => {
    statusBar.text = "$(error) Vectora";
    statusBar.tooltip = "Vectora AI - desconectado";
  });

  client.on("debug", (msg: string) => {
    // Só loga se debug habilitado
    if (vscode.workspace.getConfiguration("vectora").get("debug")) {
      console.log("[Vectora debug]", msg);
    }
  });

  // Sidebar provider
  const provider = new VectoraProvider(client);
  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider(VectoraProvider.viewType, provider)
  );

  // Registrar comandos
  context.subscriptions.push(
    vscode.commands.registerCommand("vectora.search", () => vectoraSearch(client!)),
    vscode.commands.registerCommand("vectora.docs", () => vectoraDocs(client!)),
    vscode.commands.registerCommand("vectora.analyze", () => vectoraAnalyze(client!)),
    vscode.commands.registerCommand("vectora.refactor", () => vectoraRefactor(client!)),
    vscode.commands.registerCommand("vectora.selectTool", async () => {
      const selection = await showToolSelector(client!);
      if (!selection) return;

      await vscode.window.withProgress(
        { location: vscode.ProgressLocation.Notification, title: `Vectora: ${selection.tool.name}...` },
        async () => {
          const result = await client!.callTool(selection.tool.name, {
            query: selection.input,
            path: selection.input,
          });
          const text = result.content.map((c) => c.text).join("\n");
          const doc = await vscode.workspace.openTextDocument({
            content: `# Vectora: ${selection.tool.name}\n\n**Input:** ${selection.input}\n\n${text}`,
            language: "markdown",
          });
          await vscode.window.showTextDocument(doc, { preview: true });
        }
      );
    }),
    vscode.commands.registerCommand("vectora.reconnect", async () => {
      client!.disconnect();
      statusBar.text = "$(loading~spin) Vectora";
      await client!.connect();
    })
  );

  // Conectar ao Vectora Core
  if (workspacePath) {
    client.connect().catch((err) => {
      statusBar.text = "$(error) Vectora";
      vscode.window.showWarningMessage(`Vectora: falha ao conectar — ${err.message}`);
    });
  } else {
    statusBar.text = "$(warning) Vectora";
    statusBar.tooltip = "Vectora AI - abra um workspace para conectar";
  }
}

export function deactivate(): void {
  client?.disconnect();
}
