// Package main implements the Vectora ACP Agent binary.
// This binary is invoked by IDE clients (VS Code, Claude Code, Antigravity, etc.)
// to provide agent-like functionality through the Agent Client Protocol (ACP).
//
// Unlike `vectora` CLI which is interactive, `vectora-agent` runs as a subprocess
// and communicates via JSON-RPC on stdin/stdout with the parent IDE client.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
)

func main() {
	// Parse flags
	verbose := flag.Bool("v", false, "Enable verbose logging")
	logFormat := flag.String("log", "text", "Log format: text or json")
	flag.Parse()

	// Setup logger
	var logHandler slog.Handler
	if *logFormat == "json" {
		logHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	if *verbose {
		logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	logger := slog.New(logHandler)

	// Initialize app manager
	osMgr, err := vecos.NewManager()
	if err != nil {
		logger.Error("Failed to initialize OS manager", slog.Any("error", err))
		os.Exit(1)
	}

	// Setup data directories
	appDataDir, err := osMgr.GetAppDataDir()
	if err != nil {
		logger.Error("Failed to get app data directory", slog.Any("error", err))
		os.Exit(1)
	}

	dbPath := appDataDir + "/vectora.db"
	indexPath := appDataDir + "/index"

	// Initialize databases
	kvStore, err := db.NewKVStoreAtPath(dbPath)
	if err != nil {
		logger.Error("Failed to initialize KV store", slog.Any("error", err))
		os.Exit(1)
	}
	defer kvStore.Close()

	vecStore, err := db.NewVectorStoreAtPath(indexPath)
	if err != nil {
		logger.Error("Failed to initialize vector store", slog.Any("error", err))
		os.Exit(1)
	}

	// Initialize LLM router (providers will be registered separately)
	router := llm.NewRouter()

	// Initialize message service
	msgService := llm.NewMessageService(kvStore)

	// Create ACP agent
	agent := acp.NewVectoraAgent(
		"vectora-agent",
		getVersion(),
		kvStore,
		vecStore,
		router,
		msgService,
		logger,
	)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received signal, shutting down", slog.String("signal", sig.String()))
		cancel()
	}()

	// Start ACP agent (blocks until client disconnects)
	if err := acp.StartACPAgent(ctx, agent, logger); err != nil {
		logger.Error("ACP agent error", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("ACP agent shutdown complete")
}

func getVersion() string {
	// This could be populated at build time with -ldflags
	// For now, return a default version
	return "0.1.0"
}
