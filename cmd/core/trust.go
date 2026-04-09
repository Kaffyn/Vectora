package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Manage trusted workspaces for Vectora ACP",
	Long:  "Whitelist or remove directories allowed to connect to the background core via ACP.",
}

var trustAddCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Add a directory to the trust list",
	Long:  "Allow IDEs operating in this directory to connect to the Vectora Core.",
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Println("Error resolving path:", err)
			return
		}

		trustList, err := readTrustList()
		if err != nil {
			fmt.Println("Error reading trust list:", err)
			return
		}

		for _, p := range trustList {
			if strings.EqualFold(p, absPath) {
				fmt.Println("Path is already trusted:", absPath)
				return
			}
		}

		trustList = append(trustList, absPath)
		if err := writeTrustList(trustList); err != nil {
			fmt.Println("Error writing trust list:", err)
			return
		}

		fmt.Println("Path added to trust list:", absPath)
	},
}

var trustRmCmd = &cobra.Command{
	Use:   "rm [path]",
	Short: "Remove a directory from the trust list",
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Println("Error resolving path:", err)
			return
		}

		trustList, err := readTrustList()
		if err != nil {
			fmt.Println("Error reading trust list:", err)
			return
		}

		var newList []string
		found := false
		for _, p := range trustList {
			if strings.EqualFold(p, absPath) {
				found = true
				continue
			}
			newList = append(newList, p)
		}

		if !found {
			fmt.Println("Path not found in trust list:", absPath)
			return
		}

		if err := writeTrustList(newList); err != nil {
			fmt.Println("Error writing trust list:", err)
			return
		}

		fmt.Println("Path removed from trust list:", absPath)
	},
}

var trustLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all trusted workspaces",
	Run: func(cmd *cobra.Command, args []string) {
		trustList, err := readTrustList()
		if err != nil {
			fmt.Println("Error reading trust list:", err)
			return
		}

		if len(trustList) == 0 {
			fmt.Println("No trusted directories found.")
			return
		}

		fmt.Println("Trusted workspaces:")
		for _, p := range trustList {
			fmt.Println(" -", p)
		}
	},
}

func init() {
	trustCmd.AddCommand(trustAddCmd)
	trustCmd.AddCommand(trustRmCmd)
	trustCmd.AddCommand(trustLsCmd)
}

func getTrustFilePath() string {
	userProfile, _ := os.UserHomeDir()
	return filepath.Join(userProfile, ".Vectora", "trust_paths.txt")
}

func readTrustList() ([]string, error) {
	path := getTrustFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var validLines []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" {
			validLines = append(validLines, t)
		}
	}
	return validLines, nil
}

func writeTrustList(list []string) error {
	path := getTrustFilePath()
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, []byte(strings.Join(list, "\n")), 0644)
}
