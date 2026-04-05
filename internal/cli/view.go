package cli

import (
	"strings"
)

func (m Model) View() string {
	// Header with title and model info
	header := m.renderHeader()

	// Main content area
	content := m.renderContent()

	// Input area
	inputArea := m.renderInput()

	// Combine all
	return header + "\n" + content + "\n" + inputArea
}

func (m Model) renderHeader() string {
	title := StyleTitle.Render("VECTORA AI")
	model := StyleModelBadge.Render("LOCAL_SYSTEM")

	headerLine := title + " " + model
	divider := StyleDivider.Render(strings.Repeat("─", 70))

	return headerLine + "\n" + divider
}

func (m Model) renderContent() string {
	if len(m.messages) == 0 {
		welcomeMsg := StyleSystemMsg.Render("Welcome to Vectora AI!")
		help := StyleHelp.Render("Type your message or '/?' for commands")
		return welcomeMsg + "\n" + help
	}

	var sb strings.Builder
	for _, msg := range m.messages {
		if msg.Role == "user" {
			// User message with "YOU:" prefix
			userLabel := StyleUserMsg.Render("YOU:")
			userContent := StyleUserMsg.Render(msg.Content)
			sb.WriteString(userLabel + " " + userContent + "\n\n")
		} else if msg.Role == "assistant" {
			// Bot message with border
			botContent := StyleBotMsg.Render(msg.Content)
			sb.WriteString(botContent + "\n")
		}
	}
	return sb.String()
}

func (m Model) renderInput() string {
	if m.loading {
		loader := StyleLoading.Render(m.spinner.View())
		status := StyleLoading.Render(" Processing...")
		return loader + status
	}

	inputPrompt := StylePrompt.Render(">>> ")
	inputField := m.textInput.View()

	inputBox := StyleInput.Render(inputPrompt + inputField)

	// Add help text
	var help string
	if m.showMenu {
		help = StyleHelp.Render("\n📋 Commands: /help  /model  /reset  /quit")
	} else {
		help = StyleHelp.Render("\n💡 Type /? for help")
	}

	return inputBox + help
}

func (m Model) renderChat() string {
	// Deprecated - use renderContent instead
	return m.renderContent()
}
