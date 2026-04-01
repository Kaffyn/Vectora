# Contribuindo para o Vectora

> [!TIP]
> Read this file in another language.
> English [CONTRIBUTING.md] | Portuguese [CONTRIBUTING.pt.md]

Este documento é destinado a desenvolvedores que queiram contribuir com o Vectora. Ele cobre a filosofia do projeto, estrutura do repositório, decisões de arquitetura e as regras que mantêm o código consistente.

---

## Filosofia

O Vectora é construído em torno de três princípios inegociáveis:

**Local primeiro.** Toda feature central deve funcionar sem acesso à internet. Provedores cloud (Gemini) são extensões opt-in, nunca dependências.

**Baixo consumo.** O daemon do systray deve permanecer abaixo de 5MB de RAM. O sistema completo deve operar abaixo de 4GB de RAM em hardware modesto. Toda dependência adicionada deve justificar seu custo em memória e tamanho de binário.

**Go puro onde possível.** Evite bindings CGO, dependências pesadas em C++ e interpretadores em runtime, a menos que não exista alternativa viável em Go. As exceções são `llama.cpp` (sidecar de inferência) e `Fyne` (instalador, CGO/OpenGL).

---

## Estrutura do Repositório

O Vectora segue as convenções de layout padrão de projetos Go.

```markdown
vectora/
├── cmd/                        # Pontos de entrada dos binários (um por executável)
│   ├── vectora/                # Orquestrador principal (daemon systray)
│   │   └── main.go
│   ├── vectora-cli/            # Binário CLI (Bubbletea)
│   │   └── main.go
│   ├── vectora-web/            # Binário Web UI (Wails)
│   │   └── main.go
│   └── vectora-installer/      # Binário do instalador (Fyne)
│       └── main.go
│
├── internal/                   # Packages privados (não importáveis externamente)
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
│   │   └── qwen.go             # Implementação Qwen/llama.cpp
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
├── web/                        # Frontend Next.js (embarcado via Wails)
│   ├── src/
│   └── package.json
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
- `web/` é o frontend Next.js. Ele é embarcado no binário Wails em tempo de build e não deve depender de nenhum CDN externo em runtime.
- Preocupações transversais (logging, config, erros) pertencem a sub-packages em `internal/`, não espalhadas por features.

---

## Arquitetura

### O Systray como Daemon Central

O binário `cmd/vectora` é a única fonte de verdade para o estado da aplicação. Ele roda como processo em background desde o login e expõe um servidor IPC ao qual todas as outras interfaces se conectam.

```markdown
cmd/vectora (daemon systray)
    └── Servidor IPC
            ├── cmd/vectora-cli   (iniciado sob demanda)
            ├── cmd/vectora-web   (iniciado sob demanda)
            └── clientes MCP / ACP (externos)
```

Nenhum binário de interface mantém estado. Eles são clientes sem estado. Se uma interface crasha, o daemon e seu estado não são afetados.

### Contrato IPC

Toda comunicação entre o daemon e as interfaces passa pela camada IPC em `internal/ipc`. Adicionar uma nova interface significa implementar o cliente IPC, não duplicar a lógica central.

### Interface do Provedor LLM

Todas as interações com LLM passam pela interface `internal/llm.Provider`. Novos provedores devem implementar essa interface. Nenhum código específico de provedor vaza para `internal/core`.

```markdowngo
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
    Embed(ctx context.Context, input string) ([]float32, error)
    Name() string
}
```

### Execução de Tools e GitBridge

Toda tool em `internal/tools` que realiza uma operação de escrita ou shell deve chamar `internal/git.Bridge.Snapshot()` antes da execução. Isso é obrigatório, não opcional. O snapshot habilita o comando `undo` em todas as interfaces.

### Isolamento de Workspaces

Cada workspace mapeia para uma coleção isolada do chromem-go e um bucket dedicado no bbolt. Workspaces nunca devem compartilhar uma coleção ou fazer consultas cruzadas sem intenção explícita do usuário. O package `internal/core` impõe esse limite.

---

## Regras de Negócio

Estas regras definem restrições que devem ser respeitadas em todo o codebase.

**RN-01 — Orçamento de RAM**
O daemon do systray não deve exceder 5MB RSS em idle. O sistema completo (daemon + uma interface ativa + um workspace carregado) não deve exceder 4GB RSS. Qualquer PR que degrade isso mensuravelmente deve incluir justificativa.

**RN-02 — Sem Rede no Core**
`internal/core`, `internal/db`, `internal/tools` e `internal/git` não devem fazer chamadas de rede de saída. O acesso à rede é restrito a `internal/llm` (chamadas de provedor) e `internal/index` (cliente do Index).

**RN-03 — Isolamento de Workspaces**
Um workspace só pode ler de sua própria coleção vetorial. Consultas entre workspaces não são permitidas na camada de armazenamento. Agregação, se necessária, acontece na camada RAG em `internal/core`, nunca em `internal/db`.

**RN-04 — Proteção de Escrita**
Nenhuma tool pode escrever no sistema de arquivos ou executar comandos shell sem um snapshot Git prévio via `internal/git.Bridge`. Os testes devem verificar a criação do snapshot antes de operações de escrita.

**RN-05 — Abstração de Provedor**
Nenhum arquivo fora de `internal/llm` pode importar um SDK de provedor diretamente (SDK do Gemini, bindings do llama.cpp). Detalhes de provedor são totalmente encapsulados.

**RN-06 — Interfaces Sem Estado**
Binários de interface (`vectora-cli`, `vectora-web`) não mantêm estado da aplicação. Todo estado vive no daemon. Interfaces são descartáveis.

**RN-07 — Frontend Embarcado**
O frontend Next.js deve ser totalmente buildável e funcional sem nenhuma dependência de CDN externo. Todos os assets devem ser embarcáveis via `go:embed`.

**RN-08 — Curadoria do Index**
O cliente do Vectora Index (`internal/index`) só pode baixar datasets que possuam uma assinatura de revisão Kaffyn válida. O cliente deve rejeitar datasets não assinados ou adulterados no momento do download.

---

## Configuração do Ambiente

### Requisitos

- Go 1.22+
- Node.js 20+ (para o frontend web)
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Binário `llama.cpp` (para inferência local, opcional para desenvolvimento)

### Build

```bash
# Build de todos os binários
make build

# Build de um binário específico
make build-tray
make build-cli
make build-web
make build-installer

# Executar o daemon em modo de desenvolvimento
make dev
```

### Testes

```bash
# Executar todos os testes
make test

# Executar apenas testes de integração
make test-integration

# Executar com detector de race condition
make test-race
```

Todos os PRs devem passar na suite completa de testes, incluindo o detector de race condition, antes da revisão.

---

## Convenção de Commits

O Vectora usa [Conventional Commits](https://www.conventionalcommits.org/).

```markdown
<tipo>(escopo): <descrição>

Tipos: feat, fix, docs, refactor, test, chore, perf
Escopo: core, cli, web, tray, installer, mcp, acp, ipc, db, tools, index, git
```

**Exemplos:**

```
feat(core): adiciona agregação de consultas multi-workspace
fix(tools): garante snapshot git antes da execução shell
perf(db): reduz tempo de carregamento de coleção no chromem-go
docs(contributing): adiciona regras de isolamento de workspace
```

Commits que não seguirem a convenção não serão mergeados.

---

## Regras de Pull Request

- Todo PR deve referenciar uma issue aberta.
- PRs que tocam `internal/core` ou `internal/db` requerem duas aprovações.
- PRs não devem introduzir novas dependências CGO sem discussão prévia na issue.
- Todas as novas funções públicas em `internal/` devem ter comentários Go doc.
- Mudanças que quebram o contrato IPC requerem um caminho de migração documentado no PR.

---

## O Que Não Fazer

- Não adicione estado a binários de interface.
- Não chame SDKs de provedor fora de `internal/llm`.
- Não contorne o GitBridge para operações de escrita.
- Não adicione packages a `pkg/` sem discussão com o time.
- Não embuta URLs de CDN externo no frontend.
- Não introduza dependências que exijam um processo de servidor em execução (sem Postgres, sem Redis, sem bancos vetoriais externos).

---

_Parte da organização open source [Kaffyn](https://github.com/Kaffyn)._
