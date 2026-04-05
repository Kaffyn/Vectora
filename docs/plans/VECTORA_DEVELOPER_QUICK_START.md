# VECTORA: DEVELOPER QUICK START

**Status:** Guia Prático — Setup Rápido para Desenvolvimento
**Versão:** 1.0
**Data:** 2026-04-05
**Idioma:** Português (PT-BR)
**Escopo:** 30 minutos até primeira execução do Vectora

---

## 0. PRÉ-REQUISITOS (5 minutos)

### Sistema Operacional
- **Windows 10+** (recomendado para testes iniciais)
- **macOS 11+** (opcional)
- **Linux Ubuntu/Debian** (opcional)

### Ferramentas Necessárias

Instale uma vez (global):

```bash
# Windows (use PowerShell como Administrator)
winget install golang.go               # Go 1.22+
winget install nodejs.nodejs           # Node.js 20+
winget install git.git                 # Git
winget install oven-sh.bun             # Bun

# macOS (use Homebrew)
brew install go node git bun

# Linux Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y golang-go nodejs git npm
npm install -g bun
```

### Verificar Instalação

```bash
go version                             # go1.22+
node --version                         # v20+
npm --version                          # v10+
bun --version                          # v1.0+
git --version                          # v2.40+
```

---

## 1. CLONE REPOSITÓRIO (2 minutos)

```bash
# Clone com depth para mais rápido
git clone --depth 1 https://github.com/Kaffyn/Vectora.git vectora
cd vectora

# Checkout branch de desenvolvimento (ou main)
git checkout develop   # ou git checkout main
```

---

## 2. SETUP AMBIENTE (5 minutos)

### 2.1 Variáveis de Ambiente

Crie arquivo `.env` na raiz do projeto:

```bash
# Windows (PowerShell)
@"
GEMINI_API_KEY=
MAX_RAM_DAEMON=4294967296
MAX_RAM_INDEXING=536870912
PREFERRED_LLM_PROVIDER=qwen_local
LOG_LEVEL=DEBUG
LOG_FORMAT=json
"@ | Set-Content -Path .env

# macOS/Linux (bash)
cat > .env << 'EOF'
GEMINI_API_KEY=
MAX_RAM_DAEMON=4294967296
MAX_RAM_INDEXING=536870912
PREFERRED_LLM_PROVIDER=qwen_local
LOG_LEVEL=DEBUG
LOG_FORMAT=json
EOF
```

### 2.2 Instalar Dependências Go

```bash
# Download dependências
go mod download
go mod tidy

# Isso vai baixar todos os pacotes Go necessários
# (pode levar 1-2 minutos na primeira vez)
```

### 2.3 Instalar Dependências Frontend

```bash
# Navegar para diretório Web UI
cd internal/app

# Instalar dependências com Bun
bun install
# ou npm install (mais lento)

# Voltar para raiz
cd ../..
```

---

## 3. BUILD (5 minutos)

### 3.1 Build Daemon (Núcleo)

```bash
# Windows (PowerShell)
.\build.ps1 -step daemon

# macOS/Linux
make build-daemon
```

Resultado: `./build/vectora` (binary executável)

### 3.2 Build Web UI (Opcional para Dev)

```bash
# Compilar Next.js
cd internal/app
bun run build          # cria ./out com SSG
cd ../..

# Depois build o wrapper Wails
.\build.ps1 -step app

# macOS/Linux
make build-app
```

Resultado: `./build/vectora-app` (desktop app)

### 3.3 Quick Build All

```bash
# Windows (PowerShell)
.\build.ps1

# macOS/Linux
make build-all
```

---

## 4. PRIMEIRA EXECUÇÃO (3 minutos)

### 4.1 Run Daemon Diretamente

```bash
# Na raiz do projeto
./build/vectora

# Esperado:
# [INFO] Daemon started on socket: ~/.Vectora/run/vectora.sock
# [INFO] Listening for connections...
```

Deixe rodando em um terminal.

### 4.2 Test IPC Connection (Em outro terminal)

```bash
# Teste se daemon está respondendo
./build/vectora --test-ipc

# Esperado:
# [INFO] Connected to daemon ✓
# [INFO] IPC latency: 0.45ms
```

### 4.3 Run First Query (Opcional)

```bash
# Se tem workspace disponível
./build/vectora query --workspace "test" --message "Hello?"

# ou com CLI interativa
./build/vectora chat

# Tipo: seu primeira pergunta
# Esperado: resposta do LLM
```

---

## 5. DESENVOLVIMENTO LOCAL

### 5.1 Mode de Desenvolvimento (Recomendado)

**Terminal 1 - Daemon com Watch:**

```bash
# Instalar air para auto-reload
go install github.com/cosmtrek/air@latest

# Executar com auto-reload
air
```

Air monitora mudanças em `.go` files e recompila automaticamente.

**Terminal 2 - Frontend Dev Server:**

```bash
cd internal/app

# Modo desenvolvimento com hot-reload
bun run dev

# Acessa em http://localhost:3000
```

**Terminal 3 - Testes:**

```bash
# Watch mode para testes
go test -v -race ./... -watch

# ou rodar testes específicos
go test -v ./internal/ipc/...
```

### 5.2 Estrutura de Pastas para Edição

```
vectora/
├── cmd/vectora/
│   ├── main.go                 # ← Editar para adicionar flags
│   ├── app.go                  # ← Wails bindings
│   └── cli.go                  # ← Bubbletea UI
│
├── internal/infra/
│   ├── config.go               # ← Configuração
│   └── logger.go               # ← Logging
│
├── internal/ipc/
│   ├── server.go               # ← IPC Server
│   └── handlers.go             # ← Handle methods
│
├── internal/core/
│   ├── rag_pipeline.go         # ← RAG Logic
│   └── workspace.go            # ← Workspace CRUD
│
├── internal/app/
│   ├── app/                    # ← Páginas (Next.js)
│   ├── components/             # ← Componentes React
│   └── hooks/                  # ← Hooks customizados
│
└── tests/
    ├── integration/            # ← Testes IPC
    └── e2e/                    # ← Testes Playwright
```

### 5.3 Adicionar Novo Arquivo

Exemplo: Adicionar novo LLM Provider

```bash
# 1. Criar arquivo
touch internal/llm/new_provider.go

# 2. Adicionar ao código:
cat > internal/llm/new_provider.go << 'EOF'
package llm

type NewProvider struct {
    // fields
}

func NewNewProvider(config string) (*NewProvider, error) {
    return &NewProvider{}, nil
}

func (p *NewProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // implementation
    return nil, nil
}
EOF

# 3. Testar
go test ./internal/llm/... -v

# 4. Build
make build-daemon
```

---

## 6. TESTES

### 6.1 Unit Tests

```bash
# Rodar todos os testes
go test -v ./...

# Com coverage
go test -v -cover ./...

# Com race detection
go test -v -race ./...

# Teste específico
go test -v ./internal/ipc -run TestServer
```

### 6.2 Integration Tests

```bash
# Testes de integração (mais lentos)
go test -v -tags=integration ./tests/integration/...
```

### 6.3 System Tests (Completo)

```bash
# Roda suite completa de testes do daemon
./build/vectora --tests

# Com output verboso
./build/vectora --tests --verbose

# Apenas suite específico
./build/vectora --tests --suite core
```

### 6.4 Frontend Tests

```bash
cd internal/app

# Jest unit tests
bun test

# E2E tests com Playwright
bun exec playwright test

# Apenas um arquivo de teste
bun test components/Chat/__tests__/ChatFeed.test.tsx
```

---

## 7. DEBUGGING

### 7.1 Logs

```bash
# Aumentar verbosidade
export LOG_LEVEL=DEBUG

./build/vectora

# Logs são salvos em:
# ~/.Vectora/logs/daemon.log
# ~/.Vectora/logs/ipc.log
```

### 7.2 Profiling CPU/Memory

```bash
# Profile de CPU (30 segundos)
./build/vectora --cpuprofile=cpu.prof

# Depois, analizar:
go tool pprof cpu.prof

# Dentro do pprof:
(pprof) top10
(pprof) graph

# Profile de memória
./build/vectora --memprofile=mem.prof
go tool pprof mem.prof
```

### 7.3 Debugger com VS Code

Instale extension `Go`:

1. Abra VS Code na pasta `vectora`
2. Create `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Daemon",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/vectora",
            "env": {},
            "args": []
        }
    ]
}
```

3. Pressione F5 para debug
4. Coloque breakpoints nos arquivos `.go`

### 7.4 Inspect IPC Messages

```bash
# Monitor IPC traffic (simula client)
./build/vectora --ipc-monitor

# Mostra todas as mensagens sendo trocadas
# [IPC] → workspace.list {}
# [IPC] ← {workspaces: [...]}
```

---

## 8. CHECKLIST PARA PRIMEIRA SEMANA

### Dia 1: Setup
- [ ] Clonar repositório
- [ ] Instalar ferramentas (Go, Node, Bun)
- [ ] Setup .env
- [ ] `go mod download`
- [ ] `bun install` (Web UI)
- [ ] `make build-daemon` ou `.\build.ps1 -step daemon`
- [ ] Testar daemon: `./build/vectora --test-ipc`

### Dia 2: Entender Arquitetura
- [ ] Ler `VECTORA_ARCHITECTURE_OVERVIEW.md`
- [ ] Traçar fluxo IPC (IPC server → handlers)
- [ ] Olhar estrutura de pastas
- [ ] Rodar `go test ./internal/ipc -v` para entender

### Dia 3: Modificar Código
- [ ] Criar um novo handler IPC (exemplo: `info.version`)
- [ ] Testar com `./build/vectora`
- [ ] Fazer pequeña mudança em logger e recompilar
- [ ] Confirmar que hot-reload funciona com air

### Dia 4-5: Frontend
- [ ] Abrir `internal/app` no VS Code
- [ ] Rodar `bun run dev`
- [ ] Modifique um componente (e.g., cor da sidebar)
- [ ] Confirmar hot-reload no navegador
- [ ] Entender Zustand stores

---

## 9. TROUBLESHOOTING

### Problema: "go: command not found"
**Solução:**
```bash
# Reinstalar Go
# Windows: winget install golang.go
# macOS: brew install go

# Adicionar ao PATH (se necessário)
# Windows: C:\Program Files\Go\bin
# macOS: /usr/local/go/bin
```

### Problema: "bun: command not found"
**Solução:**
```bash
# Reinstalar Bun
npm install -g bun

# ou
# macOS: brew install bun
```

### Problema: IPC connection timeout
**Solução:**
```bash
# Verificar se daemon está rodando
ps aux | grep vectora          # macOS/Linux
tasklist | findstr vectora     # Windows

# Remover socket antigo
rm ~/.Vectora/run/vectora.sock
# ou
del %USERPROFILE%\.Vectora\run\vectora.sock

# Reiniciar daemon
./build/vectora
```

### Problema: "permission denied" em build.ps1
**Solução:**
```powershell
# Permitir scripts (PowerShell como Admin)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Depois rodar
.\build.ps1
```

### Problema: "No space left on device"
**Solução:**
```bash
# Limpar caches
rm -rf ~/.Vectora/temp/*
rm -rf ./build/*
go clean -cache

# Limpar node_modules se necessário
cd internal/app && rm -rf node_modules && bun install
```

---

## 10. COMANDOS ÚTEIS (Referência Rápida)

```bash
# BUILD
make build-daemon                # Linux/macOS
make build-app                   # Linux/macOS
make build-all                   # Linux/macOS
.\build.ps1                      # Windows
.\build.ps1 -step daemon         # Windows (daemon only)

# TEST
go test ./...                    # All tests
go test -race ./...              # Race detection
./build/vectora --tests          # System integrity

# DEV
air                              # Auto-reload daemon
cd internal/app && bun run dev   # Frontend dev server
go run ./cmd/vectora/...         # Run directly

# DEBUG
go test -v ./internal/ipc        # Verbose test
./build/vectora --test-ipc       # Test IPC
tail ~/.Vectora/logs/daemon.log  # View logs

# FORMAT
go fmt ./...                     # Format code
go vet ./...                     # Lint
cd internal/app && bun run lint  # Lint frontend

# CLEAN
make clean                       # Remove builds
go clean -cache                  # Clear cache
rm -rf internal/app/out          # Clear Next.js build
```

---

## 11. DOCUMENTAÇÃO RELACIONADA

- **VECTORA_ARCHITECTURE_OVERVIEW.md** - Diagrama de componentes e fluxos
- **CONTRIBUTING.md** / **CONTRIBUTING.pt.md** - Setup detalhado (para primeira vez)
- **VECTORA_IMPLEMENTATION_PLAN.md** - Especificação técnica completa
- **VECTORA_DEVELOPMENT_TIMELINE.md** - Roadmap e milestones

---

## 12. PRÓXIMAS AÇÕES APÓS SETUP

### Primeira Tarefa: Implementar Simple Test Handler

1. Abra `internal/ipc/handlers.go`
2. Adicione novo handler:

```go
func (s *Server) handleTestPing(msg *IPCMessage) (*IPCMessage, error) {
    return &IPCMessage{
        ID:      msg.ID,
        Type:    "response",
        Payload: map[string]interface{}{"pong": true},
    }, nil
}

// No handleMessage():
case "test.ping":
    return s.handleTestPing(msg)
```

3. Compile: `make build-daemon`
4. Test:
```bash
./build/vectora --test-ipc test.ping
```

### Segunda Tarefa: Adicionar Novo Componente React

1. Crie `internal/app/components/Example.tsx`:

```typescript
export function Example() {
  return (
    <div className="p-4 bg-zinc-800 rounded">
      <p className="text-white">Exemplo de componente</p>
    </div>
  );
}
```

2. Importe em `internal/app/app/chat/page.tsx`
3. Rode `cd internal/app && bun run dev`
4. Verá componente renderizado em http://localhost:3000

---

**Status:** ✅ Pronto para começar desenvolvimento
**Tempo Total:** ~30 minutos até primeira execução

**Próximo:** Abra `CONTRIBUTING.md` para setup mais detalhado se precisar

---

