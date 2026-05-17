# Deployment e GitHub Actions

Este documento descreve como configurar o CI/CD pipeline do Vectora.

## GitHub Secrets Necessários

Para que o pipeline funcione corretamente, configure os seguintes secrets no repositório GitHub:

### 🔐 Secrets Obrigatórios

1. **`GITHUB_TOKEN`** (automático)

   - Fornecido automaticamente pelo GitHub Actions
   - Usado para: GHCR login, GitHub API

2. **`VPS_SSH_KEY`**

   - Chave privada SSH para conectar no VPS
   - Tipo: RSA private key
   - Comando para gerar:
     ```bash
     ssh-keygen -t rsa -b 4096 -f ~/.ssh/vectora-deploy -C "vectora-deploy"
     ```
   - Adicionar chave pública (`vectora-deploy.pub`) ao `~/.ssh/authorized_keys` do VPS

3. **`VPS_HOST`**

   - Hostname ou IP do VPS
   - Exemplo: `srv1640150.hstgr.cloud`

4. **`VPS_USER`**
   - Usuário SSH do VPS
   - Exemplo: `root` ou `deploy`

### 🌐 Secrets Opcionais (Recomendados)

5. **`VPS_GOOGLE_API_KEY`**

   - Google Gemini API key para LLM
   - Obtém em: https://aistudio.google.com/app/apikey

6. **`VPS_COHERE_API_KEY`**

   - Cohere API key para embeddings e reranking
   - Obtém em: https://dashboard.cohere.com/api-keys

7. **`VPS_LANGSMITH_API_KEY`**

   - LangSmith API key para rastreamento
   - Obtém em: https://smith.langchain.com

8. **`GHCR_TOKEN`**

   - Token Classic do GitHub para push em GHCR
   - Pode ser o mesmo que `GITHUB_TOKEN` ou um token específico
   - Crie em: https://github.com/settings/tokens
   - Permissões necessárias:
     - `write:packages` (push Docker images)
     - `read:packages` (pull Docker images)

9. **`MCP_REGISTRY_TOKEN`** (se usar registro MCP)
   - Token para registrar servidor no MCP Registry oficial
   - Obtém em: https://registry.modelcontextprotocol.io

## Como Configurar Secrets

### Via CLI (gh)

```bash
# GitHub Token
gh secret set GITHUB_TOKEN --body "$GITHUB_TOKEN"

# VPS SSH Key (arquivo)
gh secret set VPS_SSH_KEY --body-file ~/.ssh/vectora-deploy

# Outros secrets
gh secret set VPS_HOST -b "srv1640150.hstgr.cloud"
gh secret set VPS_USER -b "root"
gh secret set VPS_GOOGLE_API_KEY -b "seu-api-key..."
```

### Via GitHub Web

1. Vá para: **Settings** → **Secrets and variables** → **Actions**
2. Clique em **"New repository secret"**
3. Adicione cada secret com seu nome e valor

## CI/CD Pipeline

### Fluxo Normal (Sem `[deploy]`)

Quando você faz push para `main`, `develop` ou `dev`:

1. ✅ **Setup** - Instala dependências Python
2. ✅ **Lint** - Executa Ruff, Isort
3. ✅ **Type Check** - Executa Mypy
4. ✅ **Unit Tests** - Testes unitários
5. ✅ **Integration Tests** - Testes de integração
6. ✅ **E2E Tests** - Testes MCP Resources
7. ✅ **Security** - Verificação de segurança
8. ✅ **Report** - Relatório do pipeline

**Resultado:** Branch protegido até tudo passar

### Deploy Automático (`[deploy]`)

Quando você faz commit com `[deploy]` no título:

```bash
git commit -m "feat: nova feature [deploy]"
git push
```

O pipeline executa:

1. ✅ Todos os passos acima
2. 🐳 **Docker Build & Push** - Build imagem Docker
3. 📢 **Publish MCP** - Publica no MCP Registry
4. 🚀 **Deploy VPS** - Faz deploy automático
5. 🔍 **Health Check** - Verifica saúde do serviço

## Estrutura do VPS

O deploy assume a seguinte estrutura:

```
/var/lib/vectora/
├── config.toml           # Configurações
├── .env                  # Variáveis (via secrets)
├── data/
│   ├── vectora.db        # Banco de dados
│   ├── lancedb/          # Vector store
│   └── embedding_queue.db
└── logs/                 # Logs estruturados
```

## Exemplo de Deploy

### 1. Preparar commit com [deploy]

```bash
# Fazer mudanças no código
git add .
git commit -m "fix: resolver issue crítica [deploy]"
```

### 2. Push dispara pipeline

```bash
git push origin main
```

### 3. GitHub Actions executa

- Testes rodam automaticamente
- Docker image é buildada
- Imagem é pushada para GHCR
- Deploy acontece no VPS

### 4. Verificar logs

```bash
# Ver logs do pipeline
gh run list --limit 5

# Ver detalhes de um run
gh run view <run-id> --log
```

## Troubleshooting

### ❌ Docker Push Falha

**Verificar:** Credenciais GHCR e permissões do token

```bash
# Testar login local
echo $GITHUB_TOKEN | docker login ghcr.io -u Kaffyn --password-stdin

# Verificar imagens
docker image ls
```

### ❌ Deploy SSH Falha

**Verificar:** SSH key e host accessibility

```bash
# Testar conexão SSH local
ssh -i ~/.ssh/vectora-deploy root@$VPS_HOST "whoami"
```

### ❌ VPS Container não inicia

**Verificar:** Logs do container

```bash
# SSH no VPS
ssh root@$VPS_HOST

# Ver logs
docker logs vectora-mcp

# Verificar se está rodando
docker ps | grep vectora
```

## Monitoramento

### Health Check do Serviço

```bash
# Verificar se MCP server está respondendo
ssh root@$VPS_HOST "docker exec vectora-mcp curl -f http://localhost:8000/health"
```

### Logs em Tempo Real

```bash
# SSH no VPS
ssh root@$VPS_HOST

# Ver logs do container
docker logs -f vectora-mcp

# Ver logs estruturados
tail -f /var/lib/vectora/logs/vectora.log
```

## Rollback Manual

Se o deploy der problema:

```bash
ssh root@$VPS_HOST bash -s << 'EOF'
# Parar container atual
docker stop vectora-mcp
docker rm vectora-mcp

# Usar versão anterior
docker run -d \
  --name vectora-mcp \
  --restart unless-stopped \
  -e GOOGLE_API_KEY=$GOOGLE_API_KEY \
  -v /var/lib/vectora:/root/.vectora \
  ghcr.io/kaffyn/vectora:vX.X.X \
  vectora mcp-server

# Verificar
docker ps | grep vectora
EOF
```

## Próximos Passos

- [ ] Configurar secrets no GitHub
- [ ] Testificar pipeline em `develop` primeiro
- [ ] Fazer primeiro deploy em `main` com `[deploy]`
- [ ] Configurar monitoring/alertas
- [ ] Documentar runbooks de operação

---

**Última atualização:** 2026-05-14
