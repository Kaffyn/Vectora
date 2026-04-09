# Issues: Blueprint vs Implementação Real do Core — Status Atualizado

> **Data:** 2026-04-08 (atualizado após resolução das issues)
> **Escopo:** Comparação entre os blueprints em `docs/` e a implementação efetiva em `core/` > **Arquivos blueprint analisados:** `Storage.md`, `Config_Manager.md`, `Observabilidade.md`, `POLICIES.md`, `LLM_Gateway.md`, `Ingestion.md`, `Tools_Executors.md`, `API.md`, `agentic_tools.md` > **Excluídos:** `implementation_plan.md`, `EXTENSIONS_PLAN.pt.md`, `golang-timeline.md`, `rust-timeline.md`

---

## Resumo Executivo (APÓS RESOLUÇÃO)

| Módulo              | Blueprint                                              | Implementação                                  | Status       |
| ------------------- | ------------------------------------------------------ | ---------------------------------------------- | ------------ |
| **Storage**         | `core/storage/` (BBolt + Chromem + Engine)             | `core/db/` (consolidado)                       | ✅ Resolvido |
| **Config**          | `core/config/` (YAML + AES-256-GCM)                    | `core/infra/config.go` (.env documentado)      | ✅ Resolvido |
| **Observabilidade** | `core/telemetry/` (RotatingWriter + slog JSON)         | `core/telemetry/` (usado via `infra.Logger()`) | ✅ Resolvido |
| **Policies**        | `core/policies/` (Guardian + YAML rules)               | `core/policies/` completo com 4 YAMLs          | ✅ Sempre OK |
| **LLM Gateway**     | `core/llm/` (Provider + ContextManager + Gemini)       | `core/llm/` com Gemini + Claude + Qwen         | ✅ Superior  |
| **Ingestion**       | `core/ingestion/` (Parser + DependencyGraph + Indexer) | `core/ingestion/` (Parser + DependencyGraph)   | ✅ Resolvido |
| **Tools**           | `core/tools/` (3 tools)                                | `core/tools/` com 10 tools + Guardian          | ✅ Superior  |
| **API**             | `core/api/` (Router + JSON-RPC + gRPC + IPC)           | `core/api/` com ACP integrado                  | ✅ Resolvido |

---

## Resumo das Issues Resolvidas

| #   | Módulo              | Severidade | Status       | Ação Tomada                                                                 |
| --- | ------------------- | ---------- | ------------ | --------------------------------------------------------------------------- |
| 1   | **Config**          | 🔴 Crítico | ✅ Resolvido | Removido `core/config/`. `infra/config.go` documentado como oficial (.env). |
| 2   | **Storage**         | ⚠️ Médio   | ✅ Resolvido | Removido `core/storage/`. Daemon usa `core/db/` exclusivamente.             |
| 3   | **Observabilidade** | ⚠️ Médio   | ✅ Resolvido | `infra/logger.go` delega para `telemetry/` com RotatingWriter (10MB).       |
| 4   | **API duplicata**   | ⚠️ Baixo   | ✅ Resolvido | Removido `core/api/methods/` (duplicata de `handlers/`).                    |
| 5   | **Ingestion**       | ⚠️ Baixo   | ✅ Resolvido | Removidos parâmetros LLM e Storage do Indexer.                              |
| 6   | **gRPC stub**       | ℹ️ Info    | ✅ Resolvido | Removido `core/api/grpc/` (stub sem proto gerado).                          |
| 7   | **ACP server**      | ℹ️ Info    | ✅ Resolvido | `vectora acp` command integrado no daemon.                                  |
| 8   | **MCP server**      | ℹ️ Info    | ⏳ Pendente  | `extensions/mcp-server/` conforme EXTENSIONS_PLAN.pt.md.                    |

---

## O Que Foi Removido (Código Morto)

| Diretório                       | Motivo                                                     |
| ------------------------------- | ---------------------------------------------------------- |
| `core/storage/` (5 arquivos)    | Duplicata de `core/db/` — nunca importado pelo daemon      |
| `core/config/` (4 arquivos)     | Nunca importado — daemon usa `core/infra/config.go` (.env) |
| `core/api/methods/` (1 arquivo) | Duplicata de `core/api/handlers/tools_call.go`             |
| `core/api/grpc/` (2 arquivos)   | Stub sem proto gerado — sem uso real                       |

## O Que Foi Adicionado/Melhorado

| Adição                               | Descrição                                                      |
| ------------------------------------ | -------------------------------------------------------------- |
| `core/infra/logger.go` → `telemetry` | Agora usa RotatingWriter (10MB rotação, 1 backup)              |
| `vectora acp` command                | ACP server over stdio para IDEs (VS Code, JetBrains, Zed)      |
| `core/llm/claude_provider.go`        | Claude provider via HTTP API (não previsto no blueprint)       |
| `core/api/acp/`                      | ACP server completo com initialize, session, prompt, fs, tools |

## Pendências

| Item              | Descrição                                | Documento                    |
| ----------------- | ---------------------------------------- | ---------------------------- |
| MCP Server        | `extensions/mcp-server/` para Gemini CLI | `docs/EXTENSIONS_PLAN.pt.md` |
| VS Code Extension | `extensions/vscode/` ACP client          | `docs/EXTENSIONS_PLAN.pt.md` |
