# Vectora Index Service

Backend de indexação vetorial escalável em nuvem. Microsserviço Go rodando em Vercel com storage em Supabase PostgreSQL + pgvector.

## 🏗️ Arquitetura

```
Desktop/Daemon ← gRPC → Vercel (Index Service) ← SQL ← Supabase PostgreSQL (pgvector)
```

## 📋 Requisitos

- **Go 1.26+**
- **PostgreSQL 14+** (Supabase)
- **Protocol Buffers** (protoc)
- **Vercel CLI** (deploy)

## 🚀 Setup Local

### 1. Clonar e Preparar

```bash
cd services/index
cp .env.example .env
# Editar .env com suas credenciais Supabase
```

### 2. Criar Database em Supabase

1. Criar novo projeto em [Supabase](https://supabase.com)
2. Copiar `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_KEY`
3. Na aba SQL Editor do Supabase, rodar `schema.sql`

```bash
# Ou via CLI:
psql "postgresql://postgres:PASSWORD@ENDPOINT:5432/postgres" < schema.sql
```

### 3. Instalar Dependências

```bash
go mod download
go mod tidy
```

### 4. Gerar Código gRPC

```bash
# Instalar protoc (se não tiver)
# Windows: choco install protobuf
# macOS: brew install protobuf
# Linux: apt-get install protobuf-compiler

# Instalar plugins Go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Gerar código
make generate

# Ou manual:
mkdir -p api/gen/go/vectora/index/v1
protoc -I api/v1 \
  --go_out=api/gen/go \
  --go-grpc_out=api/gen/go \
  api/v1/index.proto
```

### 5. Rodar Localmente

```bash
make run

# Ou:
GRPC_PORT=3000 go run cmd/server/main.go
```

### 6. Testar com grpcurl

```bash
# Instalar grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Listar serviços
grpcurl -plaintext localhost:3000 list

# Chamar rpc (quando implementado)
grpcurl -plaintext \
  -d '{"name": "my-workspace", "owner_id": "user123"}' \
  localhost:3000 vectora.index.v1.IndexService/CreateWorkspace
```

## 📦 Estrutura de Pastas

```
services/index/
├── api/
│   └── v1/
│       ├── index.proto           # Definições gRPC
│       └── gen/                  # Código gerado (gitignored)
├── cmd/
│   └── server/
│       └── main.go               # Entry point
├── internal/
│   ├── db/
│   │   ├── postgres.go           # Conexão DB
│   │   └── migrations.go         # Schema management
│   ├── service/
│   │   ├── workspace.go          # Workspace CRUD
│   │   ├── index.go              # Index CRUD
│   │   ├── document.go           # Document management
│   │   └── search.go             # Vector search
│   ├── embedding/
│   │   └── provider.go           # Gemini/OpenAI
│   └── grpc/
│       └── server.go             # gRPC server setup
├── go.mod                        # Go dependencies
├── go.sum
├── Makefile                      # Comandos úteis
├── schema.sql                    # PostgreSQL schema
├── vercel.json                   # Vercel config
├── .env.example                  # Template de variáveis
└── README.md
```

## 🔧 Makefile Targets

```bash
make help              # Listar todos os targets
make run               # Rodar servidor localmente
make generate          # Gerar código protobuf
make test              # Rodar testes
make lint              # Verificar código (golangci-lint)
make build             # Build para produção
make deploy            # Deploy em Vercel
make docker-build      # Build Docker image
make docker-run        # Rodar em Docker
```

## 📐 Endpoints gRPC

### Workspace Management
- `CreateWorkspace(CreateWorkspaceRequest) → WorkspaceResponse`
- `GetWorkspace(GetWorkspaceRequest) → WorkspaceResponse`
- `ListWorkspaces(ListWorkspacesRequest) → ListWorkspacesResponse`
- `DeleteWorkspace(DeleteWorkspaceRequest) → Empty`

### Index Management
- `CreateIndex(CreateIndexRequest) → IndexResponse`
- `GetIndex(GetIndexRequest) → IndexResponse`
- `ListIndexes(ListIndexesRequest) → ListIndexesResponse`
- `DeleteIndex(DeleteIndexRequest) → Empty`

### Document Operations
- `UploadDocument(stream UploadDocumentRequest) → UploadDocumentResponse` (streaming)
- `GetDocument(GetDocumentRequest) → DocumentResponse`
- `ListDocuments(ListDocumentsRequest) → ListDocumentsResponse`
- `DeleteDocument(DeleteDocumentRequest) → Empty`

### Search & Indexing
- `SearchDocuments(SearchDocumentsRequest) → SearchDocumentsResponse`
- `GetIndexingStatus(GetIndexingStatusRequest) → IndexingStatusResponse`

### Health
- `Health(Empty) → HealthResponse`

## 🌐 Variáveis de Ambiente

Veja `.env.example` para lista completa. Principais:

```env
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=...
SUPABASE_SERVICE_KEY=...
GEMINI_API_KEY=...
GRPC_PORT=3000
EMBEDDING_PROVIDER=gemini
```

## 🚢 Deploy em Vercel

```bash
# Conectar repositório
vercel link

# Deploy
vercel deploy

# Production
vercel deploy --prod

# Configurar secrets
vercel env add SUPABASE_URL
vercel env add SUPABASE_ANON_KEY
vercel env add GEMINI_API_KEY
```

## 🧪 Testes

```bash
# Rodar testes unitários
make test

# Com cobertura
go test ./... -cover

# Teste E2E (requer servidor rodando)
make test-e2e
```

## 📊 Performance

### Índices de Banco de Dados

- **IVFFLAT**: Para busca vetorial (padrão, escalável)
- **Índices compostos**: workspace_id + document_id para filtros

### Otimizações

```sql
-- Aumentar work_mem durante operações de índice
SET work_mem = '256MB';

-- Vacuumar periodicamente
VACUUM ANALYZE chunks;
```

## 🔐 Segurança

- ✅ gRPC TLS (em produção)
- ✅ Rate limiting por workspace
- ✅ Row-level security (RLS) no PostgreSQL
- ✅ Validação de input
- ✅ Auditoria de operações

## 📚 Documentação

- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [gRPC Go Docs](https://grpc.io/docs/languages/go/)
- [Supabase Docs](https://supabase.com/docs)
- [pgvector Documentation](https://github.com/pgvector/pgvector)

## 🤝 Integração Desktop

Veja `desktop/index_client.go` para exemplo de client gRPC em Go.

## 📞 Support

Issues e PRs bem-vindos!

## 📄 License

MIT - Kaffyn Ecosystem
