# Vectora - Unified Test Suite Explorer
# Unifica testes do Go Core e da Extensão VS Code

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
    Write-Header "Vectora Master Test Suite"

    # --- PHASE 1: Go Core ---
    Write-Section "Go Core (Unit & Integration)"
    & go test ./... -race -v
    if ($LASTEXITCODE -ne 0) {
        Write-Host "e[0;31m[FAIL]e[0m Go Core tests failed."
        $allPass = $false
    } else {
        Write-Host "e[0;32m[PASS]e[0m Go Core tests completed."
    }

    # --- PHASE 2: VS Code Extension ---
    Write-Section "VS Code Extension (Unit, UI & E2E)"
    Push-Location extensions/vscode
    try {
        & npm test
        if ($LASTEXITCODE -ne 0) {
            Write-Host "e[0;31m[FAIL]e[0m VS Code Extension tests failed."
            $allPass = $false
        } else {
            Write-Host "e[0;32m[PASS]e[0m VS Code Extension tests completed."
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
        Write-Host "`n  🏆 ALL TESTS PASSED! 🚀" -ForegroundColor Green
        exit 0
    } else {
        Write-Host "`n  ❌ SOME TESTS FAILED. Check logs above." -ForegroundColor Red
        exit 1
    }

} catch {
    Write-Host "`n  🛑 Critical Failure: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
