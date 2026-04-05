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

## 13. Business Case & Motivation

### Why MPM?

Currently, Vectora users have two challenges:

1. **Manual Model Management**: Without MPM, users must:
   - Manually visit Hugging Face and find GGUF models
   - Download large files (~1-8GB) manually
   - Verify SHA256 hashes by hand
   - Extract files to correct directories
   - Manually update configuration
   - Manually restart Vectora daemon

2. **No Hardware Awareness**: Users don't know which model fits their hardware:
   - 16GB RAM + CUDA → Should use Qwen3-8B
   - 8GB RAM + CPU only → Should use Qwen3-4B
   - 4GB RAM → Should use Qwen3-1.7B

### With MPM (Future)

Users simply run:
```bash
mpm recommend              # Get recommendation based on hardware
mpm install --model qwen3-8b   # Download, verify, extract
```

**Impact:**
- ⏱️ Time: 30 minutes (manual) → 2 minutes (MPM)
- ✅ Automation: Repeatable, scriptable, reliable
- 🔒 Safety: Automatic SHA256 verification
- 🎯 Intelligence: Hardware-aware recommendations

### Comparison

| Aspect | Without MPM | With MPM |
|--------|------------|----------|
| **Time** | 30 min + downloads | 2 min + downloads |
| **Verification** | Manual or skipped | Automatic |
| **Recommendations** | User guesses | Hardware-aware |
| **Model Switching** | Manual config edit | `mpm set-active` |
| **Repeatability** | Error-prone | Automated |

---

## 14. Model Support Details

### Qwen3 Family (Phase 1 - MVP)

**Qwen3 Chat Models**
| Model | Size | RAM | GPU | Use Case |
|-------|------|-----|-----|----------|
| Qwen3-0.6B | 600MB | 2GB | Optional | Ultra-lightweight, IoT, embedded |
| Qwen3-1.7B | 1.7GB | 4GB | Optional | Mobile, resource-constrained |
| **Qwen3-4B** | **4GB** | **8GB** | **Recommended** | **General purpose, balanced** |
| Qwen3-8B | 8GB | 16GB | Recommended | High-performance, reasoning |

**Qwen3 Embedding Models**
| Model | Size | Use Case |
|-------|------|----------|
| Qwen3-Embedding-0.6B | 600MB | Lightweight embeddings |
| Qwen3-Embedding-4B | 4GB | High-quality embeddings |

**Qwen3-Coder-Next (Phase 2 - Q2 2026)**
| Model | Size | Use Case |
|-------|------|----------|
| Qwen3-Coder-Next | 32B-80B | Code generation & reasoning |

### Hardware Detection & Recommendation

MPM automatically detects:
- CPU cores and features (AVX, AVX2, NEON, etc.)
- Total RAM available
- GPU type and VRAM (CUDA, Metal, Vulkan)
- Free disk space

**4-Tier Recommendation Algorithm:**
1. **Perfect fit**: Model with recommended RAM <= available RAM
2. **Best smaller**: Largest model that still fits
3. **Largest possible**: Maximum model that doesn't crash system
4. **Fallback**: Always recommend Qwen3-0.6B (2GB footprint)

**Examples:**
- 32GB RAM + CUDA 12.0 → Qwen3-8B (recommended)
- 16GB RAM + CPU only → Qwen3-4B
- 8GB RAM → Qwen3-1.7B
- 4GB RAM → Qwen3-0.6B

---

## 15. Implementation Checklist

### Phase 1: Package Structure Setup
- [ ] Create `cmd/mpm/` directory
- [ ] Create `internal/models/` directory
- [ ] Create empty stub files for all modules
- [ ] Set up Go module imports
- [ ] Create Makefile targets

**Estimated time:** 1 hour

### Phase 2: Core Types & Data Structures
- [ ] `types.go` (60 lines)
  - `Model` struct with all metadata
  - `Catalog` struct for model list
  - `Hardware` struct for system info
  - `ModelManager` interface
  - Error codes enum

- [ ] `catalog.go` (100 lines)
  - `LoadCatalog()` from embedded JSON
  - `GetCatalog()` singleton
  - `FindModel(id)` lookup
  - `SearchModels(query)` search
  - `go:embed` directive

**Estimated time:** 2 hours

### Phase 3: Hardware Detection
- [ ] `detector.go` (50 lines)
  - Import `internal/engines.DetectHardware()`
  - Wrapper for model detection
  - Test with different hardware profiles

**Estimated time:** 1 hour

### Phase 4: Download & Integrity
- [ ] `downloader.go` (160 lines)
  - `NewDownloader()` with timeout
  - `Download()` with resume support
  - `.partial` file handling
  - Exponential backoff retry (3x)
  - Progress callbacks
  - Bandwidth throttling (optional)

- [ ] `integrity.go` (60 lines)
  - `VerifyFile()` SHA256 check
  - `ComputeFileSHA256()` hash computation
  - Reuse existing crypto code

**Estimated time:** 3 hours

### Phase 5: Manager & Orchestration
- [ ] `manager.go` (180 lines) - **Most critical**
  - `NewModelManager()` singleton
  - `Install(ctx, modelID, onProgress)` orchestrator
  - `GetActive()` read metadata.json
  - `SetActive(modelID)` update metadata
  - `GetModelPath(modelID)` resolve full path
  - `RecommendModel(hw)` 4-tier algorithm
  - `ListInstalled()` scan directory

**Estimated time:** 4 hours

### Phase 6: Search & Filtering
- [ ] `search.go` (80 lines)
  - `SearchByTag(tag)` filter by tag
  - `SearchByCapability(cap)` capability match
  - `SearchByName(query)` substring match
  - Combine filters with AND/OR logic

**Estimated time:** 1.5 hours

### Phase 7: CLI Implementation
- [ ] `main.go` (100 lines)
  - Parse subcommands
  - Global flags (--json, --silent, --log-level)
  - Route to command handlers
  - Version info
  - Error handling

- [ ] `commands.go` (200 lines)
  - `cmdList()` - list with filters
  - `cmdDetect()` - show hardware
  - `cmdRecommend()` - recommend model
  - `cmdInstall()` - download & install
  - `cmdActive()` - show active
  - `cmdSetActive()` - switch models
  - `cmdSearch()` - search catalog
  - `cmdVersions()` - quantization list
  - Pretty-print + JSON output

**Estimated time:** 3 hours

### Phase 8: Testing
- [ ] `manager_test.go` (150 lines)
  - `TestLoadCatalog` - embedding verification
  - `TestDetectHardware` - hardware detection
  - `TestRecommendModel` - 4 scenarios
  - `TestVerifyFile` - SHA256 validation
  - `TestSearchModels` - search functionality

- [ ] `integration_test.go` (100 lines)
  - `TestFullInstallationFlow` - end-to-end
  - `TestModelSwitching` - set-active flow
  - `TestDownloadResume` - retry logic
  - Table-driven tests

**Estimated time:** 2.5 hours

### Phase 9: Build & Integration
- [ ] Update `build.ps1`
  - Add step [7/11]: `go build -o mpm.exe ./cmd/mpm`
  - Copy `mpm.exe` to installer directory before embedding
  - Generate SHA256 for mpm.exe
  - Update progress output

- [ ] Update `embed_windows.go`
  - Add `//go:embed mpm.exe`
  - Add to embedding loop
  - Test extraction during setup

- [ ] Verify full build pipeline
  - All 11 steps complete
  - All binaries present
  - SHA256 manifest includes mpm

**Estimated time:** 2 hours

### Total Implementation Time

| Scenario | Time | Team |
|----------|------|------|
| Solo developer | 14-15 days | 1 person |
| With pair programming | 7-8 days | 2 people |
| Full team | 5-6 days | 3-4 people |

---

## 16. Success Criteria

### Code Quality (Mandatory)
- [ ] All tests passing (13+ tests)
- [ ] Race detector clean: `go test -race ./...`
- [ ] No unused imports or variables
- [ ] 90%+ code coverage on `internal/models`
- [ ] All exported functions have doc comments
- [ ] No linting errors: `golangci-lint run ./...`

### Functionality (Mandatory)
- [ ] `mpm list` displays 7 models correctly
- [ ] `mpm detect` shows hardware specs (cores, RAM, GPU)
- [ ] `mpm recommend` makes sensible recommendations
- [ ] `mpm install qwen3-4b` completes successfully
- [ ] SHA256 verified after download
- [ ] `mpm set-active` switches models
- [ ] Model metadata persisted in `metadata.json`
- [ ] `mpm search` finds models by tag
- [ ] JSON output valid with `--json` flag

### Integration (Mandatory)
- [ ] `mpm.exe` is 6MB or smaller
- [ ] `build.ps1` completes in <5 minutes
- [ ] `vectora-setup.exe` contains mpm.exe
- [ ] Installer extracts mpm correctly
- [ ] SHA256 manifest includes mpm entry
- [ ] No breaking changes to LPM or daemon

### Performance (Recommended)
- [ ] Hardware detection < 1 second
- [ ] Model recommendations < 100ms
- [ ] Catalog search < 500ms
- [ ] Download progress updates every 1 second
- [ ] No memory leaks in long operations

### Documentation (Mandatory)
- [ ] 800+ lines technical specification
- [ ] CLI examples for all 8 commands
- [ ] Catalog format documented
- [ ] Integration points with daemon/installer
- [ ] README mentions MPM in model section

---

## 17. Development Timeline & Phases

### Phase 2 Schedule (Q2 2026, Week 1-2)

**Week 1: Core Implementation**
- Monday-Tuesday: Types + Catalog (Phases 1-2)
- Wednesday: Hardware detection (Phase 3)
- Thursday-Friday: Download + Manager (Phases 4-5)

**Week 2: Completion**
- Monday: Search + CLI (Phases 6-7)
- Tuesday-Wednesday: Testing (Phase 8)
- Thursday-Friday: Build integration (Phase 9)

**Week 3: Polish & Integration**
- Bug fixes from testing
- Performance optimization
- Documentation finalization
- Full system integration testing

### Deliverables by EOW April 11

1. ✅ `mpm.exe` binary (6MB)
2. ✅ Catalog with 7 Qwen3 models
3. ✅ Full test suite (13+ tests)
4. ✅ Updated `build.ps1` (11 steps)
5. ✅ Installer embedding configured
6. ✅ SHA256 manifest updated
7. ✅ Complete documentation

### Future Phases

**Phase 3 (Q2 2026, Week 3-4)**: Qwen3-Coder-Next, batch operations
**Phase 4 (Q3 2026)**: Qwen3-VL (vision models)
**Phase 5 (Q4 2026)**: Llama3.2, Mistral, custom catalogs

---

## 18. Integration Points

### Setup Installer Integration
```
vectora-setup.exe
    ↓
[1] Detect hardware (shared with LPM)
    ↓
[2] Recommend llama.cpp build (LPM)
    ↓
[3] Recommend AI model (MPM) ← NEW
    ↓
[4] User confirms
    ↓
[5] Download both via LPM + MPM
    ↓
[6] Save active model to metadata.json
    ↓
[7] Setup complete
```

### Daemon Integration
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

### CLI Integration
```bash
vectora chat --model qwen3-8b              # Use specific model
vectora chat --detect                      # Auto-recommend
mpm list                                   # List all models
mpm recommend                              # Get recommendation
mpm install --model $(mpm recommend)       # One-command install
```

---

## 19. Maintenance & Updates

### When to Update Catalog

1. **New Qwen models released**: Add to `catalog.json`, rebuild
2. **New GGUF quantizations available**: Update model entries
3. **Hugging Face URLs change**: Edit `catalog.json` entries
4. **Removal of deprecated models**: Remove from catalog

### Version Management

MPM uses `Version` field in each model entry for tracking:
- When downloaded and verified
- Which quantization was installed
- Compatibility notes with llama.cpp versions

### Future: Auto-Update

Planned for Phase 3:
- `mpm update-catalog` command
- Fetch latest catalog from Hugging Face
- Notify user of new model versions
- One-click upgrades with automatic migration

---

**Specification Version:** 1.0
**Last Updated:** 2026-04-05
**Status:** Ready for Implementation
**Consolidated From:** MPM_SPECIFICATION.md + MPM_EXECUTIVE_SUMMARY.md + MPM_IMPLEMENTATION_CHECKLIST.md + MPM_PLANNING_COMPLETE.md
