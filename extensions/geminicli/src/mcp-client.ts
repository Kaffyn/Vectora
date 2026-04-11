import * as cp from "child_process";
import {
  InitializeRequest,
  InitializeResponse,
  ToolsListResponse,
  Tool,
  ToolCallRequest,
  ToolCallResponse,
  JsonRpcRequest,
  JsonRpcResponse,
} from "./types/mcp";

/**
 * McpClient manages the connection to a Vectora MCP server over stdio.
 * Implements the Model Context Protocol (MCP) client side.
 */
export class McpClient {
  private process: cp.ChildProcess | null = null;
  private buffer = "";
  private pendingRequests = new Map<number, { resolve: (value: any) => void; reject: (reason: any) => void }>();
  private nextId = 0;
  private _sessionId: string | undefined;

  public serverInfo: { name: string; version: string } = { name: "unknown", version: "0.0.0" };

  constructor(private corePath: string) {}

  get sessionId(): string | undefined {
    return this._sessionId;
  }

  /**
   * Starts the Vectora MCP server as a subprocess and performs initialization.
   */
  async start(workspacePath: string): Promise<InitializeResponse> {
    this.process = cp.spawn(this.corePath, ["mcp", workspacePath], {
      stdio: ["pipe", "pipe", "pipe"],
      cwd: workspacePath,
    });

    this.process.stdout!.on("data", (data: Buffer) => this.onStdoutData(data));
    this.process.stderr!.on("data", (data: Buffer) => {
      process.stderr.write(`[Vectora MCP] ${data.toString()}`);
    });
    this.process.on("error", (err) => {
      throw new Error(`Failed to start Vectora MCP: ${err.message}`);
    });

    // Perform initialization handshake
    const result = await this.request<InitializeRequest, InitializeResponse>("initialize", {
      protocolVersion: "2024-11-05",
      capabilities: {
        roots: { listChanged: true },
      },
      clientInfo: {
        name: "vectora-geminicli",
        version: "0.1.0",
      },
    });

    this.serverInfo = result.serverInfo;
    return result;
  }

  /**
   * Creates a new session (ACP-specific, wrapped for MCP).
   */
  async newSession(cwd: string): Promise<string> {
    const result = await this.request<{ cwd: string }, { sessionId: string }>("session/new", { cwd });
    this._sessionId = result.sessionId;
    return result.sessionId;
  }

  /**
   * Lists all available MCP tools.
   */
  async listTools(): Promise<Tool[]> {
    const result = await this.request<{}, ToolsListResponse>("tools/list", {});
    return result.tools || [];
  }

  /**
   * Calls an MCP tool by name with the given arguments.
   */
  async callTool(name: string, args: Record<string, any>): Promise<ToolCallResponse> {
    return this.request<ToolCallRequest, ToolCallResponse>("tools/call", {
      name,
      arguments: args,
    });
  }

  /**
   * Sends a prompt to the agent (ACP-specific, wrapped for MCP).
   */
  async prompt(sessionId: string, text: string): Promise<{ stopReason: string }> {
    return this.request("session/prompt", {
      sessionId,
      prompt: [{ type: "text", text }],
    });
  }

  /**
   * Sends a JSON-RPC request and waits for the response.
   */
  private async request<TParams, TResult>(method: string, params: TParams): Promise<TResult> {
    return new Promise<TResult>((resolve, reject) => {
      const id = this.nextId++;
      this.pendingRequests.set(id, { resolve, reject });

      const msg: JsonRpcRequest = { jsonrpc: "2.0", id, method, params };
      this.write(JSON.stringify(msg) + "\n");
    });
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
   */
  private onStdoutData(chunk: Buffer): void {
    this.buffer += chunk.toString();

    while (true) {
      const idx = this.buffer.indexOf("\n");
      if (idx === -1) break;

      const line = this.buffer.substring(0, idx).trim();
      this.buffer = this.buffer.substring(idx + 1);

      if (!line) continue;

      try {
        const msg = JSON.parse(line) as JsonRpcResponse;
        const pending = this.pendingRequests.get(msg.id);
        if (pending) {
          this.pendingRequests.delete(msg.id);
          if (msg.error) {
            pending.reject(new Error(msg.error.message));
          } else {
            pending.resolve(msg.result);
          }
        }
      } catch {
        // Skip malformed lines
      }
    }
  }

  /**
   * Stops the MCP client and kills the subprocess.
   */
  stop(): void {
    this.pendingRequests.forEach((p) => p.reject(new Error("Client stopped")));
    this.pendingRequests.clear();
    this.process?.kill();
    this.process = null;
  }
}
