# Vectora Issue Report - Bug Collection & Requests

Este documento consolidada as falhas e solicitações de melhoria do ecossistema Vectora, organizado por categoria.

---

## 🐞 Reports (Bugs)

### 1. Falha no Carregamento da Webview (vectora.chatView)

**Status**: Pendente
**Descrição**: Erro "Ocorreu um erro ao carregar o modo de exibição". Independente do estado do Core.
**Causa**: Registro tardio do provedor no `activate`.

### 2. Falha na Inicialização Automática do Core

**Status**: Pendente
**Descrição**: Status mostra "Ready" mas o processo de background (`vectora start`) não foi iniciado (Tray ausente).
**Causa**: Inconsistência de caminhos entre `build.ps1` e `BinaryManager.ts`.

### 3. Múltiplas Instâncias do Core

**Status**: Pendente
**Descrição**: Possibilidade de iniciar vários processos principais simultaneamente.
**Causa**: Falta de check de instância única (lock) no Core.

### 4. Proliferação de Processos e Discrepância de Binários

**Status**: Pendente
**Descrição**: Rodar o VS Code com Core manual gera processos redundantes com nomes diferentes (`vectora.exe` e `vectora-windows-amd64.exe`).

### 5. Ambiguidade na Configuração via CLI

**Status**: Pendente
**Descrição**: `config set` não indica as chaves válidas. Falta de esquema/lista de LLMs.

### 6. Opacidade no Comando `workspace ls`

**Status**: Pendente
**Descrição**: O comando exibe apenas o ID (hash) e não o caminho da pasta associada.

### 7. Falta de Alias para `workspaces`

**Status**: Pendente
**Descrição**: O comando no plural não funciona.

### 8. Falso Positivo de Antivírus

**Status**: Contornado (Recompilação)
**Descrição**: Windows Defender sinaliza o binário como `Trojan:Win32/Bearfoos.A!ml`.

### 9. Erro 404 Gemini (Nome do Modelo)

**Status**: Pendente
**Descrição**: O sistema tenta usar `gemini-3-flash`, que é inválido/inexistente na API v1beta.

---

## ❓ Questions (Arquitetural)

### 10. Uso do SDK Oficial do Gemini

**Status**: Aprovado para Migração
**Questão**: Por que não usamos `google.golang.org/genai` diretamente para simplificar o suporte a modelos de raciocínio (Thinking) e evitar erros de montagem de URL?

---

## 🚀 Requests (Melhorias)

### 11. Revisão e Normalização de Modelos (Docs)

**Status**: Atribuído
**Descrição**: Revisar os identificadores baseados na [documentação oficial](https://ai.google.dev/gemini-api/docs/models?hl=pt-br) (ex: `gemini-2.0-flash`, `gemini-1.5-pro`).

### 12. Migração para o SDK de Embeddings

**Status**: Atribuído
**Descrição**: Migrar a lógica de embeddings em `core/llm/gemini_provider.go` para o SDK oficial utilizando `client.Models.EmbedContent` conforme a [documentação de Go](https://ai.google.dev/gemini-api/docs/embeddings?hl=pt-br#go).

### 13. Suporte a ThinkingConfig

**Status**: Atribuído
**Descrição**: Adicionar suporte nativo à configuração de raciocínio (Thinking) para os modelos Gemini 2.0 que a suportam.

---

_Este relatório será atualizado conforme novos bugs forem reportados._
