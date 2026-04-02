# Revisão Arquitetural Sistêmica: Ecossistema de APIs Vectora

Após a implementação basal dos Sockets e Módulos Estruturais, detectamos interdependências profundas que requerem orquestração impecável. Este plano definitivo formaliza as regras de tráfego de dados cruzados entre todo o Backend (Daemon/Tray -> LLMs -> Bancos BD -> Ferramentas).

## User Review Required

> [!IMPORTANT]
>
> - O Wails IPC (_Web UI_) irá abstrair chamadas do React e enviá-las num pipeline nativo do Go diretamente para o socket IPC Socket UNIX do Daemon. Assim não abrimos portas HTTP na máquina do usuário. Valide se concorda.
> - O MCP Server roda isoladamente quando invocado pelo VSCode/Cursor (`vectora.exe --mcp`). Ele deverá renegociar acesso seguro ao Chromem-Go independentemente, mas não instanciará _llama.cpp_ para evitar conflitos de GPU com o Daemon.

---

## Revisão de Fluxo Integrado (Como os módulos se tocam)

### 1. Comunicação Wails Front-End ↔ RPC IPC Daemon (`vectora-web` -> `internal/ipc`)

- O Next.js/React fará chamada assíncrona p/ Wails: `window.go.main.App.WorkspaceQuery("id", "pergunta")`.
- A Go Struct do Wails agirá puramente como um Proxy (Router): Constrói a mensagem ND-JSON e emite no `client.Send("workspace.query")`.
- O Daemon `vectora.exe` principal (Unix Socket Master) recebe o pulso, invoca o pipeline RAG, aciona evento de Stream pro UI e devolve a resposta final.

### 2. Ciclo Sidecar (`tray` ↔ `internal/llm/qwen` ↔ `llama-server.exe`)

- Quando o usuário no Systray escolhe o modo **Qwen GGUF Offline**:
  1. A `tray` inicializa a Factory `NewQwenProvider()`.
  2. O O.S. Manager entra em ação, busca o binário nativo na raiz `~/.Vectora/` e starta o Subprocesso local (42781 tcp).
  3. O provedor injeta um client `langchaingo` falsificando requisições OpenAI `v1/chat/completions` em apontamento direto para o localhost limpo.
  4. Quando o Systray é finalizado (Quit) ou há troca de motor (Gemini), a interface atesta limpeza matando a Thread OS do `llama-server`.

### 3. Workflow de I/O em LangChain (`internal/llm` ↔ `internal/acp` ↔ `internal/tools`)

- Todas as Tool calls descritas no schema do Langchaingo devem realizar unmarshal nativo nos structs do Go.
- Quando a LLM solicita acesso (Ex: `write_file`):
  1. O `internal/acp` (Agent Context) invoca a Tool através do "ExecuteStringArgs".
  2. A própria ferramenta instigada executa antes um _Backup_ automático raw gerando SnapID em `%USERPROFILE%/.Vectora/backups`.
  3. Só então o FS nativo é tocado via O.S. (Garantindo Regra Zero-Risco de modificação sem reverter).

### 4. Persistência de Contexto Duplo (`internal/db`)

- **Memória Linear (`bbolt`)**: Persiste a estrutura de preferência do usuário (Keys, Workspaces Meta) e Histórico de Chat. Injetado na Tool `save_memory` via AgentContext pra dar lembrança pro LLM.
- **RAG Geométrico (`chromem-go`)**: Atuará de mãos dadas com a Action `workspace.query`. Assim que um Dataset é indexado/baixado pelo Cliente HTTP `internal/index`, ele vai pra gaveta vetorial com Namespaces Isolados. A função de Embebing do Langchaingo mapea todo documento em "Chunks", que mergulharam individualmente encapsulados no Struct `db.Chunk`.

---

## Proximas Frentes de Adoção (Checklist Arquitetural)

Para solidificar os links que acabam de nascer nos pacotes desacoplados da `/internal`:

#### [MODIFY] `internal/acp/agent.go`

- Injetar o repositório DB físico (BBoltStore) no inicializador do Agent Context, permitindo que a _Tool_ `save_memory` funcione perfeitamente.

#### [NEW] `internal/core/rag_pipeline.go`

- Será a Placa Mãe Real! Fiará as conexões em tempo real do DB Vectorial (Chromem) -> Chamada ao Prompt -> Invoca a LLM (Gemini/Qwen) -> Ativa o Evento do IPC pra desenhar na tela Web.

## Open Questions

> [!WARNING]
> Dado que o Wails é o empacotador de Janelas do Windows/Mac, a única dependência que o instalador vai colocar além do Daemon (`vectora.exe`) na raiz do `%USERPROFILE%/.Vectora` é o binário principal `vectora-web.exe`. Ou você prefere fundir o Wails e o Daemon no MESMO código binário unificado usando as Views Nativas do Wails invés de processos IPC independentes? (Recomendo fortemente IPC Separado pra não "sujar" o Daemon Systray com WebView memory footprint).

## Verification Plan

1. Revisaremos o Mockup Pipeline do Core injetando Chunks artificais na camada `internal/db` para aferir busca em menos de 100ms.
2. Validar que o Wails Router dispara perfeitamente no Socket Linux/Win32 via IPC Ping.
