/**
 * ACP (Agent Client Protocol) Implementation for VS Code Extension
 *
 * Phase 7F: TypeScript ACP na VS Code Extension
 *
 * Implementa o lado cliente do ACP para que a extensão VS Code
 * se comunique com o Vectora Core funcionando como agente ACP.
 */

import * as vscode from 'vscode';
import { EventEmitter } from 'events';
import { ChildProcessWithoutNullStreams, spawn } from 'child_process';

/**
 * VectoraACPClient - Cliente ACP que se conecta a um agente ACP (Vectora Core)
 */
export class VectoraACPClient extends EventEmitter {
  private messageId = 0;
  private pendingRequests = new Map<number, any>();
  private sessionId: string | null = null;
  private initialized = false;

  constructor(
    private process: ChildProcessWithoutNullStreams,
    private logger: vscode.OutputChannel
  ) {
    super();
    this.setupListeners();
  }

  /**
   * Configurar listeners para mensagens do processo
   */
  private setupListeners(): void {
    this.process.stdout?.on('data', (data: Buffer) => {
      const messages = data.toString().split('\n').filter(line => line.trim());
      messages.forEach(message => {
        try {
          const json = JSON.parse(message);
          this.handleMessage(json);
        } catch (e) {
          this.logger.appendLine(`Erro ao parsear mensagem: ${message}`);
        }
      });
    });

    this.process.stderr?.on('data', (data: Buffer) => {
      this.logger.appendLine(`[STDERR] ${data.toString()}`);
    });
  }

  /**
   * Enviar requisição JSON-RPC ao agente
   */
  private sendRequest(method: string, params?: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const id = ++this.messageId;
      const message = {
        jsonrpc: '2.0',
        method,
        params,
        id
      };

      this.pendingRequests.set(id, { resolve, reject, timeout: setTimeout(() => {
        this.pendingRequests.delete(id);
        reject(new Error(`Timeout na requisição ${method}`));
      }, 30000) });

      this.process.stdin?.write(JSON.stringify(message) + '\n');
      this.logger.appendLine(`[SEND] ${id}: ${method}`);
    });
  }

  /**
   * Tratar mensagens recebidas do agente
   */
  private handleMessage(json: any): void {
    if (json.id && this.pendingRequests.has(json.id)) {
      const pending = this.pendingRequests.get(json.id);
      clearTimeout(pending.timeout);
      this.pendingRequests.delete(json.id);

      if (json.error) {
        this.logger.appendLine(`[ERROR] ${json.error.message}`);
        pending.reject(new Error(json.error.message));
      } else {
        this.logger.appendLine(`[RECV] Resposta recebida`);
        pending.resolve(json.result);
      }
    }
  }

  /**
   * Inicializar conexão ACP com o agente
   */
  async initialize(clientName: string, clientVersion: string): Promise<void> {
    try {
      const response = await this.sendRequest('initialize', {
        protocolVersion: 1,
        clientInfo: {
          name: clientName,
          version: clientVersion
        }
      });

      this.initialized = true;
      this.logger.appendLine(`Agente inicializado: ${response.agentInfo.name}`);
      this.emit('initialized', response);
    } catch (error) {
      this.logger.appendLine(`Erro ao inicializar: ${error}`);
      throw error;
    }
  }

  /**
   * Criar nova sessão com o agente
   */
  async createSession(): Promise<string> {
    if (!this.initialized) {
      throw new Error('Cliente não inicializado');
    }

    try {
      const response = await this.sendRequest('session/new', {});
      this.sessionId = response.sessionId;
      this.logger.appendLine(`Sessão criada: ${this.sessionId}`);
      return this.sessionId;
    } catch (error) {
      this.logger.appendLine(`Erro ao criar sessão: ${error}`);
      throw error;
    }
  }

  /**
   * Enviar prompt ao agente
   */
  async sendPrompt(prompt: string): Promise<any> {
    if (!this.sessionId) {
      throw new Error('Nenhuma sessão ativa');
    }

    try {
      const response = await this.sendRequest('session/prompt', {
        sessionId: this.sessionId,
        prompt: [
          {
            type: 'text',
            text: prompt
          }
        ]
      });

      this.logger.appendLine(`Resposta: ${response.stopReason}`);
      return response;
    } catch (error) {
      this.logger.appendLine(`Erro ao enviar prompt: ${error}`);
      throw error;
    }
  }

  /**
   * Cancelar operação em sessão
   */
  async cancel(): Promise<void> {
    if (!this.sessionId) {
      return;
    }

    try {
      await this.sendRequest('session/cancel', {
        sessionId: this.sessionId
      });
      this.logger.appendLine('Operação cancelada');
    } catch (error) {
      this.logger.appendLine(`Erro ao cancelar: ${error}`);
    }
  }

  /**
   * Fechar cliente e desconectar
   */
  close(): void {
    this.process.kill();
    this.initialized = false;
    this.sessionId = null;
    this.pendingRequests.clear();
    this.logger.appendLine('Cliente ACP desconectado');
  }
}

/**
 * Gerenciador de conexões ACP para a extensão
 */
export class ACPConnectionManager {
  private client: VectoraACPClient | null = null;

  constructor(private logger: vscode.OutputChannel) {}

  /**
   * Criar nova conexão ACP com o Vectora Core
   */
  async connect(agentPath: string): Promise<VectoraACPClient> {
    try {
      const process = spawn(agentPath, ['--log', 'json']);

      this.client = new VectoraACPClient(process, this.logger);
      await this.client.initialize('vectora-vscode', '0.1.0');

      return this.client;
    } catch (error) {
      this.logger.appendLine(`Erro ao conectar: ${error}`);
      throw error;
    }
  }

  /**
   * Desconectar do agente
   */
  disconnect(): void {
    if (this.client) {
      this.client.close();
      this.client = null;
    }
  }

  /**
   * Obter cliente atual
   */
  getClient(): VectoraACPClient | null {
    return this.client;
  }
}
