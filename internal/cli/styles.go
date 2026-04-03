package cli

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Kaffyn Color Palette
	Blue      = lipgloss.Color("#4B3AF0")
	DeepBlue  = lipgloss.Color("#2E23B0")
	Purple    = lipgloss.Color("#7D42D1")
	LightGray = lipgloss.Color("#C8C8C8")
	DarkGray  = lipgloss.Color("#1B1B1B")
	BgColor   = lipgloss.Color("#0B0A10")
	Accent    = lipgloss.Color("#00FFD1")

	// Estilos do Terminal
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Blue).
			Padding(1, 2)

	StyleStatus = lipgloss.NewStyle().
			Foreground(LightGray).
			PaddingLeft(2)

	StyleInput = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DeepBlue).
			Padding(0, 1)

	StyleBotMsg = lipgloss.NewStyle().
			Foreground(LightGray).
			Border(lipgloss.Border{Left: "┃"}).
			BorderForeground(Purple).
			PaddingLeft(2).
			MarginBottom(1)

	StyleUserMsg = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true).
			MarginTop(1)

	StyleModelBadge = lipgloss.NewStyle().
			Background(DeepBlue).
			Foreground(LightGray).
			Padding(0, 1).
			Bold(true).
			MarginLeft(2)
)
