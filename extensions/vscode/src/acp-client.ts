import * as cp from 'child_process';
import * as vscode from 'vscode';
import {
  AcpClientEvents,
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

/**
 * ACPClient manages the connection to the Vectora ACP server over stdio.
 * It handles JSON-RPC 2.0 message framing, request/response correlation,
 * and event dispatching for session updates and permission requests.
 */
export class ACPClient {
  private process: cp.ChildProcess | null = null;
  private buffer = '';
  private pendingRequests = new Map<number, { resolve: (value: any) => void; reject: (reason: any) => void }>();
  private nextId = 0;
  private _sessionId: string | undefined;

  // Event emitters
  public readonly onSessionUpdate = new vscode.EventEmitter<SessionUpdate>();
  public readonly onPermissionRequest = new vscode.EventEmitter<RequestPermissionRequest>();
  public readonly onError = new vscode.EventEmitter<string>();
  public readonly onInitialized = new vscode.EventEmitter<InitializeResponse>();

  constructor(private corePath: string) {}

  get sessionId(): string | undefined {
    return this._sessionId;
  }

  /**
   * Starts the Vectora ACP server as a subprocess and performs the initialization handshake.
   */
  async start(workspacePath: string): Promise<InitializeResponse> {
    this.process = cp.spawn(this.corePath, ['acp', workspacePath], {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: workspacePath,
      env: { ...process.env },
    });

    this.process.stdout!.on('data', (data: Buffer) => this.onStdoutData(data));
    this.process.stderr!.on('data', (data: Buffer) => {
      vscode.window.showWarningMessage(`Vectora: ${data.toString().trim()}`);
    });
    this.process.on('error', (err) => {
      this.onError.fire(`Failed to start Vectora: ${err.message}`);
    });
    this.process.on('exit', (code) => {
      if (code !== 0 && code !== null) {
        this.onError.fire(`Vectora exited with code ${code}`);
      }
    });

    // Perform initialization handshake
    const result = await this.request<InitializeRequest, InitializeResponse>('initialize', {
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
    });

    this.onInitialized.fire(result);
    return result;
  }

  /**
   * Creates a new ACP session for the given working directory.
   */
  async newSession(cwd: string): Promise<string> {
    const result = await this.request<SessionNewRequest, SessionNewResponse>('session/new', { cwd });
    this._sessionId = result.sessionId;
    return result.sessionId;
  }

  /**
   * Sends a prompt to the agent and returns the stop reason.
   * Responses are streamed via onSessionUpdate events.
   */
  async prompt(sessionId: string, text: string): Promise<PromptResponse> {
    return this.request<SessionPromptRequest, PromptResponse>('session/prompt', {
      sessionId,
      prompt: [{ type: 'text', text }],
    });
  }

  /**
   * Cancels an ongoing prompt for the given session.
   */
  cancel(sessionId: string): void {
    this.notify('session/cancel', { sessionId });
  }

  /**
   * Reads a file from the workspace.
   */
  async readFile(sessionId: string, path: string): Promise<string> {
    const result = await this.request<FSReadRequest, FSReadResponse>('fs/read_text_file', { sessionId, path });
    return result.content;
  }

  /**
   * Writes content to a file in the workspace.
   */
  async writeFile(sessionId: string, path: string, content: string): Promise<void> {
    await this.request<FSWriteRequest, void>('fs/write_text_file', { sessionId, path, content });
  }

  /**
   * Sends a JSON-RPC request and waits for the response.
   */
  private async request<TParams, TResult>(method: string, params: TParams): Promise<TResult> {
    return new Promise<TResult>((resolve, reject) => {
      const id = this.nextId++;
      this.pendingRequests.set(id, { resolve, reject });

      const msg: JsonRpcRequest = { jsonrpc: '2.0', id, method, params };
      this.write(JSON.stringify(msg) + '\n');
    });
  }

  /**
   * Sends a JSON-RPC notification (no response expected).
   */
  private notify(method: string, params: any): void {
    const msg: JsonRpcNotification = { jsonrpc: '2.0', method, params };
    this.write(JSON.stringify(msg) + '\n');
  }

  /**
   * Writes data to the subprocess stdin.
   */
  private write(data: string): void {
    if (this.process?.stdin?.writable) {
      this.process.stdin.write(data);
    }
  }

  /**
   * Processes incoming data from the subprocess stdout.
   * Handles newline-delimited JSON-RPC messages.
   */
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
        // Skip malformed lines (stderr leakage or partial data)
      }
    }
  }

  /**
   * Dispatches incoming messages to the appropriate handler.
   */
  private handleMessage(msg: JsonRpcResponse | JsonRpcNotification): void {
    if ('id' in msg && msg.id !== undefined) {
      // Response to our request
      const pending = this.pendingRequests.get(msg.id as number);
      if (pending) {
        this.pendingRequests.delete(msg.id as number);
        if (msg.error) {
          pending.reject(new Error(msg.error.message));
        } else {
          pending.resolve(msg.result);
        }
      }
    } else if ('method' in msg) {
      // Notification from the agent
      if (msg.method === 'session/update') {
        this.onSessionUpdate.fire(msg.params as SessionUpdate);
      } else if (msg.method === 'session/request_permission') {
        this.onPermissionRequest.fire(msg.params as RequestPermissionRequest);
      }
    }
  }

  /**
   * Stops the ACP client and kills the subprocess.
   */
  stop(): void {
    this.pendingRequests.forEach((p) => p.reject(new Error('Client stopped')));
    this.pendingRequests.clear();
    this.process?.kill();
    this.process = null;
  }
}
