package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kaffyn/Vectora/internal/core"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type queryResponseMsg struct {
	Answer string
	Err    error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		sCmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			input := m.textInput.Value()
			if strings.TrimSpace(input) == "" {
				return m, nil
			}

			// '/' Commands Handling
			if strings.HasPrefix(input, "/") {
				return m.handleSlashCommand(input)
			}

			if m.loading {
				return m, nil
			}

			m.messages = append(m.messages, ChatMessage{Role: "user", Content: input})
			m.textInput.Reset()
			m.loading = true
			m.viewport.SetContent(m.renderChat())
			m.viewport.GotoBottom()

			return m, m.sendQuery(input)
		}

	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		headerHeight := 4
		footerHeight := 4
		m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
		m.viewport.SetContent(m.renderChat())
		m.textInput.Width = msg.Width - 10

	case queryResponseMsg:
		m.loading = false
		if msg.Err != nil {
			m.messages = append(m.messages, ChatMessage{Role: "bot", Content: fmt.Sprintf("❌ Error: %v", msg.Err)})
		} else {
			m.messages = append(m.messages, ChatMessage{Role: "bot", Content: msg.Answer})
		}
		m.viewport.SetContent(m.renderChat())
		m.viewport.GotoBottom()

	case spinner.TickMsg:
		m.spinner, sCmd = m.spinner.Update(msg)
		return m, sCmd
	}

	// Smart menu when typing /
	m.showMenu = m.textInput.Value() == "/"

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) sendQuery(query string) tea.Cmd {
	return func() tea.Msg {
		var res core.QueryResponse
		err := m.ipcClient.Send(context.Background(), "workspace.query", core.QueryRequest{
			WorkspaceID: "default",
			Query:       query,
		}, &res)

		return queryResponseMsg{Answer: res.Answer, Err: err}
	}
}
