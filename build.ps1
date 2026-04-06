# Vectora Build Script - Unified cross-platform compilation
# RULES:
# 1. Binaries ONLY in bin/ with convention: {app}-{os}-{arch}{suffix}
# 2. Executable ONLY via build.ps1
# 3. CGO=1 only for Fyne (setup, desktop)
# 4. Clean bin/ before compiling

$ErrorActionPreference = "Stop"

# Use ANSI escape codes for colors
$RED = "`e[0;31m"
$GREEN = "`e[0;32m"
$YELLOW = "`e[1;33m"
$NC = "`e[0m" # No Color

$BIN_DIR = "bin"

# Cleanup bin/
if (Test-Path "$BIN_DIR") {
    Write-Host "${YELLOW}[CLEAN] Cleaning bin/...${NC}"
    Remove-Item -Recurse -Force "$BIN_DIR"
}
New-Item -ItemType Directory -Force "$BIN_DIR" | Out-Null

# System info
$OS = "windows"
$ARCH = $env:PROCESSOR_ARCHITECTURE
if ($ARCH -eq "AMD64") {
    $GOARCH_HOST = "amd64"
} elseif ($ARCH -eq "ARM64") {
    $GOARCH_HOST = "arm64"
} else {
    $GOARCH_HOST = $ARCH.ToLower()
}

Write-Host ""
Write-Host "================================================================"
Write-Host "        Vectora Build - Unified Universal Compilation         "
Write-Host "================================================================"
Write-Host ""
Write-Host "System: Windows ($ARCH -> $GOARCH_HOST)"
Write-Host ""

# Function to build a single binary
function Build-Binary {
    param (
        [string]$cmdName,      # vectora, lpm, mpm, etc
        [string]$path,         # cmd/daemon, cmd/lpm, etc
        [string]$target_os,    # windows, linux, darwin
        [string]$target_arch,  # amd64, arm64, etc
        [string]$suffix        # .exe for Windows, "" for Unix
    )

    $env:GOOS = $target_os
    $env:GOARCH = $target_arch

    # CGO_ENABLED=1 ONLY for Fyne apps (setup, desktop)
    if ($cmdName -eq "vectora-setup" -or $cmdName -eq "vectora-desktop") {
        $env:CGO_ENABLED = "1"
    } else {
        $env:CGO_ENABLED = "0"
    }

    # Naming convention: {app}-{os}-{arch}{suffix}
    $output = "$BIN_DIR/${cmdName}-${target_os}-${target_arch}${suffix}"

    Write-Host ("  {0,-40}" -f "Building ${cmdName}-${target_os}-${target_arch}...") -NoNewline

    # Invoke 'go build'
    & go build -ldflags="-s -w" -o "$output" "./$path" 2>$null

    if ($LASTEXITCODE -eq 0 -and (Test-Path "$output")) {
        $file = Get-Item "$output"
        $size = if ($file.Length -gt 1MB) { "$("{0:N2}" -f ($file.Length / 1MB)) MB" } else { "$("{0:N2}" -f ($file.Length / 1KB)) KB" }
        Write-Host "${GREEN}OK${NC} ($size)"
        return $true
    }

    Write-Host "${RED}FAIL${NC}"
    return $false
}

Write-Host "Compiling for Windows (Native Targets)..."
Write-Host ""

# PHASE 1: Compile dependencies (mpm, lpm, vectora daemon - required by setup for embedding)
Write-Host "${YELLOW}[PHASE 1] Compiling dependencies (vectora, mpm, lpm)...${NC}"
Build-Binary "vectora" "cmd/daemon" "windows" "amd64" ".exe"
Build-Binary "mpm" "cmd/mpm" "windows" "amd64" ".exe"
Build-Binary "lpm" "cmd/lpm" "windows" "amd64" ".exe"
Build-Binary "mpm" "cmd/mpm" "windows" "arm64" ".exe"
Build-Binary "lpm" "cmd/lpm" "windows" "arm64" ".exe"
Build-Binary "vectora" "cmd/daemon" "windows" "arm64" ".exe"

# PHASE 2: Copy compiled binaries to cmd/setup/ for embedding
Write-Host ""
Write-Host "${YELLOW}[PHASE 2] Copying binaries to cmd/setup/ for embedding...${NC}"
Copy-Item -Path "$BIN_DIR/vectora-windows-amd64.exe" -Destination "cmd/setup/vectora.exe" -Force
Copy-Item -Path "$BIN_DIR/mpm-windows-amd64.exe" -Destination "cmd/setup/mpm.exe" -Force
Copy-Item -Path "$BIN_DIR/lpm-windows-amd64.exe" -Destination "cmd/setup/lpm.exe" -Force
Write-Host "  ${GREEN}OK${NC} Binaries copied"

# PHASE 3: Compile setup and desktop (now have updated embedded binaries)
Write-Host ""
Write-Host "${YELLOW}[PHASE 3] Compiling setup and desktop...${NC}"
Build-Binary "vectora-setup" "cmd/setup" "windows" "amd64" ".exe"
Build-Binary "vectora-desktop" "cmd/desktop" "windows" "amd64" ".exe"

# PHASE 4: Compile TUI for all targets
Write-Host ""
Write-Host "${YELLOW}[PHASE 4] Compiling TUI...${NC}"
Build-Binary "vectora-tui" "cmd/tui" "windows" "amd64" ".exe"
Build-Binary "vectora-tui" "cmd/tui" "windows" "arm64" ".exe"

# Cleanup env variables
$env:GOOS = ""
$env:GOARCH = ""
$env:CGO_ENABLED = ""

Write-Host ""
Write-Host "================================================================"
Write-Host "             Build Completed Successfully!                  "
Write-Host "================================================================"
Write-Host ""

# Summary
$files = @(Get-ChildItem "$BIN_DIR" -File)
$count = $files.Count
$totalSize = ($files | Measure-Object -Property Length -Sum).Sum
$sizeInMB = [math]::Round($totalSize / 1MB, 2)

Write-Host "Summary:"
Write-Host "  * Compiled binaries: $count"
Write-Host "  * Total size: $sizeInMB MB"
Write-Host "  * Location: ./$BIN_DIR/"
Write-Host ""
