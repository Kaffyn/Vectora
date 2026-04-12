# Vectora Build Script - Core Only (Quick Build)
# This is a quick build script that skips the problematic VS Code extension build
# and focuses on the core binary + minimal extension setup

$ErrorActionPreference = "Stop"

$RED = "`e[0;31m"
$GREEN = "`e[0;32m"
$YELLOW = "`e[1;33m"
$NC = "`e[0m"

$BIN_DIR = "bin"
$APP_NAME = "vectora"
$CMD_PATH = "./cmd/core"
$VERSION = "0.1.0"

Write-Host ""
Write-Host "================================================================"
Write-Host "   Vectora Build (Core + Installation) - QUICK MODE        "
Write-Host "================================================================"
Write-Host ""

# Stop running instances
Write-Host "${YELLOW}[PRE-CHECK] Stopping active Vectora processes...${NC}"
$procNames = @("vectora", "vectora-windows-amd64", "vectora-windows-arm64")
foreach ($name in $procNames) {
    Get-Process $name -ErrorAction SilentlyContinue | Stop-Process -Force
}
Start-Sleep -Milliseconds 500

# Cleanup bin/
if (Test-Path "$BIN_DIR") {
    Write-Host "${YELLOW}[CLEAN] Cleaning bin/...${NC}"
    Remove-Item -Recurse -Force "$BIN_DIR"
}
New-Item -ItemType Directory -Force "$BIN_DIR" | Out-Null

# System info
$ARCH = $env:PROCESSOR_ARCHITECTURE
if ($ARCH -eq "AMD64") {
    $GOARCH_HOST = "amd64"
} else {
    $GOARCH_HOST = $ARCH.ToLower()
}

Write-Host "System: Windows ($ARCH -> $GOARCH_HOST)"
Write-Host ""

# PHASE 1: Compile core binaries only
Write-Host "${YELLOW}[PHASE 1] Compiling Vectora Core...${NC}"

function Build-Binary {
    param (
        [string]$os,
        [string]$arch,
        [string]$ext
    )
    $env:GOOS = $os
    $env:GOARCH = $arch
    $env:CGO_ENABLED = "0"
    $outputName = "${APP_NAME}-${os}-${arch}${ext}"
    $outputPath = Join-Path $BIN_DIR $outputName
    Write-Host ("  Building {0,-30} ..." -f $outputName) -NoNewline
    & go build -ldflags="-s -w" -o $outputPath $CMD_PATH 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0 -and (Test-Path "$outputPath")) {
        $file = Get-Item "$outputPath"
        $size = if ($file.Length -gt 1MB) { "$("{0:N2}" -f ($file.Length / 1MB)) MB" } else { "$("{0:N2}" -f ($file.Length / 1KB)) KB" }
        Write-Host " ${GREEN}OK${NC} ($size)"
        return $true
    }
    Write-Host " ${RED}FAIL${NC}"
    return $false
}

Build-Binary -os "windows" -arch "amd64" -ext ".exe"
Build-Binary -os "linux" -arch "amd64" -ext ""

# Cleanup env
$env:GOOS = ""
$env:GOARCH = ""
$env:CGO_ENABLED = ""

# PHASE 2: Install Core
Write-Host ""
Write-Host "${YELLOW}[PHASE 2] Installing Vectora Core...${NC}"
$InstallDir = Join-Path $env:USERPROFILE "AppData\Local\Vectora"
New-Item -ItemType Directory -Force $InstallDir | Out-Null
Copy-Item "$BIN_DIR/${APP_NAME}-windows-amd64.exe" "$InstallDir\vectora.exe" -Force
Write-Host "  Installed to: ${GREEN}$InstallDir\vectora.exe${NC}"

# PHASE 3: VS Code Extension
Write-Host ""
Write-Host "${YELLOW}[PHASE 3] VS Code Extension Setup...${NC}"

# Check if code CLI is available
Write-Host "  Checking VS Code installation..." -NoNewline
$codeAvailable = Get-Command code -ErrorAction SilentlyContinue
if ($codeAvailable) {
    Write-Host " ${GREEN}OK${NC}"

    # Try to find a pre-built .vsix or create a minimal one
    Push-Location "extensions/vscode"

    $vsixPath = "../../bin/vectora-vscode.vsix"
    Write-Host "  VS Code Extension Status:" -NoNewline

    # Check if .vsix already exists in bin/
    if (Test-Path $vsixPath) {
        Write-Host " ${GREEN}Using existing .vsix${NC}"
    } else {
        Write-Host " ${YELLOW}Creating package...${NC}"

        # Try to package without compiling (if dist/ exists)
        $distExists = Test-Path "./dist"
        if ($distExists) {
            Write-Host "    Found dist/ directory, packaging..." -NoNewline
            & npx vsce package --out $vsixPath 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0 -and (Test-Path $vsixPath)) {
                Write-Host " ${GREEN}OK${NC}"
            } else {
                Write-Host " ${RED}FAIL${NC} - run 'npm run compile' manually"
            }
        } else {
            Write-Host "    ${YELLOW}dist/ not found - skipping package${NC}"
            Write-Host "    To build the extension: cd extensions/vscode && npm run compile && npm run package"
        }
    }

    # Install extension if .vsix exists
    if (Test-Path $vsixPath) {
        Write-Host "  Installing to VS Code..." -NoNewline
        & code --install-extension $vsixPath --force 2>&1 | Out-Null
        Write-Host " ${GREEN}OK${NC}"
    }

    Pop-Location
} else {
    Write-Host " ${YELLOW}NOT FOUND${NC}"
    Write-Host "  VS Code 'code' CLI not available in PATH"
    Write-Host "  You can install it manually:"
    Write-Host "    1. cd extensions/vscode"
    Write-Host "    2. npm install"
    Write-Host "    3. npm run compile"
    Write-Host "    4. code --install-extension bin/vectora-vscode.vsix --force"
}

Write-Host ""
Write-Host "================================================================"
Write-Host "           Build Complete!                                   "
Write-Host "================================================================"
Write-Host ""
Write-Host "Core binary installed to: ${GREEN}$InstallDir${NC}"
Write-Host "Run: ${GREEN}vectora --help${NC}"
Write-Host ""
