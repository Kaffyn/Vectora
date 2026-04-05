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
	Gold      = lipgloss.Color("#FFB800")
	Green     = lipgloss.Color("#00D64F")
	Red       = lipgloss.Color("#FF4565")

	// Main Title Style
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Blue).
			FontSize(16)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(LightGray).
			FontSize(12)

	// Model/Status Badge
	StyleModelBadge = lipgloss.NewStyle().
			Background(DeepBlue).
			Foreground(Accent).
			Padding(0, 1).
			Bold(true).
			MarginLeft(1)

	StylePlanBadge = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	// Message Styles
	StyleBotMsg = lipgloss.NewStyle().
			Foreground(LightGray).
			Border(lipgloss.Border{Left: "▌"}).
			BorderForeground(Purple).
			PaddingLeft(2).
			PaddingRight(1).
			MarginBottom(1)

	StyleUserMsg = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true).
			MarginTop(1).
			MarginBottom(0)

	StyleSystemMsg = lipgloss.NewStyle().
			Foreground(Gold).
			Italic(true).
			MarginTop(1).
			MarginBottom(1)

	// Input Style
	StyleInput = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DeepBlue).
			Padding(0, 1).
			MarginTop(1)

	StylePrompt = lipgloss.NewStyle().
			Foreground(Blue).
			Bold(true)

	// Info/Help Style
	StyleHelp = lipgloss.NewStyle().
			Foreground(LightGray).
			Italic(true)

	StyleLoading = lipgloss.NewStyle().
			Foreground(Accent)

	// Header divider
	StyleDivider = lipgloss.NewStyle().
			Foreground(DeepBlue)
)
