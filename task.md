# Vectora Implementation Task List

## 🟢 SDK & Gateway Consolidation (Standard 2026)
- [x] Unificar lista de modelos em `AGENTS.md` (Top 10 Famílias)
- [x] Remover Claude Mythos (Preview Restrito) de toda a documentação
- [x] Atualizar README.md e README.pt.md com a tabela completa de 10 famílias
- [x] Implementar `GatewayProvider` no Core (OpenAI SDK compatible)
- [x] Implementar detecção de família e roteamento inteligente de embeddings no `Router`
- [x] Expor `models list` via CLI e IPC
- [ ] Verificar consistência de tipos em `core/llm/provider.go` (CompletionRequest deve suportar novos modelos)
- [ ] Testar integração real com chaves OpenRouter/Anannas

## 🟡 Phase 9: Multi-Tenancy (MTP) Foundation
- [ ] Criar `core/manager/tenant.go` para gerenciar namespaces de workspace
- [ ] Implementar `WorkspaceID` determinístico (hash do path absoluto)
- [ ] Modificar `core/infra/config.go` para suportar diretórios de dados dinâmicos
- [ ] Criar estrutra de subdiretórios em `appdata/workspaces/<id>/`
- [ ] Integrar Tenant no `context.Context` dos handlers IPC em `core/api/ipc/server.go`

## ⚪ Phase 10: Storage Isolation
- [ ] Isolar `chat_history.db` (bbolt) por tenant
- [ ] Isolar `chromadb/` (chromem-go) por tenant
- [ ] Implementar rotina de Eviction (unload de memória após inatividade)

## ⚪ Phase 11: Security & Guardian
- [ ] Implementar interceptor de path traversal em `core/os/guardian.go`
- [ ] Validar todas as operações de arquivo (read/write/edit) contra a Root do Tenant

## ⚪ Phase 7: Protocol Integration (ACP/MCP)
- [ ] Implementar Servidor MCP no Core para modo Sub-Agent
- [ ] Implementar Agente ACP no Core para modo CLI/Extension
- [ ] Adicionar suporte a `vectora-agent` binary
