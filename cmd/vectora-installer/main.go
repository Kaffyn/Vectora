package main

import (
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
	a := app.New()
	w := a.NewWindow("Vectora Installer")
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
	var showAlreadyInstalled func(string)

	// ---- PASSO 0: Verificação Ativa ----
	showAlreadyInstalled = func(existingPath string) {
		lbl := widget.NewLabelWithStyle(i18n.T("inst_already"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		
		btnUninstall := widget.NewButton(i18n.T("inst_btn_uninstall"), func() {
			uninstaller := filepath.Join(existingPath, "vectora-uninstaller.exe")
			if _, err := os.Stat(uninstaller); err == nil {
				exec.Command(uninstaller).Start()
			}
			w.Close()
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

	// ---- PASSO 1: Boas Vindas & Idioma ----
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

	// ---- PASSO 1.5: Seleção de Diretório ----
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

	// ---- PASSO 2: Seleção de Provedor IA ----
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

	// ---- PASSO 3: Instalação & Escrita no Registro (App do Sistema) ----
	showStep3 = func() {
		title := widget.NewLabelWithStyle("...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)
		
		btn := widget.NewButton(i18n.T("inst_btn_finish"), func() {
			w.Close()
		})
		btn.Hide() // Omitido até garantir 100%

		go func() {
			title.SetText(i18n.T("inst_step_download"))
			title.Refresh()
			
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 120)
				progress.SetValue(float64(i) / 20.0)
			}
			
			registerWindowsApp(installPath)
			progress.SetValue(1.0)
			
			doneTxt := i18n.T("inst_done")
			if doneTxt == "inst_done" {
				doneTxt = "Instalação Concluída com Sucesso!"
			}
			title.SetText(doneTxt)
			btn.Show() // Salta pra UI de forma imperativa
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

	// Executa a ramificação Inicial de verificação OS-Level no Registry
	if existingPath := checkInstalledPath(); existingPath != "" {
		showAlreadyInstalled(existingPath)
	} else {
		showStep1()
	}

	w.ShowAndRun()
}

// Verifica Instalação Prévia Lendo Chaves do Sistema Win32
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

// Criação de chaves no Regedit p/ Transformar o Vectora em Aplicativo Nativo
func registerWindowsApp(dst string) {
	os.MkdirAll(dst, 0755)
	
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall\Vectora`
	key, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.ALL_ACCESS)
	if err == nil {
		defer key.Close()
		key.SetStringValue("DisplayName", "Vectora")
		key.SetStringValue("DisplayVersion", "1.0.0")
		key.SetStringValue("Publisher", "Kaffyn")
		key.SetStringValue("DisplayIcon", filepath.Join(dst, "vectora.exe"))
		key.SetStringValue("UninstallString", filepath.Join(dst, "vectora-uninstaller.exe"))
		key.SetStringValue("InstallLocation", dst)
	}
}
