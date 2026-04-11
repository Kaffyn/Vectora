# Vectora Issue Report - Bug Collection & Requests

Este documento consolida as falhas, decisões arquiteturais e requisitos estratégicos do ecossistema Vectora.

---

## 🐞 Reports (Bugs)

### 1. Falha no Carregamento da Webview (vectora.chatView)

**Status**: Identificado
**Descrição**: Erro "Ocorreu um erro ao carregar o modo de exibição". Independente do estado do Core.
**Causa Provável**: Registro tardio do provedor no ciclo de vida da extensão.

### 2. Falha na Inicialização Automática do Core

**Status**: Identificado
**Descrição**: A extensão sinaliza prontidão, mas o processo de background (`vectora start`) não é iniciado.
**Causa Provável**: Divergência entre o diretório de instalação do build e o `BinaryManager`.

### 3. Concorrência de Processos (Múltiplas Instâncias)

**Status**: Identificado
**Descrição**: O Core permite a execução de múltiplas instâncias simultâneas, causando conflitos de porta e UI.
**Causa Provável**: Ausência de controle de Singleton (Lock/Socket).

### 4. Resíduos de Processos e Binários Duplicados

**Status**: Identificado
**Descrição**: Convivência de processos manuais e automáticos com nomes distintos (`vectora.exe` vs `vectora-windows-amd64.exe`).

### 5. Lacunas na CLI (`config set`)

**Status**: Identificado
**Descrição**: Falta de clareza sobre chaves válidas e formatos de configuração.

### 6. Opacidade no Comando `workspace ls`

**Status**: Identificado
**Descrição**: Exibição apenas de hashes (IDs) sem o caminho físico (`path`) correspondente.

### 7. Comandos no Plural

**Status**: Identificado
**Descrição**: Comandos comuns como `workspaces` não possuem aliases.

### 8. Bloqueio por Antivírus (Windows Defender)

**Status**: Identificado/Contornado
**Descrição**: O binário é falsamente detectado como Trojan após operações de `start`.

### 9. Erro 404 Gemini (Modelo Inválido)

**Status**: Identificado
**Descrição**: Uso de identificadores inexistentes como `gemini-3-flash`.

---

## ❓ Questions (Discussões Arquiteturais Críticas)

### 10. Método de Singleton no Core

**Decisão**: **Abordagem Híbrida (File Lock + PID Validation)**.

- O Daemon tenta criar o arquivo `.vectora.lock`. Se existir, valida se o PID gravado ainda está ativo no SO.
- Resolve problemas de _Race Condition_ do TCP e _Orphaned Locks_ de crashes.

### 11. Estratégia de Fallback LLM

**Decisão**: **Migração 100% para SDKs oficiais**.

- Eliminar fallbacks HTTP manuais para reduzir complexidade. Confiar nos SDKs e tratar erros de rede de forma genérica.

### 12. Gerenciamento de Memória em Long-Running Daemons

**Questão**: Como lidar com potencial vazamento de memória em sessões longas ao usar SDKs pesados?

- **Opção A**: "Soft Restart" do worker de LLM a cada X tokens.
- **Opção B**: Confiar no GC do Go e monitorar via pprof local.

### 13. Estratégia de Atualização de Binários (Windows)

**Questão**: Como o Daemon deve se atualizar (`vectora.exe`) visto que o Windows bloqueia a sobrescrita de arquivos em execução?

- **Recomendação**: **Opção B (Processo Auxiliar Updater)**. Um updater mata o Daemon, substitui o binário e reinicia com rollback automático em caso de falha.

### 14. Isolamento de Contexto (Workspaces Privados vs Públicos)

**Questão**: Como garantir que workspaces "Privados" não vazem metadados/hashes estruturais durante sincronização?

- **Decisão sugerida**: Uso de **Salting** nos hashes antes de enviar checksums para o servidor de Index.

### 15. Tratamento de Erros em SDKs Assíncronos (Streaming)

**Questão**: Como padronizar o tratamento de erros parciais em streams do Gemini/Claude?

- **Opção A**: Reconnect automático transparente pelo SDK.
- **Opção B**: Daemon intercepta o erro, fecha o stream e solicita "Retry" na UI.

### 16. Segurança do Canal IPC (Named Pipes/Sockets)

**Questão**: Como impedir que outros processos do mesmo usuário local interceptem o canal IPC?

- **Decisão sugerida**: Implementar um **Handshake de Autenticação** (token gerado no startup e passado via env vars para os processos filhos).

### 17. Versionamento de Schema do Banco Vetorial (Chromem-go)

**Questão**: Como lidar com atualizações que mudam a dimensão do embedding ou formato do índice?

- **Decisão sugerida**: Detecção automática de versão do schema e trigger de re-indexação automática (lenta).

### 18. Observabilidade e Logs Sensíveis

**Questão**: Como evitar que logs de debug capturem payloads sensíveis do usuário?

- **Decisão sugerida**: Middleware de **sanitização de logs** que mascara conteúdos, mantendo apenas metadados estruturais.

### 19. Seleção de Bibliotecas JSON-RPC

**Decisão**: Padronização das libs para garantir conformidade JSON-RPC 2.0 e suporte a streaming.

- **Core (Go)**: `sourcegraph/jsonrpc2` (Padrão LSP, robusto para streams bidirecionais).
- **Extensions (TS)**: `vscode-jsonrpc` (Nativo Microsoft, integração perfeita com VS Code API).

---

## 🚀 Requests (Modernização e Requisitos)

### 20. Consolidação da Comunicação (IPC + JSON-RPC + SDKs)

**Status**: Requisito de Modernização
**Descrição**: Unificar toda a comunicação em **IPC + JSON-RPC** entre Core e Extensões (ACP/MCP). O SDK de cada provedor deve ser um método interno e privado do Core. Extensões e chat consomem apenas a nossa API unificada.

**SDKs Alvo (Chat & Embeddings)**:

- **Gemini**: [google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai)
- **Claude**: [github.com/anthropics/anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)
- **Voyage AI**: [github.com/austinfhunter/voyageai](https://pkg.go.dev/github.com/austinfhunter/voyageai)

### 21. Revisão de Modelos e Funcionalidades via Docs Oficiais

**Status**: Requisito de Modernização
**Descrição**: Revisar e alinhar identificadores de modelos e configurações (Thinking, Caching) com base nas documentações oficiais.

**Documentação de Referência**:

- **Gemini (Models & Thinking)**: [ai.google.dev/gemini-api/docs/models](https://ai.google.dev/gemini-api/docs/models?hl=pt-br)
- **Claude (Models & Caching)**: [platform.claude.com/docs/en/api/sdks/go](https://platform.claude.com/docs/en/api/sdks/go)
- **Voyage (Embedding Docs)**: [pkg.go.dev/github.com/austinfhunter/voyageai](https://pkg.go.dev/github.com/austinfhunter/voyageai)

### 22. Auditoria Geral de Security Patterns e Tools

**Status**: Requisito de Modernização
**Descrição**: Realizar uma auditoria completa nos padrões de segurança e ferramentas utilizadas, integrando as decisões tomadas nas Questions 10-19.

---

_Este relatório servirá como base para o planejamento detalhado da fase de execução._
