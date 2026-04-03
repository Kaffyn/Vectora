# Plano de Implementação: Camada de API e IPC (Inter-Process Communication)

Este plano detalha a arquitetura de comunicação entre o Daemon Central e suas Interfaces (Web, CLI, MCP), assegurando o fluxo de dados em tempo real e isolamento total de estado.

---

## 1. Definição Tecnológica

O Vectora rejeita o uso de servidores HTTP locais para o core de comunicação devido à latência de handshake TCP e à exposição de portas locais.

- **Windows:** Named Pipes (`\\.\pipe\vectora`).
- **Linux/macOS:** Unix Domain Sockets (`~/.Vectora/run/vectora.sock`).
- **Codificação:** JSON-ND (Newline Delimited) para suporte nativo a streaming sem cabeçalhos pesados.

---

## 2. Estrutura do Servidor IPC (`internal/ipc/server.go`)

O servidor é implementado em Go puro e gerenciado pelo Daemon.

### 2.1 O Ciclo de Vida do Socket

1. **Iniciação:** O Daemon cria o arquivo de socket (ou Named Pipe) no boot.
2. **Registro de Rotas:** O `RegisterRoutes()` mapeia strings de comando (`workspace.query`, etc.) para funções do `internal/core`.
3. **Loop de Aceite:** Escuta e redireciona conexões para goroutines paralelas.
4. **Encerramento:** Garante o `cleanup` do arquivo de socket para evitar conflitos no próximo boot.

---

## 3. Contrato de Mensagens (O Envelope)

Toda mensagem IPC segue o mesmo envelope:

```json
{
  "id": "uuid-v4-string",
  "method": "string.name",
  "type": "request | response | event",
  "payload": { ... },
  "error": null
}
```

### 3.1 Lista de Métodos e payloads Revisitada

| Método                | Payload In           | Resposta                    |
| --------------------- | -------------------- | --------------------------- |
| **`workspace.query`** | `{"ws_id", "query"}` | `{"answer", "sources"}`     |
| **`workspace.index`** | `{"ws_id", "path"}`  | `{"job_id", "status"}`      |
| **`tool.execute`**    | `{"tool", "args"}`   | `{"result", "snapshot_id"}` |
| **`provider.set`**    | `{"name", "key"}`    | `{"configured": true}`      |

---

## 4. O Sistema de Eventos (Push Notifications)

O Daemon pode empurrar eventos para as interfaces sem um pedido prévio (ex: progresso de indexação).

```json
{
  "id": "event-uuid",
  "type": "event",
  "method": "index.progress",
  "payload": { "ws_id": "godot", "percent": 45 }
}
```

---

## 5. Próximos Passos de Implementação (API Refactor)

1.  [ ] **Fallback IPC:** Se o Named Pipe falhar no Windows (ex: privilégio), disparar um fallback para TCP Local-Only via flag.
2.  [ ] **Criptografia Simétrica (Opcional):** Implementar um handshake de chave simples (Handshake Secret) gerado pelo Daemon no boot para encriptar payloads que contenham chaves de API.
3.  [ ] **Multiplexação de Clientes:** Suportar que o CLI e o Web UI estejam conectados simultaneamente e vejam o mesmo histórico de chat em tempo real.

---

## 6. Regras de Negócio (API)

- **RN-API-01:** O payload de erro deve sempre conter um `code` amigável à Web UI (snake_case) e uma `message` legível.
- **RN-API-02:** Conexões inativas por mais de 30 minutos devem ser fechadas para preservar RAM no Daemon.
- **RN-API-03:** Todas as chamadas ao IPC devem ser logadas em `internal/infra/logger.go` para auditoria de segurança (GitBridge).

[Fim do Plano de API - Revisão 2026.04.03]
