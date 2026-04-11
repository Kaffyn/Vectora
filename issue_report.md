# Vectora Issue Report - Bug Collection & Requests

Este documento consolida as falhas, discussões e requisitos do ecossistema Vectora.

---

## 🐞 Reports (Bugs)

### 1. Falha no Carregamento da Webview (vectora.chatView)

**Status**: Identificado
**Descrição**: Erro "Ocorreu um erro ao carregar o modo de exibição". Independente do estado do Core.
**Causa Provável**: Registro tardio do provedor no ciclo de vida da extensão.

### 2. Falha na Inicialização Automática do Core

**Status**: Identificado
**Descrição**: A extensão sinaliza prontidão, mas o processo de background (`vectora start`) não foi iniciado.
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

## ❓ Questions (Discussão)

### 10. Método de Singleton no Core

**Questão**: Para garantir instância única no Windows, prefere-se o uso de um arquivo de lock (`.vectora.lock`) ou uma tentativa de bind em porta TCP específica?

### 11. Estratégia de Fallback

**Questão**: Devemos manter a implementação HTTP manual como fallback de segurança ou migrar 100% para os SDKs oficiais?

---

## 🚀 Requests (Modernização e Requisitos)

### 12. Migração para SDKs Oficiais (Chat & Embeddings)

Requisito de migrar todas as implementações HTTP manuais para os SDKs oficiais de cada provedor, abrangendo tanto as funcionalidades de Chat quanto as de Embeddings.

**Referências Oficiais**:

- **Gemini (Chat & Embeddings)**: [google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai)
- **Claude (SDK)**: [github.com/anthropics/anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)
- **Voyage AI (SDK & Embeddings)**: [github.com/austinfhunter/voyageai](https://pkg.go.dev/github.com/austinfhunter/voyageai)

### 13. Revisão e Normalização de Modelos e Funcionalidades

Revisar e alinhar os identificadores de modelos e configurações avançadas com base nas documentações oficiais dos 3 provedores.

**Documentação de Referência**:

- **Gemini**: [Modelos e ThinkingConfig](https://ai.google.dev/gemini-api/docs/models?hl=pt-br)
- **Claude**: [Modelos e Prompt Caching](https://platform.claude.com/docs/en/api/sdks/go)
- **Voyage**: [Modelos de Embedding Profissional](https://pkg.go.dev/github.com/austinfhunter/voyageai)

---

_Este relatório será atualizado conforme novas interações._
