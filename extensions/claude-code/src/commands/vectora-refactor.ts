import * as vscode from "vscode";
import { MCPClient } from "../utils/mcpClient";

export async function vectoraRefactor(client: MCPClient): Promise<void> {
  if (!client.isConnected()) {
    vscode.window.showErrorMessage("Vectora não está conectado.");
    return;
  }

  const editor = vscode.window.activeTextEditor;
  if (!editor) {
    vscode.window.showWarningMessage("Abra um arquivo e selecione o código para refatorar.");
    return;
  }

  const selection = editor.document.getText(editor.selection);
  if (!selection) {
    vscode.window.showWarningMessage("Selecione o código que deseja refatorar.");
    return;
  }

  const goal = await vscode.window.showInputBox({
    prompt: "Qual o objetivo da refatoração?",
    placeHolder: "Ex: padronizar tratamento de erros, adicionar logging, melhorar legibilidade",
  });

  if (!goal) return;

  await vscode.window.withProgress(
    { location: vscode.ProgressLocation.Notification, title: "Vectora: refatorando..." },
    async () => {
      try {
        const result = await client.callTool("refactor_with_context", {
          code: selection,
          goal,
          file_path: editor.document.uri.fsPath,
        });

        const refactored = result.content.map((c) => c.text).join("\n");

        // Mostra diff em documento temporário
        const text = [
          "# Vectora: Proposta de Refatoração",
          "",
          `**Objetivo:** ${goal}`,
          "",
          "## Código Original",
          "```",
          selection,
          "```",
          "",
          "## Código Refatorado",
          refactored,
          "",
          "---",
          "_Copie o código refatorado acima e substitua manualmente, ou use Ctrl+Z para desfazer._",
        ].join("\n");

        const doc = await vscode.workspace.openTextDocument({
          content: text,
          language: "markdown",
        });
        await vscode.window.showTextDocument(doc, { preview: true, viewColumn: vscode.ViewColumn.Beside });
      } catch (err) {
        vscode.window.showErrorMessage(`Erro na refatoração: ${err}`);
      }
    }
  );
}
