# Vectora Master Verification Script
$ErrorActionPreference = "Stop"

Write-Host "================================================================" -ForegroundColor Magenta
Write-Host "        Vectora - Unified Verification Suite" -ForegroundColor Magenta
Write-Host "================================================================" -ForegroundColor Magenta

# 1. Run Go Core Tests
Write-Host ""
Write-Host " [PHASE 1] Running Core Go Tests..." -ForegroundColor Cyan
& go test ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host " FAIL: Go tests did not pass." -ForegroundColor Red
    exit 1
}
Write-Host " OK" -ForegroundColor Green

# 2. Run Full Build
Write-Host ""
Write-Host " [PHASE 2] Building Ecosystem..." -ForegroundColor Cyan
& ./build.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host " FAIL: Build failed." -ForegroundColor Red
    exit 1
}

# 3. Run Smoke Test
Write-Host ""
Write-Host " [PHASE 3] Running Installation Smoke Test..." -ForegroundColor Cyan
& ./scripts/smoke-test.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host " FAIL: Smoke test failed." -ForegroundColor Red
    exit 1
}

# 4. Run VS Code Integration Tests
Write-Host ""
Write-Host " [PHASE 4] Running VS Code Integration Tests..." -ForegroundColor Cyan
Push-Location "extensions/vscode"
Write-Host "  Installing test dependencies..."
& npm install --no-audit --no-fund
Write-Host "  Compiling tests..."
& npm run compile
Write-Host "  Executing integration tests (Mocha)..."
& npm test
if ($LASTEXITCODE -ne 0) {
    Write-Host " FAIL: VS Code integration tests failed." -ForegroundColor Red
    Pop-Location
    exit 1
}
Pop-Location

Write-Host ""
Write-Host "================================================================" -ForegroundColor Green
Write-Host "        ALL SYSTEMS GO: Vectora is stable and verified!" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
