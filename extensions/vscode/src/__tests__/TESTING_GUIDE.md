# Testing Guide - Vectora Phase 6

Guia prático para escrever e executar testes na suíte Phase 6.

## Quick Start

### Instalar dependências de teste

```bash
npm install --save-dev jest ts-jest @types/jest
```

### Rodar todos os testes

```bash
npm test
```

### Rodar testes específicos

```bash
# Apenas testes unitários
npm run test:unit

# Apenas testes de integração
npm run test:integration

# Apenas testes e2e
npm run test:e2e

# Arquivo específico
npm test -- client.test.ts

# Padrão específico
npm test -- --testNamePattern="Session Management"

# Com watch mode
npm run test:watch

# Com cobertura
npm run test:coverage
```

## Escrevendo Testes

### 1. Teste Unitário Básico

```typescript
import { mockPromptResponse } from "@/fixtures";

describe("Client", () => {
  it("should handle prompt response", () => {
    // Arrange
    const response = mockPromptResponse;

    // Act
    const result = parseResponse(response);

    // Assert
    expect(result.content).toBe(response.content);
  });
});
```

### 2. Usando Fixtures com Builders

```typescript
import { PromptResponseBuilder, mockPromptResponse } from "@/fixtures";

describe("Streaming", () => {
  it("should handle streaming response", () => {
    const response = new PromptResponseBuilder(mockPromptResponse)
      .withContent("Streaming content...")
      .addTokenUsage({ inputTokens: 200, outputTokens: 300 })
      .build();

    expect(response.content).toContain("Streaming");
  });
});
```

### 3. Teste de Integração

```typescript
import { MockRpcConnection } from "@/__tests__/helpers";
import { mockSessionNewResponse, mockPromptResponse } from "@/fixtures";

describe("ACP Protocol", () => {
  it("should handle session creation", async () => {
    const rpc = new MockRpcConnection();

    rpc.onRequest("session.new", () => mockSessionNewResponse);
    rpc.onRequest("session.prompt", () => mockPromptResponse);

    const sessionReq = {
      jsonrpc: "2.0",
      id: 1,
      method: "session.new",
      params: { workspaceId: "test" },
    };

    const response = await rpc.handleRequest(sessionReq);

    expect(response.result).toEqual(mockSessionNewResponse);
  });
});
```

### 4. Teste com Timeouts e Async

```typescript
import { TimeoutHelpers } from "@/__tests__/helpers";

describe("Async Operations", () => {
  it("should timeout long operations", async () => {
    const slowOperation = new Promise((resolve) =>
      setTimeout(() => resolve("done"), 10000),
    );

    await expect(
      TimeoutHelpers.withTimeout(slowOperation, 1000),
    ).rejects.toThrow("Timeout");
  });

  it("should wait for condition", async () => {
    let value = false;
    setTimeout(() => {
      value = true;
    }, 100);

    await TimeoutHelpers.waitUntil(() => value);
    expect(value).toBe(true);
  });
});
```

### 5. Teste com Mocks de Process

```typescript
import { MockHelpers } from "@/__tests__/helpers";

describe("Core Process", () => {
  it("should handle process events", () => {
    const mockProcess = MockHelpers.createMockProcess();
    const handler = jest.fn();

    mockProcess.stdout.on("data", handler);
    mockProcess.stdout.emit("data", Buffer.from("test"));

    expect(handler).toHaveBeenCalledWith(Buffer.from("test"));
  });
});
```

### 6. Teste com Performance Tracking

```typescript
import { PerformanceTracker } from "@/__tests__/helpers";

describe("Performance", () => {
  it("should track operation speed", () => {
    const tracker = new PerformanceTracker();

    tracker.start("operation");
    // ... fazer algo
    tracker.end("operation");

    const stats = tracker.stats("operation");
    expect(stats!.avg).toBeLessThan(100); // < 100ms
  });
});
```

### 7. Teste com Test Context

```typescript
import { TestContext } from "@/__tests__/helpers";

describe("Complex Setup", () => {
  let ctx: TestContext;

  beforeEach(() => {
    ctx = new TestContext();
    ctx.set("sessionId", "test-123");
  });

  afterEach(async () => {
    await ctx.runCleanup();
  });

  it("should use context", () => {
    const sessionId = ctx.get("sessionId");
    expect(sessionId).toBe("test-123");
  });
});
```

## Padrões Comuns

### Pattern: Setup e Teardown

```typescript
describe("Feature", () => {
  let client: AcpClient;
  let mockRpc: MockRpcConnection;

  beforeEach(() => {
    mockRpc = new MockRpcConnection();
    client = new AcpClient(mockRpc);
  });

  afterEach(async () => {
    await client.dispose();
  });

  it("should do something", async () => {
    // Test code
  });
});
```

### Pattern: Parameterized Tests

```typescript
describe("Multiple Providers", () => {
  const providers = ["anthropic", "openai", "ollama"];

  providers.forEach((provider) => {
    it(`should work with ${provider}`, () => {
      const config = { provider };
      // Test with provider
    });
  });
});
```

### Pattern: Mocking HTTP Calls

```typescript
describe("API Integration", () => {
  beforeEach(() => {
    jest.spyOn(global, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ success: true })),
    );
  });

  it("should fetch data", async () => {
    const result = await fetchData();
    expect(result).toEqual({ success: true });
  });
});
```

### Pattern: Testing Errors

```typescript
describe("Error Handling", () => {
  it("should handle connection error", async () => {
    const rpc = new MockRpcConnection();
    rpc.onRequest("test", () => {
      throw new Error("Connection failed");
    });

    const response = await rpc.handleRequest({
      jsonrpc: "2.0",
      id: 1,
      method: "test",
      params: {},
    });

    expect(response.error).toBeDefined();
    expect(response.error.message).toContain("Connection");
  });
});
```

## Debugging Tests

### Rodar teste específico em debug

```bash
npm run test:debug -- --testNamePattern="Session Management"
```

Então abrir `chrome://inspect` no Chrome para debug.

### Verbose output

```bash
VERBOSE_TESTS=1 npm test -- client.test.ts
```

### Log específico

```typescript
it("should debug", () => {
  console.log("Debug info"); // Visível se VERBOSE_TESTS=1
  expect(true).toBe(true);
});
```

## Coverage

### Gerar relatório de cobertura

```bash
npm run test:coverage
```

Então abrir `coverage/index.html` no navegador.

### Aumentar cobertura

1. Identificar linhas não cobertas em `coverage/index.html`
2. Adicionar testes para cobrir essas linhas
3. Rodar coverage novamente

## Performance Testing

```typescript
import { PerformanceTracker } from "@/__tests__/helpers";

describe("Performance", () => {
  let tracker: PerformanceTracker;

  beforeEach(() => {
    tracker = new PerformanceTracker();
  });

  it("should be fast", async () => {
    tracker.start("operation");
    await someOperation();
    const duration = tracker.end("operation");

    expect(duration).toBeLessThan(1000); // < 1s
  });

  it("should show stats", () => {
    for (let i = 0; i < 10; i++) {
      tracker.start("op");
      someQuickOperation();
      tracker.end("op");
    }

    const stats = tracker.stats("op");
    console.log(`Avg: ${stats.avg}ms, P95: ${stats.p95}ms`);
  });
});
```

## Best Practices

### 1. Um conceito por teste

```typescript
// ✓ Bom
it("should parse valid JSON", () => {
  const result = parseJSON('{"key": "value"}');
  expect(result.key).toBe("value");
});

// ✗ Ruim
it("should parse JSON and validate schema", () => {
  const result = parseJSON('{"key": "value"}');
  expect(result.key).toBe("value");
  expect(validateSchema(result)).toBe(true);
});
```

### 2. Nomes descritivos

```typescript
// ✓ Bom
it("should timeout after 5 seconds if no response", () => {
  // Test code
});

// ✗ Ruim
it("should timeout", () => {
  // Test code
});
```

### 3. Fixtures reutilizáveis

```typescript
// ✓ Bom - Use fixtures
const response = new PromptResponseBuilder(mockPromptResponse)
  .withContent("Custom")
  .build();

// ✗ Ruim - Hardcode
const response = {
  sessionId: "test",
  model: "test",
  content: "Custom",
};
```

### 4. Limpe recursos

```typescript
// ✓ Bom
afterEach(async () => {
  await client.dispose();
  jest.clearAllMocks();
});

// ✗ Ruim - Deixa lixo
// (sem cleanup)
```

## Troubleshooting

### Problema: Testes timeout

**Solução**: Aumentar timeout global ou específico

```typescript
jest.setTimeout(10000); // 10 segundos

it("should do something slow", async () => {
  // ...
}, 15000); // 15 segundos para este teste
```

### Problema: Mocks não funcionam

**Solução**: Garantir que jest.mock() é no escopo certo

```typescript
// ✓ Correto - mock no topo
jest.mock("vscode");

describe("Test", () => {
  it("should use mock", () => {
    expect(vscode.mock).toBeDefined();
  });
});
```

### Problema: Testes passam isolados mas falham juntos

**Solução**: Usar beforeEach/afterEach corretamente

```typescript
beforeEach(() => {
  jest.clearAllMocks();
  // Reset estado global
});
```

## CI/CD Integration

Testes rodam automaticamente em:

- PRs para `main`/`master`
- Commits diretos
- Build pre-release

Ver `.github/workflows/test.yml` para detalhes.

## Próximos Passos

1. Implementar testes práticos com dados reais
2. Adicionar testes de performance
3. Implementar testes visuais
4. Adicionar mutation testing
5. Integrar com ferramentas de qualidade (SonarQube)
