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
**Descrição**: Uso de identificadores inexistentes ou obsoletos como `gemini-3-flash`. Devem ser usados os identificadores de 2026 como `gemini-3.1-pro-preview`.

---

## ❓ Questions (Discussões Arquiteturais Críticas)

### 10. Método de Singleton no Core

**Decisão: Abordagem Híbrida (File Lock + PID Validation)**.

- **Implementação:** O Daemon tenta criar o arquivo `.vectora.lock`. Se existir, valida se o PID gravado ainda está ativo via SO. Resolve o problema de sockets presos no Windows e locks órfãos.

### 11. Estratégia de Fallback LLM

**Decisão: Migração 100% para SDKs oficiais**.

- **Justificativa:** Os SDKs são mais estáveis e mantidos. Manter fallback HTTP manual duplicaria a complexidade e o risco de bugs.

### 12. Gerenciamento de Memória em Long-Running Daemons

**Decisão: Opção B (Monitoramento via `pprof` + GC Agressivo)**.

- **Justificativa:** O "Soft Restart" interromperia o fluxo do usuário e perderia o contexto da conversa.
- **Execução:** Confiar no GC do Go, limpar buffers explicitamente após cada resposta e expor `pprof` localmente para auditoria.

### 13. Estratégia de Atualização de Binários (Windows)

**Decisão: Processo Auxiliar Updater com Rollback Automático**.

- **Execução:** O Daemon baixa o novo `.exe`, valida hash e spawna um `updater.exe` independente. O `updater` aguarda o fim do Daemon, substitui o binário e o reinicia.
- **Rollback:** Se a nova versão falhar no health check em 10 segundos, o `updater` restaura a versão estável anterior.

### 14. Isolamento de Contexto (Workspaces Privados vs Públicos)

**Decisão: Uso de Salting** nos hashes antes de enviar checksums para o servidor de Index, impedindo o vazamento de metadados estruturais de workspaces privados.

### 15. Tratamento de Erros em SDKs Assíncronos (Streaming)

**Questão:** Como padronizar o tratamento de erros parciais em streams do Gemini/Claude?

- **Opção A:** Reconnect automático transparente pelo SDK.
- **Opção B:** Daemon intercepta o erro, fecha o stream e solicita "Retry" na UI.

### 16. Segurança do Canal IPC (Named Pipes/Sockets)

**Decisão: Handshake de Autenticação**.

- **Execução:** Token gerado no startup e passado via env vars para os processos filhos, validando que apenas UIs legítimas conectem ao canal IPC.

### 17. Versionamento de Schema do Banco Vetorial (Chromem-go)

**Decisão: Re-indexação Automática (Lenta) com Aviso ao Usuário**.

- **Justificativa:** Evita "alucinações silenciosas" causadas por índices incompatíveis.
- **Execução:** Detecta mismatch no startup, marca o workspace como "Rebuilding..." na UI e processa o re-embedding em background com baixa prioridade.

### 18. Observabilidade e Logs Sensíveis

**Decisão: Middleware de Sanitização de Logs**.

- **Execução:** Mascaramento automático de strings que pareçam conteúdo de usuário, mantendo apenas metadados técnicos estruturais.

### 19. Seleção de Bibliotecas JSON-RPC

**Decisão: Padronização das libs para conformidade JSON-RPC 2.0 e streaming**.

- **Core (Go):** `sourcegraph/jsonrpc2` (Robustez em streams bidirecionais).
- **Extensions (TS):** `vscode-jsonrpc` (Integração nativa VS Code).

---

## 🚀 Requests (Modernização e Requisitos)

### 20. Consolidação da Comunicação (IPC + JSON-RPC + SDKs)

**Status**: Requisito de Modernização
**Descrição**: Unificar a comunicação em IPC + JSON-RPC entre Core e Extensões. O SDK de cada provedor deve ser um método interno e privado do Core.

**SDKs Alvo (Chat & Embeddings):**

- **Gemini:** [google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai)
- **Claude:** [github.com/anthropics/anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)
- **OpenAI:** [github.com/openai/openai-go](https://github.com/openai/openai-go)
- **Voyage AI:** [github.com/austinfhunter/voyageai](https://pkg.go.dev/github.com/austinfhunter/voyageai)

### 21. Revisão de Modelos e Funcionalidades via Docs Oficiais

**Status**: Requisito de Modernização
**Descrição**: Revisar e alinhar identificadores de modelos e configurações (Thinking, Caching) com base nas documentações oficiais.

- **Gemini (Models & Thinking)**: [ai.google.dev/gemini-api/docs/models](https://ai.google.dev/gemini-api/docs/models?hl=pt-br)
- **Claude (Models & Caching)**: [platform.claude.com/docs/en/api/sdks/go](https://platform.claude.com/docs/en/api/sdks/go)
- **OpenAI (GPT-5.4 family)**: [platform.openai.com/docs/models](https://platform.openai.com/docs/models)
- **Qwen / Alibaba (Qwen 3.6 & Embeddings v4)**: [alibabacloud.com/help/en/model-studio](https://www.alibabacloud.com/help/en/model-studio/)
- **Voyage (Embedding Docs)**: [pkg.go.dev/github.com/austinfhunter/voyageai](https://pkg.go.dev/github.com/austinfhunter/voyageai)

### 22. Auditoria Geral de Security Patterns e Tools

**Status**: Requisito de Modernização
**Descrição**: Auditoria completa nos padrões de segurança e ferramentas, integrando as decisões 10-19.

### 23. Padronização de Protocolos e Integração de SDKs

**Status**: Requisito de Implementação
**Descrição**: Efetivar a migração integral das rotinas de inferência e comunicação para os SDKs oficiais (Anthropic, Gemini, Voyage, OpenAI), utilizando as definições de `vectora-protocol-sdks.md`.

### 24. Multi-Tenancy Protocol (MTP)

**Status**: Requisito de Arquitetura
**Descrição**: Implementar isolamento lógico rigoroso no nível do Core (Singleton Daemon). Cada conexão IPC deve atrelar seu respectivo *Tenant* a instâncias isoladas de banco vetorial (Chromem-go), históricos locais, fila de requisições restritas (Semaphores) e limites rígidos de sistema de arquivos via Guardian (Trust Folders), protegendo totalmente o contexto ativo de operações externas cruzadas de outros projetos em andamento na máquina. Detalhes em `multi-tenant.md`.

### 25. Gateway & Aggregator Support (OpenRouter/Anannas)

**Status**: Requisito de Implementação
**Descrição**: Integrar suporte nativo para gateways como OpenRouter e Anannas via SDK oficial da OpenAI, permitindo testar múltiplos provedores (Claude, Gemini, Qwen) com uma única arquitetura unificada e chaves de teste.

---

_Este relatório é a especificação técnica final e aprovada para a fase de implementação._
