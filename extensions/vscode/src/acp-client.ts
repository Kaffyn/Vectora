import * as cp from 'child_process';
import * as vscode from 'vscode';
import {
  InitializeRequest,
  InitializeResponse,
  SessionNewRequest,
  SessionNewResponse,
  SessionPromptRequest,
  PromptResponse,
  SessionUpdate,
  RequestPermissionRequest,
  FSReadRequest,
  FSReadResponse,
  FSWriteRequest,
  JsonRpcRequest,
  JsonRpcResponse,
  JsonRpcNotification,
} from './types/acp';

const REQUEST_TIMEOUT_MS = 60000;

export class ACPClient {
  private process: cp.ChildProcess | null = null;
  private buffer = '';
  private pendingRequests = new Map<number, {
    resolve: (value: any) => void;
    reject: (reason: any) => void;
    timeout: NodeJS.Timeout;
  }>();
  private nextId = 0;
  private _sessionId: string | undefined;
  private _isDisposed = false;

  public readonly onSessionUpdate = new vscode.EventEmitter<SessionUpdate>();
  public readonly onPermissionRequest = new vscode.EventEmitter<RequestPermissionRequest>();
  public readonly onError = new vscode.EventEmitter<string>();
  public readonly onInitialized = new vscode.EventEmitter<InitializeResponse>();
  public readonly onProcessExit = new vscode.EventEmitter<number | null>();

  constructor(private corePath: string) {}

  get sessionId(): string | undefined { return this._sessionId; }
  get isRunning(): boolean { return this.process !== null && !this._isDisposed; }

  async start(workspacePath: string): Promise<InitializeResponse> {
    if (this._isDisposed) throw new Error('Client has been disposed.');

    return new Promise((resolve, reject) => {
      try {
        this.process = cp.spawn(this.corePath, ['acp', workspacePath], {
          stdio: ['pipe', 'pipe', 'pipe'],
          cwd: workspacePath,
          env: { ...process.env },
        });

        this.process.on('error', (err) => {
          const msg = `Failed to start Vectora: ${err.message}`;
          this.onError.fire(msg);
          reject(new Error(msg));
        });

        this.process.on('exit', (code) => {
          this._isDisposed = true;
          this.cleanupPendingRequests(`Process exited with code ${code}`);
          this.onProcessExit.fire(code);
          if (code !== 0 && code !== null) {
            this.onError.fire(`Vectora exited unexpectedly (Code: ${code})`);
          }
        });

        this.process.stdout!.on('data', (data: Buffer) => this.onStdoutData(data));
        this.process.stderr!.on('data', (data: Buffer) => {
          console.error(`Vectora: ${data.toString().trim()}`);
        });

        this.request<InitializeRequest, InitializeResponse>('initialize', {
          protocolVersion: 1,
          clientCapabilities: {
            fs: { readTextFile: true, writeTextFile: true },
            terminal: true,
          },
          clientInfo: {
            name: 'vectora-vscode',
            title: 'Vectora VS Code',
            version: '0.1.0',
          },
        })
        .then((result) => {
          this.onInitialized.fire(result);
          resolve(result);
        })
        .catch((err) => reject(err));

      } catch (err: any) {
        reject(err);
      }
    });
  }

  async newSession(cwd: string): Promise<string> {
    const result = await this.request<SessionNewRequest, SessionNewResponse>('session/new', { cwd });
    this._sessionId = result.sessionId;
    return result.sessionId;
  }

  async prompt(sessionId: string, text: string): Promise<PromptResponse> {
    return this.request<SessionPromptRequest, PromptResponse>('session/prompt', {
      sessionId,
      prompt: [{ type: 'text', text }],
    });
  }

  cancel(sessionId: string): void {
    this.notify('session/cancel', { sessionId });
  }

  async readFile(sessionId: string, path: string): Promise<string> {
    const result = await this.request<FSReadRequest, FSReadResponse>('fs/read_text_file', { sessionId, path });
    return result.content;
  }

  async writeFile(sessionId: string, path: string, content: string): Promise<void> {
    await this.request<FSWriteRequest, void>('fs/write_text_file', { sessionId, path, content });
  }

  async getCompletion(sessionId: string, path: string, prefix: string, suffix: string, language: string): Promise<string> {
    const result = await this.request<FSCompletionRequest, FSCompletionResponse>('fs/completion', {
      sessionId,
      path,
      prefix,
      suffix,
      language,
    });
    return result.content;
  }

  private async request<TParams, TResult>(method: string, params: TParams): Promise<TResult> {
    if (!this.isRunning) throw new Error('ACP Client is not running');

    return new Promise<TResult>((resolve, reject) => {
      const id = this.nextId++;
      const timeout = setTimeout(() => {
        if (this.pendingRequests.has(id)) {
          this.pendingRequests.delete(id);
          reject(new Error(`Request '${method}' timed out after ${REQUEST_TIMEOUT_MS}ms`));
        }
      }, REQUEST_TIMEOUT_MS);

      this.pendingRequests.set(id, { resolve, reject, timeout });
      const msg: JsonRpcRequest = { jsonrpc: '2.0', id, method, params };
      this.write(JSON.stringify(msg) + '\n');
    });
  }

  private notify(method: string, params: any): void {
    if (!this.isRunning) return;
    const msg: JsonRpcNotification = { jsonrpc: '2.0', method, params };
    this.write(JSON.stringify(msg) + '\n');
  }

  private write(data: string): void {
    if (this.process?.stdin?.writable) {
      this.process.stdin.write(data);
    }
  }

  private onStdoutData(chunk: Buffer): void {
    this.buffer += chunk.toString();
    while (true) {
      const idx = this.buffer.indexOf('\n');
      if (idx === -1) break;
      const line = this.buffer.substring(0, idx).trim();
      this.buffer = this.buffer.substring(idx + 1);
      if (!line) continue;
      try {
        const msg = JSON.parse(line) as JsonRpcResponse | JsonRpcNotification;
        this.handleMessage(msg);
      } catch {
        // Skip malformed lines
      }
    }
  }

  private handleMessage(msg: JsonRpcResponse | JsonRpcNotification): void {
    if ('id' in msg && msg.id !== undefined) {
      const pending = this.pendingRequests.get(msg.id as number);
      if (pending) {
        clearTimeout(pending.timeout);
        this.pendingRequests.delete(msg.id as number);
        if (msg.error) {
          pending.reject(new Error(msg.error.message));
        } else {
          pending.resolve(msg.result);
        }
      }
    } else if ('method' in msg) {
      if (msg.method === 'session/update') {
        this.onSessionUpdate.fire(msg.params as SessionUpdate);
      } else if (msg.method === 'session/request_permission') {
        this.onPermissionRequest.fire(msg.params as RequestPermissionRequest);
      }
    }
  }

  private cleanupPendingRequests(reason: string): void {
    this.pendingRequests.forEach((p) => {
      clearTimeout(p.timeout);
      p.reject(new Error(reason));
    });
    this.pendingRequests.clear();
  }

  stop(): void {
    this._isDisposed = true;
    this.cleanupPendingRequests('Client stopped');
    if (this.process) {
      this.process.kill();
      this.process = null;
    }
  }
}
