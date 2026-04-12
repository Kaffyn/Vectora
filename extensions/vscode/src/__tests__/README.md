# Phase 6: Runtime Testing & Vectora Core Integration

Suíte completa de testes para a extensão Vectora VS Code, incluindo testes unitários, de integração e end-to-end.

## Estrutura

### Fixtures (`/fixtures`)

Dados de teste reutilizáveis para toda a suíte:

- **mockCoreResponses.ts**: Respostas simuladas do Vectora Core
  - Session management responses
  - Prompt responses (com e sem tool calls)
  - Stream deltas
  - RPC responses (sucesso e erro)
  - Provider configuration
  - Token usage data

- **testMessages.ts**: Mensagens de teste para interações do usuário
  - User messages de diferentes tipos
  - Conversation sequences
  - System prompts
  - RPC requests
  - Special characters e unicode

- **testWorkspaces.ts**: Dados de workspace simulados
  - Workspaces básicos
  - Projetos React/Node/Fullstack
  - Monorepos
  - Workspaces com RAG habilitado
  - Projetos multi-linguagem

- **index.ts**: Exportação central + builders para dados customizados
  - SessionResponseBuilder
  - PromptResponseBuilder
  - MessageBuilder
  - WorkspaceContextBuilder

### Testes Unitários (`/unit`)

Testes isolados para componentes individuais:

- **client.test.ts**: Testes da classe AcpClient
  - Connection management
  - Session management
  - Prompt requests
  - Streaming
  - RPC communication
  - Notification handling
  - Error handling
  - Message parsing
  - Timeout handling
  - Resource management

- **historyStorage.test.ts**: Testes de persistência de histórico
  - Save/retrieval operations
  - Message history
  - Workspace association
  - Session management
  - Cleanup operations
  - Data integrity
  - Performance
  - Search operations
  - Export/import

- **errorRecovery.test.ts**: Testes de recuperação de erro
  - Connection errors
  - Request errors
  - Session errors
  - Streaming errors
  - Tool execution errors
  - Resource exhaustion
  - Data corruption
  - Error propagation
  - Recovery strategies
  - Graceful degradation

### Testes de Integração (`/integration`)

Testes de fluxos entre componentes:

- **acp-protocol.test.ts**: Testes do protocolo ACP
  - Session lifecycle
  - Prompt flow
  - Tool call handling
  - Streaming integration
  - Message sequences
  - RAG integration
  - Token usage tracking
  - Provider management
  - Notification handling
  - Error handling
  - Request/response matching

- **rag-integration.test.ts**: Testes de RAG
  - Workspace indexing
  - Document retrieval
  - Query processing
  - Context integration
  - Workspace filtering
  - Embedding management
  - Performance
  - Quality metrics
  - RAG com tools
  - User experience

- **streaming.test.ts**: Testes de streaming
  - Stream initialization
  - Content streaming
  - Tool call streaming
  - Token usage streaming
  - Stream completion
  - Error handling
  - Performance
  - UI integration
  - Cancellation
  - Pause/resume
  - Mixed streaming

### Testes End-to-End (`/e2e`)

Testes de fluxos completos:

- **chat-flow.test.ts**: Fluxos de chat completos
  - Simple chat flow
  - Code analysis flow
  - Tool-using flow
  - RAG-enabled flow
  - Streaming response flow
  - Error handling flow
  - Long conversations
  - Session management
  - Workspace context
  - Settings & configuration
  - Real-world scenarios

- **core-connection.test.ts**: Integração com Core binary
  - Core binary discovery
  - Core startup
  - Core communication
  - Session with Core
  - Streaming with Core
  - Core state management
  - Core resource usage
  - Core error handling
  - Core configuration
  - Provider management
  - Performance
  - Integration tests
  - Real-world scenarios

## Rodando os Testes

### Todos os testes

```bash
npm test
```

### Testes específicos

```bash
# Testes unitários
npm test -- unit

# Testes de integração
npm test -- integration

# Testes end-to-end
npm test -- e2e

# Arquivo específico
npm test -- client.test.ts

# Padrão específico
npm test -- --testNamePattern="Session Management"
```

### Com cobertura

```bash
npm test -- --coverage
```

### Modo watch

```bash
npm test -- --watch
```

### Modo verbose

```bash
VERBOSE_TESTS=1 npm test
```

## Escrevendo Testes

### Usando Fixtures

```typescript
import {
  mockPromptResponse,
  mockWorkspaceWithEmbeddings,
  PromptResponseBuilder,
} from "@/fixtures";

describe("My Test", () => {
  it("should handle response", () => {
    const response = new PromptResponseBuilder(mockPromptResponse)
      .withContent("Custom content")
      .withModel("custom-model")
      .build();

    expect(response.content).toBe("Custom content");
  });
});
```

### Criando Builders

```typescript
import { MessageBuilder } from "@/fixtures";

it("should handle message", () => {
  const message = new MessageBuilder("user")
    .withContent("Hello")
    .withName("TestUser")
    .build();

  expect(message.content).toBe("Hello");
});
```

### Usando Test Utilities

```typescript
import { delay, waitFor } from "@/__tests__/setup";

it("should wait for condition", async () => {
  let flag = false;
  setTimeout(() => {
    flag = true;
  }, 100);

  await waitFor(() => flag, 5000);
  expect(flag).toBe(true);
});
```

## Conventions

### Nomeação

- **Test files**: `*.test.ts`
- **Fixtures**: `mock*.ts` ou `test*.ts`
- **Test suites**: Describe blocks por feature
- **Test cases**: Describe behavior clearly

### Estrutura de Teste

```typescript
describe("Feature", () => {
  beforeEach(() => {
    // Setup
  });

  afterEach(() => {
    // Cleanup
  });

  describe("Sub-feature", () => {
    it("should do something", () => {
      // Arrange
      const data = createTestData();

      // Act
      const result = performAction(data);

      // Assert
      expect(result).toBe(expectedValue);
    });
  });
});
```

### Fixtures

- Reutilizáveis e imutáveis
- Documentadas com JSDoc
- Organizadas por tipo
- Incluir variações (sucesso, erro, edge cases)

## Coverage

Alvos de cobertura:

- **Branches**: 60%+
- **Functions**: 60%+
- **Lines**: 60%+
- **Statements**: 60%+

Ver cobertura detalhada em `coverage/index.html` após rodar testes com `--coverage`.

## Debugging

### VS Code Debugger

Adicionar ao `.vscode/launch.json`:

```json
{
  "type": "node",
  "request": "launch",
  "name": "Jest Debug",
  "program": "${workspaceFolder}/node_modules/.bin/jest",
  "args": ["--runInBand", "--no-cache"],
  "console": "integratedTerminal",
  "internalConsoleOptions": "neverOpen"
}
```

### Logs Detalhados

```bash
VERBOSE_TESTS=1 npm test -- client.test.ts --testNamePattern="Session"
```

## CI/CD

Os testes rodam automaticamente em:

- Pull requests
- Commits para main/master
- Builds de release

Ver `.github/workflows/test.yml` para configuração do CI.

## Roadmap

- [ ] Testes com providers reais (mock + integration)
- [ ] Testes de performance e benchmarks
- [ ] Testes de acessibilidade (a11y)
- [ ] Testes visuais (visual regression)
- [ ] Snapshot testing para respostas complexas
- [ ] Mutation testing para qualidade
- [ ] Testes de segurança (OWASP)
- [ ] Load testing para streaming

## Recursos

- [Jest Documentation](https://jestjs.io/)
- [Testing Library](https://testing-library.com/)
- [Vectora Core API](../../../core/AGENTS.md)
- [Extension Development](https://code.visualstudio.com/api)
