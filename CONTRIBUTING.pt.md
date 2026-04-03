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
│   ├── vectora-web/            # Binário Web UI (Wails)
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
            ├── cmd/vectora-web   (iniciado sob demanda)
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

### Requisitos

- Go 1.22+
- Node.js 20+ (Frontend)
- Bun (Build do Frontend)
- Wails CLI
- Binário `llama-cli` (Gerenciado em `internal/engines`)

### Build

```bash
# Build coletivo via powershell
./build.ps1

# Build individual
make build-tray
```

### Testes

```bash
# Executar auditoria de integridade completa
go run ./cmd/vectora --tests
```

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
