import * as vscode from "vscode";
import { Client } from "../acpClient";
import { FSCompletionRequest, FSCompletionResponse } from "../types";

export class VectoraInlineProvider implements vscode.InlineCompletionItemProvider {
  constructor(private client: Client) {}

  async provideInlineCompletionItems(
    document: vscode.TextDocument,
    position: vscode.Position,
    context: vscode.InlineCompletionContext,
    token: vscode.CancellationToken,
  ): Promise<vscode.InlineCompletionItem[] | null> {
    // 1. Verifica se a request foi cancelada pelo VS Code (ex: usuário continuou digitando)
    if (token.isCancellationRequested) {
      return null;
    }

    // Opcional: Ignorar triggers automáticos para economizar API calls
    if (context.triggerKind === vscode.InlineCompletionTriggerKind.Automatic) {
      // return null;
    }

    // 2. Obter contexto do documento
    const prefix = document.getText(new vscode.Range(new vscode.Position(0, 0), position));
    const suffix = document.getText(new vscode.Range(position, document.positionAt(document.getText().length)));

    // Limita o prefixo para não sobrecarregar o contexto (últimos 2000 chars)
    const contextPrefix = prefix.slice(-2000);

    // 3. Chamar o Vectora Core
    try {
      const suggestion = await this.requestCompletion(contextPrefix, suffix, document.languageId, token);

      // Se cancelado durante a espera da rede, retorna null
      if (token.isCancellationRequested || !suggestion || suggestion.trim().length === 0) {
        return null;
      }

      // 4. Retornar o Item de Completions
      return [new vscode.InlineCompletionItem(suggestion, new vscode.Range(position, position))];
    } catch (error) {
      // Silencioso para não poluir logs do usuário com erros de rede frequentes
      console.debug("Vectora inline completion error:", error);
      return null;
    }
  }

  /**
   * Solicita uma completion ao Vectora Core.
   *
   * Nota: Para latência ideal (<300ms), considere uma chamada direta à API do LLM
   * em vez de passar pelo loop completo do agente via ACPClient.
   *
   * TODO: Implementar chamada real — ex: return await this.client.getCompletion(prefix, suffix, lang);
   */
  private async requestCompletion(
    prefix: string,
    suffix: string,
    lang: string,
    token: vscode.CancellationToken,
  ): Promise<string> {
    if (!this.client.sessionId) {
      return "";
    }

    try {
      // Call the real completion endpoint on Vectora Core
      const resp = await this.client.request<FSCompletionRequest, FSCompletionResponse>("fs/completion", {
        sessionId: this.client.sessionId,
        path: vscode.window.activeTextEditor?.document.uri.fsPath || "",
        language: lang,
      });
      const suggestion = resp.content;

      // Check for cancellation after network call
      if (token.isCancellationRequested) {
        return "";
      }

      return suggestion;
    } catch (error) {
      console.debug("Vectora completion error:", error);
      return "";
    }
  }
}
