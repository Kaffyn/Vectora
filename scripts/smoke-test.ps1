# Vectora Smoke Test - Installation Validation
$BIN_DIR = "bin"
$EXT_VSIX = "$BIN_DIR/vectora-vscode.vsix"

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "        Vectora Smoke Test - Installation Validation" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Check if VSIX exists
if (-not (Test-Path $EXT_VSIX)) {
    Write-Host " [FAIL] extension .vsix not found in $BIN_DIR. Run ./build.ps1 first." -ForegroundColor Red
    exit 1
}

Write-Host " [STEP 1] Validating VSIX package..." -NoNewline
$size = (Get-Item $EXT_VSIX).Length
Write-Host " OK ($size bytes)" -ForegroundColor Green

# 2. Try to install the extension
Write-Host " [STEP 2] Attempting to install extension in local VS Code..." -NoNewline
& code --install-extension $EXT_VSIX --force 2>$null | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Host " OK" -ForegroundColor Green
} else {
    Write-Host " FAIL" -ForegroundColor Red
    Write-Host " Make sure the 'code' command is in your PATH." -ForegroundColor Yellow
    exit 1
}

# 3. Verify it appears in the list
Write-Host " [STEP 3] Verifying extension registration..." -NoNewline
$list = & code --list-extensions
if ($list -like "*kaffyn.vectora*") {
    Write-Host " OK (kaffyn.vectora is active)" -ForegroundColor Green
} else {
    Write-Host " FAIL (Not found in extension list)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "================================================================" -ForegroundColor Green
Write-Host "        SMOKE TEST PASSED: Extension is properly installable!" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
