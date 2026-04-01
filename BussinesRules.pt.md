# REGRAS DE NEGÓCIO: VECTORA — CONTRATO DE GOVERNANÇA

> [!TIP]
> Leia esse arquivo em outro idioma.
> Português [BUSINESS_RULES.pt.md] | Inglês [BUSINESS_RULES.md]

Este documento estabelece as fronteiras arquiteturais, contratos de API interna e regras de negócio obrigatórias do Vectora. Qualquer implementação que viole estes limites deve ser refatorada imediatamente. Este arquivo é o Single Source of Truth (SSOT). Nenhuma mudança complexa pode ser implementada antes de sua regra estar documentada aqui.

---

## 1. FILOSOFIA E RIGOR DE ENGENHARIA

O Vectora rejeita o **"Vibe-Coding"** — programação por intuição, suposição ou conveniência. Cada linha de lógica de negócio é tratada como um compromisso de engenharia industrial.

### 1.1 Pair Programming e Governança

- **Desapego Radical ao Código:** Se o código falha, a falha é na comunicação ou na arquitetura. A correção é feita via diálogo e ajuste da documentação, nunca por remendos manuais.
- **SSOT (Single Source of Truth):** Este arquivo é a Lei de Ferro. Antes de qualquer mudança complexa, a regra deve estar documentada aqui primeiro.
- **Idioma:** Código e documentação técnica em **Inglês**. Diálogo e tom de pair programming em **Português**.

### 1.2 Protocolo TDD (Red-Green-Refactor)

Nenhuma lógica de negócio existe sem um teste que a justifique.

1. **RED:** Escrever o teste que falha, definindo o contrato.
2. **GREEN:** Implementar o código mínimo para passar.
3. **REFACTOR:** Otimizar mantendo o status de aprovação.

### 1.3 O Padrão 300% (Lei de Ferro)

Cada funcionalidade deve ser provada por pelo menos **3 variações** na mesma suite de testes:

1. **Happy Path:** Cenário base ideal.
2. **Negative:** Entrada inválida ou falha esperada.
3. **Edge Case:** Combinações complexas, concorrência, limites de borda.

---

## 2. PILARES ARQUITETURAIS

Estas são as restrições inegociáveis das quais todas as regras derivam.

| Pilar                        | Restrição                                                                                                                                                                        |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Local Primeiro**           | Toda feature central deve funcionar sem internet. Provedores cloud são opt-in, nunca dependências.                                                                               |
| **Baixo Consumo**            | Daemon do systray ≤ 5MB RSS em idle. Sistema completo ≤ 4GB RSS sob carga.                                                                                                       |
| **Go Puro**                  | Sem CGO, sem dependências C++, sem interpretadores em runtime, a menos que não exista alternativa Go viável. Exceções: `llama.cpp` (sidecar de inferência), `Fyne` (instalador). |
| **Sem Estado Compartilhado** | Binários de interface mantêm zero estado de aplicação. Todo estado vive no daemon.                                                                                               |
| **Proteção de Escrita**      | Nenhuma escrita no sistema de arquivos ou execução shell pode ocorrer sem um snapshot prévio do GitBridge.                                                                       |
| **Isolamento de Workspace**  | Nenhum workspace pode ler da coleção vetorial de outro workspace na camada de armazenamento.                                                                                     |

---

## 3. ESTRUTURA DO REPOSITÓRIO (OBRIGATÓRIA)

```markdown
vectora/
├── cmd/                        # Apenas pontos de entrada. Sem lógica de negócio.
│   ├── vectora/                # Daemon systray (orquestrador central)
│   ├── vectora-cli/            # Binário CLI (Bubbletea)
│   ├── vectora-web/            # Binário Web UI (Wails)
│   └── vectora-installer/      # Binário do instalador (Fyne)
│
├── internal/                   # Toda a lógica de negócio. Não importável externamente.
│   ├── core/                   # Pipeline RAG, gerenciamento de workspaces, lógica de sessão
│   ├── db/                     # Wrappers do chromem-go e bbolt
│   ├── llm/                    # Abstração de provedores baseada em langchaingo
│   ├── ipc/                    # Servidor IPC (daemon) e cliente (interfaces)
│   ├── tray/                   # UI do systray e ciclo de vida dos processos
│   ├── tools/                  # Toolkit agêntico (filesystem, busca, shell, web, memória)
│   ├── git/                    # GitBridge: snapshot e rollback
│   ├── mcp/                    # Implementação do servidor MCP
│   ├── acp/                    # Implementação do agente ACP
│   ├── index/                  # Cliente HTTP do Vectora Index
│   └── infra/                  # Transversal: logging, config, tipos de erro
│
├── pkg/                        # Packages públicos. Adicionar apenas após discussão do time.
│   └── vectorakit/             # SDK externo futuro
│
├── web/                        # Frontend Next.js (export estático, embarcado via go:embed)
├── index-server/               # Index Server HTTP do Vectora (Go, net/http)
├── assets/                     # Assets estáticos embarcados (ícones, configs padrão)
├── scripts/                    # Scripts de build, release e setup
├── tests/                      # Suites de testes de integração e end-to-end
├── docs/                       # Documentação para desenvolvedores
├── go.mod
├── go.sum
└── Makefile
```

### 3.1 Regras de Fronteira de Package

- `cmd/` contém apenas: parsing de flags, cabeamento de dependências e a chamada para iniciar o serviço raiz. Sem `if`, sem lógica de negócio.
- Packages em `internal/` não devem se importar circularmente. O grafo de dependências flui: `core → db`, `core → llm`, `tools → git`, `mcp/acp → core`, `ipc → core`.
- `internal/infra/` pode ser importado por qualquer package em `internal/`. Nunca deve importar outros packages `internal/`.
- `pkg/` está congelado até o primeiro release estável.
- `index-server/` é um módulo Go independente. Não compartilha packages `internal/` com o módulo principal.

---

## 4. O DAEMON: SYSTRAY COMO ORQUESTRADOR CENTRAL

O binário `cmd/vectora` é a única fonte de verdade para todo o estado em runtime. Deve ser o primeiro processo a iniciar e o último a sair.

### 4.1 Responsabilidades

- Possui e gerencia todo o estado de workspaces.
- Possui e gerencia todas as conexões ativas com provedores LLM via `internal/llm`.
- Expõe o servidor IPC para todas as conexões de interface.
- Inicializa e encerra processos de interface sob demanda.
- Executa o GitBridge antes de qualquer operação de escrita de tool.

### 4.2 Ciclo de Vida do Processo

```markdown
Login do Sistema
    └── cmd/vectora inicia
            └── Servidor IPC vincula ao socket
            └── Workspaces hidratados do bbolt
            └── Provedor LLM inicializado via langchaingo
            └── Sidecar llama.cpp iniciado (se modo Qwen)
            └── Ícone do systray renderizado
                    └── Usuário aciona CLI / Web UI
                            └── Daemon inicializa processo de interface
                            └── Interface conecta via IPC
                            └── Interface sai → Daemon permanece
```

### 4.3 Regras de Inicialização de Interface

- Interfaces são iniciadas apenas quando explicitamente solicitadas pelo usuário via menu do systray.
- Apenas uma instância de cada binário de interface pode rodar por vez. Tentar iniciar uma segunda instância deve focar a existente.
- Se um binário de interface crasha, o daemon registra o evento e limpa o handle do processo. Nenhum crash em uma interface pode afetar a estabilidade do daemon.

---

## 5. WEB UI: ARQUITETURA WAILS + NEXT.JS

O Web UI (`cmd/vectora-web`) é uma aplicação Wails que embarca o frontend Next.js como export estático.

### 5.1 Modelo de Build

- O Next.js deve ser configurado com `output: 'export'`. Isso produz um site completamente estático sem necessidade de servidor Node.js.
- O output estático é embarcado no binário Wails via `//go:embed`.
- Nenhum runtime Node.js roda em nenhum momento durante a operação normal da aplicação.
- O frontend deve ser totalmente funcional sem dependências de CDN externo. Todos os assets são auto-contidos.

### 5.2 Comunicação Frontend ↔ Go

O frontend se comunica com o backend Go exclusivamente através de **Wails bindings**. Não há servidor HTTP, não há fetch para localhost, não há WebSocket.

```go
// Go — métodos da struct App são expostos ao frontend via wails.Bind
type App struct {
    ipcClient *ipc.Client
}

func (app *App) QueryWorkspace(workspaceID string, query string) (QueryResponse, error) {
    return app.ipcClient.Send("workspace.query", map[string]any{
        "workspace_id": workspaceID,
        "query":        query,
    })
}
```

```typescript
// TypeScript — Wails gera bindings tipados automaticamente a partir dos métodos Go
import { QueryWorkspace } from "../wailsjs/go/main/App";

const response = await QueryWorkspace("godot-4.2", "como usar sinais?");
```

O Wails gera os bindings TypeScript automaticamente a partir dos métodos Go exportados. A camada de binding é type-safe e não requer manutenção manual.

### 5.3 Escopo do Web UI

O Web UI é a interface principal voltada ao usuário. Seu escopo inclui: a experiência principal de chat, criação e gerenciamento de workspaces, configuração de provedor (entrada de API key, seleção de modelo), e navegação e download de datasets do Vectora Index.

### 5.4 Regras do Web UI

- **RN-WEB-01:** A struct App em `cmd/vectora-web` não deve conter lógica de negócio. É uma camada fina de binding que delega todas as chamadas ao cliente IPC.
- **RN-WEB-02:** Nenhuma chamada direta a banco de dados ou LLM pode originar da struct App do Wails. Todos os dados fluem via IPC para o daemon.
- **RN-WEB-03:** O build Next.js deve passar `next build` com `output: 'export'` sem erros antes de qualquer PR tocando `web/` ser mergeado.
- **RN-WEB-04:** Métodos de binding do Wails devem seguir as mesmas convenções de nomenclatura definidas na Seção 12.

---

## 6. MOTOR DE IA: LANGCHAINGO + LLAMA.CPP

Todas as capacidades de IA — completação, embedding e tool calling — são mediadas por `internal/llm`, construído sobre `langchaingo`.

### 6.1 langchaingo como Abstração de Provedor

`langchaingo` fornece a interface unificada para todos os provedores LLM e de embedding. É o equivalente Go do LangChain/LlamaIndex, abstraindo SDKs específicos de provedores por trás de interfaces comuns.

- **Gemini** é integrado via o provedor Google AI do langchaingo, usando a API Key do usuário.
- **Provedores futuros** (Claude, OpenAI, Ollama, etc.) são adicionados implementando ou configurando o provedor langchaingo correspondente. Nenhuma mudança em `internal/core` é necessária.
- `langchaingo` vive inteiramente dentro de `internal/llm`. Nenhum outro package o importa diretamente.

### 6.2 llama.cpp como Sidecar de Inferência Local

`llama.cpp` gerencia a execução offline de modelos para o Qwen. Roda como processo separado (sidecar) e expõe um servidor HTTP local com o qual `internal/llm` se comunica.

```
cmd/vectora (daemon)
    └── internal/llm
            └── processo llama.cpp (sidecar)
                    └── servidor HTTP em localhost (apenas loopback)
                    └── modelo Qwen GGUF carregado
```

- O sidecar é iniciado pelo daemon na inicialização quando o modo Qwen está ativo.
- O sidecar vincula apenas ao loopback (`127.0.0.1`). Nunca deve ser exposto em uma interface de rede.
- Se o sidecar crasha, o daemon tenta uma reinicialização automática antes de apresentar um erro ao usuário.
- A porta do sidecar é alocada dinamicamente e armazenada no estado do daemon. Nunca é hardcoded.

### 6.3 Interface do Provedor

Apesar de usar langchaingo internamente, `internal/llm` ainda expõe sua própria interface para o resto do sistema. Isso isola `internal/core` de mudanças na API do langchaingo.

```go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
    Embed(ctx context.Context, input string) ([]float32, error)
    Name() string
    IsConfigured() bool
}

type CompletionRequest struct {
    Messages     []Message
    SystemPrompt string
    MaxTokens    int
    Temperature  float32
    Tools        []ToolDefinition
}

type CompletionResponse struct {
    Content   string
    ToolCalls []ToolCall
    Usage     TokenUsage
}
```

### 6.4 Gerenciamento da Gemini API Key

- A Gemini API Key é armazenada em `~/.vectora/config.json`, criptografada em repouso usando o keychain do SO onde disponível (macOS Keychain, Windows Credential Manager, Linux Secret Service).
- A chave é carregada em memória uma vez na inicialização do provedor e nunca escrita em logs, payloads IPC ou mensagens de erro.
- A chave é passada ao provedor Google AI do langchaingo na inicialização e não é armazenada em nenhum outro package `internal/`.
- Rotação: o usuário pode atualizar a chave via painel de configurações do Web UI. O provedor é re-inicializado imediatamente com a nova chave sem necessidade de reiniciar o daemon.

### 6.5 Regras do Motor de IA

- **RN-LLM-01:** Nenhum package fora de `internal/llm` pode importar `langchaingo` ou qualquer SDK de provedor diretamente.
- **RN-LLM-02:** `Complete` deve ser cancelável por contexto. Um contexto cancelado deve abortar a requisição e retornar `ctx.Err()`.
- **RN-LLM-03:** `Embed` deve retornar um vetor determinístico para o mesmo input dado o mesmo modelo.
- **RN-LLM-04:** A Gemini API Key nunca deve aparecer em logs, mensagens de erro, payloads IPC ou relatórios de crash.
- **RN-LLM-05:** A porta do sidecar llama.cpp deve ser alocada dinamicamente. A porta 8080 ou qualquer outra porta fixa não deve ser hardcoded.
- **RN-LLM-06:** Adicionar um novo provedor requer apenas uma nova implementação de `Provider` em `internal/llm`. Nenhuma mudança em `internal/core`, `internal/ipc` ou qualquer outro package é permitida como parte de uma adição de provedor.

---

## 7. CONTRATO IPC

A camada IPC (`internal/ipc`) é a espinha dorsal de comunicação entre o daemon e todas as interfaces.

### 7.1 Transporte

O IPC usa **Unix Domain Sockets** no Linux/macOS e **Named Pipes** no Windows:

```
~/.vectora/run/vectora.sock   (Linux/macOS)
\\.\pipe\vectora               (Windows)
```

### 7.2 Formato de Mensagem

Todas as mensagens são **JSON delimitado por newline** (`\n` como delimitador de frame):

```json
{
  "id": "string (UUIDv4, único por requisição)",
  "type": "string (request | response | event)",
  "method": "string (apenas para type=request)",
  "payload": "object (específico do método)",
  "error": "object | null (apenas para type=response)"
}
```

### 7.3 Métodos IPC (Request/Response)

#### Workspace

| Método                 | Payload                                        | Resposta                      |
| ---------------------- | ---------------------------------------------- | ----------------------------- |
| `workspace.list`       | `{}`                                           | `{ workspaces: Workspace[] }` |
| `workspace.create`     | `{ name, source_path }`                        | `{ workspace_id }`            |
| `workspace.delete`     | `{ workspace_id }`                             | `{}`                          |
| `workspace.activate`   | `{ workspace_id }`                             | `{}`                          |
| `workspace.deactivate` | `{ workspace_id }`                             | `{}`                          |
| `workspace.query`      | `{ workspace_id, query, active_workspaces[] }` | `{ answer, sources[] }`       |
| `workspace.index`      | `{ workspace_id }`                             | `{ job_id }` (async)          |

#### Provedor

| Método         | Payload                  | Resposta                                             |
| -------------- | ------------------------ | ---------------------------------------------------- |
| `provider.get` | `{}`                     | `{ provider: "qwen" \| "gemini", configured: bool }` |
| `provider.set` | `{ provider, api_key? }` | `{}`                                                 |

#### Tools

| Método         | Payload                       | Resposta                  |
| -------------- | ----------------------------- | ------------------------- |
| `tool.execute` | `{ tool_name, args: object }` | `{ result, snapshot_id }` |
| `tool.undo`    | `{ snapshot_id }`             | `{ restored: bool }`      |

#### Index

| Método           | Payload                | Resposta                  |
| ---------------- | ---------------------- | ------------------------- |
| `index.browse`   | `{ query?, filters? }` | `{ datasets: Dataset[] }` |
| `index.download` | `{ dataset_id }`       | `{ job_id }` (async)      |
| `index.publish`  | `{ path, metadata }`   | `{ submission_id }`       |

#### Sessão

| Método            | Payload      | Resposta                  |
| ----------------- | ------------ | ------------------------- |
| `session.history` | `{ limit? }` | `{ messages: Message[] }` |
| `session.clear`   | `{}`         | `{}`                      |

### 7.4 Eventos IPC (Daemon → Interface, Proativos)

| Método do Evento          | Payload                         | Descrição                    |
| ------------------------- | ------------------------------- | ---------------------------- |
| `workspace.indexed`       | `{ workspace_id, chunk_count }` | Job de indexação concluído   |
| `workspace.index_failed`  | `{ workspace_id, error }`       | Job de indexação falhou      |
| `index.download_progress` | `{ job_id, percent }`           | Atualização de progresso     |
| `index.download_complete` | `{ job_id, workspace_id }`      | Download finalizado          |
| `tool.snapshot_created`   | `{ snapshot_id, tool_name }`    | Snapshot criado              |
| `daemon.status`           | `{ ram_mb, workspaces_loaded }` | Broadcast periódico de saúde |

### 7.5 Objeto de Erro

```json
{
  "code": "string (legível por máquina, snake_case)",
  "message": "string (legível por humano)",
  "detail": "object | null"
}
```

**Códigos Canônicos:**

| Código                     | Significado                                 |
| -------------------------- | ------------------------------------------- |
| `workspace_not_found`      | workspace_id não existe                     |
| `workspace_already_active` | Workspace já está no conjunto ativo         |
| `provider_not_configured`  | Nenhum provedor LLM configurado             |
| `tool_not_found`           | tool_name não existe                        |
| `snapshot_failed`          | GitBridge não conseguiu criar snapshot      |
| `index_signature_invalid`  | Dataset falhou na verificação de assinatura |
| `ipc_method_unknown`       | Método não registrado                       |
| `ipc_payload_invalid`      | Payload falhou na validação de schema       |
| `internal_error`           | Erro não tratado do daemon                  |

### 7.6 Regras IPC

- **RN-IPC-01:** Toda requisição deve receber exatamente uma resposta com o `id` correspondente.
- **RN-IPC-02:** Eventos são broadcast a todos os clientes conectados. Clientes não devem assumir ordenação.
- **RN-IPC-03:** O servidor IPC deve tratar desconexão de clientes graciosamente.
- **RN-IPC-04:** Limite de tamanho de mensagem: 4MB por frame. Payloads maiores usam streaming.
- **RN-IPC-05:** O daemon valida todo payload recebido antes de processar.

---

## 8. CONTRATOS DE API INTERNA

### 8.1 `internal/db` — Camada de Armazenamento

```go
type VectorStore interface {
    UpsertChunk(ctx context.Context, collection string, chunk Chunk) error
    Query(ctx context.Context, collection string, query string, topK int) ([]ScoredChunk, error)
    DeleteCollection(ctx context.Context, collection string) error
    CollectionExists(ctx context.Context, collection string) bool
}

type KVStore interface {
    Set(ctx context.Context, bucket string, key string, value []byte) error
    Get(ctx context.Context, bucket string, key string) ([]byte, error)
    Delete(ctx context.Context, bucket string, key string) error
    List(ctx context.Context, bucket string, prefix string) ([]string, error)
}
```

Cada workspace mapeia para exatamente uma coleção chromem-go e um bucket bbolt, ambos com nome `ws:<workspace_id>`. Deleção deve ser atômica entre os dois.

### 8.2 `internal/core` — Pipeline RAG

```go
type RAGPipeline interface {
    Query(ctx context.Context, req QueryRequest) (QueryResponse, error)
    IndexWorkspace(ctx context.Context, workspaceID string) error
    ActiveWorkspaces() []string
    ActivateWorkspace(workspaceID string) error
    DeactivateWorkspace(workspaceID string) error
}
```

`Query` recupera em paralelo de todos os workspaces ativos, mescla e re-ranqueia antes da chamada ao LLM. `IndexWorkspace` é assíncrono e reporta progresso via channel.

### 8.3 `internal/tools` — Toolkit Agêntico

```go
type Tool interface {
    Name() string
    Description() string
    Schema() json.RawMessage
    Execute(ctx context.Context, args map[string]any) (ToolResult, error)
}
```

**Tools Registradas:**

| Nome                | Categoria  | Muta Estado |
| ------------------- | ---------- | ----------- |
| `read_file`         | Filesystem | Não         |
| `write_file`        | Filesystem | **Sim**     |
| `read_folder`       | Filesystem | Não         |
| `edit`              | Filesystem | **Sim**     |
| `find_files`        | Busca      | Não         |
| `grep_search`       | Busca      | Não         |
| `google_search`     | Web        | Não         |
| `web_fetch`         | Web        | Não         |
| `run_shell_command` | Sistema    | **Sim**     |
| `save_memory`       | Memória    | **Sim**     |
| `enter_plan_mode`   | Memória    | Não         |

Tools que mutam estado devem chamar `git.Bridge.Snapshot()` como primeira operação. Falha no snapshot cancela a execução.

### 8.4 `internal/git` — GitBridge

```go
type Bridge interface {
    Snapshot(ctx context.Context, label string) (snapshotID string, err error)
    Restore(ctx context.
```
