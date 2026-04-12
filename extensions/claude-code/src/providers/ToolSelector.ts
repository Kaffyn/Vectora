import * as vscode from "vscode";
import { MCPClient } from "../utils/mcpClient";
import { MCPTool } from "../utils/types";

export interface ToolSelection {
  tool: MCPTool;
  input: string;
}

export async function showToolSelector(client: MCPClient): Promise<ToolSelection | undefined> {
  const tools = client.getTools();

  if (tools.length === 0) {
    vscode.window.showWarningMessage("Nenhuma ferramenta Vectora disponível.");
    return undefined;
  }

  const items: vscode.QuickPickItem[] = tools.map((t) => ({
    label: `$(tools) ${t.name}`,
    description: t.description,
    detail: Object.keys((t.inputSchema as { properties?: Record<string, unknown> }).properties ?? {})
      .map((k) => `$(symbol-parameter) ${k}`)
      .join("  "),
  }));

  const picked = await vscode.window.showQuickPick(items, {
    title: "Vectora: Selecionar Ferramenta",
    placeHolder: "Escolha uma ferramenta para executar...",
    matchOnDescription: true,
    matchOnDetail: false,
  });

  if (!picked) return undefined;

  const toolName = picked.label.replace("$(tools) ", "");
  const tool = tools.find((t) => t.name === toolName);
  if (!tool) return undefined;

  const input = await vscode.window.showInputBox({
    title: `Vectora: ${tool.name}`,
    prompt: tool.description,
    placeHolder: "Digite o input para a ferramenta...",
  });

  if (input === undefined) return undefined;

  return { tool, input };
}
