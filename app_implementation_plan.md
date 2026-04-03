# Plano de Implementação: Web UI (Vectora Desktop)

Este plano descreve o desenvolvimento da interface gráfica principal do Vectora, utilizando **Wails**, **Next.js 14**, **TailwindCSS** e **Framer Motion**.

---

## 1. Stack Tecnológica e Localização

O Web UI é embarcado no binário `cmd/vectora-web`.

- **Frontend:** `internal/app/` (Next.js 14, App Router).
- **Backend Bridge:** Go + Wails v3.
- **Build Tool:** Bun para velocidade de exportação estática.
- **Estética:** Kaffyn Dark Premium (Zinc, Slate, Emerald).

---

## 2. Arquitetura de Comunicação (Wails Bridges)

O Web UI é um **cliente IPC** que utiliza o Wails como transportador.

1.  **Go Bindings (`cmd/vectora-web/app.go`):** O Go expõe uma struct `App` que contém o `ipcClient`.
2.  **JS Proxy:** O Frontend chama `window.go.main.App.CallIPC(method, payload)`.
3.  **Encapsulamento:** O Next.js não sabe que o Daemon existe; ele fala apenas com o Proxy do Wails.

---

## 3. Estrutura de Componentes (`internal/app/`)

### 3.1 Chat Interface (A Experiência Principal)

- **Chat Feed:** Mensagens com Markdown, Highlighting de código e ícones de Workspace.
- **Sidebar (Workspace Switcher):** Listagem de bases ativas e botão de Indexação.
- **Input Area:** Textarea expandível com suporte a Drag-and-Drop de arquivos.

### 3.2 Workspace Manager

- **Index Browser:** Galeria de datasets do Vectora Index com readme.
- **Local Indexer:** Tela de progresso para pastas locais.
- **Settings:** Configuração de Gemini/Qwen via formulário persistente no Daemon.

---

## 4. O Sistema de Build (Embedded Assets)

Para manter o binário único:

1.  Rodar `bun run build` dentro de `internal/app/`.
2.  O Next.js gera a pasta `out/` (Static Export).
3.  O `cmd/vectora-web/main.go` usa `//go:embed internal/app/out` para injetar o site no binário.

---

## 5. Regras de Estética (Rich Aesthetics)

- **RN-UI-01:** Uso de **Glassmorphism** em barras laterais.
- **RN-UI-02:** Micro-animações de entrada (Framer Motion) para mensagens do Agente.
- **RN-UI-03:** Modo Escuro (`dark mode`) forçado via CSS para preservar a identidade visual Kaffyn.

---

## 6. Próximos Passos (App Refactor)

1.  [ ] **Implementar o State Manager:** Usar **Zustand** no frontend para sincronizar o estado das Workspaces com os eventos do IPC.
2.  [ ] **Diff Viewer:** Criar um componente de visualização de alterações para a ferramenta `edit`, permitindo ao usuário ver o snapshot do GitBridge antes de aceitar.
3.  [ ] **Systray Sync:** Sincronizar o ícone de status da bandeja com a atividade do Next.js (idle, thinking, indexing).

---

## 7. Verificação de Sucesso

- O binário final `vectora-web.exe` deve ter menos de 50MB.
- O tempo de boot (de clique até a tela de chat) deve ser inferior a 1.5s.
- Nenhuma chamada `fetch` deve ser feita para domínios externos (Gemini é via Go).

[Fim do Plano do App - Revisão 2026.04.03]
