# Blueprint: Ecossistema de Extensões & Clientes

**Status:** Extensão VS Code Implementada (v0.1.0)
**Mesa:** `extensions/vscode/`
**Tecnologia:** TypeScript, Webpack, JSON-RPC (ACP)

Este documento descreve como o Vectora Core se integra às interfaces de usuário, com foco especial no cliente VS Code e na estratégia de expansão para outras plataformas.

---

## 1. Extensão VS Code (O Cliente Primário)

A extensão VS Code atua como a interface principal de interação para desenvolvedores. Ela utiliza a interface visual do **Roo Code** como base estética premium, mas foi submetida a uma refatoração total para operar de forma nativa e enxuta no ecossistema Vectora.

### A Abordagem "Vectora-Shell":

- **Base de UI Premium:** Herdamos a estética rica de chat do Roo Code, garantindo uma experiência visual de estado da arte desde o primeiro dia.
- **Refatoração Estrutural:** Todo o código de suporte foi reconstruído do zero. A extensão atua como um cliente "fino", delegando 100% da lógica de IA, RAG e execução de ferramentas para o Core Go via JSON-RPC.
- **Purga de Código Legado:** Removemos mais de 200 arquivos e subpastas exclusivos do Roo (como Cloud, Marketplace, Worktrees e MCP proprietário), eliminando o "peso morto" e garantindo estabilidade técnica.
- **Comunicação Unificada:** A interface foi remapeada para falar diretamente com o nosso protocolo IPC, garantindo que ações de UI (como criar novas tarefas ou aprovar permissões) fluam corretamente para o daemon local.

### Componentes Implementados:

- **Chat Panel (Webview):** Interface rica (premium aesthetics) totalmente descomentada e funcional, suportando streaming de Markdown e feedback visual de ferramentas.
- **Inline Completion:** Assistente de escrita em tempo real integrado ao editor.
- **Binary Manager:** Responsável por detectar e auto-iniciar o binário do Core (Daemon) de forma transparente.

---

## 2. Estratégia de Comunicação (Client Engine)

Para garantir estabilidade, a extensão utiliza um protocolo de transporte sobre Stdio.

- **ACP (Agent Client Protocol):** Permite lidar com múltiplas sessões (sess\_...) e interrupções de usuário de forma graciosa.
- **Streaming:** O texto é renderizado conforme os tokens chegam do Core, sem aguardar a conclusão total da resposta.

---

## 3. Próximos Clientes (Roadmap)

### A. CLI Agêntica (`cmd/cli`)

- Uma interface de terminal interativa para quem prefere não sair do shell.
- Já possui o comando `vectora chat` integrado ao IPC do Core.

### B. Extensão Chrome/Browser

- Para permitir o uso do Vectora em plataformas de documentação web, injetando contexto do projeto em sites como GitHub ou StackOverflow.

### C. Desktop Dashboard (Tauri/Electron)

- Um painel centralizado para gerenciar todos os workspaces, visualizar estatísticas de indexação e configurar políticas `Guardian` de forma visual.

---

## 4. Portabilidade e Empacotamento

A extensão é distribuída como um arquivo `.vsix` independente.

- **Self-Contained:** O processo de build (`build.ps1`) gera tanto o pacote da extensão quanto o binário do core para a plataforma alvo, garantindo que o usuário tenha tudo o que precisa em uma única instalação.
