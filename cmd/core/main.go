package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/api/ipc"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/engine"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/Kaffyn/Vectora/core/policies"
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

var (
	embedInclude   string
	embedExclude   string
	embedWorkspace string
	embedForce     bool
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

func init() {
	cobra.MousetrapHelpText = ""

	// Embed flags
	embedCmd.Flags().StringVar(&embedInclude, "include", "", "Comma-separated file patterns to include (e.g. *.go,*.md,*.py)")
	embedCmd.Flags().StringVar(&embedExclude, "exclude", "", "Comma-separated patterns to exclude (e.g. node_modules,.git,*.log)")
	embedCmd.Flags().StringVar(&embedWorkspace, "workspace", "default", "Workspace ID for embedding isolation")
	embedCmd.Flags().BoolVar(&embedForce, "force", false, "Re-embed files even if already indexed")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(embedCmd)
	rootCmd.AddCommand(acpCmd)
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(trustCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(chatCmd)

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

func getGeminiProvider(ctx context.Context) (llm.Provider, error) {
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not configured. Set it in %%USERPROFILE%%\\.Vectora\\.env")
	}
	return llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
}

func runEmbed(rootPath string) error {
	ctx := context.Background()

	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fmt.Printf("Scanning: %s\n", absPath)
	fmt.Println()

	// Initialize stores
	kvStore, err := db.NewKVStore()
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			fmt.Println("Error: Vectora Core is running in the background and managing the database.")
			fmt.Println("To use the bulk embedding command, please stop the Core temporarily:")
			fmt.Println("   1. vectora stop")
			fmt.Println("   2. vectora embed")
			fmt.Println("   3. vectora start")
			return fmt.Errorf("database locked by core")
		}
		return fmt.Errorf("failed to init KV store: %w", err)
	}
	defer kvStore.Close()

	vecStore, err := db.NewVectorStore()
	if err != nil {
		return fmt.Errorf("failed to init vector store: %w", err)
	}

	// Get Gemini provider for embeddings
	provider, err := getGeminiProvider(ctx)
	if err != nil {
		return fmt.Errorf("embedding requires Gemini API: %w", err)
	}

	guardian := policies.NewGuardian(absPath)

	fmt.Printf("Embedding provider: %s\n", provider.Name())
	fmt.Println()

	// Parse include/exclude patterns
	var includePatterns []string
	if embedInclude != "" {
		includePatterns = strings.Split(embedInclude, ",")
		for i, p := range includePatterns {
			includePatterns[i] = strings.TrimSpace(p)
		}
	}

	var excludePatterns []string
	var unignorePatterns []string
	if embedExclude != "" {
		parts := strings.Split(embedExclude, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(p, "!") {
				unignorePatterns = append(unignorePatterns, strings.TrimPrefix(p, "!"))
			} else {
				excludePatterns = append(excludePatterns, p)
			}
		}
	}

	// Tenta carregar .embedignore
	ignorePath := filepath.Join(absPath, ".embedignore")
	if data, err := os.ReadFile(ignorePath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "!") {
				unignorePatterns = append(unignorePatterns, strings.TrimPrefix(line, "!"))
			} else {
				excludePatterns = append(excludePatterns, line)
			}
		}
	}

	// Collect files
	var filesToEmbed []string
	var filesSkipped int
	var filesAlreadyIndexed int

	err = filepath.WalkDir(absPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			relPath, _ := filepath.Rel(absPath, path)
			relPath = filepath.ToSlash(relPath)
			name := d.Name()

			// Always skip known large dirs
			if guardian.IsExcludedDir(name) {
				return filepath.SkipDir
			}

			// Check exclusions
			isIgnored := false
			for _, pattern := range excludePatterns {
				if m, _ := filepath.Match(pattern, relPath); m {
					isIgnored = true
				} else if m, _ := filepath.Match(pattern, name); m {
					isIgnored = true
				} else if strings.Contains(pattern, "/") {
					if strings.HasPrefix(relPath, pattern) {
						isIgnored = true
					} else {
						m, _ := filepath.Match(pattern+"/*", relPath)
						if m {
							isIgnored = true
						}
					}
				}
			}

			if isIgnored {
				for _, pattern := range unignorePatterns {
					if m, _ := filepath.Match(pattern, relPath); m {
						isIgnored = false
					} else if m, _ := filepath.Match(pattern, name); m {
						isIgnored = false
					} else if strings.Contains(pattern, "/") {
						if strings.HasPrefix(relPath, pattern) {
							isIgnored = false
						} else {
							m, _ := filepath.Match(pattern+"/*", relPath)
							if m {
								isIgnored = false
							}
						}
					}
				}
			}

			if isIgnored {
				return filepath.SkipDir
			}
			return nil
		}

		// Check Guardian blocks
		if guardian.IsProtected(path) {
			filesSkipped++
			return nil
		}

		// Check include patterns
		if len(includePatterns) > 0 {
			name := d.Name()
			matched := false
			for _, pattern := range includePatterns {
				if m, _ := filepath.Match(pattern, name); m {
					matched = true
					break
				}
			}
			if !matched {
				filesSkipped++
				return nil
			}
		}

		// Check exclude patterns
		relPath, _ := filepath.Rel(absPath, path)
		relPath = filepath.ToSlash(relPath)
		name := d.Name()

		isIgnored := false
		for _, pattern := range excludePatterns {
			if m, _ := filepath.Match(pattern, relPath); m {
				isIgnored = true
			} else if m, _ := filepath.Match(pattern, name); m {
				isIgnored = true
			} else if strings.Contains(pattern, "/") {
				if strings.HasPrefix(relPath, pattern) {
					isIgnored = true
				} else {
					m, _ := filepath.Match(pattern+"/*", relPath)
					if m {
						isIgnored = true
					}
				}
			}
		}

		if isIgnored {
			for _, pattern := range unignorePatterns {
				if m, _ := filepath.Match(pattern, relPath); m {
					isIgnored = false
				} else if m, _ := filepath.Match(pattern, name); m {
					isIgnored = false
				} else if strings.Contains(pattern, "/") {
					if strings.HasPrefix(relPath, pattern) {
						isIgnored = false
					} else {
						m, _ := filepath.Match(pattern+"/*", relPath)
						if m {
							isIgnored = false
						}
					}
				}
			}
		}

		if isIgnored {
			filesSkipped++
			return nil
		}

		// relPath already initialized above
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			filesSkipped++
			return nil
		}

		if !embedForce {
			existing, _ := kvStore.Get(ctx, "file_index", relPath)
			if existing != nil {
				// Compare hash to detect changes
				var entry db.FileIndexEntry
				if err := json.Unmarshal(existing, &entry); err == nil {
					contentHash := db.CalculateHash(string(content))
					if entry.ContentHash == contentHash {
						filesAlreadyIndexed++
						return nil
					}
				}
			}
		}

		filesToEmbed = append(filesToEmbed, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	if len(filesToEmbed) == 0 {
		fmt.Printf("OK: No new files to embed. (%d skipped, %d already indexed)\n", filesSkipped, filesAlreadyIndexed)
		return nil
	}

	fmt.Printf("INFO: Found %d files to embed (%d skipped, %d already indexed)\n", len(filesToEmbed), filesSkipped, filesAlreadyIndexed)
	fmt.Println()

	// Embed each file
	collectionName := "ws_" + embedWorkspace
	totalEmbedded := 0
	totalChunks := 0
	totalErrors := 0

	for i, filePath := range filesToEmbed {
		relPath, _ := filepath.Abs(filePath)
		if rel, err := filepath.Rel(absPath, filePath); err == nil {
			relPath = rel
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("  ERR %s: read error\n", relPath)
			totalErrors++
			continue
		}

		// Chunk the content directly
		chunks := chunkContent(string(content), 800, 100)
		fileChunks := 0

		// Detect language from extension
		language := "text"
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".go":
			language = "go"
		case ".py":
			language = "python"
		case ".js", ".ts":
			language = "javascript"
		case ".md":
			language = "markdown"
		}

		for j, chunk := range chunks {
			// Generate embedding via Gemini API
			vec, err := provider.Embed(ctx, chunk)
			if err != nil {
				fmt.Printf("  ERR %s[%d]: embed error: %v\n", relPath, j, err)
				totalErrors++
				break
			}

			docID := fmt.Sprintf("%s:%d", relPath, j)
			err = vecStore.UpsertChunk(ctx, collectionName, db.Chunk{
				ID:       docID,
				Content:  chunk,
				Metadata: map[string]string{"source": relPath, "filename": filepath.Base(filePath), "language": language},
				Vector:   vec,
			})
			if err != nil {
				fmt.Printf("  ERR %s[%d]: store error: %v\n", relPath, j, err)
				totalErrors++
				break
			}
			fileChunks++
		}

		if fileChunks > 0 {
			// Save index metadata
			contentHash := db.CalculateHash(string(content))
			entry := db.FileIndexEntry{
				AbsolutePath: filePath,
				ContentHash:  contentHash,
				SizeBytes:    int64(len(content)),
			}
			entryBytes, _ := json.Marshal(entry)
			kvStore.Set(ctx, "file_index", relPath, entryBytes)

			totalEmbedded++
			totalChunks += fileChunks
			fmt.Printf("  OK  %s → %d chunks\n", relPath, fileChunks)
		}

		// Progress indicator
		if (i+1)%10 == 0 || i+1 == len(filesToEmbed) {
			fmt.Printf("  ... %d/%d files processed\n", i+1, len(filesToEmbed))
		}
	}

	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Printf("  Embedding complete!\n")
	fmt.Printf("  OK  %d files embedded\n", totalEmbedded)
	fmt.Printf("  --- %d total chunks\n", totalChunks)
	fmt.Printf("  ERR %d errors\n", totalErrors)
	fmt.Printf("  DIR Workspace: %s\n", collectionName)
	fmt.Println("==========================================================")

	return nil
}

func chunkContent(content string, maxTokens int, overlap int) []string {
	if len(content) == 0 {
		return []string{}
	}

	// ~4 chars per token approximation
	maxChars := maxTokens * 4
	overlapChars := overlap * 4

	runes := []rune(content)
	totalRunes := len(runes)

	if maxChars >= totalRunes {
		return []string{content}
	}

	var chunks []string
	start := 0
	for start < totalRunes {
		end := start + maxChars
		if end > totalRunes {
			end = totalRunes
		}

		// Try to break at newline
		chunkEnd := end
		if end < totalRunes {
			for i := end; i < totalRunes && i < end+maxChars/2; i++ {
				if runes[i] == '\n' {
					chunkEnd = i + 1
					break
				}
			}
		}

		chunks = append(chunks, string(runes[start:chunkEnd]))
		start = chunkEnd - overlapChars
		if start < 0 {
			start = 0
		}
		if start >= chunkEnd {
			break
		}
	}

	return chunks
}

func runCore() {
	systemManager, err := vecos.NewManager()
	if err != nil {
		log.Fatalf("Critical Hardware OS Failure: %v", err)
	}

	if err := systemManager.EnforceSingleInstance(); err != nil {
		infra.NotifyOS("Vectora Running", "Vectora is already running.")
		os.Exit(0)
	}

	infra.SetupLogger()

	appDataDir, _ := systemManager.GetAppDataDir()
	envPath := filepath.Join(appDataDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		infra.Logger().Warn(fmt.Sprintf("Global config (.env) not found at: %s. Using defaults.", envPath))
	}

	infra.Logger().Info("Starting Vectora Core...")

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

	// ---- Core-hosted ACP Server ----
	go func() {
		var l net.Listener
		var err error
		if runtime.GOOS == "windows" {
			l, err = net.Listen("tcp", "127.0.0.1:42782")
		} else {
			sockPath := filepath.Join(appDataDir, "run", "vectora_acp.sock")
			os.Remove(sockPath)
			l, err = net.Listen("unix", sockPath)
		}
		if err == nil {
			infra.Logger().Info("✅ ACP Server listening on background socket/pipe")
			for {
				conn, err := l.Accept()
				if err != nil {
					continue
				}
				go func(c net.Conn) {
					defer c.Close()
					reader := bufio.NewReader(c)
					workspaceLine, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					workspace := strings.TrimSpace(workspaceLine)
					eng := initCoreClientEngine(appCtx, workspace, vecStore, kvStore)
					if eng == nil {
						return
					}
					acpSrv := acp.NewServer(eng)
					_ = acpSrv.RunStream(appCtx, reader, c)
				}(conn)
			}
		}
	}()

	// Dev HTTP bridge for debugging
	go ipcServer.StartDevHTTP(startPort)

	infra.NotifyOS("Vectora", "Operational Assistant.")
	tray.Setup()
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
	llmRouter := llm.NewRouter()
	if cfg.GeminiAPIKey != "" {
		p, _ := llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
		llmRouter.RegisterProvider("gemini", p, true)
	}
	if cfg.ClaudeAPIKey != "" {
		p, _ := llm.NewClaudeProvider(ctx, cfg.ClaudeAPIKey)
		llmRouter.RegisterProvider("claude", p, false)
	}

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

func runStop() {
	if os.Getenv("OS") == "Windows_NT" {
		exec.Command("taskkill", "/F", "/IM", "vectora.exe", "/FI", fmt.Sprintf("PID ne %d", os.Getpid())).Run()
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

	fmt.Printf("Query: %s\n", query)
	fmt.Printf("[session: %s]\n", conversationID)
	fmt.Print("Thinking...")

	client, err := ensureCoreConnected()
	if err != nil {
		fmt.Printf("\rError: %v\n", err)
		fmt.Println("Please try running: vectora start")
		return err
	}
	defer client.Close()

	// Persist user message
	appendConversationEntry(conversationID, "user", query)

	req := map[string]string{
		"workspace_id":    workspaceID,
		"query":           query,
		"conversation_id": conversationID,
	}

	var resp struct {
		Answer string `json:"answer"`
	}

	err = client.Send(ctx, "workspace.query", req, &resp)
	if err != nil {
		if strings.Contains(err.Error(), "No LLM provider has been configured") || strings.Contains(err.Error(), "provider_not_configured") {
			fmt.Println("\rError: Vectora requires an API key to work.")
			runSetup()
			return fmt.Errorf("please try your query again after configuration")
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

func runSetup() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n--- Vectora Initial Setup ---")
	fmt.Println("Which AI provider do you want to use?")
	fmt.Println("[1] Google Gemini (Recommended, accepts free keys)")
	fmt.Println("[2] Anthropic Claude")
	fmt.Print("Choice (1 or 2): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var envKey string
	if choice == "2" {
		envKey = "CLAUDE_API_KEY"
		fmt.Print("\nPaste your CLAUDE_API_KEY: ")
	} else {
		envKey = "GEMINI_API_KEY"
		fmt.Print("\nPaste your GEMINI_API_KEY: ")
	}

	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)

	systemManager, _ := vecos.NewManager()
	appDataDir, _ := systemManager.GetAppDataDir()
	envPath := filepath.Join(appDataDir, ".env")

	f, err := os.OpenFile(envPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		f.WriteString(fmt.Sprintf("\n%s=%s\n", envKey, key))
		f.Close()
		fmt.Println("Key safely saved to", envPath)
	}

	// Auto Restart
	fmt.Println("Restarting Core with new API Key...")
	runStop()
	spawnDetached()
}

func initEngine(ctx context.Context, workspace string) (*engine.Engine, func(), error) {
	absPath, err := filepath.Abs(workspace)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid workspace path: %w", err)
	}

	cfg := infra.LoadConfig()

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

	llmRouter := llm.NewRouter()
	if cfg.GeminiAPIKey != "" {
		p, _ := llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
		llmRouter.RegisterProvider("gemini", p, true)
	}
	if cfg.ClaudeAPIKey != "" {
		p, _ := llm.NewClaudeProvider(ctx, cfg.ClaudeAPIKey)
		llmRouter.RegisterProvider("claude", p, false)
	}

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
