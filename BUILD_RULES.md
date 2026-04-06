# 🔨 Regras de Build - Vectora

## ⚡ Regras Rígidas e Obrigatórias

### 1. **EXECUTÁVEIS APENAS VIA `build.sh`**
```bash
# ✅ CORRETO:
bash build.sh

# ❌ ERRADO:
go build -o bin/vectora ./cmd/daemon
CGO_ENABLED=1 go build -o ./vectora-desktop ./cmd/desktop
```

**Razão:** Garante versioning, padronização de nomes e configuração correta de CGO.

---

### 2. **Estrutura de Nomes - Convenção Obrigatória**

```
{app}-{os}-{arch}{suffix}
```

**Exemplos válidos:**
- `vectora-windows-amd64.exe`
- `vectora-linux-arm64`
- `vectora-macos-amd64`
- `vectora-tui-windows-amd64.exe`
- `vectora-setup-linux-amd64`
- `vectora-desktop-macos-arm64`
- `lpm-windows-amd64.exe`
- `mpm-linux-amd64`

**Componentes:**
- `{app}`: nome do aplicativo (vectora, vectora-tui, vectora-setup, vectora-desktop, lpm, mpm)
- `{os}`: windows, linux, macos
- `{arch}`: amd64, arm64, armv7
- `{suffix}`: `.exe` para Windows, vazio para Unix

---

### 3. **Localização: `bin/` Apenas**

```
Vectora/
├── bin/
│   ├── vectora-windows-amd64.exe
│   ├── vectora-windows-arm64.exe
│   ├── vectora-linux-amd64
│   ├── vectora-linux-arm64
│   ├── vectora-macos-amd64
│   ├── vectora-macos-arm64
│   ├── vectora-tui-windows-amd64.exe
│   ├── vectora-tui-linux-amd64
│   ├── vectora-setup-windows-amd64.exe
│   ├── vectora-setup-linux-amd64
│   ├── vectora-desktop-windows-amd64.exe
│   ├── vectora-desktop-linux-amd64
│   ├── vectora-desktop-macos-amd64
│   ├── vectora-desktop-macos-arm64
│   ├── lpm-windows-amd64.exe
│   ├── lpm-linux-amd64
│   ├── mpm-windows-amd64.exe
│   └── mpm-linux-amd64
```

**Regra:** Nenhum outro lugar. `bin/` é limpado completamente antes de cada build.

---

### 4. **CGO - Apenas para Fyne Apps**

| App | CGO_ENABLED | Razão |
|-----|-------------|-------|
| `vectora` (daemon) | 0 | Go puro |
| `vectora-tui` | 0 | Go puro (Bubbletea) |
| `vectora-setup` | 1 | Fyne GUI |
| `vectora-desktop` | 1 | Fyne GUI |
| `lpm` | 0 | Go puro |
| `mpm` | 0 | Go puro |

**Nota:** Fyne apps (setup, desktop) com ARM64 no Windows não compilam com CGO. CLI tools (lpm, mpm) compilam para ambas arquiteturas.

---

### 5. **Compilação Nativa**

O `build.sh` compila **APENAS para o sistema detectado** (nativo).

```bash
# No macOS compila: macos-amd64 + macos-arm64
# No Linux compila: linux-amd64 + linux-arm64
# No Windows compila: windows-amd64 + windows-arm64 (CLI apenas)
```

---

## 📋 Checklist para Commits

Antes de commitar mudanças no código:

- [ ] **Código compila** sem erros: `bash build.sh`
- [ ] **Todos os binários** estão em `bin/` com nomes corretos
- [ ] **Nenhum binário em outras pastas** (root, src/, cmd/, etc)
- [ ] **E2E tests passam**: `go test ./test/e2e/...`
- [ ] **Commit message** segue o padrão (feat:, fix:, refactor:, etc)
- [ ] **build.sh não foi modificado** sem necessidade

---

## 🚀 Comandos Rápidos

```bash
# Build completo (nativo)
bash build.sh

# Verificar binários gerados
ls -lah bin/

# Limpar e recompilar
rm -rf bin/ && bash build.sh

# Rodar testes
go test ./...

# Rodar E2E tests
go test ./test/e2e/...
```

---

## ⚠️ Anti-Padrões (NÃO FAZER)

```bash
# ❌ NÃO compilar manualmente
go build -o vectora ./cmd/daemon

# ❌ NÃO usar binários fora de bin/
./vectora-daemon  # ERRADO

# ❌ NÃO nomear sem SO-arquitetura
vectora.exe  # ERRADO (deveria ser: vectora-windows-amd64.exe)

# ❌ NÃO usar CGO para CLI tools
CGO_ENABLED=1 go build -o bin/lpm ./cmd/lpm  # ERRADO

# ❌ NÃO colocar binários em root
./vectora  # ERRADO (deveria estar em bin/)
```

---

## 📝 Histórico de Builds

| Commit | Build Script | Nomes | CGO |
|--------|--------------|-------|-----|
| 55d59e7 | Antigo (sufixo -arm64) | Parcial | ✓ |
| 2f52b1a | Antigo (sufixo -arm64) | Parcial | ✓ |
| b2d0f92 | **NOVO ({app}-{os}-{arch})** | **Completo** | ✓ |

---

## 🔄 Fluxo de Desenvolvimento

```
Modify Code
    ↓
Run: bash build.sh
    ↓
Verify: ls -lah bin/
    ↓
Check names: {app}-{os}-{arch}{suffix}
    ↓
Run tests: go test ./...
    ↓
Commit + Push
```

---

**Versão:** 1.0
**Data:** 2026-04-05
**Status:** ✅ Ativo
