import * as cp from "child_process";
import * as vscode from "vscode";

// Legacy JSON-RPC 2.0 types for backward compatibility
export interface JsonRpcRequest {
  jsonrpc: "2.0";
  id?: number | string;
  method: string;
  params?: any;
}

export interface JsonRpcResponse {
  jsonrpc: "2.0";
  id: number | string;
  result?: any;
  error?: {
    code: number;
    message: string;
    data?: any;
  };
}

export interface JsonRpcNotification {
  jsonrpc: "2.0";
  method: string;
  params?: any;
}

const DEFAULT_TIMEOUT_MS = 60000;

/**
 * Unified Client for JSON-RPC 2.0 over Stdio.
 * Now uses official vscode-jsonrpc library for robust transport and message handling.
 *
 * Maintains backward compatibility with old interface while using modern library internally.
 */
export class Client {
  private process: cp.ChildProcess | null = null;
  private connection: MessageConnection | null = null;
  private nextId = 0;
  private _isDisposed = false;

  public readonly onNotification = new vscode.EventEmitter<JsonRpcNotification>();
  public readonly onError = new vscode.EventEmitter<string>();
  public readonly onExit = new vscode.EventEmitter<number | null>();

  public sessionId?: string;

  constructor(
    private readonly name: string,
    private readonly command: string,
    private readonly args: string[] = [],
    private readonly cwd?: string,
    private readonly env?: Record<string, string>,
  ) {}

  public get isRunning(): boolean {
    return this.process !== null && !this._isDisposed && this.connection !== null;
  }

  /**
   * Spawns the child process and establishes JSON-RPC connection.
   */
  public async start(): Promise<void> {
    if (this._isDisposed) throw new Error(`${this.name} client has been disposed.`);
    if (this.process) return;

    return new Promise((resolve, reject) => {
      try {
        this.process = cp.spawn(this.command, this.args, {
          stdio: ["pipe", "pipe", "pipe"],
          cwd: this.cwd,
          env: { ...process.env, ...this.env },
        });

        this.process.on("error", (err) => {
          const msg = `Failed to start ${this.name}: ${err.message}`;
          this.onError.fire(msg);
          this._isDisposed = true;
          reject(new Error(msg));
        });

        this.process.on("exit", (code) => {
          this._isDisposed = true;
          if (this.connection) {
            this.connection.dispose();
            this.connection = null;
          }
          this.onExit.fire(code);
        });

        // Setup JSON-RPC connection using vscode-jsonrpc
        const reader = new StreamMessageReader(this.process.stdout!);
        const writer = new StreamMessageWriter(this.process.stdin!);
        this.connection = createMessageConnection(reader, writer);

        // Listen for notifications
        this.connection.onNotification((method, params) => {
          this.onNotification.fire({
            jsonrpc: "2.0",
          });
        });

        // Handle stderr
        this.process.stderr!.on("data", (data: Buffer) => {
          const stderr = data.toString().trim();
          if (stderr) console.error(`[${this.name} Stderr]: ${stderr}`);
        });

        // Start connection and resolve
        this.connection.listen();
        resolve();
      } catch (err: any) {
        this._isDisposed = true;
        reject(err);
      }
    });
  }

  /**
   * Sends a JSON-RPC request and waits for a response.
   * Maintains backward-compatible interface while using vscode-jsonrpc internally.
   */
  public async request<TParams = any, TResult = any>(
    method: string,
    params: TParams,
    timeoutMs: number = DEFAULT_TIMEOUT_MS,
  ): Promise<TResult> {
    if (!this.isRunning) throw new Error(`${this.name} is not running`);
    if (!this.connection) throw new Error(`${this.name} connection not established`);

    try {
      // Create request type dynamically
      const requestType = new RequestType<TParams, TResult, any>(method);

      // Send request with timeout
      const result = await Promise.race([
        this.connection.sendRequest(requestType, params),
        new Promise<TResult>((_, reject) =>
          setTimeout(
            () => reject(new Error(`Request '${method}' timed out after ${timeoutMs}ms`)),
          ),
        ),
      ]);

      return result;
    } catch (err: any) {
      // Convert vscode-jsonrpc errors to standard format
      if (err.code !== undefined) {
        throw new Error(`JSON-RPC error ${err.code}: ${err.message}`);
      }
      throw err;
    }
  }

  /**
   * Sends a JSON-RPC notification (no response expected).
   */
  public notify<TParams = any>(method: string, params: TParams): void {
    if (!this.isRunning) return;
    if (!this.connection) return;

    try {
      const notificationType = new NotificationType<TParams>(method);
      this.connection.sendNotification(notificationType, params);
    } catch (err) {
      console.error(`Failed to send notification '${method}':`, err);
    }
  }

  /**
   * Stops the connection and process.
   */
  public stop(): void {
    this._isDisposed = true;

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
  }
}
