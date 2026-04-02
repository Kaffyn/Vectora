# Blueprint de Engenharia do Sistema Vectora

Este documento centraliza o plano de implementação ponta-a-ponta para o projeto **Vectora**, atuando como nosso Single Source of Truth (SSOT). Ele reflete rigorosamente as "Iron Rules" (Regras de Negócio Inegociáveis) detalhadas na documentação oficial e os contratos de infraestrutura (Local First, Baixo Consumo, Go Puro, Sem Estado).

## Visão Geral e Background

O Vectora é um assistente de IA offline, focado na ingestão privativa de contexto corporativo ou pessoal. Em vez da clássica arquitetura cliente/servidor baseada em APIs cloud, o Vectora consolida suas operações em um poderoso e leve daemon via systray (≤ 5MB idle) que controla e possui a orquestração via IPC (Inter-Process Communication) para todo o resto do sistema. Este plano cobre como criaremos, passo-a-passo, os componentes essenciais em Go garantindo as restrições arquiteturais.

---

## User Review Required

> [!IMPORTANT]
> Aprovação Arquitetural Solicitada
> Analise os pacotes e limites da arquitetura. Como o sistema exige 'Zero Estado' fora do Daemon, a estrutura em _Wails + IPC_ descrita na seção Web UI (e outras interfaces) foge do trivial. Solicito especial atenção:
>
> - Na definição de inicialização do sidecar `llama.cpp` (Qwen3).
> - Na garantia de que o `GitBridge` interceptará sempre a Camada Agêntica.

---

## Proposed Changes

Abaixo está o fluxo tático de estruturação dos pacotes no repositório. Lançamentos serão feitos sequencialmente seguindo a árvore de dependência sem furos.

### 1. Daemon Systray e Servidor IPC

**Componente Central**. Deve ser implementado inicialmente, já que nenhum Web UI pode subsistir sem o daemon principal para possuir o estado.

#### [NEW] [cmd/vectora/main.go](file:///c:/Users/bruno/Desktop/Vectora/cmd/vectora/main.go)

Script limpo que apenas realiza o parsing de flags, inicializa e acopla a camada IPC ao handler `systray`, além de invocar o setup do backend (Bbolt/Chroma via interfaces internas).

#### [NEW] [internal/tray/tray.go](file:///c:/Users/bruno/Desktop/Vectora/internal/tray/tray.go)

Gerenciador do ícone e ciclo de vida. Mantém a máquina de estados local, gerenciando a alocação do Processo Sidecar (LLaMA) ou de novas UIs Wails quando acionadas pelos botões do menu.

#### [NEW] [internal/ipc/server.go](file:///c:/Users/bruno/Desktop/Vectora/internal/ipc/server.go)

#### [NEW] [internal/ipc/client.go](file:///c:/Users/bruno/Desktop/Vectora/internal/ipc/client.go)

Implementação de transporte via Unix Domain Sockets e Named Pipes (`\\.\pipe\vectora` no Windows). Irão suportar streams de JSON (New Line Delimited) contendo eventos síncronos e broadcasts assíncronos.

---

### 2. Camada de Storage Local e Pipeline RAG

**Componente Lógico e Persistente**, estrita e totalmente enclausurado contra acessos em redes e sem acoplamento entre os workspaces na camada de hardware.

#### [NEW] [internal/db/vector.go](file:///c:/Users/bruno/Desktop/Vectora/internal/db/vector.go)

Wrapper ao `chromem-go`. Define acesso isolado ao Storage de vetores para embeddings em namespaces `ws:<workspace_id>`.

#### [NEW] [internal/db/store.go](file:///c:/Users/bruno/Desktop/Vectora/internal/db/store.go)

Wrapper ao `bbolt` para salvamento de Configurações do workspace e logs estruturados em chave-valor.

#### [NEW] [internal/core/workspace.go](file:///c:/Users/bruno/Desktop/Vectora/internal/core/workspace.go)

Gerencia a ativação/desativação simultânea de partições vetoriais, negando intersecções acidentais antes da etapa heurística RAG.

#### [NEW] [internal/core/rag.go](file:///c:/Users/bruno/Desktop/Vectora/internal/core/rag.go)

Lógica massiva: Busca distribuídas de N workspaces ativos simultâneos, ordenação cruzada semântica e composição de Prompt consolidado a ser despachado ao LLM.

---

### 3. Integração LLM Base (Langchaingo) e Sidecar

**Componente Inteligente**, abstraído rigidamente via interface Go. Módulo não é exposto a outras dependências diretamente, a não ser via IPC/Tray.

#### [NEW] [internal/llm/provider.go](file:///c:/Users/bruno/Desktop/Vectora/internal/llm/provider.go)

Interface consolidada Go com métodos obrigatórios: `Complete(...)` (para Prompt) e `Embed(...)` (para Vetores).

#### [NEW] [internal/llm/gemini.go](file:///c:/Users/bruno/Desktop/Vectora/internal/llm/gemini.go)

Implementação da API via API-Key gerenciada pelo serviço Crypto Keyring nativo do Sistema Operacional (nenhum dado vaza em RAM permanente).

#### [NEW] [internal/llm/qwen.go](file:///c:/Users/bruno/Desktop/Vectora/internal/llm/qwen.go)

Manuseio robusto do executável externo de inferência `llama.cpp` — binding dinâmico de porta localhost, ciclo de watchdog e repasse HTTP das inferências locais multimodais.

---

### 4. GitBridge & Kit de Ferramentas Agênticas

**Execução Segura**. A base do zero-risk do usuário; nenhuma operação no PC pode ser destruída pelas inferências do agente LLM inadvertidamente.

#### [NEW] [internal/git/bridge.go](file:///c:/Users/bruno/Desktop/Vectora/internal/git/bridge.go)

Encapsulamento de versionamento temporário do diretório atual de trabalho. Provê os métodos `.Snapshot()` e `.Restore()`.

#### [NEW] [internal/tools/filesystem.go](file:///c:/Users/bruno/Desktop/Vectora/internal/tools/filesystem.go)

Implementa interfaces de manipulação (`write_file`, `edit`, `read_folder`). Se os métodos Write existirem, o código irá falhar imediatamente caso não receba um Ticket de autorização via Snapshot concluído do GitBridge.

#### [NEW] [internal/tools/shell.go](file:///c:/Users/bruno/Desktop/Vectora/internal/tools/shell.go)

Dispensador e avaliador de sub-processos shell ativados nos endpoints e workspaces.

---

### 5. Interfaces Voláteis: Web UI & CLI

**Componente Front**. Sem dependência cloud, construído com foco em estética agressiva e modernidade.

#### [NEW] [web/package.json](file:///c:/Users/bruno/Desktop/Vectora/web/package.json)

Iniciaremos um projeto Next.js no schema SSR/Export (totalmente offline, via out static) que terá design estético premium.

#### [NEW] [cmd/vectora-web/main.go](file:///c:/Users/bruno/Desktop/Vectora/cmd/vectora-web/main.go)

Implementação da engine de visualização web utilizando a lib **Wails**. Iremos incorporar (`//go:embed`) o build pre-renderizado (Next.js out). Implementará os bindings TS->Go onde _SOMENTE_ enviará dados formatados em mapa para o serviço `internal/ipc`.

#### [NEW] [cmd/vectora-cli/main.go](file:///c:/Users/bruno/Desktop/Vectora/cmd/vectora-cli/main.go)

Interface interativa minimalista via TUI `Bubbletea`.

---

## Open Questions

> [!WARNING]
> Antes de programarmos e partirmos para implementação pesada, responda ou reflita sobre os pontos:
>
> 1. **Binário do `llama.cpp`:** Na documentação prevemos a inicialização via "sidecar". O binário e arquivos GGUF farão preload via arquivo bundle, ou serão baixados na primeira instalação por meio do instalador (`cmd/vectora-installer`) localizados no `%AppData%/.vectora/`?
> 2. **Chaves (API-Key/Criptografia):** Posso assumir a utilização da lib `zalando/go-keyring` para persistir dados confidenciais nos keyrings oficiais dos SOs (Windows Credential Manager / macOS Keychain / Linux Secret Service)?
> 3. **Protocolo Wails:** Como o Wails já gera um IPC próprio do browser ao AppDesktop por baixo dos panos (bindings diretos JS<->Go), nós injetaremos explicitamente a passagem de mensagens para o `internal/ipc/client.go` Go-para-Go da interface do projeto Wails até o binário do Daemon `cmd/vectora`, correto? Ou seja -> (Next.js --_WailsIPC_--> `cmd/vectora-web` --_DomainSocket/NamedPipe_--> `cmd/vectora`).

---

## Verification Plan

### Automated Tests

1. **Regra dos 300% (Go Tests)**: Utilização agressiva da suíte padrão de testes focando explicitamente: The Happy Path, The Failure Path, and the Edge Constraints.
2. Comandos cruciais: `make test`, `make test-integration` e `make test-race` (concorrência e deadlocks do acesso bbolt aos arquivos).

### Manual Verification

1. Lançar no Windows o executável principal gerado em `cmd/vectora`, acompanhando via Gerenciador de Tarefas se sua pegada de RAM não excederá os sub-critérios de ~5-15MB previstos na inatividade (Daemon Idle Mode).
2. Lançar `cmd/vectora-web`, provar que emulando uma desconexão ou fechamento do Web UI no gerenciador de Tarefas, o processo Systray mantém todo o chat histórico perfeitamente integro.
3. Testar a proteção _GitBridge_: invocar uma tool via CLI com um LLM para destruir intencionalmente um arquivo mock de texto e em seguida rodar `.Restore()` local para documentar a prova de sobrevivência dos blocos em rollback.
