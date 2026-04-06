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
)

func main() {
	systemManager, _ := vecos.NewManager()
	if systemManager != nil && !systemManager.IsRunningAsAdmin() {
		// Attempt to restart as admin
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

	if len(os.Args) < 2 {
		// Default to daemon mode if double-clicked or run without args
		runDaemon()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "daemon", "start":
		runDaemon()
	case "status":
		runStatus()
	case "stop":
		runStop()
	case "--tests":
		runSystemIntegrityTests()
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Vectora AI - Local Engineering Assistant")
	fmt.Println("\nUsage:")
	fmt.Println("  vectora <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  daemon       Start background service (Tray)")
	fmt.Println("  start        Start background service (Tray)")
	fmt.Println("  stop         Shutdown background service")
	fmt.Println("  status       Verify health of micro-services")
	fmt.Println("  --tests      Run integrity suite")
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
	go ipcServer.StartDevHTTP(42700)

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
	// ... Reuse previous code ...
}
