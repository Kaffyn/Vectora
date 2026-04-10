package main

import (
	"fmt"
	"os"
	"path/filepath"

	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage local indexing namespaces",
	Long:  "View and manage Vectora memory collections and vector shards.",
}

var workspaceLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all indexed workspaces",
	Run: func(cmd *cobra.Command, args []string) {
		systemManager, _ := vecos.NewManager()
		appDataDir, _ := systemManager.GetAppDataDir()
		chromaDir := filepath.Join(appDataDir, "data", "chroma")

		entries, err := os.ReadDir(chromaDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No workspaces found.")
				return
			}
			fmt.Println("Error reading workspaces:", err)
			return
		}

		fmt.Println("Indexed Workspaces:")
		count := 0
		for _, e := range entries {
			if e.IsDir() {
				info, _ := e.Info()
				sizeStr := ""
				if info != nil {
					sizeStr = fmt.Sprintf("\t(Last modified: %s)", info.ModTime().Format("2006-01-02 15:04:05"))
				}
				fmt.Printf(" - %s%s\n", e.Name(), sizeStr)
				count++
			}
		}

		if count == 0 {
			fmt.Println("No workspaces found.")
		}
	},
}

var workspaceHard bool
var workspaceRmCmd = &cobra.Command{
	Use:   "rm [workspace_id]",
	Short: "Delete a workspace completely",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		if !workspaceHard {
			fmt.Println("Error: This removes all RAG context for this workspace permanently.")
			fmt.Println("Please run with --hard to confirm.")
			os.Exit(1)
		}

		systemManager, _ := vecos.NewManager()
		appDataDir, _ := systemManager.GetAppDataDir()
		wsChromaDir := filepath.Join(appDataDir, "data", "chroma", id)

		fmt.Printf("Deleting workspace '%s'...\n", id)
		err := os.RemoveAll(wsChromaDir)
		if err != nil {
			fmt.Println("Error removing workspace metrics:", err)
			return
		}

		fmt.Println("Workspace deleted successfully.")
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceLsCmd)
	workspaceCmd.AddCommand(workspaceRmCmd)
	workspaceRmCmd.Flags().BoolVar(&workspaceHard, "hard", false, "Confirm irreversible deletion")
}
