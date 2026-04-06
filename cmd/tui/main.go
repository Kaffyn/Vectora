package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Kaffyn/Vectora/internal/ipc"
	"github.com/Kaffyn/Vectora/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	client, err := ipc.NewClient()
	if err != nil {
		fmt.Printf("Error starting client: %v\n", err)
		return
	}

	// Attempt to connect. If fails, start the daemon silently.
	if err := client.Connect(); err != nil {
		fmt.Println(">> Backend offline. Starting Daemon in background...")
		self, _ := os.Executable()
		cmd := exec.Command(self, "daemon")
		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to start daemon: %v\n", err)
			return
		}

		// Wait for the socket to be ready
		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			time.Sleep(500 * time.Millisecond)
			if err := client.Connect(); err == nil {
				goto connected
			}
		}
		fmt.Println("❌ Failed to connect to Backend after boot.")
		return
	}

connected:
	p := tea.NewProgram(tui.NewModel(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("UI Error: %v\n", err)
	}
}
