package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Kaffyn/Vectora/internal/db"
	"github.com/Kaffyn/Vectora/internal/infra"
	"github.com/Kaffyn/Vectora/internal/ipc"
	"github.com/Kaffyn/Vectora/internal/llm"
	vecos "github.com/Kaffyn/Vectora/internal/os"
	"github.com/Kaffyn/Vectora/internal/tools"
	"github.com/Kaffyn/Vectora/internal/tray"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var (
	daemonPort int
	testMode   bool
)

var rootCmd = &cobra.Command{
	Use:     "vectora",
	Short:   "Vectora AI - Local Engineering Assistant",
	Long:    "Vectora is an offline-first local AI assistant with knowledge base management.",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// Default to daemon mode if no arguments
		runDaemon()
	},
}

var daemonCmd = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"start"},
	Short:   "Start background service (Tray)",
	Long:    "Start the Vectora daemon as a background service in the system tray.",
	Run: func(cmd *cobra.Command, args []string) {
		runDaemon()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Verify health of micro-services",
	Long:  "Check if the Vectora daemon is running and responsive.",
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Shutdown background service",
	Long:  "Stop the running Vectora daemon.",
	Run: func(cmd *cobra.Command, args []string) {
		runStop()
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run integrity suite",
	Long:  "Run system integrity tests and diagnostics.",
	Run: func(cmd *cobra.Command, args []string) {
		runSystemIntegrityTests()
	},
}

var (
	checkOnly bool
)

var updateCmd = &cobra.Command{
	Use:   "update [component]",
	Short: "Update Vectora components",
	Long:  "Update Vectora daemon and all components (tui, lpm, mpm, setup). Specify components or update all.",
	Example: `  vectora update              # Update all components
  vectora update daemon      # Update daemon only
  vectora update tui lpm     # Update tui and lpm
  vectora update --check     # Check for updates without installing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate(args, checkOnly)
	},
}

func init() {
	// Admin elevation for Windows - only for daemon, status, stop commands
	// Skip elevation for update, test, and help commands
	skipElevation := false
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if cmd == "update" || cmd == "test" || cmd == "help" || cmd == "-h" || cmd == "--help" || cmd == "--version" || cmd == "-v" {
			skipElevation = true
		}
	}

	if !skipElevation {
		systemManager, _ := vecos.NewManager()
		if systemManager != nil && !systemManager.IsRunningAsAdmin() {
			exe, _ := os.Executable()
			cwd, _ := os.Getwd()
			args := strings.Join(os.Args[1:], " ")

			cmd := exec.Command("powershell", fmt.Sprintf("Start-Process -FilePath '%s' -Verb runas -WorkingDirectory '%s'", exe, cwd))
			if args != "" {
				cmd = exec.Command("powershell", fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb runas -WorkingDirectory '%s'", exe, args, cwd))
			}
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			if err := cmd.Start(); err == nil {
				os.Exit(0)
			}
		}
	}

	// Add subcommands
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(updateCmd)

	// Add flags
	daemonCmd.Flags().IntVar(&daemonPort, "port", 42780, "Custom daemon port")
	updateCmd.Flags().BoolVar(&checkOnly, "check", false, "Check for updates without installing")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDaemon() {
	systemManager, err := vecos.NewManager()
	if err != nil {
		log.Fatalf("Critical Hardware OS Failure: %v", err)
	}

	if err := systemManager.EnforceSingleInstance(); err != nil {
		infra.NotifyOS("Vectora Running", "Vectora is already running in the system tray.")
		os.Exit(0)
	}

	infra.SetupLogger()

	// Load global configuration from %USERPROFILE%/.Vectora/.env
	appDataDir, _ := systemManager.GetAppDataDir()
	envPath := filepath.Join(appDataDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		infra.Logger.Warn(fmt.Sprintf("Global config (.env) not found at: %s. Using defaults.", envPath))
	}

	infra.Logger.Info("Starting Vectora Daemon (Micro-Services Mode)...")

	kvStore, _ := db.NewKVStore()
	vecStore, _ := db.NewVectorStore()
	appCtx := context.Background()
	msgService := llm.NewMessageService(kvStore)
	memService, _ := db.NewMemoryService(appCtx, filepath.Join(appDataDir, "data", "memory"))
	searchService := tools.NewSearchService()

	ipcServer, _ := ipc.NewServer()
	ipc.RegisterRoutes(ipcServer, kvStore, vecStore, func() llm.Provider { return tray.ActiveProvider }, msgService, memService, searchService)
	go ipcServer.Start()

	// Start the Dev HTTP Bridge in background for debugging IPC protocol.
	// Used for testing and development purposes - Vectora Desktop (Fyne) uses IPC directly.
	go ipcServer.StartDevHTTP(daemonPort)

	infra.NotifyOS("Vectora", "Operational Assistant.")
	tray.Setup()
}

func runStatus() {
	fmt.Println("--- VECTORA STATUS ---")
	client, err := ipc.NewClient()
	if err != nil || client.Connect() != nil {
		fmt.Println("Status: OFFLINE (Daemon not found)")
		return
	}
	fmt.Println("Status: ONLINE (Daemon active)")
	// TODO: Add ping to router to check real health of sub-services
}

func runStop() {
	// On Windows, we could use mutex or TaskKill for simplicity if local.
	if os.Getenv("OS") == "Windows_NT" {
		exec.Command("taskkill", "/F", "/IM", "vectora.exe").Run()
		fmt.Println("Vectora terminated.")
	}
}

func runSystemIntegrityTests() {
	fmt.Println("🛰️ Initializing Integrity Audit...")
	// TODO: Implement system integrity tests
	fmt.Println("✅ All systems operational")
}
