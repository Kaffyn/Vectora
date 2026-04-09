package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Vectora core configurations",
	Long:  "Read or modify global configuration keys stored in ~/.Vectora/.env",
}

var configSetCmd = &cobra.Command{
	Use:   "set [KEY] [VALUE]",
	Short: "Set a configuration key",
	Long:  "Example: vectora config set GEMINI_API_KEY AIzaSy...",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		envPath := getConfigPath()
		envMap, err := godotenv.Read(envPath)
		if err != nil {
			if os.IsNotExist(err) {
				envMap = make(map[string]string)
			} else {
				fmt.Println("Error reading .env:", err)
				return
			}
		}

		envMap[key] = value
		if err := godotenv.Write(envMap, envPath); err != nil {
			fmt.Println("Error writing .env:", err)
			return
		}

		fmt.Printf("Key %s has been set successfully.\n", key)
		fmt.Println("Restarting Core to apply changes...")
		runStop()
		spawnDetached()
		fmt.Println("Core restarted.")
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration keys",
	Run: func(cmd *cobra.Command, args []string) {
		envPath := getConfigPath()
		envMap, err := godotenv.Read(envPath)
		if err != nil {
			fmt.Println("No configuration found or error reading it.")
			return
		}

		for k, v := range envMap {
			displayVal := v
			if len(v) > 10 {
				displayVal = v[:4] + "..." + v[len(v)-4:]
			}
			fmt.Printf("%s=%s\n", k, displayVal)
		}
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

func getConfigPath() string {
	userProfile, _ := os.UserHomeDir()
	return filepath.Join(userProfile, ".Vectora", ".env")
}
