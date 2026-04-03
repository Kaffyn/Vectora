# 0. Ambiente de Compilação (Enforce Architecture)
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# 1. Limpeza
Write-Host "[1/8] Limpando binarios antigos..." -ForegroundColor Gray
Remove-Item vectora.exe, vectora-web.exe, vectora-installer.exe, vectora-tests.exe, llama-installer.exe -ErrorAction SilentlyContinue

# 2. Build do Frontend (Next.js via Bun)
Write-Host "[2/8] Exportando Frontend (Next.js SSG via Bun)..." -ForegroundColor Yellow
Push-Location internal/app
# Garante que o Bun esteja pronto e o build ocorra para exportação
bun install
bun run build
if (-not (Test-Path "out")) { throw "FALHA: Pasta 'out' não gerada pelo Next.js." }
Pop-Location

# 3. Build do App Desktop (Wails)
Write-Host "[3/8] Compilando App Desktop (Wails Interface)..." -ForegroundColor Yellow
# O Wails build na pasta raiz lê o wails.json que já configuramos
wails build -clean -platform windows/amd64 -o vectora-web.exe
if (-not (Test-Path "build/bin/vectora-web.exe")) { throw "FALHA: vectora-web.exe não foi gerado pelo Wails." }
Copy-Item "build/bin/vectora-web.exe" "vectora-web.exe" -Force

# 4. Build do Daemon (vectora.exe)
Write-Host "[4/8] Compilando Daemon (Vectora Engine)..." -ForegroundColor Yellow
go build -ldflags="-s -w -H windowsgui" -o vectora.exe ./cmd/vectora
if (-not (Test-Path "vectora.exe")) { throw "FALHA: vectora.exe não foi gerado." }

# 5. Build do Llama-Installer
Write-Host "[5/8] Compilando Llama-Installer (Standalone)..." -ForegroundColor Yellow
if (!(Test-Path "cmd/llama/assets")) { New-Item -ItemType Directory -Path "cmd/llama/assets" | Out-Null }
Copy-Item "internal/os/windows/llama-b8583/*" "cmd/llama/assets" -Force
go build -o llama-installer.exe ./cmd/llama
if (-not (Test-Path "llama-installer.exe")) { throw "FALHA: llama-installer.exe não foi gerado." }
Remove-Item -Path "cmd/llama/assets" -Recurse -Force

# Sincronização de Staging para o Instalador Principal (Fyne)
Copy-Item "vectora.exe" "cmd/vectora-installer/vectora.exe" -Force
Copy-Item "vectora-web.exe" "cmd/vectora-installer/vectora-web.exe" -Force
Copy-Item "llama-installer.exe" "cmd/vectora-installer/llama-installer.exe" -Force

# 6. Build do Instalador (UAC + Admin)
Write-Host "[6/8] Compilando Instalador Administrativo (Fyne)..." -ForegroundColor Yellow
Push-Location cmd/vectora-installer
go run github.com/tc-hib/go-winres@latest make --in winres/winres.json
Pop-Location
go build -ldflags="-s -w -H windowsgui" -o vectora-installer.exe ./cmd/vectora-installer
if (-not (Test-Path "vectora-installer.exe")) { throw "FALHA: vectora-installer.exe não foi gerado." }

# 7. Build dos Testes
Write-Host "[7/8] Compilando Suite de Testes..." -ForegroundColor Yellow
go build -o vectora-tests.exe ./cmd/tests
if (-not (Test-Path "vectora-tests.exe")) { throw "FALHA: vectora-tests.exe não foi gerado." }

# 8. Geração de Manifestos SHA256
Write-Host "`n[8/8] Gerando Manifestos SHA256..." -ForegroundColor Cyan
if (!(Test-Path "assets")) { New-Item -ItemType Directory -Path "assets" | Out-Null }
$hashes = @{}
$hashes["vectora.exe"] = (Get-FileHash vectora.exe -Algorithm SHA256).Hash.ToLower()
$hashes["vectora-web.exe"] = (Get-FileHash vectora-web.exe -Algorithm SHA256).Hash.ToLower()
$hashes | ConvertTo-Json | Out-File "assets/engines.sha256" -Encoding utf8

Write-Host "`n--- FULL STACK BUILD CONCLUIDO COM SUCESSO! ---" -ForegroundColor Green
Write-Host "Binarios (Daemon, App, CLI, Installer) prontos na raiz." -ForegroundColor Gray
