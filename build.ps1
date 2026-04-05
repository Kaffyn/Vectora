# 0. Ambiente de Compilação (Enforce Architecture)
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# 1. Limpeza
Write-Host "[1/10] Limpando binarios antigos..." -ForegroundColor Gray
Remove-Item vectora.exe, vectora-app.exe, vectora-cli.exe, lpm.exe, vectora-setup.exe, vectora-tests.exe -ErrorAction SilentlyContinue

# 2. Build do Frontend (Next.js via Bun)
Write-Host "[2/10] Exportando Frontend (Next.js SSG via Bun)..." -ForegroundColor Yellow
Push-Location internal/app
# Garante que o Bun esteja pronto e o build ocorra para exportação
bun install
bun run build
if (-not (Test-Path "out")) { throw "FALHA: Pasta 'out' não gerada pelo Next.js." }
Pop-Location

# 3. Build do App Desktop (Wails)
Write-Host "[3/10] Compilando App Desktop (Wails Interface)..." -ForegroundColor Yellow
Push-Location cmd/vectora-app
wails build -clean -platform windows/amd64 -o vectora-app.exe
Pop-Location
if (-not (Test-Path "cmd/vectora-app/build/bin/vectora-app.exe")) { throw "FALHA: vectora-app.exe não foi gerado pelo Wails." }
Copy-Item "cmd/vectora-app/build/bin/vectora-app.exe" "vectora-app.exe" -Force

# 4. Build do Daemon (vectora.exe)
Write-Host "[4/10] Compilando Daemon (Vectora Engine)..." -ForegroundColor Yellow
go build -ldflags="-s -w -H windowsgui" -o vectora.exe ./cmd/vectora
if (-not (Test-Path "vectora.exe")) { throw "FALHA: vectora.exe não foi gerado." }

# 5. Teste do Llama.cpp Package Manager
Write-Host "[5/10] Testando Llama.cpp Package Manager (internal/engines)..." -ForegroundColor Yellow
go test -v ./internal/engines -short
if ($LASTEXITCODE -ne 0) { throw "FALHA: Testes do engines package falharam." }
Write-Host "✅ Engines package manager testado com sucesso!" -ForegroundColor Green

# 6. Build do LPM (Llama Package Manager - CLI)
Write-Host "[6/10] Compilando LPM - Llama Package Manager (cmd/lpm)..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o lpm.exe ./cmd/lpm
if (-not (Test-Path "lpm.exe")) { throw "FALHA: lpm.exe não foi gerado." }
Write-Host "✅ LPM compilado com sucesso!" -ForegroundColor Green

# 7. Build do CLI (wrapper que chama o Daemon)
Write-Host "[7/10] Compilando CLI (wrapper para vectora chat)..." -ForegroundColor Yellow
go build -o vectora-cli.exe ./cmd/vectora-cli
if (-not (Test-Path "vectora-cli.exe")) { throw "FALHA: vectora-cli.exe não foi gerado." }
Write-Host "✅ CLI compilado com sucesso!" -ForegroundColor Green

# Sincronização de Staging para o Instalador Principal (Fyne)
Write-Host "Sincronizando binarios para o Instalador..." -ForegroundColor Gray
Copy-Item "vectora.exe" "cmd/vectora-installer/vectora.exe" -Force
Copy-Item "vectora-app.exe" "cmd/vectora-installer/vectora-app.exe" -Force
Copy-Item "vectora-cli.exe" "cmd/vectora-installer/vectora-cli.exe" -Force
Copy-Item "lpm.exe" "cmd/vectora-installer/lpm.exe" -Force

# 8. Build do Instalador (UAC + Admin)
Write-Host "[8/10] Compilando Instalador Administrativo (Fyne)..." -ForegroundColor Yellow
Push-Location cmd/vectora-installer
go run github.com/tc-hib/go-winres@latest make --in winres/winres.json
Pop-Location
go build -ldflags="-s -w -H windowsgui" -o vectora-setup.exe ./cmd/vectora-installer
if (-not (Test-Path "vectora-setup.exe")) { throw "FALHA: vectora-setup.exe não foi gerado." }

# 9. Build dos Testes
Write-Host "[9/10] Compilando Suite de Testes (cmd/tests)..." -ForegroundColor Yellow
go build -o vectora-tests.exe ./cmd/tests
if (-not (Test-Path "vectora-tests.exe")) { throw "FALHA: vectora-tests.exe não foi gerado." }

# 10. Geração de Manifestos SHA256
Write-Host "`n[10/10] Gerando Manifestos SHA256..." -ForegroundColor Cyan
if (!(Test-Path "assets")) { New-Item -ItemType Directory -Path "assets" | Out-Null }
$hashes = @{}
$hashes["vectora.exe"] = (Get-FileHash vectora.exe -Algorithm SHA256).Hash.ToLower()
$hashes["vectora-app.exe"] = (Get-FileHash vectora-app.exe -Algorithm SHA256).Hash.ToLower()
$hashes["vectora-cli.exe"] = (Get-FileHash vectora-cli.exe -Algorithm SHA256).Hash.ToLower()
$hashes["lpm.exe"] = (Get-FileHash lpm.exe -Algorithm SHA256).Hash.ToLower()
$hashes["vectora-setup.exe"] = (Get-FileHash vectora-setup.exe -Algorithm SHA256).Hash.ToLower()
$hashes["vectora-tests.exe"] = (Get-FileHash vectora-tests.exe -Algorithm SHA256).Hash.ToLower()
$hashes | ConvertTo-Json | Out-File "assets/binaries.sha256" -Encoding utf8

Write-Host "`n--- FULL STACK BUILD CONCLUIDO COM SUCESSO! ---" -ForegroundColor Green
Write-Host "✅ Daemon (vectora.exe) compilado e pronto" -ForegroundColor Green
Write-Host "✅ App Desktop (vectora-app.exe) compilado e pronto" -ForegroundColor Green
Write-Host "✅ CLI (vectora-cli.exe) compilado e pronto" -ForegroundColor Green
Write-Host "✅ LPM - Llama Package Manager (lpm.exe) compilado e pronto" -ForegroundColor Green
Write-Host "✅ Llama.cpp Package Manager (internal/engines) testado e integrado" -ForegroundColor Green
Write-Host "✅ Testes (vectora-tests.exe) compilados e prontos" -ForegroundColor Green
Write-Host "✅ Instalador (vectora-setup.exe) compilado e pronto" -ForegroundColor Green
Write-Host "`nBinarios prontos na raiz (vectora.exe, vectora-app.exe, vectora-cli.exe, lpm.exe, vectora-setup.exe, vectora-tests.exe)" -ForegroundColor Gray
