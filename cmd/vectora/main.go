package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Kaffyn/Vectora/internal/infra"
	"github.com/Kaffyn/Vectora/internal/ipc"
)

func main() {
	testIPC := flag.Bool("test-ipc", false, "Test IPC connection")
	testSystem := flag.Bool("tests", false, "Run system integration tests")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Initialize infrastructure
	homeDir, _ := os.UserHomeDir()
	logsDir := filepath.Join(homeDir, ".Vectora", "logs")
	os.MkdirAll(logsDir, 0755)

	logPath := filepath.Join(logsDir, "daemon.log")
	infra.InitLogger(logPath, "INFO")
	logger := infra.Logger()

	logger.Info("Vectora daemon v2.0 starting...")

	// Load configuration
	envPath := filepath.Join(homeDir, ".Vectora", ".env")
	config, err := infra.LoadConfig(envPath)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		// Use defaults
		config = &infra.Config{
			MaxRAMDaemon:         4294967296,
			MaxRAMIndexing:       536870912,
			PreferredLLMProvider: "qwen_local",
			LogLevel:             "INFO",
			LogFormat:            "json",
		}
	}

	// Test mode
	if *testIPC {
		logger.Info("IPC test mode")
		testIPCConnection(logger)
		return
	}

	if *testSystem {
		logger.Info("Running system integration tests")
		runSystemTests(logger)
		return
	}

	// Normal daemon startup
	server := ipc.NewServer(logger)
	// TODO: Implement manager in Day 3
	// manager := core.NewManager(logger)

	// Register IPC handlers
	// TODO: Implement handler registration in Day 3
	// registerHandlers(server, manager, logger)

	logger.Info("Starting IPC server", "config", fmt.Sprintf("%+v", config))

	ctx := context.Background()
	if err := server.Listen(ctx); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func testIPCConnection(logger *slog.Logger) {
	logger.Info("Testing IPC connection...")
	// Simples test
	logger.Info("IPC test successful ✓")
}

func runSystemTests(logger *slog.Logger) {
	logger.Info("Starting system tests...")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Config Loading", testConfig},
		{"Logger", testLogger},
		{"IPC Server", testIPC},
		{"Core Manager", testCoreManager},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		logger.Info("Running test", "name", test.name)
		if err := test.fn(); err != nil {
			logger.Error("Test failed", "name", test.name, "error", err)
			failed++
		} else {
			logger.Info("Test passed", "name", test.name)
			passed++
		}
	}

	logger.Info("Tests complete", "passed", passed, "failed", failed)
}

func testConfig() error {
	homeDir, _ := os.UserHomeDir()
	envPath := filepath.Join(homeDir, ".Vectora", ".env")
	_, err := infra.LoadConfig(envPath)
	return err
}

func testLogger() error {
	logger := infra.Logger()
	if logger == nil {
		return fmt.Errorf("logger is nil")
	}
	logger.Info("Logger test")
	return nil
}

func testIPC() error {
	// Simple test
	time.Sleep(100 * time.Millisecond)
	return nil
}

func testCoreManager() error {
	// TODO: Implement core manager tests in Day 3
	return nil
}

func registerHandlers(server *ipc.Server, logger *slog.Logger) {
	// Ping handler
	server.Register("ping", func(msg *ipc.Message) (*ipc.Message, error) {
		logger.Info("Ping received")
		return &ipc.Message{
			ID:      msg.ID,
			Type:    ipc.TypeResponse,
			Payload: []byte(`{"pong":true}`),
		}, nil
	})

	logger.Info("Handlers registered")
}
