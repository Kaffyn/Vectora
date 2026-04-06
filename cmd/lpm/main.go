package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Kaffyn/Vectora/engines"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var (
	ctx = context.Background()
)

var rootCmd = &cobra.Command{
	Use:     "lpm",
	Short:   "Llama.cpp Package Manager",
	Long:    "LPM (Llama Package Manager) - Manage llama.cpp builds for your system",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available builds",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleList(ctx)
	},
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect system hardware capabilities",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleDetect(ctx)
	},
}

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Recommend best build for this system",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleRecommend(ctx)
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a specific build",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		buildID, _ := cmd.Flags().GetString("build")
		if buildID == "" {
			return fmt.Errorf("--build flag is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		buildID, _ := cmd.Flags().GetString("build")
		silent, _ := cmd.Flags().GetBool("silent")
		return handleInstall(ctx, buildID, !silent)
	},
}

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Show currently active build",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleGetActive(ctx)
	},
}

var setActiveCmd = &cobra.Command{
	Use:   "set-active",
	Short: "Set a build as active",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		buildID, _ := cmd.Flags().GetString("build")
		if buildID == "" {
			return fmt.Errorf("--build flag is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		buildID, _ := cmd.Flags().GetString("build")
		return handleSetActive(ctx, buildID)
	},
}

func init() {
	// Add subcommands to root
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(recommendCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(activeCmd)
	rootCmd.AddCommand(setActiveCmd)

	// Add flags to subcommands
	installCmd.Flags().StringP("build", "b", "", "Build ID to install")
	installCmd.Flags().BoolP("silent", "s", false, "Silent mode (no output)")

	setActiveCmd.Flags().StringP("build", "b", "", "Build ID to set as active")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleList(ctx context.Context) error {
	if err := engines.LoadCatalog(); err != nil {
		return fmt.Errorf("failed to load catalog: %w", err)
	}

	catalog, err := engines.GetCatalog()
	if err != nil {
		return fmt.Errorf("failed to get catalog: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tOS\tArch\tGPU\tDescription")
	fmt.Fprintln(w, "---\t---\t----\t---\t-----------")

	for _, build := range catalog.Builds {
		gpu := build.GPU
		if gpu == "" {
			gpu = "CPU"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			build.ID,
			build.OS,
			build.Architecture,
			gpu,
			build.Description,
		)
	}
	w.Flush()

	return nil
}

func handleDetect(ctx context.Context) error {
	hw, err := engines.DetectHardware()
	if err != nil {
		return fmt.Errorf("hardware detection failed: %w", err)
	}

	fmt.Printf("System Hardware:\n")
	fmt.Printf("  OS:           %s\n", hw.OS)
	fmt.Printf("  Architecture: %s\n", hw.Architecture)
	fmt.Printf("  CPU Cores:    %d\n", hw.CoreCount)
	fmt.Printf("  RAM:          %.2f GB\n", float64(hw.RAM)/(1024*1024*1024))
	fmt.Printf("  GPU Type:     %s\n", hw.GPUType)
	if hw.GPUVersion != "" {
		fmt.Printf("  GPU Version:  %s\n", hw.GPUVersion)
	}
	if len(hw.CPUFeatures) > 0 {
		fmt.Printf("  CPU Features: %v\n", hw.CPUFeatures)
	}

	return nil
}

func handleRecommend(ctx context.Context) error {
	if err := engines.LoadCatalog(); err != nil {
		return fmt.Errorf("failed to load catalog: %w", err)
	}

	hw, err := engines.DetectHardware()
	if err != nil {
		return fmt.Errorf("hardware detection failed: %w", err)
	}

	build, err := engines.RecommendedBuild(hw)
	if err != nil {
		return fmt.Errorf("no suitable build found: %w", err)
	}

	// Print just the ID for script usage
	fmt.Println(build.ID)

	return nil
}

func handleInstall(ctx context.Context, buildID string, verbose bool) error {
	mgr, err := engines.NewEngineManager()
	if err != nil {
		return fmt.Errorf("failed to create engine manager: %w", err)
	}

	if verbose {
		fmt.Printf("Installing build: %s\n", buildID)
	}

	onProgress := func(p *engines.DownloadProgress) error {
		if verbose && p.Total > 0 {
			percent := float64(p.Current) / float64(p.Total) * 100
			fmt.Printf("\rDownloading: %.1f%% (%d/%d bytes)", percent, p.Current, p.Total)
		}
		return nil
	}

	if err := mgr.Install(ctx, buildID, onProgress); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("\n✅ Build %s installed successfully\n", buildID)
	}

	return nil
}

func handleGetActive(ctx context.Context) error {
	mgr, err := engines.NewEngineManager()
	if err != nil {
		return fmt.Errorf("failed to create engine manager: %w", err)
	}

	info, err := mgr.GetActive()
	if err != nil {
		return fmt.Errorf("no active build: %w", err)
	}

	fmt.Printf("Active Build: %s\n", info.BuildID)
	fmt.Printf("Path:         %s\n", info.Path)
	fmt.Printf("Installed:    %s\n", info.Installed.Format("2006-01-02 15:04:05"))

	return nil
}

func handleSetActive(ctx context.Context, buildID string) error {
	mgr, err := engines.NewEngineManager()
	if err != nil {
		return fmt.Errorf("failed to create engine manager: %w", err)
	}

	if err := mgr.SetActive(buildID); err != nil {
		return err
	}

	fmt.Printf("✅ Build %s is now active\n", buildID)

	return nil
}
