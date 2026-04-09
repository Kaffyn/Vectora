# Plano de Implementação: Estratégia de Testes Avançada (Fase 1)

Este plano detalha a implementação da Fase 1 da nova estratégia de testes, focando em robustez no parsing de LLMs e simulação de estado no servidor ACP.

## Propostas de Mudanças

### 1. Fixtures de LLM

#### [NEW] `core/llm/fixtures/gemini_response.json`

- Adicionar um exemplo real de JSON retornado pela API do Gemini.

#### [NEW] `core/llm/fixtures/claude_tool_call.json`

- Adicionar um exemplo de JSON do Claude contendo um bloco `tool_use`.

### 2. Testes de Parsing de Provedores

#### [NEW] `core/llm/provider_parsing_test.go`

- Implementar testes que usam `httptest.NewServer` para servir os arquivos de fixture e validar que os métodos `Complete` decodificam os dados corretamente para as structs internas.

### 3. Stateful ACP Mock

#### [NEW] `core/api/acp/stateful_mock.go`

- Criar a struct `StatefulMockEngine` que implementa a interface `Engine`.
- **Estado Interno**:
  - `Files map[string]string`: Simula o sistema de arquivos.
  - `Sessions map[string][]string`: Histórico de mensagens por sessão.
- **Comportamento**:
  - `WriteFile`: Atualiza o mapa de arquivos.
  - `ReadFile`: Retorna o conteúdo do mapa.
  - `Query`: Simula um loop de pensamento com atraso (`time.Sleep`).

### 4. Atualização de Testes de Integração

#### [MODIFY] `core/api/acp/server_test.go`

- Substituir o uso de `mockEngine` pelo `StatefulMockEngine` para garantir que o fluxo de leitura/escrita de arquivos via protocolo JSON-RPC seja testado com persistência de estado.

## Verificação

### Automated Tests

- `go test ./core/llm/...`
- `go test ./core/api/acp/...`
