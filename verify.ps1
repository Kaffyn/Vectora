# Vectora Verification Script
# This is the single source of truth for CI and release verification.

$ErrorActionPreference = "Stop"

$startTime = Get-Date

function Write-Header($text) {
    Write-Host "`n================================================================" -ForegroundColor Magenta
    Write-Host "        $text" -ForegroundColor Magenta
    Write-Host "================================================================`n" -ForegroundColor Magenta
}

Write-Header "Vectora Verification Suite"

# --- PHASE 0: Static Analysis & Linting ---
Write-Host " [PHASE 0] Running Static Analysis & Linting..." -ForegroundColor Cyan
Write-Host "  Checking Go Core (vet & fmt)..." -NoNewline
& go vet ./...
& go fmt ./...
Write-Host " OK" -ForegroundColor Green

Write-Host "  Checking VS Code Extension (lint)..." -NoNewline
Push-Location "extensions/vscode"
try {
    & npm run lint 2>$null | Out-Null
    Write-Host " OK" -ForegroundColor Green
} finally {
    Pop-Location
}

# --- PHASE 1: Go Core Tests ---
Write-Host ""
Write-Host " [PHASE 1] Running Core Go Unit Tests..." -ForegroundColor Cyan
& go test ./... -race
if ($LASTEXITCODE -ne 0) {
    Write-Host " FAIL: Go tests did not pass." -ForegroundColor Red
    exit 1
}
Write-Host " OK" -ForegroundColor Green

# --- PHASE 2: Full Build & Packaging ---
Write-Host ""
Write-Host " [PHASE 2] Building Ecosystem..." -ForegroundColor Cyan
& ./build.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host " FAIL: Build failed." -ForegroundColor Red
    exit 1
}

# --- PHASE 3: Smoke Test ---
Write-Host ""
Write-Host " [PHASE 3] Running Installation Smoke Test..." -ForegroundColor Cyan
if (Test-Path "./scripts/smoke-test.ps1") {
    & ./scripts/smoke-test.ps1
    if ($LASTEXITCODE -ne 0) {
        Write-Host " FAIL: Smoke test failed." -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host " SKIP: scripts/smoke-test.ps1 not found." -ForegroundColor Yellow
}

# --- PHASE 4: VS Code Integration Tests ---
Write-Host ""
Write-Host " [PHASE 4] Running VS Code Integration Tests..." -ForegroundColor Cyan
Push-Location "extensions/vscode"
try {
    Write-Host "  Executing integration tests (Mocha/Vitest)..."
    & npm test
    if ($LASTEXITCODE -ne 0) {
        Write-Host " FAIL: VS Code integration tests failed." -ForegroundColor Red
        exit 1
    }
} finally {
    Pop-Location
}

$endTime = Get-Date
$duration = $endTime - $startTime

Write-Host ""
Write-Host "================================================================" -ForegroundColor Green
Write-Host "  ALL SYSTEMS GO: Vectora is stable and verified! ($($duration.TotalSeconds.ToString("F2"))s)" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
