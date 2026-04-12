/**
 * Jest Setup File
 * Configuração global para testes
 */

// Silence console during tests unless explicitly needed
const originalLog = console.log;
const originalWarn = console.warn;
const originalError = console.error;

// Allow debugging with VERBOSE_TESTS=1
if (process.env.VERBOSE_TESTS !== "1") {
  console.log = jest.fn();
  console.warn = jest.fn();
  console.error = jest.fn();
}

// Restore for actual errors
afterAll(() => {
  if (process.env.VERBOSE_TESTS !== "1") {
    console.log = originalLog;
    console.warn = originalWarn;
    console.error = originalError;
  }
});

// Global test utilities
global.TEST_TIMEOUT = 5000;

// Mock VS Code API if not already mocked
if (!global.vscode) {
  global.vscode = {
    window: {
      showErrorMessage: jest.fn(),
      showWarningMessage: jest.fn(),
      showInformationMessage: jest.fn(),
      showInputBox: jest.fn(),
      createOutputChannel: jest.fn(() => ({
        appendLine: jest.fn(),
        append: jest.fn(),
        show: jest.fn(),
        dispose: jest.fn(),
      })),
    },
    workspace: {
      getConfiguration: jest.fn(),
      workspaceFolders: [],
      onDidChangeConfiguration: jest.fn(),
    },
    EventEmitter: class {
      private listeners: any[] = [];
      fire(event: any) {
        this.listeners.forEach((listener) => listener(event));
      }
      get onDidChange() {
        return (listener: any) => {
          this.listeners.push(listener);
          return { dispose: () => {} };
        };
      }
    },
  } as any;
}

/**
 * Test utilities
 */
export function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function waitFor(
  condition: () => boolean,
  timeout: number = 5000,
): Promise<void> {
  const start = Date.now();
  while (!condition()) {
    if (Date.now() - start > timeout) {
      throw new Error("Timeout waiting for condition");
    }
    await delay(50);
  }
}

export function createMockEventEmitter() {
  const listeners = new Map<string, Set<Function>>();

  return {
    on(event: string, listener: Function) {
      if (!listeners.has(event)) {
        listeners.set(event, new Set());
      }
      listeners.get(event)!.add(listener);
      return this;
    },
    emit(event: string, ...args: any[]) {
      if (listeners.has(event)) {
        listeners.get(event)!.forEach((listener) => listener(...args));
      }
      return true;
    },
    off(event: string, listener: Function) {
      if (listeners.has(event)) {
        listeners.get(event)!.delete(listener);
      }
      return this;
    },
    removeAllListeners(event?: string) {
      if (event) {
        listeners.delete(event);
      } else {
        listeners.clear();
      }
      return this;
    },
  };
}

export function createMockStream() {
  const emitter = createMockEventEmitter();
  return {
    on: emitter.on.bind(emitter),
    emit: emitter.emit.bind(emitter),
    off: emitter.off.bind(emitter),
    write: jest.fn(),
    end: jest.fn(),
  };
}

/**
 * Assertion helpers
 */
export function expectDeepEqual(actual: any, expected: any) {
  expect(actual).toEqual(expected);
}

export function expectToContain(array: any[], item: any) {
  expect(array).toContainEqual(item);
}

export function expectToHaveBeenCalledWith(mock: jest.Mock, ...args: any[]) {
  expect(mock).toHaveBeenCalledWith(...args);
}
