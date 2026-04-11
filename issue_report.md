# Vectora Issue Report - Bug Collection

Este documento consolida os bugs identificados durante a migração e teste da interface Antigravity no Vectora.

---

## 1. Falha no Carregamento da Webview (vectora.chatView)

**Status**: Identificado
**Componente**: Extension Host / `ChatViewProvider`

### Descrição

Ao abrir o VS Code, a barra lateral do Vectora exibe a mensagem de erro: _"Ocorreu um erro ao carregar o modo de exibição: vectora.chatView"_.

### Causa Provável

O registro do provedor de view (`registerWebviewViewProvider`) está ocorrendo dentro da função assíncrona `startVectora` no `extension.ts`. Se o VS Code tentar renderizar a UI antes que a inicialização do Core termine, o provedor ainda não está registrado, resultando em falha de resolução da view.

### Sugestão de Correção

- Mover o registro do provedor para o início da função `activate`.
- Implementar um estado "Loading" ou "Offline" na Webview enquanto o Core inicializa.

---

## 2. Falha na Inicialização Automática do Core

**Status**: Identificado
**Componente**: `BinaryManager` / Build Script

### Descrição

A extensão indica "Vectora: Ready" na barra de status, mas o processo de background (`vectora start`) não é iniciado automaticamente. O ícone do Tray não aparece conforme esperado.

### Causa Provável

**Inconsistência de caminhos**:

- O script `build.ps1` instala o binário em `%LOCALAPPDATA%\Vectora\vectora.exe`.
- O `BinaryManager.ts` procura o binário em `~/.vectora/bin/vectora.exe`.
  Como o binário não é encontrado no local esperado pela extensão, o `spawn` do comando `start` falha ou não ocorre.

### Sugestão de Correção

- Sincronizar os caminhos de busca entre o script de build e o `BinaryManager`.
- Adicionar verificação de erro explícita no `spawn` do processo de background.

---

## 3. Múltiplas Instâncias do Core (Core Bug)

**Status**: Identificado
**Componente**: Vectora Core (Binary)

### Descrição

É possível iniciar múltiplos processos principais do Vectora simultaneamente (ex: ao abrir o VS Code após já ter iniciado o Vectora manualmente). Isso resulta em múltiplos ícones no Tray e potencial conflito de recursos.

### Causa Provável

O executável do Core não possui um mecanismo de trava (pidfile ou socket lick) que impeça a execução de uma segunda instância principal.

### Sugestão de Correção

- Implementar um check de instância única no Core (ex: tentar escutar em uma porta fixa ou criar um arquivo de trava em uma pasta global temporária).
- Se uma instância já estiver rodando, o novo processo deve apenas focar a instância existente ou sair silenciosamente.

---

_Este relatório será atualizado conforme novos bugs forem reportados._
