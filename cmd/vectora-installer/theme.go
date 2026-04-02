package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type zyrisTheme struct{}

func (m zyrisTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return color.RGBA{R: 14, G: 13, B: 21, A: 255}    // #0E0D15 - Fundo Geral Dark Navy
	case theme.ColorNamePrimary:
		return color.RGBA{R: 75, G: 58, B: 240, A: 255}   // #4B3AF0 - Blurple Primary
	case theme.ColorNameButton:
		return color.RGBA{R: 28, G: 27, B: 41, A: 255}    // #1C1B29 - Botão Escuro
	case theme.ColorNameHover:
		return color.RGBA{R: 45, G: 43, B: 66, A: 255}    // Hover sobre componentes
	case theme.ColorNameForeground:
		return color.RGBA{R: 230, G: 230, B: 230, A: 255} // Texto claro
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 20, G: 19, B: 30, A: 255}    // #14131E - Input Entry
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 35, G: 34, B: 45, A: 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 22, G: 21, B: 30, A: 255} 
	default:
		return theme.DefaultTheme().Color(n, theme.VariantDark)
	}
}

func (m zyrisTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m zyrisTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 12
	}
	return theme.DefaultTheme().Size(name)
}

func (m zyrisTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
