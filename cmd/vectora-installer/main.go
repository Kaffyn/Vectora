package main

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Kaffyn/vectora/assets"
	"github.com/Kaffyn/vectora/internal/i18n"
	"github.com/jeandeaual/go-locale"
)

func main() {
	a := app.New()
	w := a.NewWindow("Vectora Installer")
	w.Resize(fyne.NewSize(600, 350))

	// Associa o resource icone compativel à Window Tool do Fyne (Para Desktop Taskbar e Window Chrome)
	w.SetIcon(fyne.NewStaticResource("logo", assets.IconData))

	// Define idioma pautado na API Nativa do S.O (Windows GetSystemDefaultUILanguage ou registry)
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
		// Fallback internacional seguro
		i18n.SetLanguage("en")
	}

	var showStep1, showStep2, showStep3 func()

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
			showStep2()
		})

		content := container.NewVBox(
			welcome,
			widget.NewLabel(""), // Spacer
			widget.NewLabelWithStyle(i18n.T("inst_select_lang"), fyne.TextAlignCenter, fyne.TextStyle{}),
			container.NewCenter(langSelect),
			widget.NewLabel(""), // Spacer
			container.NewCenter(btn),
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

	// ---- PASSO 3: Download & Workspace Config ----
	showStep3 = func() {
		title := widget.NewLabelWithStyle(i18n.T("inst_step_download"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		progress := widget.NewProgressBar()
		
		go func() {
			for i := 0.0; i <= 1.0; i += 0.05 {
				time.Sleep(time.Millisecond * 120)
				progress.SetValue(i)
			}
			title.SetText(i18n.T("inst_done"))
		}()

		btn := widget.NewButton(i18n.T("inst_btn_finish"), func() {
			w.Close()
		})

		content := container.NewVBox(
			title,
			widget.NewLabel(""),
			progress,
			widget.NewLabel(""),
			container.NewCenter(btn),
		)
		w.SetContent(content)
	}

	showStep1()
	w.ShowAndRun()
}
