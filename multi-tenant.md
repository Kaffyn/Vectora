# 🏗️ Vectora Multi-Tenancy Protocol (MTP)

Este documento especifica a arquitetura e o protocolo para gerenciar **múltiplos projetos simultâneos (Multi-Tenancy)** em uma **única instância singleton** do Vectora Daemon.

O objetivo é manter o consumo de memória baixo (usando um único daemon rodando em background) enquanto se garante isolamento absoluto de estados de conversa, índices vetoriais de código, concorrência a provedores LLM, e rotinas de leitura de discos (Trust Folders).

---

## 1. O Modelo de Tenant baseado em Conexão (Connection-Bound Tenancy)

No Vectora, um "Tenant" representa um Projeto ou Workspace aberto no editor (como uma janela do VS Code). Quando a extensão do VS Code se conecta ao Daemon via IPC (Named Pipes ou Unix Sockets), o protocolo MTP estabelece o escopo de atuação daquela conexão.

### A. Handshake de Autenticação e Contexto (IPC)

A mensagem `ipc.auth` ou uma nova mensagem `workspace.init` serve para estabelecer o contexto do Tenant de forma contínua durante essa conexão.

**Request (Client -> Daemon):**

```json
{
  "type": "request",
  "method": "workspace.init",
  "payload": {
    "workspace_root": "C:\\Users\\bruno\\Projects\\MyApp",
    "project_name": "MyApp"
  }
}
```

O Daemon gera um **WorkspaceID** previsível e consistente baseado no path absoluto (e.g. `sha256(workspace_root)`).
A partir desse momento, todas as trocas de mensagens na conexão daquele _socket_ estarão amarradas a este isolamento lógico.

---

## 2. Abstração de Armazenamento e Estado (Storage & State Isolation)

Para garantir que o Projeto A nunca sobrescreva os dados do Projeto B, toda persistência e gerenciamento em memória é dinamicamente enclausurado.

### A. Estrutura de Pastas de Dados

Os dados no `%APPDATA%\Vectora\` (ou Unix equivalente) serão reestruturados:

```text
%APPDATA%\Vectora\
├── global.db                  # BBolt: Configs globais do Daemon e chaves API
├── ipc.token                  # Token de segurança local
└── workspaces\                # NÓS DO MULTI-TENANCY
    ├── <workspace_id_A>\      # Namespace fixo para o Projeto A
    │   ├── chromadb\          # Índice vetorial (Chromem-go) local isolado
    │   ├── chat_history.db    # Histórico de agentes e sessões
    │   └── guardian.json      # Regras de limite de acesso (Trust Folder)
    └── <workspace_id_B>\      # Namespace fixo para o Projeto B
        ├── chromadb\
        ├── chat_history.db
        └── guardian.json
```

### B. Gestor de Ciclo de Vida dos Workspaces em Memória

O Daemon manterá um registro Singleton de "Workspaces Ativos".

- Quando a primeira query `workspace.init` para `<id_A>` chega, o Daemon carega a coleção do `chromem-go` A em memória.
- Uma política de **Eviction** descarregará o Workspace B da memória RAM após 30 minutos sem nenhuma conexão IPC ativa solicitando-o. Isso economiza RAM.

---

## 3. Segurança e Limites Absolutos (Trust Folders & Guardian)

O conceito mais perigoso em um Singleton é um agente num Projeto A ser comprometido e tentar ler as variáveis `.env` do Projeto B do usuário. Para evitar isso:

1. **Restrição por Contexto IPC:** A rotina do IPC nunca permite repassar `<workspace_id_B>` com a conexão originada e autorizada no `<workspace_id_A>`. O contexto `context.Context` passado aos Handlers já contém a Root de Segurança fixada.
2. **Guardian File Interceptor:** Qualquer acesso de File System feito pelos Handlers ou por chamadas da LLM vai passar por um `guardian.ValidatePath(requestedPath, tenantRoot)`. Se tentarem `../../../ProjetoSecreto/`, o daemon corta e retorna um `IPCError`.

---

## 4. Paralelismo Seguro e Pool de Recursos (Resource Throttling)

Como usamos requisições externas para LLMs (ou locações de memória na GPU local), o Projeto A pode gerar bloqueios ou exceder os Rate Limits de uma Cloud (Anthropic/Voyage), afetando o Projeto B.

### A. CPU-Bound Priority Queue (Indexadores)

- Indexação e Embedding será delegada a Background Workers de baixa prioridade.
- A Janela Ativa tem uma requisição Foreground impulsionada na pool. A interface IPC enviará sinais passivos para sinalizar quem está no "Foco" do OS.

### B. IO-Bound Rate Limiting Semaphore (LLMs Calls)

- O daemon terá Semáforos por **Workspace**. O limite padrão pode ser 2 requisições paralelas ativas _por projeto_. Se houver pico num arquivo, o projeto enfileira suas próprias chamadas, enchendo a própria cota, enquanto o _Projeto B_ no painel ao lado tem a própria cota intocável e fluida.

---

## 5. Fluxo da Arquitetura para Implementação

**Passo 1: `core/manager/tenant.go`**
Criar um construtor de tenants `GetOrCreateTenant(workspaceRoot)`. Este objeto encapsula suas próprias classes do DB, LLM Histories e Storage Engine (evitando passagem de parametros massiva nas rotinas).

**Passo 2: Injetar Tenant no Pipeline de Handlers IPC**
Modificar o Event Loop em `core/api/ipc/server.go`. Quando ler uma mensagem conectada num descritor, ele insere o respectivo `*Tenant` no Context usando `context.WithValue(ctx, TenantKey, activeTenant)`. Assim os Handlers só batem na DB daquele Context.

**Passo 3: Mapeamento Dinâmico de Database e Index**
Separar os diretórios de forma limpa. A inicialização do Chromem-Go deixará de ser Global para ser mapeada no `GetOrCreateTenant()`.

**Passo 4: Monitor de Auto-Desligamento (Eviction)**
Rotina em background (Ticker de x mins) que fecha e consolida/salva estados de um Tenant persistido cujo socket de interface (Janela do editor) não deu sinais de vida ou bateu em um timeout de socket.
