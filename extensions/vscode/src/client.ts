import * as cp from "child_process";
import * as vscode from "vscode";
import {
  MessageConnection,
  createMessageConnection,
  StreamMessageReader,
  StreamMessageWriter,
  RequestType,
  NotificationType,
  RequestType0,
  NotificationType0,
} from "vscode-jsonrpc/lib/common/connection";

import type {
  RpcRequest,
  RpcResponse,
  RpcNotification,
  SessionPromptRequest,
  PromptResponse,
  SessionUpdate,
  SessionNewRequest,
  SessionNewResponse,
} from "@/types/core";

const DEFAULT_TIMEOUT_MS = 60000;

/**
 * Vectora ACP Client - JSON-RPC 2.0 Protocol
 *
 * Handles communication with Vectora Core binary via JSON-RPC 2.0 over stdio.
 * Implements request/response, notifications, and streaming support.
 */
export class AcpClient {
  private process: cp.ChildProcess | null = null;
  private connection: MessageConnection | null = null;
  private requestId = 0;
  private _isDisposed = false;
  private notificationHandlers = new Map<string, (params: any) => void>();

  // Event emitters
  public readonly onSessionUpdate = new vscode.EventEmitter<SessionUpdate>();
  public readonly onError = new vscode.EventEmitter<string>();
  public readonly onConnectionChange = new vscode.EventEmitter<boolean>();

  constructor(
    private readonly name: string,
    private readonly command: string,
    private readonly args: string[] = [],
    private readonly cwd?: string,
    private readonly env?: Record<string, string>,
  ) {}

  public get isConnected(): boolean {
    return this.process !== null && !this._isDisposed && this.connection !== null;
  }

  /**
   * Conecta ao Core binary e estabelece comunicação JSON-RPC
   */
  public async connect(): Promise<void> {
    if (this._isDisposed) throw new Error(`${this.name} client has been disposed.`);
    if (this.isConnected) return;

    return new Promise((resolve, reject) => {
      try {
        this.process = cp.spawn(this.command, this.args, {
          stdio: ["pipe", "pipe", "pipe"],
          cwd: this.cwd,
          env: { ...process.env, ...this.env },
        });

        this.process.on("error", (err) => {
          const msg = `Failed to start ${this.name}: ${err.message}`;
          console.error(msg);
          this.onError.fire(msg);
          this._isDisposed = true;
          this.onConnectionChange.fire(false);
          reject(new Error(msg));
        });

        this.process.on("exit", (code) => {
          this._isDisposed = true;
          if (this.connection) {
            this.connection.dispose();
            this.connection = null;
          }
          this.onConnectionChange.fire(false);
        });

        // Configurar JSON-RPC
        const reader = new StreamMessageReader(this.process.stdout!);
        const writer = new StreamMessageWriter(this.process.stdin!);
        this.connection = createMessageConnection(reader, writer);

        // Registrar handlers de notificação
        this.setupNotificationHandlers();

        // Handle stderr
        this.process.stderr!.on("data", (data: Buffer) => {
          const msg = data.toString().trim();
          if (msg) {
            console.error(`[${this.name}] ${msg}`);
          }
        });

        // Iniciar conexão
        this.connection.listen();
        this.onConnectionChange.fire(true);
        resolve();
      } catch (err: any) {
        this._isDisposed = true;
        reject(err);
      }
    });
  }

  /**
   * Configura handlers para notificações do Core
   */
  private setupNotificationHandlers(): void {
    if (!this.connection) return;

    // Handle generic notifications
    this.connection.onNotification((method: string, params: any) => {
      // Route to specific handler if registered
      const handler = this.notificationHandlers.get(method);
      if (handler) {
        try {
          handler(params);
        } catch (err) {
          console.error(`Error in notification handler for '${method}':`, err);
        }
      }

      // Also emit SessionUpdate if it matches the pattern
      if (method === "session/update") {
        this.onSessionUpdate.fire(params as SessionUpdate);
      }
    });
  }

  /**
   * Registra handler para notificação específica
   */
  public onNotification(method: string, handler: (params: any) => void): vscode.Disposable {
    this.notificationHandlers.set(method, handler);
    return {
      dispose: () => this.notificationHandlers.delete(method),
    };
  }

  /**
   * Envia requisição RPC e aguarda resposta
   */
  public async request<TParams = any, TResult = any>(
    method: string,
    params?: TParams,
    timeoutMs: number = DEFAULT_TIMEOUT_MS,
  ): Promise<TResult> {
    if (!this.isConnected) throw new Error(`${this.name} is not connected`);
    if (!this.connection) throw new Error(`${this.name} connection not established`);

    try {
      const requestType = new RequestType<TParams, TResult, any>(method);

      const result = await Promise.race([
        this.connection.sendRequest(requestType, params),
        new Promise<TResult>((_, reject) =>
          setTimeout(
            () => reject(new Error(`Request '${method}' timed out after ${timeoutMs}ms`)),
            timeoutMs,
          ),
        ),
      ]);

      return result;
    } catch (err: any) {
      if (err.code !== undefined) {
        throw new Error(`JSON-RPC error ${err.code}: ${err.message}`);
      }
      throw err;
    }
  }

  /**
   * Envia notificação (sem aguardar resposta)
   */
  public notify<TParams = any>(method: string, params?: TParams): void {
    if (!this.isConnected || !this.connection) return;

    try {
      if (params === undefined) {
        const notificationType = new NotificationType0(method);
        this.connection.sendNotification(notificationType);
      } else {
        const notificationType = new NotificationType<TParams>(method);
        this.connection.sendNotification(notificationType, params);
      }
    } catch (err) {
      console.error(`Failed to send notification '${method}':`, err);
    }
  }

  /**
   * Cria nova sessão no Core
   */
  public async createSession(request: SessionNewRequest): Promise<SessionNewResponse> {
    return this.request<SessionNewRequest, SessionNewResponse>("session/new", request);
  }

  /**
   * Envia prompt para processamento
   */
  public async prompt(request: SessionPromptRequest): Promise<PromptResponse> {
    return this.request<SessionPromptRequest, PromptResponse>("session/prompt", request);
  }

  /**
   * Cancela sessão ativa
   */
  public cancelSession(sessionId: string): void {
    this.notify("session/cancel", { sessionId });
  }

  /**
   * Desconecta do Core e limpa recursos
   */
  public disconnect(): void {
    this._isDisposed = true;
    this.notificationHandlers.clear();

    if (this.connection) {
      try {
        this.connection.dispose();
      } catch {
        /* ignore */
      }
      this.connection = null;
    }

    if (this.process) {
      try {
        this.process.kill();
      } catch {
        /* ignore */
      }
      this.process = null;
    }

    this.onConnectionChange.fire(false);
  }

  /**
   * Verifica se cliente foi descartado
   */
  public get isDisposed(): boolean {
    return this._isDisposed;
  }
}
