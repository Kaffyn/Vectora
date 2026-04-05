# Cross-compilation script for Vectora
# Targets: Windows (amd64, arm), Linux (amd64, arm), macOS (arm64 only)

$binDir = "bin"
if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir -ErrorAction SilentlyContinue | Out-Null
    Write-Host "✓ Diretório $binDir criado" -ForegroundColor Green
}

Write-Host "`n╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║       Vectora Cross-Compilation (Go Native)               ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# Targets
$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Name = "windows-amd64" },
    @{ GOOS = "windows"; GOARCH = "arm"; Name = "windows-arm" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Name = "linux-amd64" },
    @{ GOOS = "linux"; GOARCH = "arm"; Name = "linux-arm" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; Name = "macos-arm64" }
)

# Binários
$binaries = @(
    @{ Name = "vectora"; Path = "cmd/vectora" },
    @{ Name = "vectora-cli"; Path = "cmd/vectora-cli" },
    @{ Name = "lpm"; Path = "cmd/lpm" },
    @{ Name = "mpm"; Path = "cmd/mpm" }
)

# Build each combination
$count = 0
foreach ($target in $targets) {
    Write-Host "📦 Target: $($target.Name)" -ForegroundColor Yellow
    
    foreach ($binary in $binaries) {
        $env:GOOS = $target.GOOS
        $env:GOARCH = $target.GOARCH
        
        $ext = if ($target.GOOS -eq "windows") { ".exe" } else { "" }
        $out = "$binDir/$($binary.Name)-$($target.Name)$ext"
        
        Write-Host "  Building $($binary.Name)..." -NoNewline
        go build -ldflags "-w -s" -o "$out" "./$($binary.Path)" 2>$null
        
        if ($?) {
            $size = (Get-Item $out).Length / 1MB
            Write-Host " ✓ ($([math]::Round($size, 1)) MB)" -ForegroundColor Green
            $count++
        } else {
            Write-Host " ✗" -ForegroundColor Red
        }
    }
}

# Build installer (Windows amd64 only)
Write-Host "`n📦 Target: windows-installer" -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
Write-Host "  Building vectora-installer..." -NoNewline
go build -o "$binDir/vectora-installer-windows-amd64.exe" "./cmd/vectora-installer" 2>$null

if ($?) {
    $size = (Get-Item "$binDir/vectora-installer-windows-amd64.exe").Length / 1MB
    Write-Host " ✓ ($([math]::Round($size, 1)) MB)" -ForegroundColor Green
    $count++
} else {
    Write-Host " ✗" -ForegroundColor Red
}

# Clear env
$env:GOOS = $null
$env:GOARCH = $null

# Summary
Write-Host "`n╔════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║              Build Completo!                              ║" -ForegroundColor Green
Write-Host "╚════════════════════════════════════════════════════════════╝`n" -ForegroundColor Green

$files = Get-ChildItem $binDir -File
$total = ($files | Measure-Object -Property Length -Sum).Sum / 1MB

Write-Host "✅ Binários compilados: $($files.Count)" -ForegroundColor Cyan
Write-Host "📊 Tamanho total: $([math]::Round($total, 1)) MB" -ForegroundColor Cyan
Write-Host "📁 Localização: ./$binDir/" -ForegroundColor Cyan
Write-Host ""
