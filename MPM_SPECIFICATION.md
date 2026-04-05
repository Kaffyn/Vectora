# Model Package Manager (MPM) - Specification

**Status:** Design Specification
**Version:** 1.0
**Date:** 2026-04-05
**Component:** `cmd/mpm/` + `internal/models/`

---

## Overview

**MPM (Model Package Manager)** is a standalone CLI tool for downloading, managing, and switching between AI models compatible with Vectora. It handles GGUF format models from Hugging Face, with initial support for Qwen3 family models.

Like **LPM** (Llama Package Manager), MPM is:
- A **separate executable** (`mpm.exe`)
- **Independent** - works standalone or called by setup/daemon
- **Extensible** - future support for Llama, Mistral, etc.
- **Scriptable** - outputs machine-readable JSON for automation

---

## Architecture

### 1.1 Component Structure

```
cmd/mpm/
├── main.go              # CLI entry point, subcommands
├── commands.go          # list, install, active, search
└── version.go           # version info

internal/models/
├── types.go             # Model, Catalog, Hardware definitions
├── catalog.go           # Model catalog loading (embedded)
├── detector.go          # Hardware specs + RAM detection
├── downloader.go        # Hugging Face GGUF downloader
├── integrity.go         # SHA256 verification
├── manager.go           # EngineManager → ModelManager pattern
├── search.go            # Semantic search in model catalog
├── catalog.json         # Embedded model definitions (80+ models)
├── manager_test.go      # Unit tests
└── integration_test.go  # Full flow tests
```

### 1.2 Data Layout

```
%USERPROFILE%/.Vectora/
├── models/
│   ├── catalog.json              # Cached catalog metadata
│   ├── metadata.json             # Active model info + versions
│   ├── qwen3-0.6b/
│   │   ├── qwen3-0.6b.gguf       # Model weights (~0.6GB)
│   │   ├── qwen3-0.6b.gguf.sha256 # Integrity check
│   │   └── model.json            # Model metadata
│   ├── qwen3-4b/
│   │   ├── qwen3-4b.gguf         # Model weights (~4GB)
│   │   ├── qwen3-4b.gguf.sha256  # Integrity check
│   │   └── model.json
│   └── mistral-7b/               # Future
│       └── ...
├── engines/                       # LPM manages these
│   └── llama-cpp-b3430/
└── backups/
```

---

## 2. Core Features

### 2.1 Model Catalog

**Embedded at compile time** via `go:embed`, never requires network fetch:

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
      "description": "Qwen3 7B instruction-following model"
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
      "description": "Qwen3 1.7B - lightweight instruction-following model"
    }
  ]
}
```

**Initial Model Support (April 2026):**

| Model | Sizes | Quantizations | Status |
|-------|-------|----------------|--------|
| **Qwen3** | 0.6B, 1.7B, 4B, 8B | Q4_K_M, Q5_K_M, Q6_K | ✅ Supported |
| **Qwen3-Embedding** | 0.6B, 4B | Q4_K_M, Q5_K_M, Q6_K | ✅ Supported |
| Qwen3-Coder-Next | 32B, 80B | Q4_K_M, Q6_K | ✅ Roadmap Q2 |
| Qwen3-VL | 2B, 8B | Q4_K_M, Q5_K_M | ✅ Roadmap Q3 |
| Llama3.2 | 3B, 8B, 70B | Q4_K_M, Q6_K | 🔄 Future |
| Mistral | 7B | Q4_K_M, Q6_K | 🔄 Future |

### 2.2 Hardware Detection Integration

Leverages `internal/engines/detector.go` already built for LPM:

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

**Returns:**
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

### 2.3 Model Recommendation Algorithm

```go
func RecommendModel(hw *Hardware) (*Model, error) {
    // 4-tier strategy:
    // 1. Exact match (RAM fits, GPU compatible)
    // 2. Best smaller model that fits RAM
    // 3. Largest model that fits available RAM
    // 4. Fallback to Qwen3-0.6B (always fits)

    availableRAM := hw.RAM - 2.0 // Reserve 2GB for system

    // Tier 1: Perfect fit
    for _, m := range catalog.Models {
        if m.Requirements.RecommendedRAM <= availableRAM {
            if hardwareMatches(hw, m) {
                return m, nil
            }
        }
    }

    // Tier 2: Fallback progressively smaller
    // ...
    return fallbackModel, nil
}
```

### 2.4 Download & Integrity

**Features:**
- Resume support via `.partial` files
- SHA256 verification before extraction
- Progress callbacks
- Exponential backoff (3 retries)
- Bandwidth throttling (optional)

```go
type DownloadProgress struct {
    Current int64        // Bytes downloaded
    Total   int64        // Total bytes
    Speed   float64      // MB/s
    ETA     time.Duration
}

func (m *ModelManager) Download(ctx context.Context, modelID string,
    onProgress func(*DownloadProgress) error) error
```

---

## 3. CLI Interface

### 3.1 Subcommands

#### `mpm list`
List all available models with filtering.

```bash
$ mpm list
$ mpm list --family qwen3
$ mpm list --filter "7b|1.5b"
$ mpm list --json
```

Output:
```
ID                    Family     Size  RAM    GPU    Status
qwen3-8b              qwen3      8B    8GB    ✅     Available
qwen3-4b              qwen3      4B    6GB    ✅     Available
qwen3-1.7b            qwen3      1.7B  4GB    ✅     Available
qwen3-0.6b            qwen3      0.6B  2GB    ✅     Available
```

#### `mpm detect`
Detect local hardware and capabilities.

```bash
$ mpm detect
```

Output:
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
Recommend best model for detected hardware.

```bash
$ mpm recommend
$ mpm recommend --json
$ mpm recommend --size 7b  # Force specific size
```

Output:
```
Recommended Model: qwen3-7b
Reason: Perfect fit for 16GB RAM + CUDA support
```

Or JSON:
```json
{
  "model_id": "qwen3-7b",
  "reason": "Perfect fit for 16GB RAM + CUDA support",
  "requires_ram_gb": 8,
  "size_bytes": 7000000000
}
```

#### `mpm install`
Download and install a model.

```bash
$ mpm install --model qwen3-8b
$ mpm install --model qwen3-4b --silent
$ mpm install --model $(mpm recommend)
```

Output:
```
Installing: qwen3-7b
Downloading: ████████████░░░░░░ 65% (4.5GB / 7GB) @ 25.3 MB/s
ETA: 2m 15s
✓ Download complete
✓ SHA256 verified
✓ Model installed at ~/.Vectora/models/qwen3-7b/
```

#### `mpm active`
Show currently active model.

```bash
$ mpm active
$ mpm active --json
```

Output:
```
Active Model: qwen3-7b
Path:         ~/.Vectora/models/qwen3-7b/qwen3-7b-q6_k.gguf
Size:         7.0 GB
RAM Required: 8 GB
Installed:    2026-04-05 14:30:45
```

#### `mpm set-active`
Switch to a different installed model.

```bash
$ mpm set-active --model qwen3-4b
```

#### `mpm search`
Search models by name, capability, or tag.

```bash
$ mpm search "coding"
$ mpm search --tag "lightweight"
$ mpm search --capability "instruct"
```

Output:
```
Found 4 models:
1. qwen3-8b (8B) - General purpose model
2. qwen3-4b (4B) - Lightweight, instruction-following
3. qwen3-1.7b (1.7B) - Very lightweight
4. qwen3-coder-next (80B) - Specialized for code generation [Roadmap]
```

#### `mpm update-catalog`
Refresh the embedded catalog (checks Hugging Face).

```bash
$ mpm update-catalog
✓ Catalog updated. Now 85 models available.
```

#### `mpm versions`
List available quantizations for a model.

```bash
$ mpm versions qwen3-7b
```

Output:
```
Quantizations for qwen3-7b:
- Q4_K_M  3.8 GB
- Q5_K_M  4.5 GB
- Q6_K    5.2 GB
- Q8_0    6.8 GB
```

### 3.2 Global Flags

```
--json              Output machine-readable JSON
--silent            Suppress progress output (for automation)
--log-level DEBUG   Set logging level
--timeout 3600      Timeout in seconds for downloads
--threads 4         Parallel download threads
```

---

## 4. Integration with Vectora Ecosystem

### 4.1 Setup Installer Integration

During `vectora-setup.exe`:

```
1. Detect hardware (shared with LPM)
   └─ Call: internal/engines.DetectHardware()

2. Recommend llama.cpp build
   └─ Call: lpm.recommend()

3. Recommend model
   └─ Call: mpm.recommend()

4. User selects provider
   ├─ Gemini → Skip model install
   └─ Qwen Local
       ├─ Download llama.cpp build via LPM
       └─ Download model via MPM

5. Store active model in ~/.Vectora/models/metadata.json
```

### 4.2 Daemon Integration

The daemon (`vectora.exe`) calls MPM programmatically:

```go
import "github.com/Kaffyn/Vectora/internal/models"

// In llm/qwen.go or similar
manager, err := models.NewModelManager()
if err != nil {
    return err
}

// Check if model is installed
info, err := manager.GetActive()
if err != nil {
    // Auto-download recommended model
    hw, _ := detectHardware()
    model, _ := manager.RecommendModel(hw)
    manager.Install(ctx, model.ID)
}

// Use model path for llama-cli
modelPath := manager.GetModelPath(modelID)
// → /Users/bruno/.Vectora/models/qwen3-7b/qwen3-7b-q6_k.gguf
```

### 4.3 CLI Integration

When user types `mpm` command:

```bash
vectora chat --model qwen3-8b
# Or
vectora chat --model-list
# Or
vectora chat --detect  # Auto-detect and recommend
```

---

## 5. Implementation Phases

### Phase 1: MVP (Week 1-2)
- ✅ Core CLI structure (`cmd/mpm/main.go`)
- ✅ Model catalog with Qwen3/Qwen3.5 (20+ models)
- ✅ Hardware detection (reuse from LPM)
- ✅ Download + SHA256 verification
- ✅ `list`, `detect`, `recommend`, `install`, `active` subcommands
- ✅ Basic tests

### Phase 2: Polish (Week 2-3)
- JSON output for all commands
- Search and filter functionality
- Better progress feedback
- Bandwidth throttling
- Integration tests with daemon
- Documentation

### Phase 3: Expansion (Week 4+)
- Llama3.2 support
- Mistral support
- Custom catalog sources (self-hosted)
- Model quantization conversion (GGML → GGUF)
- Caching of model metadata

---

## 6. File Structure

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

## 7. Build Integration

### 7.1 build.ps1 Changes

```powershell
# Step 6: Build MPM
Write-Host "[6/11] Compilando MPM - Model Package Manager (cmd/mpm)..."
go build -ldflags="-s -w" -o mpm.exe ./cmd/mpm
if (-not (Test-Path "mpm.exe")) { throw "FALHA: mpm.exe não foi gerado." }

# Step 7: LPM (unchanged)

# Step 8: Sync to installer
Copy-Item "mpm.exe" "cmd/vectora-installer/mpm.exe" -Force

# Step 9: Build installer with mpm.exe embedded
```

### 7.2 Installer Embedding

```go
// cmd/vectora-installer/embed_windows.go
//go:embed mpm.exe
var mpmExe []byte

// Extract during setup
for _, binName := range []string{"vectora.exe", "mpm.exe", "lpm.exe"} {
    if binData, ok := assets[binName]; ok {
        os.WriteFile(filepath.Join(installPath, binName), binData, 0755)
    }
}
```

---

## 8. Dependencies

**New Go imports:**
```go
import (
    "net/http"              // HTTP downloads
    "io"                    // I/O operations
    "crypto/sha256"         // Hash verification
    "os"                    // File operations
    "path/filepath"         // Path utilities
)
```

**Reused from `internal/engines`:**
```go
import (
    "github.com/Kaffyn/Vectora/internal/engines"
    // Reuse: DetectHardware, DownloadProgress pattern, SHA256 utils
)
```

---

## 9. Testing Strategy

### Unit Tests (`manager_test.go`)
- `TestLoadCatalog` - Verify catalog loading
- `TestDetectHardware` - Verify hardware detection
- `TestRecommendModel` - 4+ test cases for recommendation logic
- `TestVerifyFile` - SHA256 validation
- `TestPaths` - Directory resolution

### Integration Tests (`integration_test.go`)
- `TestFullInstallationFlow` - Download + verify + extract
- `TestModelSwitching` - Install multiple, switch between
- `TestDownloadResume` - Resume interrupted download
- `TestSearchFiltering` - Search and filter models

---

## 10. Future Roadmap

**Q2 2026:**
- Qwen3-Coder support
- Batch install multiple models
- Model auto-update checks

**Q3 2026:**
- Qwen3-VL (Vision) support
- Llama3.2 models
- Mistral models

**Q4 2026:**
- Self-hosted catalog sources
- Model quantization tools
- Community model registry

---

## 11. Error Handling

**Download Errors:**
```
✗ Download failed: Connection timeout
  Retrying in 5 seconds... (Attempt 2/3)

✗ SHA256 mismatch
  Expected: abc123...
  Got:      def456...
  File corrupted or incomplete. Deleting...
```

**Disk Space:**
```
✗ Insufficient disk space
  Required: 7.0 GB
  Available: 2.3 GB
  Please free up ~5 GB and try again
```

**Network:**
```
✗ Hugging Face unreachable
  Using cached catalog from: 2026-04-04
  Run 'mpm update-catalog' to refresh when online
```

---

## 12. Example Usage Flow

```bash
# First time user
$ mpm detect
$ mpm recommend
  → Recommends: qwen3-4b

$ mpm install --model qwen3-4b
  → Downloads 4 GB
  → Verifies SHA256
  → Extracts to ~/.Vectora/models/

# Switch models
$ mpm list
$ mpm set-active --model qwen3-7b

# Use with Vectora
$ vectora          # Daemon starts, uses active model
$ vectora-app      # Web UI
$ vectora-cli      # Terminal interface

# Automation
$ mpm list --json | jq '.models[] | select(.family=="qwen3")'
$ mpm recommend --json | jq '.model_id'
$ mpm install --model $(mpm recommend) --silent
```

---

**End of Specification**
