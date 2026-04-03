package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	header := StyleTitle.Render("VECTORA AI") + StyleModelBadge.Render("LOCAL SYSTEM")
	
	var footer strings.Builder
	if m.showMenu {
		footer.WriteString(lipgloss.NewStyle().Foreground(Purple).Bold(true).Render("\n COMMANDS: /quit  /models  /tools  /mcp \n"))
	}

	if m.loading {
		footer.WriteString("\n " + m.spinner.View() + " PROCESSANDO...")
	} else {
		footer.WriteString("\n" + StyleInput.Render(m.textInput.View()))
	}

	return header + "\n\n" + m.viewport.View() + footer.String()
}

func (m Model) renderChat() string {
	var sb strings.Builder
	for _, msg := range m.messages {
		if msg.Role == "user" {
			sb.WriteString(StyleUserMsg.Render("YOU: " + msg.Content))
		} else {
			sb.WriteString(StyleBotMsg.Render(msg.Content))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
