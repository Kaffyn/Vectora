# Implementação de Comunicação (API, IPC, e RPC)

Este documento dita a elaboração estruturada das camadas de comunicação do Vectora, fundamentais para a arquitetura Zero-State. Aqui detalhamos o fluxo de trânsito de dados entre o Orquestrador Central (Daemon) e os seus "braços" (Web UI, Terminal CLI e Clientes HTTP externos).

## User Review Required

> [!IMPORTANT]
> Precisamos validar a decisão de uso do pacote `net` padrão em Go para o Named Pipe no Windows. E também validar a taxonomia de Pastas para os roteadores (Handlers) IPC. Analise e aprove antes da execução!

## Proposed Changes

### 1. Servidor e Cliente IPC (`internal/ipc`)

A comunicação base local entre binários do Vectora. Extingue APIs REST arcaicas que custariam performance local desnecessária.

#### [NEW] `internal/ipc/protocol.go`

- Definição estrutural dos payloads (Request, Response, Eventos).
- Definição da constante `Delimiter = '\n'` (Newline-Delimited JSON).

#### [NEW] `internal/ipc/server.go`

- O Coração do Daemon. Escuta em **Unix Sockets** (`~/.Vectora/run/vectora.sock`) ou **Named Pipes** (`\\.\pipe\vectora` no Windows).
- Mantém um mapa em memória de Conexões de Clientes Ativas concorrentemente.
- Fornece módulo Broadcast para o Daemon emitir streaming (como `shell_stream_chunk`).

#### [NEW] `internal/ipc/client.go`

- O consumidor, utilizado pelo Wails (`cmd/vectora-web`) e Bubbletea (`cmd/vectora-cli`).
- Fornece funções atreladas via Future/Promises gerando Requests (UUID) e aguardando na Channel lock correspondente até o daemon responder.

---

### 2. Controladores de Rota (Handlers da API)

Os mapeamentos isolados para processamento de requisição. A placa mãe será responsável por ligar pacotes `internal/core`, `internal/tools` para responder o JSON via `internal/ipc`.

#### [NEW] `internal/ipc/handlers/workspace.go`

- Resolve chamadas `workspace.list`, `workspace.create`, e o poderoso `workspace.query` (Invoca Chunking RAG e LLM Predict).

#### [NEW] `internal/ipc/handlers/provider.go`

- Lida com a gravação segura de `GEMINI_API_KEY` (`provider.set`) e leitura de configuração (`provider.get`).

#### [NEW] `internal/ipc/handlers/session.go`

- Realiza parse do banco `BBolt` para empacotar todo o histórico de conversas em respostas JSON sob o método `session.history`.

---

### 3. Integração Externa (HTTP e RPC)

#### [NEW] `internal/index/client.go` (HTTP API)

- Realizará Requisições HTTPS para o "Vectora Index" (O Marketplace de datasets).
- Funcionalidade principal: `DownloadArchive(id string)`, baixando metadados e arquivos pesados chunk a chunk para evitar consumo brutal de RAM. Retorna percentual emitindo os frames pelo Evento IPC de Progresso na rede.

#### [NEW] `internal/mcp/server.go` (JSON-RPC)

- Modelo Model Context Protocol (Padrão Cursor / Claude).
- Lê o `StdIn/StdOut` em vez de Sockets e encapsula o RAG (`workspace.query`) num Schema de "Tool" inteligível pela IDE do usuário.

## Open Questions

> [!WARNING]
>
> 1. O cliente HTTP de download do Marketplace de Datasets (Index) precisará de Proxy Handling automático pelo OS para redes corporativas rígidas?
> 2. O IPC Protocol precisa suportar SSL Mútuo de criptografia? Como são Sockets estritos no próprio sistema, acreditamos que Sockets locais já são isolados no kernel OS para processos do mesmo User, evitando o gasto brutal com criptografia TLS IPC in-memory. Concorda em fazermos Plaint-text por Sockets UNIX e Named Pipes?

## Verification Plan

### Automated Tests

- Testes unitários puros isolados entre cliente e servidor subindo um TCP emulando Pipe, garantindo que enviar UUID-A recebe exatamente na Channel do cliente UUID-A.
- Execução do Ping-Pong via binários cli/daemon pra verificação manual de Sockets em Background.
