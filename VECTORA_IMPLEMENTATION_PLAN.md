# VECTORA: PLANO DE IMPLEMENTAÇÃO CONSOLIDADO (v2.0)

**Status:** Engineering Specification — Unified Blueprint
**Versão:** 2.0 (Consolidated from 7 distributed plans)
**Data:** 2026-04-05
**Single Source of Truth:** Este documento substitui todos os planos descentralizados anteriores

---

## CONTEXTO E MOTIVAÇÃO

O Vectora estava fragmentado em **7 planos de implementação isolados** (`agent_`, `api_`, `app_`, `cli_`, `llama_`, `index_`, `test_`), causando:

1. **Inconsistência Arquitetural:** Decisões de design eram locais e não globais
2. **Duplicação de Conceitos:** Regras de negócio repetidas ou conflitantes
3. **Overhead Cognitivo:** Impossível ter visão holística sem 7 leituras paralelas
4. **Manutenção Custosa:** Mudanças em um plano não refletiam nos outros

**Solução:** Consolidar em um **único documento denso (880+ linhas)** que detalha cada subsistema em ordem de implementação e interdependência, mantendo rigor arquitetural absoluto.

---

## 1. TOPOLOGIA E ARQUITETURA GLOBAL

### 1.1 Visão de Sistema Integrado

O Vectora é um assistente IA desktop completo com três interfaces primárias, um daemon central e subsistemas especializados:

```
┌─────────────────────────────────────────────────────────────┐
│                    APLICAÇÃO VECTORA                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  Web UI      │  │  CLI (TUI)   │  │  MCP Servidor    │  │
│  │ (Wails+Next) │  │ (Bubbletea)  │  │ (Cursor/VSCode)  │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘  │
│         │                 │                    │             │
│         └─────────────────┼────────────────────┘             │
│                           │ IPC (JSON-ND)                    │
│                    ┌──────▼──────┐                           │
│                    │   DAEMON    │                           │
│                    │  (Go Core)  │                           │
│                    │  ~100MB RAM │                           │
│                    └──────┬──────┘                           │
│                           │                                  │
│       ┌───────────────────┼───────────────────┐              │
│       │                   │                   │              │
│   ┌───▼────┐         ┌───▼──────┐      ┌────▼────┐         │
│   │   IPC  │         │   Core   │      │  Tools  │         │
│   │ Server │         │   RAG    │      │ & ACP   │         │
│   └────────┘         │ Pipeline │      └────┬────┘         │
│                      └───┬──────┘            │               │
│                          │                  │               │
│       ┌──────────────────┼──────────────────┘               │
│       │                  │                                  │
│   ┌───▼────────┐  ┌─────▼────────┐                         │
│   │   Storage  │  │     LLM      │                         │
│   │ (bbolt +   │  │  Providers   │                         │
│   │  chromem)  │  │ (Qwen + Gemi)│                         │
│   └────────────┘  └──────┬───────┘                         │
│                          │                                 │
│                 ┌────────▼────────┐                        │
│                 │  Sidecar Llama  │                        │
│                 │  (Processo Filho)                        │
│                 └─────────────────┘                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Binários Finais

| Binário | Localização | Rol | Tamanho | RAM |
|---------|-------------|-----|--------|-----|
| `vectora` | `cmd/vectora/` | Daemon central, orquestrador IPC, sistema de bandeja | ~5MB | ~100MB |
| `vectora-app` | `cmd/vectora-app/` | Web UI embarcado (Wails + Next.js SSG) | ~40MB | ~150MB (com renderer) |
| `vectora-cli` | `cmd/vectora/` (subcomando) | Interface Terminal baseada em Bubbletea | ~8MB | <10MB |
| `vectora-installer` | `cmd/vectora-installer/` | Setup wizard (Fyne GUI) com downloader de llama.cpp | ~15MB | ~50MB |
| `llama-sidecar` | `cmd/llama/` | Processo filho gerenciado, wrapper STDIO para llama.cpp | ~2MB | Dinâmico |

### 1.3 Padrão de Armazenamento e Diretórios

```
%USERPROFILE%/.Vectora/
├── .env                           # API Keys, flags globais
├── data/
│   ├── vectora.db                 # bbolt (workspaces, sessions, settings)
│   └── chroma_ws_{id}/            # Chromem-Go vector store por workspace
├── engines/
│   ├── catalog.json               # Embarcado, lista de llama.cpp builds
│   ├── llama-cpp-{version}/       # Binários do llama.cpp instalados
│   └── qwen-{model}.gguf          # Modelos GGUF cachados
├── backups/
│   └── {uuid}-{filename}.bak      # GitBridge snapshots de alterações
├── logs/
│   ├── daemon.log                 # Rotação a cada 10MB, 5 arquivos máx
│   ├── wails.log
│   └── ipc.log
├── run/
│   └── vectora.sock               # UDS (Unix) ou \\.\pipe\vectora (Windows)
└── temp/
    └── {workspace_id}/            # Downloads/staging temporários
```

---

## 2. CAMADA DE INFRAESTRUTURA (`internal/infra`)

### 2.1 ConfigLoader

**Arquivo:** `internal/infra/config.go`

**Responsabilidade:** Carregar configurações do arquivo `.env` e prover defaults sensatos.

```go
type Config struct {
    // API Keys
    GeminiAPIKey          string

    // Limites de Sistema
    MaxRAMDaemon          int64  // Default: 4GB
    MaxRAMIndexing        int64  // Default: 512MB

    // Engines e Providers
    PreferredLLMProvider  string // "gemini" | "qwen_local"

    // Networking
    IPCPipeWindows        string // Default: "\\.\pipe\vectora"
    IPCSocketUnix         string // Default: "$HOME/.Vectora/run/vectora.sock"

    // Logging
    LogLevel              string // "DEBUG" | "INFO" | "WARN" | "ERROR"
    LogFormat             string // "json" | "text"
}

func LoadConfig(envPath string) (*Config, error)
func (c *Config) Validate() error
func (c *Config) ApplyDefaults()
```

**Testes:**
- Validação de arquivo `.env` malformado
- Fallback para defaults se arquivo não existe
- Suporte a variáveis de ambiente override

### 2.2 Logger Estruturado (slog + JSON)

**Arquivo:** `internal/infra/logger.go`

**Responsabilidade:** Logging centralizado em formato JSON, amigável a parsers e inspeções.

```go
type LoggerConfig struct {
    Level   slog.Level
    Handler slog.Handler // JSONHandler para produção
    Path    string       // Arquivo de saída
}

var globalLogger *slog.Logger

func InitLogger(cfg *LoggerConfig) error
func Debug(ctx context.Context, msg string, args ...any)
func Info(ctx context.Context, msg string, args ...any)
func Warn(ctx context.Context, msg string, args ...any)
func Error(ctx context.Context, msg string, args ...any)
```

**Atributos JSON obrigatórios:**
- `timestamp`: RFC3339 com milissegundos
- `level`: "DEBUG", "INFO", "WARN", "ERROR"
- `module`: Nome do pacote (ex: "internal.ipc")
- `function`: Nome da função
- `line`: Número da linha
- `message`: Mensagem legível
- `trace_id`: UUID para rastreamento distribuído (quando aplicável)

**Rotação de Logs:**
- Máximo 10MB por arquivo
- Guardar últimos 5 arquivos
- Timestamp automático no nome (daemon.2026-04-05T14-30.log)

---

## 3. CAMADA IPC: COMUNICAÇÃO INTER-PROCESSO (`internal/ipc`)

### 3.1 Servidor IPC (`internal/ipc/server.go`)

**Responsabilidade:** Ponte central entre as interfaces (Web, CLI, MCP) e o daemon.

**Transporte:**
- **Windows:** Named Pipes (`\\.\pipe\vectora`)
- **Unix:** Domain Sockets (`~/.Vectora/run/vectora.sock`)
- **Protocolo:** JSON-ND (Newline Delimited JSON) sem cabeçalhos HTTP

**Ciclo de Vida do Socket:**

```
1. Boot do Daemon
   → Criar socket (ou named pipe)
   → Registrar rotas (workspace.query, tool.execute, etc.)
   → Escutar em loop de aceite

2. Cliente Conecta
   → Daemon spawna goroutine para handleConnection()
   → Ler mensagens JSON-ND do cliente
   → Despachar para rota registrada
   → Escrever resposta ao cliente

3. Erro de Conexão
   → Cleanup automático
   → Log estruturado do erro
   → Retry de cliente gerenciado pelo próprio cliente
```

### 3.2 Contrato de Mensagens (IPCMessage)

```json
{
  "id": "uuid-v4",
  "type": "request | response | event",
  "method": "workspace.query | workspace.index | tool.execute | ...",
  "payload": {},
  "error": null
}
```

**Tipos de Mensagem:**

| Tipo | Fluxo | Exemplo |
|------|-------|---------|
| `request` | Cliente → Daemon | `{"method": "workspace.query", "payload": {...}}` |
| `response` | Daemon → Cliente | `{"id": "...", "type": "response", "payload": {...}, "error": null}` |
| `event` | Daemon → Cliente (push) | `{"method": "index.progress", "payload": {...}}` |

### 3.3 Métodos e Contratos

| Método | Payload In | Resposta | Descrição |
|--------|-----------|----------|-----------|
| `workspace.query` | `{"ws_id": "...", "query": "..."}` | `{"answer": "...", "sources": [...], "thinking": "..."}` | Consulta RAG |
| `workspace.index` | `{"ws_id": "...", "path": "..."}` | `{"job_id": "...", "status": "running"}` | Iniciar indexação |
| `workspace.create` | `{"name": "...", "description": "..."}` | `{"ws_id": "...", "indexed": false}` | Criar workspace |
| `tool.execute` | `{"tool": "read_file", "args": {...}}` | `{"result": "...", "snapshot_id": "..."}` | Executar ferramenta |
| `provider.set` | `{"provider": "gemini", "key": "..."}` | `{"configured": true}` | Configurar LLM |
| `history.list` | `{"ws_id": "...", "limit": 20}` | `{"messages": [...]}` | Listar histórico |

### 3.4 Sistema de Eventos (Push Notifications)

O Daemon pode **empurrar eventos** para clientes sem que eles tenham feito uma requisição:

```json
{
  "id": "event-uuid",
  "type": "event",
  "method": "index.progress",
  "payload": {
    "ws_id": "godot",
    "files_done": 12,
    "files_total": 45,
    "percent": 26
  }
}
```

**Eventos principais:**

- `index.progress`: Progresso de indexação de um workspace
- `index.completed`: Indexação finalizada
- `shell_stream_chunk`: Chunk de saída de ferramenta shell (streaming)
- `tool_completed`: Ferramenta finalizou execução
- `llm_thinking`: (Futuro) Stream de raciocínio do agente

### 3.5 Tratamento de Erros

```json
{
  "id": "...",
  "type": "response",
  "error": {
    "code": "workspace_not_found",
    "message": "Workspace 'unknown' não existe",
    "details": {...}
  }
}
```

**Códigos de erro padrão:**
- `invalid_request`: Payload inválido
- `method_not_found`: Método desconhecido
- `workspace_not_found`: Workspace não existe
- `index_in_progress`: Workspace já está indexando
- `llm_error`: Erro ao chamar LLM
- `tool_execution_error`: Ferramenta falhou
- `authentication_required`: API Key não configurada

### 3.6 Segurança e Isolamento

- **Timeout de Conexão:** Conexões inativas por 30 minutos são fechadas
- **Limite de Requisições:** Máximo 1000 req/min por cliente
- **Validação de Payload:** JSON schema validation antes de despacho
- **Logging Obrigatório:** Todas as chamadas IPC são logadas para auditoria (GitBridge)
- **Criptografia Opcional:** Handshake de chave simétrica se API keys são transmitidas

---

## 4. BANCO DE DADOS (`internal/db`)

### 4.1 Armazenamento Chave-Valor (`internal/db/store.go`)

**Tecnologia:** bbolt (BoltDB) — imutável, ACID, embedded.

**Buckets e Schemas:**

```
buckets:
  ├── workspaces        # ws_id → WorkspaceMetadata JSON
  ├── sessions          # session_uuid → [Message, Message, ...] (histórico de chat)
  ├── settings          # global_key → JSON value (preferências do usuário)
  └── snapshots         # snapshot_uuid → {before, after, timestamp}
```

**Estruturas de Dados:**

```go
type WorkspaceMetadata struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    CreatedAt    time.Time `json:"created_at"`
    IndexedAt    *time.Time `json:"indexed_at,omitempty"`
    ChunkCount   int       `json:"chunk_count"`
    IndexStatus  string    `json:"status"` // "idle", "indexing", "done", "error"
    SourcePath   string    `json:"source_path"` // Caminho originário
}

type Message struct {
    ID        string                 `json:"id"`
    Timestamp time.Time              `json:"timestamp"`
    Role      string                 `json:"role"` // "user", "assistant", "system"
    Content   string                 `json:"content"`
    Metadata  map[string]interface{} `json:"metadata"` // tool_calls, thinking, etc
}

type SnapshotMetadata struct {
    ID         string                 `json:"id"`
    FilePath   string                 `json:"file_path"`
    Timestamp  time.Time              `json:"timestamp"`
    Operation  string                 `json:"operation"` // "write", "edit", "delete"
    BeforeHash string                 `json:"before_hash"`
    AfterHash  string                 `json:"after_hash"`
}
```

**Interface Base:**

```go
type Store interface {
    // Workspaces
    CreateWorkspace(ctx context.Context, ws *WorkspaceMetadata) error
    GetWorkspace(ctx context.Context, wsID string) (*WorkspaceMetadata, error)
    ListWorkspaces(ctx context.Context) ([]*WorkspaceMetadata, error)
    UpdateWorkspace(ctx context.Context, ws *WorkspaceMetadata) error
    DeleteWorkspace(ctx context.Context, wsID string) error

    // Sessions
    AppendMessage(ctx context.Context, sessionID string, msg *Message) error
    GetMessageHistory(ctx context.Context, sessionID string, limit int) ([]*Message, error)

    // Snapshots
    SaveSnapshot(ctx context.Context, snap *SnapshotMetadata) error
    GetSnapshot(ctx context.Context, snapID string) (*SnapshotMetadata, error)
    ListSnapshotsByFile(ctx context.Context, filePath string, limit int) ([]*SnapshotMetadata, error)
}
```

### 4.2 Armazenamento Vetorial (`internal/db/vector.go`)

**Tecnologia:** chromem-go — vector database em memory com persistência.

**Isolamento por Workspace:**

Cada workspace possui sua própria collection isolada:

```go
type VectorStore interface {
    // Chunking e Ingestão
    IndexDocuments(ctx context.Context, wsID string, chunks []*Chunk) error

    // Busca Semântica
    SearchSimilar(ctx context.Context, wsID string, query string, topK int) ([]*SearchResult, error)

    // Metadata
    GetCollectionStats(ctx context.Context, wsID string) (*CollectionStats, error)

    // Lifecycle
    CreateCollection(ctx context.Context, wsID string) error
    DeleteCollection(ctx context.Context, wsID string) error
}

type Chunk struct {
    ID           string  `json:"id"`
    Content      string  `json:"content"`
    Embedding    []float32 `json:"embedding"` // 384D ou 1536D conforme modelo
    Metadata     map[string]string `json:"metadata"`
    SourceFile   string  `json:"source_file"`
    ChunkIndex   int     `json:"chunk_index"`
}

type SearchResult struct {
    ChunkID  string
    Content  string
    Score    float32
    Metadata map[string]string
}
```

**Estratégia de Chunking:**

Todos os documentos passam por semantic chunking:
- **Max chunk:** 512 tokens
- **Overlap:** 50 tokens
- **Estratégia:** Baseada em marcadores semânticos (títulos, funções, blocos de código)

### 4.3 Gerenciamento de Memória

**Limite de Raio de Ação:**
- Máximo 4GB total do sistema
- Máximo 512MB durante indexação
- Máximo 100MB para o daemon idle

**Monitora via:**
- `runtime.MemStats` no Go
- Goroutines de limpeza que trigam GC agressivo se threshold é atingido
- Alert ao usuário se próximo do limite

---

## 5. PIPELINE RAG E ORQUESTRAÇÃO (`internal/core`)

### 5.1 RAG Pipeline (`internal/core/rag_pipeline.go`)

**Responsabilidade:** Orquestrar o fluxo: Embedding → Busca → Contexto → LLM → Resposta.

**Fluxo Padrão:**

```
User Query
    ↓
[1] Embed Query (LLM Embedding Provider)
    ↓ Vetor 384D
[2] Search Similar (Chromem-Go KNN, TopK=5)
    ↓ [Chunk1, Chunk2, ...]
[3] Construct Context (Flatten chunks em single string)
    ↓ "Context: ...chunk1...\n...chunk2..."
[4] Get Tool Specs (ACP Registry)
    ↓ [Tool1 JSON Schema, Tool2 JSON Schema, ...]
[5] Build Final Prompt (System + Context + Query + Tools)
    ↓ Mensagem completa
[6] Call LLM Provider (Gemini ou Qwen)
    ↓ Tool calls ou resposta texto
[7] If Tool Call:
        → Execute tool (Filesystem, Web, Shell)
        → Add result ao histórico (ReAct loop v1: single-shot)
[8] Return Answer to User
```

**Constraints:**

- **TOP K = 5:** Fixo para proteger RAM
- **Latência Alvo:** < 2s até início de resposta
- **Privacy:** Zero chamadas externas durante busca/contexto

**Interface Principal:**

```go
type RAGEngine interface {
    // Query única (modo sync)
    Query(ctx context.Context, wsID, userQuery string) (*QueryResult, error)

    // Streaming (para UI)
    QueryStream(ctx context.Context, wsID, userQuery string,
                onChunk func(chunk string) error) (*QueryResult, error)
}

type QueryResult struct {
    ID            string
    Answer        string
    Sources       []*SourceReference // Chunks usados
    Thinking      string // Raciocínio do agente (se disponível)
    ToolCalls     []*ToolCall
    ExecutionTime time.Duration
    Model         string // "gemini" ou "qwen_local"
}

type SourceReference struct {
    ChunkID  string
    Content  string
    Score    float32
    SourceFile string
    LineNumber int
}
```

### 5.2 Workspace Manager (`internal/core/workspace.go`)

**Responsabilidade:** CRUD de workspaces e coordenação de indexação.

```go
type WorkspaceManager interface {
    Create(ctx context.Context, name, description string) (*Workspace, error)
    Get(ctx context.Context, wsID string) (*Workspace, error)
    List(ctx context.Context) ([]*Workspace, error)
    Delete(ctx context.Context, wsID string) error

    // Indexing
    IndexDirectory(ctx context.Context, wsID, dirPath string) (*IndexJob, error)
    CancelIndexing(ctx context.Context, wsID string) error
    GetIndexStatus(ctx context.Context, wsID string) (*IndexStatus, error)
}

type Workspace struct {
    ID          string
    Name        string
    Description string
    CreatedAt   time.Time
    IndexedAt   *time.Time
    Status      WorkspaceStatus // idle, indexing, done, error
    IndexJob    *IndexJob `json:"-"`
}
```

---

## 6. TOOLKIT AGÊNTICO E CONTROLE AUTÔNOMO (`internal/acp` e `internal/tools`)

### 6.1 ACP: Autonomous Control Protocol (`internal/acp/registry.go`)

**Responsabilidade:** Registrar e gerenciar o "arsenal" de tools disponíveis para o LLM.

```go
type ACPRegistry interface {
    // Tool Registration
    Register(tool Tool) error
    Unregister(toolName string) error
    List() []*ToolSpec
    GetSpec(toolName string) (*ToolSpec, error)

    // Tool Execution
    Execute(ctx context.Context, toolName string, args json.RawMessage) (*ExecutionResult, error)
}

type ToolSpec struct {
    Name          string                 `json:"name"`
    Description   string                 `json:"description"`
    InputSchema   json.RawMessage        `json:"input_schema"` // JSON Schema
    OutputSchema  json.RawMessage        `json:"output_schema"`
    Category      string                 `json:"category"` // filesystem, web, shell, etc
}

type ExecutionResult struct {
    ToolName   string
    Result     json.RawMessage
    SnapshotID string `json:",omitempty"` // Para undo via GitBridge
    Error      *ToolError `json:",omitempty"`
}
```

### 6.2 Toolkit (`internal/tools`)

**Categorias de Tools:**

#### 6.2.1 Filesystem Tools

```go
// read_file: Ler arquivo
{
    "name": "read_file",
    "description": "Lê conteúdo de um arquivo",
    "input": {"path": "string"}
}

// write_file: Escrever arquivo (cria se não existe)
{
    "name": "write_file",
    "description": "Escreve conteúdo para arquivo",
    "input": {"path": "string", "content": "string"}
}

// edit_file: Editar seção específica (via busca + replace)
{
    "name": "edit_file",
    "description": "Edita uma seção de arquivo",
    "input": {"path": "string", "old_text": "string", "new_text": "string"}
}

// find_files: Buscar arquivos por padrão
{
    "name": "find_files",
    "description": "Encontra arquivos por glob ou regex",
    "input": {"pattern": "string", "dir": "string"}
}
```

#### 6.2.2 Information Retrieval Tools

```go
// grep_search: Busca full-text recursiva
{
    "name": "grep_search",
    "description": "Busca texto recursivo em diretório",
    "input": {"pattern": "string", "dir": "string"}
}

// web_search: Busca web via DuckDuckGo (sem key)
{
    "name": "web_search",
    "description": "Busca na web",
    "input": {"query": "string"}
}

// web_fetch: Lê HTML completo de URL
{
    "name": "web_fetch",
    "description": "Extrai texto de página web",
    "input": {"url": "string"}
}
```

#### 6.2.3 Shell & Execution

```go
// shell_command: Executar comando no terminal
{
    "name": "shell_command",
    "description": "Executa comando shell",
    "input": {"command": "string", "cwd": "string", "timeout_sec": "number"},
    "output": {"stdout": "string", "stderr": "string", "exit_code": "number"}
}
```

### 6.3 GitBridge: Snapshot & Undo (`internal/git/bridge.go`)

**Responsabilidade:** Capture antes de cada mutação de arquivo para permitir undo.

```go
type GitBridge interface {
    // Snapshot antes de qualquer mutação
    Snapshot(ctx context.Context, filePath string) (*SnapshotID, error)

    // Restaurar arquivo do snapshot
    Restore(ctx context.Context, snapID SnapshotID) error

    // Listar snapshots de um arquivo
    ListSnapshots(ctx context.Context, filePath string) ([]*SnapshotMetadata, error)
}
```

**Fluxo de Execução de Tool:**

```
1. LLM decide: execute write_file("/path/to/file.txt", "new content")
2. ACP Registry intercepta
3. GitBridge.Snapshot("/path/to/file.txt") → sha256 do conteúdo anterior
4. write_file executa
5. Retorna SnapshotID para UI
6. Se usuário clica "Undo":
   → GitBridge.Restore(snapID)
   → Arquivo restaurado
```

### 6.4 Segurança de Execução de Tools

**Regras de Negócio:**

- **RN-ACP-01:** Toda execução de `shell_command` sem sanção visual prévia é bloqueada
- **RN-ACP-02:** Comando com `rm -rf`, `sudo`, ou `format` requer dupla confirmação
- **RN-ACP-03:** Tools de escrita acionam GitBridge obrigatoriamente
- **RN-ACP-04:** Histórico de pensamentos (Thought stream) é persistido no bbolt para auditoria

---

## 7. MOTOR DE IA: LLM PROVIDERS (`internal/llm`)

### 7.1 Interface Base (`internal/llm/provider.go`)

```go
type Provider interface {
    // Completude com suporte a tool calling
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

    // Embedding (vetorização de texto)
    Embed(ctx context.Context, texts []string) ([][]float32, error)

    // Info do modelo
    ModelInfo() *ModelInfo

    // Configuração
    IsConfigured() bool
    Configure(apiKey string) error
}

type CompletionRequest struct {
    SystemPrompt string
    UserMessage  string
    Context      string // RAG context
    Tools        []*ToolSpec
    MaxTokens    int
    Temperature  float32
}

type CompletionResponse struct {
    FinishReason string // "stop", "tool_calls", "length"
    Content      string // Resposta em texto
    ToolCalls    []*ToolCall
    Thinking     string // (Futuro) Raciocínio do modelo
}

type ToolCall struct {
    ID        string
    ToolName  string
    Arguments json.RawMessage
}

type ModelInfo struct {
    Name           string
    Provider       string // "gemini" ou "qwen"
    MaxContextSize int
    EmbeddingDim   int // 384 ou 1536
    Supports       []string // "tool_use", "vision", "streaming"
}
```

### 7.2 Implementação Gemini (`internal/llm/gemini.go`)

**Tecnologia:** Google Gemini API (via langchaingo/providers/google)

**Ciclo de Vida:**

```go
type GeminiProvider struct {
    client  *generativelanguage.Client
    apiKey  string
    model   string // "gemini-2.0-flash"
}

func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
    // Valida key, cria client
}

func (p *GeminiProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // 1. Build system prompt
    // 2. Convert tools to Gemini format
    // 3. Call API with streaming
    // 4. Parse response (text ou tool calls)
    // 5. Return
}

func (p *GeminiProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
    // Usa embedding model do Gemini (dimensão 768)
}
```

**Características:**

- **Tool Calling:** Via `function_calling` do Gemini
- **Vision:** Suporta embeddings de imagens
- **Streaming:** Respostas streamed para UI
- **Fallback:** Se API falha, retorna erro que UI pode exibir

**Configuração:**
- API Key carregada de `.env` ou configurado via Web UI
- Validação antes de salvar no `.env`
- Log de falhas de autenticação para debug

### 7.3 Implementação Qwen Local (`internal/llm/qwen.go`)

**Tecnologia:** llama.cpp sidecar + Qwen3 4B GGUF

**Processo de Boot:**

```
1. Daemon startup
2. LLMManager detecta que Qwen é o provider preferido
3. Engines manager checa se llama.cpp binary existe
4. Se não existe, emite event "engine.setup_required"
5. Se existe, spawna sidecar llama-cli:
   → cmd: llama-cpp --model qwen.gguf --port 127.0.0.1:0 --embeddings
   → Localhost port é aleatório, registrado no QwenProvider
```

**Interface do Sidecar (STDIO JSON-ND):**

```json
Request:
{
  "method": "complete",
  "payload": {
    "prompt": "...",
    "max_tokens": 1024,
    "temperature": 0.7
  }
}

Response:
{
  "id": "req-uuid",
  "text": "...",
  "finish_reason": "stop"
}
```

**Ciclo de Vida do Provider:**

```go
type QwenProvider struct {
    process      *exec.Cmd
    stdin        io.Writer
    stdout       io.ReadCloser
    tempDir      string
    modelPath    string
}

func NewQwenProvider(modelPath string) (*QwenProvider, error) {
    // 1. Valida GGUF
    // 2. Detecta porta livre
    // 3. Spawna llama-cpp
    // 4. Testa conexão com health check
}

func (p *QwenProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // 1. Serializa req como JSON
    // 2. Escreve em stdin do sidecar
    // 3. Lê resposta do stdout
    // 4. Parse resposta
    // 5. Return
}

func (p *QwenProvider) Shutdown(ctx context.Context) error {
    // SIGTERM ao processo
    // Wait com timeout
    // SIGKILL se não responde
}
```

### 7.4 Seleção e Fallback

**Lógica de Provider:**

```go
func (m *LLMManager) SelectProvider(ctx context.Context) (Provider, error) {
    preferredProvider := config.PreferredLLMProvider // "gemini" ou "qwen_local"

    switch preferredProvider {
    case "gemini":
        if !geminiProvider.IsConfigured() {
            return nil, errors.New("Gemini API Key não configurada")
        }
        // Test connection
        if err := testGemini(ctx); err != nil {
            // Fallback para Qwen local se disponível
            if qwenProvider.IsAvailable() {
                return qwenProvider, nil
            }
            return nil, errors.New("Nenhum LLM disponível")
        }
        return geminiProvider, nil

    case "qwen_local":
        if !qwenProvider.IsAvailable() {
            return nil, errors.New("Qwen local não instalado")
        }
        return qwenProvider, nil
    }
}
```

---

## 8. ENGINES: GERENCIADOR DO LLAMA.CPP (`internal/engines`)

### 8.1 Arquitetura e Responsabilidades

**Objetivo:** Abstrair a complexidade do llama.cpp (múltiplas builds, versões, backends GPU/CPU) em uma interface simples.

**Pacotes:**

```
internal/engines/
├── catalog.go       # Definições de versão e builds
├── detector.go      # Detecção de capacidades de hardware
├── downloader.go    # HTTP downloader com resume
├── extractor.go     # ZIP extractor seletivo
├── integrity.go     # SHA256 verification
├── manager.go       # Orquestrador principal
├── process.go       # Subprocess management
└── paths.go         # Resolução de caminhos
```

### 8.2 Detector de Hardware (`internal/engines/detector.go`)

**Responsabilidade:** Detectar CPU/GPU disponível e recomendar melhor build.

```go
type Hardware struct {
    OS            string   // "windows", "linux", "darwin"
    Architecture  string   // "x86_64", "arm64"
    CPUFeatures   []string // "avx2", "avx512", "neon"
    GPUType       string   // "none", "cuda", "vulkan", "metal"
    GPUVersion    string   // "11.8", "12.0" (para CUDA)
    RAM           int64    // Em bytes
}

func DetectHardware(ctx context.Context) (*Hardware, error) {
    // Windows: Checa registry, usa cpuid, GetLogicalProcessorInformation
    // Linux: /proc/cpuinfo, nvidia-smi
    // macOS: system_profiler, sysctl
}

func RecommendBuild(hw *Hardware) string {
    // Retorna ID da build mais adequada
}
```

### 8.3 Catálogo (`internal/engines/catalog.go`)

**Embarcado no binário via `go:embed`.**

```json
{
  "version": "b3430",
  "timestamp": "2026-04-05T00:00:00Z",
  "builds": [
    {
      "id": "llama-windows-x86-cuda-12-q6",
      "os": "windows",
      "arch": "x86_64",
      "cpu_features": ["avx2"],
      "gpu": "cuda",
      "gpu_version": "12.0",
      "download_url": "https://github.com/ggerganov/llama.cpp/releases/...",
      "sha256": "abc123...",
      "size_bytes": 2500000000,
      "description": "CUDA 12 optimized"
    }
  ]
}
```

### 8.4 Manager Principal (`internal/engines/manager.go`)

```go
type EngineManager interface {
    // Installation
    Install(ctx context.Context, buildID string, onProgress func(current, total int64)) error

    // Activation
    SetActive(ctx context.Context, buildID string) error
    GetActive(ctx context.Context) (*EngineInfo, error)

    // Lifecycle
    Start(ctx context.Context) (*EngineProcess, error)
    Stop(ctx context.Context) error
    IsRunning(ctx context.Context) bool

    // Registry
    ListInstalled(ctx context.Context) ([]*EngineInfo, error)
    Uninstall(ctx context.Context, buildID string) error
}
```

---

## 9. INTERFACE WEB: WAILS + NEXT.JS (`cmd/vectora-app` e `internal/app`)

### 9.1 Stack Tecnológico

- **Framework:** Wails v3 (Go + Webview2/WebKit)
- **Frontend:** Next.js 14 (App Router, SSG)
- **Styling:** TailwindCSS + Shadcn/UI
- **State:** Zustand
- **Animations:** Framer Motion
- **Build Tool:** Bun

### 9.2 Arquitetura de Binding (`cmd/vectora-app/app.go`)

```go
type App struct {
    ipcClient *ipc.Client
    ctx       context.Context
}

func (a *App) CallIPC(method string, payload json.RawMessage) (json.RawMessage, error) {
    return a.ipcClient.Request(method, payload)
}

func (a *App) OnEvent(eventName string, callback func(json.RawMessage)) {
    a.ipcClient.Subscribe(eventName, callback)
}
```

### 9.3 Estrutura de Componentes

**Rotas:**
- `/` (chat principal)
- `/workspaces` (galeria)
- `/settings` (configurações)

**Componentes:** ChatFeed, InputArea, Sidebar, CodeDiffVisualizer

### 9.4 Estética

- **Color Palette:** Kaffyn Dark (Zinc, Slate, Emerald)
- **RN-UI-01:** Glassmorphism em sidebars
- **RN-UI-02:** Micro-animações (Framer Motion)
- **RN-UI-03:** Dark mode obrigatório

---

## 10. INTERFACE CLI: BUBBLETEA (`cmd/vectora` subcomando)

### 10.1 Arquitetura TUI

**Stack:** Bubbletea + Lipgloss + Bubbles

### 10.2 Flags

```bash
vectora chat --workspace godot     # Carregar workspace
vectora chat --index /path/to/dir  # Indexar + chat
vectora chat --undo                # Reverter última ferramenta
```

### 10.3 Performance

- RAM: < 10MB
- Latência: < 100ms
- Suporte a No-Color mode

---

## 11. VECTORA INDEX: MARKETPLACE (`index-server`)

### 11.1 Endpoints

```
GET /api/v1/catalog         # Listar datasets
POST /api/v1/download/{id}  # Download dataset
POST /api/v1/publish        # Publicar dataset
```

### 11.2 Workflow

1. Browse datasets
2. Download com verificação MD5
3. Extrai para chromem-go
4. Registra novo workspace

---

## 12. SUITE DE TESTES (`cmd/tests`)

### 12.1 Flag `--tests`

```bash
vectora --tests              # Suite completa
vectora --tests --suite core # Apenas RAG
```

### 12.2 Suites

- Core RAG: indexação, queries, relevância
- Tool Execution: read, write, edit, find
- IPC Protocol: 100 reqs, timeout, rate limit
- Concurrency: race detection, GC, cleanup

**Critérios:** Todos passam, < 3 min, zero memory leaks

---

## 13. BUILD E ORQUESTRAÇÃO

### 13.1 Makefile

```makefile
make setup      # Setup inicial
make dev-web    # Dev com hot reload
make build-all  # Compilar todos binários
make test       # Rodar suite de testes
```

### 13.2 PowerShell (build.ps1)

- Detecta Go, Node, Bun
- Compila sequencialmente
- Gera checksums SHA256
- Empacota .zip

---

## 14. CONFIGURAÇÃO E SEGURANÇA

### 14.1 Arquivo `.env`

```env
GEMINI_API_KEY=sk-...
MAX_RAM_DAEMON=4294967296
PREFERRED_LLM_PROVIDER=qwen_local
LOG_LEVEL=INFO
```

### 14.2 Isolamento de Workspaces

- Iron Rule 3: Nenhum vetor vaza entre workspaces
- Cada workspace: collection isolada em chromem-go
- IPC valida `ws_id` antes de acesso

### 14.3 GitBridge Audit Trail

- UUIDs + timestamps
- Armazenados em `~/.Vectora/backups/`
- Logados para auditoria
- Nunca deletados (purga após 90 dias)

---

## 15. REGRAS DE NEGÓCIO CONSOLIDADAS

### 15.1 Arquitetura
- **RN-ARCH-01:** Single daemon Go; interfaces são clientes IPC
- **RN-ARCH-02:** Max 4GB RAM sistema, 512MB indexação
- **RN-ARCH-03:** JSON-ND, nunca HTTP internamente
- **RN-ARCH-04:** Go 1.22+

### 15.2 IPC
- **RN-IPC-01:** Erro com `code` + `message`
- **RN-IPC-02:** Conexões inativas > 30min fechadas
- **RN-IPC-03:** Todas calls logadas para auditoria
- **RN-IPC-04:** Rate limit: 1000 req/min

### 15.3 Database
- **RN-DB-01:** bbolt (metadata) + chromem-go (vetores)
- **RN-DB-02:** Max 512 tokens/chunk, 50 overlap
- **RN-DB-03:** Top-K = 5 (RAG)
- **RN-DB-04:** Isolamento total workspaces

### 15.4 LLM & Engines
- **RN-LLM-01:** Gemini API com fallback Qwen
- **RN-LLM-02:** Qwen sidecar subprocess (stdin/stdout)
- **RN-LLM-03:** Llama.cpp < 20MB download
- **RN-LLM-04:** SHA256 obrigatório

### 15.5 Tools & Agency
- **RN-TOOL-01:** Mutação aciona GitBridge.Snapshot()
- **RN-TOOL-02:** shell_command sem sanção bloqueado
- **RN-TOOL-03:** Comandos perigosos (rm, sudo) dupla confirmação
- **RN-TOOL-04:** Thought stream persistido em bbolt

### 15.6 UI & UX
- **RN-UI-01:** Web UI < 50MB, boot < 1.5s
- **RN-UI-02:** CLI < 10MB, latência < 100ms
- **RN-UI-03:** Dark mode forçado
- **RN-UI-04:** Zero fetch externos (Gemini via Go)

### 15.7 Testing & QA
- **RN-TEST-01:** Nenhum teste falha por rede
- **RN-TEST-02:** Suite < 3 minutos
- **RN-TEST-03:** Logs de falha em `tests/logs/latest_fail.txt`
- **RN-TEST-04:** Flag `--tests` auditoria completa

---

## 16. SEQUENCIAMENTO E INTERDEPENDÊNCIAS

### Fase 1: Fundação (2-3 sem)
1. infra (config, logger)
2. ipc (server, client)
3. db (bbolt, chromem-go)
4. tests (setup isolado)

### Fase 2: IA e RAG (3-4 sem)
1. llm (provider interface, Gemini)
2. engines (detector, manager, llama)
3. core (RAG pipeline, workspace)
4. index (index client)

### Fase 3: Tools e Agency (2-3 sem)
1. tools (filesystem, web, shell)
2. git (GitBridge)
3. acp (registry, execution)
4. mcp (MCP server)

### Fase 4: Interfaces (3-4 sem)
1. app (Next.js + Wails)
2. cli (Bubbletea)
3. tray (system tray)
4. installer (Fyne + setup)

### Fase 5: Integração (1-2 sem)
1. Build unificado
2. Testes E2E
3. Performance profiling
4. Documentação

---

## 17. VERIFICAÇÃO DE SUCESSO

### Funcionalidade
- [ ] Daemon roda 24h estável
- [ ] RAG query < 2s (embed + search + contexto + LLM)
- [ ] Tool execution + GitBridge undo ponta-a-ponta
- [ ] CLI e Web UI sincronizam em tempo real
- [ ] Qwen e Gemini intercambiáveis sem restart

### Performance
- [ ] Daemon idle: < 100MB
- [ ] 5 workspaces: < 500MB
- [ ] 1000 arquivos indexados: 2-5 min
- [ ] IPC: < 1ms latência
- [ ] Web UI boot: < 1.5s

### Qualidade
- [ ] Suite `--tests` 100% pass
- [ ] Zero memory leaks
- [ ] Race condition free
- [ ] Coverage > 80% (core)
- [ ] Logs parseáveis

### Segurança
- [ ] GitBridge snapshots funcionais
- [ ] Tool logging completo
- [ ] Sem vazamento de API keys
- [ ] Isolamento total workspaces

### Documentação
- [ ] README PT + EN
- [ ] CONTRIBUTING com setup guide
- [ ] Inline comments em funções complexas
- [ ] ADRs para trade-offs

---

## CONCLUSÃO

Este documento é o **Single Source of Truth** para Vectora v2.0. Implementação deve aderir rigorosamente aos esquemas, interfaces e regras descritos.

**Próximo:** Aprovação arquitetural → Fase 1 (Infra + IPC)

---

**Versão:** 2.0
**Data:** 2026-04-05
**Autores:** Kaffyn Engineering
**Status:** Engineering Specification
