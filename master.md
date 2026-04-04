# VECTORA: MASTER IMPLEMENTATION PLAN

Este documento é a fonte autoritativa de verdade para o ecossistema Vectora. Ele detalha a evolução da arquitetura, o estado atual de cada componente e a sequência lógica de expansão.

---

## 1. O Caminho Até Aqui (Histórico de Estabilização)

O projeto evoluiu de uma arquitetura fragmentada para um ecossistema unificado e profissional no Windows.

### 1.1 Unificação de Backend & Frontend
- **Ação**: O backend (Daemon) e o frontend (Next.js) foram integrados via IPC (Named Pipes) para eliminar a dependência de servidores HTTP intermediários em produção.
- **Resultado**: Performance de latência zero e segurança via pipes nativos do sistema.

### 1.2 Estabilização do Pipeline de Build (Wails v2)
- **Ação**: O `wails.json` foi movido para `cmd/vectora-app` e os caminhos relativos de assets (`../../internal/app/out`) foram recalibrados.
- **Resultado**: O Wails CLI agora identifica corretamente o ponto de entrada Go, gerando o `vectora-app.exe` de forma consistente.

### 1.3 Padronização de Módulo (Case-Sensitivity)
- **Ação**: O módulo Go foi padronizado globalmente como `github.com/Kaffyn/Vectora` (Maiúsculo).
- **Resultado**: Fim dos conflitos de "module not found" e erros de compilador no Windows.

### 1.4 Segurança e Privilégios (UAC Auto-Elevation)
- **Ação**: Injetada lógica de auto-elevação UAC via PowerShell `runas` em todos os pontos de entrada (`vectora.exe`, `vectora-app.exe`, `vectora-setup.exe`).
- **Resultado**: Os binários agora solicitam privilégios administrativos no início (se necessário), garantindo acesso a `Program Files` e manipulação de serviços.

---

## 2. Estado Atual do Projeto

Todos os binários abaixo foram verificados e estão operacionais com suporte a ícones e metadados profissionais (Kaffyn 1.0.0.0).

| Binário | Caminho Fonte | Função | Status |
|---------|---------------|--------|--------|
| `vectora.exe` | `cmd/vectora/` | Motor/Daemon + Tray Icon | **ESTÁVEL** (Inicia tray por padrão) |
| `vectora-app.exe` | `cmd/vectora-app/` | Interface Wails (Next.js SSG) | **ESTÁVEL** |
| `vectora-setup.exe` | `cmd/vectora-installer/` | Instalador/Desinstalador Fyne | **ESTÁVEL** (Auto-UAC) |
| `llama-installer.exe` | `cmd/llama/` | Instalador de Motor Legado | **OBSOLETO** (Será substituído pelo LPM) |

---

## 3. Próxima Fase: LLAMA PACKAGE MANAGER (LPM)

Esta é a nossa prioridade atual: desacoplar o motor Llama do instalador principal.

### 3.1 Arquitetura LPM (`internal/engines/`)
- `catalog.json`: Catálogo de versões b8583 (CUDA, Vulkan, AVX2).
- `detector.go`: Detecção de hardware (GPU/CPU).
- `downloader.go`: Motor de download com suporte a **Resume (.partial)**.
- `process.go`: Comunicação via **STDIO JSON-ND** (Sem portas TCP).

### 3.2 Metas Sequenciais
1. [ ] Implementar Catálogo e Detector de Hardware.
2. [ ] Substituir o `llama-installer.exe` legado pela lógica do LPM.
3. [ ] Atualizar o `build.ps1` para remover binários embutidos de Llama.

---

## 4. Caminhos de Infraestrutura

- **AppData**: `%USERPROFILE%/.Vectora` (Configurações, Logs, .env).
- **Engines**: `%USERPROFILE%/.Vectora/packages/llama/` (Binários baixados pelo LPM).
- **Install Dir**: `C:\Program Files\Vectora` (Executáveis do sistema).

---
*Atualizado em: 2026-04-03 21:23 - Antigravity Engine*
