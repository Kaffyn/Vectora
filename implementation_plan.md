# Relatório de Problemas do Vectora - Plano de Implementação

## Contexto

O projeto Vectora possui 9 bugs, 10 decisões arquiteturais que precisam de implementação e 3 solicitações de modernização. Os principais problemas são: identificadores de modelos quebrados causando 404s, problemas no registro da webview, todos os provedores de LLM usando HTTP bruto em vez de SDKs oficiais, implementação manual de JSON-RPC e falta de polimento na CLI. Este plano aborda todos os 22 itens do relatório de problemas em ordem de dependência.

---

## Fase 0: Correções de Bugs Críticos (Imediato)

### 0A. Corrigir Identificadores de Modelo do Gemini (Issue #9)

- **Arquivo:** `core/llm/gemini_provider.go:31-37`
- Atual: `"gemini-3-flash"`, `"gemini-3.1-pro"`, `"gemini-embedding-2-preview"` - estes modelos não existem.
- Correção: Atualizar para IDs de modelos reais da documentação do Google (ex: `"gemini-2.0-flash"`, `"gemini-2.5-pro"`, `"text-embedding-004"`).
- Também corrigir os aliases do Claude em `core/llm/claude_provider.go:76-86` - os aliases "4.6" apontam para IDs de modelos desatualizados.

### 0B. Corrigir Falha no Carregamento da Webview (Issue #1)

- **Arquivo:** `extensions/vscode/src/extension.ts:25`
- Problema: `new ChatViewProvider(undefined as any, context)` passa um cliente nulo.
- Correção: O `ChatViewProvider` deve lidar com o cliente nulo de forma graciosa - exibir estado de "Conectando..." em vez de travar.
- **Arquivo:** `extensions/vscode/src/chat-panel.ts` - adicionar proteção contra cliente nulo nos manipuladores de mensagens.

### 0C. Corrigir Inconsistência na Nomeação do Binário (Issues #2, #4)

- **Arquivo:** `extensions/vscode/src/binary-manager.ts`
- Problema: Procura por `vectora.exe`, mas o build produz `vectora-windows-amd64.exe`.
- Correção: Padronizar para `vectora.exe` na saída do build OU adicionar ambos os nomes à cadeia de resolução.
- Também limpar o comando `stop` para encerrar processos que correspondam a ambos os nomes.

---

## Fase 1: UX da CLI (Ganhos Rápidos)

### 1A. Validação de Chaves de Configuração (Issue #5)

- **Arquivo:** `cmd/core/config.go`
- Adicionar lista branca de chaves válidas + texto de ajuda mostrando as chaves aceitas.

### 1B. Exibição do Caminho do Workspace (Issue #6)

- **Arquivo:** `cmd/core/workspace.go`
- Armazenar metadados do caminho junto com as coleções de workspace, exibir `ID → /caminho` em `workspace ls`.

### 1C. Aliases de Comandos (Issue #7)

- **Arquivo:** `cmd/core/main.go` (registro de comandos)
- Adicionar `Aliases` do Cobra: `workspaceCmd.Aliases = []string{"workspaces", "ws"}`.

### 1D. Documentação do Windows Defender (Issue #8)

- Documentar o processo de assinatura de código em `CONTRIBUTING.md` ou nos documentos de release (não é uma correção de código).

---

## Fase 2: Singleton e Gerenciamento de Processos (Decisões #10, Issues #3, #4)

### 2A. File Lock Híbrido + Validação de PID

- **Arquivos:** `core/os/linux/linux.go`, `core/os/macos/macos.go` (substituir o bind de porta TCP).
- Novo multi-plataforma: Gravar PID em `~/.vectora/vectora.pid` + `flock()` no Unix.
- Manter o mutex do Windows como está (já funciona), adicionar arquivo de PID como suplementar.

### 2B. Encerramento Gracioso (Graceful Shutdown)

- **Arquivo:** `cmd/core/main.go` - manipuladores de sinais para limpar o arquivo de PID no SIGTERM/SIGINT.

---

## Fase 3: Migração da Biblioteca JSON-RPC (Decisão #19)

### 3A. Core (Go): Adotar `sourcegraph/jsonrpc2`

- **Arquivos:** `core/api/jsonrpc/`, `core/api/ipc/server.go`, `core/api/ipc/router.go`.
- Adicionar dependência, reescrever o registro de handlers para usar a interface de handler da biblioteca.
- Migrar método por método.

### 3B. Extensão VS Code: Adotar `vscode-jsonrpc`

- **Arquivo:** `extensions/vscode/src/client.ts`.
- Substituir o enquadramento (framing) manual por `createMessageConnection`.

### 3C. Handshake de Segurança do IPC (Decisão #16)

- Core: Gerar token na inicialização → `~/.vectora/ipc.token`.
- Extensão: Ler o token, enviá-lo na requisição `initialize`.
- Core: Rejeitar conexões sem um token válido.

---

## Fase 4: Migração para SDKs de LLM (Decisões #11, #20, #21)

### 4A. Gemini → `google.golang.org/genai`

- **Arquivo:** `core/llm/gemini_provider.go` - reescrita completa usando o SDK.
- Corrige validação de modelo, streaming e tratamento de erros em um único passo.

### 4B. Claude → `github.com/anthropics/anthropic-sdk-go`

- **Arquivo:** `core/llm/claude_provider.go` - reescrita completa usando o SDK.
- Atualizar identificadores de modelo para os valores atuais.

### 4C. Voyage → `github.com/austinfhunter/voyageai`

- **Arquivo:** `core/llm/voyage_provider.go` - reescrita usando o SDK.

### 4D. Tratamento de Erros em Streaming (Decisão #15)

- Implementar em cada provedor de SDK: em caso de erro no stream, enviar notificação de erro JSON-RPC com conteúdo parcial.
- A UI da extensão mostra "Resposta interrompida" com botão de tentar novamente.

---

## Fase 5: Observabilidade e Segurança (Decisões #12, #17, #18)

### 5A. Integração com pprof (Decisão #12)

- **Arquivo:** `cmd/core/main.go` - adicionar `net/http/pprof` na porta de debug do localhost.

### 5B. Sanitização de Logs (Decisão #18)

- Novo middleware em `core/infra/` - redigir (mascarar) chaves de API e informações de identificação pessoal (PII) dos logs.

### 5C. Versionamento de Schema do Vector DB (Decisão #17)

- Armazenar versão do schema no bbolt. Em caso de incompatibilidade → re-indexação automática com notificação ao usuário.

---

## Fase 6: Sistema de Atualização e Segurança (Decisões #13, #14, #22)

### 6A. Auto-Atualizador com Rollback (Decisão #13)

- Novo pacote: `core/updater/` - verificar releases no GitHub, baixar, trocar binário, realizar health check e rollback se necessário.

### 6B. Hashes Salteados (Salted) de Workspace (Decision #14)

- Sal (salt) por instalação em `~/.vectora/salt`, usar `SHA256(salt + path)` para IDs de workspace.

### 6C. Auditoria de Segurança (Decisão #22)

- Revisar todas as mudanças: autenticação IPC, sanitização de logs, aplicação do Guardian, verificações de travessia de caminho (path traversal).

---

## Gráfico de Dependências

```
Fase 0 ──┐
Fase 1 ──┼── (todas paralelas, sem dependências)
Fase 2 ──┘
           │
Fase 3 ───── (portão para a Fase 4: SDKs precisam de propagação de erro adequada)
           │
Fase 4 ───── (depende da Fase 3)
           │
Fase 5 ───── (depende da Fase 2 para a porta pprof)
Fase 6 ───── (depende da Fase 2 + Fase 5)
```

## Verificação

- **Fase 0:** `go build ./...` tem sucesso; a extensão carrega a webview sem erro; `vectora ask "test"` não retorna 404.
- **Fase 1:** `vectora workspace ls` mostra os caminhos; `vectora workspaces` funciona; `vectora config set INVALID x` avisa sobre chave inválida.
- **Fase 2:** Iniciar duas instâncias mostra "already running"; `vectora status` relata o estado correto.
- **Fase 3:** A extensão se conecta via `vscode-jsonrpc`; a autenticação de token IPC rejeita clientes não autorizados.
- **Fase 4:** `go test ./core/llm/...` passa com cada SDK; o streaming funciona de ponta a ponta.
- **Fase 5:** `curl localhost:<debug-port>/debug/pprof/` funciona; os logs não mostram chaves de API.
- **Fase 6:** Atualização de binário + rollback testados manualmente; os IDs de workspace diferem entre instalações.
