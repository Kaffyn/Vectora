import * as cp from "child_process";
import * as fs from "fs";
import * as path from "path";
import * as os from "os";
import { EventEmitter } from "events";
import { MCPRequest, MCPResponse, MCPTool, ToolCallResult } from "./types";

/** Resolve o caminho do binário vectora em ordem de prioridade */
function resolveBinary(configPath?: string): string {
  // 1. Config explícita (settings.json do VS Code)
  if (configPath && fs.existsSync(configPath)) return configPath;

  // 2. Env var (para CI/custom installs)
  const envPath = process.env["VECTORA_BINARY_PATH"];
  if (envPath && fs.existsSync(envPath)) return envPath;

  // 3. Instalação automática (~/.vectora/bin/vectora)
  const suffix = os.platform() === "win32" ? ".exe" : "";
  const autoInstall = path.join(os.homedir(), ".vectora", "bin", `vectora${suffix}`);
  if (fs.existsSync(autoInstall)) return autoInstall;

  // 4. AppData (Windows — instalador oficial)
  if (os.platform() === "win32") {
    const appData = path.join(
      process.env["LOCALAPPDATA"] ?? "",
      "Programs",
      "Vectora",
      "vectora.exe"
    );
    if (fs.existsSync(appData)) return appData;
  }

  // 5. Fallback: assume no PATH
  return "vectora";
}

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
  private readonly binary: string;

  constructor(
    private readonly workspacePath: string,
    configBinaryPath?: string
  ) {
    super();
    this.binary = resolveBinary(configBinaryPath);
  }

  async connect(): Promise<void> {
    if (this.process) return;

    this.process = cp.spawn(this.binary, ["mcp", this.workspacePath], {
      stdio: ["pipe", "pipe", "pipe"],
    });

    this.process.on("error", (err) => {
      const hint =
        err.message.includes("ENOENT")
          ? `Binário não encontrado: "${this.binary}". Instale o Vectora ou configure "vectora.binaryPath" nas settings.`
          : err.message;
      this.emit("error", new Error(hint));
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
            if (msg.error) pending.reject(new Error(msg.error.message));
            else pending.resolve(msg.result);
          }
        } catch {
          // linha de debug do servidor — ignorar
        }
      }
    });

    this.process.stderr!.on("data", (d: Buffer) => {
      this.emit("debug", d.toString());
    });

    this.process.on("exit", (code) => {
      this.initialized = false;
      this.process = null;
      this.emit("disconnected", code);
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

  getBinaryPath(): string {
    return this.binary;
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
          reject(new Error(`Timeout (30s) aguardando resposta de "${method}"`));
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
    return this.send<ToolCallResult>("tools/call", { name, arguments: args });
  }
}
