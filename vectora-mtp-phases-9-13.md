# Vectora Multi-Tenancy Protocol (MTP) - Phases 9-13

## Visão Geral

Implementar um **protocolo de isolamento de tenants** para que uma **única instância singleton** do Vectora Daemon gerencie múltiplos projetos/workspaces simultâneos, mantendo baixo consumo de memória e isolamento absoluto de:

- Estados de conversa
- Índices vetoriais de código
- Limites de concorrência em provedores LLM
- Acesso a arquivos no disco (Trust Folders)

---

## Phase 9: Modelo de Tenant Baseado em Conexão (Connection-Bound Tenancy)

### Objetivo

Estabelecer o escopo de atuação de cada conexão IPC associando-a a um Workspace/Projeto específico.

### Implementação

**File:** `core/api/ipc/workspace.go`

```go
// WorkspaceContext define o contexto de um tenant conectado
type WorkspaceContext struct {
    WorkspaceID string
    WorkspaceRoot string
    ProjectName string
    ContextCancel context.CancelFunc
}

// WorkspaceInitRequest - mensagem de inicialização
type WorkspaceInitRequest struct {
    WorkspaceRoot string `json:"workspace_root"`
    ProjectName string `json:"project_name"`
}

// GenerateWorkspaceID cria ID único baseado no path
func GenerateWorkspaceID(workspaceRoot string) string {
    hash := sha256.Sum256([]byte(workspaceRoot))
    return hex.EncodeToString(hash[:])
}

// HandleWorkspaceInit processa inicialização de workspace
func (h *IPCHandler) HandleWorkspaceInit(req WorkspaceInitRequest) (*WorkspaceContext, error) {
    wsID := GenerateWorkspaceID(req.WorkspaceRoot)

    ctx := &WorkspaceContext{
        WorkspaceID: wsID,
        WorkspaceRoot: req.WorkspaceRoot,
        ProjectName: req.ProjectName,
    }

    // Passar workspace context para handlers posteriores
    return ctx, nil
}
```

### Fluxo

1. Cliente (VS Code Extension) se conecta ao IPC socket
2. Envia `workspace.init` com `workspace_root` e `project_name`
3. Daemon gera `WorkspaceID` consistente (SHA256 do path)
4. Todas as mensagens subsequentes naquele socket estão vinculadas a esse tenant
5. Cuando desconecta, o tenant é evicted da memória após timeout

### Verificação

```bash
# Testar inicialização de workspace
vectora-test workspace.init "C:\Users\bruno\Projects\MyApp" "MyApp"
# Deve retornar: WorkspaceID: <sha256hash>
```

---

## Phase 10: Abstração de Armazenamento e Isolamento de Estado (Storage & State Isolation)

### Objetivo

Garantir que cada tenant tenha seu próprio espaço de armazenamento isolado para índices, históricos e configurações.

### Estrutura de Pastas

```
%APPDATA%\Vectora\
├── global.db                    # Configurações globais e chaves API
├── ipc.token                    # Token de segurança local
└── workspaces/                  # Multi-tenancy storage
    ├── <workspace_id_A>/
    │   ├── chromadb/            # Índice vetorial isolado
    │   ├── chat_history.db      # Histórico de conversas
    │   ├── guardian.json        # Regras de acesso
    │   └── state.json           # Estado da sessão
    └── <workspace_id_B>/
        ├── chromadb/
        ├── chat_history.db
        ├── guardian.json
        └── state.json
```

### Implementação

**File:** `core/manager/tenant.go`

```go
// TenantManager gerencia lifecycle dos tenants
type TenantManager struct {
    mu sync.RWMutex
    activeTenants map[string]*Tenant
    evictionPolicy EvictionPolicy
}

// Tenant representa um workspace isolado
type Tenant struct {
    ID string
    Root string
    ProjectName string
    VectorStore db.VectorStore
    KVStore db.KVStore
    Guardian *policies.Guardian
    LastActivity time.Time
}

// GetOrCreateTenant obtém ou cria tenant para workspace
func (tm *TenantManager) GetOrCreateTenant(wsID, wsRoot, projName string) (*Tenant, error) {
    tm.mu.Lock()
    defer tm.mu.Unlock()

    if t, exists := tm.activeTenants[wsID]; exists {
        t.LastActivity = time.Now()
        return t, nil
    }

    // Criar novo tenant
    tenant := &Tenant{
        ID: wsID,
        Root: wsRoot,
        ProjectName: projName,
        LastActivity: time.Now(),
    }

    // Carregar stores específicos do workspace
    dataDir := fmt.Sprintf("%%APPDATA%%\\Vectora\\workspaces\\%s", wsID)

    vecStore, err := db.NewVectorStoreAtPath(filepath.Join(dataDir, "chromadb"))
    if err != nil {
        return nil, err
    }

    kvStore, err := db.NewKVStoreAtPath(filepath.Join(dataDir, "chat_history.db"))
    if err != nil {
        return nil, err
    }

    tenant.VectorStore = vecStore
    tenant.KVStore = kvStore
    tenant.Guardian = policies.NewGuardian(wsRoot)

    tm.activeTenants[wsID] = tenant
    return tenant, nil
}

// ReleaseTenant descarrega tenant da memória
func (tm *TenantManager) ReleaseTenant(wsID string) error {
    tm.mu.Lock()
    defer tm.mu.Unlock()

    if t, exists := tm.activeTenants[wsID]; exists {
        // Fechar stores
        t.VectorStore.Close()
        t.KVStore.Close()
        delete(tm.activeTenants, wsID)
    }

    return nil
}

// EvictionPolicy define política de descarregamento
type EvictionPolicy struct {
    IdleTimeout time.Duration // padrão: 30 minutos
    MaxTenants int
}
```

### Verificação

```bash
# Testar isolamento de dados
ls %APPDATA%\Vectora\workspaces\
# Deve mostrar diretórios por workspace_id
```

---

## Phase 11: Segurança e Limites Absolutos (Trust Folders & Guardian)

### Objetivo

Garantir que um tenant comprometido não consiga acessar dados de outro tenant.

### Implementação

**File:** `core/manager/tenant_security.go`

```go
// TenantSecurityInterceptor intercepta acesso a recursos
type TenantSecurityInterceptor struct {
    tenantID string
    tenantRoot string
}

// ValidateFilePath verifica se path está dentro do workspace
func (tsi *TenantSecurityInterceptor) ValidateFilePath(reqPath string) (string, error) {
    absPath, err := filepath.Abs(reqPath)
    if err != nil {
        return "", fmt.Errorf("invalid path: %w", err)
    }

    // Resolver path absoluto
    absTenantRoot, err := filepath.Abs(tsi.tenantRoot)
    if err != nil {
        return "", err
    }

    // Verificar se está dentro do workspace
    if !strings.HasPrefix(absPath, absTenantRoot) {
        return "", fmt.Errorf("path escapes workspace bounds: %s", absPath)
    }

    return absPath, nil
}

// ContextWithTenant injeta tenant no context
func ContextWithTenant(ctx context.Context, tenant *Tenant) context.Context {
    return context.WithValue(ctx, "tenant", tenant)
}

// TenantFromContext extrai tenant do context
func TenantFromContext(ctx context.Context) *Tenant {
    t := ctx.Value("tenant")
    if t == nil {
        return nil
    }
    return t.(*Tenant)
}

// GuardianValidate valida operação contra regras do tenant
func GuardianValidate(ctx context.Context, filePath string) error {
    tenant := TenantFromContext(ctx)
    if tenant == nil {
        return errors.New("no tenant in context")
    }

    validPath, err := tenant.Guardian.ValidatePath(filePath)
    if err != nil {
        return fmt.Errorf("access denied: %w", err)
    }

    return nil
}
```

### Modificar Handlers IPC

```go
// Em core/api/ipc/server.go
func (s *IPCServer) handleRequest(req JSONRPCRequest, conn net.Conn, tenant *Tenant) {
    // Injeta tenant no context
    ctx := ContextWithTenant(context.Background(), tenant)

    // Todos os handlers subsequentes respeitam o isolamento
    result := handleMethod(ctx, req.Method, req.Params)

    s.sendResponse(conn, result)
}
```

### Verificação

```bash
# Testar proteção contra path traversal
vectora-test workspace.query "C:\..\..\SensitiveData\secret.env"
# Deve retornar: ERROR - path escapes workspace bounds
```

---

## Phase 12: Paralelismo Seguro e Pool de Recursos (Resource Throttling)

### Objetivo

Prevenir que um tenant consuma todos os recursos (LLM rate limits, CPU, memória) afetando outros tenants.

### Implementação

**File:** `core/manager/resource_pool.go`

```go
// ResourcePool gerencia recursos por tenant
type ResourcePool struct {
    mu sync.RWMutex

    // Semáforos por tenant
    llmSemaphores map[string]*semaphore.Weighted

    // Prioridades de indexação
    indexQueue *PriorityQueue

    config ResourceConfig
}

type ResourceConfig struct {
    MaxParallelLLMPerTenant int // padrão: 2
    MaxConcurrentIndexing int // padrão: 4
    LLMRateLimit time.Duration // padrão: 100ms entre calls
}

// AcquireLLMSlot adquire permissão para chamar LLM
func (rp *ResourcePool) AcquireLLMSlot(tenantID string, ctx context.Context) error {
    rp.mu.Lock()
    sem, exists := rp.llmSemaphores[tenantID]
    if !exists {
        sem = semaphore.NewWeighted(int64(rp.config.MaxParallelLLMPerTenant))
        rp.llmSemaphores[tenantID] = sem
    }
    rp.mu.Unlock()

    // Tenant só pode usar sua própria cota
    return sem.Acquire(ctx, 1)
}

// ReleaseLLMSlot libera slot de LLM
func (rp *ResourcePool) ReleaseLLMSlot(tenantID string) {
    rp.mu.RLock()
    sem := rp.llmSemaphores[tenantID]
    rp.mu.RUnlock()

    if sem != nil {
        sem.Release(1)
    }
}

// EnqueueIndexing enfileira tarefa de indexação com prioridade
func (rp *ResourcePool) EnqueueIndexing(tenantID string, priority int, task func()) {
    rp.indexQueue.Enqueue(&IndexTask{
        TenantID: tenantID,
        Priority: priority,
        Callback: task,
    })
}

// PriorityQueue executa tasks respeitando prioridade e cota por tenant
type PriorityQueue struct {
    mu sync.Mutex
    items []*IndexTask
    semaphore *semaphore.Weighted
}

type IndexTask struct {
    TenantID string
    Priority int
    Callback func()
}
```

### Modificar Provider de LLM

```go
// Em core/llm/router.go
func (r *Router) QueryWithTenantLimit(ctx context.Context, model string, prompt string) (string, error) {
    tenant := TenantFromContext(ctx)
    if tenant == nil {
        return "", errors.New("no tenant in context")
    }

    // Adquirir slot da pool de recursos
    if err := r.resourcePool.AcquireLLMSlot(tenant.ID, ctx); err != nil {
        return "", fmt.Errorf("rate limited: %w", err)
    }
    defer r.resourcePool.ReleaseLLMSlot(tenant.ID)

    // Fazer requisição ao LLM
    return r.queryLLM(model, prompt)
}
```

### Verificação

```bash
# Testar rate limiting por tenant
# Terminal 1: Fazer 5 requisições simultâneas no Projeto A
# Terminal 2: Fazer 5 requisições simultâneas no Projeto B
# Esperado: Projeto A usa sua cota (2), Projeto B usa sua cota (2), resto enfileira
```

---

## Phase 13: Fluxo Completo da Arquitetura para Implementação

### Objetivo

Integrar todas as Phases 9-12 em um sistema coeso onde o daemon singleton gerencia múltiplos tenants com perfeito isolamento.

### Passo 1: Inicializar Manager de Tenants

**File:** `cmd/core/main.go`

```go
func main() {
    // ... existente ...

    // Inicializar manager de tenants
    tenantMgr := manager.NewTenantManager(
        manager.EvictionPolicy{
            IdleTimeout: 30 * time.Minute,
            MaxTenants: 10,
        },
    )

    // Inicializar resource pool
    resourcePool := manager.NewResourcePool(
        manager.ResourceConfig{
            MaxParallelLLMPerTenant: 2,
            MaxConcurrentIndexing: 4,
        },
    )

    // Injetar nas dependências do IPC
    ipcServer := ipc.NewServer(tenantMgr, resourcePool)

    go ipcServer.Start()
}
```

### Passo 2: Modificar IPC Handler

**File:** `core/api/ipc/server.go`

```go
type IPCServer struct {
    tenantManager *manager.TenantManager
    resourcePool *manager.ResourcePool
    listener net.Listener
}

func (s *IPCServer) handleConnection(conn net.Conn) {
    defer conn.Close()

    var activeTenant *manager.Tenant

    for {
        var req JSONRPCRequest
        json.NewDecoder(conn).Decode(&req)

        // Mensagem workspace.init inicializa tenant
        if req.Method == "workspace.init" {
            var initReq manager.WorkspaceInitRequest
            json.Unmarshal(req.Params, &initReq)

            tenant, err := s.tenantManager.GetOrCreateTenant(
                manager.GenerateWorkspaceID(initReq.WorkspaceRoot),
                initReq.WorkspaceRoot,
                initReq.ProjectName,
            )

            activeTenant = tenant
            continue
        }

        // Todas as mensagens posteriores usam tenant ativo
        if activeTenant == nil {
            s.sendError(conn, "not initialized")
            continue
        }

        // Injeta tenant no context
        ctx := manager.ContextWithTenant(context.Background(), activeTenant)

        // Processa método com contexto de tenant
        result := s.handleMethod(ctx, req.Method, req.Params)
        s.sendResponse(conn, result)
    }
}
```

### Passo 3: Background Eviction Routine

**File:** `core/manager/tenant.go`

```go
// StartEvictionRoutine inicia rotina de limpeza de tenants inativos
func (tm *TenantManager) StartEvictionRoutine(policy EvictionPolicy) {
    ticker := time.NewTicker(5 * time.Minute)

    go func() {
        for range ticker.C {
            tm.mu.Lock()
            now := time.Now()

            for wsID, tenant := range tm.activeTenants {
                if now.Sub(tenant.LastActivity) > policy.IdleTimeout {
                    tenant.VectorStore.Close()
                    tenant.KVStore.Close()
                    delete(tm.activeTenants, wsID)
                }
            }

            tm.mu.Unlock()
        }
    }()
}
```

### Integração Completa no main.go

```go
// Em cmd/core/main.go - função main() modificada
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Componentes globais
    tenantMgr := manager.NewTenantManager(manager.EvictionPolicy{
        IdleTimeout: 30 * time.Minute,
        MaxTenants: 10,
    })
    tenantMgr.StartEvictionRoutine(tenantMgr.evictionPolicy)

    resourcePool := manager.NewResourcePool(manager.ResourceConfig{
        MaxParallelLLMPerTenant: 2,
        MaxConcurrentIndexing: 4,
    })

    // IPC Server com multi-tenancy
    ipcServer := ipc.NewServer(tenantMgr, resourcePool)
    go ipcServer.Start(ctx)

    // Aguardar sinais
    <-ctx.Done()
}
```

### Verificação Completa

```bash
# 1. Abrir 2 projetos diferentes no VS Code
# 2. Fazer queries em ambos simultaneamente
# 3. Verificar que:
#    ✅ Memory usage baixo (um único daemon)
#    ✅ Projeto A não vê dados do Projeto B
#    ✅ Índices vetoriais separados
#    ✅ Rate limiting por projeto

# 4. Fechar um projeto
# 5. Esperar 30+ minutos
# 6. Verificar que Projeto A foi evicted da RAM
```

---

## Dependency Graph Atualizado

```
Phase 0-8 (Protocol SDKs)
    │
    ├── Phase 9  (Tenant Model - Connection Binding)
    │   └── Phase 10 (Storage Isolation)
    │       └── Phase 11 (Security & Guardian)
    │           └── Phase 12 (Resource Throttling)
    │               └── Phase 13 (Full Integration)
    │
    └── All working with ACP/MCP/IPC protocols
```

---

## Verification Checklist (Phases 9-13)

- [ ] Phase 9: `workspace.init` gera WorkspaceID único e consistente
- [ ] Phase 10: Múltiplos workspaces têm diretórios isolados em `%APPDATA%\Vectora\workspaces\`
- [ ] Phase 11: Path traversal (`../`) é bloqueado por Guardian
- [ ] Phase 12: Rate limiting por tenant funciona (máx 2 LLM calls simultâneos por projeto)
- [ ] Phase 13: Daemon singleton gerencia 3+ projetos abertos sem crescimento de memória
- [ ] Eviction: Tenant inativo por 30+ min é removido da RAM
- [ ] Integration: VS Code Extension inicializa corretamente cada tenant ao abrir novo projeto
- [ ] Isolation: Projeto A não consegue ler `.env` do Projeto B

---

## Files to Create/Modify

### New Files

- `core/manager/tenant.go` - Tenant model e lifecycle
- `core/manager/tenant_security.go` - Security interceptors
- `core/manager/resource_pool.go` - Rate limiting e resource management
- `core/manager/manager.go` - TenantManager principal
- `core/api/ipc/workspace.go` - Workspace initialization handlers

### Modified Files

- `cmd/core/main.go` - Inicializar TenantManager e ResourcePool
- `core/api/ipc/server.go` - Injeta tenant no context
- `core/llm/router.go` - Respeita rate limits por tenant
- `core/db/store.go` - Usa paths por workspace

---

## Success Criteria

**Single `vectora.exe` daemon rodando em background gerencia:**

1. ✅ Múltiplos projetos abertos em VS Code
2. ✅ Isolamento absoluto de dados entre projetos
3. ✅ Gerenciamento eficiente de memória (eviction após idle)
4. ✅ Rate limiting justo entre projetos
5. ✅ Segurança contra path traversal e acesso cross-project
6. ✅ Performance idêntica a single-tenant em contexto único

**Resultado:** Multi-Tenancy Protocol totalmente funcional e pronto para produção.
