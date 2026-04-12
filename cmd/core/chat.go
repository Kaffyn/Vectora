package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kaffyn/Vectora/core/api/ipc"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive TUI chat with Vectora",
	Long:  "Launch a continuous REPL session that maintains conversation history natively within the core.",
	Run: func(cmd *cobra.Command, args []string) {
		var client *ipc.Client
		var err error
		client, err = ensureCoreConnected()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Please try running: vectora start")
			return
		}
		defer client.Close()

		// Initialize workspace (Required for activeTenant check in Phase 13)
		cwd, _ := os.Getwd()
		if _, err := initWorkspace(client, cwd); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Derive stable session from cwd (same as `vectora ask`)
		absCwd, _ := filepath.Abs(cwd)
		conversationID := workspaceConversationID(absCwd)

		ctx := context.Background()

		// Load existing history
		prior, _ := loadConversationEntries(conversationID)

		fmt.Println("=====================================================")
		fmt.Printf(" Vectora Chat  [session: %s]\n", conversationID)
		fmt.Printf(" Workspace: %s\n", absCwd)
		if len(prior) > 0 {
			fmt.Printf(" Resuming conversation (%d messages)\n", len(prior))
		} else {
			fmt.Println(" New conversation")
		}
		fmt.Println(" /exit to quit   /clear to reset   /history to view")
		fmt.Println("=====================================================")

		scanner := bufio.NewScanner(os.Stdin)

		for {
			fmt.Print("\n> ")
			if !scanner.Scan() {
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			switch input {
			case "/exit", "exit", "quit":
				fmt.Println("Goodbye!")
				return

			case "/clear":
				clearConversation(conversationID)
				conversationID = workspaceConversationID(absCwd) // same ID, file deleted
				fmt.Println("Conversation history cleared.")
				continue

			case "/history":
				entries, _ := loadConversationEntries(conversationID)
				if len(entries) == 0 {
					fmt.Println("No history yet.")
				}
				for _, e := range entries {
					fmt.Printf("[%s] %s: %s\n", e.Timestamp.Format("15:04"), e.Role, e.Content[:min(80, len(e.Content))])
				}
				continue
			}

			fmt.Print("...thinking\r")
			appendConversationEntry(conversationID, "user", input)

			req := map[string]string{
				"workspace_id":    conversationID,
				"query":           input,
				"conversation_id": conversationID,
			}

			var resp struct {
				Answer string `json:"answer"`
			}

			start := time.Now()
			err = client.Send(ctx, "workspace.query", req, &resp)
			duration := time.Since(start)

			fmt.Printf("%-20s\r", " ") // clear thinking

			if err != nil {
				if strings.Contains(err.Error(), "No LLM provider has been configured") || strings.Contains(err.Error(), "provider_not_configured") {
					fmt.Println("Error: Vectora requires an API key to work.")
					runConfigInteractive()
					continue
				}
				fmt.Println("Error:", err)
				continue
			}

			appendConversationEntry(conversationID, "assistant", resp.Answer)
			fmt.Printf("Vectora: %s\n\n[%v]\n", resp.Answer, duration)
		}
	},
}
