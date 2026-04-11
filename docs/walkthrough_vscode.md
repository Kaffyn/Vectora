# Walkthrough: Estabilização da Extensão VS Code (Vectora)

Este documento detalha a transformação técnica realizada na extensão do VS Code para purgar o código legado do Roo Code e estabelecer a base sólida do **Vectora**.

## 1. Purga de Código Legado (Remoção)

Para garantir que a extensão não carregasse lógica desnecessária ou proprietária do Roo, removemos as seguintes pastas e componentes:

- **Pastas Inteiras**:
  - `src/utils`, `src/services`, `src/shared`, `src/hooks`, `src/context`, `src/core`.
  - `src/components/marketplace`, `src/components/cloud`, `src/components/worktrees`, `src/components/modes`, `src/components/mcp`, `src/components/settings`.
- **Componentes de Chat**:
  - `ModeSelector.tsx`, `McpExecution.tsx`, `WorktreeSelector.tsx`, `UpdateTodoListToolBlock.tsx`, `TaskHeader.tsx`, `BatchListFilesPermission.tsx`.
- **Lógica Comercial/Roo-specific**:
  - `RooHero.tsx`, `RooTips.tsx`, `Announcement.tsx`, `CloudUpsellDialog.tsx`, `TelemetryBanner.tsx`.

## 2. Refatoração de Arquitetura (Modificações)

### Extension & IPC

- **Padrão ES6**: Refatoramos o `extension.ts` para utilizar imports modernos (ES6) e remover referências `require` legadas.
- **Sincronização IPC**: O `ChatViewProvider` foi atualizado para reconhecer as mensagens disparadas pela UI do Roo (`newTask`, `askResponse`, `cancelTask`, `clearTask`) e mapeá-las corretamente para as rotas do Core (`session/new`, `session/prompt`, `session/cancel`).

### UI (ChatRow & ChatView)

- **Descomentação Total**: Após estabilizar os tipos, descomentamos 100% dos blocos de lógica visual no `ChatRow.tsx` e `ChatView.tsx`.
- **Escopos de Bloco**: Corrigimos centenas de erros `no-case-declarations` envolvendo variáveis léxicas dentro de blocos `switch/case` usando chaves `{}`.
- **Landing Page**: Substituímos a tela inicial do Roo Code por uma tela de boas-vindas limpa do **Vectora AI v0.1.0**.

## 3. Adições e Restaurações (Vectora-Ready)

Como a UI dependia de certas funções e componentes para não quebrar, recriamos versões **simplificadas e independentes**:

- **Componentes de UI**:
  - `Mention.tsx`: Renderização simples de menções.
  - `FollowUpSuggest.tsx`: Sugestões de ação via IPC.
  - `BatchFilePermission.tsx`: Interface de aprovação de leitura/escrita de arquivos.
  - `BatchDiffApproval.tsx`: Interface de revisão de alterações de código.
- **Utilitários (src/utils)**:
  - `useDebounceEffect.ts`, `imageUtils.ts`, `costFormatting.ts`, `batchConsecutive.ts`.

## 4. Estado de Qualidade (Linting)

- **Configuração**: Modernizamos o `.eslintrc.json` para suportar React 18 e desativar regras obsoletas (`react-in-jsx-scope`).
- **Resultado**: Reduzimos o contador de erros de **+3000** para **Zero Erros**.
- **Automação**: O hook `VS Code Extension Lint` no `pre-commit` agora garante que nenhum erro de regressão seja introduzido.

---

**Status Final**: A extensão é agora um shell visual premium, leve, estável e 100% integrado ao Core via JSON-RPC.
