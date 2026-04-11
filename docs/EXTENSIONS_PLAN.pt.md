# Blueprint: Ecossistema de Extensões & Clientes

**Status:** Extensão VS Code Implementada (v0.1.0)
**Mesa:** `extensions/vscode/`
**Tecnologia:** TypeScript, Webpack, JSON-RPC (ACP)

Este documento descreve como o Vectora Core se integra às interfaces de usuário, com foco especial no cliente VS Code e na estratégia de expansão para outras plataformas.

---

## 1. Extensão VS Code (O Cliente Primário)

A extensão VS Code atua como a interface principal de interação para desenvolvedores. Ela não contém lógica de inteligência artificial; em vez disso, ela é um cliente fino para o daemon local.

### Componentes Implementados:

- **Unified Client:** Uma classe única que gerencia o ciclo de vida do processo `vectora.exe` e a comunicação via JSON-RPC.
- **Chat Panel (Webview):** Interface rica (premium aesthetics) que suporta streaming de Markdown, blocos de código e feedback visual de execução de ferramentas.
- **Inline Completion:** Assistente de escrita em tempo real integrado ao editor, utilizando modelos Flash para baixa latência (<300ms).
- **Binary Manager:** Responsável por detectar, baixar e auto-iniciar o binário do Core se ele não estiver rodando.

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
