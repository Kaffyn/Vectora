package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Kaffyn/Vectora/internal/db"
	"github.com/Kaffyn/Vectora/internal/infra"
	"github.com/Kaffyn/Vectora/internal/ipc"
)

// Version will be set during build
var Version = "0.1.0-dev"

func main() {
	testMode := flag.Bool("test", false, "Run in test mode (30 second timeout)")
	version := flag.Bool("version", false, "Print version and exit")
	logLevel := flag.String("log-level", "INFO", "Log level (DEBUG, INFO, WARN, ERROR)")
	flag.Parse()

	if *version {
		fmt.Printf("Vectora Daemon v%s\n", Version)
		os.Exit(0)
	}

	// Setup paths
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".Vectora/logs/daemon.log")
	dbPath := filepath.Join(homeDir, ".Vectora/db/vectora.db")

	// Initialize logger
	if err := infra.InitLogger(logPath, *logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	logger := infra.Logger()

	logger.Info("Vectora daemon starting", "version", Version)

	// Load configuration
	envPath := filepath.Join(homeDir, ".Vectora/.env")
	config, err := infra.LoadConfig(envPath)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("Configuration loaded", "config", config.String())

	// Validate configuration
	if err := config.Validate(); err != nil {
		logger.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	// Initialize database store
	store, err := db.NewStore(dbPath, logger)
	if err != nil {
		logger.Error("Failed to initialize database store", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	logger.Info("Database store initialized", "path", dbPath)

	// Create IPC server
	server := ipc.NewServerWithConfig(&ipc.ServerConfig{
		MaxClients: 50,
		Logger:     logger,
	})

	// Register handlers
	registerHandlers(server, store, logger, config)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Test mode
	if *testMode {
		logger.Info("Running in test mode (30 second timeout)")
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// Start IPC server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Listen(ctx)
	}()

	logger.Info("Vectora daemon ready and accepting connections")

	// Wait for signals or errors
	select {
	case sig := <-sigChan:
		logger.Info("Received signal", "signal", sig)
		cancel()
	case err := <-errChan:
		if err != context.Canceled && err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	logger.Info("Vectora daemon shutdown complete")
}

// registerHandlers registers all IPC handlers
func registerHandlers(server *ipc.Server, store *db.Store, logger *slog.Logger, config *infra.Config) {
	// Health check
	server.Register("ping", func(msg *ipc.Message) (*ipc.Message, error) {
		logger.Debug("Handling ping request", "id", msg.ID)
		payload := map[string]interface{}{
			"pong":      true,
			"timestamp": time.Now(),
		}
		msg.Payload, _ = json.Marshal(payload)
		response := &ipc.Message{
			ID:        msg.ID,
			Type:      ipc.TypeResponse,
			Payload:   msg.Payload,
			Timestamp: time.Now(),
		}
		return response, nil
	})

	// Get server info
	server.Register("server.info", func(msg *ipc.Message) (*ipc.Message, error) {
		logger.Debug("Handling server.info request")
		payload := map[string]interface{}{
			"version":     Version,
			"uptime":      time.Now().Unix(),
			"maxClients":  50,
			"environment": config.PreferredLLMProvider,
		}
		data, _ := json.Marshal(payload)
		return &ipc.Message{
			ID:        msg.ID,
			Type:      ipc.TypeResponse,
			Payload:   data,
			Timestamp: time.Now(),
		}, nil
	})

	// Get database stats
	server.Register("db.stats", func(msg *ipc.Message) (*ipc.Message, error) {
		logger.Debug("Handling db.stats request")
		stats := store.Stats()
		data, _ := json.Marshal(stats)
		return &ipc.Message{
			ID:        msg.ID,
			Type:      ipc.TypeResponse,
			Payload:   data,
			Timestamp: time.Now(),
		}, nil
	})

	// Save workspace
	server.Register("workspace.save", func(msg *ipc.Message) (*ipc.Message, error) {
		logger.Debug("Handling workspace.save request", "id", msg.ID)
		var req ipc.RequestPayload
		if err := msg.UnmarshalPayload(&req); err != nil {
			return nil, fmt.Errorf("invalid payload: %w", err)
		}

		if err := store.SaveWorkspace(context.Background(), req.Workspace, req.Data); err != nil {
			return ipc.NewErrorMessage(msg.ID, "save_error", err.Error()), nil
		}

		response := map[string]interface{}{
			"success":   true,
			"workspace": req.Workspace,
		}
		data, _ := json.Marshal(response)
		return &ipc.Message{
			ID:        msg.ID,
			Type:      ipc.TypeResponse,
			Payload:   data,
			Timestamp: time.Now(),
		}, nil
	})

	// Get workspace
	server.Register("workspace.get", func(msg *ipc.Message) (*ipc.Message, error) {
		logger.Debug("Handling workspace.get request", "id", msg.ID)
		var req ipc.RequestPayload
		if err := msg.UnmarshalPayload(&req); err != nil {
			return nil, fmt.Errorf("invalid payload: %w", err)
		}

		data, err := store.GetWorkspace(context.Background(), req.Workspace)
		if err != nil {
			return ipc.NewErrorMessage(msg.ID, "not_found", err.Error()), nil
		}

		return &ipc.Message{
			ID:        msg.ID,
			Type:      ipc.TypeResponse,
			Payload:   data,
			Timestamp: time.Now(),
		}, nil
	})

	// List workspaces
	server.Register("workspace.list", func(msg *ipc.Message) (*ipc.Message, error) {
		logger.Debug("Handling workspace.list request")
		workspaces, err := store.ListWorkspaces(context.Background())
		if err != nil {
			return ipc.NewErrorMessage(msg.ID, "list_error", err.Error()), nil
		}

		response := map[string]interface{}{
			"workspaces": workspaces,
			"count":      len(workspaces),
		}
		data, _ := json.Marshal(response)
		return &ipc.Message{
			ID:        msg.ID,
			Type:      ipc.TypeResponse,
			Payload:   data,
			Timestamp: time.Now(),
		}, nil
	})

	// Delete workspace
	server.Register("workspace.delete", func(msg *ipc.Message) (*ipc.Message, error) {
		logger.Debug("Handling workspace.delete request", "id", msg.ID)
		var req ipc.RequestPayload
		if err := msg.UnmarshalPayload(&req); err != nil {
			return nil, fmt.Errorf("invalid payload: %w", err)
		}

		if err := store.DeleteWorkspace(context.Background(), req.Workspace); err != nil {
			return ipc.NewErrorMessage(msg.ID, "delete_error", err.Error()), nil
		}

		response := map[string]interface{}{
			"success":   true,
			"workspace": req.Workspace,
		}
		data, _ := json.Marshal(response)
		return &ipc.Message{
			ID:        msg.ID,
			Type:      ipc.TypeResponse,
			Payload:   data,
			Timestamp: time.Now(),
		}, nil
	})

	logger.Info("Handlers registered", "count", 6)
}
