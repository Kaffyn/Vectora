package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/Kaffyn/vectora/assets"
	"github.com/Kaffyn/vectora/internal/i18n"
	"github.com/jeandeaual/go-locale"
	"golang.org/x/sys/windows/registry"
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
	isUninstallMode := false
	if len(os.Args) >= 2 && os.Args[1] == "--uninstall" {
		isUninstallMode = true
	}

	a := app.New()
	a.Settings().SetTheme(&zyrisTheme{}) 
	
	w = a.NewWindow("Setup Vectora")
	w.Resize(fyne.NewSize(620, 480))
	w.SetIcon(fyne.NewStaticResource("logo", assets.IconData))

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

	var showStepWelcome, showStepLang, showStepPath, showStepEngine, showStepInstall func()
	var showUninstallProgress, showAlreadyInstalled func(string)

	showUninstallProgress = func(existingPath string) {
		progress := widget.NewProgressBar()
		content := container.NewVBox(
			widget.NewLabel("Destruindo Instância Local do Vectora e Links do Windows..."),
			progress,
		)

		var backFunc func() = nil
		var nextFunc func() = nil 

		wrapper := createLayout("Uninstalling Vectora", content, backFunc, nextFunc, "Finalizar")
		w.SetContent(wrapper)

		go func() {
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 60)
				progress.SetValue(float64(i) / 20.0)
			}
			
			registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`)
			os.Remove(filepath.Join(existingPath, "vectora.exe"))
			removeStartMenuShortcut()
			
			progress.SetValue(1.0)
			
			finalContent := container.NewVBox(widget.NewLabel("Desinstalação e Limpeza Concluídas com Sucesso."))
			w.SetContent(createLayout("Desinstalado", finalContent, nil, func(){ w.Close() }, "Vazar"))
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
		text1 := widget.NewLabel("Bem-vindo(a) ao Assistente de Instalação do Vectora.")
		text2 := widget.NewLabel("Este assistente ajudará você a configurar o ambiente de IA\nno seu sistema local.")
		text3 := widget.NewLabel("Clique em Avançar para continuar.")

		content := container.NewVBox(
			text1, widget.NewLabel(""),
			text2, widget.NewLabel(""), widget.NewLabel(""),
			text3,
		)
		
		w.SetContent(createLayout("Vectora Setup", content, nil, showStepLang, "Avançar >"))
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

		content := container.NewVBox(
			widget.NewLabel(i18n.T("inst_select_lang")),
			langSelect,
		)
		
		w.SetContent(createLayout("Linguagem Padrão", content, showStepWelcome, showStepPath, "Avançar >"))
	}

	showStepPath = func() {
		home, _ := os.UserHomeDir()
		defaultPath := filepath.Join(home, "AppData", "Local", "Programs", "Vectora")
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
			showStepEngine()
		}

		w.SetContent(createLayout("Local de Instalação", content, showStepLang, nextCmd, "Avançar >"))
	}

	showStepEngine = func() {
		radio := widget.NewRadioGroup([]string{
			i18n.T("inst_desc_gemini"),
			i18n.T("inst_desc_qwen"),
		}, func(s string) {})
		radio.SetSelected(i18n.T("inst_desc_gemini"))

		content := container.NewVBox(widget.NewLabel("Defina qual Core fará o processamento local LLM:"), radio)
		w.SetContent(createLayout("Motor Cognitivo", content, showStepPath, showStepInstall, "Avançar >"))
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
				time.Sleep(time.Millisecond * 120)
				progress.SetValue(float64(i) / 20.0)
			}
			
			os.MkdirAll(installPath, 0755)
			srcApp, _ := os.Executable()
			copySysFile(srcApp, filepath.Join(installPath, "vectora-uninstaller.exe"))

			srcDaemon := filepath.Join(filepath.Dir(srcApp), "vectora.exe")
			if _, err := os.Stat(srcDaemon); err == nil {
				copySysFile(srcDaemon, filepath.Join(installPath, "vectora.exe"))
			}

			registerWindowsApp(installPath)
			createStartMenuShortcut(installPath)

			progress.SetValue(1.0)
			
			doneContent := container.NewVBox(widget.NewLabel("Ambiente construído fisicamente no seu Driver C:\\ com sucesso!"))
			w.SetContent(createLayout("Sucesso", doneContent, nil, func(){ w.Close() }, "Encerrar"))
		}()
	}

	if isUninstallMode {
		existingPath := checkInstalledPath()
		if existingPath != "" {
			showUninstallProgress(existingPath)
		} else {
			execPath, _ := os.Executable()
			showUninstallProgress(filepath.Dir(execPath))
		}
	} else if existingPath := checkInstalledPath(); existingPath != "" {
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

func checkInstalledPath() string {
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
	if err == nil {
		defer key.Close()
		val, _, err := key.GetStringValue("InstallLocation")
		if err == nil && val != "" {
			return val
		}
	}
	return ""
}

func registerWindowsApp(dst string) {
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`
	key, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.ALL_ACCESS)
	if err == nil {
		defer key.Close()
		key.SetStringValue("DisplayName", "Vectora")
		key.SetStringValue("DisplayVersion", "1.0.0")
		key.SetStringValue("Publisher", "Kaffyn")
		key.SetStringValue("DisplayIcon", filepath.Join(dst, "vectora.exe"))
		key.SetStringValue("UninstallString", filepath.Join(dst, "vectora-uninstaller.exe")+" --uninstall")
		key.SetStringValue("InstallLocation", dst)
	}
}

func createStartMenuShortcut(installDir string) {
	appData := os.Getenv("APPDATA")
	if appData == "" { return }
	programsDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Vectora")
	os.MkdirAll(programsDir, 0755)

	script := `
$wshell = New-Object -ComObject WScript.Shell
$shortcut = $wshell.CreateShortcut("` + filepath.Join(programsDir, "Vectora.lnk") + `")
$shortcut.TargetPath = "` + filepath.Join(installDir, "vectora.exe") + `"
$shortcut.WorkingDirectory = "` + installDir + `"
$shortcut.IconLocation = "` + filepath.Join(installDir, "vectora.exe") + `,0"
$shortcut.Save()
`
	exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script).Run()
}

func removeStartMenuShortcut() {
	appData := os.Getenv("APPDATA")
	if appData == "" { return }
	os.RemoveAll(filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Vectora"))
}
