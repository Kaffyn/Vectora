# Setup Installer Vectora - Especificação Completa

**Status:** Especificação de Arquitetura
**Versão:** 2.0 - Com Integração de LPM/MPM
**Data:** 2026-04-05
**Componente:** `cmd/vectora-setup/` (Fyne GUI)

---

## 1. Visão Geral

O **Vectora Setup Installer** é uma aplicação desktop (Fyne) que guia usuários através da instalação completa do Vectora, incluindo:

1. **Detecção de Hardware** (CPU, RAM, GPU)
2. **Seleção e Download de Llama.cpp** (via LPM CLI)
3. **Seleção e Download de Modelo de IA** (via MPM CLI)
4. **Configuração Inicial** (Gemini API key, etc)
5. **Inicialização do Daemon**

O Setup é **inteligente**: oferece recomendações automáticas baseadas em hardware, mas permite customização manual.

---

## 2. Architecture & Components

### 2.1 Estrutura de Diretórios

```
cmd/vectora-setup/
├── main.go                 # Entry point, inicializa Fyne window
├── wizard.go               # SetupWizard controller (state machine)
├── screens/
│   ├── welcome.go         # Tela 1: Boas-vindas
│   ├── hardware.go        # Tela 2: Detecção
│   ├── build_select.go    # Tela 3: Seleção Llama.cpp
│   ├── build_install.go   # Tela 4: Download Llama.cpp
│   ├── model_select.go    # Tela 5: Seleção Modelo
│   ├── model_install.go   # Tela 6: Download Modelo
│   ├── config.go          # Tela 7: Configuração final
│   └── finished.go        # Tela 8: Sucesso!
│
├── controllers/
│   ├── lpm_controller.go  # Executor de LPM CLI
│   └── mpm_controller.go  # Executor de MPM CLI
│
├── widgets/
│   ├── progress_bar.go    # Barra de progresso
│   ├── hardware_display.go # Exibe specs
│   ├── selector_list.go   # List com radio buttons
│   └── status_box.go      # Box com status/erro
│
└── embed/
    ├── lpm.exe            # LPM binary (embarcado)
    ├── mpm.exe            # MPM binary (embarcado)
    ├── icon.png           # Ícone do instalador
    └── banner.png         # Banner de boas-vindas
```

### 2.2 Fluxo de Telas

```
┌─ Tela 1: Boas-vindas ──────────────────┐
│ • Mostrar logo Vectora                 │
│ • Explicar o que vai acontecer         │
│ • Botões: [Próximo] [Cancelar]        │
└─────────────────────────────────────────┘
                    ↓
┌─ Tela 2: Detecção de Hardware ────────┐
│ • Escanear CPU, RAM, GPU              │
│ • Mostrar: "Detectando..."            │
│ • Exibir specs detectados             │
│ • Botões: [Próximo] [Voltar]         │
└─────────────────────────────────────────┘
                    ↓
┌─ Tela 3: Seleção Llama.cpp ───────────┐
│ • Listar builds disponíveis           │
│ • Mostrar recomendação automática     │
│ • Radio buttons para seleção          │
│ • Permitir "Recomendar automaticamente"│
│ • Botões: [Próximo] [Voltar]         │
└─────────────────────────────────────────┘
                    ↓
┌─ Tela 4: Download Llama.cpp ──────────┐
│ • Mostrar: "Baixando llama-cuda-12..."│
│ • Barra de progresso (%)              │
│ • Tempo restante estimado             │
│ • Velocidade (MB/s)                   │
│ • Botão: [Cancelar]                  │
│ • Depois [Próximo]                    │
└─────────────────────────────────────────┘
                    ↓
┌─ Tela 5: Seleção Modelo de IA ────────┐
│ • Listar modelos Qwen3 disponíveis    │
│ • Mostrar recomendação automática     │
│ • Radio buttons para seleção          │
│ • Permitir "Recomendar automaticamente"│
│ • Botões: [Próximo] [Voltar]         │
└─────────────────────────────────────────┘
                    ↓
┌─ Tela 6: Download Modelo ─────────────┐
│ • Mostrar: "Baixando qwen3-8b..."     │
│ • Barra de progresso (%)              │
│ • Tempo restante estimado             │
│ • Velocidade (MB/s)                   │
│ • Botão: [Cancelar]                  │
│ • Depois [Próximo]                    │
└─────────────────────────────────────────┘
                    ↓
┌─ Tela 7: Configuração Final ──────────┐
│ • Input: Gemini API key (opcional)    │
│ • Checkbox: Criar atalho na área de trab │
│ • Checkbox: Iniciar Vectora agora     │
│ • Botões: [Instalar] [Voltar]        │
└─────────────────────────────────────────┘
                    ↓
┌─ Tela 8: Sucesso! ────────────────────┐
│ • Mostrar: "✓ Vectora instalado!"     │
│ • Log do que foi instalado            │
│ • Botões: [Fechar] [Abrir Vectora]   │
└─────────────────────────────────────────┘
```

---

## 3. Estado Machine (SetupWizard)

### 3.1 Definição de Estado

```go
// cmd/vectora-setup/wizard.go

type SetupWizard struct {
    // Window
    app    fyne.App
    window fyne.Window

    // Estado da instalação
    currentStep int

    // Hardware detectado
    hardware *Hardware

    // Seleções do usuário
    selectedBuild string
    selectedModel string
    geminiKey     string

    // Controllers
    lpmController *LPMController
    mpmController *MPMController

    // Progresso
    isDownloading bool
    downloadProgress float32
    downloadSpeed    float32
    downloadETA      string

    // Callbacks
    onStepChange func(step int)
    onError      func(error)
}

type Hardware struct {
    OS          string      // "windows", "linux", "darwin"
    Arch        string      // "x86_64", "arm64"
    CPUCores    int
    RAM         int64       // em bytes
    GPUType     string      // "none", "cuda", "metal", "vulkan"
    GPUVersion  string      // "12.0", "11.8", etc
    FreeDisk    int64       // em bytes
}
```

### 3.2 Transições de Estado

```
State 0 (Welcome)
    └─→ [Next] → State 1 (Hardware Detection)
    └─→ [Cancel] → Exit

State 1 (Hardware Detection)
    └─→ [Next] → State 2 (Build Selection) [após detectar]
    └─→ [Back] → State 0

State 2 (Build Selection)
    └─→ [Next] → State 3 (Build Install)
    └─→ [Back] → State 1
    └─→ [Recommend] → Auto-select + [Next]

State 3 (Build Install)
    └─→ [After install] → [Next] → State 4 (Model Selection)
    └─→ [Cancel] → State 2 [retry]

State 4 (Model Selection)
    └─→ [Next] → State 5 (Model Install)
    └─→ [Back] → State 3
    └─→ [Recommend] → Auto-select + [Next]

State 5 (Model Install)
    └─→ [After install] → [Next] → State 6 (Config)
    └─→ [Cancel] → State 4 [retry]

State 6 (Config)
    └─→ [Install] → State 7 (Finished)
    └─→ [Back] → State 5

State 7 (Finished)
    └─→ [Close] → Exit
    └─→ [Open Vectora] → Launch daemon + Exit
```

---

## 4. Telas Detalhadas

### 4.1 Tela 1: Boas-vindas

```go
// cmd/vectora-setup/screens/welcome.go

func NewWelcomeScreen() fyne.CanvasObject {
    banner := canvas.NewImageFromFile("embed/banner.png")
    banner.FillMode = canvas.ImageFillContain

    title := widget.NewRichTextFromMarkdown(
        "# Bem-vindo ao Vectora\n\n" +
        "Um assistente de IA que roda 100% localmente no seu computador.",
    )

    description := widget.NewLabel(
        "Este instalador vai:\n\n" +
        "1. Detectar seu hardware\n" +
        "2. Baixar Llama.cpp otimizado\n" +
        "3. Baixar modelo de IA (Qwen3)\n" +
        "4. Configurar tudo automaticamente\n\n" +
        "Processo leva ~10-30 minutos (depende da conexão)",
    )

    nextBtn := widget.NewButton("Próximo", func() {
        wizard.GoToStep(1)
    })

    cancelBtn := widget.NewButton("Cancelar", func() {
        wizard.window.Close()
    })

    buttons := container.NewHBox(nextBtn, cancelBtn)

    return container.NewVBox(
        banner,
        title,
        description,
        container.NewCenter(buttons),
    )
}
```

### 4.2 Tela 2: Detecção de Hardware

```go
// cmd/vectora-setup/screens/hardware.go

func NewHardwareScreen(wizard *SetupWizard) fyne.CanvasObject {
    statusBox := widget.NewLabel("Detectando hardware...")

    hardwareDisplay := container.NewVBox()

    // Começar detecção em goroutine
    go func() {
        // Executar LPM detect
        hw, err := wizard.lpmController.Detect()
        if err != nil {
            statusBox.SetText("❌ Erro ao detectar: " + err.Error())
            return
        }

        wizard.hardware = hw

        // Atualizar UI
        hardwareDisplay.Objects = []fyne.CanvasObject{
            widget.NewLabel("✓ Hardware detectado:"),
            widget.NewLabel(fmt.Sprintf("  OS: %s (%s)", hw.OS, hw.Arch)),
            widget.NewLabel(fmt.Sprintf("  CPU: %d cores", hw.CPUCores)),
            widget.NewLabel(fmt.Sprintf("  RAM: %.1f GB", float64(hw.RAM)/(1024*1024*1024))),
            widget.NewLabel(fmt.Sprintf("  GPU: %s", hw.GPUType)),
            widget.NewLabel(fmt.Sprintf("  Disco livre: %.1f GB", float64(hw.FreeDisk)/(1024*1024*1024))),
        }
        hardwareDisplay.Refresh()

        statusBox.SetText("✓ Hardware detectado com sucesso!")
    }()

    nextBtn := widget.NewButton("Próximo", func() {
        if wizard.hardware == nil {
            wizard.onError(errors.New("Hardware não detectado"))
            return
        }
        wizard.GoToStep(2)
    })

    backBtn := widget.NewButton("Voltar", func() {
        wizard.GoToStep(0)
    })

    return container.NewVBox(
        widget.NewLabel("Detectando hardware..."),
        statusBox,
        hardwareDisplay,
        container.NewHBox(backBtn, nextBtn),
    )
}
```

### 4.3 Tela 3: Seleção Llama.cpp

```go
// cmd/vectora-setup/screens/build_select.go

type BuildSelectScreen struct {
    wizard *SetupWizard
    builds []*Build
    selectedBuild *Build
}

func (s *BuildSelectScreen) Build() fyne.CanvasObject {
    // Carregar lista de builds
    builds, err := s.wizard.lpmController.List()
    if err != nil {
        return container.NewVBox(
            widget.NewLabel("❌ Erro ao carregar builds: " + err.Error()),
            widget.NewButton("Voltar", func() {
                s.wizard.GoToStep(1)
            }),
        )
    }

    s.builds = builds

    title := widget.NewLabel("Selecione Llama.cpp Build")

    // Recomendação automática
    recommendedBuild, _ := s.wizard.lpmController.Recommend()
    recommendText := widget.NewLabel(
        fmt.Sprintf("📍 Recomendado para seu hardware: %s", recommendedBuild.Name),
    )

    // List de builds com radio buttons
    buildItems := container.NewVBox()
    for _, build := range builds {
        item := NewBuildSelectorItem(build, func() {
            s.selectedBuild = build
        })
        buildItems.Add(item)
    }

    scroll := container.NewScroll(buildItems)
    scroll.SetMinSize(fyne.NewSize(500, 300))

    autoBtn := widget.NewButton("Usar Recomendado", func() {
        s.selectedBuild = recommendedBuild
        s.wizard.selectedBuild = recommendedBuild.ID
        s.wizard.GoToStep(3) // Ir para download
    })

    nextBtn := widget.NewButton("Próximo", func() {
        if s.selectedBuild == nil {
            s.wizard.onError(errors.New("Selecione um build"))
            return
        }
        s.wizard.selectedBuild = s.selectedBuild.ID
        s.wizard.GoToStep(3)
    })

    backBtn := widget.NewButton("Voltar", func() {
        s.wizard.GoToStep(1)
    })

    return container.NewVBox(
        title,
        recommendText,
        scroll,
        container.NewHBox(autoBtn),
        container.NewHBox(backBtn, nextBtn),
    )
}
```

### 4.4 Tela 4: Download Llama.cpp

```go
// cmd/vectora-setup/screens/build_install.go

func NewBuildInstallScreen(wizard *SetupWizard) fyne.CanvasObject {
    title := widget.NewLabel(
        fmt.Sprintf("Baixando %s...", wizard.selectedBuild),
    )

    progressBar := widget.NewProgressBar()
    progressBar.Min = 0
    progressBar.Max = 100

    statusLabel := widget.NewLabel("Conectando ao servidor...")
    speedLabel := widget.NewLabel("")
    etaLabel := widget.NewLabel("")

    cancelBtn := widget.NewButton("Cancelar", func() {
        wizard.GoToStep(2) // Voltar para seleção
    })

    nextBtn := widget.NewButton("Próximo", func() {
        wizard.GoToStep(4) // Próxima tela
    })
    nextBtn.Disable()

    // Começar download em goroutine
    go func() {
        err := wizard.lpmController.Install(
            wizard.selectedBuild,
            func(event *ProgressEvent) {
                // Atualizar barra de progresso
                progressBar.Value = float64(event.Percent)
                progressBar.Refresh()

                // Atualizar labels
                statusLabel.SetText(event.Status)
                speedLabel.SetText(fmt.Sprintf("%.1f MB/s", event.Speed))
                etaLabel.SetText(fmt.Sprintf("⏱️ ETA: %s", event.ETA))

                // Se completo
                if event.Percent == 100 {
                    statusLabel.SetText("✓ Download completo!")
                    cancelBtn.Disable()
                    nextBtn.Enable()
                }
            },
        )

        if err != nil {
            statusLabel.SetText("❌ Erro: " + err.Error())
            cancelBtn.Enable()
        }
    }()

    return container.NewVBox(
        title,
        widget.NewLabel(fmt.Sprintf("Tamanho: ~2.5 GB")),
        progressBar,
        statusLabel,
        speedLabel,
        etaLabel,
        container.NewHBox(cancelBtn, nextBtn),
    )
}
```

### 4.5 Tela 5: Seleção de Modelo

Similar à tela 3, mas para modelos (MPM em vez de LPM).

```go
func NewModelSelectScreen(wizard *SetupWizard) fyne.CanvasObject {
    // Similar a BuildSelectScreen, mas usando mpmController
    // Listar modelos com mpm list --json
    // Recomendar com mpm recommend --json
}
```

### 4.6 Tela 6: Download Modelo

Similar à tela 4, mas para modelos.

```go
func NewModelInstallScreen(wizard *SetupWizard) fyne.CanvasObject {
    // Similar a BuildInstallScreen, mas usando mpmController.Install()
}
```

### 4.7 Tela 7: Configuração Final

```go
// cmd/vectora-setup/screens/config.go

func NewConfigScreen(wizard *SetupWizard) fyne.CanvasObject {
    title := widget.NewLabel("Configuração Final")

    // Input para Gemini API key (opcional)
    geminiLabel := widget.NewLabel("Gemini API Key (opcional):")
    geminiInput := widget.NewEntry()
    geminiInput.PlaceHolder = "Deixar em branco para usar apenas Qwen3 local"

    geminiHelp := widget.NewLabel(
        "💡 Para usar Gemini Vision (análise de imagens), cole sua API key aqui.\n" +
        "   Você pode configurar depois nas configurações do Vectora.",
    )

    // Checkboxes
    shortcutCheck := widget.NewCheck(
        "Criar atalho na área de trabalho",
        func(b bool) {},
    )
    shortcutCheck.SetChecked(true)

    launchCheck := widget.NewCheck(
        "Iniciar Vectora após instalação",
        func(b bool) {},
    )
    launchCheck.SetChecked(true)

    installBtn := widget.NewButton("Instalar", func() {
        wizard.geminiKey = geminiInput.Text

        // Salvar configuração
        if geminiInput.Text != "" {
            saveGeminiKey(geminiInput.Text)
        }

        // Criar atalho se selecionado
        if shortcutCheck.Checked {
            createDesktopShortcut()
        }

        // Iniciar daemon se selecionado
        if launchCheck.Checked {
            startVectoraDaemon()
        }

        wizard.GoToStep(7)
    })

    backBtn := widget.NewButton("Voltar", func() {
        wizard.GoToStep(5)
    })

    return container.NewVBox(
        title,
        geminiLabel,
        geminiInput,
        geminiHelp,
        widget.NewSeparator(),
        shortcutCheck,
        launchCheck,
        container.NewHBox(backBtn, installBtn),
    )
}
```

### 4.8 Tela 8: Sucesso!

```go
// cmd/vectora-setup/screens/finished.go

func NewFinishedScreen(wizard *SetupWizard) fyne.CanvasObject {
    successIcon := canvas.NewText("✓", color.White)
    successIcon.TextSize = 48

    title := widget.NewRichTextFromMarkdown(
        "# Parabéns!\n\n**Vectora foi instalado com sucesso!**",
    )

    summary := container.NewVBox(
        widget.NewLabel("📦 Instalado:"),
        widget.NewLabel(fmt.Sprintf("  • Llama.cpp %s", wizard.selectedBuild)),
        widget.NewLabel(fmt.Sprintf("  • Modelo %s", wizard.selectedModel)),
        widget.NewLabel("  • Vectora Daemon"),
        widget.NewLabel("  • Vectora App"),
    )

    nextSteps := widget.NewRichTextFromMarkdown(
        "## Próximos passos:\n\n" +
        "1. Clique em **Abrir Vectora** para iniciar\n" +
        "2. Acesse http://localhost:3000 no navegador\n" +
        "3. Comece a fazer perguntas!\n\n" +
        "Para mais informações, visite https://github.com/Kaffyn/Vectora",
    )

    closeBtn := widget.NewButton("Fechar", func() {
        wizard.window.Close()
    })

    openBtn := widget.NewButton("Abrir Vectora", func() {
        startVectoraDaemon()
        time.Sleep(2 * time.Second)
        openBrowser("http://localhost:3000")
        wizard.window.Close()
    })

    return container.NewVBox(
        container.NewCenter(successIcon),
        title,
        widget.NewSeparator(),
        summary,
        widget.NewSeparator(),
        nextSteps,
        container.NewHBox(closeBtn, openBtn),
    )
}
```

---

## 5. Controllers para CLI Execution

### 5.1 LPM Controller

```go
// cmd/vectora-setup/controllers/lpm_controller.go

type LPMController struct {
    lpmPath string
}

func NewLPMController() *LPMController {
    // Desembarcar LPM do embedding
    lpmPath := filepath.Join(os.TempDir(), "lpm.exe")
    os.WriteFile(lpmPath, lpmBinary, 0755)

    return &LPMController{
        lpmPath: lpmPath,
    }
}

func (c *LPMController) Detect() (*Hardware, error) {
    cmd := exec.Command(c.lpmPath, "detect", "--json")

    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var result struct {
        Data struct {
            Hardware Hardware `json:"hardware"`
        } `json:"data"`
    }

    if err := json.Unmarshal(output, &result); err != nil {
        return nil, err
    }

    return &result.Data.Hardware, nil
}

func (c *LPMController) List() ([]*Build, error) {
    cmd := exec.Command(c.lpmPath, "list", "--json")

    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var result struct {
        Data struct {
            Builds []*Build `json:"builds"`
        } `json:"data"`
    }

    if err := json.Unmarshal(output, &result); err != nil {
        return nil, err
    }

    return result.Data.Builds, nil
}

func (c *LPMController) Recommend() (*Build, error) {
    cmd := exec.Command(c.lpmPath, "recommend", "--json")

    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var result struct {
        Data struct {
            Build *Build `json:"build"`
        } `json:"data"`
    }

    if err := json.Unmarshal(output, &result); err != nil {
        return nil, err
    }

    return result.Data.Build, nil
}

func (c *LPMController) Install(
    buildID string,
    onProgress func(*ProgressEvent),
) error {
    cmd := exec.Command(c.lpmPath, "install", "--id", buildID, "--json")

    // Monitorar stdout para progresso
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }

    if err := cmd.Start(); err != nil {
        return err
    }

    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        line := scanner.Bytes()

        var progress ProgressEvent
        if err := json.Unmarshal(line, &progress); err != nil {
            continue
        }

        onProgress(&progress)
    }

    return cmd.Wait()
}

type Build struct {
    ID           string `json:"id"`
    Name         string `json:"name"`
    Version      string `json:"version"`
    OS           string `json:"os"`
    Arch         string `json:"arch"`
    GPU          string `json:"gpu"`
    GPUVersion   string `json:"gpu_version"`
    SizeBytes    int64  `json:"size_bytes"`
    Installed    bool   `json:"installed"`
    Active       bool   `json:"active"`
    Description  string `json:"description"`
}

type ProgressEvent struct {
    Percent int     `json:"percent"`
    Status  string  `json:"status"`
    Speed   float32 `json:"speed_mbps"`
    ETA     string  `json:"eta"`
    Downloaded int64 `json:"downloaded_bytes"`
    Total   int64   `json:"total_bytes"`
}
```

### 5.2 MPM Controller

Idêntico a LPMController, mas usa `mpmPath` em vez de `lpmPath` e chama MPM em vez de LPM.

```go
// cmd/vectora-setup/controllers/mpm_controller.go

type MPMController struct {
    mpmPath string
}

func NewMPMController() *MPMController {
    // Desembarcar MPM do embedding
    mpmPath := filepath.Join(os.TempDir(), "mpm.exe")
    os.WriteFile(mpmPath, mpmBinary, 0755)

    return &MPMController{
        mpmPath: mpmPath,
    }
}

// Métodos idênticos aos de LPMController
// detect(), list(), recommend(), install()
```

---

## 6. Build Integration

### 6.1 Compilation Steps (build.ps1)

```powershell
# [6/11] Build LPM
go build -o ./build/lpm.exe ./cmd/lpm

# [7/11] Build MPM
go build -o ./build/mpm.exe ./cmd/mpm

# [8/11] Copy para embedding
Copy-Item ./build/lpm.exe ./cmd/vectora-setup/embed/
Copy-Item ./build/mpm.exe ./cmd/vectora-setup/embed/

# [9/11] Build Setup Installer (Fyne)
go build -o ./build/vectora-setup.exe ./cmd/vectora-setup
```

### 6.2 Embedding Assets

```go
// cmd/vectora-setup/main.go

import (
    _ "embed"
)

//go:embed embed/lpm.exe
var lpmBinary []byte

//go:embed embed/mpm.exe
var mpmBinary []byte

//go:embed embed/icon.png
var iconBytes []byte

//go:embed embed/banner.png
var bannerBytes []byte
```

---

## 7. Error Handling & Recovery

### 7.1 Erros Comuns

| Erro | Solução |
|------|---------|
| Network timeout | Retry com backoff exponencial |
| SHA256 mismatch | Delete arquivo, retry download |
| Disk space insufficient | Mostrar mensagem, pedir liberar espaço |
| LPM/MPM não encontrado | Re-desembarcar do embed |
| Invalid JSON | Log e mostrar erro genérico |

### 7.2 Retry Logic

```go
func (c *LPMController) InstallWithRetry(
    buildID string,
    maxRetries int,
    onProgress func(*ProgressEvent),
) error {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        err := c.Install(buildID, onProgress)
        if err == nil {
            return nil
        }

        lastErr = err

        if !isRetryable(err) {
            return err
        }

        // Backoff exponencial
        backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        time.Sleep(backoff)
    }

    return lastErr
}

func isRetryable(err error) bool {
    // Erros de rede são retryable
    // Erros de validation não são
    return strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "connection")
}
```

---

## 8. Platform-Specific Implementation

### 8.1 Windows (Recomendado)

```powershell
# Installer é .exe standalone
# Fyne suporta bem em Windows
# Embedding funciona perfeitamente
```

### 8.2 macOS

```bash
# Compilar como .app bundle
# Code signing (opcional)
# Notarização (opcional)
```

### 8.3 Linux

```bash
# Compilar como binary standalone
# Pode distribuir como AppImage
```

---

## 9. Testing Strategy

### 9.1 Unit Tests

```bash
go test ./cmd/vectora-setup/controllers
go test ./cmd/vectora-setup/screens
```

### 9.2 Integration Tests

- Setup com LPM offline (mock)
- Setup com MPM offline (mock)
- Detector de hardware
- JSON parsing

### 9.3 E2E Tests (Manual)

- Executar setup completo
- Verificar arquivos instalados
- Verificar daemon inicia
- Verificar app abre

---

## 10. Success Criteria

### Funcionalidade
- ✅ Detecção automática de hardware
- ✅ Recomendação inteligente de build/modelo
- ✅ Download robusto com retry
- ✅ Interface amigável (Fyne)
- ✅ Progresso visual
- ✅ Tratamento de erros
- ✅ Criação de atalho
- ✅ Inicialização automática do daemon

### UX
- ✅ Setup fluidez (não trava)
- ✅ Tempos estimados precisos
- ✅ Mensagens de erro claras
- ✅ Sugestões para resolver problemas
- ✅ Opções para customização

### Quality
- ✅ Sem memory leaks
- ✅ Timeouts implementados
- ✅ Testes passando
- ✅ CI/CD integrado

---

**Versão:** 2.0 - Com integração de LPM/MPM (CLI)
**Última Atualização:** 2026-04-05
**Status:** Pronto para implementação
