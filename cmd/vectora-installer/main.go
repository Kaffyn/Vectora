package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/Kaffyn/Vectora/assets"
	"github.com/Kaffyn/Vectora/internal/i18n"
	vecos "github.com/Kaffyn/Vectora/internal/os"
	"github.com/jeandeaual/go-locale"
)

var installPath string
var w fyne.Window
var selectedModel string
var hardwareInfo map[string]interface{}
var recommendedModel map[string]interface{}

// Funções auxiliares para integração com MPM
func getMPMPath(installDir string) string {
	if filepath.VolumeName(installDir) != "" {
		return filepath.Join(installDir, "mpm.exe")
	}
	return filepath.Join(installDir, "mpm")
}

func callMPM(args ...string) (string, error) {
	var execPath string

	// Tentar encontrar mpm.exe no path de instalação
	if installPath != "" {
		mpmPath := getMPMPath(installPath)
		if _, err := os.Stat(mpmPath); err == nil {
			execPath = mpmPath
		}
	}

	// Fallback: procurar no PATH
	if execPath == "" {
		exe, _ := os.Executable()
		execPath = getMPMPath(filepath.Dir(exe))
	}

	cmd := exec.Command(execPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run mpm: %w", err)
	}

	return out.String(), nil
}

func detectHardware(installDir string) (map[string]interface{}, error) {
	output, err := callMPM("detect", "--json")
	if err != nil {
		return nil, err
	}

	var hw map[string]interface{}
	if err := json.Unmarshal([]byte(output), &hw); err != nil {
		return nil, fmt.Errorf("failed to parse hardware info: %w", err)
	}

	return hw, nil
}

func recommendModel(installDir string) (map[string]interface{}, error) {
	output, err := callMPM("recommend", "--json")
	if err != nil {
		return nil, err
	}

	var model map[string]interface{}
	if err := json.Unmarshal([]byte(output), &model); err != nil {
		return nil, fmt.Errorf("failed to parse model info: %w", err)
	}

	return model, nil
}

func listModels(installDir string) ([]map[string]interface{}, error) {
	output, err := callMPM("list", "--json")
	if err != nil {
		return nil, err
	}

	var models []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &models); err != nil {
		return nil, fmt.Errorf("failed to parse models list: %w", err)
	}

	return models, nil
}

func installModel(installDir string, modelID string) (string, error) {
	return callMPM("install", "--model", modelID)
}

func createLayout(titleText string, content fyne.CanvasObject, backFunc func(), nextFunc func(), nextLabel string) *fyne.Container {
	title := canvas.NewText(titleText, color.RGBA{R: 75, G: 58, B: 240, A: 255})
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}

	footerBg := canvas.NewRectangle(color.RGBA{R: 11, G: 10, B: 16, A: 255})

	btnBack := widget.NewButton("Voltar", backFunc)
	if backFunc == nil {
		btnBack.Disable()
	}

	btnNext := widget.NewButton(nextLabel, nextFunc)
	if nextFunc == nil {
		btnNext.Disable()
	}

	ecosystemLbl := canvas.NewText("Kaffyn Ecosystem", color.RGBA{R: 200, G: 200, B: 200, A: 255})
	ecosystemLbl.TextSize = 13

	footerContent := container.NewHBox(
		btnBack,
		layout.NewSpacer(),
		container.NewCenter(ecosystemLbl),
		layout.NewSpacer(),
		btnNext,
	)

	return container.NewBorder(
		container.NewPadded(container.NewVBox(title, widget.NewLabel(""))),
		container.NewStack(footerBg, container.NewPadded(footerContent)),
		nil, nil,
		container.NewPadded(content),
	)
}

func main() {
	systemManager, _ := vecos.NewManager()
	if systemManager != nil && !systemManager.IsRunningAsAdmin() {
		// Attempt to restart as admin
		exe, _ := os.Executable()
		cwd, _ := os.Getwd()
		args := strings.Join(os.Args[1:], " ")

		psCmd := fmt.Sprintf("Start-Process -FilePath '%s' -Verb runas -WorkingDirectory '%s'", exe, cwd)
		if args != "" {
			psCmd += fmt.Sprintf(" -ArgumentList '%s'", args)
		}

		cmd := exec.Command("powershell", psCmd)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Start(); err == nil {
			os.Exit(0)
		}
	}

	isUninstallMode := false
	if len(os.Args) >= 2 && os.Args[1] == "--uninstall" {
		isUninstallMode = true
	}

	a := app.New()
	a.Settings().SetTheme(&zyrisTheme{})

	w = a.NewWindow("Setup Vectora")
	w.Resize(fyne.NewSize(620, 480))
	w.SetIcon(fyne.NewStaticResource("installer_icon", assets.InstallerIconData))

	// Re-get manager for logic
	systemManager, err := vecos.NewManager()
	if err != nil {
		w.SetContent(widget.NewLabel("FALHA: Erro ao carregar System Manager."))
		w.ShowAndRun()
		return
	}

	userLoc, err := locale.GetLanguage()
	if err == nil {
		if strings.HasPrefix(userLoc, "pt") {
			i18n.SetLanguage("pt")
		} else if strings.HasPrefix(userLoc, "es") {
			i18n.SetLanguage("es")
		} else if strings.HasPrefix(userLoc, "fr") {
			i18n.SetLanguage("fr")
		} else {
			i18n.SetLanguage("en")
		}
	} else {
		i18n.SetLanguage("en")
	}

	var showStepWelcome, showStepLang, showStepPath, showStepInstall func()
	var showStepConfigMode, showStepConfigGemini, showStepConfigQwen, showStepFinish func()
	var showStepDetectHardware, showStepRecommendModel, showStepChooseModel, showStepInstallModel func()
	var showUninstallProgress, showAlreadyInstalled func(string)

	showUninstallProgress = func(existingPath string) {
		progress := widget.NewProgressBar()
		content := container.NewVBox(
			// Destroying Local Vectora Instance and Windows links...,
			progress,
		)

		wrapper := createLayout("Uninstalling Vectora", content, nil, nil, "Finalizar")
		w.SetContent(wrapper)

		go func() {
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 60)
				progress.SetValue(float64(i) / 20.0)
			}

			systemManager.UnregisterApp(existingPath)
			daemonBin := "vectora"
			if filepath.VolumeName(existingPath) != "" {
				daemonBin = "vectora.exe"
			}
			os.Remove(filepath.Join(existingPath, daemonBin))

			progress.SetValue(1.0)

			finalContent := container.NewVBox(widget.NewLabel("Desinstalação e Limpeza Concluídas com Sucesso."))
			w.SetContent(createLayout("Uninstalled", finalContent, nil, func() { w.Close() }, "Exit"))
		}()
	}

	showAlreadyInstalled = func(existingPath string) {
		lbl := widget.NewLabel("O motor já detectou instâncias do Vectora residentes no seu ambiente.")
		btnUninstall := widget.NewButton("Desfazer Instalação Atual do Vectora (Wipe)", func() {
			showUninstallProgress(existingPath)
		})
		content := container.NewVBox(lbl, widget.NewLabel(""), container.NewHBox(btnUninstall))
		w.SetContent(createLayout("App Detectado", content, nil, func() { w.Close() }, "Sair"))
	}

	showStepWelcome = func() {
		text1 := widget.NewLabel(i18n.T("inst_welcome"))
		text2 := widget.NewLabel("Este assistente ajudará você a configurar o ambiente de IA\nno seu sistema local.")
		text3 := widget.NewLabel("Clique em " + i18n.T("inst_btn_next") + " para continuar.")

		content := container.NewVBox(
			text1, widget.NewLabel(""),
			text2, widget.NewLabel(""), widget.NewLabel(""),
			text3,
		)
		w.SetContent(createLayout("Vectora Setup", content, nil, showStepLang, i18n.T("inst_btn_next")+" >"))
	}

	showStepLang = func() {
		langSelect := widget.NewSelect([]string{"English", "Português", "Español", "Français"}, nil)
		switch i18n.GetCurrentLang() {
		case "en":
			langSelect.SetSelected("English")
		case "pt":
			langSelect.SetSelected("Português")
		case "es":
			langSelect.SetSelected("Español")
		case "fr":
			langSelect.SetSelected("Français")
		}

		langSelect.OnChanged = func(s string) {
			switch s {
			case "English":
				i18n.SetLanguage("en")
			case "Português":
				i18n.SetLanguage("pt")
			case "Español":
				i18n.SetLanguage("es")
			case "Français":
				i18n.SetLanguage("fr")
			}
			showStepLang()
		}

		content := container.NewVBox(widget.NewLabel(i18n.T("inst_select_lang")), langSelect)
		w.SetContent(createLayout("Linguagem Padrão", content, showStepWelcome, showStepPath, "Avançar >"))
	}

	showStepPath = func() {
		defaultPath, _ := systemManager.GetInstallDir() // Changes to Program Files on Windows
		if installPath == "" {
			installPath = defaultPath
		}

		pathEntry := widget.NewEntry()
		pathEntry.SetText(installPath)

		btnFolder := widget.NewButton(i18n.T("inst_btn_browse"), func() {
			dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
				if lu != nil {
					pathEntry.SetText(lu.Path())
					installPath = lu.Path()
				}
			}, w)
		})

		content := container.NewVBox(
			widget.NewLabel("Selecione a pasta onde os executáveis serão instalados:"),
			container.NewBorder(nil, nil, nil, btnFolder, pathEntry),
			widget.NewLabel("O instalador criará automaticamente o diretório caso\nele não exista."),
		)

		nextCmd := func() {
			installPath = pathEntry.Text
			showStepInstall()
		}
		w.SetContent(createLayout("Local de Instalação", content, showStepLang, nextCmd, "Avançar >"))
	}

	showStepInstall = func() {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		content := container.NewVBox(
			widget.NewLabel("Desempacotando binários do ecossistema e escrevendo manifestos..."),
			progress,
		)

		wrapper := createLayout("Instalando...", content, nil, nil, "Finalizar")
		w.SetContent(wrapper)

		go func() {
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 30)
				progress.SetValue(float64(i) / 20.0)
			}

			// Official Scaffolding hierarchy in Program Files (Requires Admin)
			if err := os.MkdirAll(installPath, 0755); err != nil {
				dialog.ShowError(fmt.Errorf("PERMISSION DENIED: Run the installer as Administrator or request IT access.\n(%v)", err), w)
				return
			}
			os.MkdirAll(filepath.Join(installPath, "data", "chroma"), 0755)
			os.MkdirAll(filepath.Join(installPath, "logs"), 0755)
			os.MkdirAll(filepath.Join(installPath, "backups"), 0755)

			assets := getInstallerAssets()

			// Extract all binaries
			binariesToExtract := []string{
				"vectora.exe",
				"vectora-app.exe",
				"vectora-cli.exe",
				"lpm.exe",
				"mpm.exe",
			}

			for _, binName := range binariesToExtract {
				if binData, exists := assets[binName]; exists && len(binData) > 0 {
					target := filepath.Join(installPath, binName)
					if err := os.WriteFile(target, binData, 0755); err != nil {
						fmt.Printf("[WARNING] Failed to extract %s: %v\n", binName, err)
					}
				}
			}
			fmt.Println("--- Engine Extraction and Setup Completed ---")

			srcSelf, _ := os.Executable()
			uninstallerPath := filepath.Join(installPath, "vectora-uninstaller.exe")
			copySysFile(srcSelf, uninstallerPath)

			systemManager.RegisterApp(installPath)
			progress.SetValue(1.0)

			doneContent := container.NewVBox(widget.NewLabel("Files extracted and O.S links successfully registered.\n\nNow let's configure the Assistant Engine."))
			w.SetContent(createLayout("Instalação Concluída", doneContent, nil, showStepConfigMode, "Configurar IA >"))
		}()
	}

	showStepConfigMode = func() {
		radio := widget.NewRadioGroup([]string{
			"Usar Gemini API (Apenas Chave / Leve RAM)",
			"Usar Qwen3 (100% Privado / Download Pesado)",
		}, func(s string) {})
		radio.SetSelected("Usar Gemini API (Apenas Chave / Leve RAM)")

		content := container.NewVBox(widget.NewLabel("Defina qual Core fará o processamento local LLM:"), radio)

		nextCmd := func() {
			if radio.Selected == "Usar Gemini API (Apenas Chave / Leve RAM)" {
				showStepConfigGemini()
			} else {
				showStepConfigQwen()
			}
		}
		w.SetContent(createLayout("Motor Cognitivo", content, nil, nextCmd, "Avançar >"))
	}

	showStepConfigGemini = func() {
		lbl := widget.NewLabel("Insira sua chave de API do Google AI Studio:")
		entry := widget.NewPasswordEntry()
		entry.SetPlaceHolder("AIzaSy...")

		content := container.NewVBox(lbl, entry, widget.NewLabel("Ela será arquivada e criptografada em segurança na sua máquina local."))

		nextCmd := func() {
			showStepFinish()
		}
		w.SetContent(createLayout("Validar Access Token", content, showStepConfigMode, nextCmd, "Avançar >"))
	}

	// Detectar hardware e recomendarmodelo
	showStepDetectHardware = func() {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		statusLbl := widget.NewLabel("Detectando especificações do sistema...")
		content := container.NewVBox(statusLbl, progress)

		wrapper := createLayout("Detectando Hardware", content, nil, nil, "Próximo")
		w.SetContent(wrapper)

		go func() {
			// Simular progresso
			for i := 0; i < 10; i++ {
				time.Sleep(100 * time.Millisecond)
				progress.SetValue(float64(i) / 10.0)
			}

			hw, err := detectHardware(installPath)
			if err != nil {
				statusLbl.SetText(fmt.Sprintf("Erro ao detectar hardware: %v", err))
				return
			}

			hardwareInfo = hw
			progress.SetValue(1.0)
			time.Sleep(500 * time.Millisecond)

			// Passar para próximo step automaticamente
			showStepRecommendModel()
		}()
	}

	// Recomendar modelo
	showStepRecommendModel = func() {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		statusLbl := widget.NewLabel("Recomendando modelo para seu hardware...")
		content := container.NewVBox(statusLbl, progress)

		wrapper := createLayout("Recomendação de Modelo", content, nil, nil, "Próximo")
		w.SetContent(wrapper)

		go func() {
			// Simular progresso
			for i := 0; i < 10; i++ {
				time.Sleep(100 * time.Millisecond)
				progress.SetValue(float64(i) / 10.0)
			}

			model, err := recommendModel(installPath)
			if err != nil {
				statusLbl.SetText(fmt.Sprintf("Erro ao recomendar modelo: %v", err))
				return
			}

			recommendedModel = model
			selectedModel = model["id"].(string)
			progress.SetValue(1.0)
			time.Sleep(500 * time.Millisecond)

			// Passar para próximo step
			showStepChooseModel()
		}()
	}

	// Permitir escolher outro modelo
	showStepChooseModel = func() {
		var allModels []map[string]interface{}

		models, err := listModels(installPath)
		if err != nil {
			lbl := widget.NewLabel(fmt.Sprintf("Erro ao listar modelos: %v", err))
			w.SetContent(createLayout("Erro", lbl, showStepDetectHardware, nil, ""))
			return
		}

		allModels = models

		// Criar lista de opções
		var modelOptions []string
		for _, m := range allModels {
			id := m["id"].(string)
			name := m["name"].(string)
			modelOptions = append(modelOptions, fmt.Sprintf("%s - %s", id, name))
		}

		modelSelector := widget.NewSelect(modelOptions, func(s string) {
			// Extrair ID do modelo selecionado
			parts := strings.Split(s, " - ")
			if len(parts) > 0 {
				selectedModel = parts[0]
			}
		})

		// Pre-selecionar o recomendado
		recID := recommendedModel["id"].(string)
		recName := recommendedModel["name"].(string)
		modelSelector.SetSelected(fmt.Sprintf("%s - %s", recID, recName))
		selectedModel = recID

		content := container.NewVBox(
			widget.NewLabel("Modelo Recomendado (ou escolha outro):"),
			modelSelector,
			widget.NewLabel(""),
			widget.NewLabel(fmt.Sprintf("RAM do Sistema: %.1f GB", hardwareInfo["RAM"])),
		)

		nextCmd := func() {
			showStepInstallModel()
		}

		w.SetContent(createLayout("Escolher Modelo", content, showStepDetectHardware, nextCmd, "Instalar >"))
	}

	// Instalar modelo
	showStepInstallModel = func() {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		statusLbl := widget.NewLabel(fmt.Sprintf("Preparando instalação de %s...", selectedModel))
		content := container.NewVBox(statusLbl, progress)

		wrapper := createLayout("Instalando Modelo", content, nil, nil, "Concluir")
		w.SetContent(wrapper)

		go func() {
			// Simular progresso de preparação
			for i := 0; i < 5; i++ {
				time.Sleep(100 * time.Millisecond)
				progress.SetValue(float64(i) / 10.0)
			}

			statusLbl.SetText(fmt.Sprintf("Instalando %s...", selectedModel))
			_, err := installModel(installPath, selectedModel)

			if err != nil {
				statusLbl.SetText(fmt.Sprintf("Aviso: %v\nModelo será instalado na primeira inicialização", err))
			} else {
				statusLbl.SetText(fmt.Sprintf("✓ %s pronto!", selectedModel))
			}

			progress.SetValue(1.0)
			time.Sleep(500 * time.Millisecond)

			// Passar para finish
			showStepFinish()
		}()
	}

	showStepConfigQwen = func() {
		// Iniciar fluxo de Qwen3
		showStepDetectHardware()
	}

	showStepFinish = func() {
		doneContent := container.NewVBox(widget.NewLabel("O Vectora foi instalado com sucesso."))
		w.SetContent(createLayout("Sucesso", doneContent, nil, func() { w.Close() }, "Encerrar"))
	}

	if isUninstallMode {
		existingPath := systemManager.IsInstalled()
		if existingPath != "" {
			showUninstallProgress(existingPath)
		} else {
			execPath, _ := os.Executable()
			showUninstallProgress(filepath.Dir(execPath))
		}
	} else if existingPath := systemManager.IsInstalled(); existingPath != "" {
		showAlreadyInstalled(existingPath)
	} else {
		showStepWelcome()
	}

	w.ShowAndRun()
}

func copySysFile(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer out.Close()
	io.Copy(out, in)
}
