# Vectora Build Script - Unified cross-platform compilation
# RULES:
# 1. Windows: .exe (PE binary with embedded icon)
# 2. Linux: no extension (ELF binary)
# 3. macOS: .app bundle (directory structure with Info.plist)
# 4. Clean bin/ before compiling

$ErrorActionPreference = "Stop"

# Use ANSI escape codes for colors
$RED = "`e[0;31m"
$GREEN = "`e[0;32m"
$YELLOW = "`e[1;33m"
$NC = "`e[0m" # No Color

$BIN_DIR = "bin"
$APP_NAME = "vectora"
$CMD_PATH = "./cmd/daemon"
$VERSION = "0.1.0"

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

# Function to generate deterministic build hash
function Generate-BuildHash {
    param([string]$BinDir)

    $hashFiles = @(
        "$BinDir/${APP_NAME}-windows-amd64.exe",
        "$BinDir/${APP_NAME}",
        "$BinDir/${APP_NAME}-darwin-amd64.app/Contents/MacOS/${APP_NAME}"
    )

    $combinedHash = ""
    foreach ($file in $hashFiles) {
        if (Test-Path $file) {
            $fileHash = (Get-FileHash $file -Algorithm SHA256).Hash
            $combinedHash += $fileHash
        }
    }

    if ($combinedHash -ne "") {
        $utf8 = [System.Text.Encoding]::UTF8
        $hashBytes = $utf8.GetBytes($combinedHash)
        $sha256 = [System.Security.Cryptography.SHA256]::Create()
        $hashResult = $sha256.ComputeHash($hashBytes)
        $finalHash = -join ($hashResult | ForEach-Object { "{0:x2}" -f $_ })
        return $finalHash.Substring(0, [Math]::Min(16, $finalHash.Length))
    }

    return "development"
}

# Function to build a single binary
function Build-Binary {
    param (
        [string]$os,
        [string]$arch,
        [string]$ext,
        [string]$ldflags
    )

    $env:GOOS = $os
    $env:GOARCH = $arch
    $env:CGO_ENABLED = "0"

    $outputName = "${APP_NAME}-${os}-${arch}${ext}"
    $outputPath = Join-Path $BIN_DIR $outputName

    Write-Host ("Building {0,-35} ..." -f $outputName) -NoNewline

    $finalFlags = "-s -w"
    if ($os -eq "windows") {
        $finalFlags += " -H=windowsgui"
    }
    if ($ldflags -ne "") {
        $finalFlags += " $ldflags"
    }

    & go build -ldflags="$finalFlags" -o $outputPath $CMD_PATH 2>$null

    if ($LASTEXITCODE -eq 0 -and (Test-Path "$outputPath")) {
        $file = Get-Item "$outputPath"
        $size = if ($file.Length -gt 1MB) { "$("{0:N2}" -f ($file.Length / 1MB)) MB" } else { "$("{0:N2}" -f ($file.Length / 1KB)) KB" }
        Write-Host "${GREEN}OK${NC} ($size)"
        return $true
    }

    Write-Host "${RED}FAIL${NC}"
    return $false
}

# Function to create macOS .app bundle
function Build-MacOSAppBundle {
    param (
        [string]$arch
    )

    $bundleName = "${APP_NAME}-darwin-${arch}.app"
    $bundlePath = Join-Path $BIN_DIR $bundleName
    $contentsDir = Join-Path $bundlePath "Contents"
    $macosDir = Join-Path $contentsDir "MacOS"
    $resourcesDir = Join-Path $contentsDir "Resources"
    $binaryName = $APP_NAME

    Write-Host ("Building {0,-35} ..." -f $bundleName) -NoNewline

    $env:GOOS = "darwin"
    $env:GOARCH = $arch
    $env:CGO_ENABLED = "0"

    # Create bundle structure
    New-Item -ItemType Directory -Force $macosDir | Out-Null
    New-Item -ItemType Directory -Force $resourcesDir | Out-Null

    # Compile binary directly into bundle
    $binaryPath = Join-Path $macosDir $binaryName
    & go build -ldflags="-s -w" -o $binaryPath $CMD_PATH 2>$null

    if ($LASTEXITCODE -ne 0) {
        Write-Host "${RED}FAIL${NC}"
        return $false
    }

    # Generate Info.plist
    $infoPlist = @"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>${binaryName}</string>
    <key>CFBundleIdentifier</key>
    <string>com.vectora.app</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>Vectora</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
"@

    $infoPlistPath = Join-Path $contentsDir "Info.plist"
    $infoPlist | Out-File -FilePath $infoPlistPath -Encoding utf8

    $size = (Get-ChildItem $macosDir -Recurse | Measure-Object -Property Length -Sum).Sum
    $sizeStr = if ($size -gt 1MB) { "$("{0:N2}" -f ($size / 1MB)) MB" } else { "$("{0:N2}" -f ($size / 1KB)) KB" }
    Write-Host "${GREEN}OK${NC} ($sizeStr)"
    return $true
}

Write-Host "Compiling for all platforms (amd64)..."
Write-Host ""

# PHASE 0: Generate Windows resource file (icon + version info)
Write-Host "${YELLOW}[PHASE 0] Generating Windows resource file...${NC}"
if (Test-Path "cmd/vectora/versioninfo.json") {
    Push-Location "cmd/vectora"
    & goversioninfo -64 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ${GREEN}OK${NC} - resource file generated (amd64)"
    } else {
        Write-Host "  ${YELLOW}WARN${NC} - goversioninfo unavailable, skipping icon"
    }
    Pop-Location
} else {
    Write-Host "  ${YELLOW}WARN${NC} - versioninfo.json not found, skipping icon"
}

# PHASE 1: Compile binaries
Write-Host "${YELLOW}[PHASE 1] Compiling binaries...${NC}"
Build-Binary -os "windows" -arch "amd64" -ext ".exe" -ldflags ""

# Linux: no extension (Unix convention)
Build-Binary -os "linux" -arch "amd64" -ext "" -ldflags ""

# macOS: .app bundle (Apple convention)
Build-MacOSAppBundle -arch "amd64"

# Generate build hash
Write-Host ""
Write-Host "${YELLOW}[HASH] Generating build hash...${NC}"
$buildHash = Generate-BuildHash $BIN_DIR
Write-Host "  Build Hash: ${GREEN}$buildHash${NC}"

# Cleanup
$env:GOOS = ""
$env:GOARCH = ""
$env:CGO_ENABLED = ""

# PHASE 2: Package for distribution
Write-Host ""
Write-Host "${YELLOW}[PHASE 2] Packaging for distribution...${NC}"

# --- Windows: Install to user-local directory + PATH ---
$InstallDir = Join-Path $env:USERPROFILE "AppData\Local\Vectora"
New-Item -ItemType Directory -Force $InstallDir | Out-Null
Copy-Item "$BIN_DIR/${APP_NAME}-windows-amd64.exe" "$InstallDir\vectora.exe" -Force

# Add to PATH if not already there
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$InstallDir*") {
    Write-Host "  Adding $InstallDir to User PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
    $env:Path += ";$InstallDir"
    Write-Host "  ${GREEN}OK${NC} - PATH updated (restart terminal to use)"
} else {
    Write-Host "  ${GREEN}OK${NC} - Install dir already in PATH"
}

# --- Linux: Create .tar.gz for distribution ---
$linuxBin = "$BIN_DIR/${APP_NAME}-linux-amd64"
if (Test-Path $linuxBin) {
    $tmpDir = Join-Path $BIN_DIR "linux-tmp"
    New-Item -ItemType Directory -Force "$tmpDir\usr\local\bin" | Out-Null
    Copy-Item $linuxBin "$tmpDir\usr\local\bin\${APP_NAME}" -Force
    
    $outPath = Resolve-Path $BIN_DIR
    $tarOut = Join-Path $outPath "${APP_NAME}-linux-amd64.tar.gz"
    
    if (Get-Command tar -ErrorAction SilentlyContinue) {
        Push-Location $tmpDir
        tar -czf $tarOut usr 2>$null
        Pop-Location
        Remove-Item $tmpDir -Recurse -Force
        if (Test-Path $tarOut) {
            Write-Host "  ${GREEN}OK${NC} - Linux .tar.gz created"
        } else {
            Write-Host "  ${YELLOW}WARN${NC} - tar failed, skipping Linux package"
        }
    } else {
        Remove-Item $tmpDir -Recurse -Force
        Write-Host "  ${YELLOW}WARN${NC} - tar not available, skipping Linux package"
    }
}

# --- macOS: Create .tar.gz of .app bundle ---
$macApp = "$BIN_DIR/${APP_NAME}-darwin-amd64.app"
if (Test-Path $macApp) {
    $outPath = Resolve-Path $BIN_DIR
    $macTarOut = Join-Path $outPath "${APP_NAME}-darwin-amd64.app.tar.gz"
    
    if (Get-Command tar -ErrorAction SilentlyContinue) {
        Push-Location $BIN_DIR
        tar -czf $macTarOut "${APP_NAME}-darwin-amd64.app" 2>$null
        Pop-Location
        if (Test-Path $macTarOut) {
            Write-Host "  ${GREEN}OK${NC} - macOS .app.tar.gz created"
        }
    }
}

Write-Host ""
Write-Host "================================================================"
Write-Host "             Build Completed Successfully!                  "
Write-Host "================================================================"
Write-Host ""

# Summary
$files = @(Get-ChildItem "$BIN_DIR" -File -Recurse)
$count = $files.Count
$totalSize = ($files | Measure-Object -Property Length -Sum).Sum
$sizeInMB = [math]::Round($totalSize / 1MB, 2)

Write-Host "Summary:"
Write-Host "  * Compiled binaries: $count"
Write-Host "  * Total size: $sizeInMB MB"
Write-Host "  * Location: ./$BIN_DIR/"
Write-Host "  * Installed to: $InstallDir\vectora.exe"
Write-Host ""
