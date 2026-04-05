# Integração de Package Managers (MPM e LPM)

**Status:** Especificação de Arquitetura
**Versão:** 2.0 - Sem GUI Própria
**Data:** 2026-04-05
**Componentes:** `cmd/mpm/` + `cmd/lpm/` + Setup Installer + App Manager

---

## 1. Visão Geral

**MPM (Model Package Manager)** e **LPM (Llama Package Manager)** são **ferramentas CLI puras** sem interface gráfica própria. Eles são **controlados por**:

1. **Setup Installer** (Fyne GUI) - durante instalação inicial
2. **Vectora App** (Web UI - Manager tab) - durante uso normal

Esta arquitetura permite:
- ✅ CLI simples e testável
- ✅ Integração unificada com Setup
- ✅ Gerenciamento via App sem duplicação
- ✅ Automação via scripts
- ✅ Sem dependências GUI pesadas em MPM/LPM

---

## 2. Arquitetura de Package Managers

### 2.1 Estrutura de Componentes

```
Vectora/
├── cmd/
│   ├── mpm/                    # CLI puro (sem GUI)
│   │   ├── main.go
│   │   ├── commands.go
│   │   └── output.go           # JSON output apenas
│   │
│   ├── lpm/                    # CLI puro (sem GUI)
│   │   ├── main.go
│   │   ├── commands.go
│   │   └── output.go
│   │
│   └── vectora-setup/          # Fyne GUI que controla LPM/MPM
│       ├── main.go
│       ├── wizard.go
│       └── manager.go
│
├── internal/
│   ├── models/                 # Lógica de MPM
│   ├── engines/                # Lógica de LPM
│   │
│   └── app/                    # Vectora App Web UI
│       └── manager/
│           ├── model_controller.go   # Chama MPM CLI
│           └── engine_controller.go  # Chama LPM CLI
│
└── build.ps1                   # Compila: LPM + MPM + Setup + App
```

### 2.2 Padrão de Comunicação

MPM e LPM **não compartilham estado** com Setup ou App. A comunicação é:

```
Setup Installer / App Manager
        ↓
   Chama CLI (spawn process)
        ↓
   mpm.exe / lpm.exe
        ↓
   Executa operação
        ↓
   Escreve JSON em stdout
        ↓
   Process termina
        ↓
   Setup/App parseia resultado
```

**Nunca há servidor rodando**, tudo é **one-shot CLI execution**.

---

## 3. LPM (Llama Package Manager) - Especificação Atualizada

### 3.1 Responsabilidades Principais

| Responsabilidade | Quem Controla | Como |
|-----------------|---------------|------|
| **Listar builds** | App + Setup | `lpm list --json` |
| **Detectar hardware** | App + Setup | `lpm detect --json` |
| **Recomendar build** | App + Setup | `lpm recommend --json` |
| **Baixar build** | App + Setup | `lpm install --id <id>` |
| **Ativar build** | App + Setup | `lpm set-active --id <id>` |
| **Listar instalados** | App + Setup | `lpm list --installed --json` |

### 3.2 CLI Interface (sem GUI)

```bash
# Listagem
lpm list                           # Formato texto (tty)
lpm list --json                    # Formato JSON (para App/Setup)
lpm list --installed               # Apenas instalados

# Hardware
lpm detect                         # Mostrar capabilities
lpm detect --json                  # JSON para parsing

# Recomendação
lpm recommend                      # Texto legível
lpm recommend --json               # JSON com ID + nome

# Instalação
lpm install --id llama-windows-cuda-12
lpm install --id <id> --silent     # Sem output (Setup)
lpm install --id <id> --json       # Output JSON

# Ativação
lpm set-active --id llama-windows-cuda-12
lpm active                         # Mostrar build ativo atual
lpm active --json                  # JSON com info do build ativo

# Info
lpm version                        # Versão do LPM
lpm info --id <id> --json         # Detalhes do build
```

### 3.3 Output JSON Padrão

```json
{
  "success": true,
  "command": "list",
  "data": {
    "builds": [
      {
        "id": "llama-windows-cuda-12",
        "name": "Llama.cpp CUDA 12.0",
        "version": "b3430",
        "os": "windows",
        "arch": "x86_64",
        "gpu": "cuda",
        "gpu_version": "12.0",
        "size_bytes": 2500000000,
        "installed": true,
        "active": true
      }
    ]
  },
  "error": null
}
```

### 3.4 Fluxo de Instalação (via Setup/App)

```
Setup/App: "Instale llama-windows-cuda-12"
    ↓
App executa: lpm install --id llama-windows-cuda-12 --silent
    ↓
LPM:
  1. Validar build existe em catálogo
  2. Detectar local de instalação (~/.Vectora/engines/)
  3. Baixar arquivo .zip
  4. Verificar SHA256
  5. Extrair para ~/.Vectora/engines/llama-<id>/
  6. Criar metadata.json
  7. Output JSON com sucesso
    ↓
App/Setup: parseia JSON e continua
```

---

## 4. MPM (Model Package Manager) - Especificação Atualizada

### 4.1 Responsabilidades Principais

| Responsabilidade | Quem Controla | Como |
|-----------------|---------------|------|
| **Listar modelos** | App + Setup | `mpm list --json` |
| **Detectar hardware** | App + Setup | `mpm detect --json` |
| **Recomendar modelo** | App + Setup | `mpm recommend --json` |
| **Baixar modelo** | App + Setup | `mpm install --model <id>` |
| **Ativar modelo** | App + Setup | `mpm set-active --model <id>` |
| **Listar instalados** | App + Setup | `mpm list --installed --json` |
| **Buscar modelos** | App | `mpm search --query <q> --json` |

### 4.2 CLI Interface (sem GUI)

```bash
# Listagem
mpm list                          # Formato texto (tty)
mpm list --json                   # Formato JSON
mpm list --installed              # Apenas instalados
mpm list --family qwen3           # Filtrar por família

# Hardware
mpm detect                        # Mostrar specs
mpm detect --json                 # JSON para parsing

# Recomendação
mpm recommend                     # Texto legível
mpm recommend --json              # JSON com ID + nome

# Busca
mpm search --query "coding" --json
mpm search --tag "lightweight" --json

# Instalação
mpm install --model qwen3-8b
mpm install --model qwen3-4b --silent
mpm install --model qwen3-0.6b --json

# Ativação
mpm set-active --model qwen3-8b
mpm active                        # Mostrar modelo ativo
mpm active --json                 # JSON com info

# Info
mpm version
mpm info --model qwen3-8b --json

# Versões/quantizações
mpm versions --model qwen3-8b --json
```

### 4.3 Output JSON Padrão

```json
{
  "success": true,
  "command": "list",
  "data": {
    "models": [
      {
        "id": "qwen3-8b",
        "name": "Qwen3 8B",
        "family": "qwen3",
        "size_bytes": 8000000000,
        "huggingface_id": "Qwen/Qwen3-8B-GGUF",
        "installed": true,
        "active": true,
        "capabilities": ["chat", "instruct"],
        "recommended_ram_gb": 16,
        "estimated_speed": "fast"
      }
    ]
  },
  "error": null
}
```

### 4.4 Fluxo de Instalação (via Setup/App)

```
Setup/App: "Instale qwen3-8b"
    ↓
App executa: mpm install --model qwen3-8b --silent
    ↓
MPM:
  1. Validar modelo existe em catálogo
  2. Detectar local de instalação (~/.Vectora/models/)
  3. Baixar GGUF do Hugging Face
  4. Verificar SHA256
  5. Extrair para ~/.Vectora/models/qwen3-8b/
  6. Criar model.json com metadata
  7. Output JSON com sucesso
    ↓
App/Setup: parseia JSON e continua
```

---

## 5. Integração Setup Installer → Package Managers

### 5.1 Fluxo Completo do Setup

```
Usuário executa: vectora-setup.exe
    ↓
[1] Tela de Boas-vindas
    ↓
[2] Hardware Detection
    → Detectar CPU, RAM, GPU
    → Chamar: lpm detect --json
    → Chamar: mpm detect --json
    ↓
[3] Seleção de Llama.cpp
    → Exibir: lpm list --json
    → Permitir seleção ou aceitar recomendação
    → Chamar: lpm recommend --json (se automático)
    ↓
[4] Download de Llama.cpp
    → Executar: lpm install --id <selected> --silent
    → Monitorar progresso
    → Validar sucesso
    ↓
[5] Seleção de Modelo
    → Exibir: mpm list --json
    → Permitir seleção ou aceitar recomendação
    → Chamar: mpm recommend --json (se automático)
    ↓
[6] Download de Modelo
    → Executar: mpm install --model <selected> --silent
    → Monitorar progresso
    → Validar sucesso
    ↓
[7] Finalização
    → Criar atalho do Vectora
    → Iniciar daemon
    → Sucesso!
```

### 5.2 Estrutura do Setup Installer (Fyne)

```go
// cmd/vectora-setup/wizard.go

type SetupWizard struct {
    window fyne.Window

    // Estado atual
    currentStep int
    hardware *Hardware
    selectedBuild string
    selectedModel string

    // Controllers para chamar CLIs
    lpmController *LPMController
    mpmController *MPMController
}

func (w *SetupWizard) StepHardwareDetection() {
    // Executa:
    // - lpm detect --json
    // - mpm detect --json
    // - Parseia resultados
}

func (w *SetupWizard) StepSelectBuild() {
    // Executa:
    // - lpm list --json
    // - Exibe lista em UI
    // - Permite seleção
}

func (w *SetupWizard) StepDownloadBuild() {
    // Executa:
    // - lpm install --id <id> --silent --json
    // - Monitora stdout para progresso
    // - Valida resultado
}

func (w *SetupWizard) StepSelectModel() {
    // Executa:
    // - mpm list --json
    // - Exibe lista em UI
    // - Permite seleção
}

func (w *SetupWizard) StepDownloadModel() {
    // Executa:
    // - mpm install --model <id> --silent --json
    // - Monitora stdout para progresso
    // - Valida resultado
}
```

### 5.3 Controllers para Package Managers

```go
// internal/setup/lpm_controller.go

type LPMController struct {
    lpmPath string  // Caminho do lpm.exe
}

func (c *LPMController) List() ([]Build, error) {
    cmd := exec.Command(c.lpmPath, "list", "--json")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var result struct {
        Data struct {
            Builds []Build `json:"builds"`
        } `json:"data"`
    }
    json.Unmarshal(output, &result)
    return result.Data.Builds, nil
}

func (c *LPMController) Detect() (*Hardware, error) {
    cmd := exec.Command(c.lpmPath, "detect", "--json")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    // Parse JSON...
}

func (c *LPMController) Install(buildID string, onProgress func(percent int)) error {
    cmd := exec.Command(c.lpmPath, "install", "--id", buildID, "--json")

    // Monitora stdout e chama callback com progresso
    // Ex: {"progress": 45, "status": "downloading"}
}
```

---

## 6. Integração App → Package Managers

### 6.1 Manager Tab Architecture

```
App Web UI (Next.js)
    ↓
Manager Tab (React Component)
    ├── Model Manager Section
    │   ├── Listar modelos instalados
    │   ├── Seleção de modelo ativo
    │   ├── Botão "Instalar novo modelo"
    │   └── Busca de modelos
    │
    └── Engine Manager Section
        ├── Listar builds instalados
        ├── Seleção de build ativo
        ├── Botão "Instalar novo build"
        └── Informações de hardware
```

### 6.2 IPC Bridge (App ↔ Daemon)

O App **não chama MPM/LPM diretamente**. Em vez disso, passa requisições via IPC para o daemon, que controla os CLIs:

```
App Web UI (navegador)
    ↓ IPC Request (JSON)
Vectora Daemon (Go)
    ├─ manager.go (Manager Controller)
    │  ├─ HandleModelList()
    │  │  └─ mpm list --json
    │  ├─ HandleModelInstall()
    │  │  └─ mpm install --model <id>
    │  ├─ HandleEngineList()
    │  │  └─ lpm list --json
    │  └─ HandleEngineInstall()
    │     └─ lpm install --id <id>
    ↓ IPC Response (JSON)
App Web UI (navegador)
```

### 6.3 Fluxo de Instalação de Modelo (via App)

```
Usuário clica: "Instalar modelo"
    ↓
Modal de seleção abre
    ↓
App envia IPC: {method: "model.list", json: true}
    ↓
Daemon executa: mpm list --json
    ↓
App recebe lista e exibe em modal
    ↓
Usuário seleciona: "qwen3-8b"
    ↓
App envia IPC: {method: "model.install", model_id: "qwen3-8b"}
    ↓
Daemon executa: mpm install --model qwen3-8b --json
    ↓
Daemon envia eventos de progresso via IPC (streaming)
    ↓
App exibe barra de progresso
    ↓
Instalação completa → Sucesso!
```

### 6.4 Manager Tab Component (React TypeScript)

```typescript
// internal/app/components/Manager.tsx

export function Manager() {
  const [models, setModels] = useState<Model[]>([]);
  const [activeModel, setActiveModel] = useState<string>("");
  const [engines, setEngines] = useState<Engine[]>([]);
  const [activeEngine, setActiveEngine] = useState<string>("");
  const [installing, setInstalling] = useState(false);
  const [progress, setProgress] = useState(0);

  // Carrega lista de modelos ao montar
  useEffect(() => {
    ipcRequest("model.list").then(data => {
      setModels(data.models);
      setActiveModel(data.active);
    });
  }, []);

  // Carrega lista de engines ao montar
  useEffect(() => {
    ipcRequest("engine.list").then(data => {
      setEngines(data.engines);
      setActiveEngine(data.active);
    });
  }, []);

  const handleInstallModel = async (modelId: string) => {
    setInstalling(true);

    // Subscribe para eventos de progresso
    ipcSubscribe("model.install.progress", (event) => {
      setProgress(event.percent);
    });

    // Envia requisição
    await ipcRequest("model.install", { model_id: modelId });

    setInstalling(false);
    setProgress(0);

    // Recarrega lista
    const data = await ipcRequest("model.list");
    setModels(data.models);
    setActiveModel(data.active);
  };

  return (
    <div className="manager">
      <section>
        <h2>Modelos de IA</h2>
        <ModelSelector
          models={models}
          active={activeModel}
          onSelect={setActiveModel}
          onInstall={handleInstallModel}
          installing={installing}
          progress={progress}
        />
      </section>

      <section>
        <h2>Build Llama.cpp</h2>
        <EngineSelector
          engines={engines}
          active={activeEngine}
          onSelect={setActiveEngine}
          onInstall={handleInstallEngine}
        />
      </section>
    </div>
  );
}
```

---

## 7. Build Integration (build.ps1)

### 7.1 Sequência de Compilação

```powershell
# build.ps1

# [1] Clean
# [2] Frontend (Next.js)
# [3] App (Wails)
# [4] Daemon
# [5] Tests

# [6] LPM CLI ← IMPORTANTE: Compilar antes do Setup
Write-Host "[6/11] Building LPM..."
& go build -o ./build/lpm.exe ./cmd/lpm
if ($LASTEXITCODE -ne 0) { exit 1 }

# [7] MPM CLI ← IMPORTANTE: Compilar antes do Setup
Write-Host "[7/11] Building MPM..."
& go build -o ./build/mpm.exe ./cmd/mpm
if ($LASTEXITCODE -ne 0) { exit 1 }

# [8] Setup Installer ← IMPORTANTE: Depois de LPM/MPM
# O Setup embarca LPM e MPM dentro dele
Write-Host "[8/11] Building Setup Installer..."
# Copiar LPM e MPM para embedding
Copy-Item ./build/lpm.exe ./cmd/vectora-setup/embed/
Copy-Item ./build/mpm.exe ./cmd/vectora-setup/embed/

# Compilar Setup (Fyne precisa de resources embarcados)
& go build -o ./build/vectora-setup.exe ./cmd/vectora-setup
if ($LASTEXITCODE -ne 0) { exit 1 }

# [9] CLI
# [10] Tests
# [11] SHA256
```

### 7.2 Embedding no Setup

```go
// cmd/vectora-setup/main.go

import (
    _ "embed"
)

//go:embed embed/lpm.exe
var lpmBinary []byte

//go:embed embed/mpm.exe
var mpmBinary []byte

func init() {
    // Setup desembarca LPM e MPM em temp
    lpmPath := filepath.Join(os.TempDir(), "lpm.exe")
    mpmPath := filepath.Join(os.TempDir(), "mpm.exe")

    os.WriteFile(lpmPath, lpmBinary, 0755)
    os.WriteFile(mpmPath, mpmBinary, 0755)
}
```

---

## 8. Error Handling & Retry Logic

### 8.1 Tratamento de Erros CLI

```json
{
  "success": false,
  "command": "install",
  "error": {
    "code": "download_failed",
    "message": "Falha ao baixar de Hugging Face",
    "details": "Connection timeout after 30s",
    "retry": true
  }
}
```

### 8.2 Retry Policy

- **Transient errors** (timeout, network): Retry até 3x com backoff exponencial
- **Permanent errors** (not found, invalid): Não retry, apenas erro
- **Disk space**: Erro com sugestão de liberar espaço
- **Corruption**: Delete arquivo parcial, retry download

---

## 9. Testing Strategy

### 9.1 Unit Tests para CLI

```bash
go test ./cmd/mpm
go test ./cmd/lpm
go test ./internal/models
go test ./internal/engines
```

### 9.2 Integration Tests

- Setup Installer com LPM/MPM
- App Manager com LPM/MPM
- IPC requests → CLI execution → Response parsing

### 9.3 End-to-End Tests

- Full Setup wizard (LPM install + MPM install)
- Full App Manager workflow (Model switching + Engine switching)

---

## 10. Success Criteria

### Funcionalidade
- ✅ LPM CLI completo (list, detect, recommend, install, active, set-active)
- ✅ MPM CLI completo (list, detect, recommend, install, active, set-active, search)
- ✅ Setup Installer controla ambos via CLI
- ✅ App Manager tab controla ambos via IPC → CLI
- ✅ JSON output parseable
- ✅ Erro handling robusto

### Qualidade
- ✅ Todos os testes passando
- ✅ Sem memory leaks
- ✅ Timeouts implementados
- ✅ Retry logic funcionando

### User Experience
- ✅ Setup wizard fluidez
- ✅ Manager tab responsivo
- ✅ Progresso visual durante downloads
- ✅ Mensagens de erro claras

---

**Versão:** 2.0 - Sem GUI Própria (LPM/MPM = CLI puro)
**Última Atualização:** 2026-04-05
**Status:** Pronto para implementação
