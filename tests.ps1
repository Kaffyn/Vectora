# Vectora - Unified Master Test Suite
# Unifica testes do Go Core, Extensão VS Code, Linting e Análise Estática

$ErrorActionPreference = "Stop"

function Write-Header($text) {
    Write-Host "`n================================================================" -ForegroundColor Cyan
    Write-Host "        $text" -ForegroundColor Cyan
    Write-Host "================================================================`n" -ForegroundColor Cyan
}

function Write-Section($text) {
    Write-Host "`n[TEST PHASE] $text..." -ForegroundColor Yellow
}

$startTime = Get-Date
$allPass = $true

try {
    Write-Header "Vectora Master Test Suite (Comprehensive Version)"

    # --- PHASE 1: Go Quality & Static Analysis ---
    Write-Section "Go Static Analysis (Vet & Fmt)"
    & go vet ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ [FAIL] Go Vet found issues." -ForegroundColor Red
        $allPass = $false
    }
    
    & go fmt ./...
    Write-Host "✅ [PASS] Go Analysis completed." -ForegroundColor Green

    # --- PHASE 2: Go Core Tests with Coverage ---
    Write-Section "Go Core (Unit, Integration & Coverage)"
    & go test ./... -race -cover -v
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ [FAIL] Go Core tests failed." -ForegroundColor Red
        $allPass = $false
    } else {
        Write-Host "✅ [PASS] Go Core tests completed." -ForegroundColor Green
    }

    # --- PHASE 3: VS Code Extension Linting ---
    Write-Section "VS Code Extension Linting"
    Push-Location extensions/vscode
    try {
        & npm run lint
        if ($LASTEXITCODE -ne 0) {
            Write-Host "❌ [FAIL] Extension Linting found issues." -ForegroundColor Red
            $allPass = $false
        } else {
            Write-Host "✅ [PASS] Extension Linting completed." -ForegroundColor Green
        }
    } finally {
        Pop-Location
    }

    # --- PHASE 4: VS Code Extension Tests ---
    Write-Section "VS Code Extension (Unit, UI & E2E)"
    Push-Location extensions/vscode
    try {
        & npm test
        if ($LASTEXITCODE -ne 0) {
            Write-Host "❌ [FAIL] Extension tests failed." -ForegroundColor Red
            $allPass = $false
        } else {
            Write-Host "✅ [PASS] Extension tests completed." -ForegroundColor Green
        }
    } finally {
        Pop-Location
    }

    # --- SUMMARY ---
    $endTime = Get-Date
    $duration = $endTime - $startTime
    
    Write-Header "Test Summary"
    Write-Host "Duration: $($duration.TotalSeconds.ToString("F2"))s"
    
    if ($allPass) {
        Write-Host "`n  🏆 ALL QUALITY CHECKS PASSED! 🚀" -ForegroundColor Green
        exit 0
    } else {
        Write-Host "`n  ❌ QUALITY GATE FAILED. Check logs above." -ForegroundColor Red
        exit 1
    }

} catch {
    Write-Host "`n  🛑 Critical Failure: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
