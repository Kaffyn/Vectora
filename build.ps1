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

# host (x86_64) - all binaries
Build-Binary "vectora" "cmd/daemon" "windows" "amd64" ".exe"
Build-Binary "vectora-tui" "cmd/tui" "windows" "amd64" ".exe"
Build-Binary "vectora-setup" "cmd/setup" "windows" "amd64" ".exe"
Build-Binary "vectora-desktop" "cmd/desktop" "windows" "amd64" ".exe"
Build-Binary "mpm" "cmd/mpm" "windows" "amd64" ".exe"
Build-Binary "lpm" "cmd/lpm" "windows" "amd64" ".exe"

# arm64 (target) - only CLI tools
Build-Binary "vectora" "cmd/daemon" "windows" "arm64" ".exe"
Build-Binary "vectora-tui" "cmd/tui" "windows" "arm64" ".exe"
Build-Binary "mpm" "cmd/mpm" "windows" "arm64" ".exe"
Build-Binary "lpm" "cmd/lpm" "windows" "arm64" ".exe"

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
if (Test-Path "$BIN_DIR") {
    $files = Get-ChildItem "$BIN_DIR" -File
    $count = $files.Count
    $totalSize = ($files | Measure-Object -Property Length -Sum).Sum
    $totalSizeStr = if ($totalSize -gt 1MB) { "$("{0:N2}" -f ($totalSize / 1MB)) MB" } else { "$("{0:N2}" -f ($totalSize / 1KB)) KB" }
    
    Write-Host "Summary:"
    Write-Host "  * Compiled binaries: $count"
    Write-Host "  * Total size: $totalSizeStr"
    Write-Host "  * Location: ./$BIN_DIR/"
    Write-Host ""
    Write-Host "Generated binaries:"
    foreach ($f in $files) {
        Write-Host "  - $($f.Name)"
    }
    Write-Host ""
}
