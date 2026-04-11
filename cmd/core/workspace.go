package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/spf13/cobra"
)

// loadWorkspaceRegistry reads the workspace registry (workspaces.json) from disk.
// This file is written by the engine during embed jobs.
func loadWorkspaceRegistry() map[string]string {
	userProfile, _ := os.UserHomeDir()
	regPath := filepath.Join(userProfile, ".Vectora", "workspaces.json")
	data, err := os.ReadFile(regPath)
	if err != nil {
		return make(map[string]string)
	}
	var reg map[string]string
	if json.Unmarshal(data, &reg) != nil {
		return make(map[string]string)
	}
	return reg
}

var workspaceCmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"workspaces", "ws"},
	Short:   "Manage local indexing namespaces",
	Long:    "View and manage Vectora memory collections and vector shards.",
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

		registry := loadWorkspaceRegistry()

		fmt.Println("Indexed Workspaces:")
		count := 0
		for _, e := range entries {
			if e.IsDir() {
				info, _ := e.Info()
				modTime := ""
				if info != nil {
					modTime = info.ModTime().Format("2006-01-02 15:04:05")
				}

				sourcePath := registry[e.Name()]
				if sourcePath != "" {
					fmt.Printf(" - %s → %s\t(%s)\n", e.Name(), sourcePath, modTime)
				} else {
					fmt.Printf(" - %s\t(%s)\n", e.Name(), modTime)
				}
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
