package cli

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleSlashCommand(input string) (tea.Model, tea.Cmd) {
	cmd := strings.TrimPrefix(input, "/")
	m.textInput.Reset()

	switch cmd {
	case "quit", "exit", "q":
		return m, tea.Quit
		
	case "models":
		m.messages = append(m.messages, ChatMessage{
			Role: "bot", 
			Content: "🧩 AVAILABLE PROVIDERS:\n\n" +
				"1. [LOCAL] Qwen 0.5B (GGUF via Llama-cpp)\n" +
				"2. [CLOUD] Gemini 1.5 Flash (Default)\n" +
				"3. [CLOUD] Gemini 1.5 Pro (Precision Mode)",
		})
		
	case "tools":
		m.messages = append(m.messages, ChatMessage{
			Role: "bot", 
			Content: "🛠️ AGENT TOOLS:\n\n" +
				"- search_web: Brave/Google Search integration\n" +
				"- memory_rag: Dynamic semantic search\n" +
				"- code_writer: Local file manipulation\n" +
				"- git_bridge: Workspace version control",
		})
		
	case "mcp":
		m.messages = append(m.messages, ChatMessage{
			Role: "bot", 
			Content: "📦 MCP PROTOCOL (Model Context Protocol):\n\n" +
				"Vectora operating as [CLIENT].\n" +
				"Connections detected in the ecosystem:\n" +
				"✅ github-server-v1\n" +
				"✅ google-search-mcp\n" +
				"⚠️  filesystem-mcp (Permission denied)",
		})
		
	default:
		m.messages = append(m.messages, ChatMessage{
			Role: "bot", 
			Content: "❓ INVALID COMMAND.\nUse /models, /tools, /mcp or /quit.",
		})
	}
	
	m.viewport.SetContent(m.renderChat())
	m.viewport.GotoBottom()
	return m, nil
}
