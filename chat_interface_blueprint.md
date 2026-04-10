# Blueprint: Vectora Next-Gen Chat Interface

Este blueprint define a arquitetura visual e funcional da interface de chat do Vectora, focada em densidade de informação, estética premium (Antigravity/Cursor) e transparência de processos em segundo plano.

## 1. Filosofia de Design: "The Invisible Engine"

A interface deve ser minimalista, utilizando variáveis nativas do VS Code para garantir que o foco permaneça no conteúdo.

- **Contraste Dinâmico**: Uso de `--vscode-descriptionForeground` para metadados e `--vscode-editor-foreground` para conteúdo principal.
- **Glassmorphism**: Uso de `backdrop-blur` em botões flutuantes e containers de sistema.
- **Micro-animações**: Transições suaves de entrada (`fade-in`, `slide-in-from-bottom`).

---

## 2. Componentes Antigravity (Copiados Literalmente)

Os componentes essenciais para a experiência "Antigravity" foram migrados integralmente para garantir fidelidade visual.

### A. Orquestração e Viewers

- **[ChatRow.tsx](file:///extensions/vscode/src/components/chat/ChatRow.tsx)**: O componente principal de cada mensagem. Gerencia estados complexos como pensamento (`ReasoningBlock`), execuções de ferramentas e placeholders.
- **[ChatView.tsx](file:///extensions/vscode/src/components/chat/ChatView.tsx)**: Container principal do chat, responsável pelo scroll inteligente e renderização da lista de mensagens filtradas.
- **[TerminalOutput.tsx](file:///extensions/vscode/src/components/chat/TerminalOutput.tsx)**: Emulador de terminal ANSI para exibir resultados de comandos internos no chat.

### B. Blocos de Processamento

- **[ReasoningBlock.tsx](file:///extensions/vscode/src/components/chat/ReasoningBlock.tsx)**: Implementa a "chain-of-thought" visível, com timer live ("Thought for X s") e colapso automático.
- **[CodeAccordion.tsx](file:///extensions/vscode/src/components/common/CodeAccordion.tsx)**: Usado para exibir diffs de arquivos e buscas no código com sintaxe highlight integrada.

---

## 3. Internacionalização (translations.csv)

Diferente do projeto original que usa `react-i18next`, o Vectora utiliza um sistema baseado em CSV centralizado no Core Go.

### Estratégia de Adaptação:

1. **useTranslation Hook**: Portamos um hook adaptador customizado sob `src/hooks/useTranslation.ts`.
2. **Key Mapping**: O hook emula a API `t()` do i18next mas consulta as chaves injetadas via `translations.csv`.
3. **Data Bridge**: As chaves do CSV são lidas pela extensão e enviadas ao Webview via `postMessage` no evento de inicialização.

---

---

## 6. Fluxo de Integração (IPC Bridge)

1. **Extension Host**: Monitora o processo Go do Core.
2. **IPC Client**: Recebe notificações JSON-RPC (ex: `{"method": "progress", "params": {"file": "main.go", "percent": 45}}`).
3. **Webview Provider**: Faz o disparo de `postMessage` para a interface React.
4. **App Container**: Atualiza o estado global das mensagens, injetando o componente `EmbedStatus` na conversa ativa.

---

> [!IMPORTANT]
> O blueprint garante que a interface do Vectora seja informativa sem ser intrusiva. O segredo está na tipografia condensada e no uso estratégico de cores de "segundo plano" (description colors).
