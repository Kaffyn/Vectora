package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/Kaffyn/Vectora/internal/engines"
)

func main() {
	// Subcommands
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	activeCmd := flag.NewFlagSet("active", flag.ExitOnError)
	detectCmd := flag.NewFlagSet("detect", flag.ExitOnError)
	setActiveCmd := flag.NewFlagSet("set-active", flag.ExitOnError)

	// Install subcommand flags
	installBuildID := installCmd.String("build", "", "Build ID to install")
	installSilent := installCmd.Bool("silent", false, "Silent mode (no output)")

	// Set-active subcommand flags
	setActiveBuildID := setActiveCmd.String("build", "", "Build ID to set as active")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		if err := handleList(ctx); err != nil {
			log.Fatalf("Error listing builds: %v", err)
		}

	case "install":
		installCmd.Parse(os.Args[2:])
		if *installBuildID == "" {
			fmt.Println("Error: --build flag is required")
			installCmd.Usage()
			os.Exit(1)
		}
		if err := handleInstall(ctx, *installBuildID, !*installSilent); err != nil {
			log.Fatalf("Installation failed: %v", err)
		}

	case "active":
		activeCmd.Parse(os.Args[2:])
		if err := handleGetActive(ctx); err != nil {
			log.Fatalf("Error getting active build: %v", err)
		}

	case "set-active":
		setActiveCmd.Parse(os.Args[2:])
		if *setActiveBuildID == "" {
			fmt.Println("Error: --build flag is required")
			setActiveCmd.Usage()
			os.Exit(1)
		}
		if err := handleSetActive(ctx, *setActiveBuildID); err != nil {
			log.Fatalf("Error setting active build: %v", err)
		}

	case "detect":
		detectCmd.Parse(os.Args[2:])
		if err := handleDetect(ctx); err != nil {
			log.Fatalf("Error detecting hardware: %v", err)
		}

	case "recommend":
		if err := handleRecommend(ctx); err != nil {
			log.Fatalf("Error recommending build: %v", err)
		}

	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	usage := `Vectora Llama.cpp Package Manager (LPM)

Usage:
  lpm <command> [options]

Commands:
  list              List all available builds
  detect            Detect system hardware capabilities
  recommend         Recommend best build for this system
  install           Install a specific build
  active            Show currently active build
  set-active        Set a build as active
  help              Show this help message

Examples:
  lpm list
  lpm detect
  lpm recommend
  lpm install --build llama-windows-x86-cuda-12-q6
  lpm active
  lpm set-active --build llama-windows-x86-cuda-12-q6

Install a recommended build:
  lpm install --build $(lpm recommend)
`
	fmt.Print(usage)
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
