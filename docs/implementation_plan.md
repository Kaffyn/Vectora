# Plano de Implementação - Vectora Core (v1.2)

**Objetivo:** Consolidar o Vectora como um motor RAG híbrido, local-first, com arquitetura de API modular, políticas de segurança executáveis e observabilidade robusta. Este documento serve como o "Mapa Mestre" do projeto, interligando todos os blueprints técnicos.

## 1. Resolução das Open Questions

Para eliminar ambiguidades no desenvolvimento, definimos as seguintes estratégias arquiteturais:

### Q1: Estratégia do GitBridge para Mutações

**Decisão:** **Commits Atômicos com Tag de Snapshot.**

- **Por que não Stash?** `git stash` é volátil e difícil de gerenciar programaticamente em sessões concorrentes.
- **Por que não Branch?** Criar branches para cada edição gera overhead excessivo de gerenciamento de refs e "sujeira" visual no repositório do usuário.
- **A Solução:** O Vectora executará `git add <file>` seguido de `git commit -m "chore(vectora): snapshot pre-edit [TIMESTAMP]"`.
  - **Vantagem:** Cria um ponto de reversão claro e histórico auditável.
  - **Cleanup:** Opcionalmente, o Vectora pode oferecer uma ferramenta `cleanup_snapshots` que faz um `git reset --soft` para o estado anterior à sessão, mantendo o working directory limpo se o usuário aprovar as mudanças finais.
  - **Segurança:** Se o usuário não tiver commits anteriores, o Vectora avisa que não pode criar um ponto de restauração seguro e pede confirmação explícita.

### Q2: Canal de Comunicação MCP

**Decisão:** **Stdio Primário, TCP Secundário (Dev Mode).**

- **MCP Padrão:** A especificação MCP exige `stdio` para integrações locais seguras (como Claude Desktop ou VS Code). Este será o foco principal da v1.0.
- **TCP (Opcional):** Uma flag `--tcp-port` será adicionada ao daemon para permitir depuração ou conexões remotas controladas, mas não será o padrão para plugins de IDE por questões de segurança e complexidade de firewall.

---

## 2. Índice Mestre de Documentação Técnica

Este plano consolida e referencia todos os documentos de arquitetura e implementação definidos até o momento. Cada link abaixo representa um contrato técnico fechado e pronto para codificação.

| Documento                                            | Módulo            | Descrição Técnica                                                                                                                                            | Status          |
| :--------------------------------------------------- | :---------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------- | :-------------- |
| **[POLICIES.md](POLICIES.md)**                       | `core/policies/`  | Motor de Políticas Executáveis. Define regras imutáveis de escopo (Trust Folder), Guardian (bloqueio de binários/secrets) e prioridade de RAG.               | ✅ Concluído    |
| **[Storage.md](Storage.md)**                         | `core/storage/`   | Arquitetura de Persistência Dual-Store. Uso de **BBolt** para metadados/estado e **Chromem-go** para vetores. Inclui lógica de chunking com overlap.         | ✅ Concluído    |
| **[Tools_Executors.md](Tools_Executors.md)**         | `core/tools/`     | Implementação dos executores agênticos (`read_file`, `grep_search`, `terminal_run`). Foca em segurança, truncagem de output e portabilidade Go-native.       | ✅ Concluído    |
| **[LLM_Gateway.md](LLM_Gateway.md)**                 | `core/llm/`       | Orquestração de Modelos. Interface unificada `LLMProvider`, gerenciamento de janela de contexto (truncagem agressiva) e fábrica de System Prompts dinâmicos. | ✅ Concluído    |
| **[Ingestion.md](Ingestion.md)**                     | `core/ingestion/` | Pipeline de Indexação On-Demand. Parser selector, extração de grafo de dependências via Regex e integração com o Storage Engine.                             | ✅ Concluído    |
| **[Config_Manager.md](Config_Manager.md)**           | `core/config/`    | Gerenciamento de Estado e Segredos. Isolamento de Workspaces via Hash de Path e criptografia AES-GCM para API Keys no `config.yaml`.                         | ✅ Concluído    |
| **[Observabilidade.md](Observabilidade.md)**         | `core/telemetry/` | Sistema de Logging "Black Box". Implementação de `slog` com writer rotativo personalizado para evitar crescimento infinito de logs.                          | ✅ Concluído    |
| **[API.md](API.md)**                                 | `core/api/`       | Contratos de Multi-Protocolo. Especificação de JSON-RPC (MCP/ACP), gRPC (Streaming) e IPC (UI Interna).                                                      | ✅ Concluído    |
| **[Agentic_Tools.md](Agentic_Tools.md)**             | `internal/tools/` | Visão conceitual do Toolkit Agêntico. Define a dinâmica Tier 1 (Agente) vs Tier 2 (Sub-Agente) e a integração com IDEs.                                      | ✅ Concluído    |
| **[Implementation_Plan.md](Implementation_Plan.md)** | Raiz              | Este documento. Serve como roteiro unificado e checklist de integração.                                                                                      | 🔄 Em Progresso |

_(Nota: `golang-timeline.md` refere-se ao cronograma de desenvolvimento e milestones, que é derivado deste plano.)_

---

## 3. Estrutura Física do Motor de Políticas (Fase 2.1)

Já iniciada, esta fase garante que as regras sejam código, não documentação.

### Arquivos Chave:

1.  **`core/policies/rules/01-scope-guardian.yaml`**: Define extensões bloqueadas, symlinks inseguros e limites de diretório.
2.  **`core/policies/rules/02-git-passive.yaml`**: Define comportamento passivo do Git e estratégia de snapshots atômicos.
3.  **`core/policies/rules/03-rag-priority.yaml`**: Define pesos de contexto, fallbacks de busca e sanitização de output.
4.  **`core/policies/schema.go`**: Structs Go com tags `yaml` para mapeamento das regras.
5.  **`core/policies/loader.go`**: Usa `//go:embed rules/*.yaml` para carregar as regras no binário final, garantindo integridade.

---

## 4. Camada de API Multi-Transporte (Fase 3)

Esta é a espinha dorsal da comunicação. A arquitetura segue o princípio de **"Handlers Atômicos"** e **"Roteamento Agnóstico"**.

### 4.1. Documento Mestre: `core/api/API.md`

Este arquivo serve como a "Single Source of Truth" para desenvolvedores de clientes (IDEs, CLIs). Ele detalha os contratos JSON-RPC, definições `.proto` para gRPC e mensagens IPC.

### 4.2. Estrutura de Diretórios e Implementação Go

A estrutura abaixo evita arquivos monolíticos. Cada handler é responsável por uma única ação ou grupo coeso.

```text
core/api/
├── API.md                  # Contratos documentados
├── router.go               # Interface comum para todos os protocolos
├── grpc/
│   ├── server.go           # Setup do listener gRPC
│   ├── proto/
│   │   └── vectora.proto   # Definição .proto
│   └── handlers/
│       ├── query_handler.go    # Implementa stream de RAG
│       └── index_handler.go    # Implementa stream de Indexação
├── jsonrpc/
│   ├── server.go           # Loop de leitura stdio/tcp
│   └── methods/
│       ├── init_method.go      # Handler para 'initialize'
│       ├── tools_list_method.go# Handler para 'tools/list'
│       └── tools_call_method.go# Handler para 'tools/call'
├── ipc/
│   ├── server.go           # Listener de Named Pipes/Unix Socket
│   └── events/
│       ├── status_event.go     # Struct e serializer de Status
│       └── auth_event.go       # Struct e serializer de Auth Request
└── shared/
    └── middleware.go       # Middleware comum (Logging, Policy Check)
```

### 4.3. Exemplo de Código Modular

#### A. O Roteador Abstrato (`core/api/router.go`)

```go
package api

import (
    "context"
    "vectora/core/engine" // O cérebro real
)

// Router interface define o contrato mínimo que qualquer protocolo deve atender
type Router interface {
    Start(ctx context.Context) error
    Stop() error
}

// CoreDeps injeta as dependências necessárias para os handlers
type CoreDeps struct {
    Engine *engine.Engine
    // Outras deps comuns (Logger, PolicyEngine)
}
```

#### B. Handler JSON-RPC Atômico (`core/api/jsonrpc/methods/tools_call_method.go`)

```go
package methods

import (
    "context"
    "encoding/json"
    "vectora/core/api/shared"
    "vectora/core/engine"
)

// HandleToolsCall é uma função pura que recebe params brutos e retorna resultado
func HandleToolsCall(ctx context.Context, deps *shared.CoreDeps, params json.RawMessage) (interface{}, error) {
    var req engine.ToolCallRequest
    if err := json.Unmarshal(params, &req); err != nil {
        return nil, err
    }

    // Chama o engine centralizado
    result, err := deps.Engine.ExecuteTool(ctx, req)
    if err != nil {
        return nil, err
    }

    // Formata resposta padrão JSON-RPC
    return map[string]interface{}{
        "content": []map[string]string{{"type": "text", "text": result.Output}},
        "isError": result.Error != nil,
    }, nil
}
```

#### C. Handler gRPC Específico (`core/api/grpc/handlers/query_handler.go`)

```go
package handlers

import (
    "context"
    pb "vectora/core/api/grpc/proto"
    "vectora/core/engine"
)

type QueryHandler struct {
    pb.UnimplementedVectoraServiceServer
    Engine *engine.Engine
}

func (h *QueryHandler) Query(req *pb.QueryRequest, stream pb.VectoraService_QueryServer) error {
    // Usa o mesmo engine centralizado, mas em modo stream
    resultStream, err := h.Engine.StreamQuery(stream.Context(), req.Query)
    if err != nil {
        return err
    }

    for chunk := range resultStream {
        if err := stream.Send(&pb.QueryResponse{
            Token: chunk.Token,
            SourceRef: chunk.SourceRef,
        }); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 5. Próximos Passos Imediatos (Action Plan)

Com toda a arquitetura definida e documentada, o foco agora é a integração vertical:

1.  **Implementar `core/policies/loader.go`**: Garantir que os YAMLs sejam carregados via `go:embed` e validados na inicialização.
2.  **Criar `core/api/` Skeleton**: Estabelecer as pastas e arquivos vazios conforme a estrutura acima.
3.  **Desenvolver `jsonrpc/server.go`**: Implementar o loop básico de leitura de `stdin` para suportar o protocolo MCP inicial.
4.  **Integrar `tools/call`**: Conectar o handler JSON-RPC ao `engine.ExecuteTool` (que orquestra as tools definidas em `Tools_Executors.md`).
5.  **Teste de Ponta a Ponta**: Rodar o daemon em modo MCP e conectar via Claude Desktop ou Cursor para validar o ciclo completo: _Pergunta -> RAG -> Tool Call -> Resposta_.

Este plano representa a conclusão da fase de arquitetura. O Vectora Core está pronto para ser construído linha por linha seguindo estes blueprints.
