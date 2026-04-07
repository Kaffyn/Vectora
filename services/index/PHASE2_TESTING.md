# Phase 2 - Testing Guide

Guia para testar a implementação da Fase 2 (gRPC Server Base).

---

## Pré-requisitos

### 1. Instalar protoc

**Windows (via Chocolatey)**:
```powershell
choco install protobuf
protoc --version
```

**macOS (via Homebrew)**:
```bash
brew install protobuf
protoc --version
```

**Linux (Ubuntu/Debian)**:
```bash
apt-get install protobuf-compiler
protoc --version
```

### 2. Instalar grpcurl

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
grpcurl --version
```

### 3. Criar projeto Supabase

1. Ir para https://supabase.com
2. Criar novo projeto
3. Copiar database credentials
4. Executar `schema.sql` na aba SQL Editor

### 4. Configurar .env

Copiar `.env.example` para `.env` e preencher:
```
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@YOUR_HOST:5432/postgres?sslmode=require
GRPC_PORT=3000
GRPC_HOST=0.0.0.0
```

---

## Testing Steps

### Passo 1: Gerar código Protobuf

```bash
cd services/index

# Instalar generators
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Gerar código
mkdir -p api/gen/go/vectora/index/v1
protoc -I api/v1 \
  --go_out=api/gen/go \
  --go-grpc_out=api/gen/go \
  api/v1/index.proto

# Verificar que os arquivos foram gerados
ls -la api/gen/go/vectora/index/v1/
```

**Esperado**: 2 arquivos `.pb.go`:
- `index.pb.go` (~1500 LOC)
- `index_grpc.pb.go` (~500 LOC)

### Passo 2: Compilar servidor

```bash
cd services/index
go build -o server.exe ./cmd/server

# Verificar que compilou
ls -lh server.exe
```

**Esperado**: Executável de ~20-25 MB

### Passo 3: Iniciar servidor

```bash
# Terminal 1: Iniciar servidor
cd services/index
.\server.exe

# Esperado: Logs como
# [INFO] Conectando no banco de dados...
# [INFO] Conexão com banco de dados estabelecida com sucesso
# [INFO] Iniciando Vectora Index Service em 0.0.0.0:3000
```

### Passo 4: Testar com grpcurl

Em outro terminal:

```bash
# Listar serviços
grpcurl -plaintext localhost:3000 list
# Esperado: vectora.index.v1.IndexService

# Obter detalhes do serviço
grpcurl -plaintext localhost:3000 describe vectora.index.v1.IndexService
# Esperado: Lista de RPCs (CreateWorkspace, GetWorkspace, etc)

# Test 1: CreateWorkspace
grpcurl -plaintext -d '{
  "name": "test-workspace",
  "owner_id": "user123"
}' localhost:3000 vectora.index.v1.IndexService/CreateWorkspace

# Esperado:
# {
#   "workspace": {
#     "id": "uuid-here",
#     "name": "test-workspace",
#     "owner_id": "user123",
#     "created_at": "2026-04-06T...",
#     "updated_at": "2026-04-06T..."
#   }
# }

# Test 2: ListWorkspaces
grpcurl -plaintext -d '{
  "owner_id": "user123",
  "page": 1,
  "page_size": 10
}' localhost:3000 vectora.index.v1.IndexService/ListWorkspaces

# Esperado:
# {
#   "workspaces": [
#     {
#       "id": "uuid-here",
#       "name": "test-workspace",
#       ...
#     }
#   ],
#   "total": 1
# }

# Test 3: CreateIndex
# Primeiro, get o workspace ID da resposta anterior
WORKSPACE_ID="uuid-from-above"

grpcurl -plaintext -d "{
  \"workspace_id\": \"$WORKSPACE_ID\",
  \"name\": \"test-index\",
  \"description\": \"Test index for documents\"
}" localhost:3000 vectora.index.v1.IndexService/CreateIndex

# Esperado:
# {
#   "index": {
#     "id": "uuid-here",
#     "workspace_id": "...",
#     "name": "test-index",
#     "description": "Test index for documents",
#     "document_count": 0,
#     "size_bytes": "0",
#     "created_at": "2026-04-06T...",
#     "updated_at": "2026-04-06T..."
#   }
# }

# Test 4: ListIndexes
grpcurl -plaintext -d "{
  \"workspace_id\": \"$WORKSPACE_ID\"
}" localhost:3000 vectora.index.v1.IndexService/ListIndexes

# Esperado:
# {
#   "indexes": [
#     {
#       "id": "uuid-here",
#       "name": "test-index",
#       ...
#     }
#   ],
#   "total": 1
# }
```

### Passo 5: Verificar dados em Supabase

No dashboard Supabase:
1. Abrir SQL Editor
2. Executar:
   ```sql
   SELECT * FROM workspaces;
   SELECT * FROM indexes;
   ```
3. Verificar que os dados aparecem com os valores inseridos via gRPC

### Passo 6: Testes de Validação

```bash
# Test: CreateWorkspace sem nome
grpcurl -plaintext -d '{
  "name": "",
  "owner_id": "user123"
}' localhost:3000 vectora.index.v1.IndexService/CreateWorkspace

# Esperado: Error com mensagem "nome do workspace é obrigatório"

# Test: CreateIndex sem workspace_id
grpcurl -plaintext -d '{
  "workspace_id": "nonexistent-id",
  "name": "test",
  "description": "test"
}' localhost:3000 vectora.index.v1.IndexService/CreateIndex

# Esperado: Error com mensagem "workspace não encontrado"
```

---

## Verificação Checklist

- [ ] `protoc --version` mostra versão instalada
- [ ] `make generate` compila sem erros (ou manual protoc command)
- [ ] `api/gen/go/vectora/index/v1/index.pb.go` existe (~1500 LOC)
- [ ] `api/gen/go/vectora/index/v1/index_grpc.pb.go` existe (~500 LOC)
- [ ] `go build ./cmd/server` cria binário de ~20-25 MB
- [ ] Servidor inicia com `.env` corrigido
- [ ] `grpcurl -plaintext localhost:3000 list` mostra `vectora.index.v1.IndexService`
- [ ] `CreateWorkspace` RPC funciona e retorna workspace com ID
- [ ] `ListWorkspaces` RPC retorna lista de workspaces
- [ ] `CreateIndex` RPC funciona e retorna índice com ID
- [ ] `ListIndexes` RPC retorna lista de índices
- [ ] Dados aparecem em Supabase (SELECT nas tabelas)
- [ ] Validações funcionam (erros para campos obrigatórios)
- [ ] Servidor para gracefully com Ctrl+C

---

## Troubleshooting

### Erro: "protoc: command not found"
**Solução**: Instalar protoc conforme instruções acima

### Erro: "DATABASE_URL is required"
**Solução**: Criar arquivo `.env` com `DATABASE_URL` válida

### Erro: "connection refused"
**Solução**: Verificar que:
1. Servidor está rodando na porta 3000
2. DATABASE_URL está conectável (ping ao Supabase)
3. Firewall não está bloqueando localhost:3000

### Erro: "workspace não encontrado"
**Solução**: CreateIndex requer workspace existente. Criar workspace primeiro e usar seu ID.

### Erro: "unknown service"
**Solução**: Gerar código protobuf com `protoc` antes de compilar

---

## Performance Notes

- CreateWorkspace: ~10-50ms (dependendo da latência Supabase)
- ListWorkspaces: ~20-100ms (dependendo de quantos registros)
- Max message size: 100 MB (para uploads de documentos)

---

## Próximas Ações (Phase 3)

Após verificar que Phase 2 funciona:

1. Implementar UploadDocument (streaming)
2. Implementar SearchDocuments (vector search com pgvector)
3. Integrar provider de embeddings (Gemini/OpenAI)
4. Adicionar testes E2E

---

**Data**: 2026-04-06
**Versão**: Phase 2 (gRPC Server Base)
