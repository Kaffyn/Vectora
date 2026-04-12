import * as vscode from "vscode";
import { MCPClient } from "../utils/mcpClient";

export async function vectoraAnalyze(client: MCPClient): Promise<void> {
  if (!client.isConnected()) {
    vscode.window.showErrorMessage("Vectora não está conectado.");
    return;
  }

  const editor = vscode.window.activeTextEditor;
  const filePath = editor?.document.uri.fsPath;

  if (!filePath) {
    vscode.window.showWarningMessage("Abra um arquivo para analisar.");
    return;
  }

  await vscode.window.withProgress(
    { location: vscode.ProgressLocation.Notification, title: "Vectora: analisando bugs e padrões..." },
    async () => {
      try {
        const [bugs, patterns] = await Promise.all([
          client.callTool("bug_pattern_detection", { path: filePath }),
          client.callTool("analyze_code_patterns", { path: filePath, pattern_type: "all" }),
        ]);

        const text = [
          "# Vectora: Análise de Código",
          "",
          `**Arquivo:** \`${filePath}\``,
          "",
          "## Problemas Detectados",
          bugs.content.map((c) => c.text).join("\n"),
          "",
          "## Padrões de Código",
          patterns.content.map((c) => c.text).join("\n"),
        ].join("\n");

        const doc = await vscode.workspace.openTextDocument({
          content: text,
          language: "markdown",
        });
        await vscode.window.showTextDocument(doc, { preview: true });
      } catch (err) {
        vscode.window.showErrorMessage(`Erro na análise: ${err}`);
      }
    }
  );
}
