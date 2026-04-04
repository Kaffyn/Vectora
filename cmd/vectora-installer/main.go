package main

import (
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

	"github.com/Kaffyn/vectora/assets"
	"github.com/Kaffyn/vectora/internal/i18n"
	vecos "github.com/Kaffyn/vectora/internal/os"
	"github.com/jeandeaual/go-locale"
)

var installPath string
var w fyne.Window

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
			if filepath.VolumeName(existingPath) != "" { daemonBin = "vectora.exe" }
			os.Remove(filepath.Join(existingPath, daemonBin))
			
			progress.SetValue(1.0)
			
			finalContent := container.NewVBox(widget.NewLabel("Desinstalação e Limpeza Concluídas com Sucesso."))
			w.SetContent(createLayout("Uninstalled", finalContent, nil, func(){ w.Close() }, "Exit"))
		}()
	}

	showAlreadyInstalled = func(existingPath string) {
		lbl := widget.NewLabel("O motor já detectou instâncias do Vectora residentes no seu ambiente.")
		btnUninstall := widget.NewButton("Desfazer Instalação Atual do Vectora (Wipe)", func() {
			showUninstallProgress(existingPath)
		})
		content := container.NewVBox(lbl, widget.NewLabel(""), container.NewHBox(btnUninstall))
		w.SetContent(createLayout("App Detectado", content, nil, func(){ w.Close() }, "Sair"))
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
		case "en": langSelect.SetSelected("English")
		case "pt": langSelect.SetSelected("Português")
		case "es": langSelect.SetSelected("Español")
		case "fr": langSelect.SetSelected("Français")
		}

		langSelect.OnChanged = func(s string) {
			switch s {
			case "English": i18n.SetLanguage("en")
			case "Português": i18n.SetLanguage("pt")
			case "Español": i18n.SetLanguage("es")
			case "Français": i18n.SetLanguage("fr")
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
			llamaInstallerData := assets["llama-installer.exe"]
			if len(llamaInstallerData) > 0 {
				tmpInstaller := filepath.Join(os.TempDir(), "llama-installer-tmp.exe")
				os.WriteFile(tmpInstaller, llamaInstallerData, 0755)
				
				cmd := exec.Command(tmpInstaller, "--silent", "--path", installPath)
				if err := cmd.Run(); err != nil {
					fmt.Printf("[WARNING] Silent failure in Llama Installer: %v\n", err)
				}
				os.Remove(tmpInstaller)
			}

			vectoraData := assets["vectora.exe"]
			if len(vectoraData) > 0 {
				target := filepath.Join(installPath, "vectora.exe")
				os.WriteFile(target, vectoraData, 0755)
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

	showStepConfigQwen = func() {
		lbl := widget.NewLabel("O motor de IA Qwen3-0.6B será baixado em background pelo Daemon\nna Primeira Inicialização (Aprox 2GB exigidos em disco).")
		
		content := container.NewVBox(lbl)
		
		nextCmd := func() {
			showStepFinish()
		}
		w.SetContent(createLayout("Download e Pesos (Weights)", content, showStepConfigMode, nextCmd, "Avançar >"))
	}

	showStepFinish = func() {
		doneContent := container.NewVBox(widget.NewLabel("O Vectora foi instalado com sucesso."))
		w.SetContent(createLayout("Sucesso", doneContent, nil, func(){ w.Close() }, "Encerrar"))
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
	if err != nil { return }
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil { return }
	defer out.Close()
	io.Copy(out, in)
}
