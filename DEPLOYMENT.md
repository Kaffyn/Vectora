# 🚀 Deployment Guide - Vectora

Este documento explica como fazer deploy do Vectora para produção, incluindo publicação no PyPI, GHCR, MCP Registry e VPS.

## Pré-requisitos

Você precisa ter acesso ou criar tokens para:

1. **PyPI Token** - Para publicar o pacote Python
2. **GHCR Token** - Para publicar a imagem Docker
3. **VPS SSH Key** - Para fazer deploy no servidor
4. **MCP Registry Token** (opcional) - Para registrar no MCP Registry

---

## 1️⃣ Configurar Secrets no GitHub

### PyPI Token

1. Acesse [pypi.org](https://pypi.org/)
2. Faça login na sua conta
3. Vá para **Account settings** → **API tokens**
4. Clique em **Create token**
5. Escolha o escopo: **Entire account** ou apenas **vectora**
6. Copie o token gerado (formato: `pypi-...`)

**Adicionar ao GitHub:**

1. Vá para seu repositório no GitHub
2. **Settings** → **Secrets and variables** → **Actions**
3. Clique em **New repository secret**
4. Nome: `PYPI_TOKEN`
5. Valor: Cole o token do PyPI
6. Clique em **Add secret**

### GHCR Token (GitHub Container Registry)

1. Vá para **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Clique em **Generate new token (classic)**
3. Escolha os escopos:
   - `write:packages` - Para escrever pacotes
   - `read:packages` - Para ler pacotes
   - `delete:packages` - Para deletar (opcional)
4. Copie o token gerado
5. Adicione ao GitHub como secret `GHCR_TOKEN`

### VPS SSH Key

1. Gere uma chave SSH se ainda não tiver:

   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/vectora-vps -C "vectora-deploy"
   ```

2. Adicione a chave pública ao VPS:

   ```bash
   cat ~/.ssh/vectora-vps.pub | ssh user@vps-host 'cat >> ~/.ssh/authorized_keys'
   ```

3. Adicione o conteúdo da chave privada ao GitHub:
   ```bash
   cat ~/.ssh/vectora-vps
   ```
   - Vá para **Settings** → **Secrets and variables** → **Actions**
   - Adicione como secret `VPS_SSH_KEY`

### Outras Secrets Necessárias

```
VPS_HOST          → IP ou hostname do seu VPS
VPS_USER          → Usuário SSH (ex: root, ubuntu)
VPS_BASE_URL      → URL base do VPS (ex: http://192.168.1.100:8000)
MCP_REGISTRY_TOKEN → Token do MCP Registry (opcional)
```

---

## 2️⃣ Triggar Deploy Automático

O deploy é **automático** quando você:

1. **Faz push para `main`** com mensagem contendo `[deploy]`
2. O GitHub Actions roda a pipeline completa:
   - ✅ Lint & Format
   - ✅ Type Check
   - ✅ Testes (unit, integration, e2e)
   - ✅ Security scan
   - 🐳 Build Docker image
   - 📦 **Publish to PyPI** ← NOVO
   - 🔌 Publish to MCP Registry
   - 🌐 Deploy to VPS

### Exemplo de Commit para Deploy

```bash
git add .
git commit -m "feat: add new feature [deploy]"
git push origin main
```

Ou:

```bash
git commit --amend --no-edit && git push origin main -f
```

---

## 3️⃣ O que Acontece em Cada Stage

### 📦 Publish to PyPI

1. **Extract version** - Lê a versão de `pyproject.toml`
2. **Build package** - Cria distribuição com `uv build`
   - Gera `.whl` (wheel distribution)
   - Gera `.tar.gz` (source distribution)
3. **Publish** - Upload para PyPI usando `twine`
   - Username: `__token__`
   - Password: `${{ secrets.PYPI_TOKEN }}`

**Resultado:**

- Pacote disponível em: https://pypi.org/project/vectora/
- Instalável com: `pip install vectora`

### 🐳 Build & Push Docker

1. Login no GHCR
2. Build da imagem multi-stage
3. Push para: `ghcr.io/kaffyn/vectora:latest` e `ghcr.io/kaffyn/vectora:0.1.0`

**Resultado:**

- Imagem no registry: https://ghcr.io/kaffyn/vectora

### 🌐 Deploy to VPS

1. SSH para o VPS
2. Login no GHCR
3. Pull da imagem latest
4. Para container antigo
5. Inicia novo container com a versão
6. Valida health check

**Resultado:**

- MCP Server rodando no VPS
- Acessível via: `ssh user@vps-host`

---

## 4️⃣ Verificar Deploy

### Verificar PyPI

```bash
pip index versions vectora  # Mostrar todas as versões
pip install vectora --upgrade  # Instalar versão mais recente
vectora --version  # Verificar versão instalada
```

### Verificar GHCR

```bash
docker pull ghcr.io/kaffyn/vectora:latest
docker run -it ghcr.io/kaffyn/vectora:latest vectora --help
```

### Verificar VPS

```bash
ssh user@vps-host
docker ps  # Ver containers
docker logs vectora-mcp  # Ver logs
```

---

## 5️⃣ Troubleshooting

### PyPI Upload Fails

**Erro:** `Invalid or expired authentication token`

- **Solução:** Regenerar token no pypi.org e atualizar secret

**Erro:** `File already exists`

- **Solução:** Incrementar versão em `pyproject.toml` (não pode reusar versão)
- **Exemplo:** `0.1.0` → `0.1.1`

### GHCR Push Fails

**Erro:** `denied: User does not have push access`

- **Solução:** Verificar permissões do GHCR_TOKEN
- **Confirmar:** Escopos `write:packages` estão ativados

### VPS Deploy Fails

**Erro:** `Permission denied (publickey)`

- **Solução:** Verificar VPS_SSH_KEY está correto
- **Confirmar:** Chave pública foi adicionada ao `~/.ssh/authorized_keys`

**Erro:** `docker: command not found`

- **Solução:** Instalar Docker no VPS

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
```

---

## 6️⃣ Fluxo Completo Exemplo

```bash
# 1. Fazer mudanças
echo "# Nova feature" >> README.md
git add .
git commit -m "feat: add documentation [deploy]"

# 2. Push para main (trigga workflow)
git push origin main

# 3. Monitorar workflow
# Ir para: https://github.com/Kaffyn/vectora/actions
# Clicar no último workflow run

# 4. Após sucesso, validar:
pip install --upgrade vectora
docker pull ghcr.io/kaffyn/vectora:latest
ssh user@vps-host docker ps
```

---

## 7️⃣ Versionamento Semântico

Recomendamos seguir [Semantic Versioning](https://semver.org/):

- **0.1.0** → Primeira versão (MVP)
- **0.1.1** → Bug fix
- **0.2.0** → Feature nova
- **1.0.0** → Versão estável (remover "0.")

Atualize em `pyproject.toml`:

```toml
[project]
version = "0.1.1"  # ← Mude aqui
```

---

## 📚 Links Úteis

- [PyPI Upload Docs](https://packaging.python.org/en/latest/tutorials/packaging-projects/)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-container-registry)
- [GitHub Secrets](https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions)
- [UV Build Tool](https://github.com/astral-sh/uv)

---

## ❓ FAQ

**P: Posso fazer deploy sem `[deploy]` tag?**
A: Sim! Remova a condição `contains(github.event.head_commit.message, '[deploy]')` do workflow.

**P: Como fazer rollback se algo der errado?**
A: No PyPI, não é possível remover versões (por segurança). Incremente para nova versão. No VPS, você pode usar `docker pull ghcr.io/kaffyn/vectora:0.1.0` para voltar versão anterior.

**P: Como fazer deploy sem esperar testes?**
A: Não recomendado! Mas você pode comentar os `needs:` do job deploy se realmente precisar.

---

**Última atualização:** 2026-05-14  
**Mantenedor:** @Kaffyn  
**Status:** ✅ Pronto para produção
