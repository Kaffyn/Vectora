package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Kaffyn/vectora/assets"
	"github.com/Kaffyn/vectora/internal/i18n"
	"github.com/jeandeaual/go-locale"
	"golang.org/x/sys/windows/registry"
)

var installPath string

func main() {
	isUninstallMode := false
	if len(os.Args) >= 2 && os.Args[1] == "--uninstall" {
		isUninstallMode = true
	}

	a := app.New()
	w := a.NewWindow("Vectora Setup Engine")
	w.Resize(fyne.NewSize(600, 350))
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

	var showStep1, showStepPath, showStep2, showStep3 func()
	var showAlreadyInstalled, showUninstallProgress func(string)

	// ---- PASSO: Desinstalação Oficial UI ----
	showUninstallProgress = func(existingPath string) {
		title := widget.NewLabelWithStyle("Uninstalling / Removendo Vectora...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		progress := widget.NewProgressBar()
		
		btn := widget.NewButton("Close / Sair", func() {
			w.Close()
		})
		btn.Hide()

		go func() {
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 60)
				progress.SetValue(float64(i) / 20.0)
			}
			
			registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`)
			
			daemonPath := filepath.Join(existingPath, "vectora.exe")
			os.Remove(daemonPath)
			
			// Deleta o Ícone do Menu Iniciar!
			removeStartMenuShortcut()
			
			progress.SetValue(1.0)
			title.SetText("Successfully uninstalled / Desinstalado com sucesso.")
			btn.Show()
			title.Refresh()
		}()

		w.SetContent(container.NewVBox(
			widget.NewLabel(""),
			title,
			widget.NewLabel(""),
			progress,
			widget.NewLabel(""),
			container.NewCenter(btn),
		))
	}

	// ---- PASSO 0: Verificação Ativa ----
	showAlreadyInstalled = func(existingPath string) {
		lbl := widget.NewLabelWithStyle(i18n.T("inst_already"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		
		btnUninstall := widget.NewButton(i18n.T("inst_btn_uninstall"), func() {
			showUninstallProgress(existingPath)
		})
		
		btnCancel := widget.NewButton("Cancel / Sair", func() {
			w.Close()
		})
		
		content := container.NewVBox(
			widget.NewLabel(""),
			lbl,
			widget.NewLabel(""),
			container.NewCenter(container.NewHBox(btnUninstall, btnCancel)),
		)
		w.SetContent(content)
	}

	showStep1 = func() {
		welcome := widget.NewLabelWithStyle(i18n.T("inst_welcome"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
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
			showStep1()
		}

		btn := widget.NewButton(i18n.T("inst_btn_next"), func() {
			showStepPath()
		})

		content := container.NewVBox(
			welcome,
			widget.NewLabel(""), 
			widget.NewLabelWithStyle(i18n.T("inst_select_lang"), fyne.TextAlignCenter, fyne.TextStyle{}),
			container.NewCenter(langSelect),
			widget.NewLabel(""),
			container.NewCenter(btn),
		)
		w.SetContent(content)
	}

	showStepPath = func() {
		title := widget.NewLabelWithStyle(i18n.T("inst_step_path"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		
		home, _ := os.UserHomeDir()
		defaultPath := filepath.Join(home, "AppData", "Local", "Programs", "Vectora")
		installPath = defaultPath

		pathEntry := widget.NewEntry()
		pathEntry.SetText(defaultPath)

		btnFolder := widget.NewButton(i18n.T("inst_btn_browse"), func() {
			dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
				if lu != nil {
					pathEntry.SetText(lu.Path())
					installPath = lu.Path()
				}
			}, w)
		})

		btnNext := widget.NewButton(i18n.T("inst_btn_next"), func() {
			installPath = pathEntry.Text
			showStep2() 
		})

		content := container.NewVBox(
			title,
			widget.NewLabel(""),
			pathEntry,
			btnFolder,
			widget.NewLabel(""),
			container.NewCenter(btnNext),
		)
		w.SetContent(content)
	}

	showStep2 = func() {
		title := widget.NewLabelWithStyle(i18n.T("inst_step_engine"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

		radio := widget.NewRadioGroup([]string{
			i18n.T("inst_desc_gemini"),
			i18n.T("inst_desc_qwen"),
		}, func(s string) {})
		radio.SetSelected(i18n.T("inst_desc_gemini"))

		btn := widget.NewButton(i18n.T("inst_btn_next"), func() {
			showStep3()
		})

		content := container.NewVBox(
			title,
			widget.NewLabel(""),
			radio,
			widget.NewLabel(""),
			container.NewCenter(btn),
		)
		w.SetContent(content)
	}

	// ---- PASSO 3: Instalação & Engine Cópias e Atalhos ----
	showStep3 = func() {
		title := widget.NewLabelWithStyle("...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)
		
		btn := widget.NewButton(i18n.T("inst_btn_finish"), func() {
			w.Close()
		})
		btn.Hide()

		go func() {
			title.SetText(i18n.T("inst_step_download"))
			title.Refresh()
			
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 120)
				progress.SetValue(float64(i) / 20.0)
			}
			
			os.MkdirAll(installPath, 0755)
			srcApp, _ := os.Executable()
			
			// 1. Instala o Desinstalador copiando o instalador para o diretório de destino
			copySysFile(srcApp, filepath.Join(installPath, "vectora-uninstaller.exe"))

			// 2. Transfere o Daemon real (vectora.exe) se ele estiver ao lado do Installer local!
			srcDaemon := filepath.Join(filepath.Dir(srcApp), "vectora.exe")
			if _, err := os.Stat(srcDaemon); err == nil {
				copySysFile(srcDaemon, filepath.Join(installPath, "vectora.exe"))
			}

			// 3. Cadastra o App formalmente para ele ganhar ícone e Start Menu Searches
			registerWindowsApp(installPath)
			createStartMenuShortcut(installPath)

			progress.SetValue(1.0)
			
			doneTxt := i18n.T("inst_done")
			if doneTxt == "inst_done" {
				doneTxt = "Instalação Concluída com Sucesso!"
			}
			title.SetText(doneTxt)
			btn.Show()
			title.Refresh()
		}()

		content := container.NewVBox(
			title,
			widget.NewLabel(""),
			progress,
			widget.NewLabel(""),
			container.NewCenter(btn),
		)
		w.SetContent(content)
	}

	if isUninstallMode {
		existingPath := checkInstalledPath()
		if existingPath != "" {
			showUninstallProgress(existingPath)
		} else {
			// Se o registry falhar o path atual serve como safety net fallback
			execPath, _ := os.Executable()
			showUninstallProgress(filepath.Dir(execPath))
		}
	} else if existingPath := checkInstalledPath(); existingPath != "" {
		showAlreadyInstalled(existingPath)
	} else {
		showStep1()
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
		
		// Mapeando do vectora.exe instalado oficial
		targetExe := filepath.Join(dst, "vectora.exe")
		key.SetStringValue("DisplayIcon", targetExe)
		key.SetStringValue("UninstallString", filepath.Join(dst, "vectora-uninstaller.exe")+" --uninstall")
		key.SetStringValue("InstallLocation", dst)
	}
}

func createStartMenuShortcut(installDir string) {
	appData := os.Getenv("APPDATA")
	if appData == "" { return }
	
	programsDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Vectora")
	os.MkdirAll(programsDir, 0755)

	shortcutPath := filepath.Join(programsDir, "Vectora.lnk")
	targetPath := filepath.Join(installDir, "vectora.exe")
	
	// OLE Automation wrapper invisível via PS pra criar Symlink Win32 sem sujar dependências CGO
	script := `
$wshell = New-Object -ComObject WScript.Shell
$shortcut = $wshell.CreateShortcut("` + shortcutPath + `")
$shortcut.TargetPath = "` + targetPath + `"
$shortcut.WorkingDirectory = "` + installDir + `"
$shortcut.IconLocation = "` + targetPath + `,0"
$shortcut.Save()
`
	exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script).Run()
}

func removeStartMenuShortcut() {
	appData := os.Getenv("APPDATA")
	if appData == "" { return }
	programsDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Vectora")
	os.RemoveAll(programsDir)
}
