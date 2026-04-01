# Contribuindo para o Vectora

> [!TIP]
> Read this file in another language.
> English [CONTRIBUTING.md] | Portuguese [CONTRIBUTING.pt.md]

Este documento é destinado a desenvolvedores que queiram contribuir com o Vectora. Ele cobre a filosofia do projeto, estrutura do repositório, decisões de arquitetura e as regras que mantêm o código consistente.

---

## Filosofia

O Vectora é construído em torno de três princípios inegociáveis:

**Local primeiro.** Toda feature central deve funcionar sem acesso à internet. Provedores cloud (Gemini) são extensões opt-in, nunca dependências.

**Baixo consumo.** O daemon do systray deve permanecer abaixo de 100MB de RAM RSS em idle. O sistema completo deve operar abaixo de 4GB de RAM em hardware modesto. Toda dependência adicionada deve justificar seu custo em memória e tamanho de binário.

**Go puro onde possível.** Evite bindings CGO, dependências pesadas em C++ e interpretadores em runtime, a menos que não exista alternativa viável em Go. As exceções são `Fyne` (instalador, CGO/OpenGL) e `llama-cli` (usado como sidecar através de pipes).

---

## Estrutura do Repositório

O Vectora segue as convenções de layout padrão de projetos Go.

```markdown
vectora/
├── cmd/                        # Pontos de entrada dos binários (um por executável)
│   ├── vectora/                # Orquestrador principal (daemon systray) & CLI
│   │   └── main.go
│   ├── vectora-app/            # Binário Web UI (Wails)
│   │   └── main.go
│   └── vectora-installer/      # Binário do instalador (Fyne)
│       └── main.go
│
├── internal/                   # Packages privados (não importáveis externamente)
│   ├── app/                    # Frontend Next.js (Código fonte e Assets)
│   │   ├── app/                # Next.js App Router
│   │   ├── components/         # Componentes React
│   │   └── public/             # Assets Estáticos
│   ├── core/                   # Lógica de negócio: pipeline RAG, gerenciamento de workspaces
│   │   ├── rag.go
│   │   ├── workspace.go
│   │   └── indexer.go
│   ├── db/                     # Camada de banco de dados
│   │   ├── vector.go           # Wrapper do chromem-go
│   │   └── store.go            # Wrapper do bbolt
│   ├── llm/                    # Abstração de provedor LLM
│   │   ├── provider.go         # Definição da interface
│   │   ├── gemini.go           # Implementação Gemini
│   │   └── qwen.go             # Implementação Qwen/llama-cli via Pipes
│   ├── ipc/                    # Servidor e cliente IPC (systray ↔ interfaces)
│   │   ├── server.go
│   │   └── client.go
│   ├── tray/                   # UI e gerenciamento de ciclo de vida do systray
│   │   └── tray.go
│   ├── tools/                  # Toolkit agêntico (compartilhado entre MCP, ACP, CLI)
│   │   ├── filesystem.go       # read_file, write_file, read_folder, edit
│   │   ├── search.go           # find_files, grep_search
│   │   ├── web.go              # google_search, web_fetch
│   │   ├── shell.go            # run_shell_command
│   │   └── memory.go           # save_memory, enter_plan_mode
│   ├── git/                    # GitBridge: snapshot e rollback
│   │   └── bridge.go
│   ├── engines/                # Gestão de binários externos/sidecars
│   │   └── manager.go
│   ├── mcp/                    # Implementação do servidor MCP
│   │   └── server.go
│   ├── acp/                    # Implementação do agente ACP
│   │   └── agent.go
│   └── index/                  # Cliente do Vectora Index (catálogo, download, publicação)
│       └── client.go
│
├── pkg/                        # Packages públicos (seguros para importação externa)
│   └── vectorakit/             # SDK para integrações externas (futuro)
│
├── assets/                     # Assets estáticos embarcados (ícones, configs padrão)
├── scripts/                    # Scripts de build, release e setup
├── docs/                       # Documentação para desenvolvedores
├── tests/                      # Testes de integração e end-to-end
│   └── suite/
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── CONTRIBUTING.md
```

### Regras Fundamentais

- `cmd/` contém apenas pontos de entrada. Nenhuma lógica de negócio vive aqui — apenas inicialização, parsing de flags e cabeamento.
- `internal/` é onde toda a lógica vive. Packages aqui não podem ser importados por projetos externos.
- `pkg/` é reservado para código intencionalmente exposto a consumidores externos. Não adicione a `pkg/` sem discussão prévia.
- `internal/app/` é o frontend Next.js. Ele é embarcado no binário Wails em tempo de build a partir da exportação estática.
- Preocupações transversais (logging, config, erros) pertencem a sub-packages em `internal/infra/`, não espalhadas por features.

---

## Arquitetura

### O Systray como Daemon Central

O binário `cmd/vectora` é a única fonte de verdade para o estado da aplicação. Ele roda como processo em background e expõe um servidor IPC (Named Pipes no Windows ou Sockets no Unix).

```markdown
cmd/vectora (daemon systray)
    └── Servidor IPC
            ├── cmd/vectora-app   (iniciado sob demanda)
            ├── vectora chat / cli (embutido no daemon)
            └── clientes MCP / ACP (externos)
```

Nenhum binário de interface mantém estado. Eles são clientes sem estado que devem reconectar ao IPC se reiniciados.

### Contrato IPC

Toda comunicação entre o daemon e as interfaces passa pela camada IPC em `internal/ipc`. O protocolo é JSON-ND para eficiência em streaming.

### Interface do Provedor LLM

Todas as interações com LLM passam pela interface `internal/llm.Provider`. A inferência local (Qwen) é tratada via pipes de processo de standard I/O (Arquitetura Zero-Port).

```markdowngo
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
    Embed(ctx context.Context, input string) ([]float32, error)
    Name() string
}
```

### Execução de Tools e GitBridge

Toda tool em `internal/tools` que realiza uma operação de escrita ou shell deve chamar `internal/git.Bridge.Snapshot()` antes da execução. O snapshot habilita a reversão via `undo`.

---

## Regras de Negócio

**RN-01 — Orçamento de RAM**
O daemon deve permanecer abaixo de 100MB RSS em idle. O sistema completo não deve exceder 4GB RSS sob carga.

**RN-02 — Sem Rede no Core**
`internal/core`, `internal/db`, `internal/tools` e `internal/git` não devem fazer chamadas de rede. O acesso à rede é restrito a `internal/llm` (provedor) e `internal/index` (cliente Index).

**RN-03 — Isolamento de Workspaces**
Agregação, se necessária, acontece na camada RAG em `internal/core`, nunca no banco de dados vetorial.

**RN-04 — Proteção de Escrita**
Nenhuma ferramenta de escrita funciona sem um snapshot de proteção prévio via Git.

**RN-05 — Abstração de Provedor**
Nenhum arquivo fora de `internal/llm` deve importar chaves de API ou detalhes de drivers de inferência.

---

## Configuração do Ambiente

### Requisitos de Sistema

Antes de começar, certifique-se de ter os seguintes itens instalados:

- **Go 1.22+** (linguagem de programação principal do backend)
- **Node.js 20+** (ferramentas do frontend)
- **Bun 1.0+** (gerenciador de pacotes JavaScript rápido)
- **Wails CLI 2.5+** (framework Go + Webview para desktop)
- **Git 2.30+** (controle de versão)
- **CMake 3.20+** (opcional, para compilar dependências C/C++)

### Instruções de Instalação

Escolha seu sistema operacional abaixo e siga o guia passo-a-passo.

---

#### Windows (PowerShell)

**Passo 1: Instale Go**

```powershell
# Opção A: Usando Chocolatey
choco install golang

# Opção B: Instalação manual
# Baixe de https://go.dev/dl/
# Execute o instalador e siga as instruções
# Adicione Go ao PATH se não foi feito automaticamente

# Verifique a instalação
go version
```

**Passo 2: Instale Git**

```powershell
# Usando Chocolatey
choco install git

# Ou baixe de https://git-scm.com/download/win
# Verifique a instalação
git --version
```

**Passo 3: Instale Node.js e Bun**

```powershell
# Instale Node.js
choco install nodejs

# Verifique Node.js
node --version
npm --version

# Instale Bun
powershell -Command "irm bun.sh/install.ps1 | iex"

# Adicione Bun ao PATH (se necessário)
$env:Path += ";$env:USERPROFILE\.bun\bin"

# Verifique Bun
bun --version
```

**Passo 4: Instale Wails CLI**

```powershell
# Instale Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verifique a instalação
wails --version

# Se wails não for encontrado, adicione Go bin ao PATH
$env:Path += ";$env:USERPROFILE\go\bin"
```

**Passo 5: Clone o Repositório e Instale Dependências**

```powershell
# Clone o repositório
git clone https://github.com/Kaffyn/Vectora.git
cd Vectora

# Instale dependências Go
go mod download
go mod tidy

# Instale dependências do frontend
cd internal/app
bun install
cd ..\..
```

---

#### macOS

**Passo 1: Instale Homebrew** (se ainda não estiver instalado)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

**Passo 2: Instale Go**

```bash
# Usando Homebrew
brew install go@1.22

# Adicione ao PATH em ~/.zshrc ou ~/.bash_profile
echo 'export PATH="/usr/local/opt/go@1.22/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Verifique a instalação
go version
```

**Passo 3: Instale Git**

```bash
# Usando Homebrew
brew install git

# Verifique a instalação
git --version
```

**Passo 4: Instale Node.js e Bun**

```bash
# Instale Node.js
brew install node

# Verifique Node.js
node --version
npm --version

# Instale Bun
curl -fsSL https://bun.sh/install | bash

# Adicione Bun ao PATH em ~/.zshrc
export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"
source ~/.zshrc

# Verifique Bun
bun --version
```

**Passo 5: Instale Wails CLI**

```bash
# Instale Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verifique a instalação
wails --version
```

**Passo 6: Clone o Repositório e Instale Dependências**

```bash
# Clone o repositório
git clone https://github.com/Kaffyn/Vectora.git
cd Vectora

# Instale dependências Go
go mod download
go mod tidy

# Instale dependências do frontend
cd internal/app
bun install
cd ../..
```

---

#### Linux (Ubuntu/Debian)

**Passo 1: Instale Go**

```bash
# Baixe a versão mais recente de Go
cd /tmp
wget https://go.dev/dl/go1.22.linux-amd64.tar.gz

# Extraia para /usr/local
sudo tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz

# Adicione Go ao PATH em ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verifique a instalação
go version
```

**Passo 2: Instale Git**

```bash
# Usando apt
sudo apt-get update
sudo apt-get install git -y

# Verifique a instalação
git --version
```

**Passo 3: Instale Node.js e Bun**

```bash
# Instale Node.js via repositório NodeSource
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install nodejs -y

# Verifique Node.js
node --version
npm --version

# Instale Bun
curl -fsSL https://bun.sh/install | bash

# Adicione Bun ao PATH em ~/.bashrc
export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"
source ~/.bashrc

# Verifique Bun
bun --version
```

**Passo 4: Instale Dependências de Build**

```bash
# Necessário para Wails no Linux
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev -y

# Opcional mas recomendado
sudo apt-get install build-essential -y
```

**Passo 5: Instale Wails CLI**

```bash
# Instale Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verifique a instalação
wails --version
```

**Passo 6: Clone o Repositório e Instale Dependências**

```bash
# Clone o repositório
git clone https://github.com/Kaffyn/Vectora.git
cd Vectora

# Instale dependências Go
go mod download
go mod tidy

# Instale dependências do frontend
cd internal/app
bun install
cd ../..
```

---

### Compilando o Projeto

Após instalar todas as dependências, você pode compilar o Vectora:

**Opção 1: Compilar Todos os Componentes (Recomendado)**

```bash
# Windows (PowerShell)
.\build.ps1

# macOS/Linux (Make)
make build-all
```

Os binários de saída estão no diretório `./build/`:
- `vectora` (Daemon)
- `vectora-app` (Interface Web)
- `vectora-cli` (Interface Terminal)
- `vectora-setup.exe` (Instalador, apenas Windows)

**Opção 2: Compilar Componentes Individuais**

```bash
# Compile apenas o daemon
go build -o ./build/vectora ./cmd/vectora

# Compile a interface web
cd internal/app && bun run build
cd ../..
go build -o ./build/vectora-app ./cmd/vectora-app

# Compile a CLI
go build -o ./build/vectora-cli ./cmd/vectora-cli
```

### Testes

```bash
# Execute todos os testes com race detector
go test -v -race -count=1 ./...

# Execute apenas um pacote específico
go test -v ./internal/ipc

# Execute com relatório de cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Execute testes de integração
go run ./cmd/vectora --tests
```

**Importante:** Todos os PRs devem passar na suite de testes completa, incluindo o race detector, antes da revisão.

### Fluxo de Desenvolvimento

```bash
# 1. Inicie o daemon em modo debug
go run ./cmd/vectora daemon --log-level=DEBUG

# 2. Em outro terminal, inicie o servidor dev da interface web
cd internal/app
bun run dev

# 3. Ou execute a CLI
go run ./cmd/vectora chat --workspace <workspace-id>
```

### Resolução de Problemas

**"command not found: go"**
- Certifique-se de que Go está instalado e no PATH
- Verifique: `echo $PATH` (Linux/macOS) ou `echo %PATH%` (Windows)
- Reabra o terminal após a instalação

**"wails: command not found"**
- Certifique-se de que `$GOPATH/bin` está no PATH
- Execute: `export PATH=$PATH:$(go env GOPATH)/bin` (Linux/macOS)
- Execute: `$env:Path += ";$(go env GOPATH)\bin"` (Windows PowerShell)

**Erros "Module not found" durante a compilação**
- Execute: `go mod download && go mod tidy`

**Falhas de compilação no Windows**
- Certifique-se de que PowerShell está sendo executado como Administrador
- Execute: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`

**Testes falham com "port already in use"**
- Mate processos Vectora existentes
- Windows: `Get-Process vectora | Stop-Process -Force`
- Linux/macOS: `pkill -f vectora`

---

## Convenção de Commits

O Vectora usa [Conventional Commits](https://www.conventionalcommits.org/).

**Exemplos:**
```
feat(core): adiciona agregação de consultas multi-workspace
fix(tools): garante snapshot git antes da execução shell
```

---

_Parte da organização open source [Kaffyn](https://github.com/Kaffyn)._
