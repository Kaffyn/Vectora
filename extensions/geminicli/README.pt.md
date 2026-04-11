# Vectora — Gemini CLI Bridge (MCP)

> **Exclusivo para Modo Sub-Agente** — Conecta o Gemini CLI ao RAG e ferramentas do Vectora usando o Model Context Protocol (MCP).

[![Versão](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/Kaffyn/Vectora)
[![Licença](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## Visão Geral

O `vectora-geminicli` é uma ponte que permite ao **Gemini CLI** (do Google Cloud SDK) se comunicar com o **Vectora Core**. Ao expor as capacidades do Vectora como ferramentas MCP, ele permite que o Gemini CLI:

1. **Pesquise sua base de código** de forma semântica (RAG).
2. **Leia e escreva arquivos** com segurança dentro de sua Trust Folder.
3. **Execute comandos verificados** para testar ou buildar código.

Nesta arquitetura, o Gemini CLI atua como o **Agente Principal**, enquanto o Vectora atua como um **Sub-Agente** especializado (Especialista) invocado sempre que contexto profundo de código ou execução de ferramentas é necessário.

---

## Instalação

### 1. Build a partir do Código-Fonte

Navegue até o diretório da extensão e compile o código TypeScript:

```bash
cd extensions/geminicli
npm install
npm run build
```

### 2. Verificar Binário

Certifique-se de que o binário do **Vectora Core** está no seu `PATH`. Você pode verificar com:

```bash
vectora --version
```

---

## Configuração

O Gemini CLI usa uma configuração JSON para descobrir servidores MCP. Você pode gerar essa configuração automaticamente usando o comando `config`.

### Configuração Automática

Execute o seguinte comando para ver o snippet JSON necessário para o seu setup:

```bash
node dist/index.js config
```

### Passos para Configuração

1. Abra (ou crie) seu arquivo de configurações do Gemini CLI:

   - **Linux/macOS**: `~/.config/gemini/settings.json`
   - **Windows**: `%APPDATA%\google-cloud-sdk\gemini\settings.json` (caminho aproximado)

2. Adicione a entrada `mcpServers`:

```json
{
  "mcpServers": {
    "vectora": {
      "command": "node",
      "args": ["/caminho/para/Vectora/extensions/geminicli/dist/index.js", "call-tool"]
    }
  }
}
```

_Nota: Certifique-se de que o caminho para o `index.js` seja absoluto._

---

## Uso

### REPL Interativo (Testes)

Você pode iniciar um REPL independente para testar se a conexão MCP com o Core está funcionando corretamente:

```bash
# Defina seu workspace antes de iniciar
export VECTORA_WORKSPACE=/caminho/para/seu/projeto
node dist/index.js
```

Dentro do REPL, tente:

- `/tools` — Lista todas as ferramentas disponíveis no Vectora Core.
- `/new` — Inicia uma nova sessão na pasta atual.
- `Como funciona a autenticação neste projeto?` — Faça uma pergunta de pesquisa.

### Variáveis de Ambiente

| Variável            | Descrição                                           | Padrão          |
| ------------------- | --------------------------------------------------- | --------------- |
| `VECTORA_CORE_PATH` | Caminho para o binário `vectora`                    | `vectora`       |
| `VECTORA_WORKSPACE` | Pasta para indexação e busca                        | `process.cwd()` |
| `DEBUG`             | Defina como `mcp:*` para ver mensagens do protocolo | _(desativado)_  |

---

## Desenvolvimento

- `npm run watch`: Recompila ao salvar alterações.
- `npm run start`: Executa a ponte diretamente.

---

## Licença

MIT — Parte do ecossistema [Vectora](https://github.com/Kaffyn/Vectora).
