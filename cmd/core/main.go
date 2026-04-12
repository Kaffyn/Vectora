package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Kaffyn/Vectora/core/api/ipc"
	"github.com/Kaffyn/Vectora/core/api/mcp"
	"github.com/Kaffyn/Vectora/core/crypto"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/engine"
	"github.com/Kaffyn/Vectora/core/i18n"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/manager"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/Kaffyn/Vectora/core/policies"
	"github.com/Kaffyn/Vectora/core/service/singleton"
	"github.com/Kaffyn/Vectora/core/tools"
	"github.com/Kaffyn/Vectora/core/tray"
	"github.com/inconshreveable/mousetrap"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var startPort int
var startDetached bool

var rootCmd = &cobra.Command{
	Use:     "vectora",
	Long:    "Vectora is an AI-powered engineering assistant with knowledge base management.",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Vectora background service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting Vectora service...")
		runStop()
		if err := spawnDetached(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start background service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Vectora service restarted.")
	},
}

var resetHard bool
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Destroy all Vectora knowledge bases",
	Run: func(cmd *cobra.Command, args []string) {
		if !resetHard {
			fmt.Fprintln(os.Stderr, "Error: This command permanently deletes all indexed data and chat histories.")
			fmt.Fprintln(os.Stderr, "Run with --hard to confirm.")
			os.Exit(1)
		}
		fmt.Println("Stopping Core before data wipe...")
		runStop()

		systemManager, _ := vecos.NewManager()
		appDataDir, _ := systemManager.GetAppDataDir()
		dataDir := filepath.Join(appDataDir, "data")

		fmt.Println("Wiping all knowledge bases at", dataDir)
		os.RemoveAll(dataDir)

		fmt.Println("Database completely destroyed. Restarting fresh instance...")
		if err := spawnDetached(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start background service: %v\n", err)
		}
	},
}

var startCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"core"},
	Short:   "Start background service (Tray)",
	Long:    "Start the Vectora service as a background process with the system tray icon.",
	Run: func(cmd *cobra.Command, args []string) {
		if startDetached {
			// We ARE the background process — run the core directly.
			runCore()
			return
		}

		// Spawn ourselves as a detached background process
		if err := spawnDetached(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start background service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Vectora service started in background.")
		fmt.Println("Use 'vectora status' to check health, 'vectora stop' to shutdown.")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Verify health of micro-services",
	Long:  "Check if the Vectora core is running and responsive.",
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Shutdown background service",
	Long:  "Stop the running Vectora core.",
	Run: func(cmd *cobra.Command, args []string) {
		runStop()
	},
}

var acpCmd = &cobra.Command{
	Use:   "acp",
	Short: "Start ACP server over stdio",
	Long: `Start the Vectora ACP (Agent Client Protocol) server over stdin/stdout.
This allows code editors (VS Code, JetBrains, Zed) to connect to Vectora
as an AI coding agent via the ACP protocol.

The server reads JSON-RPC 2.0 messages from stdin and writes responses to stdout.
All logging goes to stderr to avoid interfering with the protocol.

Usage:
  vectora acp              # Start ACP server for current directory
  vectora acp /path/to/proj # Start ACP server for specific workspace
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace := "."
		if len(args) > 0 {
			workspace = args[0]
		}
		return runAcp(workspace)
	},
}

var mcpCmd = &cobra.Command{
	Use:   "mcp [workspace]",
	Short: "Start MCP server over stdio",
	Long: `Start the Vectora MCP (Model Context Protocol) server over stdin/stdout.
This allows Claude Code and other MCP-compatible clients to connect to Vectora
as a specialized sub-agent for code analysis, semantic search, and RAG.

The server reads JSON-RPC 2.0 messages from stdin and writes responses to stdout.
All logging goes to stderr to avoid interfering with the protocol.

The MCP server exposes 11 specialized embedding and analysis tools:
  - embed, search_database, web_search_and_embed, web_fetch_and_embed
  - analyze_code_patterns, knowledge_graph_analysis, doc_coverage_analysis
  - test_generation, bug_pattern_detection, plan_mode, refactor_with_context

Usage:
  vectora mcp /path/to/project  # Start MCP server for specified workspace

Configure in Claude Code settings.json:
  {
    "mcpServers": {
      "vectora": {
        "command": "vectora",
        "args": ["mcp", "/absolute/path/to/workspace"]
      }
    }
  }
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace := "."
		if len(args) > 0 {
			workspace = args[0]
		}
		return runMcp(workspace)
	},
}

var (
	embedInclude   string
	embedExclude   string
	embedWorkspace string
	embedForce     bool
	embedDetached  bool
)

var embedCmd = &cobra.Command{
	Use:   "embed [path]",
	Short: "Embed files into vector store via LLM provider",
	Long: `Embed files into the local vector database using the configured LLM provider's embedding API.
Files are chunked, embedded via Gemini/Claude remote API, and stored locally in chromem-go.

Examples:
  vectora embed                        # Embed current directory
  vectora embed --include *.go,*.md    # Only embed Go and Markdown files
  vectora embed --exclude node_modules # Exclude node_modules directory
  vectora embed --force                # Re-embed already indexed files
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rootPath := "."
		if len(args) > 0 {
			rootPath = args[0]
		}
		return runEmbed(rootPath)
	},
}

var askCmd = &cobra.Command{
	Use:   "ask [query]",
	Short: "Query Vectora via CLI",
	Long:  "Ask a question to the Vectora RAG engine and get a direct response in the terminal.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAsk(strings.Join(args, " "))
	},
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available LLM models",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runModelsList()
	},
}

func init() {
	cobra.MousetrapHelpText = ""

	// Embed flags
	embedCmd.Flags().StringVar(&embedInclude, "include", "", "Comma-separated file patterns to include (e.g. *.go,*.md,*.py)")
	embedCmd.Flags().StringVar(&embedExclude, "exclude", "", "Comma-separated patterns to exclude (e.g. node_modules,.git,*.log)")
	embedCmd.Flags().StringVar(&embedWorkspace, "workspace", "default", "Workspace ID for embedding isolation")
	embedCmd.Flags().BoolVar(&embedForce, "force", false, "Re-embed files even if already indexed")
	embedCmd.Flags().BoolVarP(&embedDetached, "background", "d", false, "Start embedding in background and exit")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(embedCmd)
	rootCmd.AddCommand(acpCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(trustCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(modelsCmd)

	resetCmd.Flags().BoolVar(&resetHard, "hard", false, "Confirm irreversible deletion")

	startCmd.Flags().IntVar(&startPort, "port", 42780, "Custom port for dev HTTP bridge")
	startCmd.Flags().BoolVar(&startDetached, "detached", false, "Run as detached background process (internal use)")
	startCmd.Flags().MarkHidden("detached")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if mousetrap.StartedByExplorer() && len(os.Args) <= 1 {
		// Se foi clicado como executável no Windows (e não pelo terminal), o usuário
		// espera que inicie em background silenciosamente.
		if err := spawnDetached(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to silence spawn: %v\n", err)
		}
		os.Exit(0)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runEmbed(rootPath string) error {
	ctx := context.Background()

	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fmt.Printf("Scanning: %s\n", absPath)
	fmt.Println()

	client, err := ensureCoreConnected()
	if err != nil {
		return err
	}
	defer client.Close()

	// Initializar o workspace no Core antes de qualquer comando (Phase 13 MTP)
	_, err = initWorkspace(client, absPath)
	if err != nil {
		return err
	}

	if !embedDetached {
		client.OnEvent = func(method string, payload json.RawMessage) {
			if method == "embed.progress" {
				var prog engine.EmbedProgress
				if err := json.Unmarshal(payload, &prog); err == nil {
					if prog.IsComplete {
						if prog.HasError {
							fmt.Printf("\r\033[2K  ❌ Embedding Error: %s\n", prog.ErrorMsg)
						} else {
							fmt.Println("\n==========================================================")
							fmt.Printf("  Embedding complete!\n")
							fmt.Printf("  OK  %d files embedded (%d skipped, %d already indexed)\n", prog.TotalEmbedded, prog.FilesSkipped, prog.FilesAlready)
							fmt.Printf("  --- %d total chunks\n", prog.TotalChunks)
							fmt.Printf("  ERR %d errors\n", prog.TotalErrors)
							fmt.Printf("  DIR Workspace: ws_%s\n", embedWorkspace)
							fmt.Println("==========================================================")
						}
						os.Exit(0)
					}

					if prog.HasError {
						fmt.Printf("\r\033[2K  ERR [%d/%d] %s: %s\n", prog.CurrentIdx+1, prog.TotalFiles, prog.CurrentFilePath, prog.ErrorMsg)
					} else if prog.FileChunks > 0 {
						fmt.Printf("\r\033[2K  ✅ [%d/%d] %s → %d chunks (%.1fs)\n", prog.CurrentIdx+1, prog.TotalFiles, prog.CurrentFilePath, prog.FileChunks, prog.ElapsedSeconds)
					}

					// Always reprint the status line mapping the current ongoing operation so it is at the bottom
					if !prog.HasError && prog.FileChunks == 0 {
						fmt.Printf("\r\033[2K  ⏳ [%d/%d] Embedding: %s [%.1fs]", prog.CurrentIdx+1, prog.TotalFiles, prog.DisplayPath, prog.ElapsedSeconds)
					}
				}
			}
		}
	}

	reqPayload := map[string]any{
		"rootPath":  absPath,
		"include":   embedInclude,
		"exclude":   embedExclude,
		"workspace": embedWorkspace,
		"force":     embedForce,
	}

	var resp struct {
		Started bool `json:"started"`
	}

	err = client.Send(ctx, "workspace.embed.start", reqPayload, &resp)
	if err != nil {
		return fmt.Errorf("failed to start embed job: %v", err)
	}

	if embedDetached {
		fmt.Println("✅ Embedding task submitted successfully.")
		fmt.Println("The job is now running seamlessly in the background Vectora Core.")
		fmt.Println("You can safely close this terminal.")
		return nil
	}

	// Keep alive waiting for events
	select {}
}

func runCore() {
	s := singleton.New("Vectora")
	if err := s.TryLock(); err != nil {
		_ = infra.NotifyOS("Vectora Running", "Another instance of Vectora is already active.")
		fmt.Println("Error: Another instance is already running.")
		os.Exit(0)
	}
	// Ensure lock is always released — covers panics and unexpected exits.
	defer func() {
		if r := recover(); r != nil {
			infra.Logger().Error(fmt.Sprintf("Vectora Core panicked: %v", r))
			_ = s.Unlock()
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		infra.Logger().Info("Shutting down Vectora Core...")
		_ = s.Unlock()
		os.Exit(0)
	}()

	systemManager, err := vecos.NewManager()
	if err != nil {
		log.Fatalf("Critical Hardware OS Failure: %v", err)
	}

	infra.SetupLogger()
	i18n.SetLanguage(systemManager.GetSystemLanguage())

	// Load configuration using the official standardized paths (MTP Phase 13)
	envPath := infra.GetConfigPath()
	if err := godotenv.Overload(envPath); err != nil {
		if !os.IsNotExist(err) {
			infra.Logger().Warn(fmt.Sprintf("Failed to load .env from %s: %v", envPath, err))
		}
	}

	infra.Logger().Info("Starting Vectora Core...")
	appCtx := context.Background()

	cfg := infra.LoadConfig()
	prefs := infra.LoadPreferences()
	llmRouter := llm.SetupRouter(appCtx, cfg, prefs)

	// Initialize Multi-Tenancy Managers (MTP Phase 13)
	tenantMgr := manager.NewTenantManager(manager.EvictionPolicy{
		IdleTimeout: 30 * time.Minute,
		MaxTenants:  10,
	}, llmRouter)
	tenantMgr.StartEvictionRoutine()

	resourcePool := manager.NewResourcePool(manager.ResourceConfig{
		MaxParallelLLMPerTenant: 2,
		MaxConcurrentIndexing:   4,
	})

	ipcServer, _ := ipc.NewServer(tenantMgr, resourcePool)

	// Global/Daemon-level stores (legacy/global config)
	kvStore, _ := db.NewKVStore()
	vecStore, _ := db.NewVectorStore()

	// Check global vector DB schema version
	if !vecStore.CheckAndUpdateSchema(appCtx) {
		infra.Logger().Warn("Global Vector DB schema mismatch detected",
			"expected_version", db.SchemaVersion,
			"recommendation", "Consider running 'vectora reset --hard' to re-index all workspaces")
	}

	// Initialize global workspace salter for per-installation workspace ID hashing
	appDataDir, _ := systemManager.GetAppDataDir()
	salter, err := crypto.NewWorkspaceSalter(appDataDir)
	if err != nil {
		infra.Logger().Warn(fmt.Sprintf("Failed to initialize workspace salter: %v", err))
		salter, _ = crypto.NewWorkspaceSalter(filepath.Join(appDataDir, "tmp"))
	}

	ipc.RegisterRoutes(ipcServer, kvStore, salter)
	go ipcServer.Start()

	// ---- Note: ACP Server ----
	// In Phase 7+, ACP is handled by cmd/acp or cmd/agent which are invoked
	// as separate processes by IDE clients (VS Code, Claude Code, etc.)
	// These use the Coder SDK and communicate via stdio.
	// The legacy socket-based ACP server has been removed.

	// Dev HTTP bridge for debugging
	go ipcServer.StartDevHTTP(startPort)

	infra.NotifyOS("Vectora", "Operational Assistant.")
	tray.Setup()
}

func runModelsList() error {
	ctx := context.Background()
	client, err := ensureCoreConnected()
	if err != nil {
		return err
	}
	defer client.Close()

	// Initialize workspace context
	cwd, _ := os.Getwd()
	_, _ = initWorkspace(client, cwd)

	var resp struct {
		Models []string `json:"models"`
	}

	fmt.Println("\n--- Vectora: Configured Models ---")

	err = client.Send(ctx, "models.list", map[string]any{}, &resp)
	if err != nil || len(resp.Models) == 0 {
		// If Core returned error or empty, it means ActiveProvider is not configured
		// Show fallbacks based on AGENTS.md
		fmt.Println("Status: [!] API Key not configured / Provider offline")
		fmt.Println("\nRecommended Models (from AGENTS.md):")
		fmt.Println("  Google Gemini (Default):")
		fmt.Println("    - gemini-3.1-pro-preview")
		fmt.Println("    - gemini-3-flash-preview")
		fmt.Println("  Anthropic Claude:")
		fmt.Println("    - claude-4.6-sonnet")
		fmt.Println("    - claude-4.6-opus")
		fmt.Println("\nRun 'vectora config' to set your API keys.")
	} else {
		for _, m := range resp.Models {
			fmt.Printf("- %-30s [Active]\n", m)
		}
	}
	fmt.Println()
	return nil
}

func initCoreClientEngine(ctx context.Context, workspace string, vecStore *db.ChromemStore, kvStore *db.BBoltStore) *engine.Engine {
	absPath, err := filepath.Abs(workspace)
	if err != nil {
		infra.Logger().Error("Rejecting ACP Connection: Invalid path", "path", workspace)
		return nil
	}

	trustList, err := readTrustList()
	if err == nil {
		trusted := false
		for _, p := range trustList {
			if strings.EqualFold(p, absPath) {
				trusted = true
				break
			}
		}
		if !trusted && len(trustList) > 0 { // Se a lista estiver vazia, defaults para permissivo ou pedimos setup depois. Mas como seguranca, vamos ser estritos se houver ao menos 1!
			// Update: Let's block unconditionally. But what if the user never setup? The CLI "trust" command is new.
			// Ideally, we block it and the IDE will receive EOF. If len == 0, we should maybe allow since it's the first time?
			// Let's only block if there's at least one trusted path, to gently deprecate the open-trust model, OR we block always and log.
			infra.Logger().Warn("Rejecting ACP Connection: Workspace path is not trusted. Use 'vectora trust add <path>'", "path", absPath)
			return nil
		}
	}

	cfg := infra.LoadConfig()
	prefs := infra.LoadPreferences()
	llmRouter := llm.SetupRouter(ctx, cfg, prefs)

	guardian := policies.NewGuardian(absPath)
	toolRegistry := tools.NewRegistry(absPath, guardian, kvStore)
	return engine.NewEngine(vecStore, kvStore, llmRouter, toolRegistry, guardian, nil)
}

func runStatus() {
	fmt.Println("--- VECTORA STATUS ---")
	client, err := ipc.NewClient()
	if err != nil || client.Connect() != nil {
		fmt.Println("Status: OFFLINE (Core not found)")
		return
	}
	defer client.Close()
	fmt.Println("Status: ONLINE (Core active)")
}

func ensureCoreConnected() (*ipc.Client, error) {
	client, err := ipc.NewClient()
	if err != nil {
		return nil, err
	}

	// Try initial connection
	if client.Connect() == nil {
		return client, nil
	}

	// If failed, try to start it
	fmt.Println("Vectora core is offline. Starting core...")
	if err := spawnDetached(); err != nil {
		return nil, fmt.Errorf("failed to auto-start core: %v", err)
	}

	// Wait and retry
	maxRetries := 15
	for i := 0; i < maxRetries; i++ {
		time.Sleep(300 * time.Millisecond)
		if client.Connect() == nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("core failed to start in time")
}

func initWorkspace(client *ipc.Client, rootPath string) (string, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	var resp ipc.WorkspaceInitResponse
	req := ipc.WorkspaceInitRequest{
		WorkspaceRoot: absPath,
		ProjectName:   filepath.Base(absPath),
	}

	err = client.Send(context.Background(), "workspace.init", req, &resp)
	if err != nil {
		return "", fmt.Errorf("workspace initialization failed: %w", err)
	}

	return resp.WorkspaceID, nil
}

func runStop() {
	if os.Getenv("OS") == "Windows_NT" {
		myPid := fmt.Sprintf("PID ne %d", os.Getpid())
		// Kill both canonical and release-named binaries to avoid residual processes
		exec.Command("taskkill", "/F", "/IM", "vectora.exe", "/FI", myPid).Run()
		exec.Command("taskkill", "/F", "/IM", "vectora-windows-amd64.exe", "/FI", myPid).Run()
		exec.Command("taskkill", "/F", "/IM", "vectora-windows-arm64.exe", "/FI", myPid).Run()
		fmt.Println("Vectora terminated.")
	} else {
		// Unix: kill any vectora processes except self
		exec.Command("pkill", "-f", "vectora").Run()
		fmt.Println("Vectora terminated.")
	}
}

// ---- ACP Server ----

func runAsk(query string) error {
	ctx := context.Background()

	// Derive stable IDs from current working directory
	cwd, _ := os.Getwd()
	absCwd, _ := filepath.Abs(cwd)
	conversationID := workspaceConversationID(absCwd)
	workspaceID := conversationID

	fmt.Printf("User: %s\n", query)
	fmt.Print("Vectora: Connecting to core...")

	client, err := ensureCoreConnected()
	if err != nil {
		fmt.Printf("\rVectora: Error: %v\n", err)
		fmt.Println("Please try running: vectora start")
		return err
	}
	fmt.Print("\rVectora: Thinking...             ")
	defer client.Close()

	// Initialize workspace (Required for activeTenant check in Phase 13)
	_, err = initWorkspace(client, absCwd)
	if err != nil {
		fmt.Printf("\rError: %v\n", err)
		return err
	}

	// Persist user message
	appendConversationEntry(conversationID, "user", query)

	req := map[string]string{
		"workspace_id":    workspaceID,
		"query":           query,
		"conversation_id": conversationID,
	}

	var resp struct {
		Answer string `json:"answer"`
		Model  string `json:"model"`
	}

	err = client.Send(ctx, "workspace.query", req, &resp)
	if err != nil {
		if strings.Contains(err.Error(), "No LLM provider has been configured") || strings.Contains(err.Error(), "provider_not_configured") {
			fmt.Println("\rError: Vectora requires an API key to work.")
			runConfigInteractive()

			// Trigger a reload in the Core after config
			_ = client.Send(ctx, "provider.reload", map[string]any{}, nil)

			fmt.Println("Configuration updated. Retrying query...")
			return runAsk(query) // Recursive retry
		}
		fmt.Println("\rError while querying Vectora:", err)
		return err
	}

	// Persist assistant response
	appendConversationEntry(conversationID, "assistant", resp.Answer)

	fmt.Print("\r")              // Clear thinking line
	fmt.Printf("\r%-40s\r", " ") // wipe the line
	fmt.Println("Vectora:", resp.Answer)
	return nil
}


func initEngine(ctx context.Context, workspace string) (*engine.Engine, func(), error) {
	absPath, err := filepath.Abs(workspace)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid workspace path: %w", err)
	}

	cfg := infra.LoadConfig()
	prefs := infra.LoadPreferences()

	kvStore, err := db.NewKVStore()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init KV store: %w", err)
	}

	cleanup := func() {
		kvStore.Close()
	}

	vecStore, err := db.NewVectorStore()
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to init vector store: %w", err)
	}

	llmRouter := llm.SetupRouter(ctx, cfg, prefs)

	guardian := policies.NewGuardian(absPath)
	toolRegistry := tools.NewRegistry(absPath, guardian, kvStore)

	eng := engine.NewEngine(vecStore, kvStore, llmRouter, toolRegistry, guardian, nil)

	return eng, cleanup, nil
}

func runAcp(workspace string) error {
	var conn net.Conn
	var err error

	if runtime.GOOS == "windows" {
		conn, err = net.Dial("tcp", "127.0.0.1:42782")
	} else {
		systemManager, _ := vecos.NewManager()
		appDataDir, _ := systemManager.GetAppDataDir()
		sockPath := filepath.Join(appDataDir, "run", "vectora_acp.sock")
		conn, err = net.Dial("unix", sockPath)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Vectora core is not running!")
		fmt.Fprintln(os.Stderr, "Please start the core first: vectora start")
		os.Exit(1)
	}
	defer conn.Close()

	if workspace == "" {
		workspace = "."
	}
	wsAbs, _ := filepath.Abs(workspace)
	fmt.Fprintf(conn, "%s\n", wsAbs)

	// Cópia bidirecional (Bridge)
	go io.Copy(conn, os.Stdin)
	io.Copy(os.Stdout, conn)

	return nil
}

func runMcp(workspace string) error {
	ctx := context.Background()

	if workspace == "" {
		workspace = "."
	}

	absPath, err := filepath.Abs(workspace)
	if err != nil {
		return fmt.Errorf("invalid workspace path: %w", err)
	}

	// Validate workspace path exists
	if info, err := os.Stat(absPath); err != nil || !info.IsDir() {
		return fmt.Errorf("workspace path does not exist or is not a directory: %s", absPath)
	}

	// Setup logger (to stderr to not interfere with stdio protocol)
	infra.SetupLogger()

	// Load configuration
	envPath := infra.GetConfigPath()
	if err := godotenv.Overload(envPath); err != nil {
		if !os.IsNotExist(err) {
			infra.Logger().Warn(fmt.Sprintf("Failed to load .env from %s: %v", envPath, err))
		}
	}

	infra.Logger().Info("Starting Vectora MCP server", "workspace", absPath)

	// Load configuration and preferences
	cfg := infra.LoadConfig()
	prefs := infra.LoadPreferences()

	// Setup LLM router
	llmRouter := llm.SetupRouter(ctx, cfg, prefs)

	// Initialize stores
	kvStore, err := db.NewKVStore()
	if err != nil {
		return fmt.Errorf("failed to initialize KV store: %w", err)
	}
	defer kvStore.Close()

	vecStore, err := db.NewVectorStore()
	if err != nil {
		return fmt.Errorf("failed to initialize vector store: %w", err)
	}
	defer vecStore.Close()

	// Setup guardian and tools registry
	guardian := policies.NewGuardian(absPath)
	toolsRegistry := tools.NewRegistry(absPath, guardian, kvStore)

	// Create Engine with proper tool registry
	eng := engine.NewEngine(vecStore, kvStore, llmRouter, toolsRegistry, guardian, nil)

	// Create MCP server - this will integrate the Engine with embedding tools
	mcpServer := mcp.NewVectoraMCPServer(
		"Vectora Core",
		"0.1.0",
		kvStore,
		vecStore,
		llmRouter,
		nil, // msgService not needed for MCP
		infra.Logger(),
	)

	// Register embedding tools in both MCP server and Engine's tool registry
	// This ensures tools are discoverable via tools/list
	mcp.RegisterEmbeddingTools(mcpServer, llmRouter, toolsRegistry)

	// Create stdio server with the engine that has the tools
	stdioServer := mcp.NewStdioServer(eng, infra.Logger())

	infra.Logger().Info("MCP server started, listening on stdio", "workspace", absPath)

	// Start server (this blocks until EOF or error)
	err = stdioServer.Start(ctx)
	if err != nil && err != context.Canceled {
		infra.Logger().Error("MCP server error", "error", err.Error())
		return err
	}

	infra.Logger().Info("MCP server shutdown gracefully")
	return nil
}
