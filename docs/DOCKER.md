# Docker Setup Guide - Vectora

Guia completo para construir, executar e fazer deploy do Vectora com Docker.

## 📋 Componentes

O stack Docker inclui:

- **Vectora API** - FastAPI aplicação em Python 3.13
- **PostgreSQL 18** - Banco de dados relacional principal
- **Valkey** - Cache em memória (compatível com Redis)
- **Qdrant** - Banco de dados vetorial para RAG

Todos os serviços rodam em uma rede Docker isolada (`vectora_network`).

## 🚀 Início Rápido

### Desenvolvimento Local

```bash
# Clonar repositório
git clone https://github.com/seu-user/vectora.git
cd vectora

# Copiar e configurar arquivo de ambiente
cp .env.example .env

# Editar .env com suas credenciais
nano .env

# Construir e iniciar serviços
docker compose up -d

# Aguardar inicialização (30-60 segundos)
docker compose logs -f vectora

# Testar API
curl http://localhost:8000/health
```

### Parar Serviços

```bash
docker compose down
```

### Remover Volumes (dados persistentes)

```bash
# ⚠️ Isso deleta banco de dados e cache!
docker compose down -v
```

## 🔧 Configuração do Ambiente

Edite `.env` para customizar:

### LLM Provider

```env
# Google Gemini (padrão)
LLM_PROVIDER=google-genai
GOOGLE_API_KEY=seu-api-key
GOOGLE_MODEL=gemini-2.0-flash

# OU Ollama (local)
LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://host.docker.internal:11434

# OU OpenAI
LLM_PROVIDER=openai
OPENAI_API_KEY=sk-...
```

### Database

```env
POSTGRES_USER=vectora
POSTGRES_PASSWORD=sua-senha-forte
POSTGRES_DB=vectora
```

### RAG (Vector Search)

```env
VOYAGE_API_KEY=seu-voyage-key
EMBEDDING_MODEL=voyage-4
QDRANT_URL=http://qdrant:6333
```

## 📦 Dockerfile

Multi-stage otimizado:

- **Stage 1**: Builder - Compila wheels
- **Stage 2**: Runtime - Imagem final pequena (não-root user)

**Recursos:**

- Python 3.13-slim (~ 200MB base)
- Health checks automáticos
- Non-root user (`vectora:1000`)
- Logging estruturado

### Rebuild da Imagem

```bash
# Build local
docker compose build

# Build sem cache
docker compose build --no-cache

# Build específico
docker compose build --no-cache vectora
```

## 🐳 docker-compose.yml

### Ordem de Inicialização

Serviços iniciam em paralelo, mas o Vectora aguarda health checks:

```
postgres → ready
valkey   → ready  → vectora inicia
qdrant   → ready
```

### Volumes

Dados persistentes em volumes Docker:

```
postgres_data       - Base PostgreSQL
valkey_data         - Cache Valkey
qdrant_data         - Índices vetoriais
qdrant_snapshots    - Snapshots Qdrant
```

**Local (host):**

```
./data/   - Embeddings, LanceDB, etc
./logs/   - Application logs
```

### Redes

Todos os serviços usam a rede `vectora_network`:

```bash
# Inspecionar rede
docker network inspect vectora_network
```

## 🔍 Monitoramento

### Logs

```bash
# Todos os serviços
docker compose logs

# Apenas Vectora
docker compose logs vectora

# Follow real-time
docker compose logs -f vectora

# Últimas 100 linhas
docker compose logs vectora --tail 100
```

### Status dos Serviços

```bash
docker compose ps

# Detalhado
docker compose ps --format "table {{.Service}}\t{{.Status}}"
```

### Health Checks

```bash
# Verificar health
docker compose exec vectora curl http://localhost:8000/health
docker compose exec postgres pg_isready
docker compose exec valkey redis-cli ping
docker compose exec qdrant curl http://localhost:6333/health
```

## 🛠️ Operações Comuns

### Reiniciar um Serviço

```bash
docker compose restart vectora
```

### Entrar em um Container

```bash
docker compose exec vectora bash
docker compose exec postgres psql -U vectora -d vectora
docker compose exec valkey redis-cli
```

### Executar Comandos

```bash
# Inicializar banco de dados
docker compose exec vectora python -m src.main --init-db

# Rodar testes
docker compose exec vectora pytest tests/

# Ver versões
docker compose exec vectora python --version
docker compose exec vectora uv --version
```

### Backup do Banco PostgreSQL

```bash
docker compose exec postgres pg_dump -U vectora vectora > backup.sql
```

### Restaurar do Backup

```bash
cat backup.sql | docker compose exec -T postgres psql -U vectora vectora
```

## 🚢 Deployment em VPS

### Requisitos

- Docker Engine 24+
- Docker Compose 2.20+
- 2GB RAM mínimo (4GB recomendado)
- 10GB disco

### Script de Deployment

```bash
#!/bin/bash

# Setup iniciais
mkdir -p /opt/vectora
cd /opt/vectora

# Clonar repositório
git clone https://github.com/seu-user/vectora.git .

# Copiar secrets
scp seu-host:/path/to/.env .env

# Build e start
docker compose up -d

# Inicializar banco
docker compose exec -T vectora python -m src.main --init-db

# Verificar status
docker compose ps
curl http://localhost:8000/health
```

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name api.vectora.com;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Systemd Service

```ini
[Unit]
Description=Vectora Docker Compose
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
WorkingDirectory=/opt/vectora
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
```

Ativar:

```bash
sudo systemctl enable vectora
sudo systemctl start vectora
```

## 🐛 Troubleshooting

### Container não inicia

```bash
# Ver logs completos
docker compose logs vectora

# Inspecionar imagem
docker compose build --no-cache vectora
```

### Erro de conexão PostgreSQL

```bash
# Verificar health
docker compose exec postgres pg_isready

# Restart
docker compose restart postgres

# Aguardar 10s e testar novamente
```

### Espaço em disco cheio

```bash
# Limpar Docker
docker system prune -a --volumes

# Manter volumes de dados (se necessário)
docker volume prune
```

### Performance lenta

```bash
# Aumentar limites de recursos em docker-compose.yml
services:
  vectora:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

## 📊 Exemplos de Uso

### Health Check

```bash
curl -X GET http://localhost:8000/health \
  -H "Content-Type: application/json"
```

### Exemplo de API Call

```bash
curl -X POST http://localhost:8000/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Olá, como posso ajudar?",
    "user_id": "user123"
  }'
```

## 🔒 Segurança

- ✅ Non-root user inside container
- ✅ Secrets in environment variables
- ✅ Health checks automáticos
- ✅ Restart policies
- ✅ Resource limits (recomendado para prod)

**NÃO:**

- ❌ Não commit `.env` com secrets
- ❌ Não exponha portas desnecessárias em prod
- ❌ Não use SQLite em produção

## 📚 Referências

- [Docker Docs](https://docs.docker.com)
- [Docker Compose Docs](https://docs.docker.com/compose)
- [Python Docker Best Practices](https://docs.docker.com/language/python)
