package main

import (
	"embed"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Kaffyn/Vectora/src/core/config"
)

// -- Tema Customizado Premium (Vectora Dark Aesthetic) --
type vectoraTheme struct{}

var _ fyne.Theme = (*vectoraTheme)(nil)

func (m vectoraTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 13, G: 11, B: 26, A: 255} // #0D0B1A (Fundo Next.js)
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 88, G: 80, B: 236, A: 255} // #5850EC (Vectora Blurple)
	case theme.ColorNameForeground:
		return color.NRGBA{R: 240, G: 240, B: 250, A: 255} // Texto super claro
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 22, G: 19, B: 38, A: 255} // Inputs escuros acinzentados
	case theme.ColorNameButton:
		return color.NRGBA{R: 35, G: 32, B: 55, A: 255} // Botões passivos
	case theme.ColorNameHover:
		return color.NRGBA{R: 55, G: 50, B: 85, A: 255} // Hover suave
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 40, G: 35, B: 65, A: 255}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (m vectoraTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
func (m vectoraTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(s)
}
func (m vectoraTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 8.0
	}
	return theme.DefaultTheme().Size(name)
}

//go:embed bin/*
var embeddedBinaries embed.FS

func main() {
	// Cria o App
	a := app.New()

	// Força o tema Premium Vectora!!
	a.Settings().SetTheme(&vectoraTheme{})

	w := a.NewWindow("Setup Vectora")
	w.Resize(fyne.NewSize(650, 480))

	// Controle de Estado
	userHome, _ := os.UserHomeDir()
	defaultInstallDir := filepath.Join(userHome, "AppData", "Local", "Vectora", "bin")

	// Variáveis de Estado
	pathEntry := widget.NewEntry()
	pathEntry.SetText(defaultInstallDir)

	checkLocal := widget.NewCheck("Modo Local (Qwen3 0.6B - Baixa ~800MB)", nil)
	checkCloud := widget.NewCheck("Modo Cloud (Google Gemini - Leve)", nil)
	geminiKeyEntry := widget.NewPasswordEntry()
	geminiKeyEntry.SetPlaceHolder("Cole sua API Key (AIzaSy...)")
	checkLocal.SetChecked(true)

	progressBar := widget.NewProgressBar()
	statusLabel := widget.NewLabel("Pronto para instalar.")

	// Container Principal (onde os passos mudam)
	contentArea := container.NewMax()

	// Botões da Barra Inferior
	btnBack := widget.NewButton("Voltar", nil)
	var btnNext *widget.Button // Declarado antes para uso cruzado

	// Layout Base do Rodapé
	footer := container.NewBorder(nil, nil, btnBack, nil, container.NewHBox(widget.NewLabel(""), widget.NewLabel("v1.0.0")))

	// ==================== [ PASSO 1: BOAS VINDAS ] ====================
	title1 := canvas.NewText("Vectora Setup", theme.PrimaryColor())
	title1.TextSize = 28
	title1.TextStyle = fyne.TextStyle{Bold: true}

	step1 := container.NewVBox(
		title1,
		widget.NewLabel(""),
		widget.NewLabel("Bem-vindo(a) ao Assistente de Instalação do Vectora."),
		widget.NewLabel("Este assistente ajudará você a configurar o ecossistema RAG\nno seu sistema local."),
		widget.NewLabel(""),
		widget.NewLabel("Clique em Avançar para continuar."),
	)

	// ==================== [ PASSO 2: CAMINHO ] ====================
	title2 := canvas.NewText("Local de Instalação", theme.PrimaryColor())
	title2.TextSize = 22
	title2.TextStyle = fyne.TextStyle{Bold: true}

	step2 := container.NewVBox(
		title2,
		widget.NewLabel("Selecione a pasta onde os executáveis do Vectora serão instalados:"),
		pathEntry,
		widget.NewLabel("O instalador criará automaticamente o diretório caso\nele não exista."),
	)

	// ==================== [ PASSO 3: COGNIÇÃO ] ====================
	title3 := canvas.NewText("Motores Cognitivos", theme.PrimaryColor())
	title3.TextSize = 22
	title3.TextStyle = fyne.TextStyle{Bold: true}

	geminiContainer := container.NewVBox(
		widget.NewLabel("Chave Gemini:"),
		geminiKeyEntry,
	)
	geminiContainer.Hide()

	// Dropdowns da Família Qwen3
	llmSelector := widget.NewSelect([]string{
		"Qwen3-0.6B-GGUF (~600MB RAM)",
		"Qwen3-1.7B-GGUF (~1.5GB RAM)",
		"Qwen3-4B-GGUF (~3.5GB RAM)",
		"Qwen3-Coder-Next-GGUF (Especialista Código)",
	}, nil)
	llmSelector.SetSelected("Qwen3-0.6B-GGUF (~600MB RAM)")

	embSelector := widget.NewSelect([]string{
		"Qwen3-Embedding-0.6B-GGUF (Recomendado)",
		"Qwen3-Embedding-4B-GGUF (Alta Fidelidade)",
	}, nil)
	embSelector.SetSelected("Qwen3-Embedding-0.6B-GGUF (Recomendado)")

	localContainer := container.NewVBox(
		widget.NewLabelWithStyle("Modelo de Linguagem Principal (Chat):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		llmSelector,
		widget.NewLabelWithStyle("Modelo Vetorial p/ RAG (Embeddings):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		embSelector,
		widget.NewLabelWithStyle("Os arquivos GGUF serão injetados pelo llama.cpp nativo.", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
	)

	checkLocal.SetText("Modo Local Habilitado")
	checkLocal.OnChanged = func(checked bool) {
		if checked {
			localContainer.Show()
		} else {
			localContainer.Hide()
		}
	}

	checkCloud.OnChanged = func(checked bool) {
		if checked {
			geminiContainer.Show()
		} else {
			geminiContainer.Hide()
		}
	}

	step3 := container.NewVBox(
		title3,
		widget.NewLabel("Monte a sua suíte neural perfeita. Motores locais exigem Embeddings."),
		checkLocal,
		localContainer,
		widget.NewSeparator(),
		checkCloud,
		geminiContainer,
	)

	// ==================== [ PASSO 4: INSTALAÇÃO ] ====================
	title4 := canvas.NewText("Instalando", theme.PrimaryColor())
	title4.TextSize = 22
	title4.TextStyle = fyne.TextStyle{Bold: true}

	step4 := container.NewVBox(
		title4,
		widget.NewLabel("Aguarde enquanto o Vectora é configurado..."),
		widget.NewLabel(""),
		statusLabel,
		progressBar,
	)

	// ==================== LÓGICA DE NAVEGAÇÃO ====================
	currentStep := 1

	updateUI := func() {
		contentArea.Objects = nil

		switch currentStep {
		case 1:
			contentArea.Add(step1)
			btnBack.Disable()
			btnNext.SetText("Avançar >")
		case 2:
			contentArea.Add(step2)
			btnBack.Enable()
			btnNext.SetText("Avançar >")
		case 3:
			contentArea.Add(step3)
			btnBack.Enable()
			btnNext.SetText("Instalar")
			btnNext.Importance = widget.HighImportance
		case 4:
			contentArea.Add(step4)
			btnBack.Disable()
			btnNext.Disable() // Disable during install
		}
		contentArea.Refresh()
	}

	// Lógica de Extração (Quando clica em Instalar no Step 3)
	performInstallation := func() {
		currentStep = 4
		updateUI()

		go func() {
			targetDir := pathEntry.Text
			vectoraData := filepath.Join(userHome, ".Vectora")

			statusLabel.SetText("Criando arquitetura de pastas...")
			_ = os.MkdirAll(filepath.Join(vectoraData, "models"), 0755)
			_ = os.MkdirAll(targetDir, 0755)

			statusLabel.SetText("Extraindo arquivos embutidos...")
			progressBar.SetValue(0.2)

			entries, err := embeddedBinaries.ReadDir("bin")
			if err == nil {
				total := len(entries)
				for i, entry := range entries {
					if entry.IsDir() || entry.Name() == ".gitkeep" {
						continue
					}
					extracted, _ := embeddedBinaries.Open("bin/" + entry.Name())
					destPath := filepath.Join(targetDir, entry.Name())
					f, _ := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
					if f != nil {
						io.Copy(f, extracted)
						f.Close()
					}
					extracted.Close()
					progressBar.SetValue(0.2 + (float64(i)/float64(total))*0.5)
					time.Sleep(100 * time.Millisecond)
				}
			}

			statusLabel.SetText("Salvando configurações globais...")
			activeProvider := "local"
			if checkCloud.Checked && checkLocal.Checked {
				activeProvider = "hybrid"
			} else if checkCloud.Checked {
				activeProvider = "gemini"
			}

			settings := config.VectoraSettings{
				ActiveProvider: activeProvider,
				GeminiAPIKey:   geminiKeyEntry.Text,
				InstallDir:     targetDir,
			}
			config.SaveSettings(settings)

			if checkCloud.Checked && geminiKeyEntry.Text != "" {
				os.Setenv("GEMINI_API_KEY", geminiKeyEntry.Text)
			}

			statusLabel.SetText("Verificando runtime Bun...")
			progressBar.SetValue(0.9)
			_, errBun := exec.LookPath("bun")
			if errBun != nil {
				cmd := exec.Command("powershell", "-Command", "irm https://bun.sh/install.ps1 | iex")
				_ = cmd.Run()
			}

			progressBar.SetValue(1.0)
			statusLabel.SetText("✅ Vectora Instalado com Sucesso!")

			dialog.ShowInformation("Concluído", "Tudo configurado! Você pode fechar o instalador.", w)

			btnNext.SetText("Concluir e Iniciar")
			btnNext.Importance = widget.HighImportance
			btnNext.Enable()

			btnNext.OnTapped = func() {
				exec.Command(filepath.Join(targetDir, "vectora_tray.exe")).Start()
				a.Quit()
			}
		}()
	}

	btnNext = widget.NewButton("Avançar >", func() {
		if currentStep == 3 {
			if checkCloud.Checked && len(geminiKeyEntry.Text) < 10 {
				dialog.ShowError(fmt.Errorf("A chave da API Gemini parece inválida"), w)
				return
			}
			if !checkCloud.Checked && !checkLocal.Checked {
				dialog.ShowError(fmt.Errorf("Marque pelo menos um provedor"), w)
				return
			}
			performInstallation()
			return
		}

		if currentStep < 3 {
			currentStep++
			updateUI()
		}
	})

	btnBack.OnTapped = func() {
		if currentStep > 1 {
			currentStep--
			updateUI()
			btnNext.Importance = widget.MediumImportance
		}
	}

	// Layout Principal
	footer = container.NewBorder(nil, nil, btnBack, btnNext, widget.NewLabel("Vectora Engine Ecosystem"))

	// Efeito visual com fundo leve e bordas
	mainContainer := container.NewBorder(
		widget.NewSeparator(),                            // Topo
		container.NewVBox(widget.NewSeparator(), footer), // Rodapé
		nil, nil,
		container.NewPadded(contentArea),
	)

	updateUI() // Iniciar no Passo 1
	w.SetContent(mainContainer)
	w.ShowAndRun()
}
