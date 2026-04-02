package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Kaffyn/vectora/internal/i18n"
)

func main() {
	a := app.New()
	w := a.NewWindow("Vectora Installer")
	w.Resize(fyne.NewSize(600, 350))

	// Idioma default OS inferido, mantendo fixo em PT por hora
	i18n.SetLanguage("pt")

	var showStep1, showStep2, showStep3 func()

	// ---- PASSO 1: Boas Vindas & Idioma ----
	showStep1 = func() {
		welcome := widget.NewLabelWithStyle(i18n.T("inst_welcome"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		
		langSelect := widget.NewSelect([]string{"English", "Português", "Español", "Français"}, nil)
		
		// Set current selection base (before assigning hook to prevent infinite loop)
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
			showStep1() // Reemite a view pra renderizar traduções vivas
		}

		btn := widget.NewButton(i18n.T("inst_btn_next"), func() {
			showStep2()
		})

		content := container.NewVBox(
			welcome,
			widget.NewLabel(""), // Spacer
			widget.NewLabelWithStyle("Select Installation Language:", fyne.TextAlignCenter, fyne.TextStyle{}),
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
				time.Sleep(time.Millisecond * 120) // Mocking real download delay
				progress.SetValue(i)
			}
			title.SetText("Instalação Concluída e Daemon Configurado! / Done!")
		}()

		btn := widget.NewButton(i18n.T("inst_btn_finish"), func() {
			w.Close() // Fecha e finaliza pipeline, acionando shell pro Daemon subir normal
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

	// Lança pipeline
	showStep1()
	w.ShowAndRun()
}
