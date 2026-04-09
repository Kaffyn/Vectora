# Blueprint: API & Protocolo ACP (Agent Client Protocol)

**Status:** Implementado  
**Módulo:** `core/api/acp/`  
**Transporte:** JSON-RPC 2.0 sobre Stdio (IPC)

O Vectora utiliza o **Agent Client Protocol (ACP)** como sua espinha dorsal de comunicação. Este protocolo foi projetado para ser leve, assíncrono e compatível com os princípios do MCP (Model Context Protocol), otimizando a interação entre IDEs e o motor agêntico.

---

## 1. Ciclo de Vida da Conexão

A comunicação ocorre via `stdin`/`stdout`, garantindo segurança e baixo overhead.

1.  **Handshake (`initialize`):** O cliente (ex: VS Code) inicia a conexão enviando suas capacidades e recebendo as capacidades do Core.
2.  **Sessão (`session/new`):** Cria um contexto de trabalho isolado para um diretório específico (Workspace).
3.  **Interação (`session/prompt`):** O fluxo principal de chat, onde o cliente envia mensagens e recebe um stream de tokens ou chamadas de ferramentas.

---

## 2. Métodos do Protocolo (JSON-RPC)

### `initialize`

Negocia a versão do protocolo e capacities.

- **Request:** `{"method": "initialize", "params": { "protocolVersion": 1, "clientInfo": {...} }}`
- **Response:** Inclui as capacidades do Agente (se suporta RAG, Terminal, etc).

### `session/new`

Inicia uma nova sessão técnica.

- **Request:** `{"method": "session/new", "params": { "cwd": "c:/path/to/project" }}`
- **Response:** `{"sessionId": "sess_..."}`

### `session/prompt` (Streaming)

Envia uma instrução ao agente.

- **Request:** `{"method": "session/prompt", "params": { "sessionId": "...", "prompt": [...] }}`
- **Notifications:** O servidor envia notificações `session/update` contendo tokens parciais, status de ferramentas ou resultados parciais.

---

## 3. Integração com Tools (ACP -> Tools)

Quando o LLM decide agir, o Core traduz a intenção para uma chamada de método interna no ACP:

### `fs/read_text_file`

Solicitação segura de leitura de arquivo monitorada pelo Guardian.

```json
{
  "method": "fs/read_text_file",
  "params": { "sessionId": "...", "path": "src/main.go" }
}
```

### `terminal/run_command`

Execução de comando via Sub-Agente ou Tool direta.

```json
{
  "method": "terminal/run_command",
  "params": { "sessionId": "...", "command": "go test ./..." }
}
```

---

## 4. O Cliente Unificado (TypeScript)

Na extensão VS Code, a classe `Client` centraliza toda esta lógica, abstraindo o protocolo para os componentes de UI (`ChatPanel`) e editores (`InlineCompletion`).

- **Auto-Gerenciamento de ID:** Atribuídos sequencialmente pelo cliente.
- **Timeouts:** Padrão de 30s para respostas de ferramentas.
- **Session Management:** Mantém o `sessionId` ativo durante todo o ciclo de vida do painel.

---

## 5. Próximas Implementações

- **ACP over WebSocket:** Para permitir Web UIs remotas conectarem ao Core local.
- **Binary Streams:** Para transferência eficiente de arquivos e imagens (snapshots de UI).
