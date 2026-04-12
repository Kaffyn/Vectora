import * as vscode from "vscode";
import { MCPClient } from "../utils/mcpClient";

export async function vectoraSearch(client: MCPClient): Promise<void> {
  if (!client.isConnected()) {
    vscode.window.showErrorMessage("Vectora não está conectado. Aguarde a inicialização.");
    return;
  }

  // Pega texto selecionado ou pede input
  const editor = vscode.window.activeTextEditor;
  const selection = editor?.document.getText(editor.selection);

  const query = await vscode.window.showInputBox({
    prompt: "O que você quer buscar no codebase?",
    value: selection ?? "",
    placeHolder: "Ex: como o sistema de autenticação funciona?",
  });

  if (!query) return;

  await vscode.window.withProgress(
    { location: vscode.ProgressLocation.Notification, title: "Vectora: buscando..." },
    async () => {
      try {
        const result = await client.callTool("search_database", {
          query,
          workspace: vscode.workspace.workspaceFolders?.[0]?.uri.fsPath ?? "",
        });

        const text = result.content.map((c) => c.text).join("\n");

        // Abre resultado em documento temporário
        const doc = await vscode.workspace.openTextDocument({
          content: `# Vectora Search: ${query}\n\n${text}`,
          language: "markdown",
        });
        await vscode.window.showTextDocument(doc, { preview: true });
      } catch (err) {
        vscode.window.showErrorMessage(`Erro na busca: ${err}`);
      }
    }
  );
}
