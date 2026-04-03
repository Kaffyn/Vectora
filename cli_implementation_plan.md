# Plano de Implementação: CLI (Vectora Chat Terminal)

Este plano descreve o desenvolvimento da interface de terminal robusta do Vectora, utilizando **Bubbletea**, **Lipgloss** e **Bubbles**.

---

## 1. Visão Geral

O Vectora CLI (`cmd/vectora chat`) é uma interface de comando e chat de baixo consumo e resposta instantânea para desenvolvedores. Ele atua como um **cliente IPC** embutido no binário principal.

- **Stack:** Bubbletea (TUI Framework).
- **Consumo de RAM:** < 10MB.
- **Isolamento:** Fale direto com o Daemon sem UI gráfica.

---

## 2. Experiência do Usuário (TUI Design)

### 2.1 O Chat Feed (Bubbletea Model)

- **Input Line:** Campo de texto com histórico (seta para cima/baixo).
- **Mensagens:** Bolhas de texto formatadas com Lipgloss.
- **Highlight:** Deteção automática de blocos de código e formatação ANSI.

### 2.2 Gestão de Contexto no Terminal

- **Flag `--workspace`:** Permite carregar bases específicas via linha de comando.
- **Flag `--index`:** Inicia indexação silenciosa com barra de progresso no terminal.
- **Comando `undo`:** Reverte a última operação via GitBridge diretamente do CLI.

---

## 3. Arquitetura de Componentes

O CLI reutiliza o pacote `internal/ipc/client.go` para conectar-se ao Pipe:

1. **Início:** `NewCLI()` inicia o Loop do Bubbletea.
2. **Mensagem enviada:** O model do Bubbletea dispara uma `Command` Go que chama `ipcClient.Request("workspace.query", ...)`.
3. **Resposta recebida:** A mensagem retorna via Msg do Bubbletea e renderiza no View do TUI.

---

## 4. Estética de Terminal (Charm Aesthetics)

- **RN-CLI-01:** Uso de gradientes sutis nos cabeçalhos de mensagem.
- **RN-CLI-02:** Spinner animado durante o "Thinking..." do LLM.
- **RN-CLI-03:** Suporte total a No-Color mode via flags para ambientes de servidor legados.

---

## 5. Próximos Passos (CLI Refactor)

1.  [ ] **Markdown Renderer:** Implementar o suporte a Markdown enriquecido no terminal usando a lib `glamour` da Charm (parceira do Bubbletea).
2.  [ ] **Pager:** Implementar scroll infinito no terminal para sessões longas de depuração.
3.  [ ] **Autocomplemento:** Implementar autocompletar de comandos do Agente (ex: `/read`, `/find`).

---

## 6. Verificação de Sucesso

- O `vectora chat` deve ser invocado em qualquer terminal (CMD, PowerShell, Kitty, iTerm2).
- O delay entre o envio do prompt e o início do streaming no terminal deve ser inferior a 100ms.

[Fim do Plano do CLI - Revisão 2026.04.03]
