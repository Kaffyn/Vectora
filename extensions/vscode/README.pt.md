# Vectora AI — Extensão VS Code

> **Assistente de codificação com IA local** — Chat com RAG + ferramentas agentivas via ACP (Agent Client Protocol).

O Vectora opera em dois modos distintos, adaptados para diferentes necessidades de desenvolvimento:

- **Modo Agente (Agent Mode):** Painel de chat interativo de alta fidelidade para interação direta do desenvolvedor.
- **Modo Sub-Agente (Sub-Agent Mode):** Motor especialista integrado em background que pode ser invocado por outros agentes (como o **Antigravity**) para realizar pesquisas e execuções em todo o repositório.

[![Versão](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/Kaffyn/Vectora)
[![Licença](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## Por que o Vectora?

Ao contrário de assistentes genéricos que enxergam apenas a janela atual do editor, o Vectora **entende toda a sua base de código**. Ele combina RAG (Retrieval-Augmented Generation) com ferramentas agentivas, operando dentro de uma Trust Folder segura e totalmente local.

| Funcionalidade  |    Copilot / IA genérica    |                 Vectora                 |
| :-------------- | :-------------------------: | :-------------------------------------: |
| **Contexto**    |    Apenas a janela atual    |          Base completa via RAG          |
| **Privacidade** | Dados enviados para a nuvem |       Índice local + Trust Folder       |
| **Ações**       |  Sugestões de texto apenas  | Ler / Escrever / Executar com segurança |
| **Protocolo**   |        Proprietário         |       ACP aberto (interoperável)        |
| **Índice**      |           Nenhum            |   chromem-go vectors + BBolt metadata   |

---

## Arquitetura

```
┌─────────────────────┐     ACP (JSON-RPC 2.0)     ┌──────────────────────┐
│  Extensão VS Code   │ ◄──────────────────────────► │  Vectora Core        │
│  (TypeScript)       │     stdio · JSON c/ newline │  vectora acp (Go)    │
│                     │                              └──────────┬───────────┘
│  ┌───────────────┐  │                                         │
│  │  ACPClient    │  │                              ┌──────────▼───────────┐
│  │  ChatPanel    │  │                              │  Engine              │
│  │  BinaryMgr    │  │                              │  RAG · Tools · LLM   │
│  │  InlineComp   │  │                              └──────────┬───────────┘
│  └───────────────┘  │                                         │
└─────────────────────┘                              ┌──────────▼───────────┐
                                                     │  Gemini / Claude API │
                                                     │  (HTTPS, nuvem)      │
                                                     └──────────────────────┘
```

**Princípio de segurança:** Todo o tráfego de rede ocorre apenas entre o Core e as APIs de IA. O canal extensão ↔ Core é 100% local via pipes stdio.

O Vectora utiliza **SDKs oficiais** (`google.golang.org/genai`, `anthropic-sdk-go`, `voyageai`) para se comunicar com os provedores de IA, garantindo máxima estabilidade e segurança.

---

## Funcionalidades

### Chat com RAG

Faça perguntas sobre o seu projeto. O Vectora busca semanticamente os arquivos indexados, lê o código relevante e fornece respostas contextuais — em streaming, token por token.

### Ferramentas Agentivas (no Core)

O core pode invocar ferramentas automaticamente dentro da Trust Folder:

| Ferramenta     | O que faz                                            |
| -------------- | ---------------------------------------------------- |
| `read_file`    | Lê conteúdo de arquivos (máx. 50 KB, com truncagem)  |
| `grep_search`  | Busca regex em todos os arquivos do projeto          |
| `terminal_run` | Executa comandos shell (sandbox com timeout de 30 s) |

Todas as chamadas de ferramenta passam pelo **Guardian** — uma camada de políticas que bloqueia acesso a segredos, chaves, bancos de dados e binários.

### Respostas em Streaming

As respostas aparecem token por token. Chamadas de ferramenta, planos do agente e raciocínio são exibidos em tempo real no painel de chat.

### Provedor de Completions Inline (Stub)

Um `VectoraInlineProvider` é registrado como `InlineCompletionItemProvider` do VS Code. Ele respeita o token de cancelamento do VS Code e está pronto para ser conectado a qualquer backend de completion rápido.

> **Nota:** A chamada ao backend de completion inline é uma funcionalidade planejada. O provedor está estruturalmente completo; conecte-o ao endpoint LLM desejado quando o backend estiver disponível.

### Gerenciamento Automático de Binário

A classe `BinaryManager` resolve o binário na seguinte ordem de prioridade:

1. Configuração `vectora.corePath` definida pelo usuário
2. Diretório gerenciado (`~/.vectora/bin/`)
3. `PATH` do sistema (`where vectora` / `which vectora`)
4. Download automático dos GitHub Releases (com notificação de progresso)

---

## Pré-requisitos

### 1. Binário do Vectora Core

A extensão executa `vectora acp <workspace>` via stdio. Obtenha o binário de uma das formas abaixo:

**Opção A — Download automático (recomendado)**

Na primeira utilização, se o binário não for encontrado, a extensão oferece o download automático para `~/.vectora/bin/`.

**Opção B — Download manual dos Releases**

```
https://github.com/Kaffyn/Vectora/releases
```

Coloque `vectora.exe` (Windows) ou `vectora` (Linux/macOS) no seu `PATH`, ou configure `vectora.corePath`.

**Opção C — Compilar do código-fonte**

```bash
git clone https://github.com/Kaffyn/Vectora.git
cd Vectora
go build -o vectora ./cmd/core/
# Mova para o PATH ou configure vectora.corePath nas configurações do VS Code
```

**Verificar instalação:**

```bash
vectora --version
# Esperado: vectora version 0.1.0
```

### 2. Chave de API

O Vectora usa APIs de LLM e embeddings na nuvem. Configure sua chave:

```
# Windows: %USERPROFILE%\.Vectora\.env
# Linux/macOS: ~/.Vectora/.env

GEMINI_API_KEY=sua_chave_gemini_aqui
# ou
CLAUDE_API_KEY=sua_chave_claude_aqui
# ou
VOYAGE_API_KEY=sua_chave_voyage_aqui
```

**Provedores suportados:**

| Provedor           | Chat | Embeddings             |
| ------------------ | ---- | ---------------------- |
| Gemini (Google)    | ✅   | ✅ Nativo              |
| Claude (Anthropic) | ✅   | ⚠️ Embeddings fallback |
| Voyage AI          | ❌   | ✅ Alta Fidelidade     |

### 3. Pasta de Workspace Aberta

A extensão exige uma pasta de workspace aberta no VS Code. Todas as operações de ferramentas ficam restritas a esta pasta.

---

## Instalação

### Via Arquivo VSIX

1. Baixe `vectora-vscode-0.1.0.vsix` em [Releases](https://github.com/Kaffyn/Vectora/releases)
2. Abra o VS Code → **Extensões** (`Ctrl+Shift+X`)
3. Clique em **`···`** → **"Instalar do VSIX..."**
4. Selecione o arquivo `.vsix`
5. **Reinicie o VS Code** quando solicitado

### Via Linha de Comando

```bash
code --install-extension vectora-vscode-0.1.0.vsix
```

### Primeiro Uso

1. Abra uma pasta de projeto no VS Code
2. A extensão ativa automaticamente ao iniciar
3. Se o binário não for encontrado, aceite o download automático
4. Verifique se `%USERPROFILE%\.Vectora\.env` contém uma chave de API válida
5. Clique no **ícone do Vectora** na Barra de Atividades ou use o Command Palette

---

## Configuração

Abra as Configurações do VS Code (`Ctrl+,`) e pesquise por `vectora`:

| Configuração        | Tipo    | Padrão          | Descrição                                                              |
| ------------------- | ------- | --------------- | ---------------------------------------------------------------------- |
| `vectora.corePath`  | string  | `"vectora"`     | Caminho para o binário. Use o caminho completo se não estiver no PATH. |
| `vectora.workspace` | string  | _(pasta atual)_ | Sobrescreve a pasta raiz para indexação RAG.                           |
| `vectora.streaming` | boolean | `true`          | Ativa respostas em streaming (token por token).                        |

**Exemplo de `settings.json`:**

```json
{
  "vectora.corePath": "C:\\Users\\voce\\.vectora\\bin\\vectora.exe",
  "vectora.workspace": "C:\\Projetos\\meu-app",
  "vectora.streaming": true
}
```

---

## Uso

### Painel de Chat

1. Clique no **ícone do Vectora** na Barra de Atividades
2. Digite sua pergunta e pressione **Enter** (`Shift+Enter` para nova linha)
3. Acompanhe a resposta em streaming — chamadas de ferramenta aparecem em tempo real
4. Pressione **Stop** para cancelar uma requisição em andamento

### Menu de Contexto (Clique Direito)

Selecione código no editor e clique com o botão direito:

| Comando                       | Ação                                           |
| ----------------------------- | ---------------------------------------------- |
| **Vectora: Explicar Código**  | Abre o chat com uma solicitação de explicação  |
| **Vectora: Refatorar Código** | Abre o chat com uma solicitação de refatoração |

Se nenhum código estiver selecionado, o conteúdo completo do arquivo é utilizado.

### Command Palette

Abra com `Ctrl+Shift+P` e digite `Vectora`:

| Comando                     | Descrição                                  |
| --------------------------- | ------------------------------------------ |
| `Vectora: Nova Sessão`      | Limpa o histórico e inicia uma nova sessão |
| `Vectora: Explicar Código`  | Explica o código selecionado               |
| `Vectora: Refatorar Código` | Refatora o código selecionado              |

### Barra de Status

O item na barra de status (canto inferior direito) indica o estado atual:

| Ícone                    | Significado                               |
| ------------------------ | ----------------------------------------- |
| `⟳ Vectora starting...`  | Binário iniciando, handshake em progresso |
| `✓ Vectora: <modelo>`    | Conectado e pronto                        |
| `○ Vectora Disconnected` | Core não está em execução                 |
| `✗ Vectora Error`        | Erro — veja o tooltip para detalhes       |

---

## Segurança

### Trust Folder

Todas as operações de sistema de arquivos ficam restritas à pasta do workspace aberto. A extensão não consegue instruir o core a acessar arquivos fora do projeto.

### Guardian — Políticas Fixas

Os seguintes tipos de arquivo são bloqueados no nível do engine, independentemente do que o LLM solicitar:

| Categoria  | Padrões                         |
| ---------- | ------------------------------- |
| Segredos   | `.env`, `.env.*`                |
| Chaves     | `.key`, `.pem`, `.p12`, `.pfx`  |
| Bancos     | `.db`, `.sqlite`                |
| Binários   | `.exe`, `.dll`, `.so`, `.dylib` |
| Chaves SSH | `id_rsa`, `id_ed25519`, `*.pub` |

### Solicitações de Permissão

Para operações de escrita, o core emite uma notificação `session/request_permission`. A extensão exibe um diálogo modal no VS Code:

- **Permitir** — autoriza a operação específica
- **Negar** — bloqueia a operação

---

## Referência do Protocolo ACP

A extensão se comunica com o core usando o **Agent Client Protocol** (JSON-RPC 2.0 via stdio).

### Cliente → Core (Requisições)

| Método               | Descrição                                     |
| -------------------- | --------------------------------------------- |
| `initialize`         | Handshake — declara capacidades do cliente    |
| `session/new`        | Cria uma nova sessão de agente                |
| `session/prompt`     | Envia prompt do usuário; retorna `stopReason` |
| `session/cancel`     | Cancela um prompt em andamento (notify)       |
| `fs/read_text_file`  | Lê arquivo via core                           |
| `fs/write_text_file` | Escreve arquivo via core                      |

### Core → Cliente (Notificações)

| Método                       | Descrição                                        |
| ---------------------------- | ------------------------------------------------ |
| `session/update`             | Chunk de streaming, evento de tool call ou plano |
| `session/request_permission` | Gate de permissão para ferramentas de escrita    |

### Variantes de `session/update`

| Campo `sessionUpdate` | Conteúdo                                              |
| --------------------- | ----------------------------------------------------- |
| `agent_message_chunk` | Token de texto streamed do modelo                     |
| `tool_call`           | Invocação de ferramenta iniciada (id, título, status) |
| `tool_call_update`    | Mudança de status da ferramenta (pending → completed) |
| `plan`                | Entradas do plano do agente (⏳ / ✅)                 |

---

## Solução de Problemas

### "Vectora starting..." nunca termina

**Verifique o binário:**

```powershell
# Windows
where vectora

# Linux/macOS
which vectora
```

**Verifique a chave de API:**

```powershell
# Windows
type $env:USERPROFILE\.Vectora\.env

# Linux/macOS
cat ~/.Vectora/.env
```

**Verifique os logs:**

```
%USERPROFILE%\.Vectora\logs\core.log
```

### "Vectora failed to start"

1. Defina `vectora.corePath` com o caminho completo do binário
2. Certifique-se de que uma pasta de workspace está aberta
3. Confirme que `GEMINI_API_KEY` está presente no arquivo `.env`

### Respostas lentas na primeira consulta

A primeira consulta dispara a indexação da base de código (geração de embeddings). As consultas seguintes usam o índice chromem-go em cache e são significativamente mais rápidas.

Para pré-indexar manualmente:

```bash
vectora embed --workspace /caminho/do/projeto
```

---

## Desenvolvimento

### Pré-requisitos

- Node.js 20+
- npm 10+
- VS Code 1.90+
- TypeScript 5+

### Build

```bash
cd extensions/vscode
npm install
npm run compile          # Build de produção (webpack)
npm run watch            # Rebuild automático ao salvar
npx vsce package         # Gera o arquivo .vsix
```

### Debug

1. Abra `extensions/vscode` no VS Code
2. Pressione `F5` → Extension Development Host é aberto
3. Defina breakpoints nos arquivos `src/*.ts`
4. Teste na nova janela do VS Code

### Estrutura do Código-Fonte

```
extensions/vscode/
├── src/
│   ├── extension.ts         # Ativação, comandos, barra de status
│   ├── client.ts        # Cliente ACP JSON-RPC (transporte stdio)
│   ├── chat-panel.ts        # WebviewPanel com UI de streaming
│   ├── binary-manager.ts    # Resolução de binário + download automático
│   ├── inline-completion.ts # InlineCompletionItemProvider (stub)
│   └── types/
│       └── acp.d.ts         # Definições de tipos do protocolo ACP
├── dist/
│   └── extension.js         # Bundle gerado (webpack)
├── media/
│   ├── icon.svg
│   └── icon.ico
├── package.json
├── tsconfig.json
└── webpack.config.js
```

---

## Roadmap

| Funcionalidade                      | Status        |
| ----------------------------------- | ------------- |
| Chat RAG via ACP                    | ✅ Disponível |
| Respostas em streaming              | ✅ Disponível |
| Visualização de tool calls          | ✅ Disponível |
| Visualização de planos do agente    | ✅ Disponível |
| Download automático de binário      | ✅ Disponível |
| Provedor de completion inline       | ⚙️ Stub       |
| Diff provider (edições write)       | Planejado     |
| Integração com terminal             | Planejado     |
| Slash commands (`/embed`, `/clear`) | Planejado     |
| Atalho de teclado `Ctrl+Shift+V`    | Planejado     |

---

## Licença

MIT — consulte [LICENSE](LICENSE)

---

**Vectora** — Parte da organização open source [Kaffyn](https://github.com/Kaffyn).
