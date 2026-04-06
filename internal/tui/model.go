package tui

import (
	"github.com/Kaffyn/Vectora/internal/ipc"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ChatMessage struct {
	Role    string
	Content string
}

type Model struct {
	viewport       viewport.Model
	textInput      textinput.Model
	spinner        spinner.Model
	ipcClient      *ipc.Client
	messages       []ChatMessage
	err            error
	loading        bool
	showMenu       bool
	terminalWidth  int
	terminalHeight int
}

func NewModel(client *ipc.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "Say something or type '/' for commands..."
	ti.Focus()
	ti.Prompt = " >> "

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(Accent)

	return Model{
		textInput: ti,
		spinner:   s,
		ipcClient: client,
		messages:  []ChatMessage{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}
