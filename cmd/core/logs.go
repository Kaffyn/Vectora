package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var clearLogs bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View Vectora core telemetry logs",
	Long:  "Tail the background core logs in real-time or manage them.",
	Run: func(cmd *cobra.Command, args []string) {
		userProfile, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		logPath := filepath.Join(userProfile, ".Vectora", "logs", "app.log")

		// Handle clear flag
		if clearLogs {
			err := os.WriteFile(logPath, []byte(""), 0644)
			if err != nil {
				fmt.Println("Failed to clear logs:", err)
			} else {
				fmt.Println("Logs cleared successfully.")
			}
			return
		}

		// Check if file exists
		file, err := os.Open(logPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No logs found at", logPath)
				return
			}
			fmt.Println("Failed to access logs:", err)
			return
		}
		defer file.Close()

		// Go to end of file minus 2000 bytes (to print last few lines before tailing)
		stat, _ := file.Stat()
		if stat.Size() > 2000 {
			file.Seek(-2000, 2) // Seek near the end
		}

		fmt.Printf("--- Tailing logs at %s ---\n", logPath)
		fmt.Println("Press Ctrl+C to exit.")
		fmt.Println("========================================")

		// Basic tail -f logic
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				fmt.Print(string(buf[:n]))
			}
			if err == io.EOF {
				time.Sleep(500 * time.Millisecond) // Wait for more content
			} else if err != nil {
				fmt.Println("\nError reading file stream:", err)
				break
			}
		}
	},
}

func init() {
	logsCmd.Flags().BoolVar(&clearLogs, "clear", false, "Erase all existing log data")
}
