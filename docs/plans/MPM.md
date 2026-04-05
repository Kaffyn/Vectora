# Model Package Manager (MPM) - Especificação

**Status:** Especificação de Design
**Versão:** 1.0
**Data:** 05/04/2026
**Componente:** `cmd/mpm/` + `internal/models/`

---

## Visão Geral

O **MPM (Model Package Manager)** é uma ferramenta CLI autônoma para baixar, gerenciar e alternar entre modelos de IA compatíveis com o Vectora. Ele lida com modelos no formato GGUF do Hugging Face, com suporte inicial para a família de modelos Qwen3.

Assim como o **LPM** (Llama Package Manager), o MPM é:

- Um **executável separado** (`mpm.exe`)
- **Independente** - funciona de forma autônoma ou chamado pelo instalador/daemon
- **Extensível** - suporte futuro para Llama, Mistral, etc.
- **Scriptável** - gera saídas JSON legíveis por máquina para automação

---

## Arquitetura

### 1.1 Estrutura de Componentes

```
cmd/mpm/
├── main.go              # Ponto de entrada CLI, subcomandos
├── commands.go          # list, install, active, search
└── version.go           # informações de versão

internal/models/
├── types.go             # Definições de Model, Catalog, Hardware
├── catalog.go           # Carregamento do catálogo de modelos (embutido)
├── detector.go          # Especificações de hardware + detecção de RAM
├── downloader.go        # Downloader de GGUF do Hugging Face
├── integrity.go         # Verificação SHA256
├── manager.go           # Padrão EngineManager → ModelManager
├── search.go            # Busca semântica no catálogo de modelos
├── catalog.json         # Definições de modelos embutidas (80+ modelos)
├── manager_test.go      # Testes unitários
└── integration_test.go  # Testes de fluxo completo
```

### 1.2 Layout de Dados

```
%USERPROFILE%/.Vectora/
├── models/
│   ├── catalog.json              # Metadados do catálogo em cache
│   ├── metadata.json             # Informações do modelo ativo + versões
│   ├── qwen3-0.6b/
│   │   ├── qwen3-0.6b.gguf       # Pesos do modelo (~0.6GB)
│   │   ├── qwen3-0.6b.gguf.sha256 # Verificação de integridade
│   │   └── model.json            # Metadados do modelo
│   ├── qwen3-4b/
│   │   ├── qwen3-4b.gguf         # Pesos do modelo (~4GB)
│   │   ├── qwen3-4b.gguf.sha256  # Verificação de integridade
│   │   └── model.json
│   └── mistral-7b/               # Futuro
│       └── ...
├── engines/                       # LPM gerencia estes
│   └── llama-cpp-b3430/
└── backups/
```

---

## 2. Recursos Principais

### 2.1 Catálogo de Modelos

**Embutido em tempo de compilação** via `go:embed`, nunca requer busca na rede:

```json
{
  "models": [
    {
      "id": "qwen3-7b",
      "name": "Qwen3 7B",
      "family": "qwen3",
      "version": "1.0",
      "huggingface_id": "Qwen/Qwen3-7B-GGUF",
      "filename": "qwen3-7b-q6_k.gguf",
      "url": "https://huggingface.co/Qwen/Qwen3-7B-GGUF/...",
      "sha256": "abc123...",
      "size_bytes": 7000000000,
      "quantization": "Q6_K",
      "context_length": 8192,
      "parameters": 7000000000,
      "hardware_requirements": {
        "min_ram_gb": 8,
        "recommended_ram_gb": 16,
        "vram_gb": 0,
        "supports_cpu": true,
        "supports_gpu": ["cuda", "metal"]
      },
      "capabilities": ["chat", "instruct", "reasoning"],
      "tags": ["general-purpose", "qwen3-latest"],
      "release_date": "2025-04-01",
      "description": "Modelo Qwen3 7B para seguimento de instruções"
    },
    {
      "id": "qwen3-1.7b",
      "name": "Qwen3 1.7B",
      "family": "qwen3",
      "version": "1.0",
      "huggingface_id": "Qwen/Qwen3-1.7B-GGUF",
      "filename": "qwen3-1.7b-q4_k_m.gguf",
      "url": "https://huggingface.co/Qwen/Qwen3-1.7B-GGUF",
      "sha256": "def456...",
      "size_bytes": 1700000000,
      "quantization": "Q4_K_M",
      "context_length": 8192,
      "parameters": 1700000000,
      "hardware_requirements": {
        "min_ram_gb": 4,
        "recommended_ram_gb": 8,
        "vram_gb": 0,
        "supports_cpu": true,
        "supports_gpu": ["cuda", "metal"]
      },
      "capabilities": ["chat", "instruct"],
      "tags": ["general-purpose", "lightweight", "qwen3"],
      "release_date": "2025-04-01",
      "description": "Qwen3 1.7B - modelo leve para seguimento de instruções"
    }
  ]
}
```

**Suporte Inicial de Modelos (Abril 2026):**

| Modelo              | Tamanhos           | Quantizações         | Status        |
| ------------------- | ------------------ | -------------------- | ------------- |
| **Qwen3**           | 0.6B, 1.7B, 4B, 8B | Q4_K_M, Q5_K_M, Q6_K | ✅ Suportado  |
| **Qwen3-Embedding** | 0.6B, 4B           | Q4_K_M, Q5_K_M, Q6_K | ✅ Suportado  |
| Qwen3-Coder-Next    | 32B, 80B           | Q4_K_M, Q6_K         | ✅ Roadmap Q2 |
| Qwen3-VL            | 2B, 8B             | Q4_K_M, Q5_K_M       | ✅ Roadmap Q3 |
| Llama3.2            | 3B, 8B, 70B        | Q4_K_M, Q6_K         | 🔄 Futuro     |
| Mistral             | 7B                 | Q4_K_M, Q6_K         | 🔄 Futuro     |

### 2.2 Integração de Detecção de Hardware

Utiliza o `internal/engines/detector.go` já construído para o LPM:

```go
func DetectHardware() (*Hardware, error) {
    hw := &Hardware{
        OS:            runtime.GOOS,       // windows, linux, darwin
        Architecture:  runtime.GOARCH,     // amd64, arm64
        CPUCores:      runtime.NumCPU(),
        RAM:           detectTotalRAM(),   // GB
        GPUType:       detectGPU(),        // cuda, metal, none
        GPUVRAM:       detectVRAM(),
        CPUFeatures:   detectCPUFeatures(), // AVX2, AVX512, etc
    }
    return hw, nil
}
```

**Retorno:**

```json
{
  "os": "windows",
  "architecture": "amd64",
  "cpu_cores": 8,
  "ram_gb": 16,
  "gpu_type": "cuda",
  "gpu_vram_gb": 8,
  "cpu_features": ["AVX2", "SSE4.2"]
}
```

### 2.3 Algoritmo de Recomendação de Modelos

```go
func RecommendModel(hw *Hardware) (*Model, error) {
    // Estratégia de 4 níveis:
    // 1. Correspondência exata (RAM cabe, GPU compatível)
    // 2. Melhor modelo menor que caiba na RAM
    // 3. Maior modelo que caiba na RAM disponível
    // 4. Fallback para o Qwen3-0.6B (sempre cabe)

    availableRAM := hw.RAM - 2.0 // Reserva 2GB para o sistema

    // Nível 1: Ajuste perfeito
    for _, m := range catalog.Models {
        if m.Requirements.RecommendedRAM <= availableRAM {
            if hardwareMatches(hw, m) {
                return m, nil
            }
        }
    }

    // Nível 2: Fallback progressivamente menor
    // ...
    return fallbackModel, nil
}
```

### 2.4 Download e Integridade

**Recursos:**

- Suporte a retomada via arquivos `.partial`
- Verificação SHA256 antes da extração
- Callbacks de progresso
- Backoff exponencial (3 tentativas)
- Limitação de largura de banda (opcional)

```go
type DownloadProgress struct {
    Current int64        // Bytes baixados
    Total   int64        // Total de bytes
    Speed   float64      // MB/s
    ETA     time.Duration
}

func (m *ModelManager) Download(ctx context.Context, modelID string,
    onProgress func(*DownloadProgress) error) error
```

---

## 3. Interface CLI

### 3.1 Subcomandos

#### `mpm list`

Lista todos os modelos disponíveis com filtragem.

```bash
$ mpm list
$ mpm list --family qwen3
$ mpm list --filter "7b|1.5b"
$ mpm list --json
```

Saída:

```
ID                    Family     Size  RAM    GPU    Status
qwen3-8b              qwen3      8B    8GB    ✅     Disponível
qwen3-4b              qwen3      4B    6GB    ✅     Disponível
qwen3-1.7b            qwen3      1.7B  4GB    ✅     Disponível
qwen3-0.6b            qwen3      0.6B  2GB    ✅     Disponível
```

#### `mpm detect`

Detecta o hardware local e suas capacidades.

```bash
$ mpm detect
```

Saída:

```
System Hardware:
  OS:              windows
  Architecture:    amd64
  CPU Cores:       8
  RAM:             16.00 GB
  GPU Type:        cuda (NVIDIA)
  GPU VRAM:        8.00 GB
  CPU Features:    AVX2, SSE4.2
```

#### `mpm recommend`

Recomenda o melhor modelo para o hardware detectado.

```bash
$ mpm recommend
$ mpm recommend --json
$ mpm recommend --size 7b  # Força um tamanho específico
```

Saída:

```
Recommended Model: qwen3-7b
Reason: Perfect fit for 16GB RAM + CUDA support
```

Ou JSON:

```json
{
  "model_id": "qwen3-7b",
  "reason": "Perfect fit for 16GB RAM + CUDA support",
  "requires_ram_gb": 8,
  "size_bytes": 7000000000
}
```

#### `mpm install`

Baixa e instala um modelo.

```bash
$ mpm install --model qwen3-8b
$ mpm install --model qwen3-4b --silent
$ mpm install --model $(mpm recommend)
```

Saída:

```
Installing: qwen3-7b
Downloading: ████████████░░░░░░ 65% (4.5GB / 7GB) @ 25.3 MB/s
ETA: 2m 15s
✓ Download completo
✓ SHA256 verificado
✓ Modelo instalado em ~/.Vectora/models/qwen3-7b/
```

#### `mpm active`

Mostra o modelo atualmente ativo.

```bash
$ mpm active
$ mpm active --json
```

Saída:

```
Active Model: qwen3-7b
Path:         ~/.Vectora/models/qwen3-7b/qwen3-7b-q6_k.gguf
Size:         7.0 GB
RAM Required: 8 GB
Installed:    2026-04-05 14:30:45
```

#### `mpm set-active`

Alterna para um modelo instalado diferente.

```bash
$ mpm set-active --model qwen3-4b
```

#### `mpm search`

Busca modelos por nome, capacidade ou tag.

```bash
$ mpm search "coding"
$ mpm search --tag "lightweight"
$ mpm search --capability "instruct"
```

Saída:

```
Found 4 models:
1. qwen3-8b (8B) - General purpose model
2. qwen3-4b (4B) - Lightweight, instruction-following
3. qwen3-1.7b (1.7B) - Very lightweight
4. qwen3-coder-next (80B) - Specialized for code generation [Roadmap]
```

#### `mpm update-catalog`

Atualiza o catálogo embutido (verifica no Hugging Face).

```bash
$ mpm update-catalog
✓ Catalog updated. Now 85 models available.
```

#### `mpm versions`

Lista as quantizações disponíveis para um modelo.

```bash
$ mpm versions qwen3-7b
```

Saída:

```
Quantizations for qwen3-7b:
- Q4_K_M  3.8 GB
- Q5_K_M  4.5 GB
- Q6_K    5.2 GB
- Q8_0    6.8 GB
```

### 3.2 Flags Globais

```
--json              Saída em JSON legível por máquina
--silent            Suprime a saída de progresso (para automação)
--log-level DEBUG   Define o nível de log
--timeout 3600      Tempo limite em segundos para downloads
--threads 4         Threads de download paralelas
```

---

## 4. Integração com o Ecossistema Vectora

### 4.1 Integração com o Instalador

Durante o `vectora-setup.exe`:

```
1. Detecta o hardware (compartilhado com o LPM)
   └─ Chamada: internal/engines.DetectHardware()

2. Recomenda a build do llama.cpp
   └─ Chamada: lpm.recommend()

3. Recomenda o modelo
   └─ Chamada: mpm.recommend()

4. Usuário seleciona o provedor
   ├─ Gemini → Pular instalação do modelo
   └─ Qwen Local
       ├─ Baixa a build do llama.cpp via LPM
       └─ Baixa o modelo via MPM

5. Armazena o modelo ativo em ~/.Vectora/models/metadata.json
```

### 4.2 Integração com o Daemon

O daemon (`vectora.exe`) chama o MPM programaticamente:

```go
import "github.com/Kaffyn/Vectora/internal/models"

// Em llm/qwen.go ou similar
manager, err := models.NewModelManager()
if err != nil {
    return err
}

// Verifica se o modelo está instalado
info, err := manager.GetActive()
if err != nil {
    // Auto-download do modelo recomendado
    hw, _ := detectHardware()
    model, _ := manager.RecommendModel(hw)
    manager.Install(ctx, model.ID)
}

// Usa o caminho do modelo para o llama-cli
modelPath := manager.GetModelPath(modelID)
// → /Users/bruno/.Vectora/models/qwen3-7b/qwen3-7b-q6_k.gguf
```

### 4.3 Integração CLI

Quando o usuário digita o comando `mpm`:

```bash
vectora chat --model qwen3-8b
# Ou
vectora chat --model-list
# Ou
vectora chat --detect  # Auto-detecta e recomenda
```

---

## 5. Fases de Implementação

### Fase 1: MVP (Semana 1-2)

- ✅ Estrutura central da CLI (`cmd/mpm/main.go`)
- ✅ Catálogo de modelos com Qwen3/Qwen3.5 (20+ modelos)
- ✅ Detecção de hardware (reutilizada do LPM)
- ✅ Download + verificação SHA256
- ✅ Subcomandos `list`, `detect`, `recommend`, `install`, `active`
- ✅ Testes básicos

### Fase 2: Polimento (Semana 2-3)

- Saída JSON para todos os comandos
- Funcionalidade de busca e filtragem
- Melhor feedback de progresso
- Limitação de largura de banda
- Testes de integração com o daemon
- Documentação

### Fase 3: Expansão (Semana 4+)

- Suporte ao Llama3.2
- Suporte ao Mistral
- Fontes de catálogo personalizadas (auto-hospedadas)
- Conversão de quantização de modelos (GGML → GGUF)
- Cache de metadados de modelos

---

## 6. Estrutura de Arquivos

```
cmd/mpm/
├── main.go
│   └── parseSubcommands()
│   └── main()
│
├── commands.go
│   ├── cmdList(args)
│   ├── cmdDetect(args)
│   ├── cmdRecommend(args)
│   ├── cmdInstall(args)
│   ├── cmdActive(args)
│   ├── cmdSetActive(args)
│   ├── cmdSearch(args)
│   └── cmdUpdateCatalog(args)
│
└── version.go
    └── printVersion()

internal/models/
├── types.go
│   ├── Model struct
│   ├── Catalog struct
│   ├── Hardware struct
│   ├── ModelManager struct
│   └── DownloadProgress struct
│
├── catalog.go
│   ├── LoadCatalog()
│   ├── GetCatalog()
│   ├── FindModel(id)
│   └── SearchModels(query)
│
├── detector.go
│   ├── DetectHardware()
│   ├── detectRAM()
│   └── detectGPU()
│
├── downloader.go
│   ├── NewDownloader()
│   ├── Download(ctx, url, dst)
│   └── resumeDownload()
│
├── integrity.go
│   ├── VerifyFile(path, sha256)
│   └── ComputeSHA256(path)
│
├── manager.go
│   ├── NewModelManager()
│   ├── Install(ctx, modelID, callback)
│   ├── GetActive()
│   ├── SetActive(modelID)
│   ├── GetModelPath(modelID)
│   ├── RecommendModel(hw)
│   └── ListInstalled()
│
├── search.go
│   ├── SearchByTag(tag)
│   ├── SearchByCapability(cap)
│   └── SearchByName(query)
│
├── catalog.json
├── manager_test.go
└── integration_test.go
```

---

## 7. Integração de Build

### 7.1 Alterações no build.ps1

```powershell
# Passo 6: Build MPM
Write-Host "[6/11] Compilando MPM - Model Package Manager (cmd/mpm)..."
go build -ldflags="-s -w" -o mpm.exe ./cmd/mpm
if (-not (Test-Path "mpm.exe")) { throw "FALHA: mpm.exe não foi gerado." }

# Passo 7: LPM (inalterado)

# Passo 8: Sincronizar com o instalador
Copy-Item "mpm.exe" "cmd/vectora-installer/mpm.exe" -Force

# Passo 9: Build do instalador com mpm.exe embutido
```

### 7.2 Embutimento no Instalador

```go
// cmd/vectora-installer/embed_windows.go
//go:embed mpm.exe
var mpmExe []byte

// Extrair durante o setup
for _, binName := range []string{"vectora.exe", "mpm.exe", "lpm.exe"} {
    if binData, ok := assets[binName]; ok {
        os.WriteFile(filepath.Join(installPath, binName), binData, 0755)
    }
}
```

---

## 8. Dependências

**Novas importações Go:**

```go
import (
    "net/http"              // Downloads HTTP
    "io"                    // Operações de I/O
    "crypto/sha256"         // Verificação de hash
    "os"                    // Operações de arquivo
    "path/filepath"         // Utilitários de caminho
)
```

**Reutilizado de `internal/engines`:**

```go
import (
    "github.com/Kaffyn/Vectora/internal/engines"
    // Reúso: DetectHardware, padrão DownloadProgress, utilitários SHA256
)
```

---

## 9. Estratégia de Testes

### Testes Unitários (`manager_test.go`)

- `TestLoadCatalog` - Verifica o carregamento do catálogo
- `TestDetectHardware` - Verifica a detecção de hardware
- `TestRecommendModel` - 4+ casos de teste para a lógica de recomendação
- `TestVerifyFile` - Validação SHA256
- `TestPaths` - Resolução de diretórios

### Testes de Integração (`integration_test.go`)

- `TestFullInstallationFlow` - Download + verificação + extração
- `TestModelSwitching` - Instalar múltiplos, alternar entre eles
- `TestDownloadResume` - Retomar download interrompido
- `TestSearchFiltering` - Busca e filtragem de modelos

---

## 10. Roadmap Futuro

**Q2 2026:**

- Suporte ao Qwen3-Coder
- Instalação em lote de múltiplos modelos
- Verificações de atualização automática de modelos

**Q3 2026:**

- Suporte ao Qwen3-VL (Visão)
- Modelos Llama3.2
- Modelos Mistral

**Q4 2026:**

- Fontes de catálogo auto-hospedadas
- Ferramentas de quantização de modelos
- Registro de modelos da comunidade

---

## 11. Tratamento de Erros

**Erros de Download:**

```
✗ Download falhou: Tempo limite de conexão
  Tentando novamente em 5 segundos... (Tentativa 2/3)

✗ Incompatibilidade SHA256
  Esperado: abc123...
  Obtido:   def456...
  Arquivo corrompido ou incompleto. Excluindo...
```

**Espaço em Disco:**

```
✗ Espaço em disco insuficiente
  Necessário: 7.0 GB
  Disponível: 2.3 GB
  Por favor, libere ~5 GB e tente novamente
```

**Rede:**

```
✗ Hugging Face inacessível
  Usando catálogo em cache de: 04/04/2026
  Execute 'mpm update-catalog' para atualizar quando estiver online
```

---

## 12. Fluxo de Exemplo de Uso

```bash
# Usuário de primeira viagem
$ mpm detect
$ mpm recommend
  → Recomenda: qwen3-4b

$ mpm install --model qwen3-4b
  → Baixa 4 GB
  → Verifica SHA256
  → Extrai para ~/.Vectora/models/

# Alternar modelos
$ mpm list
$ mpm set-active --model qwen3-7b

# Usar com Vectora
$ vectora          # Daemon inicia, usa o modelo ativo
$ vectora-app      # Interface Web
$ vectora-cli      # Interface de Terminal

# Automação
$ mpm list --json | jq '.models[] | select(.family=="qwen3")'
$ mpm recommend --json | jq '.model_id'
$ mpm install --model $(mpm recommend) --silent
```

---

## 13. Caso de Negócio e Motivação

### Por que o MPM?

Atualmente, os usuários do Vectora têm dois desafios:

1. **Gerenciamento Manual de Modelos**: Sem o MPM, os usuários devem:
   - Visitar manualment o Hugging Face e encontrar modelos GGUF
   - Baixar arquivos grandes (~1-8GB) manualmente
   - Verificar hashes SHA256 à mão
   - Extrair arquivos para os diretórios corretos
   - Atualizar manualmente a configuração
   - Reiniciar manualmente o daemon do Vectora

2. **Sem Conhecimento de Hardware**: Os usuários não sabem qual modelo se ajusta ao seu hardware:
   - 16GB RAM + CUDA → Deve usar Qwen3-8B
   - 8GB RAM + apenas CPU → Deve usar Qwen3-4B
   - 4GB RAM → Deve usar Qwen3-1.7B

### Com o MPM (Futuro)

Os usuários simplesmente executam:

```bash
mpm recommend              # Obtém recomendação baseada no hardware
mpm install --model qwen3-8b   # Baixa, verifica, extrai
```

**Impacto:**

- ⏱️ Tempo: 30 minutos (manual) → 2 minutos (MPM)
- ✅ Automação: Repetível, scriptável, confiável
- 🔒 Segurança: Verificação automática de SHA256
- 🎯 Inteligência: Recomendações conscientes do hardware

### Comparação

| Aspecto             | Sem MPM                 | Com MPM                |
| ------------------- | ----------------------- | ---------------------- |
| **Tempo**           | 30 min + downloads      | 2 min + downloads      |
| **Verificação**     | Manual ou ignorada      | Automática             |
| **Recomendações**   | Usuário adivinha        | Consciente do hardware |
| **Troca de Modelo** | Edição manual de config | `mpm set-active`       |
| **Repetibilidade**  | Propenso a erros        | Automatizado           |

---

## 14. Detalhes de Suporte de Modelos

### Família Qwen3 (Fase 1 - MVP)

**Modelos de Chat Qwen3**
| Modelo | Tamanho | RAM | GPU | Caso de Uso |
|--------|---------|-----|-----|-------------|
| Qwen3-0.6B | 600MB | 2GB | Opcional | Ultra-leve, IoT, embutidos |
| Qwen3-1.7B | 1.7GB | 4GB | Opcional | Mobile, recursos limitados |
| **Qwen3-4B** | **4GB** | **8GB** | **Recomendado** | **Propósito geral, equilibrado** |
| Qwen3-8B | 8GB | 16GB | Recomendado | Alta performance, raciocínio |

**Modelos de Embedding Qwen3**
| Modelo | Tamanho | Caso de Uso |
|--------|---------|-------------|
| Qwen3-Embedding-0.6B | 600MB | Embeddings leves |
| Qwen3-Embedding-4B | 4GB | Embeddings de alta qualidade |

**Qwen3-Coder-Next (Fase 2 - Q2 2026)**
| Modelo | Tamanho | Caso de Uso |
|--------|---------|-------------|
| Qwen3-Coder-Next | 32B-80B | Geração de código e raciocínio |

### Detecção e Recomendação de Hardware

O MPM detecta automaticamente:

- Núcleos de CPU e recursos (AVX, AVX2, NEON, etc.)
- Total de RAM disponível
- Tipo de GPU e VRAM (CUDA, Metal, Vulkan)
- Espaço livre em disco

**Algoritmo de Recomendação de 4 Níveis:**

1. **Ajuste perfeito**: Modelo com RAM recomendada <= RAM disponível
2. **Melhor menor**: Maior modelo que ainda se ajusta
3. **Maior possível**: Modelo máximo que não trava o sistema
4. **Fallback**: Sempre recomenda Qwen3-0.6B (pegada de 2GB)

**Exemplos:**

- 32GB RAM + CUDA 12.0 → Qwen3-8B (recomendado)
- 16GB RAM + apenas CPU → Qwen3-4B
- 8GB RAM → Qwen3-1.7B
- 4GB RAM → Qwen3-0.6B

---

## 15. Lista de Verificação de Implementação

### Fase 1: Configuração da Estrutura de Pacotes

- [ ] Criar diretório `cmd/mpm/`
- [ ] Criar diretório `internal/models/`
- [ ] Criar arquivos stub vazios para todos os módulos
- [ ] Configurar importações de módulos Go
- [ ] Criar alvos no Makefile

**Tempo estimado:** 1 hora

### Fase 2: Tipos Principais e Estruturas de Dados

- [ ] `types.go` (60 linhas)
  - Struct `Model` com todos os metadados
  - Struct `Catalog` para lista de modelos
  - Struct `Hardware` para informações do sistema
  - Interface `ModelManager`
  - Enum de códigos de erro

- [ ] `catalog.go` (100 linhas)
  - `LoadCatalog()` de JSON embutido
  - Singleton `GetCatalog()`
  - Busca `FindModel(id)`
  - Busca `SearchModels(query)`
  - Diretiva `go:embed`

**Tempo estimado:** 2 horas

### Fase 3: Detecção de Hardware

- [ ] `detector.go` (50 linhas)
  - Importar `internal/engines.DetectHardware()`
  - Wrapper para detecção de modelos
  - Testar com diferentes perfis de hardware

**Tempo estimado:** 1 hora

### Fase 4: Download e Integridade

- [ ] `downloader.go` (160 linhas)
  - `NewDownloader()` com timeout
  - `Download()` com suporte a retomada
  - Tratamento de arquivo `.partial`
  - Retry de backoff exponencial (3x)
  - Callbacks de progresso
  - Limitação de largura de banda (opcional)

- [ ] `integrity.go` (60 linhas)
  - Verificação SHA256 `VerifyFile()`
  - Computação de hash `ComputeFileSHA256()`
  - Reutilizar código criptográfico existente

**Tempo estimado:** 3 horas

### Fase 5: Gerenciador e Orquestração

- [ ] `manager.go` (180 linhas) - **Mais crítico**
  - Singleton `NewModelManager()`
  - Orquestrador `Install(ctx, modelID, onProgress)`
  - `GetActive()` ler metadata.json
  - `SetActive(modelID)` atualizar metadados
  - `GetModelPath(modelID)` resolver caminho completo
  - Algoritmo de 4 níveis `RecommendModel(hw)`
  - Escanear diretório `ListInstalled()`

**Tempo estimado:** 4 horas

### Fase 6: Busca e Filtragem

- [ ] `search.go` (80 linhas)
  - Filtrar por tag `SearchByTag(tag)`
  - Correspondência de capacidade `SearchByCapability(cap)`
  - Correspondência de substring `SearchByName(query)`
  - Combinar filtros com lógica AND/OR

**Tempo estimado:** 1,5 horas

### Fase 7: Implementação da CLI

- [ ] `main.go` (100 linhas)
  - Analisar subcomandos
  - Flags globais (--json, --silent, --log-level)
  - Roteamento para manipuladores de comando
  - Informações de versão
  - Tratamento de erros

- [ ] `commands.go` (200 linhas)
  - `cmdList()` - lista com filtros
  - `cmdDetect()` - mostrar hardware
  - `cmdRecommend()` - recomendar modelo
  - `cmdInstall()` - baixar e instalar
  - `cmdActive()` - mostrar ativo
  - `cmdSetActive()` - alternar modelos
  - `cmdSearch()` - buscar catálogo
  - `cmdVersions()` - lista de quantização
  - Pretty-print + saída JSON

**Tempo estimado:** 3 horas

### Fase 8: Testes

- [ ] `manager_test.go` (150 linhas)
  - `TestLoadCatalog` - verificação de embutimento
  - `TestDetectHardware` - detecção de hardware
  - `TestRecommendModel` - 4 cenários
  - `TestVerifyFile` - validação SHA256
  - `TestSearchModels` - funcionalidade de busca

- [ ] `integration_test.go` (100 linhas)
  - `TestFullInstallationFlow` - ponta a ponta
  - `TestModelSwitching` - fluxo set-active
  - `TestDownloadResume` - lógica de retry
  - Testes baseados em tabela

**Tempo estimado:** 2,5 horas

### Fase 9: Build e Integração

- [ ] Atualizar `build.ps1`
  - Adicionar passo [7/11]: `go build -o mpm.exe ./cmd/mpm`
  - Copiar `mpm.exe` para o diretório do instalador antes do embutimento
  - Gerar SHA256 para mpm.exe
  - Atualizar saída de progresso

- [ ] Atualizar `embed_windows.go`
  - Adicionar `//go:embed mpm.exe`
  - Adicionar ao loop de embutimento
  - Testar extração durante o setup

- [ ] Verificar pipeline de build completo
  - Todos os 11 passos concluídos
  - Todos os binários presentes
  - Manifesto SHA256 inclui o mpm

**Tempo estimado:** 2 horas

### Tempo Total de Implementação

| Cenário                | Tempo      | Equipe      |
| ---------------------- | ---------- | ----------- |
| Desenvolvedor solo     | 14-15 dias | 1 pessoa    |
| Com programação em par | 7-8 dias   | 2 pessoas   |
| Equipe completa        | 5-6 dias   | 3-4 pessoas |

---

## 16. Critérios de Sucesso

### Qualidade de Código (Obrigatório)

- [ ] Todos os testes passando (13+ testes)
- [ ] Detector de race limpo: `go test -race ./...`
- [ ] Sem importações ou variáveis não utilizadas
- [ ] 90%+ de cobertura de código em `internal/models`
- [ ] Todas as funções exportadas possuem comentários de documentação
- [ ] Sem erros de linting: `golangci-lint run ./...`

### Funcionalidade (Obrigatório)

- [ ] `mpm list` exibe 7 modelos corretamente
- [ ] `mpm detect` mostra especificações de hardware (núcleos, RAM, GPU)
- [ ] `mpm recommend` faz recomendações sensatas
- [ ] `mpm install qwen3-4b` concluído com sucesso
- [ ] SHA256 verificado após o download
- [ ] `mpm set-active` alterna modelos
- [ ] Metadados do modelo persistidos em `metadata.json`
- [ ] `mpm search` encontra modelos por tag
- [ ] Saída JSON válida com a flag `--json`

### Integração (Obrigatório)

- [ ] `mpm.exe` tem 6MB ou menos
- [ ] `build.ps1` concluído em < 5 minutos
- [ ] `vectora-setup.exe` contém o mpm.exe
- [ ] Instalador extrai o mpm corretamente
- [ ] Manifesto SHA256 inclui entrada do mpm
- [ ] Sem alterações disruptivas no LPM ou daemon

### Performance (Recomendado)

- [ ] Detecção de hardware < 1 segundo
- [ ] Recomendações de modelo < 100ms
- [ ] Busca no catálogo < 500ms
- [ ] Atualizações de progresso de download a cada 1 segundo
- [ ] Sem vazamentos de memória em operações longas

### Documentação (Obrigatório)

- [ ] 800+ linhas de especificação técnica
- [ ] Exemplos de CLI para todos os 8 comandos
- [ ] Formato do catálogo documentado
- [ ] Pontos de integração com o daemon/instalador
- [ ] README menciona o MPM na seção de modelos

---

## 17. Cronograma de Desenvolvimento e Fases

### Cronograma da Fase 2 (Q2 2026, Semana 1-2)

**Semana 1: Implementação Principal**

- Segunda-Terça: Tipos + Catálogo (Fases 1-2)
- Quarta: Detecção de hardware (Fase 3)
- Quinta-Sexta: Download + Gerenciador (Fases 4-5)

**Semana 2: Conclusão**

- Segunda: Busca + CLI (Fases 6-7)
- Terça-Quarta: Testes (Fase 8)
- Quinta-Sexta: Integração de build (Passo 9)

**Semana 3: Polimento e Integração**

- Correções de bugs dos testes
- Otimização de performance
- Finalização da documentação
- Testes de integração de sistema completo

### Entregáveis até o Final da Semana de 11 de Abril

1. ✅ Binário `mpm.exe` (6MB)
2. ✅ Catálogo com 7 modelos Qwen3
3. ✅ Suíte de testes completa (13+ testes)
4. ✅ `build.ps1` atualizado (11 passos)
5. ✅ Embutimento no instalador configurado
6. ✅ Manifesto SHA256 atualizado
7. ✅ Documentação completa

### Fases Futuras

**Fase 3 (Q2 2026, Semana 3-4)**: Qwen3-Coder-Next, operações em lote
**Fase 4 (Q3 2026)**: Qwen3-VL (modelos de visão)
**Fase 5 (Q4 2026)**: Llama3.2, Mistral, catálogos personalizados

---

## 18. Pontos de Integração

### Integração com o Instalador de Setup

```
vectora-setup.exe
    ↓
[1] Detecta o hardware (compartilhado com o LPM)
    ↓
[2] Recomenda build do llama.cpp (LPM)
    ↓
[3] Recomenda modelo de IA (MPM) ← NOVO
    ↓
[4] Usuário confirma
    ↓
[5] Baixa ambos via LPM + MPM
    ↓
[6] Salva modelo ativo no metadata.json
    ↓
[7] Setup concluído
```

### Integração com o Daemon

```go
manager, _ := models.NewModelManager()
active, _ := manager.GetActive()

if active == nil {
    hw, _ := DetectHardware()
    model, _ := manager.RecommendModel(hw)
    manager.Install(ctx, model.ID)
    active, _ = manager.GetActive()
}

modelPath := manager.GetModelPath(active.ModelID)
// → ~/.Vectora/models/qwen3-4b/qwen3-4b.gguf
```

### Integração CLI

```bash
vectora chat --model qwen3-8b              # Usar modelo específico
vectora chat --detect                      # Recomendação automática
mpm list                                   # Listar todos os modelos
mpm recommend                              # Obter recomendação
mpm install --model $(mpm recommend)       # Instalação com um comando
```

---

## 19. Manutenção e Atualizações

### Quando Atualizar o Catálogo

1. **Novos modelos Qwen lançados**: Adicionar ao `catalog.json`, reconstruir
2. **Novas quantizações GGUF disponíveis**: Atualizar entradas de modelos
3. **URLs do Hugging Face alteradas**: Editar entradas do `catalog.json`
4. **Remoção de modelos obsoletos**: Remover do catálogo

### Gerenciamento de Versão

O MPM usa o campo `Version` em cada entrada de modelo para rastreamento:

- Quando baixado e verificado
- Qual quantização foi instalada
- Notas de compatibilidade com versões do llama.cpp

### Futuro: Atualização Automática

Planejado para a Fase 3:

- Comando `mpm update-catalog`
- Buscar catálogo mais recente do Hugging Face
- Notificar o usuário sobre novas versões de modelos
- Upgrades com um clique com migração automática

---

**Versão da Especificação:** 1.0
**Última Atualização:** 05/04/2026
**Status:** Pronto para Implementação
**Consolidado de:** MPM_SPECIFICATION.md + MPM_EXECUTIVE_SUMMARY.md + MPM_IMPLEMENTATION_CHECKLIST.md + MPM_PLANNING_COMPLETE.md
