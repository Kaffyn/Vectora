package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "mpm",
	Short:   "Model Package Manager",
	Long:    "MPM (Model Package Manager) - Manage AI models for Vectora",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available models",
	RunE: func(cmd *cobra.Command, args []string) error {
		handleList(args)
		return nil
	},
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect system hardware",
	RunE: func(cmd *cobra.Command, args []string) error {
		handleDetect(args)
		return nil
	},
}

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Recommend a model based on hardware",
	RunE: func(cmd *cobra.Command, args []string) error {
		handleRecommend(args)
		return nil
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search models by name/capability",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		handleSearch(args)
		return nil
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a model",
	RunE: func(cmd *cobra.Command, args []string) error {
		handleInstall(args)
		return nil
	},
}

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Show currently active model",
	RunE: func(cmd *cobra.Command, args []string) error {
		handleActive(args)
		return nil
	},
}

var setActiveCmd = &cobra.Command{
	Use:   "set-active",
	Short: "Set the active model",
	RunE: func(cmd *cobra.Command, args []string) error {
		handleSetActive(args)
		return nil
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(recommendCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(activeCmd)
	rootCmd.AddCommand(setActiveCmd)

	// Flags for list command
	listCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	listCmd.Flags().StringP("family", "f", "", "Filter by model family")

	// Flags for detect command
	detectCmd.Flags().BoolP("json", "j", false, "Output as JSON")

	// Flags for recommend command
	recommendCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	recommendCmd.Flags().StringP("size", "s", "", "Filter by size (0.6b, 1.7b, 4b, 8b)")

	// Flags for search command
	searchCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	searchCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	searchCmd.Flags().StringP("capability", "c", "", "Filter by capability")

	// Flags for install command
	installCmd.Flags().StringP("model", "m", "", "Model ID to install")

	// Flags for active command
	activeCmd.Flags().BoolP("json", "j", false, "Output as JSON")

	// Flags for set-active command
	setActiveCmd.Flags().StringP("model", "m", "", "Model ID to set as active")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
