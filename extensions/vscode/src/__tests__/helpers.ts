/**
 * Test Helpers & Utilities
 * Funções auxiliares para testes
 */

import type {
  RpcRequest,
  RpcResponse,
  SessionPromptRequest,
  PromptResponse,
  VectoraMessage,
} from "@/types/core";

/**
 * Simula um processo de requisição/resposta RPC
 */
export class MockRpcConnection {
  private requestHandlers = new Map<string, (req: any) => any>();
  private notificationHandlers = new Map<string, (params: any) => void>();

  /**
   * Registra handler para um método
   */
  onRequest(method: string, handler: (req: any) => any) {
    this.requestHandlers.set(method, handler);
    return this;
  }

  /**
   * Registra handler para notificação
   */
  onNotification(method: string, handler: (params: any) => void) {
    this.notificationHandlers.set(method, handler);
    return this;
  }

  /**
   * Simula recebimento de requisição
   */
  async handleRequest(request: RpcRequest): Promise<RpcResponse> {
    const handler = this.requestHandlers.get(request.method);

    if (!handler) {
      return {
        jsonrpc: "2.0",
        id: request.id,
        error: {
          code: -32601,
          message: "Method not found",
        },
      };
    }

    try {
      const result = handler(request.params);
      return {
        jsonrpc: "2.0",
        id: request.id,
        result,
      };
    } catch (error: any) {
      return {
        jsonrpc: "2.0",
        id: request.id,
        error: {
          code: -32603,
          message: error.message,
        },
      };
    }
  }

  /**
   * Simula envio de notificação
   */
  handleNotification(method: string, params: any) {
    const handler = this.notificationHandlers.get(method);
    if (handler) {
      handler(params);
    }
  }
}

/**
 * Mock para session do Core
 */
export class MockCoreSession {
  sessionId: string;
  messages: VectoraMessage[] = [];
  tokenCount = 0;

  constructor(sessionId: string) {
    this.sessionId = sessionId;
  }

  addMessage(message: VectoraMessage) {
    this.messages.push(message);
  }

  getMessages(): VectoraMessage[] {
    return this.messages;
  }

  updateTokenCount(inputTokens: number, outputTokens: number) {
    this.tokenCount += inputTokens + outputTokens;
  }

  reset() {
    this.messages = [];
    this.tokenCount = 0;
  }
}

/**
 * Helpers para assertions
 */
export const TestAssertions = {
  /**
   * Valida estrutura de RPC request
   */
  isValidRpcRequest(req: any): boolean {
    return (
      req.jsonrpc === "2.0" &&
      typeof req.id !== "undefined" &&
      typeof req.method === "string" &&
      typeof req.params === "object"
    );
  },

  /**
   * Valida estrutura de RPC response
   */
  isValidRpcResponse(res: any): boolean {
    return (
      res.jsonrpc === "2.0" &&
      typeof res.id !== "undefined" &&
      (typeof res.result !== "undefined" || typeof res.error !== "undefined")
    );
  },

  /**
   * Valida mensagem
   */
  isValidMessage(msg: any): boolean {
    return (
      typeof msg.role === "string" &&
      ["user", "assistant", "system", "tool"].includes(msg.role) &&
      typeof msg.content === "string"
    );
  },

  /**
   * Valida resposta de prompt
   */
  isValidPromptResponse(res: any): boolean {
    return (
      typeof res.sessionId === "string" &&
      typeof res.model === "string" &&
      typeof res.content === "string" &&
      (typeof res.stopReason === "string" ||
        typeof res.stopReason === "undefined")
    );
  },

  /**
   * Compara mensagens ignorando ordem
   */
  messagesEqual(msgs1: VectoraMessage[], msgs2: VectoraMessage[]): boolean {
    if (msgs1.length !== msgs2.length) return false;
    return msgs1.every((msg, idx) =>
      JSON.stringify(msg) === JSON.stringify(msgs2[idx]),
    );
  },
};

/**
 * Helpers para timeouts e delays
 */
export const TimeoutHelpers = {
  /**
   * Aguarda condição com timeout
   */
  async waitUntil(
    condition: () => boolean,
    timeout: number = 5000,
    interval: number = 50,
  ): Promise<void> {
    const start = Date.now();
    while (!condition()) {
      if (Date.now() - start > timeout) {
        throw new Error(`Timeout waiting for condition (${timeout}ms)`);
      }
      await new Promise((resolve) => setTimeout(resolve, interval));
    }
  },

  /**
   * Aguarda tempo específico
   */
  delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  },

  /**
   * Race between promise e timeout
   */
  async withTimeout<T>(
    promise: Promise<T>,
    timeout: number = 5000,
  ): Promise<T> {
    const timeoutPromise = new Promise<never>((_, reject) =>
      setTimeout(() => reject(new Error("Timeout")), timeout),
    );
    return Promise.race([promise, timeoutPromise]);
  },
};

/**
 * Helpers para mocking
 */
export const MockHelpers = {
  /**
   * Cria mock de EventEmitter
   */
  createEventEmitter() {
    const listeners = new Map<string, Set<Function>>();

    return {
      on(event: string, listener: Function) {
        if (!listeners.has(event)) {
          listeners.set(event, new Set());
        }
        listeners.get(event)!.add(listener);
      },
      emit(event: string, ...args: any[]) {
        if (listeners.has(event)) {
          listeners.get(event)!.forEach((l) => l(...args));
        }
      },
      once(event: string, listener: Function) {
        const wrapper = (...args: any[]) => {
          listener(...args);
          this.off(event, wrapper);
        };
        this.on(event, wrapper);
      },
      off(event: string, listener: Function) {
        if (listeners.has(event)) {
          listeners.get(event)!.delete(listener);
        }
      },
      removeAllListeners(event?: string) {
        if (event) {
          listeners.delete(event);
        } else {
          listeners.clear();
        }
      },
    };
  },

  /**
   * Cria mock de stream
   */
  createMockStream() {
    const emitter = this.createEventEmitter();
    return {
      on: emitter.on.bind(emitter),
      once: emitter.once.bind(emitter),
      emit: emitter.emit.bind(emitter),
      off: emitter.off.bind(emitter),
      write: jest.fn().mockResolvedValue(undefined),
      end: jest.fn().mockResolvedValue(undefined),
      destroy: jest.fn(),
    };
  },

  /**
   * Cria mock de processo
   */
  createMockProcess() {
    return {
      stdin: this.createMockStream(),
      stdout: this.createMockStream(),
      stderr: this.createMockStream(),
      on: jest.fn(),
      kill: jest.fn(),
      pid: 12345,
    };
  },
};

/**
 * Helpers para data generation
 */
export const DataGenerators = {
  /**
   * Gera session ID único
   */
  generateSessionId(): string {
    return `test-session-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  },

  /**
   * Gera mensagem aleatória
   */
  generateMessage(role: VectoraMessage["role"]): VectoraMessage {
    const contents = [
      "Hello, how can I help?",
      "This is a test message",
      "I need your assistance with something",
      "Can you analyze this code?",
      "What do you think about this approach?",
    ];

    return {
      role,
      content: contents[Math.floor(Math.random() * contents.length)],
    };
  },

  /**
   * Gera conversação aleatória
   */
  generateConversation(turns: number): VectoraMessage[] {
    const messages: VectoraMessage[] = [];
    for (let i = 0; i < turns; i++) {
      messages.push(this.generateMessage(i % 2 === 0 ? "user" : "assistant"));
    }
    return messages;
  },

  /**
   * Gera resposta de prompt aleatória
   */
  generatePromptResponse(override?: Partial<PromptResponse>): PromptResponse {
    return {
      sessionId: this.generateSessionId(),
      model: "test-model",
      content: "Test response content",
      usage: {
        inputTokens: 100,
        outputTokens: 50,
        totalTokens: 150,
      },
      stopReason: "end_turn",
      ...override,
    };
  },
};

/**
 * Test context manager
 */
export class TestContext {
  private data = new Map<string, any>();
  private cleanup: (() => void)[] = [];

  /**
   * Armazena dado de teste
   */
  set(key: string, value: any) {
    this.data.set(key, value);
  }

  /**
   * Recupera dado de teste
   */
  get(key: string): any {
    return this.data.get(key);
  }

  /**
   * Registra função de cleanup
   */
  onCleanup(fn: () => void) {
    this.cleanup.push(fn);
  }

  /**
   * Executa cleanup
   */
  async runCleanup() {
    for (const fn of this.cleanup.reverse()) {
      try {
        await fn();
      } catch (error) {
        console.error("Cleanup error:", error);
      }
    }
    this.data.clear();
    this.cleanup = [];
  }
}

/**
 * Performance measurement
 */
export class PerformanceTracker {
  private marks = new Map<string, number>();
  private durations = new Map<string, number[]>();

  /**
   * Marca início
   */
  start(label: string) {
    this.marks.set(label, performance.now());
  }

  /**
   * Marca fim e registra duração
   */
  end(label: string): number {
    const start = this.marks.get(label);
    if (!start) {
      throw new Error(`No start mark for ${label}`);
    }

    const duration = performance.now() - start;
    if (!this.durations.has(label)) {
      this.durations.set(label, []);
    }
    this.durations.get(label)!.push(duration);
    this.marks.delete(label);

    return duration;
  }

  /**
   * Obtém estatísticas
   */
  stats(label: string) {
    const durations = this.durations.get(label) || [];
    if (durations.length === 0) {
      return null;
    }

    const sorted = [...durations].sort((a, b) => a - b);
    const sum = durations.reduce((a, b) => a + b, 0);

    return {
      min: sorted[0],
      max: sorted[sorted.length - 1],
      avg: sum / durations.length,
      median: sorted[Math.floor(sorted.length / 2)],
      p95: sorted[Math.floor(sorted.length * 0.95)],
      p99: sorted[Math.floor(sorted.length * 0.99)],
      count: durations.length,
    };
  }
}
