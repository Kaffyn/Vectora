package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/ipc"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/Kaffyn/Vectora/core/tray"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var daemonPort int

var rootCmd = &cobra.Command{
	Use:     "vectora",
	Short:   "Vectora AI - Local Engineering Assistant",
	Long:    "Vectora is an offline-first local AI assistant with knowledge base management.",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
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

func init() {
	skipElevation := false
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if cmd == "status" || cmd == "stop" || cmd == "help" || cmd == "-h" || cmd == "--help" || cmd == "--version" || cmd == "-v" {
			skipElevation = true
		}
	}

	if !skipElevation {
		systemManager, _ := vecos.NewManager()
		if systemManager != nil && !systemManager.IsRunningAsAdmin() {
			if err := elevateAdmin(); err == nil {
				os.Exit(0)
			}
		}
	}

	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stopCmd)

	daemonCmd.Flags().IntVar(&daemonPort, "port", 42780, "Custom daemon port")
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

	appDataDir, _ := systemManager.GetAppDataDir()
	envPath := filepath.Join(appDataDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		infra.Logger.Warn(fmt.Sprintf("Global config (.env) not found at: %s. Using defaults.", envPath))
	}

	infra.Logger.Info("Starting Vectora Daemon...")

	kvStore, _ := db.NewKVStore()
	vecStore, _ := db.NewVectorStore()
	appCtx := context.Background()
	msgService := llm.NewMessageService(kvStore)
	memService, _ := db.NewMemoryService(appCtx, filepath.Join(appDataDir, "data", "memory"))

	ipcServer, _ := ipc.NewServer()

	// Provider fetcher for IPC routes
	getProvider := func() llm.Provider {
		return tray.ActiveProvider
	}

	ipc.RegisterRoutes(ipcServer, kvStore, vecStore, getProvider, msgService, memService)
	go ipcServer.Start()

	// Dev HTTP bridge for debugging
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
}

func runStop() {
	if os.Getenv("OS") == "Windows_NT" {
		exec.Command("taskkill", "/F", "/IM", "vectora.exe").Run()
		fmt.Println("Vectora terminated.")
	}
}
