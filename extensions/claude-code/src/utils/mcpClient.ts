import * as cp from "child_process";
import { EventEmitter } from "events";
import { MCPRequest, MCPResponse, MCPTool, ToolCallResult } from "./types";

export class MCPClient extends EventEmitter {
  private process: cp.ChildProcess | null = null;
  private buffer = "";
  private pendingRequests = new Map<
    number,
    { resolve: (v: unknown) => void; reject: (e: Error) => void }
  >();
  private nextId = 1;
  private initialized = false;
  private tools: MCPTool[] = [];

  constructor(private readonly workspacePath: string) {
    super();
  }

  async connect(): Promise<void> {
    if (this.process) return;

    this.process = cp.spawn("vectora", ["mcp", this.workspacePath], {
      stdio: ["pipe", "pipe", "pipe"],
    });

    this.process.stdout!.setEncoding("utf8");
    this.process.stdout!.on("data", (chunk: string) => {
      this.buffer += chunk;
      const lines = this.buffer.split("\n");
      this.buffer = lines.pop() ?? "";
      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;
        try {
          const msg = JSON.parse(trimmed) as MCPResponse;
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
          // ignore non-JSON lines (debug output)
        }
      }
    });

    this.process.stderr!.on("data", (d: Buffer) => {
      this.emit("debug", d.toString());
    });

    this.process.on("exit", () => {
      this.initialized = false;
      this.process = null;
      this.emit("disconnected");
    });

    await this.initialize();
    this.tools = await this.listTools();
    this.initialized = true;
    this.emit("connected", this.tools);
  }

  disconnect(): void {
    this.process?.kill();
    this.process = null;
    this.initialized = false;
  }

  isConnected(): boolean {
    return this.initialized && this.process !== null;
  }

  getTools(): MCPTool[] {
    return this.tools;
  }

  private send<T = unknown>(method: string, params?: Record<string, unknown>): Promise<T> {
    return new Promise((resolve, reject) => {
      const id = this.nextId++;
      const req: MCPRequest = { jsonrpc: "2.0", id, method, params };
      this.pendingRequests.set(id, {
        resolve: resolve as (v: unknown) => void,
        reject,
      });
      this.process?.stdin?.write(JSON.stringify(req) + "\n");

      setTimeout(() => {
        if (this.pendingRequests.has(id)) {
          this.pendingRequests.delete(id);
          reject(new Error(`Timeout waiting for response to ${method}`));
        }
      }, 30_000);
    });
  }

  private async initialize(): Promise<void> {
    await this.send("initialize", {
      protocolVersion: "2024-11-05",
      clientInfo: { name: "vectora-claude-code", version: "0.1.0" },
    });
  }

  private async listTools(): Promise<MCPTool[]> {
    const result = await this.send<{ tools: MCPTool[] }>("tools/list", {});
    return result.tools ?? [];
  }

  async callTool(name: string, args: Record<string, unknown>): Promise<ToolCallResult> {
    const result = await this.send<ToolCallResult>("tools/call", {
      name,
      arguments: args,
    });
    return result;
  }
}
