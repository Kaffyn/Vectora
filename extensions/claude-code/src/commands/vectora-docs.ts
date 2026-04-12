import * as vscode from "vscode";
import { MCPClient } from "../utils/mcpClient";

export async function vectoraDocs(client: MCPClient): Promise<void> {
  if (!client.isConnected()) {
    vscode.window.showErrorMessage("Vectora não está conectado.");
    return;
  }

  const editor = vscode.window.activeTextEditor;
  const filePath = editor?.document.uri.fsPath;

  if (!filePath) {
    vscode.window.showWarningMessage("Abra um arquivo para gerar documentação.");
    return;
  }

  await vscode.window.withProgress(
    { location: vscode.ProgressLocation.Notification, title: "Vectora: analisando cobertura de docs..." },
    async () => {
      try {
        const [coverage, patterns] = await Promise.all([
          client.callTool("doc_coverage_analysis", { path: filePath }),
          client.callTool("analyze_code_patterns", { path: filePath, pattern_type: "documentation" }),
        ]);

        const text = [
          "# Vectora: Análise de Documentação",
          "",
          `**Arquivo:** \`${filePath}\``,
          "",
          "## Cobertura de Documentação",
          coverage.content.map((c) => c.text).join("\n"),
          "",
          "## Padrões Detectados",
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
