import * as cp from "child_process";
import * as vscode from "vscode";

// JSON-RPC 2.0 Base Types
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
 * Used for both Vectora Core (ACP/IPC) and external MCP Servers.
 */
export class Client {
  private process: cp.ChildProcess | null = null;
  private buffer = "";
  private nextId = 0;
  private pendingRequests = new Map<
    number | string,
    {
      resolve: (value: any) => void;
      reject: (reason: any) => void;
      timeout: NodeJS.Timeout;
    }
  >();
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
    return this.process !== null && !this._isDisposed;
  }

  /**
   * Spawns the child process and attaches listeners.
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
          reject(new Error(msg));
        });

        this.process.on("exit", (code) => {
          this._isDisposed = true;
          this.cleanupPendingRequests(`Process ${this.name} exited with code ${code}`);
          this.onExit.fire(code);
        });

        this.process.stdout!.on("data", (data: Buffer) => this.onData(data));
        this.process.stderr!.on("data", (data: Buffer) => {
          const stderr = data.toString().trim();
          if (stderr) console.error(`[${this.name} Stderr]: ${stderr}`);
        });

        // Resolve once process is successfully spawned
        resolve();
      } catch (err: any) {
        reject(err);
      }
    });
  }

  /**
   * Sends a JSON-RPC request and waits for a response.
   */
  public async request<TParams = any, TResult = any>(
    method: string,
    params: TParams,
    timeoutMs: number = DEFAULT_TIMEOUT_MS,
  ): Promise<TResult> {
    if (!this.isRunning) throw new Error(`${this.name} is not running`);

    return new Promise<TResult>((resolve, reject) => {
      const id = this.nextId++;
      const timeout = setTimeout(() => {
        if (this.pendingRequests.has(id)) {
          this.pendingRequests.delete(id);
          reject(new Error(`Request '${method}' to ${this.name} timed out after ${timeoutMs}ms`));
        }
      }, timeoutMs);

      this.pendingRequests.set(id, { resolve, reject, timeout });
      const msg: JsonRpcRequest = { jsonrpc: "2.0", id, method, params };
      this.sendRaw(msg);
    });
  }

  /**
   * Sends a JSON-RPC notification (no response expected).
   */
  public notify<TParams = any>(method: string, params: TParams): void {
    if (!this.isRunning) return;
    const msg: JsonRpcNotification = { jsonrpc: "2.0", method, params };
    this.sendRaw(msg);
  }

  /**
   * Low-level write to stdin.
   */
  protected sendRaw(msg: any): void {
    if (this.process?.stdin?.writable) {
      this.process.stdin.write(JSON.stringify(msg) + "\n");
    }
  }

  private onData(chunk: Buffer): void {
    this.buffer += chunk.toString();
    let idx = this.buffer.indexOf("\n");
    while (idx !== -1) {
      const line = this.buffer.substring(0, idx).trim();
      this.buffer = this.buffer.substring(idx + 1);
      if (!line) continue;

      try {
        const msg = JSON.parse(line);
        this.handleMessage(msg);
      } catch (err) {
        console.warn(`[${this.name}] Failed to parse JSON-RPC line: ${line}`);
      }
      idx = this.buffer.indexOf("\n");
    }
  }

  private handleMessage(msg: any): void {
    if ("id" in msg && msg.id !== undefined && msg.id !== null) {
      // Response handler
      const pending = this.pendingRequests.get(msg.id);
      if (pending) {
        clearTimeout(pending.timeout);
        this.pendingRequests.delete(msg.id);
        if (msg.error) {
          pending.reject(new Error(msg.error.message || "Unknown JSON-RPC error"));
        } else {
          pending.resolve(msg.result);
        }
      }
    } else if ("method" in msg) {
      // Notification handler
      this.onNotification.fire(msg as JsonRpcNotification);
    }
  }

  private cleanupPendingRequests(reason: string): void {
    this.pendingRequests.forEach((p) => {
      clearTimeout(p.timeout);
      p.reject(new Error(reason));
    });
    this.pendingRequests.clear();
  }

  public stop(): void {
    this._isDisposed = true;
    this.cleanupPendingRequests("Client stopped manually");
    if (this.process) {
      this.process.kill();
      this.process = null;
    }
  }
}
